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
	srcTree := fs.BuildTree(Cfg.SourceDir)
	targetTree := fs.BuildTree(Cfg.TargetDir)

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

	for _, v := range srcTree.Index {
		slog.Info(fmt.Sprintf("%s %x\n", v.Path, v.Hash))
	}

	for _, v := range targetTree.Index {
		slog.Info(fmt.Sprintf("%s %x\n", v.Path, v.Hash))
	}

	exitChan := make(chan struct{})

	eventChan := make(chan fsnotify.Event)
	defer close(eventChan)

	syncChan := make(chan struct{})
	defer close(syncChan)

	wg.Go(func() {
		fs.WatchChanges(eventChan, exitChan, syncChan, Cfg)
	})
	wg.Go(func() {
		fs.SyncChanges(eventChan, exitChan, syncChan, maxFileSemaphore, Cfg)
	})

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	// Block until the signal is received
	slog.Info("Press Ctrl+C to exit...")
	<-ctx.Done()

	//Perform Cleanup
	close(exitChan)

	//Wait for routines
	wg.Wait()

	slog.Info("Filo exiting...")
}
