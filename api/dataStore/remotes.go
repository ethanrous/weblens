package dataStore

import (
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func NewRemote(id, name string, key types.WeblensApiKey) error {
	if existing := dbServer.getUsingKey(key); existing != nil {
		util.Error.Println("Using server:", existing.Name)
		return ErrKeyInUse
	}

	remote := srvInfo{
		Id:       id,
		Name:     name,
		Role:     types.Backup,
		UsingKey: key,
	}
	err := dbServer.newServer(remote)
	if err != nil {
		return err
	}

	err = SetKeyRemote(key, remote.Id)
	return err
}

func GetRemotes() ([]*srvInfo, error) {
	return dbServer.getServers()
}

func DeleteRemote(remoteId string) {
	dbServer.removeServer(remoteId)
}
