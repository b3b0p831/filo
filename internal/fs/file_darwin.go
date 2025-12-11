package fs

import "path/filepath"

func IsHiddenFile(path string) (bool, error) {
	fileName := filepath.Base(path)

	return fileName[0] == '.', nil
}
