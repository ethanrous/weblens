package embedding

import (
	"math"
	"testing"
)

func TestAtlasScoreToCosine(t *testing.T) {
	cases := []struct {
		atlas float64
		want  float64
	}{
		{atlas: 1.0, want: 1.0},
		{atlas: 0.5, want: 0.0},
		{atlas: 0.0, want: -1.0},
		{atlas: 0.65, want: 0.30},
	}

	for _, c := range cases {
		got := atlasScoreToCosine(c.atlas)
		if math.Abs(got-c.want) > 1e-9 {
			t.Fatalf("atlasScoreToCosine(%v) = %v, want %v", c.atlas, got, c.want)
		}
	}
}
