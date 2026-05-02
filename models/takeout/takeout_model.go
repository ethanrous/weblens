// Package takeout contains types for a zip file of multiple files
package takeout

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	"go.mongodb.org/mongo-driver/bson"
)

// TakeoutCollectionKey is the key for the zip collection in the database
const TakeoutCollectionKey = "takeout"

// Takeout relates a zip file to its source files
type Takeout struct {
	SourceFileIDs []string `bson:"sourceFileIDs"`
	TakeoutFileID string   `bson:"zipFileID"`
}

// NewZip returns a new Zip struct.
func NewZip(sourceFileIDs []string, zipID string) *Takeout {
	return &Takeout{
		SourceFileIDs: sourceFileIDs,
		TakeoutFileID: zipID,
	}
}

// SaveZip saves a Zip struct to the database.
func SaveZip(ctx context.Context, zip *Takeout) error {
	col, err := db.GetCollection[Takeout](ctx, TakeoutCollectionKey)
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
func GetZip(ctx context.Context, zipFileID string) (*Takeout, error) {
	col, err := db.GetCollection[*Takeout](ctx, TakeoutCollectionKey)
	if err != nil {
		return nil, db.WrapError(err, "get zip collection")
	}

	var media Takeout

	err = col.FindOne(ctx, bson.M{"zipFileID": zipFileID}).Decode(&media)
	if err != nil {
		return nil, db.WrapError(err, "find zip in collection")
	}

	return &media, nil
}
