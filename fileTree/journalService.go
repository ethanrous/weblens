package fileTree

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
	"go.mongodb.org/mongo-driver/mongo"
)

type JournalServiceImpl struct {
	lifetimes       map[LifetimeId]*Lifetime
	lifetimeMapLock sync.RWMutex

	latestUpdate  map[FileId]*Lifetime
	latestMapLock sync.Mutex

	collection *mongo.Collection
}

func NewJournalService(col *mongo.Collection) JournalService {
	return &JournalServiceImpl{
		lifetimes:    make(map[LifetimeId]*Lifetime),
		latestUpdate: make(map[FileId]*Lifetime),
		collection:   col,
	}
}

func (j *JournalServiceImpl) Init(store types.HistoryStore) error {
	j.store = store

	var lifetimes []*Lifetime
	// var updatedLifetimes []*Lifetime
	var err error
	var hasProxy bool
	var proxyStore types.ProxyStore

	sw := internal.NewStopwatch("Journal init")

	if proxyStore, hasProxy = store.(types.ProxyStore); hasProxy {
		// Get all lifetimes from the local database
		localLifetimes, err := proxyStore.GetLocalStore().GetAllLifetimes()
		if err != nil {
			return err
		}

		sw.Lap("Read all local lifetimes")

		j.lifetimeMapLock.Lock()
		j.latestMapLock.Lock()
		for _, l := range localLifetimes {
			j.lifetimes[l.ID()] = l
			j.latestUpdate[l.GetLatestFileId()] = l
		}
		j.latestMapLock.Unlock()
		j.lifetimeMapLock.Unlock()

		sw.Lap("Add all local lifetimes")

		latest, err := proxyStore.GetLatestAction()
		if err != nil {
			return err
		}

		var latestTimestamp time.Time
		if latest != nil {
			latestTimestamp = latest.GetTimestamp()
		}

		remoteLifetimes, err := types.SERV.StoreService.GetLifetimesSince(latestTimestamp)
		if err != nil {
			return err
		}

		sw.Lap("Proxy got lifetime updates")

		// Upsert lifetimes that have been updated on remote server
		for _, l := range remoteLifetimes {
			err = j.Add(l)
			if err != nil {
				return err
			}
		}

		sw.Lap("Proxy upsert lifetime updates")
	} else {
		lifetimes, err = store.GetAllLifetimes()
		if err != nil {
			return err
		}
		sw.Lap("Get all local lifetimes")

		// slices.SortFunc(
		// 	lifetimes, func(a, b *Lifetime) int {
		// 		aActions := a.GetActions()
		// 		bActions := b.GetActions()
		// 		return len(aActions[len(aActions)-1].GetDestinationPath()) - len(bActions[len(bActions)-1].GetDestinationPath())
		// 	},
		// )
		//
		// sw.Lap("Sorted all lifetimes")

		j.lifetimeMapLock.Lock()
		j.latestMapLock.Lock()
		for _, l := range lifetimes {
			j.lifetimes[l.ID()] = l
			j.latestUpdate[l.GetLatestFileId()] = l
		}
		j.latestMapLock.Unlock()
		j.lifetimeMapLock.Unlock()

		sw.Lap("Added all lifetimes")
	}

	sw.Stop()
	sw.PrintResults(false)

	return nil
}

func (j *JournalServiceImpl) GetActiveLifetimes() []*Lifetime {
	var result []*Lifetime
	for _, l := range j.lifetimes {
		if l.IsLive() {
			result = append(result, l)
		}
	}
	return result
}

func (j *JournalServiceImpl) GetAllLifetimes() []*Lifetime {
	j.lifetimeMapLock.RLock()
	defer j.lifetimeMapLock.RUnlock()
	return internal.MapToValues(j.lifetimes)
}

func (j *JournalServiceImpl) GetAllFileEvents() ([]*FileEvent, error) {
	// return util.MapToSlicePure(j.)
	// return types.SERV.StoreService.GetAllFileEvents()
	return nil, werror.NotImplemented("get all file events")
}

