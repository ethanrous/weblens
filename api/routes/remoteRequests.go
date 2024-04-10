package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

type requester struct {
	ApiKey      types.WeblensApiKey
	CoreAddress string
}

func NewRequester() types.Requester {
	srvInfo := dataStore.GetServerInfo()
	if srvInfo == nil {
		return &requester{}
	}
	addr, _ := srvInfo.GetCoreAddress()

	return &requester{
		ApiKey:      srvInfo.GetUsingKey(),
		CoreAddress: addr,
	}
}

func (r *requester) coreRequest(method string, addrExt string, body any) (*http.Response, error) {
	if r.CoreAddress == "" {
		return nil, ErrNoAddress
	}
	if r.ApiKey == "" {
		return nil, ErrNoKey
	}

	buf := &bytes.Buffer{}
	if body != nil {
		bs, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(bs)
	}
	req, err := http.NewRequest(method, r.CoreAddress+addrExt, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+string(r.ApiKey))
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func PingCore(coreAddress string) error {
	thisServer := dataStore.GetServerInfo()
	var err error
	if thisServer.ServerRole() == types.CoreMode {
		return ErrCoreOriginate
	}
	if coreAddress == "" {
		coreAddress, err = dataStore.GetServerInfo().GetCoreAddress()
		if err != nil {
			util.ShowErr(err)
			return err
		}
	}

	http.Get(coreAddress + "/ping")
	return nil
}

type filesResp struct {
	Files map[string]types.WeblensFile `json:"files"`
}

func (r *requester) AttachToCore(coreAddress string, name string, apiKey types.WeblensApiKey) error {
	r.CoreAddress = coreAddress
	r.ApiKey = apiKey

	body := gin.H{"name": name, "usingKey": apiKey}
	resp, err := r.coreRequest("POST", "/api/remote", body)
	if err != nil {
		return err
	}

	if resp.StatusCode == 201 {
		return nil
	} else {
		return errors.New("failed to attch to remote core")
	}
}

func (r *requester) GetCoreSnapshot() error {
	resp, err := r.coreRequest("GET", "/api/snapshot", nil)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		err = errors.New("bad status: " + resp.Status)
		return err
	}

	allFiles, err := readRespBody[filesResp](resp)
	if err != nil {
		return err
	}

	util.Debug.Println(util.MapToKeys(allFiles.Files))

	return nil
}
