package tag

import (
	"context"
	"strings"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TagCollectionKey is the MongoDB collection name for tags.
const TagCollectionKey = "tags"

// ErrTagNotFound is returned when a requested tag does not exist.
var ErrTagNotFound = wlerrors.New("tag not found")

// Tag represents a user-defined tag for organizing files.
type Tag struct {
	TagID   primitive.ObjectID `bson:"_id"       json:"id"`
	Name    string             `bson:"name"       json:"name"`
	Color   string             `bson:"color"      json:"color"`
	Owner   string             `bson:"owner"      json:"owner"`
	FileIDs []string           `bson:"fileIDs"    json:"fileIDs"`
	Created time.Time          `bson:"created"    json:"created"`
	Updated time.Time          `bson:"updated"    json:"updated"`
}

// CreateTag creates a new tag with the given name, color, and owner.
func CreateTag(ctx context.Context, name, color, owner string) (*Tag, error) {
	col, err := db.GetCollection[any](ctx, TagCollectionKey)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	tag := &Tag{
		TagID:   primitive.NewObjectID(),
		Name:    name,
		Color:   color,
		Owner:   owner,
		FileIDs: []string{},
		Created: now,
		Updated: now,
	}

	_, err = col.InsertOne(ctx, tag)
	if err != nil {
		return nil, db.WrapError(err, "failed to create tag")
	}

	return tag, nil
}

// GetTagByID retrieves a tag by its MongoDB ObjectID.
func GetTagByID(ctx context.Context, tagID primitive.ObjectID) (*Tag, error) {
	col, err := db.GetCollection[any](ctx, TagCollectionKey)
	if err != nil {
		return nil, err
	}

	var tag Tag
	if err := col.FindOne(ctx, bson.M{"_id": tagID}).Decode(&tag); err != nil {
		return nil, db.WrapError(err, "failed to get tag %s", tagID.Hex())
	}

	return &tag, nil
}

// GetTagsByOwner retrieves all tags belonging to the given owner.
func GetTagsByOwner(ctx context.Context, owner string) ([]*Tag, error) {
	col, err := db.GetCollection[any](ctx, TagCollectionKey)
	if err != nil {
		return nil, err
	}

	var tags []*Tag

	cursor, err := col.Find(ctx, bson.M{"owner": owner})
	if err != nil {
		return nil, db.WrapError(err, "failed to get tags for owner %s", owner)
	}

	if err := cursor.All(ctx, &tags); err != nil {
		return nil, db.WrapError(err, "failed to decode tags")
	}

	return tags, nil
}

// GetTagsForFile retrieves all tags that contain the given file ID.
func GetTagsForFile(ctx context.Context, fileID string) ([]*Tag, error) {
	col, err := db.GetCollection[any](ctx, TagCollectionKey)
	if err != nil {
		return nil, err
	}

	var tags []*Tag

	cursor, err := col.Find(ctx, bson.M{"fileIDs": fileID})
	if err != nil {
		return nil, db.WrapError(err, "failed to get tags for file %s", fileID)
	}

	if err := cursor.All(ctx, &tags); err != nil {
		return nil, db.WrapError(err, "failed to decode tags")
	}

	return tags, nil
}

// UpdateTag updates a tag's name and/or color.
func UpdateTag(ctx context.Context, tagID primitive.ObjectID, name, color string) error {
	col, err := db.GetCollection[any](ctx, TagCollectionKey)
	if err != nil {
		return err
	}

	update := bson.M{"$set": bson.M{"updated": time.Now()}}

	setFields := update["$set"].(bson.M)

	if name != "" {
		setFields["name"] = name
	}

	if color != "" {
		setFields["color"] = color
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": tagID}, update)
	if err != nil {
		if strings.Contains(err.Error(), "no documents matched") {
			return ErrTagNotFound
		}

		return db.WrapError(err, "failed to update tag")
	}

	return nil
}

// DeleteTag permanently removes a tag by its ID.
func DeleteTag(ctx context.Context, tagID primitive.ObjectID) error {
	col, err := db.GetCollection[any](ctx, TagCollectionKey)
	if err != nil {
		return err
	}

	result, err := col.DeleteOne(ctx, bson.M{"_id": tagID})
	if err != nil {
		return db.WrapError(err, "failed to delete tag")
	}

	if result.DeletedCount == 0 {
		return ErrTagNotFound
	}

	return nil
}

// AddFilesToTag adds file IDs to a tag, deduplicating via $addToSet.
func AddFilesToTag(ctx context.Context, tagID primitive.ObjectID, fileIDs []string) error {
	col, err := db.GetCollection[any](ctx, TagCollectionKey)
	if err != nil {
		return err
	}

	update := bson.M{
		"$addToSet": bson.M{"fileIDs": bson.M{"$each": fileIDs}},
		"$set":      bson.M{"updated": time.Now()},
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": tagID}, update)
	if err != nil {
		if strings.Contains(err.Error(), "no documents matched") {
			return ErrTagNotFound
		}

		return db.WrapError(err, "failed to add files to tag")
	}

	return nil
}

// RemoveFilesFromTag removes file IDs from a tag.
func RemoveFilesFromTag(ctx context.Context, tagID primitive.ObjectID, fileIDs []string) error {
	col, err := db.GetCollection[any](ctx, TagCollectionKey)
	if err != nil {
		return err
	}

	update := bson.M{
		"$pull": bson.M{"fileIDs": bson.M{"$in": fileIDs}},
		"$set":  bson.M{"updated": time.Now()},
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": tagID}, update)
	if err != nil {
		if strings.Contains(err.Error(), "no documents matched") {
			return ErrTagNotFound
		}

		return db.WrapError(err, "failed to remove files from tag")
	}

	return nil
}

// RemoveFileFromAllTags removes a file ID from every tag that contains it.
func RemoveFileFromAllTags(ctx context.Context, fileID string) error {
	col, err := db.GetCollection[any](ctx, TagCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.UpdateMany(
		ctx,
		bson.M{"fileIDs": fileID},
		bson.M{
			"$pull": bson.M{"fileIDs": fileID},
			"$set":  bson.M{"updated": time.Now()},
		},
	)
	if err != nil {
		return db.WrapError(err, "failed to remove file from all tags")
	}

	return nil
}

// GetFileIDsByTagIDs returns file IDs that appear in ALL of the given tags (AND logic).
func GetFileIDsByTagIDs(ctx context.Context, tagIDs []primitive.ObjectID) ([]string, error) {
	col, err := db.GetCollection[any](ctx, TagCollectionKey)
	if err != nil {
		return nil, err
	}

	var tags []*Tag

	cursor, err := col.Find(ctx, bson.M{"_id": bson.M{"$in": tagIDs}}, options.Find().SetProjection(bson.M{"fileIDs": 1}))
	if err != nil {
		return nil, db.WrapError(err, "failed to get file IDs by tag IDs")
	}

	if err := cursor.All(ctx, &tags); err != nil {
		return nil, db.WrapError(err, "failed to decode tags")
	}

	// Intersect file IDs across all tags (AND logic)
	if len(tags) == 0 {
		return []string{}, nil
	}

	fileSet := make(map[string]int)

	for _, tag := range tags {
		for _, fID := range tag.FileIDs {
			fileSet[fID]++
		}
	}

	var result []string

	for fID, count := range fileSet {
		if count == len(tags) {
			result = append(result, fID)
		}
	}

	return result, nil
}
