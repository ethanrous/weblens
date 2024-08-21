package database

import (
	error2 "github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/types"
	"go.mongodb.org/mongo-driver/bson"
)

func (db *databaseService) CreateApiKey(key types.ApiKeyInfo) error {
	_, err := db.apiKeys.InsertOne(db.ctx, key)
	if err != nil {
		return error2.Wrap(err)
	}

	return nil
}

func (db *databaseService) GetApiKeys() ([]types.ApiKeyInfo, error) {
	ret, err := db.apiKeys.Find(db.ctx, bson.M{})
	if err != nil {
		return nil, error2.Wrap(err)
	}

	target := []types.ApiKeyInfo{}
	err = ret.All(db.ctx, &target)
	if err != nil {
		return nil, error2.Wrap(err)
	}

	return target, nil
}

func (db *databaseService) DeleteApiKey(key types.WeblensApiKey) error {
	_, err := db.apiKeys.DeleteOne(db.ctx, bson.M{"key": key})
	if err != nil {
		return error2.Wrap(err)
	}

	return nil
}

func (db *databaseService) SetRemoteUsing(key types.WeblensApiKey, remoteId types.InstanceId) error {
	filter := bson.M{"key": key}
	update := bson.M{"$set": bson.M{"remoteUsing": remoteId}}
	_, err := db.apiKeys.UpdateOne(db.ctx, filter, update)
	if err != nil {
		return error2.Wrap(err)
	}

	return nil
}
