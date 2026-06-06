package embedding_test

import (
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/embedding"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestEmbeddingBSONRoundTrip(t *testing.T) {
	orig := embedding.Embedding{
		ID:         primitive.NewObjectID(),
		Kind:       embedding.KindFileChunk,
		SourceID:   "file-abc",
		ChunkIndex: 3,
		Snippet:    "hello world",
		Vector:     []float64{0.1, 0.2, 0.3},
		Model:      "jina-clip-v2",
		CreatedAt:  time.Now().UTC().Truncate(time.Millisecond),
	}

	raw, err := bson.Marshal(orig)
	assert.NoError(t, err)

	var got embedding.Embedding
	assert.NoError(t, bson.Unmarshal(raw, &got))
	assert.Equal(t, orig, got)
}

func TestKindConstants(t *testing.T) {
	assert.Equal(t, embedding.Kind("image"), embedding.KindImage)
	assert.Equal(t, embedding.Kind("file_chunk"), embedding.KindFileChunk)
}
