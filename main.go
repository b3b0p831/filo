package main

import (
	"log"
	"time"

	"bebop831.com/filo/util"
	"github.com/fsnotify/fsnotify"

	"github.com/shirou/gopsutil/v4/disk"
)

func main() {
	if util.Cfg.LogLevel == "debug" {
		log.Println("Starting FILO...")
	}

	util.PrintBanner()

	if len(util.Cfg.SourceDir) == 0 {
		log.Fatalln("Invalid source dir.")
	}

	if len(util.Cfg.TargetDir) == 0 {
		log.Fatalln("Invalid target dir.")
	}

	targetUsage, err := disk.Usage(util.Cfg.TargetDir)
	if err != nil {
		log.Fatalln(err)
	}

	srcUsage, err := disk.Usage(util.Cfg.SourceDir)
	if err != nil {
		log.Fatalln(err)
	}

	util.PrintConfig(util.Cfg, srcUsage, targetUsage)
	log.Printf("Starting FILO watch on '%s'...\n", util.Cfg.SourceDir)

	srcTree := util.BuildTree(util.Cfg.SourceDir)
	targetTree := util.BuildTree(util.Cfg.TargetDir)

	rn := time.Now()
	var missing map[string][]*util.FileNode = srcTree.MissingIn(targetTree, func() {
		log.Println("Elapsed:", time.Since(rn))
	})

	log.Println(missing)
	targetTree.CopyMissing(missing)

	watcher, err := fsnotify.NewWatcher()
	watcher.Add(util.Cfg.SourceDir)
	go util.OnCreate(&fsnotify.Event{Op: fsnotify.Create, Name: util.Cfg.SourceDir}, watcher)

	if err != nil {
		log.Fatalln(err)
	}

	eventChan := make(chan fsnotify.Event)
	exitChan := make(chan struct{})
	syncChan := make(chan struct{})

	go util.SyncChanges(eventChan, exitChan, syncChan, util.Cfg)

	// TODO: Turns this into go routine, go util.WatchChanges
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			switch event.Op {
			case fsnotify.Create:
				log.Println(util.CreateColor(event.Op), event.Name)
				go util.OnCreate(&event, watcher)

			case fsnotify.Rename:
				log.Println(util.RenameColor(event.Op), event.Name)

			case fsnotify.Remove:
				log.Println(util.RemoveColor(event.Op), event.Name)

			case fsnotify.Chmod:
				continue

			default:
				if util.Cfg.LogLevel == "debug" {
					log.Println(event.Op, event.Name)
				}
			}

			eventChan <- event

		case err, ok := <-watcher.Errors:
			if !ok {
				exitChan <- struct{}{}
				return
			}
			log.Println("error:", err)

		case <-syncChan:
		}
	}

}
