package fileTree

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ Journal = (*JournalImpl)(nil)

type JournalImpl struct {
	lifetimes       map[FileId]*Lifetime
	lifetimeMapLock sync.RWMutex
	eventStream     chan *FileEvent

	serverId string

	fileTree *FileTreeImpl
	col      *mongo.Collection

	// Do not register actions that happen on the local server.
	// This is used in backup servers.
	ignoreLocal bool

	hasherFactory func() Hasher
}

func NewJournal(col *mongo.Collection, serverId string, ignoreLocal bool, hasherFactory func() Hasher) (
	*JournalImpl, error,
) {
	j := &JournalImpl{
		lifetimes:     make(map[FileId]*Lifetime),
		eventStream:   make(chan *FileEvent, 10),
		col:           col,
		serverId:      serverId,
		ignoreLocal:   ignoreLocal,
		hasherFactory: hasherFactory,
	}

	indexModel := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "actions.timestamp", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "actions.originPath", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "actions.destinationPath", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "serverId", Value: 1},
			},
		},
	}
	_, err := col.Indexes().CreateMany(context.Background(), indexModel)
	if err != nil {
		return nil, err
	}

	var lifetimes []*Lifetime

	lifetimes, err = getAllLifetimes(j.col, serverId)
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
		hasher:     j.hasherFactory(),

		LoggedChan: make(chan struct{}),
	}
}

func (j *JournalImpl) SetFileTree(ft *FileTreeImpl) {
	j.fileTree = ft
}

func (j *JournalImpl) IgnoreLocal() bool {
	return j.ignoreLocal
}

func (j *JournalImpl) SetIgnoreLocal(ignore bool) {
	j.ignoreLocal = ignore
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

func (j *JournalImpl) Clear() error {
	j.lifetimeMapLock.Lock()
	defer j.lifetimeMapLock.Unlock()
	j.lifetimes = make(map[FileId]*Lifetime)

	_, err := j.col.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		return werror.WithStack(err)
	}

	return nil
}

func (j *JournalImpl) LogEvent(fe *FileEvent) {
	if fe == nil {
		log.Warning.Println("Tried to log nil event")
		return
	} else if j.ignoreLocal {
		log.Trace.Func(func(l log.Logger) { l.Printf("Ignoring local file event [%s]", fe.EventId) })
		close(fe.LoggedChan)
		return
	}

	log.Debug.Func(func(l log.Logger) { l.Printf("Dropping off event with %d actions", len(fe.Actions)) })

	if len(fe.Actions) != 0 {
		j.eventStream <- fe
	} else {
		log.Debug.Func(func(l log.Logger) { l.Printf("File Event [%s] has no actions, skipping logging", fe.EventId) })
		log.TraceCaller(1, "Empty event is from here")
		close(fe.LoggedChan)
	}
}

