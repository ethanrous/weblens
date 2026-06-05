package embedding

import (
	"context"
	"math"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/ethanrous/weblens/models/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// atlasSearchConfirmed is set true once $vectorSearch returns a non-empty result, confirming Atlas support.
var atlasSearchConfirmed atomic.Bool

// Hit is a single search result row.
type Hit struct {
	Kind       Kind    `bson:"kind"`
	SourceID   string  `bson:"sourceId"`
	ChunkIndex int     `bson:"chunkIndex"`
	Page       int     `bson:"page,omitempty"`
	Snippet    string  `bson:"snippet"`
	Score      float64 `bson:"score"`
}

// Query parameterizes a vector search.
type Query struct {
	// Vector is the query embedding to search against.
	Vector []float64
	// SourceIDs restricts results to the given source IDs. Empty means no restriction.
	SourceIDs []string
	// Kind restricts results to a specific embedding kind. Zero value means no restriction.
	Kind Kind
	// Limit is the maximum number of results to return. If zero, defaults to 10.
	Limit int
}

// Search runs Atlas $vectorSearch, falling back to a brute-force cosine scan when unavailable.
func Search(ctx context.Context, q Query) ([]Hit, error) {
	col, err := db.GetCollection[Embedding](ctx, CollectionKey)
	if err != nil {
		return nil, err
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 10
	}

	filter := buildFilter(q)

	vsStage := bson.M{
		"index":         vectorIndexName,
		"path":          "vector",
		"queryVector":   q.Vector,
		"numCandidates": limit * 10,
		"limit":         limit,
	}

	if len(filter) > 0 {
		vsStage["filter"] = filter
	}

	pipeline := bson.A{
		bson.M{"$vectorSearch": vsStage},
		bson.M{
			"$project": bson.M{
				"_id":        0,
				"kind":       1,
				"sourceId":   1,
				"chunkIndex": 1,
				"page":       1,
				"snippet":    1,
				"score":      bson.M{"$meta": "vectorSearchScore"},
			},
		},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		if isVectorSearchUnsupported(err) {
			return bruteForceCosineSearch(ctx, q, filter, limit)
		}

		return nil, err
	}

	var hits []Hit
	if err := cursor.All(ctx, &hits); err != nil {
		return nil, err
	}

	for i := range hits {
		hits[i].Score = atlasScoreToCosine(hits[i].Score)
	}

	if len(hits) > 0 {
		atlasSearchConfirmed.Store(true)

		return hits, nil
	}

	if atlasSearchConfirmed.Load() {
		return hits, nil
	}

	return bruteForceCosineSearch(ctx, q, filter, limit)
}

func buildFilter(q Query) bson.M {
	filter := bson.M{}

	if q.Kind != "" {
		filter["kind"] = string(q.Kind)
	}

	if len(q.SourceIDs) > 0 {
		filter["sourceId"] = bson.M{"$in": q.SourceIDs}
	}

	return filter
}

func isVectorSearchUnsupported(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	return strings.Contains(msg, "$vectorSearch") ||
		strings.Contains(msg, "vectorSearch") ||
		strings.Contains(msg, "Unrecognized pipeline stage") ||
		strings.Contains(msg, "unknown top level operator")
}

func bruteForceCosineSearch(ctx context.Context, q Query, filter bson.M, limit int) ([]Hit, error) {
	col, err := db.GetCollection[Embedding](ctx, CollectionKey)
	if err != nil {
		return nil, err
	}

	cursor, err := col.Find(ctx, filter, options.Find().SetProjection(bson.M{
		"kind":       1,
		"sourceId":   1,
		"chunkIndex": 1,
		"page":       1,
		"snippet":    1,
		"vector":     1,
	}))
	if err != nil {
		return nil, err
	}

	var docs []Embedding
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	type scored struct {
		hit   Hit
		score float64
	}

	results := make([]scored, 0, len(docs))

	for _, doc := range docs {
		s := cosine(q.Vector, doc.Vector)
		results = append(results, scored{
			hit: Hit{
				Kind:       doc.Kind,
				SourceID:   doc.SourceID,
				ChunkIndex: doc.ChunkIndex,
				Page:       doc.Page,
				Snippet:    doc.Snippet,
				Score:      s,
			},
			score: s,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if limit < len(results) {
		results = results[:limit]
	}

	hits := make([]Hit, len(results))
	for i, r := range results {
		hits[i] = r.hit
	}

	return hits, nil
}

func cosine(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, normA, normB float64

	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}

	return dot / denom
}
