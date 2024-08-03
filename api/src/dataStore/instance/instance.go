package instance

import (
	"github.com/ethrousseau/weblens/api/types"
)

type WeblensInstance struct {
	Id   types.InstanceId `json:"id" bson:"_id"`
	Name string           `json:"name" bson:"name"`

	// Only applies to "core" server entries. This is the apiKey that remote server is using to connect to local,
	// if local is core. If local is backup, then this is the key being used to connect to remote core
	UsingKey types.WeblensApiKey `json:"-" bson:"usingKey"`

	// Core or Backup
	Role types.ServerRole `json:"role" bson:"serverRole"`

	// If this server info represents this local server
	IsThisServer bool `json:"-" bson:"isThisServer"`

	// Address of the remote server, only if the instance is a core.
	// Not set for any remotes/backups on core server, as it IS the core
	Address string `json:"coreAddress" bson:"coreAddress"`

	service *instanceService
}

func New(
	id types.InstanceId, name string, key types.WeblensApiKey, role types.ServerRole, isThisServer bool,
	address string,
) types.Instance {
	return &WeblensInstance{
		Id:           id,
		Name:         name,
		UsingKey:     key,
		Role:         role,
		IsThisServer: isThisServer,
		Address: address,
	}
}

func (wi *WeblensInstance) Info() types.Instance {
	return wi
}

func (wi *WeblensInstance) GetName() string {
	return wi.Name
}

func (wi *WeblensInstance) IsLocal() bool {
	return wi.IsThisServer
}

func (wi *WeblensInstance) IsCore() bool {
	return wi.Role == types.Core
}

func (wi *WeblensInstance) ServerId() types.InstanceId {
	return wi.Id
}

func (wi *WeblensInstance) SetServerId(id types.InstanceId) {
	wi.Id = id
}

func (wi *WeblensInstance) GetUsingKey() types.WeblensApiKey {
	return wi.UsingKey
}

func (wi *WeblensInstance) SetUsingKey(key types.WeblensApiKey) {
	wi.UsingKey = key
}

func (wi *WeblensInstance) ServerRole() types.ServerRole {
	return wi.Role
}

func (wi *WeblensInstance) GetRole() types.ServerRole {
	return wi.Role
	// if si == nil {
	// 	return false
	// }
	// return si.Role == types.Core
}

func (wi *WeblensInstance) GetAddress() (string, error) {
	if wi.Role != types.Core {
		return "", types.WeblensErrorMsg("Cannot get address of non-core instance")
	}
	return wi.Address, nil
}

func (wi *WeblensInstance) SetAddress(address string) error {
	if wi.Role != types.Core {
		return types.WeblensErrorMsg("Cannot set address of non-core instance")
	}
	wi.Address = address
	return nil
}

// func (wi *WeblensInstance) SetUserCount(count int) {
// 	wi.UserCount = count
// }

// func SetCoreAddress(core string, key types.WeblensApiKey) error {
// 	if thisServer.Role == types.CoreMode {
// 		return ErrAlreadyCore
// 	}

// 	err := dbServer.updateCoreAddress(core, key)
// 	if err != nil {
// 		return err
// 	}
// 	thisServer.Address = core
// 	return err
// }
