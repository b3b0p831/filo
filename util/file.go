package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

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

type FileChange struct {
	Op   string
	Path string
}

func BuildTree(rootPath string) *FileTree {
	ft := &FileTree{Index: make(map[string]*FileNode), Root: &FileNode{Path: rootPath}}
	//ft.Index[rootPath] = ft.Root

	filepath.WalkDir(ft.Root.Path, func(path string, d fs.DirEntry, err error) error {

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
				log.Println(err)
				return filepath.SkipDir
			}

			currentNode.Children = make([]*FileNode, 0)
			for _, e := range entries {
				childNode := &FileNode{Path: filepath.Join(path, e.Name()), Entry: e, Parent: currentNode, Children: make([]*FileNode, 0)}
				currentNode.Children = append(currentNode.Children, childNode)
				ft.Index[childNode.Path] = childNode
			}
		}

		return nil
	})

	return ft
}

func CompareFileNodes(srcFileNode, tgtFileNode *FileNode) (bool, error) {

	if srcFileNode == nil || tgtFileNode == nil {
		return false, nil
	}

	const bufSize = 4 << 20 //4 MiB
	initSrcFileInfo, err := srcFileNode.Entry.Info()

	if err != nil {
		log.Println(err)
		return false, err
	}

	initTgtFileInfo, err := tgtFileNode.Entry.Info()

	if err != nil {
		log.Println(err)
		return false, err
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
				if srcReadErr != nil && srcReadErr != io.EOF {
					return false, srcReadErr
				}
				if tgtReadErr != nil && tgtReadErr != io.EOF {
					return false, tgtReadErr
				}
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
			log.Println(err)
			return false, err
		}

		currentTgtFileInfo, err := tgtFileNode.Entry.Info()

		if err != nil {
			log.Println(err)
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
	return false, nil
}

func GetMissingRecurse(sourceRoot, targetRoot *FileNode, missingNodes map[string][]*FileNode) {

	if sourceRoot == nil || targetRoot == nil {
		return
	}

	for _, child := range sourceRoot.Children {
		var tgtNode *FileNode

		if slices.ContainsFunc(targetRoot.Children,
			func(n *FileNode) bool {

				if child.Entry.Name() == n.Entry.Name() {
					tgtNode = n
					sameFilesB, err := CompareFileNodes(child, tgtNode)
					if err != nil {
						log.Println(err)
					}
					if sameFilesB {
						log.Println(n.Path, "=", child.Path)
					} else {
						log.Println(n.Path, "!=", child.Path)
					}
					return true
				}

				return false
			}) {

			GetMissingRecurse(child, tgtNode, missingNodes)
		} else {
			tmpChildren := missingNodes[targetRoot.Path]
			missingNodes[targetRoot.Path] = append(tmpChildren, child)
		}
	}
}

// Returns a map where the keys are paths located in otherTree, and the values are the missing children for that key
// For example, "/mnt/media" -> [tv, yt, movies] means that "/mnt/media" is missing the children 'tv', 'yt', 'movies'
func (t *FileTree) GetMissing(otherTree *FileTree) map[string][]*FileNode {
	missing := make(map[string][]*FileNode)
	GetMissingRecurse(t.Root, otherTree.Root, missing)
	return missing
}

func copyChildren(rootPath string, children []*FileNode) {

	for _, cc := range children {
		srcFilePath := filepath.Join(cc.Parent.Path, cc.Entry.Name())
		tgtTmpPath := filepath.Join(rootPath, cc.Entry.Name())
		log.Println(srcFilePath, "->", tgtTmpPath)
	}

	// if cc.Entry.IsDir() {
	// 	// Create Directory and recurse
	// 	_, err := os.Stat(tgtTmpPath)
	// 	if err == nil {
	// 		log.Println()
	// 	}

	// } else {
	// 	// Handle File Copy
	// 	log.Println(srcFilePath, "->", tgtTmpPath)
	// }
}

// IO Operations
func (t *FileTree) CopyMissing(missing map[string][]*FileNode) {
	for targetPath, currentChildren := range missing {
		log.Println(targetPath)
		log.Println(currentChildren)
		go copyChildren(targetPath, currentChildren)
		// for _, cc := range currentChildren {
		// 	srcFilePath := filepath.Join(cc.Parent.Path, cc.Entry.Name())
		// 	tgtTmpPath := filepath.Join(targetPath, cc.Entry.Name())
		// 	log.Println(srcFilePath, "->", tgtTmpPath)
		// }
	}
}

func (n *FileNode) String() string {
	var sb strings.Builder

	sb.WriteString("FileNode\n")
	//	sb.WriteString(fmt.Sprintf("  Path: %s\n", n.Path))

	if n.Entry != nil {
		sb.WriteString(fmt.Sprintf("  Entry: %s\n", n.Entry.Name()))
	} else {
		sb.WriteString("  Entry: <nil>\n")
	}

	if n.Parent != nil {
		sb.WriteString(fmt.Sprintf("  Parent: %s\n", n.Parent.Path))
	} else {
		sb.WriteString("  Parent: <nil>\n")
	}

	if n.Children != nil {
		tmpChildren := make([]string, len(n.Children))
		for _, c := range n.Children {
			if c.Entry != nil {
				tmpChildren = append(tmpChildren, c.Entry.Name())
			}
		}
		sb.WriteString(fmt.Sprintf("  Children: %v\n", tmpChildren))
	}

	return sb.String()
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
