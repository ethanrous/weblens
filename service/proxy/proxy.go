package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
)

type Request struct {
	method  string
	remote  *models.Instance
	req     *http.Request
	url     string
	body    any
	queries [][]string
}

func NewRequest(remote *models.Instance, method, endpoint string) Request {
	reqUrl, err := url.JoinPath(remote.Address, "/api/core", endpoint)
	if err != nil {
		log.ErrTrace(err)
		return Request{}
	}

	return Request{method: method, remote: remote, url: reqUrl}
}

func (r Request) WithQuery(key, val string) Request {
	r.queries = append(r.queries, []string{key, val})
	return r
}

func (r Request) WithBody(body any) Request {
	r.body = body
	return r
}

func (r Request) Call() (*http.Response, error) {
	if r.remote.UsingKey == "" {
		return nil, werror.Errorf("Trying to dial core without api key")
	}
	if len(r.url) == 0 {
		return nil, werror.Errorf("Trying to dial core without endpoint")
	}

	buf := &bytes.Buffer{}
	if r.body != nil {
		bs, err := json.Marshal(r.body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(bs)
	}

	req, err := http.NewRequest(r.method, r.url, buf)
	if err != nil {
		return nil, err
	}

	if len(r.queries) != 0 {
		q := req.URL.Query()
		for _, query := range r.queries {
			q.Add(query[0], query[1])
		}
		req.URL.RawQuery = q.Encode()
	}

	req.Header.Add("Authorization", "Bearer "+string(r.remote.UsingKey))
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, werror.Errorf("Failed to call home to [%s]: %s", req.URL.String(), resp.Status)
	}

	return resp, err
}

func CallHomeStruct[T any](req Request) (T, error) {
	res, err := req.Call()

	var target T
	if err != nil {
		return target, err
	}

	defer res.Body.Close()

	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return target, werror.WithStack(err)
	}

	err = json.Unmarshal(bs, &target)
	return target, werror.WithStack(err)
}
