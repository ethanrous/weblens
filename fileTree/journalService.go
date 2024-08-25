package fileTree

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/internal/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ JournalService = (*JournalServiceImpl)(nil)

type JournalServiceImpl struct {
	lifetimes       map[LifetimeId]*Lifetime
	lifetimeMapLock sync.RWMutex
	latestUpdate  map[FileId]*Lifetime
	latestMapLock sync.Mutex
	eventStream chan *FileEvent

	serverId string

	fileTree *FileTreeImpl
	col      *mongo.Collection
}

func NewJournalService(col *mongo.Collection, serverId string) (JournalService, error) {
	j := &JournalServiceImpl{
		lifetimes:    make(map[LifetimeId]*Lifetime),
		latestUpdate: make(map[FileId]*Lifetime),
		eventStream: make(chan *FileEvent, 10),
		col:         col,
		serverId:    serverId,
	}

	var lifetimes []*Lifetime
	// var updatedLifetimes []*Lifetime
	var err error
	// var hasProxy bool

	// if proxyStore, hasProxy = store.(types.ProxyStore); hasProxy {
	if false {
		// Get all lifetimes from the local database
		// localLifetimes, err := proxyStore.GetLocalStore().GetAllLifetimes()
		// if err != nil {
		// 	return err
		// }
		//
		// sw.Lap("Read all local lifetimes")
		//
		// j.lifetimeMapLock.Lock()
		// j.latestMapLock.Lock()
		// for _, l := range localLifetimes {
		// 	j.lifetimes[l.ID()] = l
		// 	j.latestUpdate[l.GetLatestFileId()] = l
		// }
		// j.latestMapLock.Unlock()
		// j.lifetimeMapLock.Unlock()
		//
		// sw.Lap("Add all local lifetimes")
		//
		// latest, err := proxyStore.GetLatestAction()
		// if err != nil {
		// 	return err
		// }
		//
		// var latestTimestamp time.Time
		// if latest != nil {
		// 	latestTimestamp = latest.GetTimestamp()
		// }
		//
		// remoteLifetimes, err := types.SERV.StoreService.GetLifetimesSince(latestTimestamp)
		// if err != nil {
		// 	return err
		// }
		//
		// sw.Lap("Proxy got lifetime updates")
		//
		// // Upsert lifetimes that have been updated on remote server
		// for _, l := range remoteLifetimes {
		// 	err = j.Add(l)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		//
		// sw.Lap("Proxy upsert lifetime updates")
	} else {
		lifetimes, err = getAllLifetimes(j.col)
		if err != nil {
			return nil, err
		}

		j.lifetimeMapLock.Lock()
		j.latestMapLock.Lock()
		for _, l := range lifetimes {
			j.lifetimes[l.ID()] = l
			j.latestUpdate[l.GetLatestFileId()] = l
		}
		j.latestMapLock.Unlock()
		j.lifetimeMapLock.Unlock()
	}

	return j, nil
}

