package file

import (
	"testing"

	"github.com/ethanrous/weblens/models/embedding"
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
