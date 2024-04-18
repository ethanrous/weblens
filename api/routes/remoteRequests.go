package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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

// Address extensions are of the form `BASE_CORE_ADDRESS` + "/api/core" + `addrExt`.
// So, to make a call to http://mycore.net/api/core/info, you would simply pass "/info" to `addrExt“
func (r *requester) coreRequest(method string, addrExt string, body any, baseOverride ...string) (*http.Response, error) {
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

	fullAddr := ""
	if len(baseOverride) != 0 {
		fullAddr = r.CoreAddress + baseOverride[0] + addrExt
	} else {
		fullAddr = r.CoreAddress + "/api/core" + addrExt
	}
	req, err := http.NewRequest(method, fullAddr, buf)
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

func (r *requester) PingCore() bool {
	_, err := http.Get(r.CoreAddress + "/ping")
	return err == nil
}

// type filesResp struct {
// 	Files dataStore.FileArray `json:"files"`
// }

// func (d *journalResp) UnmarshalJSON(data []byte) error {
// 	var tmp map[string]any
// 	json.Unmarshal(data, &tmp)

// 	// contents := tmp["journal"].([]map[string]any)
// 	decoded := []types.JournalEntry{}

// 	for _, e := range tmp["journal"].([]any) {
// 		je, err := dataStore.IToJE(e.(map[string]any))
// 		if err != nil {
// 			return err
// 		}
// 		decoded = append(decoded, je)
// 	}

// 	d.Journal = decoded

// 	return nil
// }

func (r *requester) AttachToCore(srvId, coreAddress, name string, apiKey types.WeblensApiKey) error {
	r.CoreAddress = coreAddress
	r.ApiKey = apiKey

	body := gin.H{"id": srvId, "name": name, "usingKey": apiKey}
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

func (r *requester) RequestCoreSnapshot() ([]types.FileJournalEntry, error) {
	latest, err := dataStore.GetLatestBackup()
	if err != nil {
		return nil, err
	}
	resp, err := r.coreRequest("GET", "/snapshot?since="+strconv.FormatInt(latest.UnixMilli(), 10), nil)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		err = errors.New("bad status: " + resp.Status)
		return nil, err
	}

	j, err := readRespBody[dataStore.JournalResp](resp)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[types.FileJournalEntry](j.Journal), nil
}

func (r *requester) GetCoreUsers() (us []types.User, err error) {
	resp, err := r.coreRequest("GET", "/users", nil)
	if err != nil {
		return
	}

	body, err := readRespBody[dataStore.UserArray](resp)
	if err != nil {
		return
	}

	return body, nil

}

func (r *requester) GetCoreFileInfos(fIds []types.FileId) ([]types.WeblensFile, error) {
	resp, err := r.coreRequest("GET", "/files", fIds)
	if err != nil {
		return nil, err
	}
	body, err := readRespBody[getFilesResp](resp)
	if err != nil {
		return nil, err
	}
	if len(body.NotFound) != 0 {
		util.Error.Println("Failed to find files at core:", body.NotFound)
	}
	return body.Files, nil
}

func (r *requester) GetCoreFileBin(f types.WeblensFile) ([][]byte, error) {
	resp, err := r.coreRequest("GET", "/file/"+string(f.Id())+"/content", nil)
	if err != nil {
		return nil, err
	}
	bs := [][]byte{}
	origFileBs, err := readRespBodyRaw(resp)
	if err != nil {
		return nil, err
	}
	bs = append(bs, origFileBs)

	if f.IsDisplayable() {
		m, err := f.GetMedia()
		if err != nil {
			return nil, err
		}

		resp, err := r.coreRequest("GET", "/media/"+string(m.Id())+"/thumbnail", nil, "/api")
		if err != nil {
			return nil, err
		}
		thumbBs, err := readRespBodyRaw(resp)
		if err != nil {
			return nil, err
		}
		bs = append(bs, thumbBs)

		resp, err = r.coreRequest("GET", "/media/"+string(m.Id())+"/fullres", nil, "/api")
		if err != nil {
			return nil, err
		}
		fullresBs, err := readRespBodyRaw(resp)
		if err != nil {
			return nil, err
		}
		bs = append(bs, fullresBs)
	}

	return bs, nil
}
