package embedding

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/startup"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EmbeddingDim is the output dimensionality of the configured multimodal model.
// Jina CLIP v2 default = 1024. If the model is swapped, update this constant
// and re-create the vector index.
const EmbeddingDim = 1024

const (
	vectorIndexName         = "embeddings_vector"
	sourceIDIndexKey        = "embeddings_sourceId_index"
	kindSourceChunkIndexKey = "embeddings_kind_sourceId_chunkIndex_unique"
)

func init() {
	startup.RegisterHook(registerEmbeddings)
}

func registerEmbeddings(ctx context.Context, _ config.Provider) error {
	col, err := db.GetCollection[any](ctx, CollectionKey)
	if err != nil {
		return err
	}

	if err := col.NewIndex(mongo.IndexModel{
		Keys:    bson.D{{Key: "sourceId", Value: 1}},
		Options: options.Index().SetName(sourceIDIndexKey),
	}); err != nil {
		return err
	}

	if err := col.NewIndex(mongo.IndexModel{
		Keys: bson.D{
			{Key: "kind", Value: 1},
			{Key: "sourceId", Value: 1},
			{Key: "chunkIndex", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName(kindSourceChunkIndexKey),
	}); err != nil {
		return err
	}

	if err := col.NewSearchIndex(mongo.SearchIndexModel{
		Definition: bson.M{
			"fields": bson.A{
				bson.M{
					"type":          "vector",
					"path":          "vector",
					"numDimensions": EmbeddingDim,
					"similarity":    "cosine",
					"quantization":  "scalar",
				},
				bson.M{"type": "filter", "path": "kind"},
				bson.M{"type": "filter", "path": "sourceId"},
			},
		},
		Options: options.SearchIndexes().SetName(vectorIndexName).SetType("vectorSearch"),
	}); err != nil {
		return err
	}

	return nil
}
