package embed_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethanrous/weblens/services/embed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hangingServer replies only once the request's context is cancelled or backstop elapses, simulating a
// stalled sidecar. The backstop bounds test cleanup time: Server.Close waits for active connections to
// finish, and a POST handler that never drains its body can otherwise leave one stuck past client-side abandonment.
func hangingServer(backstop time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(backstop):
		}
	}))
}

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
	c := embed.NewClient("http://127.0.0.1:1")
	_, _, err := c.EncodeQueryText(context.Background(), "x")
	require.Error(t, err)
	assert.True(t, c.ServiceUnavailable(), "an unreachable host should flip the unavailable flag")

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

func TestEncodeImage_RespectsCallerCancellation(t *testing.T) {
	srv := hangingServer(3 * time.Second)
	defer srv.Close()

	c := embed.NewClient(srv.URL)
	c.SetHTTPTimeoutForTesting(2 * time.Second) // safety net so a regression fails fast instead of hanging the test run

	ctx, cancel := context.WithCancel(context.Background())

	start := time.Now()
	done := make(chan error, 1)

	go func() {
		_, err := c.EncodeImage(ctx, "foo.jpg")
		done <- err
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		require.Error(t, err)
		assert.Less(t, time.Since(start), time.Second, "cancelling the caller context should abort the request immediately, not wait for the shared client timeout")
	case <-time.After(4 * time.Second):
		t.Fatal("EncodeImage did not return")
	}
}

func TestFlagUnreachable_ClientTimeoutTripsBreaker(t *testing.T) {
	srv := hangingServer(2 * time.Second)
	defer srv.Close()

	c := embed.NewClient(srv.URL)
	c.SetHTTPTimeoutForTesting(100 * time.Millisecond)

	_, _, err := c.EncodeQueryText(context.Background(), "hang")
	require.Error(t, err)
	assert.True(t, c.ServiceUnavailable(), "a client-side timeout against a live-but-hung sidecar should trip the breaker")
}

func TestFlagUnreachable_CallerCancelDoesNotTripBreaker(t *testing.T) {
	srv := hangingServer(2 * time.Second)
	defer srv.Close()

	c := embed.NewClient(srv.URL)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, _, err := c.EncodeQueryText(ctx, "hang")
	require.Error(t, err)
	assert.False(t, c.ServiceUnavailable(), "the caller cancelling its own context should not be blamed on the service")
}

func TestProbeHealth_DoesNotBlockPastOwnDeadline(t *testing.T) {
	srv := hangingServer(8 * time.Second)
	defer srv.Close()

	c := embed.NewClient(srv.URL)
	c.SetHTTPTimeoutForTesting(6 * time.Second) // safety net above any reasonable probe deadline

	start := time.Now()
	done := make(chan bool, 1)

	go func() {
		done <- embed.ProbeHealth(context.Background(), c)
	}()

	select {
	case ok := <-done:
		assert.False(t, ok)
		assert.Less(t, time.Since(start), 5*time.Second, "the probe should impose its own short deadline instead of riding the shared client timeout")
	case <-time.After(10 * time.Second):
		t.Fatal("ProbeHealth did not return")
	}
}
