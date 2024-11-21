package service

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/ethanrous/weblens/database"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/service/proxy"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ models.InstanceService = (*InstanceServiceImpl)(nil)

type InstanceServiceImpl struct {
	instanceMap     map[string]*models.Instance
	instanceMapLock sync.RWMutex
	local           *models.Instance

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

	// Must know local server before starting
	for _, server := range servers {
		if !server.IsLocal() {
			continue
		}
		is.local = server
		break
	}

	is.instanceMapLock.Lock()
	defer is.instanceMapLock.Unlock()
	if is.local == nil {
		is.local = models.NewInstance("", "", "", models.InitServerRole, true, "", "")
		is.local.CreatedBy = is.local.Id
	} else {
		for _, server := range servers {
			if server.IsLocal() {
				continue
			}
			_, ok := is.instanceMap[server.DbId.Hex()]
			if ok && server.CreatedBy != is.local.Id {
				continue
			}

			log.Trace.Func(func(l log.Logger) {l.Printf("Adding server [%s] (created by [%s]) to instance map", server.Id, server.CreatedBy)})
			is.instanceMap[server.DbId.Hex()] = server
		}
	}

	return is, nil
}

func (is *InstanceServiceImpl) Add(i *models.Instance) error {
	// Check if the instance was created on this server or not. Should only
	// be false on backup servers looking to back up the database on the core
	createdHere := i.CreatedBy == is.local.ServerId()

	// Validate
	if i.ServerId() == "" {
		return werror.WithStack(werror.ErrNoServerId)
	} else if i.CreatedBy == "" {
		return werror.WithStack(werror.ErrNoCreator)
	} else if !i.IsLocal() && i.GetUsingKey() == "" && createdHere {
		// The key and the address are ALWAYS on the remote... if it was createdHere
		return werror.WithStack(werror.ErrNoServerKey)
	} else if i.GetName() == "" {
		return werror.WithStack(werror.ErrNoServerName)
	} else if i.IsCore() && i.Address == "" && createdHere {
		// The key and the address are ALWAYS on the remote... if it was createdHere
		return werror.WithStack(werror.ErrNoCoreAddress)
	} else if !i.DbId.IsZero() {
		return werror.Errorf("instance already has a local id")
	} else if i.IsLocal() && is.local != nil && !is.local.DbId.IsZero() {
		return werror.WithStack(werror.ErrDuplicateLocalServer)
	}

	// Give this server a locally unique id
	i.DbId = primitive.NewObjectID()

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
	is.instanceMap[i.DbId.Hex()] = i

	if i.IsLocal() {
		is.local = i
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

func (is *InstanceServiceImpl) Get(dbId string) *models.Instance {
	is.instanceMapLock.RLock()
	defer is.instanceMapLock.RUnlock()
	return is.instanceMap[dbId]
}

func (is *InstanceServiceImpl) GetAllByOriginServer(originId models.InstanceId) []*models.Instance {
	is.instanceMapLock.RLock()
	defer is.instanceMapLock.RUnlock()
	ret := []*models.Instance{}
	for _, instance := range is.instanceMap {
		if instance.CreatedBy == originId {
			ret = append(ret, instance)
		}
	}

	return ret
}

func (is *InstanceServiceImpl) GetByInstanceId(serverId models.InstanceId) *models.Instance {
	is.instanceMapLock.RLock()
	defer is.instanceMapLock.RUnlock()

	if serverId == is.local.ServerId() {
		return is.local
	}

	for _, instance := range is.instanceMap {
		if instance.Id == serverId && instance.CreatedBy == is.local.ServerId() {
			return instance
		}
	}

	return nil
}

func (is *InstanceServiceImpl) GetLocal() *models.Instance {
	return is.local
}

func (is *InstanceServiceImpl) GetCores() []*models.Instance {
	if is.local == nil || is.local.IsCore() {
		return nil
	}

	var cores []*models.Instance
	is.instanceMapLock.RLock()
	defer is.instanceMapLock.RUnlock()

	for _, i := range is.instanceMap {
		log.Debug.Printf("Checking instance [%s] created by [%s]", i.ServerId(), i.CreatedBy)
		if i.IsCore() && i.CreatedBy == is.local.ServerId() {
			cores = append(cores, i)
		}
	}

	return cores
}

func (is *InstanceServiceImpl) Del(dbId primitive.ObjectID) error {
	_, err := is.col.DeleteOne(context.Background(), bson.M{"_id": dbId})
	if err != nil {
		return err
	}

	is.instanceMapLock.Lock()
	defer is.instanceMapLock.Unlock()
	delete(is.instanceMap, dbId.Hex())

	return nil
}

func (is *InstanceServiceImpl) Size() int {
	return len(is.instanceMap)
}

// GetRemotes returns all instances that are not the local server
func (is *InstanceServiceImpl) GetRemotes() []*models.Instance {
	return internal.MapToValues(is.instanceMap)
}

func (is *InstanceServiceImpl) InitCore(serverName string) error {
	local := is.GetLocal()
	if local == nil {
		return werror.WithStack(werror.ErrNoLocal)
	}

	local.Name = serverName
	local.SetRole(models.CoreServerRole)
	local.DbId = primitive.NewObjectID()

	_, err := is.col.InsertOne(context.Background(), local)
	if err != nil {
		// Revert name and role if db write fails
		local.Name = ""
		local.Role = models.InitServerRole
		return werror.WithStack(err)
	}

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
	local.SetRole(models.BackupServerRole)

	_, err := is.AttachRemoteCore(coreAddr, key)
	if err != nil {
		// Revert name and role if db write fails
		local.Name = ""
		local.Role = models.InitServerRole
		return err
	}

	err = is.Add(local)

	return err
}

func (is *InstanceServiceImpl) AttachRemoteCore(coreAddr string, key string) (*models.Instance, error) {
	local := is.GetLocal()
	core := models.NewInstance("", "", key, models.CoreServerRole, false, coreAddr, local.ServerId())
	// NewInstance will generate an Id if one is not given. We want to fill the id from what the core
	// server reports it is, not make a new one
	core.Id = ""

	type newServerBody struct {
		Id       models.InstanceId    `json:"serverId"`
		Role     models.ServerRole    `json:"role"`
		Name     string               `json:"name"`
		UsingKey models.WeblensApiKey `json:"usingKey"`
	}

	body := newServerBody{Id: local.ServerId(), Role: models.BackupServerRole, Name: local.GetName(), UsingKey: key}
	r := proxy.NewCoreRequest(core, http.MethodPost, "/servers").WithBody(body)
	newCore, err := proxy.CallHomeStruct[*models.Instance](r)
	if err != nil {
		return nil, err
	}

	newCore.UsingKey = key
	newCore.Address = coreAddr
	newCore.CreatedBy = local.ServerId()
	// newCore.DbId = primitive.NewObjectID()

	err = is.Add(newCore)
	if err != nil {
		return nil, err
	}

	return newCore, nil
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

	newLocal := models.NewInstance(localId, "", "", models.InitServerRole, true, "", localId)

	_, err = is.col.InsertOne(context.Background(), newLocal)
	if err != nil {
		return err
	}

	is.local = newLocal

	is.instanceMap[localId] = newLocal

	return nil
}

func (is *InstanceServiceImpl) SetLastBackup(id models.InstanceId, lastBackup time.Time) error {
	instance := is.GetByInstanceId(id)
	if instance == nil {
		return werror.WithStack(werror.ErrNoInstance)
	}

	lastBackupMillis := lastBackup.UnixMilli()

	_, err := is.col.UpdateOne(
		context.Background(), bson.M{"_id": instance.DbId}, bson.M{"$set": bson.M{"lastBackup": lastBackupMillis}},
	)
	if err != nil {
		return werror.WithStack(err)
	}

	instance.LastBackup = lastBackupMillis

	return nil
}
