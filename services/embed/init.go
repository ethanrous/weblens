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

			if probeHealth(ctx, c) {
				c.MarkAvailable()
			}
		}
	}
}

func probeHealth(ctx context.Context, c *Client) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL()+"/health", nil)
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
