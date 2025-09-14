package main

import (
	"log"

	"bebop831.com/filo/config"
	"bebop831.com/filo/util"
	"github.com/fsnotify/fsnotify"

	"github.com/shirou/gopsutil/v4/disk"
)

func main() {
	cfg, err := config.Load()

	if cfg.LogLevel == "debug" {
		log.Println("Starting FILO...")
	}

	util.PrintBanner()

	if err != nil {
		log.Println(err)
	}

	if len(cfg.SourceDir) == 0 {
		log.Fatalln("Invalid source dir.")
	}

	if len(cfg.TargetDir) == 0 {
		log.Fatalln("Invalid target dir.")
	}

	targetUsage, err := disk.Usage(cfg.TargetDir)
	if err != nil {
		log.Fatalln(err)
	}

	srcUsage, err := disk.Usage(cfg.SourceDir)
	if err != nil {
		log.Fatalln(err)
	}

	util.PrintConfig(cfg, srcUsage, targetUsage)
	log.Printf("Starting FILO watch on '%s'...\n", cfg.SourceDir)

	srcTree := util.BuildTree(cfg.SourceDir)
	targetTree := util.BuildTree(cfg.TargetDir)

	var missing map[string][]*util.FileNode = srcTree.GetMissing(targetTree)

	for k, v := range missing {
		log.Println(k)
		for _, e := range v {
			log.Println("\t", e.Entry.Name())
		}
	}

	targetTree.CopyMissing(missing)

	watcher, err := fsnotify.NewWatcher()
	watcher.Add(cfg.SourceDir)
	go util.OnCreate(&fsnotify.Event{Op: fsnotify.Create, Name: cfg.SourceDir}, watcher)

	if err != nil {
		log.Fatalln(err)
	}

	eventChan := make(chan fsnotify.Event)
	exitChan := make(chan struct{})
	syncChan := make(chan struct{})

	go util.SyncChanges(eventChan, exitChan, syncChan, cfg)

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
				log.Println(event.Op, event.Name)
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
