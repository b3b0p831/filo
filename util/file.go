package util

import (
	"fmt"
	"io/fs"
	"strings"
)

// FileObject represents a directory entry and its children.
// It provides a recursive view of a file system hierarchy.
type FileObject struct {
	path     string
	d        fs.DirEntry
	children []FileObject
}

// String implements fmt.Stringer, producing a hierarchical
// string representation of the FileObject tree.
func (f FileObject) String() string {
	return f.format(0)
}

// format is a helper that recursively formats the FileObject
// with indentation for readability.
func (f FileObject) format(indent int) string {
	var sb strings.Builder
	prefix := strings.Repeat("  ", indent)

	// Prefer fs.DirEntry.Name() when available, otherwise fallback to path
	if f.d != nil {
		if f.d.IsDir() {
			sb.WriteString(fmt.Sprintf("%s%s/\n", prefix, f.d.Name()))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s\n", prefix, f.d.Name()))
		}
	} else {
		sb.WriteString(fmt.Sprintf("%s%s\n", prefix, f.path))
	}

	for _, child := range f.children {
		sb.WriteString(child.format(indent + 1))
	}

	return sb.String()
}
