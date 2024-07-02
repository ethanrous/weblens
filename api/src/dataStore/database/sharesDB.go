package database

import (
	"github.com/ethrousseau/weblens/api/dataStore/share"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson"
)

func (db *databaseService) GetAllShares() ([]types.Share, error) {
	ret, err := db.shares.Find(db.ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var target []*share.FileShare
	err = ret.All(db.ctx, &target)

	return util.SliceConvert[types.Share](target), err
}

func (db *databaseService) UpdateShare(s types.Share) error {
	return types.ErrNotImplemented("UpdateShare() has not yet been implemented")
}

func (db *databaseService) SetShareEnabledById(sId types.ShareId, enabled bool) error {
	return types.ErrNotImplemented("SetShareEnabledById() has not yet been implemented")
}

func (db *databaseService) CreateShare(share types.Share) error {
	_, err := db.shares.InsertOne(db.ctx, share)
	return err
}

func (db *databaseService) AddUsersToShare(share types.Share, users []types.Username) error {
	filter := bson.M{"_id": share.GetShareId()}
	update := bson.M{"$addToSet": bson.M{"accessors": bson.M{"$each": users}}}
	_, err := db.shares.UpdateOne(db.ctx, filter, update)
	return err
}

func (db *databaseService) GetSharedWithUser(username types.Username) ([]types.Share, error) {
	filter := bson.M{"accessors": username, "shareType": "file"}
	ret, err := db.shares.Find(db.ctx, filter)
	if err != nil {
		return nil, err
	}

	var target []*share.FileShare
	err = ret.All(db.ctx, &target)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[types.Share](target), nil
}
