package history

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type journalService struct {
	lifetimes       map[types.LifetimeId]types.Lifetime
	lifetimeMapLock *sync.RWMutex
	latestUpdate    map[types.FileId]types.Lifetime
	latestMapLock   *sync.Mutex

	store types.HistoryStore
}

func NewService(fileTree types.FileTree) types.JournalService {
	if fileTree == nil {
		return nil
	}
	return &journalService{
		lifetimes:       make(map[types.LifetimeId]types.Lifetime),
		lifetimeMapLock: &sync.RWMutex{},
		latestUpdate:    make(map[types.FileId]types.Lifetime),
		latestMapLock:   new(sync.Mutex),
	}
}

func (j *journalService) Init(store types.HistoryStore) error {
	j.store = store

	var lifetimes []types.Lifetime
	// var updatedLifetimes []types.Lifetime
	var err error
	var hasProxy bool
	var proxyStore types.ProxyStore

	sw := util.NewStopwatch("Journal init")

	if proxyStore, hasProxy = store.(types.ProxyStore); hasProxy {
		// Get all lifetimes from the local database
		localLifetimes, err := proxyStore.GetLocalStore().GetAllLifetimes()
		if err != nil {
			return err
		}

		sw.Lap("Proxy read all local lifetimes")

		j.lifetimeMapLock.Lock()
		j.latestMapLock.Lock()
		for _, l := range localLifetimes {
			j.lifetimes[l.ID()] = l
			j.latestUpdate[l.GetLatestFileId()] = l
		}
		j.latestMapLock.Unlock()
		j.lifetimeMapLock.Unlock()

		sw.Lap("Proxy add all local lifetimes")

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
	}
	lifetimes, err = store.GetAllLifetimes()
	if err != nil {
		return err
	}

	sw.Lap("Got all lifetimes")

	if hasProxy {
		if len(lifetimes) != len(j.lifetimes) {
			util.Error.Println("Local lifetime count does not match remote lifetime count")
		}
	} else {
		slices.SortFunc(
			lifetimes, func(a, b types.Lifetime) int {
				aActions := a.GetActions()
				bActions := b.GetActions()
				return len(aActions[len(aActions)-1].GetDestinationPath()) - len(bActions[len(bActions)-1].GetDestinationPath())
			},
		)

		sw.Lap("Sorted all lifetimes")

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

func (j *journalService) GetActiveLifetimes() []types.Lifetime {
	var result []types.Lifetime
	for _, l := range j.lifetimes {
		if l.IsLive() {
			result = append(result, l)
		}
	}
	return result
}

func (j *journalService) GetAllLifetimes() []types.Lifetime {
	j.lifetimeMapLock.RLock()
	defer j.lifetimeMapLock.RUnlock()
	return util.MapToValues(j.lifetimes)
}

func (j *journalService) GetAllFileEvents() ([]types.FileEvent, error) {
	// return util.MapToSlicePure(j.)
	// return types.SERV.StoreService.GetAllFileEvents()
	return nil, types.ErrNotImplemented("get all file events")
}

func (j *journalService) LogEvent(fe types.FileEvent) error {
	if len(fe.GetActions()) == 0 {
		return nil
	}

	actions := fe.GetActions()
	slices.SortFunc(
		actions, func(a, b types.FileAction) int {
			return a.GetTimestamp().Compare(b.GetTimestamp())
		},
	)

	var updated []types.Lifetime

	j.latestMapLock.Lock()
	defer j.latestMapLock.Unlock()
	j.lifetimeMapLock.Lock()
	defer j.lifetimeMapLock.Unlock()

	for _, action := range fe.GetActions() {
		if action.GetFile() != nil {
			size, err := action.GetFile().Size()
			if err != nil {
				return types.WeblensErrorFromError(err)
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
				return types.WeblensErrorMsg("failed to create new lifetime")
			}

			f := types.SERV.FileTree.Get(action.GetDestinationId())
			sz, _ := f.Size()
			if !f.IsDir() && sz != 0 {
				contentId := f.GetContentId()
				if contentId == "" {
					return types.NewWeblensError(
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
				return types.WeblensErrorMsg("No content ID in lifetime update")
			}
		} else if update.GetLatestAction().GetActionType() != FileDelete {
			return types.WeblensErrorMsg("Could not find file for non-delete lifetime update")
		}
		err := types.SERV.StoreService.UpsertLifetime(update)
		if err != nil {
			return err
		}
	}

	return nil
}

func (j *journalService) GetActionsByPath(path types.WeblensFilepath) ([]types.FileAction, error) {
	return j.store.GetActionsByPath(path)
}

func (j *journalService) GetPastFolderInfo(folder types.WeblensFile, time time.Time) ([]types.FileInfo, error) {
	actions, err := types.SERV.StoreService.GetActionsByPath(folder.GetPortablePath())
	if err != nil {
		return nil, types.WeblensErrorFromError(err)
	}

	actionsMap := map[string]types.FileAction{}
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

	infos := make([]types.FileInfo, 0, len(actionsMap))
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
			types.FileInfo{
				Id:             action.GetDestinationId(),
				Imported:       true,
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

func (j *journalService) GetLifetimeByFileId(fId types.FileId) types.Lifetime {
	j.latestMapLock.Lock()
	defer j.latestMapLock.Unlock()
	return j.latestUpdate[fId]
}

func (j *journalService) Get(lId types.LifetimeId) types.Lifetime {
	j.lifetimeMapLock.RLock()
	defer j.lifetimeMapLock.RUnlock()
	return j.lifetimes[lId]
}

func (j *journalService) Add(lt types.Lifetime) error {
	existing := j.Get(lt.ID())
	if existing != nil {
		if len(lt.GetActions()) != len(existing.GetActions()) {
			newActions := lt.GetActions()
			slices.SortFunc(
				newActions, func(a, b types.FileAction) int {
					return a.GetTimestamp().Compare(b.GetTimestamp())
				},
			)
			// newActionsCount := len(newActions) - len(existing.GetActions())
			for _, a := range newActions[len(existing.GetActions()):] {
				existing.Add(a)
			}

			err := j.store.UpsertLifetime(lt)
			if err != nil {
				return err
			}
		}
		lt = existing
	} else {
		err := j.store.UpsertLifetime(lt)
		if err != nil {
			return err
		}
	}

	j.lifetimeMapLock.Lock()
	defer j.lifetimeMapLock.Unlock()
	j.lifetimes[lt.ID()] = lt

	j.latestMapLock.Lock()
	defer j.latestMapLock.Unlock()
	j.latestUpdate[lt.GetLatestFileId()] = lt

	return nil
}

func (j *journalService) Del(lId types.LifetimeId) error {
	return types.ErrNotImplemented("journal del")
}

func (j *journalService) Size() int {
	return len(j.lifetimes)
}
