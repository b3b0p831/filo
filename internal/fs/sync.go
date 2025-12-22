package fs

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"bebop831.com/filo/internal/config"

	"github.com/fsnotify/fsnotify"
)

func syncRemove(filePaths []string) {
	for _, p := range filePaths {
		slog.Info(p)
	}
}

// Sync maintains 2 directories that should be the same.
func SyncChanges(eventChan <-chan fsnotify.Event, exit <-chan struct{}, syncChan chan struct{}, maxFileSemaphore chan struct{}, cfg *config.Config) {
	minInterval := cfg.SyncDelay

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

				slog.Info(fmt.Sprintf("Syncing started: %v -> %v...", cfg.SourceDir, cfg.TargetDir))
				if syncChan != nil {
					syncChan <- struct{}{}
				}

				srcFileTree := BuildTree(cfg.SourceDir)
				dstFileTree := BuildTree(cfg.TargetDir)
				missingFiles := srcFileTree.MissingIn(dstFileTree, nil, maxFileSemaphore)

				slog.Debug(fmt.Sprint(srcFileTree))
				slog.Debug(fmt.Sprint(dstFileTree))
				slog.Debug(fmt.Sprint(missingFiles))

				for Op, Paths := range fileEvents {
					switch Op {
					case "REMOVE":
						// Delete the file where the event is Rename or Remove. Will treat same for now
						syncRemove(Paths)

					case "RENAME":
						// Rename with no matching Create? Removed from watch dir, delete file/dir
						// Rename with matching Create? File was moved from current loc to somewhere else in srcDir
						// Check if any rename matches a create content. Pop entry from fileEvents?

					case "WRITE":
						// Compare fileState info between src -> target, if different get diff from src, if missing fallthrough to create

					case "CREATE":
						// If dir, create dir. If file create file.
						// Filo should only every write in the target dir and not outside.
					}
				}

				slog.Info("Sync completed successfully")
				lastEvent = time.Time{} // reset

			}

		case <-exit:
			return
		}
	}
}
