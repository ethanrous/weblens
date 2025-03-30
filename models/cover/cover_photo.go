package cover

import (
	"context"
	"errors"
)

type CoverPhoto struct {
	FolderId string `bson:"folderId"`

	// The ID of the cover photo
	CoverPhotoId string `bson:"coverPhotoId"`
}

func GetCoverByFolderId(ctx context.Context, folderId string) (*CoverPhoto, error) {
	// This function would typically query a database to retrieve the cover photo
	// based on the folder ID present in the context.
	// For now, let's return a placeholder value.

	// Placeholder return
	return &CoverPhoto{
		FolderId:     "exampleFolderId",
		CoverPhotoId: "exampleCoverPhotoId",
	}, nil

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

	_, err := fs.folderCoverCol.UpdateOne(
		context.Background(), bson.M{"folderId": folderId}, bson.M{"$set": bson.M{"coverId": coverId}},
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
