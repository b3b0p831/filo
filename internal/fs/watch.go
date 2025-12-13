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
	// Turns this into go routine so that it matches above(i.e WatchChanges)

	watcher, err := fsnotify.NewWatcher()
	watcher.Add(cfg.SourceDir)
	go OnCreate(&fsnotify.Event{Op: fsnotify.Create, Name: cfg.SourceDir}, watcher)

	if err != nil {
		slog.Error(err.Error())
	}

	// TODO: Turns this into go routine, go util.WatchChanges
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
				slog.Debug(fmt.Sprint(event.Op, " ", event.Name))
			}

			eventChan <- event

		case err, ok := <-watcher.Errors:
			if !ok {
				exitChan <- struct{}{}
				return
			}
			log.Println("error:", err)

		case <-syncChan:
			slog.Debug("received sync message")
		}
	}
}
