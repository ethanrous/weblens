package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
)

func callHome(remote *models.Instance, method, endpoint string, body any) (*http.Response, error) {
	if remote.UsingKey == "" {
		return nil, werror.Errorf("Trying to dial core without api key")
	}
	if len(endpoint) == 0 {
		return nil, werror.Errorf("Trying to dial core without endpoint")
	}

	buf := &bytes.Buffer{}
	if body != nil {
		bs, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(bs)
	}

	reqUrl, err := url.JoinPath(remote.Address, "/api/core", endpoint)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	req, err := http.NewRequest(method, reqUrl, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+string(remote.UsingKey))
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, werror.Errorf("Failed to call home to [%s]: %s", reqUrl, resp.Status)
	}

	return resp, err
}

func CallHomeStruct[T any](remote *models.Instance, method, endpoint string, body any) (T, error) {
	r, err := callHome(remote, method, endpoint, body)

	var target T
	if err != nil {
		return target, err
	}

	defer r.Body.Close()

	bs, err := io.ReadAll(r.Body)
	if err != nil {
		return target, werror.WithStack(err)
	}

	err = json.Unmarshal(bs, &target)
	return target, werror.WithStack(err)
}
