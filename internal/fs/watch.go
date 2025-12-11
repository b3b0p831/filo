package fs

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"bebop831.com/filo/internal/config"
	"bebop831.com/filo/internal/util"

	"github.com/fsnotify/fsnotify"
)

// Watch new dirs, watching files is not reccomended in docs
func OnCreate(event *fsnotify.Event, watcher *fsnotify.Watcher, logger *log.Logger) {
	info, err := os.Lstat(filepath.Clean(event.Name)) // Stat follows symlink, Lstat returns sysmlink info
	if err != nil {
		logger.Println(err)
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

func WatchChanges(eventChan chan fsnotify.Event, exitChan chan struct{}, cfg *config.Config) {
	// Turns this into go routine so that it matches above(i.e WatchChanges)

	watcher, err := fsnotify.NewWatcher()
	watcher.Add(cfg.SourceDir)
	go OnCreate(&fsnotify.Event{Op: fsnotify.Create, Name: cfg.SourceDir}, watcher, cfg.Flogger)

	if err != nil {
		cfg.Flogger.Fatalln(err)
	}

	// TODO: Turns this into go routine, go WatchChanges
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			switch event.Op {
			case fsnotify.Create:
				cfg.Flogger.Println(util.CreateColor(event.Op), event.Name)
				go OnCreate(&event, watcher, cfg.Flogger)

			case fsnotify.Rename:
				cfg.Flogger.Println(util.RenameColor(event.Op), event.Name)

			case fsnotify.Remove:
				cfg.Flogger.Println(util.RemoveColor(event.Op), event.Name)

			// case fsnotify.Chmod:
			// 	continue

			default:
				if cfg.LogLevel == "debug" {
					cfg.Flogger.Println(event.Op, event.Name)
				}
			}

			eventChan <- event

		case err, ok := <-watcher.Errors:
			if !ok {
				exitChan <- struct{}{}
				return
			}
			cfg.Flogger.Println("error:", err)

		case <-exitChan:
		}
	}
}
