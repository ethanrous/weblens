package models

import (
	"errors"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal/werror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InstanceId string
type ServerRole string

const (
	InitServer   ServerRole = "init"
	CoreServer   ServerRole = "core"
	BackupServer ServerRole = "backup"
)

type Instance struct {
	Id   InstanceId `json:"id" bson:"_id"`
	Name string     `json:"name" bson:"name"`

	// Only applies to "core" server entries. This is the apiKey that remote server is using to connect to local,
	// if local is core. If local is backup, then this is the key being used to connect to remote core
	UsingKey WeblensApiKey `json:"-" bson:"usingKey"`

	// Core or BackupServer
	Role ServerRole `json:"role" bson:"serverRole"`

	// If this server info represents this local server
	IsThisServer bool `json:"-" bson:"isThisServer"`

	// Address of the remote server, only if the instance is a core.
	// Not set for any remotes/backups on core server, as it IS the core
	Address string `json:"coreAddress" bson:"coreAddress"`

	service InstanceService
}

func NewInstance(
	id InstanceId, name string, key WeblensApiKey, role ServerRole, isThisServer bool,
	address string,
) *Instance {
	return &Instance{
		Id:           id,
		Name:         name,
		UsingKey:     key,
		Role:         role,
		IsThisServer: isThisServer,
		Address: address,
	}
}

func (wi *Instance) Info() *Instance {
	return wi
}

func (wi *Instance) GetName() string {
	return wi.Name
}

func (wi *Instance) IsLocal() bool {
	return wi.IsThisServer
}

func (wi *Instance) IsCore() bool {
	return wi.Role == CoreServer
}

func (wi *Instance) ServerId() InstanceId {
	return wi.Id
}

func (wi *Instance) SetServerId(id InstanceId) {
	wi.Id = id
}

func (wi *Instance) GetUsingKey() WeblensApiKey {
	return wi.UsingKey
}

func (wi *Instance) SetUsingKey(key WeblensApiKey) {
	wi.UsingKey = key
}

func (wi *Instance) ServerRole() ServerRole {
	return wi.Role
}

func (wi *Instance) GetRole() ServerRole {
	return wi.Role
	// if si == nil {
	// 	return false
	// }
	// return si.Role == Core
}

func (wi *Instance) GetAddress() (string, error) {
	if wi.Role != CoreServer {
		return "", werror.WithStack(errors.New("Cannot get address of non-core instance"))
	}
	return wi.Address, nil
}

func (wi *Instance) SetAddress(address string) error {
	if wi.Role != CoreServer {
		return werror.WithStack(errors.New("Cannot set address of non-core instance"))
	}
	wi.Address = address
	return nil
}

func (wi *Instance) SocketType() string {
	return "serverClient"
}

type ServerInfo struct {
	Id   InstanceId `json:"id"`
	Name string     `json:"name"`

	// Only applies to "core" server entries. This is the apiKey that remote server is using to connect to local,
	// if local is core. If local is backup, then this is the key being used to connect to remote core
	UsingKey WeblensApiKey `json:"-"`

	// Core or Backup
	Role ServerRole `json:"role"`

	// If this server info represents this local server
	IsThisServer bool `json:"-"`

	// Address of the remote server, only if the instance is a core.
	// Not set for any remotes/backups on core server, as it IS the core
	Address string `json:"coreAddress"`

	Online bool `json:"online"`
}

type InstanceService interface {
	Size() int

	Get(id InstanceId) *Instance
	Add(instance *Instance) error
	Del(id InstanceId) error
	GetLocal() *Instance
	GetCore() *Instance
	GetRemotes() []*Instance
	GenerateNewId(name string) InstanceId
	InitCore(*Instance) error
	InitBackup(name, coreAddr string, key WeblensApiKey) error
	IsLocalLoaded() bool
	AddLoading(loadingKey string)
	RemoveLoading(loadingKey string) (doneLoading bool)
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
	Init() error
	Get(key WeblensApiKey) (ApiKeyInfo, error)
	Del(key WeblensApiKey) error
	GenerateApiKey(creator *User) (ApiKeyInfo, error)
	CanUserAccessFile(user *User, file *fileTree.WeblensFileImpl, share *FileShare) bool
	CanUserModifyShare(user *User, share Share) bool
	CanUserAccessAlbum(user *User, album *Album, share *AlbumShare) bool
	
	GetAllKeys(accessor *User) ([]ApiKeyInfo, error)
	SetKeyUsedBy(key WeblensApiKey, server *Instance) error
}
