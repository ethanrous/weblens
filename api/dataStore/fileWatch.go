package dataStore

import (
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/fsnotify/fsnotify"
)

type watcherPathMod struct {
	path string
	add  bool
}

var pathModChan = make(chan (watcherPathMod), 20)

func fileWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	watcher.Add(mediaRoot.absolutePath)

WatcherLoop:
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				break WatcherLoop
			}

			if event.Has(fsnotify.Create) {
				util.Debug.Println("Created", event.Name)
			}

			if event.Has(fsnotify.Remove) {
				util.Debug.Println("Removed", event.Name)
			}

			if event.Has(fsnotify.Rename) {
				util.Debug.Println("Renamed", event.Name)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				break WatcherLoop
			}
			util.Debug.Println(err, ok)
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
	realF := f.(*weblensFile)
	if realF.watching {
		return ErrAlreadyWatching
	}

	newMod := watcherPathMod{path: f.GetAbsPath(), add: true}
	realF.watching = true
	pathModChan <- newMod

	return nil
}
