package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
)

type Request struct {
	method  string
	remote  *models.Instance
	req     *http.Request
	url     string
	body    []byte
	queries [][]string
	err     error
}

func NewCoreRequest(remote *models.Instance, method, endpoint string) Request {
	reqUrl, err := url.JoinPath(remote.Address, "/api", endpoint)
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

func (r Request) OverwriteEndpoint(newEndpoint string) Request {
	reqUrl, err := url.JoinPath(r.remote.Address, newEndpoint)
	if err != nil {
		log.ErrTrace(err)
		return Request{}
	}
	r.url = reqUrl
	return r
}

func (r Request) WithBody(body any) Request {
	bs, err := json.Marshal(body)
	if err != nil {
		r.err = werror.WithStack(err)
		return r
	}
	r.body = bs
	return r
}

func (r Request) WithBodyBytes(bodyBytes []byte) Request {
	r.body = bodyBytes
	return r
}

func (r Request) Call() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}

	if r.remote.UsingKey == "" {
		return nil, werror.Errorf("Trying to dial core without api key")
	}
	if len(r.url) == 0 {
		return nil, werror.Errorf("Trying to dial core without endpoint")
	}

	buf := bytes.NewBuffer(r.body)
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
		log.TraceCaller(2, "Call home failed from here")

		defer resp.Body.Close()
		target := rest.WeblensErrorInfo{}
		bs, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, werror.Errorf("Failed to call home to [%s %s]: %s", r.method, req.URL.String(), resp.Status)
		}

		err = json.Unmarshal(bs, &target)
		if err != nil {
			return nil, werror.Errorf("Failed to call home to [%s %s]: %s", r.method, req.URL.String(), resp.Status)
		}

		return nil, werror.Errorf("Failed to call home to [%s %s]: %s", r.method, req.URL.String(), target.Error)
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
