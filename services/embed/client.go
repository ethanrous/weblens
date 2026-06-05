// Package embed is the Go-side client for the weblens-embed service.
package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ChunkResult is one entry returned from /extract-and-embed; Page is the 1-indexed source page.
type ChunkResult struct {
	ChunkIndex int       `json:"chunkIndex"`
	Page       int       `json:"page,omitempty"`
	Snippet    string    `json:"snippet"`
	Vector     []float64 `json:"vector"`
}

// queryVectors holds the two encodings of a single query text.
type queryVectors struct {
	Plain []float64
	Image []float64
}

// Client talks to the weblens-embed container.
type Client struct {
	baseURL string
	http    *http.Client

	unavailable atomic.Bool

	queryCacheMu sync.RWMutex
	queryCache   map[string]queryVectors
}

// ErrServiceUnavailable indicates the embed container is offline.
var ErrServiceUnavailable = fmt.Errorf("embed service unavailable")

// NewClient constructs a Client. baseURL is the http://host:port of the embed container.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		http:       &http.Client{Timeout: 5 * time.Minute},
		queryCache: make(map[string]queryVectors),
	}
}

// ServiceUnavailable reports whether the container has been marked offline.
func (c *Client) ServiceUnavailable() bool { return c.unavailable.Load() }

// MarkAvailable clears the unavailable flag — used by the health-check ticker.
func (c *Client) MarkAvailable() { c.unavailable.Store(false) }

// MarkUnavailable sets the unavailable flag — used in tests to simulate a downed service.
func (c *Client) MarkUnavailable() { c.unavailable.Store(true) }

// SetBaseURLForTesting overrides the base URL of the client. Only for use in tests.
func (c *Client) SetBaseURLForTesting(url string) { c.baseURL = url }

// BaseURL returns the configured base URL — used by the health-check ticker.
func (c *Client) BaseURL() string { return c.baseURL }

// HTTPClient returns the underlying http.Client — used by the health-check ticker.
func (c *Client) HTTPClient() *http.Client { return c.http }

// EncodeImage returns the unified-model embedding for the image at imgPath.
func (c *Client) EncodeImage(ctx context.Context, imgPath string) ([]float64, error) {
	if c.ServiceUnavailable() {
		return nil, ErrServiceUnavailable
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/encode?img-path="+imgPath, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		c.flagIfNoHost(err)

		return nil, fmt.Errorf("embed encode image: %w", err)
	}

	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("embed encode image: status %d: %s", resp.StatusCode, body)
	}

	var vec []float64
	if err := json.NewDecoder(resp.Body).Decode(&vec); err != nil {
		return nil, fmt.Errorf("embed encode image decode: %w", err)
	}

	return vec, nil
}

// EncodeQueryText returns the plain and caption-prompted query embeddings, cached by exact text.
func (c *Client) EncodeQueryText(ctx context.Context, text string) (plain, image []float64, err error) {
	c.queryCacheMu.RLock()

	if v, ok := c.queryCache[text]; ok {
		c.queryCacheMu.RUnlock()

		return v.Plain, v.Image, nil
	}

	c.queryCacheMu.RUnlock()

	if c.ServiceUnavailable() {
		return nil, nil, ErrServiceUnavailable
	}

	body, _ := json.Marshal(map[string]string{"text": text})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/encode-text", bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		c.flagIfNoHost(err)

		return nil, nil, fmt.Errorf("embed encode text: %w", err)
	}

	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		rb, _ := io.ReadAll(resp.Body)

		return nil, nil, fmt.Errorf("embed encode text: status %d: %s", resp.StatusCode, rb)
	}

	var out struct {
		TextFeatures       []float64 `json:"text_features"`
		ImageQueryFeatures []float64 `json:"image_query_features"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, nil, fmt.Errorf("embed encode text decode: %w", err)
	}

	c.queryCacheMu.Lock()
	c.queryCache[text] = queryVectors{Plain: out.TextFeatures, Image: out.ImageQueryFeatures}
	c.queryCacheMu.Unlock()

	return out.TextFeatures, out.ImageQueryFeatures, nil
}

// ExtractAndEmbedFile invokes /extract-and-embed and returns per-chunk results (nil on 422/404).
func (c *Client) ExtractAndEmbedFile(ctx context.Context, path string, mimeHint string) ([]ChunkResult, error) {
	if c.ServiceUnavailable() {
		return nil, ErrServiceUnavailable
	}

	body, _ := json.Marshal(map[string]string{"path": path, "mimeHint": mimeHint})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/extract-and-embed", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		c.flagIfNoHost(err)

		return nil, fmt.Errorf("embed extract: %w", err)
	}

	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode == http.StatusUnprocessableEntity || resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		rb, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("embed extract: status %d: %s", resp.StatusCode, rb)
	}

	var out []ChunkResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("embed extract decode: %w", err)
	}

	return out, nil
}

func (c *Client) flagIfNoHost(err error) {
	if err != nil && strings.Contains(err.Error(), "no such host") {
		c.unavailable.Store(true)
	}
}
