package fileTree

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/creativecreature/sturdyc"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ Journal = (*JournalImpl)(nil)

type JournalImpl struct {
	eventStream chan *FileEvent

	fileTree *FileTreeImpl
	col      *mongo.Collection

	hasherFactory func() Hasher

	flushCond *sync.Cond

	log log.Bundle

	serverId string

	// Do not register actions that happen on the local server.
	// This is used in backup servers.
	ignoreLocal bool

	cache *sturdyc.Client[*Lifetime]
}

func NewJournal(col *mongo.Collection, serverId string, ignoreLocal bool, hasherFactory func() Hasher, logger log.Bundle) (
	*JournalImpl, error,
) {
	j := &JournalImpl{
		eventStream:   make(chan *FileEvent, 10),
		col:           col,
		serverId:      serverId,
		ignoreLocal:   ignoreLocal,
		hasherFactory: hasherFactory,
		log:           logger,
		flushCond:     sync.NewCond(&sync.Mutex{}),
		cache:         sturdyc.New[*Lifetime](10000, 10, time.Hour*2, 10),
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

	start := time.Now()
	lifetimes, err = getAllLifetimes(j.col, serverId)
	if err != nil {
		return nil, err
	}
	logger.Trace.Printf("Get all lifetimes in %s", time.Since(start))
	start = time.Now()

	for _, lt := range lifetimes {
		j.cache.Set(lt.ID(), lt)
	}
	logger.Trace.Printf("Add lifetimes to map in %s", time.Since(start))

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
	filter := bson.M{"actions.actionType": bson.M{"$ne": "fileDelete"}}
	res, err := j.col.Find(context.Background(), filter)
	if err != nil {
		j.log.ErrTrace(err)
		return nil
	}

	var target []*Lifetime
	err = res.All(context.Background(), &target)
	if err != nil {
		j.log.ErrTrace(err)
		return nil
	}

	for _, lt := range target {
		j.cache.Set(lt.ID(), lt)
	}

	return target
}

func (j *JournalImpl) GetAllLifetimes() []*Lifetime {
	filter := bson.M{}
	res, err := j.col.Find(context.Background(), filter)
	if err != nil {
		j.log.ErrTrace(err)
		return nil
	}

	var target []*Lifetime
	err = res.All(context.Background(), &target)
	if err != nil {
		j.log.ErrTrace(err)
		return nil
	}

	for _, lt := range target {
		j.cache.Set(lt.ID(), lt)
	}

	return target
}

func (j *JournalImpl) Clear() error {
	j.cache = sturdyc.New[*Lifetime](10000, 10, time.Hour*2, 10)

	_, err := j.col.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		return werror.WithStack(err)
	}

	return nil
}

func (j *JournalImpl) LogEvent(fe *FileEvent) {
	if fe == nil {
		j.log.Warning.Println("Tried to log nil event")
		return
	} else if j.ignoreLocal {
		j.log.Trace.Func(func(l log.Logger) { l.Printf("Ignoring local file event [%s]", fe.EventId) })
		close(fe.LoggedChan)
		return
	} else if fe.LoggedChan == nil {
		j.log.Warning.Println("Tried to log which has already been logged")
	}

	if len(fe.Actions) != 0 {
		j.log.Trace.Func(func(l log.Logger) { l.Printf("Dropping off event [%s] with %d actions", fe.EventId, len(fe.Actions)) })
		j.eventStream <- fe
	} else {
		j.log.Trace.Func(func(l log.Logger) { l.Printf("File Event [%s] has no actions, not logging", fe.EventId) })
		close(fe.LoggedChan)
	}
}

func (j *JournalImpl) Flush() {
	j.log.Trace.Println("Waiting for journal flush...")

	j.flushCond.L.Lock()
	for len(j.eventStream) > 0 {
		j.flushCond.Wait()
	}
	j.flushCond.L.Unlock()

	j.log.Trace.Println("Finished journal flush...")
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
	lt := j.Get(id)
	if lt == nil {
		return nil, werror.WithStack(werror.ErrNoFileAction)
	}

	actions := lt.GetActions()

	slices.SortFunc(
		actions, func(a, b *FileAction) int {
			return a.GetTimestamp().Compare(b.GetTimestamp())
		},
	)

	var err error
	if time.Unix() != 0 && actions[0].GetTimestamp().After(time) {
		actions, err = j.getActionsByPath(lt.GetLatestPath(), true)
		if err != nil {
			return nil, err
		}
		slices.SortFunc(
			actions, func(a, b *FileAction) int {
				return a.GetTimestamp().Compare(b.GetTimestamp())
			},
		)
	}

	relevantAction := actions[len(actions)-1]
	counter := 1
	for relevantAction.GetTimestamp().UnixMilli() >= time.UnixMilli() {
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

	f := NewWeblensFile(id, path.Filename(), nil, path.IsDir())
	f.parentId = relevantAction.ParentId
	f.portablePath = path
	f.pastFile = true
	f.pastId = relevantAction.LifeId
	f.SetContentId(lt.ContentId)
	f.setModTime(relevantAction.GetTimestamp())

	children, err := j.GetPastFolderChildren(f, time)
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		err = f.AddChild(child)
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}

func (j *JournalImpl) UpdateLifetime(lifetime *Lifetime) error {
	_, err := j.col.UpdateOne(context.Background(), bson.M{"_id": lifetime.ID()}, bson.M{"$set": lifetime})
	if err != nil {
		return werror.WithStack(err)
	}

	j.cache.Set(lifetime.ID(), lifetime)
	return nil
}

func (j *JournalImpl) GetPastFolderChildren(folder *WeblensFileImpl, time time.Time) (
	[]*WeblensFileImpl, error,
) {
	var id = folder.ID()
	if pastId := folder.GetPastId(); pastId != "" {
		id = pastId
	}

	actions, err := j.getChildrenAtTime(id, time)
	if err != nil {
		return nil, err
	}

	j.log.Trace.Printf("Got %d actions", len(actions))

	lifeIdMap := map[FileId]any{}
	children := []*WeblensFileImpl{}
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
		newChild.contentId = j.Get(action.LifeId).ContentId
		children = append(
			children, newChild,
		)

		lifeIdMap[action.LifeId] = nil
	}

	return children, nil
}

