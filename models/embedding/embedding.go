// Package embedding defines the unified multimodal embedding collection.
// Each row represents a single vector
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
)

// Embedding is the unified model for all embedded vectors. Each embedding is associated with a single source:
//   - For KindImage: SourceID is the media contentID, ChunkIndex is the
//     0-indexed page (always 0 for single-page media), Snippet is empty.
//   - For KindFileChunk: SourceID is the fileID, ChunkIndex enumerates chunks
//     from 0, Snippet is a short preview of the chunk text.
//
// Page is the 1-indexed source page the chunk represents (PDF page, XLSX
// sheet, PPTX slide, image page in a multi-page document). It is set to 0
// for legacy rows written before the field existed.
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
