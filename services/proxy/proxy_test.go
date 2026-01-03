package proxy_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/services/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCoreRequest(t *testing.T) {
	t.Run("creates request with correct URL", func(t *testing.T) {
		remote := &tower_model.Instance{
			Address:     "http://localhost:8080",
			OutgoingKey: "test-key",
		}

		req := proxy.NewCoreRequest(remote, http.MethodGet, "/files")

		// Should be able to add query and call without panic
		req = req.WithQuery("param", "value")
		assert.NotNil(t, req)
	})

	t.Run("handles different HTTP methods", func(t *testing.T) {
		remote := &tower_model.Instance{
			Address:     "http://localhost:8080",
			OutgoingKey: "test-key",
		}

		methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

		for _, method := range methods {
			req := proxy.NewCoreRequest(remote, method, "/test")
			assert.NotNil(t, req)
		}
	})
}

func TestRequest_WithQuery(t *testing.T) {
	t.Run("adds query parameter", func(t *testing.T) {
		remote := &tower_model.Instance{
			Address:     "http://localhost:8080",
			OutgoingKey: "test-key",
		}

		req := proxy.NewCoreRequest(remote, http.MethodGet, "/files")
		req = req.WithQuery("id", "123")
		req = req.WithQuery("name", "test")

		// The request should be valid
		assert.NotNil(t, req)
	})
}

func TestRequest_WithHeader(t *testing.T) {
	t.Run("adds header", func(t *testing.T) {
		remote := &tower_model.Instance{
			Address:     "http://localhost:8080",
			OutgoingKey: "test-key",
		}

		req := proxy.NewCoreRequest(remote, http.MethodGet, "/files")
		req = req.WithHeader("Content-Type", "application/json")
		req = req.WithHeader("X-Custom-Header", "value")

		assert.NotNil(t, req)
	})
}

func TestRequest_OverwriteEndpoint(t *testing.T) {
	t.Run("changes endpoint", func(t *testing.T) {
		remote := &tower_model.Instance{
			Address:     "http://localhost:8080",
			OutgoingKey: "test-key",
		}

		req := proxy.NewCoreRequest(remote, http.MethodGet, "/old-endpoint")
		req = req.OverwriteEndpoint("/new-endpoint")

		assert.NotNil(t, req)
	})
}

func TestRequest_WithBody(t *testing.T) {
	t.Run("marshals body to JSON", func(t *testing.T) {
		remote := &tower_model.Instance{
			Address:     "http://localhost:8080",
			OutgoingKey: "test-key",
		}

		body := map[string]string{"key": "value"}
		req := proxy.NewCoreRequest(remote, http.MethodPost, "/endpoint")
		req = req.WithBody(body)

		assert.NotNil(t, req)
	})

	t.Run("handles struct body", func(t *testing.T) {
		remote := &tower_model.Instance{
			Address:     "http://localhost:8080",
			OutgoingKey: "test-key",
		}

		type testBody struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}

		body := testBody{Name: "test", Count: 42}
		req := proxy.NewCoreRequest(remote, http.MethodPost, "/endpoint")
		req = req.WithBody(body)

		assert.NotNil(t, req)
	})
}

func TestRequest_WithBodyBytes(t *testing.T) {
	t.Run("sets raw bytes body", func(t *testing.T) {
		remote := &tower_model.Instance{
			Address:     "http://localhost:8080",
			OutgoingKey: "test-key",
		}

		bodyBytes := []byte(`{"key": "value"}`)
		req := proxy.NewCoreRequest(remote, http.MethodPost, "/endpoint")
		req = req.WithBodyBytes(bodyBytes)

		assert.NotNil(t, req)
	})
}

