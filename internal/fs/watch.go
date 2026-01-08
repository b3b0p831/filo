package fs

import (
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"bebop831.com/filo/internal/config"
	"bebop831.com/filo/internal/util"

	"github.com/fsnotify/fsnotify"
)

// Watch new dirs, watching files is not reccomended in docs
func OnCreate(event *fsnotify.Event, watcher *fsnotify.Watcher) {
	info, err := os.Lstat(filepath.Clean(event.Name)) // Stat follows symlink, Lstat returns sysmlink info
	if err != nil {
		slog.Error(err.Error())
		return
	}

	if info.IsDir() {
		filepath.WalkDir(event.Name, func(path string, d fs.DirEntry, err error) error {
			if err == nil && d.IsDir() {
				watcher.Add(path)
			}
			return err
		})
	}
}

func WatchChanges(eventChan chan fsnotify.Event, exitChan chan struct{}, syncChan chan struct{}, cfg *config.Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error(err.Error())
	}

	watcher.Add(cfg.SourceDir)
	if err != nil {
		slog.Error(err.Error())
	}

	go OnCreate(&fsnotify.Event{Op: fsnotify.Create, Name: cfg.SourceDir}, watcher)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if !IsApprovedPath(event.Name) {
				continue
			}

			switch event.Op {
			case fsnotify.Create:
				slog.Info(fmt.Sprint(util.CreateColor(event.Op), " ", event.Name))
				go OnCreate(&event, watcher)

			case fsnotify.Rename:
				slog.Info(fmt.Sprint(util.RenameColor(event.Op), " ", event.Name))

			case fsnotify.Remove:
				slog.Info(fmt.Sprint(util.RemoveColor(event.Op), " ", event.Name))

			case fsnotify.Chmod:
				continue

			default:
				slog.Info(fmt.Sprint(event.Op, " ", event.Name))
			}

			eventChan <- event

		case err, _ := <-watcher.Errors:
			log.Println("error:", err)
			watcher.Close()
			return

		case <-syncChan:
			slog.Debug("received sync message")

		case <-exitChan:
			slog.Debug("Exiting WatchChanges goroutine...")
			watcher.Close()
			return
		}
	}
}
