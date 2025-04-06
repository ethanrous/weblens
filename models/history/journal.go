package history

// import (
//
//	"context"
//	"path/filepath"
//	"slices"
//	"sync"
//	"time"
//
//	"github.com/ethanrous/weblens/models/db"
//	file_model "github.com/ethanrous/weblens/models/file"
//	"github.com/ethanrous/weblens/modules/fs"
//	"github.com/pkg/errors"
//	"github.com/rs/zerolog"
//	"github.com/viccon/sturdyc"
//	"go.mongodb.org/mongo-driver/bson"
//	"go.mongodb.org/mongo-driver/mongo"
//	"go.mongodb.org/mongo-driver/mongo/options"
//
// )
const FileHistoryCollectionKey = "fileHistory"

//
// var ErrLifetimeAlreadyExists = errors.New("lifetime already exists")
//
// func GetActionsByTowerId(ctx context.Context, towerId string) ([]*FileAction, error) {
// 	col, err := db.GetCollection(ctx, FileHistoryCollectionKey)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	filter := bson.M{"towerId": towerId}
// 	res, err := col.Find(ctx, filter)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var target []*FileAction
// 	err = res.All(ctx, &target)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return target, nil
// }
//
// func GetLastJournalUpdate(ctx context.Context) (time.Time, error) {
// 	col, err := db.GetCollection(ctx, FileHistoryCollectionKey)
// 	if err != nil {
// 		return time.Time{}, errors.WithStack(err)
// 	}
//
// 	opts := options.FindOne().SetSort(bson.M{"actions.timestamp": -1})
// 	ret := col.FindOne(ctx, bson.M{}, opts)
// 	if ret.Err() != nil {
// 		if errors.Is(ret.Err(), mongo.ErrNoDocuments) {
// 			return time.Time{}, nil
// 		}
// 		return time.Time{}, errors.WithStack(ret.Err())
// 	}
//
// 	var target Lifetime
// 	err = ret.Decode(&target)
// 	if err != nil {
// 		return time.Time{}, errors.WithStack(err)
// 	}
//
// 	latestAction := target.GetLatestAction()
// 	if latestAction == nil {
// 		return time.Time{}, nil
// 	}
//
// 	return latestAction.GetTimestamp(), nil
// }
//
// // WriteNewLifetime inserts a new lifetime into the database.
// // If a lifetime with the same ID already exists, it will not be overwritten.
// func WriteNewLifetime(ctx context.Context, lt *Lifetime) error {
// 	col, err := db.GetCollection(ctx, FileHistoryCollectionKey)
// 	if err != nil {
// 		return errors.WithStack(err)
// 	}
//
// 	opts := options.InsertOne().SetBypassDocumentValidation(false)
// 	_, err = col.InsertOne(ctx, lt, opts)
// 	if err != nil {
// 		if mongo.IsDuplicateKeyError(err) {
// 			return errors.WithStack(ErrLifetimeAlreadyExists)
// 		}
// 		return errors.WithStack(err)
// 	}
// 	return nil
// }
//
// type JournalImpl struct {
// 	eventStream chan *FileEvent
//
// 	fileTree *FileTreeImpl
// 	col      *mongo.Collection
//
// 	hasherFactory func() Hasher
//
// 	flushCond *sync.Cond
//
// 	log *zerolog.Logger
//
// 	serverId string
//
// 	// Do not register actions that happen on the local server.
// 	// This is used in backup servers.
// 	ignoreLocal bool
//
// 	cache *sturdyc.Client[*Lifetime]
// }
//
// type JournalConfig struct {
// 	Collection    *mongo.Collection
// 	ServerId      string
// 	IgnoreLocal   bool
// 	HasherFactory func() Hasher
// 	Logger        *zerolog.Logger
// }
//
// func NewJournal(cnf JournalConfig) (
// 	*JournalImpl, error,
// ) {
// 	newLogger := cnf.Logger.With().Str("service", "journal").Logger()
// 	if cnf.HasherFactory == nil {
// 		return nil, errors.New("Hasher factory cannot be nil")
// 	}
//
// 	j := &JournalImpl{
// 		eventStream:   make(chan *FileEvent, 10),
// 		col:           cnf.Collection,
// 		serverId:      cnf.ServerId,
// 		ignoreLocal:   cnf.IgnoreLocal,
// 		hasherFactory: cnf.HasherFactory,
// 		log:           &newLogger,
// 		flushCond:     sync.NewCond(&sync.Mutex{}),
// 		cache:         sturdyc.New[*Lifetime](10000, 10, time.Hour*2, 10),
// 	}
//
// 	indexModel := []mongo.IndexModel{
// 		{
// 			Keys: bson.D{
// 				{Key: "actions.timestamp", Value: -1},
// 			},
// 		},
// 		{
// 			Keys: bson.D{
// 				{Key: "actions.originPath", Value: 1},
// 			},
// 		},
// 		{
// 			Keys: bson.D{
// 				{Key: "actions.destinationPath", Value: 1},
// 			},
// 		},
// 		{
// 			Keys: bson.D{
// 				{Key: "serverId", Value: 1},
// 			},
// 		},
// 	}
// 	_, err := j.col.Indexes().CreateMany(context.Background(), indexModel)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var lifetimes []*Lifetime
//
// 	start := time.Now()
// 	lifetimes, err = getAllLifetimes(j.col, cnf.ServerId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	j.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Get all lifetimes in %s", time.Since(start)) })
// 	start = time.Now()
//
// 	for _, lt := range lifetimes {
// 		j.cache.Set(lt.ID(), lt)
// 	}
// 	j.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Add lifetimes to map in %s", time.Since(start)) })
//
// 	go j.EventWorker()
//
// 	return j, nil
// }
//
// func (j *JournalImpl) NewEvent() *FileEvent {
// 	hasher := j.hasherFactory()
//
// 	if hasher == nil {
// 		j.log.Error().Msgf("Hasher is nil trying to create new file event")
// 		return nil
// 	}
//
// 	return NewFileEvent(j, j.serverId, hasher)
// }
//
// func (j *JournalImpl) SetFileTree(ft *FileTreeImpl) {
// 	j.fileTree = ft
// }
//
// func (j *JournalImpl) IgnoreLocal() bool {
// 	return j.ignoreLocal
// }
//
// func (j *JournalImpl) SetIgnoreLocal(ignore bool) {
// 	j.ignoreLocal = ignore
// }
//
// func (j *JournalImpl) GetActiveLifetimes() []*Lifetime {
// 	filter := bson.M{"actions.actionType": bson.M{"$ne": "fileDelete"}, "serverId": j.serverId}
// 	res, err := j.col.Find(context.Background(), filter)
// 	if err != nil {
// 		j.log.Error().Stack().Err(err).Msg("")
// 		return nil
// 	}
//
// 	var target []*Lifetime
// 	err = res.All(context.Background(), &target)
// 	if err != nil {
// 		j.log.Error().Stack().Err(err).Msg("")
// 		return nil
// 	}
//
// 	for _, lt := range target {
// 		j.cache.Set(lt.ID(), lt)
// 	}
//
// 	return target
// }
//
// func (j *JournalImpl) GetAllLifetimes() []*Lifetime {
// 	filter := bson.M{}
// 	res, err := j.col.Find(context.Background(), filter)
// 	if err != nil {
// 		j.log.Error().Stack().Err(err).Msg("")
// 		return nil
// 	}
//
// 	var target []*Lifetime
// 	err = res.All(context.Background(), &target)
// 	if err != nil {
// 		j.log.Error().Stack().Err(err).Msg("")
// 		return nil
// 	}
//
// 	for _, lt := range target {
// 		j.cache.Set(lt.ID(), lt)
// 	}
//
// 	return target
// }
//
// func (j *JournalImpl) Clear() error {
// 	j.cache = sturdyc.New[*Lifetime](10000, 10, time.Hour*2, 10)
//
// 	_, err := j.col.DeleteMany(context.Background(), bson.M{})
// 	if err != nil {
// 		return errors.WithStack(err)
// 	}
//
// 	return nil
// }
//
// func (j *JournalImpl) LogEvent(fe *FileEvent) {
// 	if fe == nil {
// 		j.log.Warn().Msgf("Tried to log nil event")
// 		return
// 	} else if fe.Logged.Load() {
// 		j.log.Warn().Msgf("Tried to log which has already been logged")
// 		return
// 	} else if j.ignoreLocal {
// 		j.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Ignoring local file event [%s]", fe.EventId) })
// 		fe.SetLogged()
// 		return
// 	}
//
// 	if len(fe.Actions) != 0 {
// 		j.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Dropping off event [%s] with %d actions", fe.EventId, len(fe.Actions)) })
// 		j.eventStream <- fe
// 	} else {
// 		j.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("File Event [%s] has no actions, not logging", fe.EventId) })
// 		fe.SetLogged()
// 	}
// }
//
// func (j *JournalImpl) Flush() {
// 	j.log.Trace().Msg("Waiting for journal flush...")
//
// 	j.flushCond.L.Lock()
// 	for len(j.eventStream) > 0 {
// 		j.flushCond.Wait()
// 	}
// 	j.flushCond.L.Unlock()
//
// 	j.log.Trace().Msg("Finished journal flush...")
// }
//
// func (j *JournalImpl) GetActionsByPath(path fs.Filepath) ([]*FileAction, error) {
// 	return j.getActionsByPath(path, false)
// }
//
// func (j *JournalImpl) GetLatestAction() (*FileAction, error) {
// 	opts := options.FindOne().SetSort(bson.M{"actions.timestamp": -1})
//
// 	ret := j.col.FindOne(context.Background(), bson.M{}, opts)
// 	if ret.Err() != nil {
// 		if errors.Is(ret.Err(), mongo.ErrNoDocuments) {
// 			return nil, nil
// 		}
// 		return nil, errors.WithStack(ret.Err())
// 	}
//
// 	var target Lifetime
// 	err := ret.Decode(&target)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
//
// 	return target.Actions[len(target.Actions)-1], nil
//
// }
//
// func (j *JournalImpl) UpdateLifetime(lifetime *Lifetime) error {
// 	_, err := j.col.UpdateOne(context.Background(), bson.M{"_id": lifetime.ID()}, bson.M{"$set": lifetime})
// 	if err != nil {
// 		return errors.WithStack(err)
// 	}
//
// 	j.cache.Set(lifetime.ID(), lifetime)
// 	return nil
// }
//
// func (j *JournalImpl) GetPastFolderChildren(folder *file_model.WeblensFileImpl, time time.Time) (
// 	[]*file_model.WeblensFileImpl, error,
// ) {
// 	var id = folder.ID()
// 	if pastId := folder.GetPastId(); pastId != "" {
// 		id = pastId
// 	}
//
// 	actions, err := j.getChildrenAtTime(id, time)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	j.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Got %d actions", len(actions)) })
//
// 	lifeIdMap := map[string]any{}
// 	children := []*file_model.WeblensFileImpl{}
// 	for _, action := range actions {
// 		if action == nil {
// 			continue
// 		}
// 		if _, ok := lifeIdMap[action.LifeId]; ok {
// 			continue
// 		}
//
// 		newChild := file_model.NewWeblensFile(
// 			action.GetLifetimeId(), filepath.Base(action.DestinationPath), folder,
// 			action.DestinationPath[len(action.DestinationPath)-1] == '/',
// 		)
// 		newChild.setModifyDate(time)
// 		newChild.setPastFile(true)
// 		newChild.size.Store(action.Size)
// 		newChild.contentId = j.Get(action.LifeId).ContentId
// 		children = append(
// 			children, newChild,
// 		)
//
// 		lifeIdMap[action.LifeId] = nil
// 	}
//
// 	return children, nil
// }
//
// func (j *JournalImpl) Get(lId string) *Lifetime {
// 	ctx := context.Background()
// 	ctx = context.WithValue(ctx, "lifetimeId", lId)
// 	lt, err := j.cache.GetOrFetch(ctx, lId, j.fetchLifetime)
// 	if err != nil {
// 		j.log.Error().Stack().Err(err).Msg("")
// 		return nil
// 	}
// 	return lt
// 	// j.lifetimeMapLock.RLock()
// 	// defer j.lifetimeMapLock.RUnlock()
// 	// return j.lifetimes[lId]
// }
//
// func (j *JournalImpl) fetchLifetime(ctx context.Context) (*Lifetime, error) {
// 	lId := ctx.Value("lifetimeId")
// 	j.log.Trace().Stack().Err(errors.New("")).Msgf("Cache miss on lifetime [%s]", lId)
// 	filter := bson.M{"_id": lId}
// 	res := j.col.FindOne(ctx, filter)
// 	if err := res.Err(); err != nil {
// 		if errors.Is(err, mongo.ErrNoDocuments) {
// 			return nil, nil
// 		}
// 		return nil, errors.WithStack(err)
// 	}
//
// 	lt := &Lifetime{}
// 	err := res.Decode(lt)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
//
// 	return lt, nil
// }
//
// func (j *JournalImpl) Add(lts ...*Lifetime) error {
// 	for _, lt := range lts {
// 		if lt.ServerId != j.serverId {
// 			return errors.WithStack(errors.ErrJournalServerMismatch)
// 		}
//
// 		err := upsertLifetime(lt, j.col)
// 		if err != nil {
// 			return err
// 		}
// 		j.cache.Set(lt.ID(), lt)
// 		// j.lifetimes[lt.ID()] = lt
// 	}
//
// 	return nil
// }
//
// func (j *JournalImpl) GetLifetimesSince(date time.Time) ([]*Lifetime, error) {
// 	return getLifetimesSince(date, j.col, j.serverId)
// }
//
// func (j *JournalImpl) Close() {
// 	close(j.eventStream)
// }
//
// func (j *JournalImpl) EventWorker() {
// 	for {
// 		fe, ok := <-j.eventStream
// 		if !ok {
// 			j.log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Event worker exiting...") })
// 			return
// 		}
//
// 		if fe == nil {
// 			j.log.Error().Msg("Got nil event in event stream...")
// 		} else {
// 			j.log.Trace().Func(func(e *zerolog.Event) {
// 				e.Msgf("Journal event worker got event starting with %s", fe.GetActions()[0].GetActionType())
// 			})
// 			j.flushCond.L.Lock()
//
// 			if err := j.handleFileEvent(fe); err != nil {
// 				j.log.Error().Stack().Err(err).Msg("")
// 			}
// 			close(fe.LoggedChan)
// 		}
//
// 		if len(j.eventStream) == 0 {
// 			j.flushCond.Broadcast()
// 		}
// 		j.log.Trace().Func(func(ze *zerolog.Event) {
// 			ze.Msgf("Journal worker finishing %s event at %s", fe.Actions[0].ActionType, fe.Actions[0].DestinationPath)
// 		})
// 		j.flushCond.L.Unlock()
// 	}
// }
//
// func (j *JournalImpl) handleFileEvent(event *FileEvent) error {
// 	if event.Logged.Load() {
// 		j.log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Skipping event already logged") })
// 		return nil
// 	}
//
// 	j.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Handling event with %d actions", len(event.GetActions())) })
//
// 	defer func() {
// 		e := recover()
// 		if e != nil {
// 			err, ok := e.(error)
// 			if !ok {
// 				j.log.Error().Msgf("%v", e)
// 			} else {
// 				j.log.Error().Stack().Err(err).Msg("")
// 			}
// 		}
// 	}()
//
// 	actions := event.GetActions()
// 	if len(actions) == 0 {
// 		return nil
// 	}
//
// 	slices.SortFunc(
// 		actions, func(a, b *FileAction) int {
// 			return a.GetTimestamp().Compare(b.GetTimestamp())
// 		},
// 	)
//
// 	// Ensrue all async tasks spawned by the hasher have finished before continuing
// 	if waitHasher, ok := event.hasher.(HashWaiter); ok {
// 		waitHasher.Wait()
// 	}
//
// 	var updated []*Lifetime
//
// 	for _, action := range actions {
// 		if action.GetFile() != nil {
// 			size := action.GetFile().Size()
// 			action.SetSize(size)
// 		}
//
// 		j.log.Trace().Func(func(e *zerolog.Event) {
// 			e.Msgf("Handling %s for [%s] [%s]", action.GetActionType(), action.GetRelevantPath(), action.GetLifetimeId())
// 		})
//
// 		actionType := action.GetActionType()
// 		switch actionType {
// 		case FileCreate, FileRestore:
// 			if action.Size == -1 {
// 				_, err := action.file.LoadStat()
// 				if err != nil {
// 					j.log.Error().Stack().Err(err).Msg("")
// 					continue
// 				}
// 				action.Size = action.file.Size()
// 			}
// 			newL, err := NewLifetime(action)
// 			if err != nil {
// 				return err
// 			}
//
// 			existing := j.Get(newL.ID())
// 			if existing != nil {
// 				return errors.Wrapf(errors.ErrLifetimeAlreadyExists, "Trying to add create action for [%s -- %s]", newL.GetLatestPath(), newL.ID())
// 			}
// 			updated = append(updated, newL)
// 		case FileDelete, FileMove, FileSizeChange:
// 			existing := j.Get(action.LifeId)
// 			if existing == nil {
// 				j.log.Error().Stack().Err(errors.WithStack(errors.ErrNoLifetime.WithArg(action.LifeId))).Msg("")
// 				continue
// 			}
// 			existing.Add(action)
//
// 			updated = append(updated, existing)
// 		default:
// 			return errors.Errorf("unknown file action type %s", actionType)
// 		}
// 	}
//
// 	err := j.Add(updated...)
// 	if err != nil {
// 		return err
// 	}
//
// 	event.Logged.Store(true)
// 	return nil
// }
//
// func getAllLifetimes(col *mongo.Collection, serverId string) ([]*Lifetime, error) {
// 	ret, err := col.Find(context.Background(), bson.M{"serverId": serverId})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var target []*Lifetime
// 	err = ret.All(context.Background(), &target)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return target, nil
// }
//
// func upsertLifetime(lt *Lifetime, col *mongo.Collection) error {
// 	filter := bson.M{"_id": lt.ID()}
// 	update := bson.M{"$set": lt}
// 	o := options.Update().SetUpsert(true)
// 	_, err := col.UpdateOne(context.Background(), filter, update, o)
//
// 	return err
// }
//
// func (j *JournalImpl) getChildrenAtTime(parentId string, time time.Time) ([]*FileAction, error) {
// 	pipe := bson.A{
// 		bson.D{{Key: "$match", Value: bson.D{{Key: "actions.parentId", Value: parentId}}}},
// 		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$actions"}}}},
// 		bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$actions"}}}},
// 		bson.D{{Key: "$match", Value: bson.D{{Key: "timestamp", Value: bson.D{{Key: "$lte", Value: time}}}}}},
// 		bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
// 		bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$lifeId"},
// 			{Key: "latest", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}},
// 		},
// 		},
// 		},
// 		bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$latest"}}}},
// 		bson.D{{Key: "$match", Value: bson.D{{Key: "parentId", Value: parentId}}}},
// 	}
//
// 	ret, err := j.col.Aggregate(context.Background(), pipe)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
//
// 	var target []*FileAction
// 	err = ret.All(context.Background(), &target)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
//
// 	return target, nil
// }
//
// func (j *JournalImpl) getActionsByPath(path fs.Filepath, noChildren bool) ([]*FileAction, error) {
// 	var pathMatch bson.A
// 	if noChildren {
// 		pathMatch = bson.A{
// 			bson.D{{Key: "actions.originPath", Value: path.ToPortable()}},
// 			bson.D{{Key: "actions.destinationPath", Value: path.ToPortable()}},
// 		}
// 	} else {
// 		pathMatch = bson.A{
// 			bson.D{{Key: "actions.originPath", Value: bson.D{{Key: "$regex", Value: path.ToPortable() + "[^/]*/?$"}}}},
// 			bson.D{{Key: "actions.destinationPath", Value: bson.D{{Key: "$regex", Value: path.ToPortable() + "[^/]*/?$"}}}},
// 		}
// 	}
//
// 	pipe := bson.A{
// 		bson.D{{Key: "$match", Value: bson.D{{Key: "serverId", Value: j.serverId}}}},
// 		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$actions"}}}},
// 		bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: pathMatch}}}},
// 		bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$actions"}}}},
// 		bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
// 	}
//
// 	ret, err := j.col.Aggregate(context.Background(), pipe)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
//
// 	var target []*FileAction
// 	err = ret.All(context.Background(), &target)
// 	if err != nil {
// 		return nil, errors.WithStack(err)
// 	}
//
// 	return target, nil
// }
