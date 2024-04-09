package routes

import (
	"net/http"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
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
