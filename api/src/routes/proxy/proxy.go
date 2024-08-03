package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ethrousseau/weblens/api/types"
)

type ProxyStore struct {
	coreAddress string
	apiKey      types.WeblensApiKey
	db          types.StoreService
}

func NewProxyStore(coreAddress string, apiKey types.WeblensApiKey) *ProxyStore {
	return &ProxyStore{
		coreAddress: coreAddress,
		apiKey:      apiKey,
	}
}

func (p *ProxyStore) Init(db types.StoreService) {
	core := types.SERV.InstanceService.GetCore()
	addr, err := core.GetAddress()
	if err != nil {
		panic(err)
	}
	p.coreAddress = addr
	p.apiKey = core.GetUsingKey()
	p.db = db
}

func (p *ProxyStore) GetLocalStore() types.StoreService {
	return p.db
}

func ReadResponseBody[T any](resp *http.Response) (T, error) {
	var target T

	if resp.StatusCode > 299 {
		return target, types.WeblensErrorMsg(
			fmt.Sprintf(
				"Trying to read response body of call to %s with bad status code: %d",
				resp.Request.URL.String(), resp.StatusCode,
			),
		)
	}

	bs, err := io.ReadAll(resp.Body)

	if err != nil {
		return target, types.WeblensErrorFromError(err)
	}

	err = resp.Body.Close()
	if err != nil {
		return target, types.WeblensErrorFromError(err)
	}

	// If the requester just wants the bytes, skip unmarshaling
	switch any(target).(type) {
	case []byte:
		return any(bs).(T), nil
	}

	err = json.Unmarshal(bs, &target)
	if err != nil {
		return target, types.WeblensErrorFromError(err)
	}

	return target, nil
}

func (p *ProxyStore) CallHome(method, endpoint string, body any) (*http.Response, error) {
	if p.coreAddress == "" {
		return nil, types.WeblensErrorMsg("Trying to dial core with no address")
	}
	if p.apiKey == "" {
		return nil, types.WeblensErrorMsg("Trying to dial core without api key")
	}

	buf := &bytes.Buffer{}
	if body != nil {
		bs, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(bs)
	}

	// Remove the leading slash from the endpoint
	if endpoint[:1] == "/" {
		endpoint = endpoint[1:]
	}

	fullAddr := p.coreAddress + endpoint

	req, err := http.NewRequest(method, fullAddr, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+string(p.apiKey))
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}
