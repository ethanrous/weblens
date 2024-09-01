package fileTree

import (
	"time"

	"github.com/ethrousseau/weblens/internal/log"
	"github.com/fsnotify/fsnotify"
)

type watcherPathMod struct {
	path string
	add  bool
}

var pathModChan = make(chan (watcherPathMod), 5)

func (j *JournalServiceImpl) FileWatcher() {
	_, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	// err = watcher.Add(j.fileTree.GetRoot().GetAbsPath())
	// if err != nil {
	// 	panic(err)
	// }

	// var holder *fsnotify.Event
	holdTimer := time.NewTimer(time.Second)
	holdTimer.Stop()

WatcherLoop:
	for {
		select {
		// case <-holdTimer.C:
		// 	jeStream <- FileEvent{action: FileCreate, postFilePath: holder.Name}
		// 	holder = nil
		// case event, ok := <-watcher.Events:
		// 	if !ok {
		// 		break WatcherLoop
		// 	}
		//
		// 	// util.Debug.Println("Got file event", event.Name)
		// 	if event.Has(fsnotify.Create) {
		// 		// Move events show up as a distinct "Create" in the destination
		// 		// followed by a "Rename" in the old location, so we hold on to
		// 		// create Actions for 100 ms to wait for the following rename.
		//
		// 		if holder == nil {
		// 			holder = &event
		// 			holdTimer = time.NewTimer(time.Millisecond * 100)
		// 			continue
		// 		}
		//
		// 		// If we are already holding onto a create event, then
		// 		// it must have been a real create event, as it was not followed
		// 		// by a rename. So we rinse and repeat
		// 		holdTimer.Stop()
		// 		jeStream <- FileEvent{action: FileCreate, postFilePath: holder.Name}
		// 		holder = &event
		// 		holdTimer = time.NewTimer(time.Millisecond * 100)
		// 		continue
		// 	}
		//
		// 	if event.Has(fsnotify.Remove) {
		// 		jeStream <- FileEvent{action: FileDelete, preFilePath: event.Name}
		// 		continue
		// 	}
		//
		// 	if event.Has(fsnotify.Rename) {
		// 		if holder == nil {
		// 			jeStream <- FileEvent{action: FileDelete, preFilePath: event.Name}
		// 		} else {
		// 			holdTimer.Stop()
		// 			jeStream <- FileEvent{action: FileMove, preFilePath: event.Name, postFilePath: holder.Name}
		// 			holder = nil
		// 		}
		// 	}
		//
		// case err, ok := <-watcher.Errors:
		// 	if !ok {
		// 		break WatcherLoop
		// 	}
		// 	util.ShowErr(err, "File watcher error")
		case _, ok := <-pathModChan:
			if !ok {
				break WatcherLoop
			}

			// if mod.add {
			// 	watcher.Add(mod.path)
			// } else {
			// 	watcher.Remove(mod.path)
			// }
		}
	}

	// Not reached
	log.Error.Panicln("File watcher exiting...")
}

func (j *JournalServiceImpl) WatchFolder(f *WeblensFileImpl) error {
	// if !f.IsDir() {
	// 	return dataStore.ErrDirectoryRequired
	// }
	// if f.Owner() == types.SERV.UserService.Get("WEBLENS") {
	// 	return nil
	// }

	// err := f.SetWatching()
	// if err != nil {
	// 	return err
	// }

	// newMod := watcherPathMod{path: f.GetAbsPath(), add: true}
	// pathModChan <- newMod

	return nil
}
