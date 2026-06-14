package embedding

import "math"

const (
	// zKeep is the minimum standard deviations above its kind's mean a hit must score to survive.
	zKeep = 2.0

	// minKindHits and minKindSigma are the floors below which a kind's score
	// distribution is too small or too uniform for z-stats to mean anything.
	minKindHits  = 3
	minKindSigma = 0.01

	// fallbackSigma stands in for the distribution width when z-stats are unusable.
	fallbackSigma = 0.05
)

// kindFloor is the per-kind raw-cosine noise ceiling; hits below it are junk regardless
// of distribution. Cross-modal (text query → image) cosines run systematically lower
// than text → text, so the floors differ per kind. Empirical for jina-clip-v2.
var kindFloor = map[Kind]float64{
	KindFileChunk: 0.40,
	KindImage:     0.22,
}

// atlasScoreToCosine inverts Atlas's (1 + cosine) / 2 score back to raw cosine in [-1, 1].
func atlasScoreToCosine(score float64) float64 {
	return 2*score - 1
}

// ScoreHits rescales raw per-kind cosine scores onto a shared (0, 1) scale so hits of
// different kinds are comparable, dropping hits that don't stand out from their own
// kind's score distribution. A hit's score becomes logistic(z), with z measured against
// the kind's mean and stddev — or against the kind's noise floor when the distribution
// can't support stats.
func ScoreHits(hits []Hit) []Hit {
	byKind := map[Kind][]Hit{}
	for _, h := range hits {
		byKind[h.Kind] = append(byKind[h.Kind], h)
	}

	out := make([]Hit, 0, len(hits))

	for kind, kindHits := range byKind {
		mean, sigma := scoreStats(kindHits)
		statsUsable := len(kindHits) >= minKindHits && sigma >= minKindSigma

		for _, h := range kindHits {
			if h.Score < kindFloor[kind] {
				continue
			}

			z := (h.Score - kindFloor[kind]) / fallbackSigma

			if statsUsable {
				z = (h.Score - mean) / sigma
				if z < zKeep {
					continue
				}
			}

			h.Score = logistic(z)
			out = append(out, h)
		}
	}

	return out
}

// scoreStats returns the mean and population standard deviation of the hits' scores.
func scoreStats(hits []Hit) (mean, sigma float64) {
	if len(hits) == 0 {
		return 0, 0
	}

	for _, h := range hits {
		mean += h.Score
	}

	mean /= float64(len(hits))

	var variance float64

	for _, h := range hits {
		variance += (h.Score - mean) * (h.Score - mean)
	}

	variance /= float64(len(hits))

	return mean, math.Sqrt(variance)
}

// logistic squashes z onto (0, 1).
func logistic(z float64) float64 {
	return 1 / (1 + math.Exp(-z))
}
