package dataStore

import (
	"slices"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

var thisServer *srvInfo
var thisOwner types.User

// Can return nil. When nil is returned, all hanlders must
// do their best to re-direct to the setup page
func GetServerInfo() types.ServerInfo {
	if thisServer == nil {
		ret, err := fddb.getThisServerInfo()
		if err == nil {
			thisServer = ret
		} else {
			return nil
		}
	}
	thisServer.UserCount = UserCount()
	return thisServer
}

func GetOwner() types.User {
	if thisOwner != nil {
		return thisOwner
	} else if thisOwner == nil && (thisServer == nil || thisServer.Role != types.BackupMode) {
		i := slices.IndexFunc(GetUsers(), func(u types.User) bool { return u.IsOwner() })
		if i != -1 {
			thisOwner = GetUsers()[i]
			return thisOwner
		}
	}
	return nil
}

func (si *srvInfo) ServerId() string {
	return si.Id
}

func (si *srvInfo) ServerRole() types.ServerRole {
	return si.Role
}

func (si *srvInfo) GetCoreAddress() (string, error) {
	if si.Role == types.CoreMode {
		return "", ErrAlreadyCore
	}
	return si.CoreAddress, nil
}

func InitServer(name string, role types.ServerRole) {
	srvId := util.GlobbyHash(12, name, time.Now().String())
	srv := srvInfo{
		Id:   srvId,
		Name: name,

		// Key is always empty when IsThisServer
		UsingKey:     "",
		IsThisServer: true,
		Role:         role,
	}

	fddb.newServer(srv)
	thisServer = &srv
}

func SetCoreAddress(core string) error {
	if thisServer.Role == types.CoreMode {
		return ErrAlreadyCore
	}

	err := fddb.updateCoreAddress(core)
	if err != nil {
		return err
	}
	thisServer.CoreAddress = core
	return err
}
