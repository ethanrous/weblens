package embed

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/startup"
)

const healthCheckInterval = 30 * time.Second

// healthProbeTimeout bounds a single /health request so a wedged sidecar can never block the health
// loop for longer than this, regardless of the shared client's own (much longer) timeout.
const healthProbeTimeout = 3 * time.Second

var (
	clientOnce sync.Once
	client     *Client
)

// Default returns the process-wide Client, lazily constructed from config.
func Default() *Client {
	clientOnce.Do(func() {
		client = NewClient(config.GetConfig().EmbedURI)
	})

	return client
}

func init() {
	startup.RegisterHook(startHealthTicker)
}

func startHealthTicker(ctx context.Context, _ config.Provider) error {
	c := Default()

	go healthLoop(ctx, c, healthCheckInterval)

	return nil
}

func healthLoop(ctx context.Context, c *Client, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if !c.ServiceUnavailable() {
				continue
			}

			if ProbeHealth(ctx, c) {
				c.MarkAvailable()
			}
		}
	}
}

// ProbeHealth queries /health with its own bounded deadline. Exported so tests can exercise the health loop's probe directly.
func ProbeHealth(ctx context.Context, c *Client) bool {
	reqCtx, cancel := context.WithTimeout(ctx, healthProbeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, c.BaseURL()+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := c.HTTPClient().Do(req)
	if err != nil {
		return false
	}

	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}
