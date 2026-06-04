package embedding

// atlasScoreToCosine converts an Atlas $vectorSearch cosine score back to raw
// cosine similarity. Atlas reports cosine similarity as (1 + cosine) / 2 so
// that scores fall in [0, 1]; this inverts that mapping to recover the raw
// cosine in [-1, 1], matching the brute-force path's scale.
func atlasScoreToCosine(score float64) float64 {
	return 2*score - 1
}
