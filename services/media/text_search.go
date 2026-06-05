package media

import (
	"github.com/ethanrous/weblens/models/embedding"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/wlslices"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/embed"
)

// ScoreWrapper represents a media item along with its similarity score.
type ScoreWrapper struct {
	Media *media_model.Media
	Score float64
}

// SortMediaByTextSimilarity ranks media by cosine similarity to the query, dropping hits below minScore.
func SortMediaByTextSimilarity(ctx context_service.AppContext, search string, ms []*media_model.Media, minScore float64) ([]ScoreWrapper, error) {
	if len(search) == 0 || len(ms) == 0 {
		return []ScoreWrapper{}, nil
	}

	client := embed.Default()
	if client.ServiceUnavailable() {
		return []ScoreWrapper{}, nil
	}

	_, vec, err := client.EncodeQueryText(ctx, search)
	if err != nil {
		return nil, err
	}

	mediaByContentID := make(map[string]*media_model.Media, len(ms))
	sourceIDs := make([]string, 0, len(ms))

	for _, m := range ms {
		mediaByContentID[string(m.ContentID)] = m
		sourceIDs = append(sourceIDs, string(m.ContentID))
	}

	hits, err := embedding.Search(ctx, embedding.Query{
		Vector:    vec,
		Kind:      embedding.KindImage,
		SourceIDs: sourceIDs,
		Limit:     len(ms),
	})
	if err != nil {
		return nil, err
	}

	scored := make([]ScoreWrapper, 0, len(hits))

	for _, h := range hits {
		m, ok := mediaByContentID[h.SourceID]
		if !ok {
			continue
		}

		if h.Score < minScore {
			continue
		}

		scored = append(scored, ScoreWrapper{Media: m, Score: h.Score})
	}

	wlslices.SortFunc(scored, func(a, b ScoreWrapper) int {
		if a.Score < b.Score {
			return 1
		} else if a.Score > b.Score {
			return -1
		}

		return 0
	})

	return scored, nil
}
