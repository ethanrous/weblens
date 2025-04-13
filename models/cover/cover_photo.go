package cover

import (
	"context"
	"errors"

	"github.com/ethanrous/weblens/models/db"
	"go.mongodb.org/mongo-driver/bson"
)

var CoverPhotoCollectionKey = "coverPhoto"

type CoverPhoto struct {
	FolderId string `bson:"folderId"`

	// The ID of the cover photo
	CoverPhotoId string `bson:"coverPhotoId"`
}

func GetCoverByFolderId(ctx context.Context, folderId string) (*CoverPhoto, error) {
	col, err := db.GetCollection(ctx, CoverPhotoCollectionKey)
	if err != nil {
		return nil, err
	}

	var coverPhoto CoverPhoto
	err = col.FindOne(ctx, bson.M{"folderId": folderId}).Decode(&coverPhoto)
	if err != nil {
		return nil, db.WrapError(err, "failed to get cover photo by folder ID %s", folderId)
	}

	return &coverPhoto, nil

}

func SetCoverPhoto(ctx context.Context, folderId string, coverPhotoId string) (*CoverPhoto, error) {
	// This function would typically update the database with the new cover photo ID
	// for the specified folder ID in the context.
	// For now, let's return a placeholder value.

	// Placeholder return
	return &CoverPhoto{
		FolderId:     folderId,
		CoverPhotoId: coverPhotoId,
	}, nil
}

func DeleteCoverByFolderId(ctx context.Context, folderId string) error {
	// 	_, err := fs.folderCoverCol.DeleteOne(context.Background(), bson.M{"folderId": folderId})
	// if err != nil {
	// 	return errors.WithStack(err)
	// }
	//
	// delete(fs.folderMedia, folderId)
	// folder.SetContentId("")
	// return nil
	//
	return errors.New("not implemented")
}

func UpsertCoverByFolderId(ctx context.Context, folderId, coverPhotoId string) error {
	col, err := db.GetCollection(ctx, CoverPhotoCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.UpdateOne(
		context.Background(), bson.M{"folderId": folderId}, bson.M{"$set": bson.M{"coverId": coverPhotoId}},
	)
	// 	_, err := fs.folderCoverCol.DeleteOne(context.Background(), bson.M{"folderId": folderId})
	// if err != nil {
	// 	return errors.WithStack(err)
	// }
	//
	// delete(fs.folderMedia, folderId)
	// folder.SetContentId("")
	// return nil
	//
	return errors.New("not implemented")
}
