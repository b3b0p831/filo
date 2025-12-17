package fs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
)

var Mu *sync.Mutex

func init() {
	Mu = &sync.Mutex{}
}

// FileNode represents a directory entry and its children.
// It provides a recursive view of a file system hierarchy.
type FileNode struct {
	Path     string
	Entry    fs.DirEntry
	Parent   *FileNode
	Children []*FileNode
}

type FileTree struct {
	Root  *FileNode
	Index map[string]*FileNode
}

func BuildTree(rootPath string) *FileTree {
	ft := &FileTree{Index: make(map[string]*FileNode), Root: &FileNode{Path: rootPath}}

	filepath.WalkDir(ft.Root.Path, func(path string, d fs.DirEntry, err error) error {
		switch {
		//Nothing to walk
		case err != nil && path == rootPath:
			ft = nil
			slog.Error(err.Error())
			return filepath.SkipAll

		case err != nil:
			slog.Error(err.Error())
			return filepath.SkipDir
		}

		if IsApprovedPath(path) {

			currentNode := ft.Index[path]
			if currentNode == nil {
				if path == rootPath {
					currentNode = ft.Root
				} else {
					currentNode = &FileNode{}
					ft.Index[path] = currentNode
				}
			}

			currentNode.Path = path
			currentNode.Entry = d

			if currentNode.Entry.IsDir() {
				entries, err := os.ReadDir(path)
				if err != nil {
					slog.Error(err.Error())
					return filepath.SkipDir
				}

				currentNode.Children = make([]*FileNode, 0)
				for _, e := range entries {
					possiblePath := filepath.Join(path, e.Name())
					if IsApprovedPath(possiblePath) {
						childNode := &FileNode{Path: possiblePath, Entry: e, Parent: currentNode, Children: make([]*FileNode, 0)}
						currentNode.Children = append(currentNode.Children, childNode)
						ft.Index[childNode.Path] = childNode
					}
				}

				// You can get with this, or you can get with that
				slices.SortFunc(currentNode.Children, func(this, that *FileNode) int {
					return strings.Compare(this.Entry.Name(), that.Entry.Name())
				})
			}
		} else if d.IsDir() {
			slog.Warn(fmt.Sprint("Unapproved directory: ", path))
			return filepath.SkipDir
		}

		return nil
	})

	return ft
}

func IsApprovedPath(path string) bool {

	cleanedFilePath := filepath.Clean(path)
	isHidden, err := IsHiddenFile(cleanedFilePath)

	if err != nil {
		return false
	}

	if path == "" || isHidden {
		return false
	}

	return true
}

// Returns true wether or not 2 filenodes are the same
// For directories, this will recursilve check each child
func compareFileNodes(srcFileNode, tgtFileNode *FileNode) (bool, error) {

	if srcFileNode == nil || tgtFileNode == nil ||
		srcFileNode.Entry.IsDir() != tgtFileNode.Entry.IsDir() {
		return false, nil
	}

	const bufSize = 4 << 20 //4 MiB
	initSrcFileInfo, err := srcFileNode.Entry.Info()
	if err != nil {
		return false, err
	}

	initTgtFileInfo, err := tgtFileNode.Entry.Info()

	if err != nil {
		return false, err
	}

	if srcFileNode.Entry.IsDir() {
		for _, sc := range srcFileNode.Children {
			for _, tc := range tgtFileNode.Children {
				if srcFileNode.Entry.Name() == tgtFileNode.Entry.Name() {
					return compareFileNodes(sc, tc)
				}
			}
		}
	}

	for attempts := 0; attempts < 2; attempts++ {
		srcFile, err := os.Open(srcFileNode.Path)
		if err != nil {
			return false, err
		}

		tgtFile, err := os.Open(tgtFileNode.Path)
		if err != nil {
			srcFile.Close()
			return false, err
		}

		srcFileBuf := make([]byte, bufSize)
		tgtFileBuf := make([]byte, bufSize)

		for {
			srcBytesRead, srcReadErr := io.ReadFull(srcFile, srcFileBuf)
			tgtBytesRead, tgtReadErr := io.ReadFull(tgtFile, tgtFileBuf)

			// normalize EOFs
			// err is ErrUnexpected if not all bytes are read
			if errors.Is(srcReadErr, io.EOF) || errors.Is(srcReadErr, io.ErrUnexpectedEOF) {
				srcReadErr = io.EOF
			}

			if errors.Is(tgtReadErr, io.EOF) || errors.Is(tgtReadErr, io.ErrUnexpectedEOF) {
				tgtReadErr = io.EOF
			}

			if (srcReadErr != nil && srcReadErr != io.EOF) || (tgtReadErr != nil && tgtReadErr != io.EOF) {
				srcFile.Close()
				tgtFile.Close()
				return false, fmt.Errorf("%s %s", srcReadErr, tgtReadErr)
			}

			if srcBytesRead != tgtBytesRead || !bytes.Equal(srcFileBuf[:srcBytesRead], tgtFileBuf[:tgtBytesRead]) {
				return false, nil
			}

			if srcReadErr == io.EOF && tgtReadErr == io.EOF {
				break
			}

		}

		currentSrcFileInfo, err := srcFileNode.Entry.Info()

		if err != nil {
			return false, err
		}

		currentTgtFileInfo, err := tgtFileNode.Entry.Info()

		if err != nil {
			return false, err
		}
		srcFile.Close()
		tgtFile.Close()

		if initSrcFileInfo.Size() != currentSrcFileInfo.Size() ||
			!initSrcFileInfo.ModTime().Equal(currentSrcFileInfo.ModTime()) ||
			initTgtFileInfo.Size() != currentTgtFileInfo.Size() ||
			!initTgtFileInfo.ModTime().Equal(currentTgtFileInfo.ModTime()) {
			// changed: retry, continue, or return an "unstable" error
			continue
		}

		return true, nil
	}

	return false, fmt.Errorf("changes detected while comparing nodes %v with %v", srcFileNode, tgtFileNode)
}

