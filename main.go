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
var maxFileSemaphore chan struct{}

func init() {
	Cfg = config.Load()
	maxFileSemaphore = make(chan struct{}, Cfg.MaxOpenFile)
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

	srcTree, targetTree := fs.BuildTree(Cfg.SourceDir), fs.BuildTree(Cfg.TargetDir)

	rightNow := time.Now()
	var missing map[string][]*fs.FileNode = srcTree.MissingIn(targetTree, func() {
		slog.Debug(fmt.Sprint("srcTree.Missingin(targetTree) Elapsed time: ", time.Since(rightNow)))
	}, maxFileSemaphore)

	//Perform Initial Sync
	if len(missing) != 0 {
		slog.Info("Performing initial file sync...")
		rightNow = time.Now()
		slog.Debug(fmt.Sprintln(missing))
		targetTree.CopyMissing(missing, maxFileSemaphore, func() {
			slog.Info(fmt.Sprint("Initial file sync complete, Elapsed time: ", time.Since(rightNow)))
		})

	}

	eventChan := make(chan fsnotify.Event)
	exitChan := make(chan struct{})
	syncChan := make(chan struct{})

	slog.Info(fmt.Sprintf("Starting FILO watch on '%s'...", Cfg.SourceDir))
	go fs.SyncChanges(eventChan, exitChan, syncChan, maxFileSemaphore, Cfg)
	go fs.WatchChanges(eventChan, exitChan, syncChan, Cfg)

	<-make(chan struct{})

}
