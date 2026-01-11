package fs

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"bebop831.com/filo/internal/config"

	"github.com/fsnotify/fsnotify"
)

func syncRemove(filesRemoved []string, src *FileTree, tgt *FileTree) {

	tgtRoot, err := os.OpenRoot(tgt.Root.Path)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer tgtRoot.Close()

	for _, fileRemoved := range filesRemoved {
		relBaseFile := src.RelBaseFile(fileRemoved)
		tgtFilePath := filepath.Join(tgt.Root.Path, relBaseFile)
		_, ok := tgt.Index[tgtFilePath]
		if !ok {
			slog.Error(fmt.Sprintf("removed filepath '%s' missing from tgt.Index, skipping", tgtFilePath))
			continue
		}

		if filepath.IsLocal(relBaseFile) {

			if err := tgtRoot.RemoveAll(relBaseFile); err != nil {
				slog.Error(err.Error())
				continue
			}

			if _, err := tgtRoot.Lstat(relBaseFile); errors.Is(err, os.ErrNotExist) {
				slog.Info(fmt.Sprintf("%s successfully deleted from %s", relBaseFile, tgt.Root.Path))
			} else {
				slog.Error(err.Error())
			}

		} else {
			slog.Info(fmt.Sprintf("failed to delete %s", filepath.Join(tgt.Root.Path, relBaseFile)))
		}
	}

}

func parseFSEvents(fsEvents map[string]string) map[string][]string {
	eventMap := make(map[string][]string)
	for filePath, fsAction := range fsEvents {
		if _, ok := eventMap[fsAction]; !ok {
			eventMap[fsAction] = []string{}
		}
		currentFilePaths := eventMap[fsAction]
		currentFilePaths = append(currentFilePaths, filePath)
		eventMap[fsAction] = currentFilePaths
	}

	return eventMap
}

// Sync maintains 2 directories that should be the same.
func SyncChanges(eventChan <-chan fsnotify.Event, exit <-chan struct{}, syncChan chan struct{}, maxFileSemaphore chan struct{}, cfg *config.Config) {
	minInterval := cfg.SyncDelay

	var lastEvent time.Time
	var wg sync.WaitGroup
	lastFSEvents := make(map[string]string)

	for {
		select {
		case e := <-eventChan:
			fsAction, filePath := e.Op.String(), e.Name //CREATE, REMOVE, WRITE, etc | filePath (i.e /tmp/tempfile1.txt)
			lastFSEvents[filePath] = fsAction

			lastEvent = time.Now()

		case <-time.After(minInterval):
			if !lastEvent.IsZero() && time.Since(lastEvent) >= minInterval {

				slog.Info(fmt.Sprintf("Syncing started: %v -> %v...", cfg.SourceDir, cfg.TargetDir))
				syncTime := time.Now()

				if syncChan != nil {
					syncChan <- struct{}{}
				}

				//Build Tree
				srcFileTree := BuildTree(cfg.SourceDir)
				targetFileTree := BuildTree(cfg.TargetDir)

				eventMap := parseFSEvents(lastFSEvents)
				for fsAction, filePaths := range eventMap {
					switch fsAction {
					case "REMOVE":
						// Delete the file where the event is Rename or Remove. Will treat same for now
						wg.Go(func() { syncRemove(filePaths, srcFileTree, targetFileTree) })

					case "RENAME":
						// Rename with no matching Create? Removed from watch dir, delete file/dir
						// Rename with matching Create? File was moved from current loc to somewhere else in srcDir
						// Check if any rename matches a create content. Pop entry from fileEvents?
						// wg.Go(func() { syncRemove(paths, srcFileTree, dstFileTree) })

					case "WRITE", "CREATE":
						// Check if exists. If exists, compare fileState info between src -> target
						// wg.Go(func() { syncRemove(paths, srcFileTree, dstFileTree) })
						// If dir, create dir. If file create file.
						// Filo should only every write in the target dir and not outside.
						//wg.Go(func() { syncCreate(srcFileTree, targetFileTree, filePaths, maxFileSemaphore, &wg) })

						wg.Go(func() {
							missing := srcFileTree.MissingIn(targetFileTree, maxFileSemaphore, nil)
							if len(missing) > 0 {
								targetFileTree.CopyFrom(srcFileTree, missing, maxFileSemaphore, nil)
							}
						})

					default:
						slog.Debug(fmt.Sprintf("Skipping file event: %s %v", fsAction, filePaths))

					}
				}

				wg.Wait()

				// reset
				lastEvent = time.Time{}
				lastFSEvents = make(map[string]string)
				slog.Info(fmt.Sprintf("Sync completed successfully, Elapsed time: %v", time.Since(syncTime)))

			}

		case <-exit:
			slog.Debug("Exiting SyncChanges goroutine...")
			return
		}
	}
}
