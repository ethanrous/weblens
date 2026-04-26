// Package zip contains types for a zip file of multiple files
package zip

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	"go.mongodb.org/mongo-driver/bson"
)

// ZipCollectionKey is the key for the zip collection in the database
const ZipCollectionKey = "zip"

// Zip relates a zip file to its source files
type Zip struct {
	SourceFileIDs []string `bson:"sourceFileIDs"`
	ZipFileID     string   `bson:"zipFileID"`
}

// NewZip returns a new Zip struct.
func NewZip(sourceFileIDs []string, zipID string) *Zip {
	return &Zip{
		SourceFileIDs: sourceFileIDs,
		ZipFileID:     zipID,
	}
}

// SaveZip saves a Zip struct to the database.
func SaveZip(ctx context.Context, zip *Zip) error {
	col, err := db.GetCollection[Zip](ctx, ZipCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.InsertOne(ctx, zip)
	if err != nil {
		return db.WrapError(err, "insert zip into collection")
	}

	return nil
}

// GetZip returns a Zip struct from the database.
func GetZip(ctx context.Context, zipFileID string) (*Zip, error) {
	col, err := db.GetCollection[*Zip](ctx, ZipCollectionKey)
	if err != nil {
		return nil, db.WrapError(err, "get zip collection")
	}

	var media Zip

	err = col.FindOne(ctx, bson.M{"zipFileID": zipFileID}).Decode(&media)
	if err != nil {
		return nil, db.WrapError(err, "find zip in collection")
	}

	return &media, nil
}