func getAllLifetimes(col *mongo.Collection) ([]*Lifetime, error) {
	ret, err := col.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}

	var target []*Lifetime
	err = ret.All(context.Background(), &target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func (j *JournalServiceImpl) NewEvent() *FileEvent {
	return &FileEvent{
		EventId:    FileEventId(primitive.NewObjectID().Hex()),
		EventBegin: time.Now(),
		journal:    j,
		ServerId:   j.serverId,
	}
}

func (j *JournalServiceImpl) SetFileTree(ft *FileTreeImpl) {
	j.fileTree = ft
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

func (j *JournalServiceImpl) LogEvent(fe *FileEvent) {
	if fe != nil && len(fe.Actions) != 0 {
		j.eventStream <- fe
	}
}

func (j *JournalServiceImpl) GetActionsByPath(path WeblensFilepath) ([]*FileAction, error) {
	return getActionsByPath(path, j.col)
}

func (j *JournalServiceImpl) GetPastFolderInfo(folder *WeblensFile, time time.Time) (
	[]*WeblensFile, error,
) {
	actions, err := getActionsByPath(folder.GetPortablePath(), j.col)
	if err != nil {
		return nil, werror.WithStack(err)
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

	children := make([]*WeblensFile, 0, len(actionsMap))
	for _, action := range actionsMap {
		children = append(children, j.fileTree.Get(action.GetDestinationId()))
		// isDir := strings.HasSuffix(action.GetDestinationPath(), "/")
		// filename := filepath.Base(action.GetDestinationPath())
		// m := types.SERV.MediaRepo.Get(j.lifetimes[action.GetLifetimeId()].GetContentId())
		//
		// var displayable bool
		// if m != nil && m.GetMediaType() != nil {
		// 	displayable = m.GetMediaType().IsDisplayable()
		// }
		//
		// infos = append(
		// 	infos,
		// 	weblens.FileInfo{
		// 		Id:             action.GetDestinationId(),
		// 		Displayable:    displayable,
		// 		IsDir:          isDir,
		// 		Modifiable:     false,
		// 		Size:           action.GetSize(),
		// 		ModTime:        action.GetTimestamp().UnixMilli(),
		// 		Filename:       filename,
		// 		ParentFolderId: action.GetParentId(),
		// 		MediaData:      m,
		// 		Owner:          "",
		// 		PathFromHome:   action.GetDestinationPath(),
		// 		ShareId:        "",
		// 		Children:       nil,
		// 		PastFile:       true,
		// 	},
		// )
	}

	return children, nil
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
				log.Error.Printf("Actions for lifetime [%s] are NOT sorted", lt.ID())
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
			err := upsertLifetime(lt, j.col)
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
		err := upsertLifetime(lt, j.col)
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

func (j *JournalServiceImpl) GetLifetimesSince(date time.Time) ([]*Lifetime, error) {
	return getLifetimesSince(date, j.col)
}

func (j *JournalServiceImpl) EventWorker() {
	for {
		e := <-j.eventStream
		if err := j.handleFileEvent(e); err != nil {
			log.ErrTrace(err)
		}
	}
}

func (j *JournalServiceImpl) handleFileEvent(event *FileEvent) error {
	if len(event.GetActions()) == 0 {
		return nil
	}

	actions := event.GetActions()
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

	for _, action := range event.GetActions() {
		if action.GetFile() != nil {
			size, err := action.GetFile().Size()
			if err != nil {
				return err
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
				return werror.New("failed to create new lifetime")
			}

			// f := types.SERV.FileTree.Get(action.GetDestinationId())
			// sz, _ := f.Size()
			// if !f.IsDir() && sz != 0 {
			// 	contentId := f.GetContentId()
			// 	if contentId == "" {
			// 		return werror.New(
			// 			fmt.Sprintln(
			// 				"No content ID in FileCreate lifetime update", action.GetDestinationPath(),
			// 			),
			// 		)
			// 	}
			// 	newL.SetContentId(contentId)
			// }

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

	for _, lt := range updated {
		f := j.fileTree.Get(lt.GetLatestFileId())
		if f != nil {
			sz, _ := f.Size()
			if lt.GetContentId() == "" && !f.IsDir() && sz != 0 {
				return werror.New("No content ID in lifetime update")
			}
		} else if lt.GetLatestAction().GetActionType() != FileDelete {
			return werror.New("Could not find file for non-delete lifetime update")
		}
		filter := bson.M{"_id": lt.ID()}
		update := bson.M{"$set": lt}
		o := options.Update().SetUpsert(true)
		_, err := j.col.UpdateOne(context.Background(), filter, update, o)
		if err != nil {
			return err
		}
	}

	return nil
}

func upsertLifetime(lt *Lifetime, col *mongo.Collection) error {
	filter := bson.M{"_id": lt.ID()}
	update := bson.M{"$set": lt}
	o := options.Update().SetUpsert(true)
	_, err := col.UpdateOne(context.Background(), filter, update, o)

	return err
}

func getActionsByPath(path WeblensFilepath, col *mongo.Collection) ([]*FileAction, error) {
	pipe := bson.A{
		bson.D{{"$unwind", bson.D{{"path", "$actions"}}}},
		bson.D{
			{
				"$match",
				bson.D{
					{
						"$or",
						bson.A{
							bson.D{{"actions.originPath", bson.D{{"$regex", path.ToPortable() + "[^/]*/?$"}}}},
							bson.D{{"actions.destinationPath", bson.D{{"$regex", path.ToPortable() + "[^/]*/?$"}}}},
						},
					},
				},
			},
		},
		bson.D{{"$replaceRoot", bson.D{{"newRoot", "$actions"}}}},
	}

	ret, err := col.Aggregate(context.TODO(), pipe)
	if err != nil {
		return nil, err
	}

	var target []*FileAction
	err = ret.All(context.Background(), &target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func getLifetimesSince(date time.Time, col *mongo.Collection) ([]*Lifetime, error) {
	pipe := bson.A{
		// bson.D{{"$unwind", bson.D{{"path", "$actions"}}}},
		bson.D{
			{
				"$match",
				bson.D{{"actions.timestamp", bson.D{{"$gt", date}}}},
			},
		},
		// bson.D{{"$replaceRoot", bson.D{{"newRoot", "$actions"}}}},
		bson.D{{"$sort", bson.D{{"actions.timestamp", 1}}}},
	}
	ret, err := col.Aggregate(context.Background(), pipe)
	if err != nil {
		return nil, err
	}

	var target []*Lifetime
	err = ret.All(context.Background(), &target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

type HollowJournalService struct{}

func (h HollowJournalService) Get(id LifetimeId) *Lifetime {
	return nil
}

func (h HollowJournalService) Add(lifetime *Lifetime) error {
	return nil
}

func (h HollowJournalService) Del(id LifetimeId) error {
	return nil
}

func (h HollowJournalService) SetFileTree(ft *FileTreeImpl) {}

func (h HollowJournalService) NewEvent() *FileEvent {
	return &FileEvent{}
}

func (h HollowJournalService) WatchFolder(f *WeblensFile) error {
	return nil
}

func (h HollowJournalService) LogEvent(fe *FileEvent) {}

func (h HollowJournalService) GetActionsByPath(filepath WeblensFilepath) ([]*FileAction, error) {
	return nil, nil
}

func (h HollowJournalService) GetPastFolderInfo(folder *WeblensFile, time time.Time) ([]*WeblensFile, error) {
	return nil, nil
}

func (h HollowJournalService) GetLifetimesSince(date time.Time) ([]*Lifetime, error) {
	return nil, nil
}

func (h HollowJournalService) EventWorker() {}

func (h HollowJournalService) FileWatcher() {}

func (h HollowJournalService) GetActiveLifetimes() []*Lifetime {
	return nil
}

func (h HollowJournalService) GetAllLifetimes() []*Lifetime {
	return nil
}

func (h HollowJournalService) GetLifetimeByFileId(fId FileId) *Lifetime {
	return nil
}

func NewHollowJournalService() JournalService {
	return &HollowJournalService{}
}

type JournalService interface {
	Get(id LifetimeId) *Lifetime
	Add(lifetime *Lifetime) error
	Del(id LifetimeId) error

	SetFileTree(ft *FileTreeImpl)

	NewEvent() *FileEvent
	WatchFolder(f *WeblensFile) error

	LogEvent(fe *FileEvent)

	GetActionsByPath(WeblensFilepath) ([]*FileAction, error)
	GetPastFolderInfo(folder *WeblensFile, time time.Time) ([]*WeblensFile, error)
	GetLifetimesSince(date time.Time) ([]*Lifetime, error)

	EventWorker()
	FileWatcher()
	GetActiveLifetimes() []*Lifetime
	GetAllLifetimes() []*Lifetime
	GetLifetimeByFileId(fId FileId) *Lifetime
}
