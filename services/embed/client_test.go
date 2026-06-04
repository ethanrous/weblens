package embed_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethanrous/weblens/services/embed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeQueryText_Cached(t *testing.T) {
	var calls int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(map[string]any{
			"text_features":        []float64{0.1, 0.2, 0.3},
			"image_query_features": []float64{0.4, 0.5, 0.6},
		})
	}))
	defer srv.Close()

	c := embed.NewClient(srv.URL)
	v1, _, err := c.EncodeQueryText(context.Background(), "hello")
	require.NoError(t, err)
	v2, _, err := c.EncodeQueryText(context.Background(), "hello")
	require.NoError(t, err)
	assert.Equal(t, v1, v2)
	assert.Equal(t, 1, calls, "second call should hit cache, not server")
}

func TestServiceUnavailable_BlocksFurtherCalls(t *testing.T) {
	c := embed.NewClient("http://nonexistent-host-for-test:1")
	_, _, err := c.EncodeQueryText(context.Background(), "x")
	require.Error(t, err)
	assert.True(t, c.ServiceUnavailable(), "no-such-host should flip the unavailable flag")

	_, _, err = c.EncodeQueryText(context.Background(), "y")
	require.Error(t, err)
}

func TestEncodeQueryText_ReturnsBothVectors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string][]float64{
			"text_features":        {0.1, 0.2},
			"image_query_features": {0.3, 0.4},
		})
	}))
	defer srv.Close()

	c := embed.NewClient(srv.URL)

	plain, image, err := c.EncodeQueryText(context.Background(), "soc2")
	require.NoError(t, err)
	assert.Equal(t, []float64{0.1, 0.2}, plain)
	assert.Equal(t, []float64{0.3, 0.4}, image)
}
