package embedding

// atlasScoreToCosine inverts Atlas's (1 + cosine) / 2 score back to raw cosine in [-1, 1].
func atlasScoreToCosine(score float64) float64 {
	return 2*score - 1
}
