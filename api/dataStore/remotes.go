package dataStore

import (
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func NewRemote(name string, key types.WeblensApiKey) error {
	if existing := fddb.getUsingKey(key); existing != nil {
		util.Error.Println("Using server:", existing.Name)
		return ErrKeyInUse
	}

	remote := srvInfo{
		Name:     name,
		Role:     types.BackupMode,
		UsingKey: key,
	}
	fddb.newServer(remote)

	return nil
}
