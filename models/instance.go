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
type ServerRole = string

const (
	InitServer    ServerRole = "init"
	CoreServer    ServerRole = "core"
	BackupServer  ServerRole = "backup"
	RestoreServer ServerRole = "restore"
)

// An "Instance" is a single Weblens server.
// For clarity: Core vs Backup are absolute server roles, and each server
// will fit into one of these categories once initialized. Local vs Remote
// are RELATIVE terms, meaning one core servers "remote" is the backup
// servers "local".
type Instance struct {
	// The ID of the server in the local database
	DbId primitive.ObjectID `json:"-" bson:"_id"`

	// The ID of the server that is shared between all servers
	Id   InstanceId `json:"id" bson:"instanceId"`
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

	// The ID of the server in which this remote instance is in reference from
	CreatedBy InstanceId `json:"createdBy" bson:"createdBy"`

	// The time of the latest backup, in milliseconds since epoch
	LastBackup int64 `json:"lastBackup" bson:"lastBackup"`

	// Role the server is currently reporting. This is used to determine if the server is online (and functional) or not
	reportedRole ServerRole

	updateMu sync.RWMutex
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
	if wi.Role != CoreServer {
		return "", werror.WithStack(werror.Errorf("Cannot get address of non-core instance"))
	}
	return wi.Address, nil
}

func (wi *Instance) SetAddress(address string) error {
	if wi.Role != CoreServer {
		return werror.WithStack(werror.Errorf("Cannot set address of non-core instance"))
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

	// Role the server is currently reporting. This is used to determine if the server is online (and functional) or not
	ReportedRole ServerRole `json:"reportedRole"`

	Online bool `json:"online"`

	LastBackup int64 `json:"lastBackup"`

	BackupSize int64 `json:"backupSize"`
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

type ApiKeyInfo struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	Key         WeblensApiKey      `bson:"key" json:"key"`
	Owner       Username           `bson:"owner" json:"owner"`
	CreatedTime time.Time          `bson:"createdTime" json:"createdTime"`
	RemoteUsing InstanceId         `bson:"remoteUsing" json:"remoteUsing"`
	CreatedBy   InstanceId         `bson:"createdBy" json:"createdBy"`
}

type AccessService interface {
	GenerateJwtToken(user *User) (string, error)
	GetApiKey(key WeblensApiKey) (ApiKeyInfo, error)
	AddApiKey(key ApiKeyInfo) error
	GetUserFromToken(token string) (*User, error)
	DeleteApiKey(key WeblensApiKey) error
	GenerateApiKey(creator *User, local *Instance) (ApiKeyInfo, error)
	CanUserAccessFile(user *User, file *fileTree.WeblensFileImpl, share *FileShare) bool
	CanUserModifyShare(user *User, share Share) bool
	CanUserAccessAlbum(user *User, album *Album, share *AlbumShare) bool

	GetAllKeys(accessor *User) ([]ApiKeyInfo, error)
	GetAllKeysByServer(accessor *User, serverId InstanceId) ([]ApiKeyInfo, error)
	SetKeyUsedBy(key WeblensApiKey, server *Instance) error
}
