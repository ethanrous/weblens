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
	panic("implement me")
}

func (db *databaseService) SetShareEnabledById(sId types.ShareId, enabled bool) error {
	panic("implement me")
}
