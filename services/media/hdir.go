package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/slices"
	context_service "github.com/ethanrous/weblens/services/context"
)

var ErrNoSimilarity = errors.Statusf(http.StatusNotFound, "no images matched the search text")

type MediaWithScore struct {
	Media *media_model.Media
	Score float64
	Index int
}

// Returns the highest values above the largest significant gap in the array.
// If no significant gap, returns all values.
func skimTop(values []MediaWithScore) []MediaWithScore {
	valueRange := values[0].Score - values[len(values)-1].Score
	minScore := values[len(values)-1].Score + valueRange*0.50 // Only scores in the top 50%
	log.GlobalLogger().Debug().Msgf("Skimming top values with min score: %f %f", values[0].Score, minScore)

	lowestItemIndex := slices.IndexFunc(values, func(ms MediaWithScore) bool {
		return ms.Score < minScore
	})

	values = values[:lowestItemIndex]

	return values
}

func SortMediaByTextSimilarity(ctx context_service.AppContext, search string, ms []*media_model.Media, folderIds []string, minScore float64) ([]MediaWithScore, error) {
	if len(search) == 0 {
		return []MediaWithScore{}, nil
	}

	scores, err := getSimilarityScores(ctx, search, ms...)
	if err != nil {
		return nil, err
	}

	if len(scores) != len(ms) {
		return nil, errors.Errorf("expected %d similarity scores, got %d", len(ms), len(scores))
	}

	msScores := slices.MapI(ms, func(m *media_model.Media, i int) MediaWithScore {
		return MediaWithScore{
			Media: m,
			Score: scores[i],
		}
	})

	msScores = slices.Filter(msScores, func(ms MediaWithScore) bool {
		return ms.Score >= minScore
	})

	slices.SortFunc(msScores, func(a, b MediaWithScore) int {
		if a.Score < b.Score {
			return 1
		} else if a.Score > b.Score {
			return -1
		}

		return 0
	})

	if len(msScores) == 0 {
		return nil, errors.WithStack(ErrNoSimilarity)
	}

	msScores = skimTop(msScores)

	ctx.Log().Debug().Msgf("Similarity scores: %v", msScores)

	return msScores, nil
}

var serviceAvailable = true

func GetHighDimensionImageEncoding(ctx context_service.AppContext, m *media_model.Media) ([]float64, error) {
	if !serviceAvailable {
		return nil, errors.WithStack(errors.Statusf(http.StatusServiceUnavailable, "HDIR service is not available"))
	}

	f, err := getCacheFile(ctx, m, media_model.LowRes, 0)
	if err != nil {
		return nil, err
	}

	hdieServerUrl := config.GetConfig().HdirUri

	resp, err := http.Get(hdieServerUrl + "/encode?img-path=" + f.GetPortablePath().String())
	if err != nil {
		if strings.Contains(err.Error(), "no such host") { // If the HDIR server is not available, we don't retry
			serviceAvailable = false
		}

		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	target := []float64{}

	err = json.Unmarshal(body, &target)
	if err != nil {
		return nil, err
	}

	m.HDIR = target

	return target, nil
}

func getSimilarityScores(ctx context_service.AppContext, text string, m ...*media_model.Media) ([]float64, error) {
	hdirs := [][]float64{}
	for _, media := range m {
		if len(media.HDIR) == 0 {
			hdir, err := GetHighDimensionImageEncoding(ctx, media)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			hdirs = append(hdirs, hdir)
		} else {
			hdirs = append(hdirs, media.HDIR)
		}
	}

	hdirBytes, err := json.Marshal(hdirs)
	if err != nil {
		return nil, err
	}

	reqBody := bytes.NewBuffer((fmt.Appendf(nil, `{"text": "%s", "image_features": %s}`, text, hdirBytes)))

	hdieServerUrl := config.GetConfig().HdirUri

	resp, err := http.Post(hdieServerUrl+"/match", "application/json", reqBody)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	target := struct {
		Similarity []float64 `json:"similarity"`
	}{}

	err = json.Unmarshal(body, &target)
	if err != nil {
		ctx.Log().Debug().Msgf("Error unmarshalling similarity response: %s", reqBody)

		return nil, err
	}

	return target.Similarity, nil
}