func (j *JournalServiceImpl) LogEvent(fe *FileEvent) error {
	if len(fe.GetActions()) == 0 {
		return nil
	}

	actions := fe.GetActions()
	slices.SortFunc(
		actions, func(a, b *FileAction) int {
			return a.GetTimestamp().Compare(b.GetTimestamp())
		},
	)

	var updated []*Lifetime

	j.latestMapLock.Lock()
	defer j.latestMapLock.Unlock()
	j.lifetimeMapLock.Lock()
	defer j.lifetimeMapLock.Unlock()

	for _, action := range fe.GetActions() {
		if action.GetFile() != nil {
			size, err := action.GetFile().Size()
			if err != nil {
				return werror.Wrap(err)
			}
			action.SetSize(size)
		}

		switch action.GetActionType() {
		case FileCreate:
			newL, err := NewLifetime("", action)
			if err != nil {
				return err
			}

			if newL == nil {
				return werror.WErrMsg("failed to create new lifetime")
			}

			f := types.SERV.FileTree.Get(action.GetDestinationId())
			sz, _ := f.Size()
			if !f.IsDir() && sz != 0 {
				contentId := f.GetContentId()
				if contentId == "" {
					return werror.NewWeblensError(
						fmt.Sprintln(
							"No content ID in FileCreate lifetime update", action.GetDestinationPath(),
						),
					)
				}
				newL.SetContentId(contentId)
			}

			j.lifetimes[newL.ID()] = newL
			j.latestUpdate[action.GetDestinationId()] = newL
			updated = append(updated, newL)
		case FileMove:
			existing := j.latestUpdate[action.GetOriginId()]
			delete(j.latestUpdate, existing.GetLatestFileId())
			existing.Add(action)
			j.latestUpdate[existing.GetLatestFileId()] = existing

			updated = append(updated, existing)
		case FileDelete:
			existing := j.latestUpdate[action.GetOriginId()]
			delete(j.latestUpdate, existing.GetLatestFileId())
			existing.Add(action)

			updated = append(updated, existing)
		}
	}

	for _, update := range updated {
		f := types.SERV.FileTree.Get(update.GetLatestFileId())
		if f != nil {
			sz, _ := f.Size()
			if update.GetContentId() == "" && !f.IsDir() && sz != 0 {
				return werror.WErrMsg("No content ID in lifetime update")
			}
		} else if update.GetLatestAction().GetActionType() != FileDelete {
			return werror.WErrMsg("Could not find file for non-delete lifetime update")
		}
		err := types.SERV.StoreService.UpsertLifetime(update)
		if err != nil {
			return err
		}
	}

	return nil
}

func (j *JournalServiceImpl) GetActionsByPath(path *WeblensFilepath) ([]*FileAction, error) {
	return j.store.GetActionsByPath(path)
}

func (j *JournalServiceImpl) GetPastFolderInfo(folder *WeblensFile, time time.Time) (
	[]weblens.FileInfo, error,
) {
	actions, err := types.SERV.StoreService.GetActionsByPath(folder.GetPortablePath())
	if err != nil {
		return nil, werror.Wrap(err)
	}

	actionsMap := map[string]*FileAction{}
	for _, action := range actions {
		if action.GetTimestamp().After(time) {
			continue
		}
		var path string
		switch action.GetActionType() {
		case FileCreate:
			path = action.GetDestinationPath()
		case FileMove:
			path = action.GetOriginPath()
		case FileDelete:
			path = action.GetOriginPath()
		}
		if existingAction, ok := actionsMap[path]; ok {
			if action.GetTimestamp().After(existingAction.GetTimestamp()) {
				actionsMap[path] = action
			}
		} else {
			actionsMap[path] = action
		}
	}

	infos := make([]weblens.FileInfo, 0, len(actionsMap))
	for _, action := range actionsMap {
		isDir := strings.HasSuffix(action.GetDestinationPath(), "/")
		filename := filepath.Base(action.GetDestinationPath())
		m := types.SERV.MediaRepo.Get(j.lifetimes[action.GetLifetimeId()].GetContentId())

		var displayable bool
		if m != nil && m.GetMediaType() != nil {
			displayable = m.GetMediaType().IsDisplayable()
		}

		infos = append(
			infos,
			weblens.FileInfo{
				Id:             action.GetDestinationId(),
				Displayable:    displayable,
				IsDir:          isDir,
				Modifiable:     false,
				Size:           action.GetSize(),
				ModTime:        action.GetTimestamp().UnixMilli(),
				Filename:       filename,
				ParentFolderId: action.GetParentId(),
				MediaData:      m,
				Owner:          "",
				PathFromHome:   action.GetDestinationPath(),
				ShareId:        "",
				Children:       nil,
				PastFile:       true,
			},
		)
	}

	return infos, nil
}

