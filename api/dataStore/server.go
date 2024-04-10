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

func (si *srvInfo) GetUsingKey() types.WeblensApiKey {
	return si.UsingKey
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

func InitServerCore(name string, username types.Username, password string) error {
	user := GetUser(username)

	// Init with existing user
	if user != nil {
		if !CheckLogin(user, password) {
			return ErrUserNotAuthenticated
		} else if !user.IsAdmin() {
			err := MakeOwner(user)
			if err != nil {
				return err
			}
		}

	} else { // create new user, this will be the case 99% of the time
		err := CreateUser(username, password, true, true)
		if err != nil {
			return err
		}
	}

	srvId := util.GlobbyHash(12, name, time.Now().String())
	srv := srvInfo{
		Id:   srvId,
		Name: name,

		IsThisServer: true,
		Role:         types.CoreMode,
	}

	fddb.newServer(srv)
	thisServer = &srv

	return nil
}

func InitServerForBackup(name, coreAddress string, key types.WeblensApiKey, rq types.Requester) error {
	err := rq.AttachToCore(coreAddress, name, key)
	if err != nil {
		return err
	}

	srvId := util.GlobbyHash(12, name, time.Now().String())
	srv := srvInfo{
		Id:   srvId,
		Name: name,

		// Key is key used for remote core when IsThisServer
		UsingKey:     key,
		Role:         types.BackupMode,
		IsThisServer: true,
		CoreAddress:  coreAddress,
	}

	fddb.newServer(srv)
	thisServer = &srv
	return nil
}

// func SetCoreAddress(core string, key types.WeblensApiKey) error {
// 	if thisServer.Role == types.CoreMode {
// 		return ErrAlreadyCore
// 	}

// 	err := fddb.updateCoreAddress(core, key)
// 	if err != nil {
// 		return err
// 	}
// 	thisServer.CoreAddress = core
// 	return err
// }
