package database

import (
	"github.com/ethrousseau/weblens/api/dataStore/media"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson"
)

func (db *databaseService) GetAllMedia() ([]types.Media, error) {
	ret, err := db.media.Find(db.ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var target []*media.Media
	err = ret.All(db.ctx, &target)
	if err != nil {
		return nil, err
	}

	for _, m := range target {
		m.SetImported(true)
	}

	return util.SliceConvert[types.Media](target), nil
}

func (db *databaseService) CreateMedia(m types.Media) error {
	_, err := db.media.InsertOne(db.ctx, m)
	return err
}

func (db *databaseService) AddFileToMedia(mId types.ContentId, fId types.FileId) error {
	filter := bson.M{"contentId": mId}
	update := bson.M{"$addToSet": bson.M{"fileIds": fId}}
	_, err := db.media.UpdateOne(db.ctx, filter, update)
	return err
}

func (db *databaseService) RemoveFileFromMedia(mId types.ContentId, fId types.FileId) error {
	filter := bson.M{"contentId": mId}
	update := bson.M{"$pull": bson.M{"fileIds": fId}}
	_, err := db.media.UpdateOne(db.ctx, filter, update)
	return err
}

func (db *databaseService) DeleteMedia(id types.ContentId) error {
	filter := bson.M{"contentId": id}
	_, err := db.media.DeleteOne(db.ctx, filter)
	return err
}

func (db *databaseService) HideMedia(id types.ContentId) error {
	filter := bson.M{"contentId": id}
	_, err := db.media.UpdateOne(db.ctx, filter, bson.M{"$set": bson.M{"hidden": true}})
	return err
}
