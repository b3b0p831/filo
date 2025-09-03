package util

import (
	"fmt"
	"log"
	"time"

	"bebop831.com/filo/config"
	"github.com/fsnotify/fsnotify"
)

func Sync(eventChan <-chan fsnotify.Event, exit <-chan struct{}, syncChan chan struct{}, cfg *config.Config) {
	minInterval, err := GetTimeInterval(cfg.SyncDelay)
	if err != nil {
		log.Fatalf("failed to parse sync_delay in config: invalid value '%v' (must use s, m, or h).", cfg.SyncDelay)
	}

	var lastEvent time.Time
	lastChange := make(map[string][]string)

	for {
		select {
		case e := <-eventChan:
			changesSlice := lastChange[e.Op.String()]
			lastChange[e.Op.String()] = append(changesSlice, e.Name)
			lastEvent = time.Now()

		case <-time.After(minInterval):
			if !lastEvent.IsZero() && time.Since(lastEvent) >= minInterval {

				fileTree := BuildTree(cfg.SourceDir)

				fmt.Println(fileTree)

				for _, v := range fileTree.Index {
					fmt.Println(v)
				}

				fmt.Println(lastChange)

				log.Printf("Syncing started: %v -> %v...\n", cfg.SourceDir, cfg.TargetDir)
				if syncChan != nil {
					syncChan <- struct{}{}
				}

				log.Printf("Sync completed successfully")

				lastEvent = time.Time{} // reset
				for k := range lastChange {
					delete(lastChange, k)
				}

			}

		case <-exit:
			return
		}
	}
}
