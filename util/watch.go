package util

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"bebop831.com/filo/config"
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

func WatchChanges(eventChan chan fsnotify.Event, exitChan chan struct{}, cfg *config.Config) {
	// Turns this into go routine so that it matches above(i.e WatchChanges)

	watcher, err := fsnotify.NewWatcher()
	watcher.Add(cfg.SourceDir)
	go OnCreate(&fsnotify.Event{Op: fsnotify.Create, Name: cfg.SourceDir}, watcher)

	if err != nil {
		log.Fatalln(err)
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
				log.Println(CreateColor(event.Op), event.Name)
				go OnCreate(&event, watcher)

			case fsnotify.Rename:
				log.Println(RenameColor(event.Op), event.Name)

			case fsnotify.Remove:
				log.Println(RemoveColor(event.Op), event.Name)

			case fsnotify.Chmod:
				continue

			default:
				log.Println(event.Op, event.Name)
			}

			eventChan <- event

		case err, ok := <-watcher.Errors:
			if !ok {
				exitChan <- struct{}{}
				return
			}
			log.Println("error:", err)

		case <-exitChan:
		}
	}
}
