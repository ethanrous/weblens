package weblens

import (
	"context"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type InstanceServiceImpl struct {
	instanceMap     map[InstanceId]*WeblensInstance
	instanceMapLock sync.RWMutex
	local           *WeblensInstance
	core            *WeblensInstance
	localLoading    map[string]bool

	collection    *mongo.Collection
	accessService AccessService
}

func NewInstanceService(accessService AccessService, col *mongo.Collection) *InstanceServiceImpl {
	return &InstanceServiceImpl{
		instanceMap:   make(map[InstanceId]*WeblensInstance),
		localLoading:  map[string]bool{"all": true},
		accessService: accessService,
		collection:    col,
	}
}

func (is *InstanceServiceImpl) Init() error {
	ret, err := is.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return errors.Wrap(err, "Failed to find instances")
	}

	var servers []*WeblensInstance
	err = ret.All(context.Background(), &servers)
	if err != nil {
		return errors.Wrap(err, "Failed to decode instances")
	}

	is.instanceMapLock.Lock()
	defer is.instanceMapLock.Unlock()
	for _, server := range servers {
		if server.IsLocal() {
			is.local = server
			continue
		}
		if server.IsCore() {
			is.core = server
		}
		is.instanceMap[server.ServerId()] = server
	}

	if is.local == nil {
		is.local = &WeblensInstance{IsThisServer: true, Role: InitServer}
	}

	return nil
}

func (is *InstanceServiceImpl) Add(i *WeblensInstance) error {
	if i.ServerId() == "" && !i.IsLocal() {
		return werror.WErrMsg("Remote server must have specified id")
	} else if i.ServerId() == "" {
		i.(*WeblensInstance).Id = is.GenerateNewId(i.GetName())
	}

	err := types.SERV.StoreService.NewServer(i)
	if err != nil {
		return err
	}

	is.accessService.
	err = types.SERV.StoreService.SetRemoteUsing(i.GetUsingKey(), i.ServerId())
	if err != nil {
		return err
	}

	is.instanceMapLock.Lock()
	defer is.instanceMapLock.Unlock()
	is.instanceMap[i.ServerId()] = i

	if i.IsLocal() {
		is.local = i
	}

	if i.IsCore() {
		is.core = i
	}

	return nil
}

// TODO - this belongs in the access service, not instance
// func (is *InstanceServiceImpl) SetRemoteUsingKey(instance *WeblensInstance, key types.WeblensApiKey) error {
// 	filter := bson.M{"key": key}
// 	update := bson.M{"$set": bson.M{"remoteUsing": instance.ServerId()}}
// 	_, err := is.collection.UpdateOne(context.Background(), filter, update)
// 	if err != nil {
// 		return err
// 	}
//
// 	instance.UsingKey = key
// }

func (is *InstanceServiceImpl) Get(iId InstanceId) *WeblensInstance {
	is.instanceMapLock.RLock()
	defer is.instanceMapLock.RUnlock()
	return is.instanceMap[iId]
}

func (is *InstanceServiceImpl) GetLocal() *WeblensInstance {
	return is.local
}

func (is *InstanceServiceImpl) GetCore() *WeblensInstance {
	return is.core
}

func (is *InstanceServiceImpl) Del(iId InstanceId) error {
	_, err := is.collection.DeleteOne(context.Background(), bson.M{"_id": iId})
	if err != nil {
		return err
	}

	is.instanceMapLock.Lock()
	defer is.instanceMapLock.Unlock()
	delete(is.instanceMap, iId)

	return nil
}

func (is *InstanceServiceImpl) Size() int {
	return len(is.instanceMap)
}

func (is *InstanceServiceImpl) IsLocalLoaded() bool {
	is.instanceMapLock.RLock()
	defer is.instanceMapLock.RUnlock()
	return len(is.localLoading) == 0
}

func (is *InstanceServiceImpl) AddLoading(loadingKey string) {
	is.instanceMapLock.Lock()
	defer is.instanceMapLock.Unlock()
	is.localLoading[loadingKey] = true
}

func (is *InstanceServiceImpl) RemoveLoading(loadingKey string) {
	is.instanceMapLock.Lock()
	delete(is.localLoading, loadingKey)
	is.instanceMapLock.Unlock()

	if is.IsLocalLoaded() {
		err := types.SERV.RestartRouter()
		if err != nil {
			wlog.ErrTrace(err)
		}
		types.SERV.Caster.PushWeblensEvent("weblens_loaded")
	}
}

func (is *InstanceServiceImpl) GenerateNewId(name string) InstanceId {
	return InstanceId(internal.GlobbyHash(12, name, time.Now().String()))
}

func (is *InstanceServiceImpl) GetRemotes() []*WeblensInstance {
	return internal.MapToValues(is.instanceMap)
}

func (is *InstanceServiceImpl) InitCore(instance *WeblensInstance) error {
	// u := us.Get(username)
	//
	// // Init with existing u
	// if u != nil {
	// 	if !u.CheckLogin(password) {
	// 		return types.ErrUserNotAuthenticated
	// 	} else if !u.IsAdmin() {
	// 		return types.NewWeblensError("TODO")
	// 		// err := u.SetOwner()
	// 		// if err != nil {
	// 		// 	return err
	// 		// }
	// 	}
	//
	// } else { // create new user, this will be the case 99% of the time
	// 	err := user.New(username, password, true, true, ft)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	//
	// srvId := InstanceId(util.GlobbyHash(12, name, time.Now().String()))
	// wi.Id = srvId
	// wi.Name = name
	// wi.IsThisServer = true
	// wi.Role = Core
	//
	// err := wi.db.NewServer(srvId, name, true, Core)
	// if err != nil {
	// 	return err
	// }
	//
	// return nil

	return werror.NotImplemented("instance InitCore")
}

func (is *InstanceServiceImpl) InitBackup(
	name, coreAddr string, key types.WeblensApiKey, store types.ProxyStore,
) error {
	is.store = store

	srvId := InstanceId(internal.GlobbyHash(12, name, time.Now().String()))
	thisServer := NewInstance(srvId, name, "", BackupServer, true, "")
	core := NewInstance("", "", key, CoreServer, false, coreAddr)
	core, err := is.store.AttachToCore(thisServer, core)
	if err != nil {
		return err
	}

	err = is.Add(thisServer)
	if err != nil {
		return err
	}

	err = is.Add(core)
	if err != nil {
		return err
	}

	return nil
}
