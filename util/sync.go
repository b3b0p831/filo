package util

import (
	"log"
	"time"

	"bebop831.com/filo/config"
)

func Sync(eventChan <-chan struct{}, exit <-chan struct{}, syncChan chan struct{}, cfg *config.Config) {
	minInterval, err := GetTimeInterval(cfg.SyncDelay)
	if err != nil {
		log.Fatalf("failed to parse sync_delay in config: invalid value '%v' (must use s, m, or h).", cfg.SyncDelay)
	}
	var timer *time.Timer = time.NewTimer(minInterval)
	var eventOccurred bool = false

	timer = time.AfterFunc(minInterval, func() {
		if eventOccurred {
			log.Println("Preparing to sync...")
			time.Sleep(2 * time.Second)
			log.Printf("Syncing started: %v -> %v...\n", cfg.SourceDir, cfg.TargetDir)
			syncChan <- struct{}{}
			time.Sleep(2 * time.Second)
			log.Printf("Sync completed successfully\n")

			eventOccurred = false //Sync Done or Failed. Reset to original state.
		}
	})

	for {
		select {
		case <-eventChan:
			eventOccurred = true
			timer.Reset(minInterval)

		case <-exit:
			timer.Stop()
			return
		}
	}

}
