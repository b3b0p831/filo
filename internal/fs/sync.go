package fs

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"bebop831.com/filo/internal/config"

	"github.com/fsnotify/fsnotify"
)

func syncRemove(filesToBeRemoved []string, src *FileTree, tgt *FileTree) {

	tgtRoot, err := os.OpenRoot(tgt.Root.Path)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer tgtRoot.Close()

	for _, fp := range filesToBeRemoved {

		relBaseFile, err := filepath.Rel(src.Root.Path, fp)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		if filepath.IsLocal(relBaseFile) {

			if err := tgtRoot.RemoveAll(relBaseFile); err != nil {
				slog.Warn(err.Error())
			}

			// tgtFilePath := filepath.Join(tgt.Root.Path, relBaseFile)
			// delete(src.Index, fp)
			// delete(tgt.Index, tgtFilePath)

			if _, err := tgtRoot.Lstat(relBaseFile); errors.Is(err, os.ErrNotExist) {
				slog.Info(fmt.Sprintf("%s successfully deleted from %s", relBaseFile, tgt.Root.Path))
			}

		}

	}
}

func syncCreate(filesToBeCreated []string, src *FileTree, tgt *FileTree) {
	tgtRoot, err := os.OpenRoot(tgt.Root.Path)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer tgtRoot.Close()

	for _, fp := range filesToBeCreated {
		slog.Info(fp)
	}

}

// Sync maintains 2 directories that should be the same.
func SyncChanges(eventChan <-chan fsnotify.Event, exit <-chan struct{}, syncChan chan struct{}, maxFileSemaphore chan struct{}, cfg *config.Config) {
	minInterval := cfg.SyncDelay

	var lastEvent time.Time
	var wg sync.WaitGroup
	fileEvents := make(map[string][]string) //path -> last operation(i.e CREATE or REMOVE)

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
				syncTime := time.Now()

				if syncChan != nil {
					syncChan <- struct{}{}
				}

				srcFileTree := BuildTree(cfg.SourceDir)
				dstFileTree := BuildTree(cfg.TargetDir)

				for op, paths := range fileEvents {
					switch op {
					case "REMOVE":
						// Delete the file where the event is Rename or Remove. Will treat same for now
						syncRemove(paths, srcFileTree, dstFileTree)

					case "RENAME":
						// Rename with no matching Create? Removed from watch dir, delete file/dir
						// Rename with matching Create? File was moved from current loc to somewhere else in srcDir
						// Check if any rename matches a create content. Pop entry from fileEvents?
						//wg.Go(func() { syncRemove(paths, srcFileTree, dstFileTree) })

					case "WRITE":
						// Compare fileState info between src -> target, if different get diff from src, if missing fallthrough to create
						//wg.Go(func() { syncRemove(paths, srcFileTree, dstFileTree) })

					case "CREATE":
						// If dir, create dir. If file create file.
						// Filo should only every write in the target dir and not outside.
						syncCreate(paths, srcFileTree, dstFileTree)

					}
				}

				wg.Wait()

				// reset
				lastEvent = time.Time{}
				fileEvents = make(map[string][]string)
				srcFileTree = BuildTree(cfg.SourceDir)
				dstFileTree = BuildTree(cfg.TargetDir)

				slog.Info(fmt.Sprintf("Sync completed successfully, Elapsed time: %v", time.Since(syncTime)))

			}

		case <-exit:
			return
		}
	}
}
