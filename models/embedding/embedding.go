// Package embedding defines the unified multimodal embedding collection; each row is a single vector.
package embedding

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CollectionKey is the MongoDB collection name.
const CollectionKey = "embeddings"

// Kind discriminates between embedding source modalities.
type Kind string

const (
	// KindImage identifies an embedding produced from an image (SourceID = media contentID).
	KindImage Kind = "image"
	// KindFileChunk identifies an embedding produced from a text chunk of a file (SourceID = fileID).
	KindFileChunk Kind = "file_chunk"

	// KindAll is a wildcard for deletion operations, not a valid value for stored rows.
	KindAll Kind = "all"
)

// Embedding is the unified model for all embedded vectors, each tied to a single source (image or file chunk).
type Embedding struct {
	ID          primitive.ObjectID `bson:"_id"`
	Kind        Kind               `bson:"kind"`
	SourceID    string             `bson:"sourceId"`
	ChunkIndex  int                `bson:"chunkIndex"`
	Page        int                `bson:"page,omitempty"`
	Snippet     string             `bson:"snippet"`
	Vector      []float64          `bson:"vector"`
	Model       string             `bson:"model"`
	CreatedAt   time.Time          `bson:"createdAt"`
	ContentHash string             `bson:"contentHash,omitempty"`
}
