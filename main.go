package main

import (
	"log"
	"sync"
	"time"

	"bebop831.com/filo/internal/config"
	"bebop831.com/filo/internal/fs"
	"bebop831.com/filo/internal/util"

	"github.com/fsnotify/fsnotify"
	"github.com/shirou/gopsutil/v4/disk"
)

var Cfg *config.Config = config.Load()
var Mu *sync.Mutex

func init() {
	Mu = &sync.Mutex{}
}

func main() {

	if Cfg.LogLevel == "debug" {
		Cfg.Flogger.Println("Starting FILO...")
	}

	util.PrintBanner()

	if len(Cfg.SourceDir) == 0 {
		Cfg.Flogger.Fatalln("Invalid source dir.")
	}

	if len(Cfg.TargetDir) == 0 {
		Cfg.Flogger.Fatalln("Invalid target dir.")
	}

	targetUsage, err := disk.Usage(Cfg.TargetDir)
	if err != nil {
		log.Fatalln(err, Cfg.TargetDir)
	}

	srcUsage, err := disk.Usage(Cfg.SourceDir)
	if err != nil {
		log.Fatalln(err, Cfg.SourceDir)
	}

	util.PrintConfig(Cfg, srcUsage, targetUsage)
	log.Printf("Starting FILO watch on '%s'...\n", Cfg.SourceDir)

	srcTree := fs.BuildTree(Cfg.SourceDir)
	targetTree := fs.BuildTree(Cfg.TargetDir)

	rn := time.Now()
	var missing map[string][]*fs.FileNode = srcTree.MissingIn(targetTree, func() {
		log.Println("Elapsed:", time.Since(rn))
	})

	Cfg.Flogger.Println(missing)
	targetTree.CopyMissing(missing)

	watcher, err := fsnotify.NewWatcher()
	watcher.Add(Cfg.SourceDir)
	go fs.OnCreate(&fsnotify.Event{Op: fsnotify.Create, Name: Cfg.SourceDir}, watcher, Cfg.Flogger)

	if err != nil {
		log.Fatalln(err)
	}

	eventChan := make(chan fsnotify.Event)
	exitChan := make(chan struct{})
	syncChan := make(chan struct{})

	go fs.SyncChanges(eventChan, exitChan, syncChan, Cfg)

	// TODO: Turns this into go routine, go util.WatchChanges
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if !fs.IsApprovedPath(event.Name) {
				continue
			}

			switch event.Op {
			case fsnotify.Create:
				log.Println(util.CreateColor(event.Op), event.Name)
				go fs.OnCreate(&event, watcher, Cfg.Flogger)

			case fsnotify.Rename:
				log.Println(util.RenameColor(event.Op), event.Name)

			case fsnotify.Remove:
				log.Println(util.RemoveColor(event.Op), event.Name)

			case fsnotify.Chmod:
				continue

			default:
				if Cfg.LogLevel == "debug" {
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