func (j *JournalImpl) GetActionsByPath(path WeblensFilepath) ([]*FileAction, error) {
	return j.getActionsByPath(path, false)
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

func (j *JournalImpl) GetPastFile(id FileId, time time.Time) (*WeblensFileImpl, error) {
	// actions, err := j.getActionsByPath(path, true)
	// if err != nil {
	// 	return nil, err
	// }

	lt := j.Get(id)
	if lt == nil {
		return nil, werror.WithStack(werror.ErrNoFileAction)
	}

	actions := lt.Actions

	slices.SortFunc(
		actions, func(a, b *FileAction) int {
			return a.GetTimestamp().Compare(b.GetTimestamp())
		},
	)

	relevantAction := actions[len(actions)-1]
	counter := 1
	for relevantAction.GetTimestamp().After(time) || relevantAction.GetTimestamp().Equal(time) {
		counter++
		if len(actions)-counter < 0 {
			break
		}
		if actions[len(actions)-counter].ActionType == FileSizeChange {
			continue
		}
		relevantAction = actions[len(actions)-counter]
	}

	path := ParsePortable(relevantAction.DestinationPath)

	f := NewWeblensFile(relevantAction.LifeId, path.Filename(), nil, path.IsDir())
	f.parentId = relevantAction.ParentId
	f.portablePath = path
	f.pastFile = true
	f.SetContentId(lt.ContentId)
	f.setModTime(relevantAction.GetTimestamp())
	return f, nil
}

func (j *JournalImpl) UpdateLifetime(lifetime *Lifetime) error {
	j.lifetimeMapLock.Lock()
	defer j.lifetimeMapLock.Unlock()

	_, err := j.col.UpdateOne(context.Background(), bson.M{"_id": lifetime.ID()}, bson.M{"$set": lifetime})
	if err != nil {
		return werror.WithStack(err)
	}
	return nil
}

func (j *JournalImpl) GetPastFolderChildren(folder *WeblensFileImpl, time time.Time) (
	[]*WeblensFileImpl, error,
) {
	actions, err := j.getChildrenAtTime(folder.ID(), time)
	if err != nil {
		return nil, err
	}

	log.Trace.Printf("Got %d actions", len(actions))
	// actions, err := j.getActionsByPath(folder.GetPortablePath(), false)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// actionsMap := map[string]*FileAction{}
	// for _, action := range actions {
	// 	if action.GetTimestamp().After(time) || action.GetLifetimeId() == folder.ID() {
	// 		continue
	// 	}
	//
	// 	log.Trace.Func(func(l log.Logger) {
	// 		l.Printf("Action %s %s (%s == %s)", action.LifeId, action.DestinationPath, action.ActionType, action.ParentId, folder.ID())
	// 	})
	//
	// 	if _, ok := actionsMap[action.LifeId]; !ok {
	// 		if action.ParentId == folder.ID() {
	// 			actionsMap[action.LifeId] = action
	// 		} else {
	// 			actionsMap[action.LifeId] = nil
	// 		}
	// 	}
	// }

	lifeIdMap := map[FileId]any{}
	children := make([]*WeblensFileImpl, 0, len(actions))
	for _, action := range actions {
		if action == nil {
			continue
		}
		if _, ok := lifeIdMap[action.LifeId]; ok {
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

		lifeIdMap[action.LifeId] = nil
	}

	return children, nil
}

func (j *JournalImpl) Get(lId FileId) *Lifetime {
	j.lifetimeMapLock.RLock()
	defer j.lifetimeMapLock.RUnlock()
	return j.lifetimes[lId]
}

func (j *JournalImpl) Add(lts ...*Lifetime) error {
	var toWrite []*Lifetime
	for _, lt := range lts {

		// Make sure the lifetime is for this journal
		if lt.ServerId != j.serverId {
			return werror.WithStack(werror.ErrJournalServerMismatch)
		}

		// Check if this is a new or existing lifetime
		existing := j.Get(lt.ID())
		if existing != nil {
			// Check if the existing lifetime has a differing number of actions.
			if len(lt.GetActions()) != len(existing.GetActions()) {
				newActions := lt.GetActions()

				// Add every action that is newer than the previously existing latest to the lifetime
				for _, a := range newActions {
					if !existing.HasEvent(a.EventId) {
						existing.Add(a)
					}
				}

				toWrite = append(toWrite, lt)
			} else {
				// If it were to have the same actions, it should not require an update
				continue
			}
			// lt = existing
		} else {
			// If the lifetime does not exist, just add it right to mongo
			toWrite = append(toWrite, lt)
		}
	}

	j.lifetimeMapLock.Lock()
	defer j.lifetimeMapLock.Unlock()

	for _, lt := range toWrite {
		err := upsertLifetime(lt, j.col)
		if err != nil {
			return err
		}
		j.lifetimes[lt.ID()] = lt
	}

	return nil
}

func (j *JournalImpl) GetLifetimesSince(date time.Time) ([]*Lifetime, error) {
	return getLifetimesSince(date, j.col, j.serverId)
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
		log.Trace.Println("Journal event worker got event starting with", e.GetActions()[0].GetActionType())
		if err := j.handleFileEvent(e); err != nil {
			log.ErrTrace(err)
		}
		close(e.LoggedChan)
	}
}

func (j *JournalImpl) handleFileEvent(event *FileEvent) error {
	j.lifetimeMapLock.Lock()
	defer j.lifetimeMapLock.Unlock()
	log.Trace.Func(func(l log.Logger) { l.Printf("Handling event with %d actions", len(event.GetActions())) })

	defer func() {
		e := recover()
		if e != nil {
			err, ok := e.(error)
			if !ok {
				log.Error.Println(e)
			} else {
				log.ErrTrace(err)
			}
		}
	}()

	if len(event.GetActions()) == 0 {
		return nil
	}

	actions := event.GetActions()
	slices.SortFunc(
		actions, func(a, b *FileAction) int {
			return a.GetTimestamp().Compare(b.GetTimestamp())
		},
	)

	// Ensrue all async tasks spawned by the hasher have finished before continuing
	if waitHasher, ok := event.hasher.(HashWaiter); ok {
		waitHasher.Wait()
	}

	var updated []*Lifetime

	for _, action := range actions {
		if action.GetFile() != nil {
			size := action.GetFile().Size()
			action.SetSize(size)
		}

		log.Trace.Func(func(l log.Logger) { l.Printf("Handling %s for %s", action.GetActionType(), action.LifeId) })

		actionType := action.GetActionType()
		if actionType == FileCreate || actionType == FileRestore {
			if action.Size == -1 {
				action.file.LoadStat()
				action.Size = action.file.Size()
			}
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

			j.lifetimes[newL.ID()] = newL
			updated = append(updated, newL)
		} else if actionType == FileDelete || actionType == FileMove || actionType == FileSizeChange {
			existing := j.lifetimes[action.LifeId]
			existing.Add(action)

			updated = append(updated, existing)
		} else {
			return werror.Errorf("unknown file action type %s", actionType)
		}
	}

	for _, lt := range updated {
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

func getAllLifetimes(col *mongo.Collection, serverId string) ([]*Lifetime, error) {
	ret, err := col.Find(context.Background(), bson.M{"serverId": serverId})
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

func upsertLifetimes(lts []*Lifetime, col *mongo.Collection) error {
	many := []mongo.WriteModel{mongo.NewUpdateManyModel().SetFilter(bson.M{}).SetUpdate(lts).SetUpsert(true)}
	_, err := col.BulkWrite(context.Background(), many)
	if err != nil {
		return werror.WithStack(err)
	}

	return nil
}

func (j *JournalImpl) getChildrenAtTime(parentId FileId, time time.Time) ([]*FileAction, error) {
	pipe := bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "serverId", Value: j.serverId}}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$actions"}}}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "$and", Value: bson.A{bson.D{{Key: "actions.parentId", Value: parentId}}, bson.D{{Key: "actions.timestamp", Value: bson.D{{Key: "$lt", Value: time}}}}}}}}},
		bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$actions"}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
	}

	ret, err := j.col.Aggregate(context.Background(), pipe)
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

func (j *JournalImpl) getActionsByPath(path WeblensFilepath, noChildren bool) ([]*FileAction, error) {
	var pathMatch bson.A
	if noChildren {
		pathMatch = bson.A{
			bson.D{{Key: "actions.originPath", Value: path.ToPortable()}},
			bson.D{{Key: "actions.destinationPath", Value: path.ToPortable()}},
		}
	} else {
		pathMatch = bson.A{
			bson.D{{Key: "actions.originPath", Value: bson.D{{Key: "$regex", Value: path.ToPortable() + "[^/]*/?$"}}}},
			bson.D{{Key: "actions.destinationPath", Value: bson.D{{Key: "$regex", Value: path.ToPortable() + "[^/]*/?$"}}}},
		}
	}

	pipe := bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "serverId", Value: j.serverId}}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$actions"}}}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: pathMatch}}}},
		bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$actions"}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
	}

	ret, err := j.col.Aggregate(context.Background(), pipe)
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

