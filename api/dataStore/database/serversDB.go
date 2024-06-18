package database

import (
	"github.com/ethrousseau/weblens/api/dataStore/instance"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson"
)

func (db *databaseService) GetAllServers() ([]types.Instance, error) {
	ret, err := db.servers.Find(db.ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var servers []*instance.WeblensInstance
	err = ret.All(db.ctx, &servers)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[types.Instance](servers), nil
}

func (db *databaseService) NewServer(id types.InstanceId, name string, isThisServer bool, role types.ServerRole) error {
	return types.ErrNotImplemented("new server db")
}
