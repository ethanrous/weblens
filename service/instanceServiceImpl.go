package service

import (
	"context"
	"net/http"
	"sync"

	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/service/proxy"
	"go.mongodb.org/mongo-driver/bson"
)

var _ models.InstanceService = (*InstanceServiceImpl)(nil)

type InstanceServiceImpl struct {
	instanceMap     map[models.InstanceId]*models.Instance
	instanceMapLock sync.RWMutex
	local           *models.Instance
	core            *models.Instance

	col database.MongoCollection
}

func NewInstanceService(col database.MongoCollection) (*InstanceServiceImpl, error) {
	is := &InstanceServiceImpl{
		instanceMap: make(map[models.InstanceId]*models.Instance),
		col:         col,
	}

	ret, err := is.col.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, werror.WithStack(err)
	}

	var servers []*models.Instance
	err = ret.All(context.Background(), &servers)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	is.instanceMapLock.Lock()
	defer is.instanceMapLock.Unlock()
	for _, server := range servers {
		if server.IsLocal() {
			is.local = server
			continue
		}
		if server.IsCore() {
			if is.core != nil {
				return nil, werror.WithStack(werror.ErrDuplicateCoreServer)
			}
			is.core = server
		}
		is.instanceMap[server.ServerId()] = server
	}

	if is.local == nil {
		is.local = models.NewInstance("", "", "", models.InitServer, true, "")
	}

	return is, nil
}

func (is *InstanceServiceImpl) Add(i *models.Instance) error {
	// Validate
	if i.ServerId() == "" {
		return werror.WithStack(werror.ErrNoServerId)
	} else if !i.IsLocal() && i.GetUsingKey() == "" {
		// The key and the address are ALWAYS on the remote
		return werror.WithStack(werror.ErrNoServerKey)
	} else if i.GetName() == "" {
		return werror.WithStack(werror.ErrNoServerName)
	} else if i.IsLocal() {
		return werror.WithStack(werror.ErrDuplicateLocalServer)
	} else if i.IsCore() && i.Address == "" {
		// The key and the address are ALWAYS on the remote
		return werror.WithStack(werror.ErrNoCoreAddress)
	}

	_, err := is.col.InsertOne(context.Background(), i)
	if err != nil {
		return werror.WithStack(err)
	}

	// err = is.accessService.SetKeyUsedBy(i.GetUsingKey(), i)
	// if err != nil {
	// 	return err
	// }

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

func (is *InstanceServiceImpl) GetRemotes() []*models.Instance {
	return internal.MapToValues(is.instanceMap)
}

func (is *InstanceServiceImpl) InitCore(serverName string) error {
	local := is.GetLocal()
	if local == nil {
		return werror.WithStack(werror.ErrNoLocal)
	}

	local.Name = serverName
	local.Role = models.CoreServer

	_, err := is.col.InsertOne(context.Background(), local)
	if err != nil {
		// Revert name and role if db write fails
		local.Name = ""
		local.Role = models.InitServer
		return werror.WithStack(err)
	}

	is.core = local

	return nil
}

func (is *InstanceServiceImpl) InitBackup(
	name, coreAddr string, key models.WeblensApiKey,
) error {
	local := is.GetLocal()
	if local == nil {
		return werror.WithStack(werror.ErrNoLocal)
	}

	local.Name = name
	local.SetRole(models.BackupServer)

	core := models.NewInstance("", "", key, models.CoreServer, false, coreAddr)
	// NewInstance will generate an Id if one is not given. We want to fill the id from what the core
	// server reports it is, not make a new one
	core.Id = ""

	type newServerBody struct {
		Id       models.InstanceId    `json:"serverId"`
		Role     models.ServerRole    `json:"role"`
		Name     string               `json:"name"`
		UsingKey models.WeblensApiKey `json:"usingKey"`
	}
	body := newServerBody{Id: local.ServerId(), Role: models.BackupServer, Name: local.GetName(), UsingKey: key}

	r := proxy.NewRequest(core, http.MethodPost, "/remote").WithBody(body)
	newCore, err := proxy.CallHomeStruct[*models.Instance](r)
	if err != nil {
		return err
	}

	newCore.UsingKey = key
	newCore.Address = coreAddr

	_, err = is.col.InsertOne(context.Background(), local)
	if err != nil {
		// Revert name and role if db write fails
		local.Name = ""
		local.Role = models.InitServer
		return werror.WithStack(err)
	}

	err = is.Add(newCore)
	if err != nil {
		return err
	}

	return nil
}

// ResetAll will clear all known servers, including the local one,
// and will reset this server to initialization mode.
func (is *InstanceServiceImpl) ResetAll() error {
	// Preserve local server id
	localId := is.GetLocal().ServerId()

	is.instanceMapLock.Lock()
	defer is.instanceMapLock.Unlock()
	_, err := is.col.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		return werror.WithStack(err)
	}

	is.instanceMap = make(map[models.InstanceId]*models.Instance)
	is.instanceMapLock.Unlock()

	newLocal := models.NewInstance(localId, "", "", models.InitServer, true, "")

	_, err = is.col.InsertOne(context.Background(), newLocal)
	if err != nil {
		return err
	}

	is.core = nil
	is.local = newLocal

	is.instanceMap[localId] = newLocal

	return nil
}