func (j *JournalImpl) Get(lId FileId) *Lifetime {
	ctx := context.Background()
	lt, err := j.cache.GetFetch(ctx, lId, j.fetchLifetime)
	if err != nil {
		j.log.ErrTrace(err)
		return nil
	}
	return lt
	// j.lifetimeMapLock.RLock()
	// defer j.lifetimeMapLock.RUnlock()
	// return j.lifetimes[lId]
}

func (j *JournalImpl) fetchLifetime(ctx context.Context) (*Lifetime, error) {
	lId := ctx.Value("lifetimeId")
	filter := bson.M{"_id": lId}
	res := j.col.FindOne(ctx, filter)
	if err := res.Err(); err != nil {
		return nil, werror.WithStack(err)
	}

	lt := &Lifetime{}
	err := res.Decode(lt)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	return lt, nil
}

func (j *JournalImpl) Add(lts ...*Lifetime) error {
	for _, lt := range lts {
		if lt.ServerId != j.serverId {
			return werror.WithStack(werror.ErrJournalServerMismatch)
		}

		err := upsertLifetime(lt, j.col)
		if err != nil {
			return err
		}
		j.log.Trace.Printf("Updating lifetime [%s] in map", lt.ID())
		j.cache.Set(lt.ID(), lt)
		// j.lifetimes[lt.ID()] = lt
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
			j.log.Debug.Println("Event worker exiting...")
			return
		}

		if e == nil {
			j.log.Error.Println("Got nil event in event stream...")
		} else {
			j.log.Trace.Println("Journal event worker got event starting with", e.GetActions()[0].GetActionType())
			j.flushCond.L.Lock()

			if err := j.handleFileEvent(e); err != nil {
				j.log.ErrTrace(err)
			}
			close(e.LoggedChan)
		}

		if len(j.eventStream) == 0 {
			j.flushCond.Broadcast()
		}
		j.log.Trace.Printf("Journal worker finishing %s event at %s", e.Actions[0].ActionType, e.Actions[0].DestinationPath)
		j.flushCond.L.Unlock()
	}
}

func (j *JournalImpl) handleFileEvent(event *FileEvent) error {
	event.LogLock.Lock()
	defer event.LogLock.Unlock()

	if event.Logged {
		j.log.Debug.Println("Skipping event already logged")
		return nil
	}

	j.log.Trace.Func(func(l log.Logger) { l.Printf("Handling event with %d actions", len(event.GetActions())) })

	defer func() {
		e := recover()
		if e != nil {
			err, ok := e.(error)
			if !ok {
				j.log.Error.Println(e)
			} else {
				j.log.ErrTrace(err)
			}
		}
	}()

	actions := event.GetActions()
	if len(actions) == 0 {
		return nil
	}

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

		j.log.Trace.Func(func(l log.Logger) {
			l.Printf("Handling %s for %s (%s)", action.GetActionType(), action.GetLifetimeId(), action.GetRelevantPath())
		})

		actionType := action.GetActionType()
		if actionType == FileCreate || actionType == FileRestore {
			if action.Size == -1 {
				_, err := action.file.LoadStat()
				if err != nil {
					j.log.ErrTrace(err)
					continue
				}
				action.Size = action.file.Size()
			}
			newL, err := NewLifetime(action)
			if err != nil {
				return err
			}

			if newL == nil {
				return werror.Errorf("failed to create new lifetime")
			}

			existing := j.Get(newL.ID())
			if existing != nil {
				panic(werror.Errorf("trying to add create action to already existing lifetime %s", newL.ID()))
				return werror.Errorf("trying to add create action to already existing lifetime: %s", newL.ID())
			}
			updated = append(updated, newL)
		} else if actionType == FileDelete || actionType == FileMove || actionType == FileSizeChange {
			existing := j.Get(action.LifeId)
			existing.Add(action)

			updated = append(updated, existing)
		} else {
			return werror.Errorf("unknown file action type %s", actionType)
		}
	}

	err := j.Add(updated...)
	if err != nil {
		return err
	}

	event.Logged = true
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
		bson.D{{Key: "$match", Value: bson.D{{Key: "actions.parentId", Value: parentId}}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$actions"}}}},
		bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$actions"}}}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "timestamp", Value: bson.D{{Key: "$lte", Value: time}}}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
		bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$lifeId"},
			{Key: "latest", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}},
		},
		},
		},
		bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$latest"}}}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "parentId", Value: parentId}}}},
	}

	// pipe := bson.A{
	// 	bson.D{{Key: "$match", Value: bson.D{{Key: "serverId", Value: j.serverId}}}},
	// 	bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$actions"}}}},
	// 	bson.D{{Key: "$match", Value: bson.D{{Key: "$and", Value: bson.A{bson.D{{Key: "actions.parentId", Value: parentId}}, bson.D{{Key: "actions.timestamp", Value: bson.D{{Key: "$lt", Value: time}}}}}}}}},
	// 	bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$actions"}}}},
	// 	bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
	// }

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
	Flush()

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