func (j *JournalServiceImpl) GetLifetimeByFileId(fId FileId) *Lifetime {
	j.latestMapLock.Lock()
	defer j.latestMapLock.Unlock()
	return j.latestUpdate[fId]
}

func (j *JournalServiceImpl) Get(lId LifetimeId) *Lifetime {
	j.lifetimeMapLock.RLock()
	defer j.lifetimeMapLock.RUnlock()
	return j.lifetimes[lId]
}

func (j *JournalServiceImpl) Add(lt *Lifetime) error {
	// Check if this is a new or existing lifetime
	existing := j.Get(lt.ID())
	if existing != nil {
		// Check if the existing lifetime has a differing number of actions.
		if len(lt.GetActions()) != len(existing.GetActions()) {
			newActions := lt.GetActions()

			/* DEBUG - TODO remove if not needed */
			if !slices.IsSortedFunc(
				newActions, func(a, b *FileAction) int {
					return a.GetTimestamp().Compare(b.GetTimestamp())
				},
			) {
				wlog.Error.Printf("Actions for lifetime [%s] are NOT sorted", lt.ID())
			}
			/* END DEBUG */

			// Ensure that the actions are in time order, so we grab only the new ones to update
			slices.SortFunc(
				newActions, func(a, b *FileAction) int {
					return a.GetTimestamp().Compare(b.GetTimestamp())
				},
			)
			// Add every action that is newer than the previously existing latest to the lifetime
			for _, a := range newActions[len(existing.GetActions()):] {
				existing.Add(a)
			}

			// Update lifetime with new actions in mongo
			err := j.store.UpsertLifetime(lt)
			if err != nil {
				return err
			}
		} else {
			// If it were to have the same actions, it should not require an update
			return nil
		}
		lt = existing
	} else {
		// If the lifetime does not exist, just add it right to mongo
		err := j.store.UpsertLifetime(lt)
		if err != nil {
			return err
		}
	}

	// Add to lifetime map
	j.lifetimeMapLock.Lock()
	defer j.lifetimeMapLock.Unlock()
	j.lifetimes[lt.ID()] = lt

	// Add to latest update map
	j.latestMapLock.Lock()
	defer j.latestMapLock.Unlock()
	j.latestUpdate[lt.GetLatestFileId()] = lt

	return nil
}

func (j *JournalServiceImpl) Del(lId LifetimeId) error {
	return werror.NotImplemented("journal del")
}

func (j *JournalServiceImpl) Size() int {
	return len(j.lifetimes)
}

type JournalService interface {
	Init(store types.HistoryStore) error
	Size() int

	Get(id LifetimeId) *Lifetime
	Add(lifetime *Lifetime) error
	Del(id LifetimeId) error

	WatchFolder(f *WeblensFile) error

	LogEvent(fe *FileEvent) error

	GetActionsByPath(*WeblensFilepath) ([]*FileAction, error)
	GetPastFolderInfo(folder *WeblensFile, time time.Time) ([]weblens.FileInfo, error)

	JournalWorker()
	FileWatcher()
	GetActiveLifetimes() []*Lifetime
	GetAllLifetimes() []*Lifetime
	GetLifetimeByFileId(fId FileId) *Lifetime
}
