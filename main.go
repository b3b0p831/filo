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
)

var Cfg *config.Config
var maxFileSemaphore chan struct{}
var wg sync.WaitGroup

func init() {
	Cfg = config.Load()
	maxFileSemaphore = make(chan struct{}, Cfg.MaxOpenFile)
}

func main() {

	util.PrintIntro(Cfg)

	slog.Debug("building initial FiloTrees...")
	srcTree, err := fs.BuildTree(Cfg.SourceDir)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	targetTree, err := fs.BuildTree(Cfg.TargetDir)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	rightNow := time.Now()
	slog.Debug(fmt.Sprintf("srcTree.Missingin(%v) ", targetTree.Root.Path))
	var missing map[string][]*fs.FileNode = srcTree.MissingIn(targetTree, maxFileSemaphore, func() {
		slog.Debug(fmt.Sprint("srcTree.Missingin(targetTree) Elapsed time: ", time.Since(rightNow)))
	})

	//Perform Initial Sync
	if len(missing) != 0 {
		slog.Info("Performing initial file sync...")
		rightNow = time.Now()
		targetTree.CopyFrom(srcTree, missing, maxFileSemaphore, func() {
			slog.Debug(fmt.Sprintln(missing))
			slog.Info(fmt.Sprint("Initial file sync complete, Elapsed time: ", time.Since(rightNow)))
		})

	}

	exitChan := make(chan struct{})

	eventChan := make(chan fsnotify.Event)
	defer close(eventChan)

	wg.Go(func() {
		fs.WatchChanges(eventChan, exitChan, Cfg)
	})
	wg.Go(func() {
		fs.SyncChanges(eventChan, exitChan, maxFileSemaphore, Cfg)
	})

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	// Block until the signal is received
	slog.Info("Press Ctrl+C to exit...")

	select {
	case <-exitChan:
	case <-ctx.Done():
		close(exitChan)
	}

	wg.Wait()

	slog.Info("Filo exiting...")
}
