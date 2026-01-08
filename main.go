package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"

	"bebop831.com/filo/internal/config"
	"bebop831.com/filo/internal/fs"
	"bebop831.com/filo/internal/util"

	"github.com/fsnotify/fsnotify"
	"github.com/shirou/gopsutil/v4/disk"
)

var Cfg *config.Config
var maxFileSemaphore chan struct{}
var wg sync.WaitGroup

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

	srcTree := fs.BuildTree(Cfg.SourceDir)
	//if Cfg.WatchOnlyInSrc BuildTargetTree else BuildTree
	targetTree := fs.BuildTree(Cfg.TargetDir)

	rightNow := time.Now()
	var missing map[string][]*fs.FileNode = srcTree.MissingIn(targetTree, maxFileSemaphore, func() {
		slog.Debug(fmt.Sprint("srcTree.Missingin(targetTree) Elapsed time: ", time.Since(rightNow)))
	})

	//Perform Initial Sync
	if len(missing) != 0 {
		slog.Info("Performing initial file sync...")
		slog.Info(fmt.Sprint(missing))
		rightNow = time.Now()
		slog.Debug(fmt.Sprintln(missing))
		targetTree.CopyFrom(srcTree, missing, maxFileSemaphore, func() {
			slog.Info(fmt.Sprint("Initial file sync complete, Elapsed time: ", time.Since(rightNow)))
		})

	}
	exitChan := make(chan struct{})

	eventChan := make(chan fsnotify.Event)
	defer close(eventChan)

	syncChan := make(chan struct{})
	defer close(syncChan)

	slog.Info(fmt.Sprintf("Starting FILO watch on '%s'...", Cfg.SourceDir))

	wg.Go(func() {
		fs.WatchChanges(eventChan, exitChan, syncChan, Cfg)
	})
	wg.Go(func() {
		fs.SyncChanges(eventChan, exitChan, syncChan, maxFileSemaphore, Cfg)
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	slog.Info("Press Ctrl+C to exit...")

	// Block until the signal is received
	<-ctx.Done()
	slog.Info("\nCleanly shutting down...")
	close(exitChan)

	wg.Wait()
	slog.Info("Filo exiting...")
}
