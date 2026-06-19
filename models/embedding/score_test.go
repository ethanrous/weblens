package embedding_test

import (
	"fmt"
	"testing"

	"github.com/ethanrous/weblens/models/embedding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noiseHits builds hits with generated source IDs, one per score.
func noiseHits(kind embedding.Kind, scores ...float64) []embedding.Hit {
	hits := make([]embedding.Hit, 0, len(scores))

	for i, s := range scores {
		hits = append(hits, embedding.Hit{
			Kind:     kind,
			SourceID: fmt.Sprintf("%s-noise-%d", kind, i),
			Score:    s,
		})
	}

	return hits
}

func findHit(hits []embedding.Hit, sourceID string) (embedding.Hit, bool) {
	for _, h := range hits {
		if h.SourceID == sourceID {
			return h, true
		}
	}

	return embedding.Hit{}, false
}

func TestScoreHits_ImageStandoutBeatsWeakerTextStandout(t *testing.T) {
	// The image standout is further above its kind's noise cluster than the text
	// standout is above its own, even though every raw text score is higher.
	hits := noiseHits(embedding.KindImage, 0.16, 0.18, 0.20, 0.18, 0.16, 0.17, 0.19, 0.18, 0.17, 0.18)
	hits = append(hits, embedding.Hit{Kind: embedding.KindImage, SourceID: "img-standout", Score: 0.34})
	hits = append(hits, noiseHits(embedding.KindFileChunk, 0.36, 0.44, 0.38, 0.42, 0.37, 0.43, 0.39, 0.41, 0.40, 0.40)...)
	hits = append(hits, embedding.Hit{Kind: embedding.KindFileChunk, SourceID: "text-standout", Score: 0.52})

	scored := embedding.ScoreHits(hits)

	img, ok := findHit(scored, "img-standout")
	require.True(t, ok, "image standout should survive")

	text, ok := findHit(scored, "text-standout")
	require.True(t, ok, "text standout should survive")

	assert.Greater(t, img.Score, text.Score)
}

func TestScoreHits_DropsHitsThatDontStandOut(t *testing.T) {
	// Noise hits sit above the kind floor but barely deviate from their own mean.
	hits := noiseHits(embedding.KindFileChunk, 0.39, 0.41, 0.40, 0.42, 0.40, 0.41, 0.39, 0.42, 0.40, 0.41)
	hits = append(hits, embedding.Hit{Kind: embedding.KindFileChunk, SourceID: "standout", Score: 0.55})

	scored := embedding.ScoreHits(hits)

	require.Len(t, scored, 1)
	assert.Equal(t, "standout", scored[0].SourceID)
}

func TestScoreHits_TightStrongClusterAllSurvive(t *testing.T) {
	// A near-uniform set of strong hits has no usable spread; all sit well above the
	// image floor (0.22), so all are kept.
	hits := noiseHits(embedding.KindImage, 0.35, 0.36, 0.35, 0.36, 0.35)

	scored := embedding.ScoreHits(hits)

	assert.Len(t, scored, len(hits))
}

func TestScoreHits_TightJunkClusterAllDropped(t *testing.T) {
	hits := noiseHits(embedding.KindImage, 0.10, 0.11, 0.10, 0.11, 0.10)

	scored := embedding.ScoreHits(hits)

	assert.Empty(t, scored)
}

func TestScoreHits_StrongClusterAllSurvive(t *testing.T) {
	// Many genuinely-strong matches clustered tightly, all well above the file_chunk
	// floor (0.40). The cluster has no statistical standout, but every hit is strong
	// in absolute terms, so none should be dropped.
	hits := noiseHits(embedding.KindFileChunk,
		0.55, 0.57, 0.56, 0.58, 0.55, 0.59, 0.56, 0.57, 0.58, 0.55, 0.59, 0.56)

	scored := embedding.ScoreHits(hits)

	assert.Len(t, scored, len(hits))
}

func TestScoreHits_FewHitsKeptByStrength(t *testing.T) {
	// Too few hits for distribution stats; only the strong one is kept.
	hits := []embedding.Hit{
		{Kind: embedding.KindFileChunk, SourceID: "strong", Score: 0.55},
		{Kind: embedding.KindFileChunk, SourceID: "weak", Score: 0.20},
	}

	scored := embedding.ScoreHits(hits)

	require.Len(t, scored, 1)
	assert.Equal(t, "strong", scored[0].SourceID)
}

func TestScoreHits_NormalizedScoresAndFieldsPreserved(t *testing.T) {
	hits := noiseHits(embedding.KindFileChunk, 0.39, 0.41, 0.40, 0.42, 0.40, 0.41, 0.39, 0.42, 0.40, 0.41)
	hits = append(hits, embedding.Hit{
		Kind:       embedding.KindFileChunk,
		SourceID:   "standout",
		ChunkIndex: 4,
		Page:       2,
		Snippet:    "the relevant passage",
		Score:      0.55,
	})

	scored := embedding.ScoreHits(hits)

	require.NotEmpty(t, scored)

	for _, h := range scored {
		assert.Greater(t, h.Score, 0.0)
		assert.Less(t, h.Score, 1.0)
	}

	h, ok := findHit(scored, "standout")
	require.True(t, ok)
	assert.Equal(t, embedding.KindFileChunk, h.Kind)
	assert.Equal(t, 4, h.ChunkIndex)
	assert.Equal(t, 2, h.Page)
	assert.Equal(t, "the relevant passage", h.Snippet)

	assert.Empty(t, embedding.ScoreHits(nil))
}
