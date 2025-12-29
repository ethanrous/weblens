// Package cover manages folder cover photos and their associations.
package cover

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CoverPhotoCollectionKey is the MongoDB collection name for cover photos.
var CoverPhotoCollectionKey = "coverPhoto"

// Photo represents a folder's cover photo mapping.
type Photo struct {
	FolderID     string `bson:"folderID"`
	CoverPhotoID string `bson:"coverPhotoID"`
}

// GetCoverByFolderID retrieves the cover photo for a folder by its ID.
func GetCoverByFolderID(ctx context.Context, folderID string) (*Photo, error) {
	col, err := db.GetCollection[any](ctx, CoverPhotoCollectionKey)
	if err != nil {
		return nil, err
	}

	var coverPhoto Photo
	if err := col.FindOne(ctx, bson.M{"folderID": folderID}).Decode(&coverPhoto); err != nil {
		return nil, db.WrapError(err, "failed to get cover photo by folder ID %s", folderID)
	}

	return &coverPhoto, nil
}

// SetCoverPhoto sets or replaces the cover photo for a folder.
func SetCoverPhoto(ctx context.Context, folderID string, coverPhotoID string) (*Photo, error) {
	col, err := db.GetCollection[any](ctx, CoverPhotoCollectionKey)
	if err != nil {
		return nil, err
	}

	coverPhoto := &Photo{
		FolderID:     folderID,
		CoverPhotoID: coverPhotoID,
	}

	_, err = col.ReplaceOne(ctx, bson.M{"folderID": folderID}, coverPhoto, options.Replace().SetUpsert(true))
	if err != nil {
		return nil, db.WrapError(err, "failed to set cover photo")
	}

	return coverPhoto, nil
}

// DeleteCoverByFolderID removes the cover photo for a folder.
func DeleteCoverByFolderID(ctx context.Context, folderID string) error {
	col, err := db.GetCollection[any](ctx, CoverPhotoCollectionKey)
	if err != nil {
		return err
	}

	result, err := col.DeleteOne(ctx, bson.M{"folderID": folderID})
	if err != nil {
		return db.WrapError(err, "failed to delete cover photo")
	}

	if result.DeletedCount == 0 {
		return db.NewNotFoundError("cover photo not found")
	}

	return nil
}

// UpsertCoverByFolderID creates or updates the cover photo for a folder.
func UpsertCoverByFolderID(ctx context.Context, folderID, coverPhotoID string) error {
	col, err := db.GetCollection[any](ctx, CoverPhotoCollectionKey)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"folderID":     folderID,
			"coverPhotoID": coverPhotoID,
		},
	}

	opts := options.Update().SetUpsert(true)
	if _, err = col.UpdateOne(ctx, bson.M{"folderID": folderID}, update, opts); err != nil {
		return db.WrapError(err, "failed to upsert cover photo")
	}

	return nil
}
