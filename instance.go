package weblens

import (
	"time"

	"github.com/ethrousseau/weblens/api/fileTree"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/ethrousseau/weblens/api/types"
)

type InstanceId string
type ServerRole string

const (
	InitServer   ServerRole = "init"
	CoreServer   ServerRole = "core"
	BackupServer ServerRole = "backup"
)

type WeblensInstance struct {
	Id   InstanceId `json:"id" bson:"_id"`
	Name string     `json:"name" bson:"name"`

	// Only applies to "core" server entries. This is the apiKey that remote server is using to connect to local,
	// if local is core. If local is backup, then this is the key being used to connect to remote core
	UsingKey types.WeblensApiKey `json:"-" bson:"usingKey"`

	// Core or BackupServer
	Role ServerRole `json:"role" bson:"serverRole"`

	// If this server info represents this local server
	IsThisServer bool `json:"-" bson:"isThisServer"`

	// Address of the remote server, only if the instance is a core.
	// Not set for any remotes/backups on core server, as it IS the core
	Address string `json:"coreAddress" bson:"coreAddress"`

	service *InstanceServiceImpl
}

func NewInstance(
	id InstanceId, name string, key types.WeblensApiKey, role ServerRole, isThisServer bool,
	address string,
) *WeblensInstance {
	return &WeblensInstance{
		Id:           id,
		Name:         name,
		UsingKey:     key,
		Role:         role,
		IsThisServer: isThisServer,
		Address: address,
	}
}

func (wi *WeblensInstance) Info() *WeblensInstance {
	return wi
}

func (wi *WeblensInstance) GetName() string {
	return wi.Name
}

func (wi *WeblensInstance) IsLocal() bool {
	return wi.IsThisServer
}

func (wi *WeblensInstance) IsCore() bool {
	return wi.Role == CoreServer
}

func (wi *WeblensInstance) ServerId() InstanceId {
	return wi.Id
}

func (wi *WeblensInstance) SetServerId(id InstanceId) {
	wi.Id = id
}

func (wi *WeblensInstance) GetUsingKey() types.WeblensApiKey {
	return wi.UsingKey
}

func (wi *WeblensInstance) SetUsingKey(key types.WeblensApiKey) {
	wi.UsingKey = key
}

func (wi *WeblensInstance) ServerRole() ServerRole {
	return wi.Role
}

func (wi *WeblensInstance) GetRole() ServerRole {
	return wi.Role
	// if si == nil {
	// 	return false
	// }
	// return si.Role == Core
}

func (wi *WeblensInstance) GetAddress() (string, error) {
	if wi.Role != CoreServer {
		return "", errors.WithStack(errors.New("Cannot get address of non-core instance"))
	}
	return wi.Address, nil
}

func (wi *WeblensInstance) SetAddress(address string) error {
	if wi.Role != CoreServer {
		return errors.WithStack(errors.New("Cannot set address of non-core instance"))
	}
	wi.Address = address
	return nil
}

type InstanceService interface {
	Size() int

	Get(id InstanceId) *WeblensInstance
	Add(instance *WeblensInstance) error
	Del(id InstanceId) error
	GetLocal() *WeblensInstance
	GetCore() *WeblensInstance
	GetRemotes() []*WeblensInstance
	GenerateNewId(name string) InstanceId
	InitCore(*WeblensInstance) error
	InitBackup(name, coreAddr string, key WeblensApiKey, store ProxyStore) error
	IsLocalLoaded() bool
	AddLoading(loadingKey string)
	RemoveLoading(loadingKey string)
}

type WeblensApiKey string
type ApiKeyInfo struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	Key         WeblensApiKey      `bson:"key" json:"key"`
	Owner       Username           `bson:"owner" json:"owner"`
	CreatedTime time.Time          `bson:"createdTime" json:"createdTime"`
	RemoteUsing InstanceId         `bson:"remoteUsing" json:"remoteUsing"`
}

type AccessService interface {
	WeblensService[WeblensApiKey, ApiKeyInfo, AccessStore]
	Del(key WeblensApiKey) error
	GenerateApiKey() (ApiKeyInfo, error)
	GetApiKeyById(key WeblensApiKey) (ApiKeyInfo, error)
	GetAllKeys() ([]ApiKeyInfo, error)
	CanUserAccessFile(user *User, file *fileTree.WeblensFile) bool
}
