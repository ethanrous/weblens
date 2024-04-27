package dataStore

import (
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/fsnotify/fsnotify"
)

type watcherPathMod struct {
	path string
	add  bool
}

var pathModChan = make(chan (watcherPathMod), 5)

func fileWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	watcher.Add(mediaRoot.absolutePath)

	var holder *fsnotify.Event
	holdTimer := time.NewTimer(time.Second)
	holdTimer.Stop()

WatcherLoop:
	for {
		select {
		case <-holdTimer.C:
			jeStream <- fileEvent{action: FileCreate, postFilePath: holder.Name}
			holder = nil
		case event, ok := <-watcher.Events:
			if !ok {
				break WatcherLoop
			}

			// util.Debug.Println("Got file event", event.Name)
			if event.Has(fsnotify.Create) {
				// Move events show up as a distinct "Create" in the destination
				// followed by a "Rename" in the old location, so we hold on to
				// create actions for 100 ms to wait for the following rename.

				if holder == nil {
					holder = &event
					holdTimer = time.NewTimer(time.Millisecond * 100)
					continue
				}

				// If we are already holding onto a create event, then
				// it must have been a real create event, as it was not followed
				// by a rename. So we rinse and repeat
				holdTimer.Stop()
				jeStream <- fileEvent{action: FileCreate, postFilePath: holder.Name}
				holder = &event
				holdTimer = time.NewTimer(time.Millisecond * 100)
				continue
			}

			if event.Has(fsnotify.Remove) {
				jeStream <- fileEvent{action: FileDelete, preFilePath: event.Name}
				continue
			}

			if event.Has(fsnotify.Rename) {
				if holder == nil {
					jeStream <- fileEvent{action: FileDelete, preFilePath: event.Name}
				} else {
					holdTimer.Stop()
					jeStream <- fileEvent{action: FileMove, preFilePath: event.Name, postFilePath: holder.Name}
					holder = nil
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				break WatcherLoop
			}
			util.ShowErr(err, "File watcher error")
		case mod, ok := <-pathModChan:
			if !ok {
				break WatcherLoop
			}

			if mod.add {
				watcher.Add(mod.path)
			} else {
				watcher.Remove(mod.path)
			}
		}
	}

	// Not reached
	util.Error.Panicln("File watcher exiting...")
}

func watcherAddDirectory(f types.WeblensFile) error {
	if !f.IsDir() {
		return ErrDirectoryRequired
	}
	if f.Owner() == WEBLENS_ROOT_USER {
		return nil
	}

	realF := f.(*weblensFile)
	if realF.watching {
		return ErrAlreadyWatching
	}

	newMod := watcherPathMod{path: f.GetAbsPath(), add: true}
	realF.watching = true
	pathModChan <- newMod

	return nil
}
