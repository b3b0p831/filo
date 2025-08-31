package util

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"time"

	"bebop831.com/filo/config"
)

func populateChildren(currentFile *FileObject) {
	filepath.WalkDir(currentFile.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Println("[ERROR]", err)
			return filepath.SkipDir
		}
		if currentFile.path != path {
			childFileObj := FileObject{path: path, d: d}
			currentFile.children = append(currentFile.children, childFileObj)
		}
		return nil
	})
}

func Sync(eventChan <-chan struct{}, exit <-chan struct{}, syncChan chan struct{}, cfg *config.Config) {
	minInterval, err := GetTimeInterval(cfg.SyncDelay)
	if err != nil {
		log.Fatalf("failed to parse sync_delay in config: invalid value '%v' (must use s, m, or h).", cfg.SyncDelay)
	}

	var lastEvent time.Time

	for {
		select {
		case <-eventChan:
			lastEvent = time.Now()

		case <-time.After(minInterval):
			if !lastEvent.IsZero() && time.Since(lastEvent) >= minInterval {

				var files []FileObject
				var fInfo map[string]fs.FileInfo = make(map[string]fs.FileInfo)
				walkErr := filepath.WalkDir(cfg.SourceDir, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						log.Println(err)
						return filepath.SkipDir
					}
					currentFileObj := FileObject{path: path, d: d}

					curFileInfo, err := currentFileObj.d.Info()
					if err != nil {
						return err
					}

					fInfo[path] = curFileInfo
					if d.IsDir() && currentFileObj.path != cfg.SourceDir {
						populateChildren(&currentFileObj)
					} else {
						currentFileObj.children = nil
					}

					files = append(files, currentFileObj)
					return nil
				})
				if walkErr != nil {
					log.Println(walkErr)
					log.Printf("Error occured while walking '%v', Skipping sync...", cfg.SourceDir)
					return
				}
				log.Printf("Syncing started: %v -> %v...\n", cfg.SourceDir, cfg.TargetDir)

				for i := range files {
					fmt.Print(files[i])
				}

				if syncChan != nil {
					syncChan <- struct{}{}
				}

				log.Printf("Sync completed successfully")
				lastEvent = time.Time{} // reset
			}

		case <-exit:
			return
		}
	}
}