func getLifetimesSince(date time.Time, col *mongo.Collection, serverId string) ([]*Lifetime, error) {
	pipe := bson.A{
		bson.D{
			{
				Key:   "$match",
				Value: bson.D{{Key: "actions.timestamp", Value: bson.D{{Key: "$gt", Value: date}}}, {Key: "serverId", Value: serverId}},
			},
		},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "actions.timestamp", Value: 1}}}},
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
	Add(lifetime ...*Lifetime) error

	SetFileTree(ft *FileTreeImpl)
	IgnoreLocal() bool
	SetIgnoreLocal(ignore bool)

	NewEvent() *FileEvent
	WatchFolder(f *WeblensFileImpl) error

	LogEvent(fe *FileEvent)

	GetPastFile(id FileId, time time.Time) (*WeblensFileImpl, error)
	GetActionsByPath(WeblensFilepath) ([]*FileAction, error)
	GetPastFolderChildren(folder *WeblensFileImpl, time time.Time) ([]*WeblensFileImpl, error)
	GetLatestAction() (*FileAction, error)
	GetLifetimesSince(date time.Time) ([]*Lifetime, error)
	UpdateLifetime(lifetime *Lifetime) error

	EventWorker()
	FileWatcher()
	GetActiveLifetimes() []*Lifetime
	GetAllLifetimes() []*Lifetime
	Clear() error
}
