package embedding

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Upsert inserts or replaces a row keyed by (kind, sourceId, chunkIndex); ID is regenerated if zero.
func Upsert(ctx context.Context, e Embedding) error {
	col, err := db.GetCollection[Embedding](ctx, CollectionKey)
	if err != nil {
		return err
	}

	if e.ID.IsZero() {
		e.ID = primitive.NewObjectID()
	}

	_, err = col.GetCollection().ReplaceOne(ctx,
		bson.M{
			"kind":       string(e.Kind),
			"sourceId":   e.SourceID,
			"chunkIndex": e.ChunkIndex,
		},
		e,
		options.Replace().SetUpsert(true),
	)

	return err
}

// DeleteForSource removes every row for one source (file or media) of the given kind.
func DeleteForSource(ctx context.Context, sourceID string, kind Kind) error {
	col, err := db.GetCollection[Embedding](ctx, CollectionKey)
	if err != nil {
		return err
	}

	_, err = col.DeleteMany(ctx, bson.M{
		"sourceId": sourceID,
		"kind":     string(kind),
	})

	return err
}

// DeleteAllOfKind removes every row of a given kind.
func DeleteAllOfKind(ctx context.Context, kind Kind) error {
	col, err := db.GetCollection[Embedding](ctx, CollectionKey)
	if err != nil {
		return err
	}

	_, err = col.DeleteMany(ctx, bson.M{"kind": string(kind)})

	return err
}

// DeleteAll removes every row from the embeddings collection.
func DeleteAll(ctx context.Context) error {
	col, err := db.GetCollection[Embedding](ctx, CollectionKey)
	if err != nil {
		return err
	}

	_, err = col.DeleteMany(ctx, bson.M{})

	return err
}

// PruneTrailingChunks deletes rows for (kind, sourceId) with chunkIndex >= keepFrom.
func PruneTrailingChunks(ctx context.Context, kind Kind, sourceID string, keepFrom int) error {
	col, err := db.GetCollection[Embedding](ctx, CollectionKey)
	if err != nil {
		return err
	}

	_, err = col.DeleteMany(ctx, bson.M{
		"kind":       string(kind),
		"sourceId":   sourceID,
		"chunkIndex": bson.M{"$gte": keepFrom},
	})

	return err
}

// CountByContentHash counts rows matching (kind, sourceId, model, contentHash) as an idempotency gate.
func CountByContentHash(ctx context.Context, kind Kind, sourceID, modelName, contentHash string) (int64, error) {
	col, err := db.GetCollection[Embedding](ctx, CollectionKey)
	if err != nil {
		return 0, err
	}

	return col.CountDocuments(ctx, bson.M{
		"kind":        string(kind),
		"sourceId":    sourceID,
		"model":       modelName,
		"contentHash": contentHash,
	})
}

// CountForSource counts rows for (kind, sourceId) across all chunks/models.
func CountForSource(ctx context.Context, kind Kind, sourceID string) (int64, error) {
	col, err := db.GetCollection[Embedding](ctx, CollectionKey)
	if err != nil {
		return 0, err
	}

	return col.CountDocuments(ctx, bson.M{
		"kind":     string(kind),
		"sourceId": sourceID,
	})
}

// GetForSource returns every embedding row for one source ID, ordered by chunkIndex ascending.
func GetForSource(ctx context.Context, sourceID string) ([]Embedding, error) {
	col, err := db.GetCollection[Embedding](ctx, CollectionKey)
	if err != nil {
		return nil, err
	}

	cursor, err := col.Find(ctx, bson.M{"sourceId": sourceID},
		options.Find().SetSort(bson.D{{Key: "chunkIndex", Value: 1}}))
	if err != nil {
		return nil, err
	}

	var rows []Embedding
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, err
	}

	return rows, nil
}

// CountForChunk counts rows matching (kind, sourceId, model, chunkIndex).
func CountForChunk(ctx context.Context, kind Kind, sourceID, modelName string, chunkIndex int) (int64, error) {
	col, err := db.GetCollection[Embedding](ctx, CollectionKey)
	if err != nil {
		return 0, err
	}

	return col.CountDocuments(ctx, bson.M{
		"kind":       string(kind),
		"sourceId":   sourceID,
		"model":      modelName,
		"chunkIndex": chunkIndex,
	})
}
