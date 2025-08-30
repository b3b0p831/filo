package util

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func OnCreate(event *fsnotify.Event, watcher *fsnotify.Watcher) {
	info, err := os.Stat(filepath.Clean(event.Name))
	if err != nil {
		log.Println(err)
		return
	}

	if info.IsDir() {
		filepath.WalkDir(event.Name, func(path string, d fs.DirEntry, err error) error {
			strings.Split(filepath.Clean(path), string(os.PathSeparator))
			if err == nil && d.IsDir() {
				watcher.Add(path)
			}
			return err
		})
	}
}