func walkMissingInBinary(sourceRoot, targetRoot *FileNode, missingNodes map[string][]*FileNode, wg *sync.WaitGroup) {

	if sourceRoot == nil || targetRoot == nil {
		return
	}

	for _, srcChildNode := range sourceRoot.Children {

		wg.Go(func() {

			didContain := false
			tgtNodeIdx := sort.Search(len(targetRoot.Children), func(i int) bool {
				// Compare names first
				cmp := strings.Compare(targetRoot.Children[i].Entry.Name(), srcChildNode.Entry.Name())
				if cmp == 0 {
					// Found exact match
					return true
				}
				return cmp >= 0 // true once we've passed or matched the target
			})

			if tgtNodeIdx < len(targetRoot.Children) {
				tgtNode := targetRoot.Children[tgtNodeIdx]
				if tgtNode.Entry.IsDir() == srcChildNode.Entry.IsDir() {
					if tgtNode.Entry.IsDir() {
						slog.Debug(fmt.Sprintf("COMPARE %s <-> %s", srcChildNode.Path, tgtNode.Path))

						walkMissingInBinary(srcChildNode, tgtNode, missingNodes, wg)
						didContain = true
					} else {
						sameFilesB, err := compareFileNodes(srcChildNode, tgtNode)
						if err != nil {
							slog.Info(err.Error())
						}
						didContain = sameFilesB
					}
				}
			}

			if !didContain {
				fp := filepath.Clean(targetRoot.Path)

				Mu.Lock()
				tmpChildren := missingNodes[fp]
				missingNodes[fp] = append(tmpChildren, srcChildNode)
				Mu.Unlock()
			}
		})
	}
}

// Returns a map where the keys are paths located in otherTree, and the values are the missing children for that key
// For example, {"/mnt/media" : [tv, yt, movies]} means that directory "/mnt/media" in tgt is missing the children 'tv', 'yt', 'movies' which are present in src
func (t *FileTree) MissingIn(otherTree *FileTree, runAfter func()) map[string][]*FileNode {
	missing := make(map[string][]*FileNode)
	var wg sync.WaitGroup
	walkMissingInBinary(t.Root, otherTree.Root, missing, &wg)
	wg.Wait()

	if runAfter != nil {
		runAfter()
	}

	return missing
}

func copyChildren(rootPath string, children []*FileNode, wg *sync.WaitGroup) {

	slog.Debug(fmt.Sprint("rootPath: ", rootPath))
	slog.Debug(fmt.Sprint("children:", children))
	for _, cc := range children {
		tgtPath := filepath.Join(rootPath, cc.Entry.Name())

		if cc.Entry.IsDir() {
			dirInfo, _ := cc.Entry.Info()
			if err := os.Mkdir(tgtPath, dirInfo.Mode().Perm()); err != nil {
				slog.Error(err.Error())
				continue
			}
			slog.Debug(fmt.Sprint(cc.Path, " -> ", tgtPath))
			copyChildren(tgtPath, cc.Children, wg)
		} else {
			wg.Go(func() {
				if _, err := copyFile(cc.Path, tgtPath); err != nil {
					slog.Error(err.Error())
				}
			})
		}
	}
}

func copyFile(filePath string, tgtPath string) (int64, error) {

	srcReader, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		slog.Error(err.Error())
		return -1, err
	}
	defer srcReader.Close()

	tgtWriter, err := os.OpenFile(tgtPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error(err.Error())
		return -1, err
	}
	defer tgtWriter.Close()

	written, err := io.Copy(tgtWriter, srcReader)
	slog.Debug(fmt.Sprint(filePath, " -> ", tgtPath))
	return written, err
}

// IO Operations
func (t *FileTree) CopyMissing(missing map[string][]*FileNode) {

	var wg sync.WaitGroup
	for targetPath, currentChildren := range missing {
		copyChildren(targetPath, currentChildren, &wg)
	}
	wg.Wait()

}

func (n *FileNode) String() string {
	if n == nil {
		return "<nil FileNode>"
	}

	entry := "<nil>"
	if n.Entry != nil {
		entry = n.Entry.Name()
	}

	parent := "<nil>"
	if n.Parent != nil {
		parent = n.Parent.Path
	}

	children := "<nil>"
	if len(n.Children) > 0 {
		names := make([]string, 0, len(n.Children))
		for _, c := range n.Children {
			if c != nil && c.Entry != nil {
				names = append(names, c.Entry.Name())
			}
		}
		children = fmt.Sprintf("%v", names)
	}

	return fmt.Sprintf(
		"FileNode{Entry=%s, Parent=%s, Children=%s}",
		entry,
		parent,
		children,
	)
}

func (t *FileTree) String() string {
	var b strings.Builder
	printTree(t.Root, 0, &b)
	return b.String()
}

func printTree(n *FileNode, level int, b *strings.Builder) {
	if n == nil {
		return
	}

	fmt.Fprintf(b, "%s-%s\n", strings.Repeat(" ", level), n.Entry.Name())
	for _, c := range n.Children {
		printTree(c, level+2, b)
	}
}
