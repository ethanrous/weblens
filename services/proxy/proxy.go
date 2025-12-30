// Package proxy provides HTTP request utilities for communicating with remote tower instances.
package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/rs/zerolog/log"
)

// Request represents an HTTP request builder for communicating with remote tower instances.
type Request struct {
	err     error
	remote  *tower_model.Instance
	req     *http.Request
	method  string
	url     string
	body    []byte
	queries [][]string
	headers [][]string
}

// NewCoreRequest creates a new request builder for a core server API call.
func NewCoreRequest(remote *tower_model.Instance, method, endpoint string) Request {
	reqURL, err := url.JoinPath(remote.Address, "/api", endpoint)
	if err != nil {
		log.Error().Stack().Err(err).Msg("")

		return Request{}
	}

	return Request{method: method, remote: remote, url: reqURL}
}

// WithQuery adds a query parameter to the request.
func (r Request) WithQuery(key, val string) Request {
	r.queries = append(r.queries, []string{key, val})

	return r
}

// WithHeader adds a header to the request.
func (r Request) WithHeader(key, val string) Request {
	r.headers = append(r.headers, []string{key, val})

	return r
}

// OverwriteEndpoint replaces the request endpoint with a new one.
func (r Request) OverwriteEndpoint(newEndpoint string) Request {
	reqURL, err := url.JoinPath(r.remote.Address, newEndpoint)
	if err != nil {
		log.Error().Stack().Err(err).Msg("")

		return Request{}
	}

	r.url = reqURL

	return r
}

// WithBody sets the request body by marshaling the provided value to JSON.
func (r Request) WithBody(body any) Request {
	bs, err := json.Marshal(body)
	if err != nil {
		r.err = wlerrors.WithStack(err)

		return r
	}

	r.body = bs

	return r
}

// WithBodyBytes sets the request body from raw bytes.
func (r Request) WithBodyBytes(bodyBytes []byte) Request {
	r.body = bodyBytes

	return r
}

// ErrorInfo represents an error response from a remote server.
type ErrorInfo struct {
	Error string `json:"error"`
}

// Call executes the HTTP request and returns the response.
func (r Request) Call() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}

	if r.remote.OutgoingKey == "" {
		return nil, wlerrors.Errorf("Trying to dial core without api key")
	}

	if len(r.url) == 0 {
		return nil, wlerrors.Errorf("Trying to dial core without endpoint")
	}

	buf := bytes.NewBuffer(r.body)

	req, err := http.NewRequest(r.method, r.url, buf)
	if err != nil {
		return nil, err
	}

	for _, header := range r.headers {
		req.Header.Add(header[0], header[1])
	}

	if len(r.queries) != 0 {
		q := req.URL.Query()
		for _, query := range r.queries {
			q.Add(query[0], query[1])
		}

		req.URL.RawQuery = q.Encode()
	}

	req.Header.Add("Authorization", "Bearer "+string(r.remote.OutgoingKey))
	// req.Header.Add("Wl-Server-ID", LocalInstance.ServerID())
	log.Debug().Msgf("Calling home to %s [%s %s] with key [%s]", r.remote.TowerID, r.method, req.URL.String(), r.remote.OutgoingKey)

	cli := &http.Client{}

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close() //nolint:errcheck

		target := ErrorInfo{}

		bs, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Stack().Err(err).Msg("")

			return nil, wlerrors.Errorf("Failed to call home to [%s %s]: %s", r.method, req.URL.String(), resp.Status)
		}

		err = json.Unmarshal(bs, &target)
		if err != nil {
			return nil, wlerrors.Errorf("Failed to call home to [%s %s]: %s", r.method, req.URL.String(), resp.Status)
		}

		return nil, wlerrors.Errorf("Failed to call home to [%s %s]: %s", r.method, req.URL.String(), target.Error)
	}

	return resp, err
}

// CallHomeStruct executes the request and unmarshals the response into the specified type.
func CallHomeStruct[T any](req Request) (T, error) {
	res, err := req.Call()

	var target T
	if err != nil {
		return target, err
	}

	defer res.Body.Close() //nolint:errcheck

	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return target, wlerrors.WithStack(err)
	}

	err = json.Unmarshal(bs, &target)

	return target, wlerrors.WithStack(err)
}
