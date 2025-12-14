package main

import (
	"fmt"
	"log/slog"
	"time"

	"bebop831.com/filo/internal/config"
	"bebop831.com/filo/internal/fs"
	"bebop831.com/filo/internal/util"

	"github.com/fsnotify/fsnotify"
	"github.com/shirou/gopsutil/v4/disk"
)

var Cfg *config.Config

func init() {
	Cfg = config.Load()
}

func main() {

	util.PrintBanner()

	if len(Cfg.SourceDir) == 0 {
		slog.Error("Invalid source dir.")
	}

	if len(Cfg.TargetDir) == 0 {
		slog.Error("Invalid target dir.")
	}

	targetUsage, err := disk.Usage(Cfg.TargetDir)
	if err != nil {
		slog.Error(err.Error() + " " + Cfg.TargetDir)
		return
	}

	srcUsage, err := disk.Usage(Cfg.SourceDir)
	if err != nil {
		slog.Error(err.Error() + " " + Cfg.SourceDir)
		return
	}

	util.PrintConfig(Cfg, srcUsage, targetUsage)
	slog.Info(fmt.Sprintf("Starting FILO watch on '%s'...", Cfg.SourceDir))

	srcTree, targetTree := fs.BuildTree(Cfg.SourceDir), fs.BuildTree(Cfg.TargetDir)

	rightNow := time.Now()
	var missing map[string][]*fs.FileNode = srcTree.MissingIn(targetTree, func() {
		slog.Info(fmt.Sprint("Elapsed time: ", time.Since(rightNow)))
	})

	slog.Debug(fmt.Sprintln(missing))
	targetTree.CopyMissing(missing)

	watcher, err := fsnotify.NewWatcher()
	watcher.Add(Cfg.SourceDir)
	go fs.OnCreate(&fsnotify.Event{Op: fsnotify.Create, Name: Cfg.SourceDir}, watcher)

	if err != nil {
		slog.Error(err.Error())
		return
	}

	eventChan := make(chan fsnotify.Event)
	exitChan := make(chan struct{})
	syncChan := make(chan struct{})

	go fs.SyncChanges(eventChan, exitChan, syncChan, Cfg)
	go fs.WatchChanges(eventChan, exitChan, syncChan, Cfg)

	<-make(chan struct{})

}
