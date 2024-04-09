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
}

func NewRequester() types.Requester {
	return &requester{}
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
	body := gin.H{"name": name, "usingKey": apiKey}
	bs, err := json.Marshal(body)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(bs)
	req, err := http.NewRequest("POST", coreAddress+"/api/remote", buf)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+string(apiKey))
	cli := &http.Client{}
	resp, err := cli.Do(req)
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
	address, err := dataStore.GetServerInfo().GetCoreAddress()
	if err != nil {
		util.ShowErr(err)
		return err
	}

	resp, err := http.Get(address + "/snapshot")
	if err != nil {
		return err
	}

	allFiles, err := readRespBody[filesResp](resp)
	if err != nil {
		return err
	}

	util.Debug.Println(util.MapToKeys(allFiles.Files))

	return nil
}
