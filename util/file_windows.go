// Source - https://stackoverflow.com/a
// Posted by rustyx, modified by community. See post 'Timeline' for change history
// Retrieved 2025-12-06, License - CC BY-SA 4.0

package util

import (
	"syscall"
)

func IsHiddenFile(filename string) (bool, error) {
	pointer, err := syscall.UTF16PtrFromString(filename)
	if err != nil {
		return false, err
	}
	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false, err
	}
	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
