package models

import (
	"sync"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InstanceId = string
type ServerRole string

const (
	InitServerRole    ServerRole = "init"
	CoreServerRole    ServerRole = "core"
	BackupServerRole  ServerRole = "backup"
	RestoreServerRole ServerRole = "restore"
)

// An "Instance" is a single Weblens server.
// For clarity: Core vs Backup are absolute server roles, and each server
// will fit into one of these categories once initialized. Local vs Remote
// are RELATIVE terms, meaning one core servers "remote" is the backup
// servers "local".
type Instance struct {

	// The ID of the server that is shared between all servers
	Id   InstanceId `json:"id" bson:"instanceId"`
	Name string     `json:"name" bson:"name"`

	// Only applies to "core" server entries. This is the apiKey that remote server is using to connect to local,
	// if local is core. If local is backup, then this is the key being used to connect to remote core
	UsingKey WeblensApiKey `json:"-" bson:"usingKey"`

	// Core or BackupServer
	Role ServerRole `json:"role" bson:"serverRole"`

	// Address of the remote server, only if the instance is a core.
	// Not set for any remotes/backups on core server, as it IS the core
	Address string `json:"coreAddress" bson:"coreAddress"`

	// The ID of the server in which this remote instance is in reference from
	CreatedBy InstanceId `json:"createdBy" bson:"createdBy"`

	// Role the server is currently reporting. This is used to determine if the server is online (and functional) or not
	reportedRole ServerRole

	// The time of the latest backup, in milliseconds since epoch
	LastBackup int64 `json:"lastBackup" bson:"lastBackup"`

	updateMu sync.RWMutex
	// The ID of the server in the local database
	DbId primitive.ObjectID `json:"-" bson:"_id"`

	// If this server info represents this local server
	IsThisServer bool `json:"-" bson:"isThisServer"`
}

func NewInstance(
	id InstanceId, name string, key WeblensApiKey, role ServerRole, isThisServer bool,
	address string, createdBy InstanceId,
) *Instance {
	if id == "" {
		id = primitive.NewObjectID().Hex()
	}
	return &Instance{
		Id:           id,
		Name:         name,
		UsingKey:     key,
		Role:         role,
		IsThisServer: isThisServer,
		Address:      address,
		CreatedBy:    createdBy,
	}
}

func (wi *Instance) GetName() string {
	return wi.Name
}

func (wi *Instance) IsLocal() bool {
	return wi.IsThisServer
}

func (wi *Instance) IsCore() bool {
	return wi.Role == CoreServerRole
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

func (wi *Instance) GetRole() ServerRole {
	wi.updateMu.RLock()
	defer wi.updateMu.RUnlock()
	return wi.Role
}

func (wi *Instance) GetReportedRole() ServerRole {
	wi.updateMu.RLock()
	defer wi.updateMu.RUnlock()
	return wi.reportedRole
}

func (wi *Instance) SetRole(role ServerRole) {
	wi.updateMu.Lock()
	defer wi.updateMu.Unlock()
	wi.Role = role
}

func (wi *Instance) SetReportedRole(role ServerRole) {
	wi.updateMu.Lock()
	defer wi.updateMu.Unlock()
	wi.reportedRole = role
	log.TraceCaller(1, "SetReportedRole %s", role)
}

func (wi *Instance) GetAddress() (string, error) {
	if wi.Role != CoreServerRole {
		return "", werror.WithStack(werror.Errorf("Cannot get address of non-core instance"))
	}
	return wi.Address, nil
}

func (wi *Instance) SetAddress(address string) error {
	if wi.Role != CoreServerRole {
		return werror.WithStack(werror.Errorf("Cannot set address of non-core instance"))
	}
	wi.Address = address
	return nil
}

func (wi *Instance) SocketType() string {
	return "serverClient"
}

type InstanceService interface {
	Size() int

	Get(dbId string) *Instance
	GetAllByOriginServer(serverId InstanceId) []*Instance
	GetByInstanceId(serverId InstanceId) *Instance
	Add(instance *Instance) error
	Del(dbId primitive.ObjectID) error
	GetLocal() *Instance
	GetCores() []*Instance
	GetRemotes() []*Instance
	InitCore(serverName string) error
	InitBackup(name, coreAddr string, key WeblensApiKey) error
	SetLastBackup(id InstanceId, time time.Time) error
	AttachRemoteCore(coreAddr string, key string) (*Instance, error)
	ResetAll() error
}

type WeblensApiKey = string

type ApiKey struct {
	CreatedTime time.Time          `bson:"createdTime"`
	Name        string             `bson:"name"`
	LastUsed    time.Time          `bson:"lastUsed"`
	Key         WeblensApiKey      `bson:"key"`
	Owner       Username           `bson:"owner"`
	RemoteUsing InstanceId         `bson:"remoteUsing"`
	CreatedBy   InstanceId         `bson:"createdBy"`
	Id          primitive.ObjectID `bson:"_id"`
}

type AccessService interface {
	GenerateJwtToken(user *User) (token string, expires time.Time, err error)
	GetApiKey(key WeblensApiKey) (ApiKey, error)
	AddApiKey(key ApiKey) error
	GetUserFromToken(token string) (*User, error)
	DeleteApiKey(key WeblensApiKey) error
	GenerateApiKey(creator *User, local *Instance, keyName string) (ApiKey, error)
	CanUserAccessFile(user *User, file *fileTree.WeblensFileImpl, share *FileShare) bool
	CanUserModifyShare(user *User, share Share) bool
	CanUserAccessAlbum(user *User, album *Album, share *AlbumShare) bool

	GetAllKeys(accessor *User) ([]ApiKey, error)
	GetAllKeysByServer(accessor *User, serverId InstanceId) ([]ApiKey, error)
	SetKeyUsedBy(key WeblensApiKey, server *Instance) error
}
