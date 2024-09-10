package fileTree

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ Journal = (*JournalImpl)(nil)

type JournalImpl struct {
	lifetimes   map[FileId]*Lifetime
	lifetimeMapLock sync.RWMutex
	eventStream chan *FileEvent

	serverId string

	fileTree *FileTreeImpl
	col      *mongo.Collection
}

func NewJournal(col *mongo.Collection, serverId string) (*JournalImpl, error) {
	j := &JournalImpl{
		lifetimes: make(map[FileId]*Lifetime),
		eventStream: make(chan *FileEvent, 10),
		col:         col,
		serverId:    serverId,
	}

	indexModel := mongo.IndexModel{
		Keys: bson.D{{"actions.timestamp", -1}},
	}
	_, err := col.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return nil, err
	}

	var lifetimes []*Lifetime

	lifetimes, err = getAllLifetimes(j.col)
	if err != nil {
		return nil, err
	}

	j.lifetimeMapLock.Lock()
	for _, l := range lifetimes {
		j.lifetimes[l.ID()] = l
	}
	j.lifetimeMapLock.Unlock()

	go j.EventWorker()

	return j, nil
}

func (j *JournalImpl) NewEvent() *FileEvent {
	return &FileEvent{
		EventId:    FileEventId(primitive.NewObjectID().Hex()),
		EventBegin: time.Now(),
		journal:    j,
		ServerId:   j.serverId,
	}
}

func (j *JournalImpl) SetFileTree(ft *FileTreeImpl) {
	j.fileTree = ft
}

func (j *JournalImpl) GetActiveLifetimes() []*Lifetime {
	var result []*Lifetime
	for _, l := range j.lifetimes {
		if l.IsLive() {
			result = append(result, l)
		}
	}
	return result
}

func (j *JournalImpl) GetAllLifetimes() []*Lifetime {
	j.lifetimeMapLock.RLock()
	defer j.lifetimeMapLock.RUnlock()
	return internal.MapToValues(j.lifetimes)
}

func (j *JournalImpl) LogEvent(fe *FileEvent) {
	log.Trace.Printf("Dropping off event with %d actions", len(fe.Actions))

	if fe != nil && len(fe.Actions) != 0 {
		j.eventStream <- fe
	}
}

func (j *JournalImpl) GetActionsByPath(path WeblensFilepath) ([]*FileAction, error) {
	return getActionsByPath(path, j.col)
}

