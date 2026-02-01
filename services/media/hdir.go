// Package media provides functionalities related to media processing and management.
package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/slices"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

// ErrNoSimilarity indicates that no images matched the search text.
var ErrNoSimilarity = wlerrors.Statusf(http.StatusNotFound, "no images matched the search text")

// ScoreWrapper represents a media item along with its similarity score.
type ScoreWrapper struct {
	Media *media_model.Media
	Score float64
}

// Returns the highest values above the largest significant gap in the array.
// If no significant gap, returns all values.
func skimTop(values []ScoreWrapper) []ScoreWrapper {
	breakIndex := -1

	for i := range values {
		if i == 0 {
			continue
		}

		gap := values[i-1].Score - values[i].Score
		if gap > 0.01 {
			breakIndex = i

			break
		}
	}

	if breakIndex != -1 {
		return values[:breakIndex]
	}

	return values
}

// SortMediaByTextSimilarity sorts media items by their similarity to the given text.
func SortMediaByTextSimilarity(ctx context_service.AppContext, search string, ms []*media_model.Media, minScore float64) ([]ScoreWrapper, error) {
	if len(search) == 0 {
		return []ScoreWrapper{}, nil
	}

	textTarget, err := GetHighDimensionTextEncoding(ctx, search)
	if err != nil {
		return nil, err
	}

	scores, err := media_model.HDIRSearch(ctx, textTarget, ms)
	if err != nil {
		return nil, err
	}

	if len(scores) == 0 {
		return []ScoreWrapper{}, nil
	}

	mediasMap := make(map[string]*media_model.Media, len(ms))
	for _, m := range ms {
		mediasMap[m.ContentID] = m
	}

	msScores := slices.MapI(scores, func(score media_model.HDIRSearchResult, _ int) ScoreWrapper {
		return ScoreWrapper{
			Media: mediasMap[score.ContentID],
			Score: score.Score,
		}
	})

	ctx.Log().Debug().Msgf("Similarity scores: Top: %.6f Bottom: %.6f", msScores[0].Score, msScores[len(msScores)-1].Score)

	msScores = slices.Filter(msScores, func(ms ScoreWrapper) bool {
		return ms.Score >= minScore
	})

	slices.SortFunc(msScores, func(a, b ScoreWrapper) int {
		if a.Score < b.Score {
			return 1
		} else if a.Score > b.Score {
			return -1
		}

		return 0
	})

	if len(msScores) == 0 {
		return nil, wlerrors.WithStack(ErrNoSimilarity)
	}

	msScores = skimTop(msScores)

	return msScores, nil
}

var serviceUnavailable atomic.Bool = atomic.Bool{}

// GetHighDimensionImageEncoding retrieves the high-dimensional image encoding for the given media.
func GetHighDimensionImageEncoding(ctx context_service.AppContext, m *media_model.Media) ([]float64, error) {
	if serviceUnavailable.Load() {
		return nil, wlerrors.WithStack(wlerrors.Statusf(http.StatusServiceUnavailable, "HDIR service is not available"))
	}

	f, err := getCacheFile(ctx, m, media_model.LowRes, 0)
	if err != nil {
		return nil, err
	}

	hdirServerURL := config.GetConfig().HdirURI

	resp, err := http.Get(hdirServerURL + "/encode?img-path=" + f.GetPortablePath().String())
	if err != nil {
		if strings.Contains(err.Error(), "no such host") { // If the HDIR server is not available, we don't retry
			serviceUnavailable.Store(true)
		}

		return nil, wlerrors.Errorf("Failed to get HDIR encoding at %s: %w", hdirServerURL, err)
	}

	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, wlerrors.Errorf("Error reading HDIR response: %w", err)
	}

	target := []float64{}

	err = json.Unmarshal(body, &target)
	if err != nil {
		return nil, wlerrors.Errorf("Error unmarshalling HDIR response: %s: %w", body, err)
	}

	m.HDIR = target

	return target, nil
}

var hdirTextEncodingCacheKey = "hdir_text_encoding"

// GetHighDimensionTextEncoding retrieves the high-dimensional text encoding for the given text.
func GetHighDimensionTextEncoding(ctx context_service.AppContext, text string) ([]float64, error) {
	if serviceUnavailable.Load() {
		return nil, wlerrors.WithStack(wlerrors.Statusf(http.StatusServiceUnavailable, "HDIR service is not available"))
	}

	cache := ctx.GetCache(hdirTextEncodingCacheKey)
	if cached, found := cache.Get(text); found {
		ctx.Log().Trace().Msgf("HDIR text encoding cache hit for text: %s", text)

		return cached.([]float64), nil
	}

	hdirServerURL := config.GetConfig().HdirURI

	reqBody := bytes.NewBuffer((fmt.Appendf(nil, `{"text": "%s"}`, text)))

	resp, err := http.Post(hdirServerURL+"/encode-text", "application/json", reqBody)
	if err != nil {
		if strings.Contains(err.Error(), "no such host") { // If the HDIR server is not available, we don't retry
			serviceUnavailable.Store(true)
		}

		return nil, wlerrors.Errorf("Failed to get HDIR text encoding at %s: %w", hdirServerURL, err)
	}

	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, wlerrors.Errorf("Error reading HDIR text response: %w", err)
	}

	target := struct {
		TextFeatures []float64 `json:"text_features"`
	}{}

	err = json.Unmarshal(body, &target)
	if err != nil {
		return nil, wlerrors.Errorf("Error unmarshalling HDIR text response: %s: %w", body, err)
	}

	cache.Set(text, target.TextFeatures)

	return target.TextFeatures, nil
}

// func getSimilarityScores(ctx context_service.AppContext, text string, m ...*media_model.Media) ([]float64, error) {
// 	hdirs := [][]float64{}
//
// 	for _, media := range m {
// 		if len(media.HDIR) == 0 {
// 			hdir, err := GetHighDimensionImageEncoding(ctx, media)
// 			if err != nil {
// 				return nil, wlerrors.WithStack(err)
// 			}
//
// 			hdirs = append(hdirs, hdir)
// 		} else {
// 			hdirs = append(hdirs, media.HDIR)
// 		}
// 	}
//
// 	hdirBytes, err := json.Marshal(hdirs)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	reqBody := bytes.NewBuffer((fmt.Appendf(nil, `{"text": "%s", "image_features": %s}`, text, hdirBytes)))
//
// 	hdieServerURL := config.GetConfig().HdirURI
//
// 	resp, err := http.Post(hdieServerURL+"/match", "application/json", reqBody)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	defer resp.Body.Close() //nolint:errcheck
//
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	target := struct {
// 		Similarity []float64 `json:"similarity"`
// 	}{}
//
// 	err = json.Unmarshal(body, &target)
// 	if err != nil {
// 		ctx.Log().Debug().Msgf("Error unmarshalling similarity response: %s", reqBody)
//
// 		return nil, err
// 	}
//
// 	return target.Similarity, nil
// }