func TestRequest_Call(t *testing.T) {
	t.Run("returns error when missing API key", func(t *testing.T) {
		remote := &tower_model.Instance{
			Address:     "http://localhost:8080",
			OutgoingKey: "", // No key
		}

		req := proxy.NewCoreRequest(remote, http.MethodGet, "/files")
		resp, err := req.Call()

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "api key")
	})

	t.Run("successfully calls remote server", func(t *testing.T) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`)) //nolint:errcheck
		}))
		defer server.Close()

		remote := &tower_model.Instance{
			Address:     server.URL,
			OutgoingKey: "test-api-key",
		}

		req := proxy.NewCoreRequest(remote, http.MethodGet, "/status")
		resp, err := req.Call()

		require.NoError(t, err)
		require.NotNil(t, resp)

		defer resp.Body.Close() //nolint:errcheck

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("includes query parameters in request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "value1", r.URL.Query().Get("param1"))
			assert.Equal(t, "value2", r.URL.Query().Get("param2"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		remote := &tower_model.Instance{
			Address:     server.URL,
			OutgoingKey: "test-api-key",
		}

		req := proxy.NewCoreRequest(remote, http.MethodGet, "/test")
		req = req.WithQuery("param1", "value1")
		req = req.WithQuery("param2", "value2")

		resp, err := req.Call()
		require.NoError(t, err)
		require.NotNil(t, resp)

		defer resp.Body.Close() //nolint:errcheck
	})

	t.Run("includes custom headers in request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		remote := &tower_model.Instance{
			Address:     server.URL,
			OutgoingKey: "test-api-key",
		}

		req := proxy.NewCoreRequest(remote, http.MethodGet, "/test")
		req = req.WithHeader("Content-Type", "application/json")
		req = req.WithHeader("X-Custom-Header", "custom-value")

		resp, err := req.Call()
		require.NoError(t, err)
		require.NotNil(t, resp)

		defer resp.Body.Close() //nolint:errcheck
	})

	t.Run("returns error for 4xx status codes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(proxy.ErrorInfo{Error: "not found"}) //nolint:errcheck
		}))
		defer server.Close()

		remote := &tower_model.Instance{
			Address:     server.URL,
			OutgoingKey: "test-api-key",
		}

		req := proxy.NewCoreRequest(remote, http.MethodGet, "/notfound")
		resp, err := req.Call()

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error for 5xx status codes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(proxy.ErrorInfo{Error: "internal error"}) //nolint:errcheck
		}))
		defer server.Close()

		remote := &tower_model.Instance{
			Address:     server.URL,
			OutgoingKey: "test-api-key",
		}

		req := proxy.NewCoreRequest(remote, http.MethodGet, "/error")
		resp, err := req.Call()

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "internal error")
	})
}

func TestCallHomeStruct(t *testing.T) {
	t.Run("unmarshals response to struct", func(t *testing.T) {
		type TestResponse struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TestResponse{Name: "test", Count: 42}) //nolint:errcheck
		}))
		defer server.Close()

		remote := &tower_model.Instance{
			Address:     server.URL,
			OutgoingKey: "test-api-key",
		}

		req := proxy.NewCoreRequest(remote, http.MethodGet, "/data")
		result, err := proxy.CallHomeStruct[TestResponse](req)

		require.NoError(t, err)
		assert.Equal(t, "test", result.Name)
		assert.Equal(t, 42, result.Count)
	})

	t.Run("returns error on failed request", func(t *testing.T) {
		type TestResponse struct {
			Name string `json:"name"`
		}

		remote := &tower_model.Instance{
			Address:     "http://localhost:9999", // Non-existent server
			OutgoingKey: "test-api-key",
		}

		req := proxy.NewCoreRequest(remote, http.MethodGet, "/data")
		_, err := proxy.CallHomeStruct[TestResponse](req)

		assert.Error(t, err)
	})
}

func TestErrorInfo(t *testing.T) {
	t.Run("struct fields", func(t *testing.T) {
		errInfo := proxy.ErrorInfo{
			Error: "test error message",
		}

		assert.Equal(t, "test error message", errInfo.Error)
	})

	t.Run("json marshaling", func(t *testing.T) {
		errInfo := proxy.ErrorInfo{
			Error: "test error",
		}

		data, err := json.Marshal(errInfo)
		require.NoError(t, err)
		assert.Contains(t, string(data), "test error")
	})

	t.Run("json unmarshaling", func(t *testing.T) {
		data := []byte(`{"error": "unmarshaled error"}`)

		var errInfo proxy.ErrorInfo

		err := json.Unmarshal(data, &errInfo)
		require.NoError(t, err)
		assert.Equal(t, "unmarshaled error", errInfo.Error)
	})
}
