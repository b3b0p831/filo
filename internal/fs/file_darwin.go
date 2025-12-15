package fs

import "path/filepath"

func IsHiddenFile(path string) (bool, error) {
	return filepath.Base(path)[0] == '.', nil
}
