package file

import (
	"testing"

	"github.com/ethanrous/weblens/models/embedding"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/modules/wlfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterContentHits_TextSuppressesMarginalImage(t *testing.T) {
	hits := []embedding.Hit{
		{Kind: embedding.KindFileChunk, SourceID: "doc1", Score: 0.55, Snippet: "SOC 2 Type II report"},
		{Kind: embedding.KindFileChunk, SourceID: "doc2", Score: 0.10},
		{Kind: embedding.KindImage, SourceID: "seal", Score: 0.30},
		{Kind: embedding.KindImage, SourceID: "other", Score: 0.12},
	}

	out := filterContentHits(hits)

	var keptSeal, keptDoc bool

	for _, h := range out {
		if h.SourceID == "seal" {
			keptSeal = true
		}

		if h.SourceID == "doc1" {
			keptDoc = true
		}
	}

	if !keptDoc {
		t.Fatal("expected confident document hit to be kept")
	}

	if keptSeal {
		t.Fatal("marginal image hit should be suppressed by a confident text match")
	}
}

func TestFilterContentHits_ImageOnlyQueryStillReturnsImages(t *testing.T) {
	hits := []embedding.Hit{
		{Kind: embedding.KindImage, SourceID: "seal", Score: 0.42},
		{Kind: embedding.KindImage, SourceID: "a", Score: 0.18},
		{Kind: embedding.KindImage, SourceID: "b", Score: 0.16},
	}

	out := filterContentHits(hits)

	var keptSeal bool

	for _, h := range out {
		if h.SourceID == "seal" {
			keptSeal = true
		}
	}

	if !keptSeal {
		t.Fatal("image-only query should still return its top image")
	}
}

func TestMergeContentHits_ImageHitPreservesTextSnippetAndPage(t *testing.T) {
	// A higher-scoring image hit must not wipe the text hit's snippet/page.
	f := file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:    wlfs.BuildFilePath("test", "report.pdf"),
		FileID:  "doc1",
		MemOnly: true,
	})
	require.NotNil(t, f)

	byMedia := map[string][]*file_model.WeblensFileImpl{"cid1": {f}}

	hits := []embedding.Hit{
		{Kind: embedding.KindFileChunk, SourceID: "doc1", Score: 0.55, Snippet: "the matched text", Page: 3},
		{Kind: embedding.KindImage, SourceID: "cid1", Score: 0.62},
	}

	out := mergeContentHits(hits, byMedia)

	got, ok := out["doc1"]
	require.True(t, ok, "the file should have a content hit")
	assert.InDelta(t, 0.62, got.Score, 1e-9, "the higher image score should win")
	assert.Equal(t, "the matched text", got.Snippet, "the text snippet must be preserved when an image out-scores it")
	assert.Equal(t, 3, got.Page, "the text page must be preserved when an image out-scores it")
}
