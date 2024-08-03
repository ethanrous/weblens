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

func (db *databaseService) NewServer(i types.Instance) error {
	_, err := db.servers.InsertOne(db.ctx, i)
	if err != nil {
		return types.WeblensErrorFromError(err)
	}
	return nil
}

func (db *databaseService) AttachToCore(i types.Instance, core types.Instance) (types.Instance, error) {
	return nil, types.ErrNotImplemented("AttachToCore local")
}

func (db *databaseService) DeleteServer(id types.InstanceId) error {
	_, err := db.servers.DeleteOne(db.ctx, bson.M{"_id": id})
	if err != nil {
		return types.WeblensErrorFromError(err)
	}

	filter := bson.M{"remoteUsing": id}
	update := bson.M{"$set": bson.M{"remoteUsing": ""}}
	_, err = db.apiKeys.UpdateMany(db.ctx, filter, update)
	if err != nil {
		return types.WeblensErrorFromError(err)
	}
	return nil
}
