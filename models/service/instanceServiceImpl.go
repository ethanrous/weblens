package service

import (
	"context"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ models.InstanceService = (*InstanceServiceImpl)(nil)

type InstanceServiceImpl struct {
	instanceMap map[models.InstanceId]*models.Instance
	instanceMapLock sync.RWMutex
	local       *models.Instance
	core        *models.Instance
	localLoading    map[string]bool

	col           *mongo.Collection
	accessService models.AccessService
}

func NewInstanceService(accessService models.AccessService, col *mongo.Collection) *InstanceServiceImpl {
	return &InstanceServiceImpl{
		instanceMap: make(map[models.InstanceId]*models.Instance),
		localLoading:  map[string]bool{"all": true},
		accessService: accessService,
		col:         col,
	}
}

func (is *InstanceServiceImpl) Init() error {
	ret, err := is.col.Find(context.Background(), bson.M{})
	if err != nil {
		return errors.Wrap(err, "Failed to find instances")
	}

	var servers []*models.Instance
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
		is.local = &models.Instance{IsThisServer: true, Role: models.InitServer}
	}

	return nil
}

func (is *InstanceServiceImpl) Add(i *models.Instance) error {
	if i.ServerId() == "" && !i.IsLocal() {
		return werror.New("Remote server must have specified id")
	} else if i.ServerId() == "" {
		i.Id = is.GenerateNewId(i.GetName())
	}

	_, err := is.col.InsertOne(context.Background(), i)
	if err != nil {
		return err
	}

	err = is.accessService.SetKeyUsedBy(i.GetUsingKey(), i)
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

func (is *InstanceServiceImpl) Get(iId models.InstanceId) *models.Instance {
	is.instanceMapLock.RLock()
	defer is.instanceMapLock.RUnlock()
	return is.instanceMap[iId]
}

func (is *InstanceServiceImpl) GetLocal() *models.Instance {
	return is.local
}

func (is *InstanceServiceImpl) GetCore() *models.Instance {
	return is.core
}

func (is *InstanceServiceImpl) Del(iId models.InstanceId) error {
	_, err := is.col.DeleteOne(context.Background(), bson.M{"_id": iId})
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

func (is *InstanceServiceImpl) RemoveLoading(loadingKey string) (doneLoading bool) {
	is.instanceMapLock.Lock()
	delete(is.localLoading, loadingKey)
	is.instanceMapLock.Unlock()

	return is.IsLocalLoaded()
}

func (is *InstanceServiceImpl) GenerateNewId(name string) models.InstanceId {
	return models.InstanceId(internal.GlobbyHash(12, name, time.Now().String()))
}

func (is *InstanceServiceImpl) GetRemotes() []*models.Instance {
	return internal.MapToValues(is.instanceMap)
}

func (is *InstanceServiceImpl) InitCore(instance *models.Instance) error {
	// u := us.Get(username)
	//
	// // Init with existing u
	// if u != nil {
	// 	if !u.CheckLogin(password) {
	// 		return types.ErrUserNotAuthenticated
	// 	} else if !u.IsAdmin() {
	// 		return types.New("TODO")
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
	name, coreAddr string, key models.WeblensApiKey,
) error {

	return werror.NotImplemented("instance InitBackup")
	// srvId := InstanceId(internal.GlobbyHash(12, name, time.Now().String()))
	// thisServer := NewInstance(srvId, name, "", BackupServer, true, "")
	// core := NewInstance("", "", key, CoreServer, false, coreAddr)
	// core, err := is.store.AttachToCore(thisServer, core)
	// if err != nil {
	// 	return err
	// }
	//
	// err = is.Add(thisServer)
	// if err != nil {
	// 	return err
	// }
	//
	// err = is.Add(core)
	// if err != nil {
	// 	return err
	// }
	//
	// return nil
}
