package embedding_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/embedding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newTestCtx(t *testing.T) context.Context {
	t.Helper()

	return db.SetupTestDB(t, embedding.CollectionKey,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "kind", Value: 1},
				{Key: "sourceId", Value: 1},
				{Key: "chunkIndex", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
}

func ones(n int) []float64 {
	v := make([]float64, n)

	norm := 1.0 / math.Sqrt(float64(n))

	for i := range v {
		v[i] = norm
	}

	return v
}

func seed(ctx context.Context, t *testing.T, docs []embedding.Embedding) {
	t.Helper()

	col, err := db.GetCollection[embedding.Embedding](ctx, embedding.CollectionKey)
	require.NoError(t, err)

	for i := range docs {
		if docs[i].ID.IsZero() {
			docs[i].ID = primitive.NewObjectID()
		}

		if docs[i].CreatedAt.IsZero() {
			docs[i].CreatedAt = time.Now()
		}

		_, err := col.InsertOne(ctx, docs[i])
		require.NoError(t, err)
	}
}

func countAllEmbeddings(ctx context.Context, t *testing.T) int {
	t.Helper()

	col, err := db.GetCollection[embedding.Embedding](ctx, embedding.CollectionKey)
	require.NoError(t, err)

	n, err := col.CountDocuments(ctx, bson.M{})
	require.NoError(t, err)

	return int(n)
}

func TestSearch_FiltersBySource(t *testing.T) {
	ctx := newTestCtx(t)

	seed(ctx, t, []embedding.Embedding{
		{Kind: embedding.KindFileChunk, SourceID: "file-1", ChunkIndex: 0, Vector: ones(1024), Model: "test"},
		{Kind: embedding.KindFileChunk, SourceID: "file-2", ChunkIndex: 0, Vector: ones(1024), Model: "test"},
	})

	hits, err := embedding.Search(ctx, embedding.Query{
		Vector:    ones(1024),
		SourceIDs: []string{"file-1"},
		Limit:     10,
	})
	require.NoError(t, err)
	assert.Len(t, hits, 1)
	assert.Equal(t, "file-1", hits[0].SourceID)
}

func TestSearch_NoFilterReturnsAll(t *testing.T) {
	ctx := newTestCtx(t)

	seed(ctx, t, []embedding.Embedding{
		{Kind: embedding.KindFileChunk, SourceID: "file-1", ChunkIndex: 0, Vector: ones(1024), Model: "test"},
		{Kind: embedding.KindFileChunk, SourceID: "file-2", ChunkIndex: 0, Vector: ones(1024), Model: "test"},
		{Kind: embedding.KindFileChunk, SourceID: "file-3", ChunkIndex: 0, Vector: ones(1024), Model: "test"},
	})

	hits, err := embedding.Search(ctx, embedding.Query{
		Vector: ones(1024),
		Limit:  10,
	})
	require.NoError(t, err)
	assert.Len(t, hits, 3)
}

func TestDeleteForSource_RemovesAllChunks(t *testing.T) {
	ctx := newTestCtx(t)

	seed(ctx, t, []embedding.Embedding{
		{Kind: embedding.KindFileChunk, SourceID: "file-x", ChunkIndex: 0, Vector: ones(1024), Model: "test"},
		{Kind: embedding.KindFileChunk, SourceID: "file-x", ChunkIndex: 1, Vector: ones(1024), Model: "test"},
		{Kind: embedding.KindFileChunk, SourceID: "file-y", ChunkIndex: 0, Vector: ones(1024), Model: "test"},
	})

	require.NoError(t, embedding.DeleteForSource(ctx, "file-x", embedding.KindFileChunk))

	remaining := countAllEmbeddings(ctx, t)
	assert.Equal(t, 1, remaining, "only file-y row should remain")
}

func TestDeleteAllOfKind_RemovesOnlyMatchingKind(t *testing.T) {
	ctx := newTestCtx(t)

	seed(ctx, t, []embedding.Embedding{
		{Kind: embedding.KindFileChunk, SourceID: "file-1", ChunkIndex: 0, Vector: ones(1024), Model: "test"},
		{Kind: embedding.KindImage, SourceID: "media-1", ChunkIndex: 0, Vector: ones(1024), Model: "test"},
	})

	require.NoError(t, embedding.DeleteAllOfKind(ctx, embedding.KindFileChunk))

	remaining := countAllEmbeddings(ctx, t)
	assert.Equal(t, 1, remaining, "only the KindImage row should remain")
}

func TestSourceIDsWithEmbeddings_ReturnsOnlyPresentOfKind(t *testing.T) {
	ctx := newTestCtx(t)

	seed(ctx, t, []embedding.Embedding{
		{Kind: embedding.KindFileChunk, SourceID: "file-1", ChunkIndex: 0, Vector: ones(1024), Model: "test"},
		{Kind: embedding.KindFileChunk, SourceID: "file-1", ChunkIndex: 1, Vector: ones(1024), Model: "test"},
		{Kind: embedding.KindFileChunk, SourceID: "file-2", ChunkIndex: 0, Vector: ones(1024), Model: "test"},
		{Kind: embedding.KindImage, SourceID: "media-1", ChunkIndex: 0, Vector: ones(1024), Model: "test"},
	})

	// file-3 has no rows; media-1 is the wrong kind, so neither should come back.
	present, err := embedding.SourceIDsWithEmbeddings(ctx, embedding.KindFileChunk,
		[]string{"file-1", "file-2", "file-3", "media-1"})
	require.NoError(t, err)

	assert.Equal(t, map[string]struct{}{"file-1": {}, "file-2": {}}, present)
}

func TestPruneTrailingChunks(t *testing.T) {
	ctx := newTestCtx(t)

	seed(ctx, t, []embedding.Embedding{
		{Kind: embedding.KindFileChunk, SourceID: "f", ChunkIndex: 0, Vector: ones(1024), Model: "test"},
		{Kind: embedding.KindFileChunk, SourceID: "f", ChunkIndex: 1, Vector: ones(1024), Model: "test"},
		{Kind: embedding.KindFileChunk, SourceID: "f", ChunkIndex: 2, Vector: ones(1024), Model: "test"},
		{Kind: embedding.KindFileChunk, SourceID: "f", ChunkIndex: 3, Vector: ones(1024), Model: "test"},
	})

	require.NoError(t, embedding.PruneTrailingChunks(ctx, embedding.KindFileChunk, "f", 2))

	remaining := countAllEmbeddings(ctx, t)
	assert.Equal(t, 2, remaining, "chunks 0,1 should remain; 2,3 pruned")
}
