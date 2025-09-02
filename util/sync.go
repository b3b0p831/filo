package util

import (
	"fmt"
	"log"
	"time"

	"bebop831.com/filo/config"
)

func Sync(eventChan <-chan struct{}, exit <-chan struct{}, syncChan chan struct{}, cfg *config.Config) {
	minInterval, err := GetTimeInterval(cfg.SyncDelay)
	if err != nil {
		log.Fatalf("failed to parse sync_delay in config: invalid value '%v' (must use s, m, or h).", cfg.SyncDelay)
	}

	var lastEvent time.Time

	for {
		select {
		case <-eventChan:
			lastEvent = time.Now()

		case <-time.After(minInterval):
			if !lastEvent.IsZero() && time.Since(lastEvent) >= minInterval {

				fileTree := BuildTree(cfg.SourceDir)

				fmt.Println(fileTree)

				for _, v := range fileTree.Index {
					fmt.Println(v)
				}

				log.Printf("Syncing started: %v -> %v...\n", cfg.SourceDir, cfg.TargetDir)
				if syncChan != nil {
					syncChan <- struct{}{}
				}

				log.Printf("Sync completed successfully")
				lastEvent = time.Time{} // reset
			}

		case <-exit:
			return
		}
	}
}
