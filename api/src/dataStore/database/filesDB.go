package database

import (
	"github.com/ethrousseau/weblens/api/types"
	"go.mongodb.org/mongo-driver/bson"
)

func (db *databaseService) NewTrashEntry(te types.TrashEntry) error {
	_, err := db.trash.InsertOne(db.ctx, te)
	return err
}

func (db *databaseService) DeleteTrashEntry(fileId types.FileId) error {
	_, err := db.trash.DeleteOne(db.ctx, bson.M{"trashFileId": fileId})
	return err
}

func (db *databaseService) GetTrashEntry(fileId types.FileId) (te types.TrashEntry, err error) {
	filter := bson.D{{"trashFileId", fileId}}
	ret := db.trash.FindOne(db.ctx, filter)
	if err = ret.Err(); err != nil {
		return
	}

	err = ret.Decode(&te)
	return
}
