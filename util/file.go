package util

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
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
	ft.Index[rootPath] = ft.Root

	filepath.WalkDir(ft.Root.Path, func(path string, d fs.DirEntry, err error) error {
		currentNode := ft.Index[path]
		if currentNode == nil {
			currentNode = &FileNode{}
			ft.Index[path] = currentNode
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
