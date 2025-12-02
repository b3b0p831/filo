package util

import (
	"fmt"
	"slices"
	"time"

	"bebop831.com/filo/config"
	"github.com/fsnotify/fsnotify"
)

// Sync maintains 2 directories that should be the same. Does not do intial sync. That should happen before arriving here.
func SyncChanges(eventChan <-chan fsnotify.Event, exit <-chan struct{}, syncChan chan struct{}, cfg *config.Config) {
	minInterval, err := GetTimeInterval(cfg.SyncDelay)
	if err != nil {
		Flogger.Fatalf("failed to parse sync_delay in config: invalid value '%v' (must use s, m, or h).", cfg.SyncDelay)
	}

	var lastEvent time.Time
	fileEvents := make(map[string][]string)

	for {
		select {
		case e := <-eventChan:
			changesSlice := fileEvents[e.Op.String()] // i.e fileEvents["CREATE"]
			if !slices.Contains(changesSlice, e.Name) {
				fileEvents[e.Op.String()] = append(changesSlice, e.Name)
			}

			lastEvent = time.Now()

		case <-time.After(minInterval):
			if !lastEvent.IsZero() && time.Since(lastEvent) >= minInterval {

				srcFileTree := BuildTree(cfg.SourceDir)
				dstFileTree := BuildTree(cfg.TargetDir)

				fmt.Println(srcFileTree)
				fmt.Println(dstFileTree)
				fmt.Println(srcFileTree.MissingIn(dstFileTree, nil))

				// for Op, Paths := range fileEvents {
				// 	switch Op {
				// 	case "REMOVE":
				// Delete the file where the event is Rename or Remove. Will treat same for now

				//  case "RENAME":
				// Rename with no matching Create? Removed from watch dir, delete file/dir
				// Rename with matching Create? File was moved from current loc to somewhere else in srcDir	
				// Check if any rename matches a create content. Pop entry from fileEvents?

				// 	case "WRITE":
				// Compare fileState info between src -> target, if different get diff from src, if missing fallthrough to create

				// 	case "CREATE":
				// If dir, create dir. If file create file.
				// Filo will only every write in the target dir. 
				// 	}
				// }

				Flogger.Printf("Syncing started: %v -> %v...\n", cfg.SourceDir, cfg.TargetDir)
				if syncChan != nil {
					syncChan <- struct{}{}
				}

				Flogger.Printf("Sync completed successfully")

				lastEvent = time.Time{} // reset
				for k := range fileEvents {
					delete(fileEvents, k)
				}

			}

		case <-exit:
			return
		}
	}
}