func (j *JournalImpl) GetLatestAction() (*FileAction, error) {
	opts := options.FindOne().SetSort(bson.M{"actions.timestamp": -1})

	ret := j.col.FindOne(context.Background(), bson.M{}, opts)
	if ret.Err() != nil {
		if errors.Is(ret.Err(), mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, werror.WithStack(ret.Err())
	}

	var target Lifetime
	err := ret.Decode(&target)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	return target.Actions[len(target.Actions)-1], nil

}

func (j *JournalImpl) GetPastFolderChildren(folder *WeblensFileImpl, time time.Time) (
	[]*WeblensFileImpl, error,
) {
	actions, err := getActionsByPath(folder.GetPortablePath(), j.col)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	actionsMap := map[string]*FileAction{}
	for _, action := range actions {
		if action.GetTimestamp().UnixMilli() >= time.UnixMilli() || action.GetLifetimeId() == folder.ID() {
			continue
		}

		if _, ok := actionsMap[action.LifeId]; !ok {
			if action.ParentId == folder.ID() {
				actionsMap[action.LifeId] = action
			} else {
				actionsMap[action.LifeId] = nil
			}
		}
	}

	children := make([]*WeblensFileImpl, 0, len(actionsMap))
	for _, action := range actionsMap {
		if action == nil {
			continue
		}

		newChild := NewWeblensFile(
			action.GetLifetimeId(), filepath.Base(action.DestinationPath), folder,
			action.DestinationPath[len(action.DestinationPath)-1] == '/',
		)
		newChild.setModTime(time)
		newChild.setPastFile(true)
		newChild.size.Store(action.Size)
		newChild.contentId = j.lifetimes[action.LifeId].ContentId

		children = append(
			children, newChild,
		)
	}

	return children, nil
}

func (j *JournalImpl) Get(lId FileId) *Lifetime {
	j.lifetimeMapLock.RLock()
	defer j.lifetimeMapLock.RUnlock()
	return j.lifetimes[lId]
}

func (j *JournalImpl) Add(lt *Lifetime) error {
	// Check if this is a new or existing lifetime
	existing := j.Get(lt.ID())
	if existing != nil {
		// Check if the existing lifetime has a differing number of actions.
		if len(lt.GetActions()) != len(existing.GetActions()) {
			newActions := lt.GetActions()

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

	return nil
}

func (j *JournalImpl) GetLifetimesSince(date time.Time) ([]*Lifetime, error) {
	return getLifetimesSince(date, j.col)
}

func (j *JournalImpl) Close() {
	close(j.eventStream)
}

func (j *JournalImpl) EventWorker() {
	for {
		e, ok := <-j.eventStream
		if !ok {
			log.Debug.Println("Event worker exiting...")
			return
		}
		if e == nil {
			log.Error.Println("Got nil event in event stream...")
			continue
		}
		if err := j.handleFileEvent(e); err != nil {
			log.ErrTrace(err)
		}
	}
}

func (j *JournalImpl) handleFileEvent(event *FileEvent) error {
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

	for _, action := range actions {
		if action.GetFile() != nil {
			size := action.GetFile().Size()
			action.SetSize(size)
		}

		switch action.GetActionType() {
		case FileCreate:
			newL, err := NewLifetime(action)
			if err != nil {
				return err
			}

			if newL == nil {
				return werror.Errorf("failed to create new lifetime")
			}

			if _, ok := j.lifetimes[newL.ID()]; ok {
				return werror.Errorf("trying to add create action to already existing lifetime")
			}

			j.lifetimeMapLock.Lock()
			j.lifetimes[newL.ID()] = newL
			j.lifetimeMapLock.Unlock()
			updated = append(updated, newL)
		case FileMove:
			j.lifetimeMapLock.RLock()
			existing := j.lifetimes[action.LifeId]
			existing.Add(action)
			j.lifetimeMapLock.RUnlock()

			updated = append(updated, existing)
		case FileDelete:
			j.lifetimeMapLock.RLock()
			existing := j.lifetimes[action.LifeId]
			existing.Add(action)
			j.lifetimeMapLock.RUnlock()

			updated = append(updated, existing)
		}
	}

	log.Trace.Printf("Updating %d lifetimes", len(updated))

	for _, lt := range updated {
		// f := j.fileTree.Get(lt.ID())
		// if f != nil {
		// 	sz := f.Size()
		// 	if lt.GetContentId() == "" && !f.IsDir() && sz != 0 {
		// 		return werror.Errorf("No content ID in lifetime update")
		// 	}
		// } else if lt.GetLatestAction().GetActionType() != FileDelete {
		// 	return werror.Errorf("Could not find file for non-delete lifetime update")
		// }
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
		bson.D{{"$sort", bson.D{{"timestamp", -1}}}},
	}

	ret, err := col.Aggregate(context.Background(), pipe)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	var target []*FileAction
	err = ret.All(context.Background(), &target)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	return target, nil
}

func getLifetimesSince(date time.Time, col *mongo.Collection) ([]*Lifetime, error) {
	pipe := bson.A{
		bson.D{
			{
				"$match",
				bson.D{{"actions.timestamp", bson.D{{"$gt", date}}}},
			},
		},
		bson.D{{"$sort", bson.D{{"actions.timestamp", 1}}}},
	}
	ret, err := col.Aggregate(context.Background(), pipe)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	var target []*Lifetime
	err = ret.All(context.Background(), &target)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	return target, nil
}

type Journal interface {
	Get(id FileId) *Lifetime
	Add(lifetime *Lifetime) error

	SetFileTree(ft *FileTreeImpl)

	NewEvent() *FileEvent
	WatchFolder(f *WeblensFileImpl) error

	LogEvent(fe *FileEvent)

	GetActionsByPath(WeblensFilepath) ([]*FileAction, error)
	GetPastFolderChildren(folder *WeblensFileImpl, time time.Time) ([]*WeblensFileImpl, error)
	GetLatestAction() (*FileAction, error)
	GetLifetimesSince(date time.Time) ([]*Lifetime, error)

	EventWorker()
	FileWatcher()
	GetActiveLifetimes() []*Lifetime
	GetAllLifetimes() []*Lifetime
}
