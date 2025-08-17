package cover

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var CoverPhotoCollectionKey = "coverPhoto"

type CoverPhoto struct {
	FolderId     string `bson:"folderId"`
	CoverPhotoId string `bson:"coverPhotoId"`
}

func GetCoverByFolderId(ctx context.Context, folderId string) (*CoverPhoto, error) {
	col, err := db.GetCollection[any](ctx, CoverPhotoCollectionKey)
	if err != nil {
		return nil, err
	}

	var coverPhoto CoverPhoto
	if err := col.FindOne(ctx, bson.M{"folderId": folderId}).Decode(&coverPhoto); err != nil {
		return nil, db.WrapError(err, "failed to get cover photo by folder ID %s", folderId)
	}

	return &coverPhoto, nil
}

func SetCoverPhoto(ctx context.Context, folderId string, coverPhotoId string) (*CoverPhoto, error) {
	col, err := db.GetCollection[any](ctx, CoverPhotoCollectionKey)
	if err != nil {
		return nil, err
	}

	coverPhoto := &CoverPhoto{
		FolderId:     folderId,
		CoverPhotoId: coverPhotoId,
	}

	_, err = col.ReplaceOne(ctx, bson.M{"folderId": folderId}, coverPhoto, options.Replace().SetUpsert(true))

	if err != nil {
		return nil, db.WrapError(err, "failed to set cover photo")
	}

	return coverPhoto, nil
}

func DeleteCoverByFolderId(ctx context.Context, folderId string) error {
	col, err := db.GetCollection[any](ctx, CoverPhotoCollectionKey)
	if err != nil {
		return err
	}

	result, err := col.DeleteOne(ctx, bson.M{"folderId": folderId})
	if err != nil {
		return db.WrapError(err, "failed to delete cover photo")
	}

	if result.DeletedCount == 0 {
		return db.NewNotFoundError("cover photo not found")
	}

	return nil
}

func UpsertCoverByFolderId(ctx context.Context, folderId, coverPhotoId string) error {
	col, err := db.GetCollection[any](ctx, CoverPhotoCollectionKey)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"folderId":     folderId,
			"coverPhotoId": coverPhotoId,
		},
	}

	opts := options.Update().SetUpsert(true)
	if _, err = col.UpdateOne(ctx, bson.M{"folderId": folderId}, update, opts); err != nil {
		return db.WrapError(err, "failed to upsert cover photo")
	}

	return nil
}
