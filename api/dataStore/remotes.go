package dataStore

import (
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func NewRemote(id, name string, key types.WeblensApiKey) error {
	if existing := fddb.getUsingKey(key); existing != nil {
		util.Error.Println("Using server:", existing.Name)
		return ErrKeyInUse
	}

	remote := srvInfo{
		Id:       id,
		Name:     name,
		Role:     types.Backup,
		UsingKey: key,
	}
	err := fddb.newServer(remote)
	if err != nil {
		return err
	}

	err = SetKeyRemote(key, remote.Id)
	return err
}

func GetRemotes() ([]*srvInfo, error) {
	return fddb.getServers()
}

func DeleteRemote(remoteId string) {
	fddb.removeServer(remoteId)
}
