package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccessService interface {
	WeblensService[WeblensApiKey, ApiKeyInfo, AccessStore]
	GenerateApiKey(AccessMeta) (ApiKeyInfo, error)
	GetAllKeys(AccessMeta) ([]ApiKeyInfo, error)
}

type ApiKeyInfo struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	Key         WeblensApiKey      `bson:"key" json:"key"`
	Owner       Username           `bson:"owner" json:"owner"`
	CreatedTime time.Time          `bson:"createdTime" json:"createdTime"`
	RemoteUsing InstanceId         `bson:"remoteUsing" json:"remoteUsing"`
}

type WeblensApiKey string
type ServerRole string
type InstanceId string

const (
	Initialization ServerRole = "init"

	Core   ServerRole = "core"
	Backup ServerRole = "backup"
)

type Instance interface {
	ServerId() InstanceId
	SetServerId(InstanceId)
	GetName() string
	GetUsingKey() WeblensApiKey
	ServerRole() ServerRole
	GetCoreAddress() (string, error)
	Info() Instance
	IsLocal() bool
	// SetUserCount(int)
}

type InstanceService interface {
	WeblensService[InstanceId, Instance, InstanceStore]
	GetLocal() Instance
	GetRemotes() []Instance
	GenerateNewId(name string) InstanceId
	InitCore(Instance) error
	InitBackup(name, coreAddr string, key WeblensApiKey, store ProxyStore) error
	IsLocalLoaded() bool
	AddLoading(loadingKey string)
	RemoveLoading(loadingKey string)
}

var ErrAlreadyCore = NewWeblensError("core server cannot have a remote core")
