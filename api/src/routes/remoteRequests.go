package routes

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/ethrousseau/weblens/api/dataStore/user"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type requester struct {
	ApiKey      types.WeblensApiKey
	CoreAddress string
}

func NewRequester() types.Requester {
	local := types.SERV.InstanceService.GetLocal()
	if local.ServerRole() == types.Initialization {
		return &requester{}
	}
	addr, _ := local.GetCoreAddress()

	return &requester{
		ApiKey:      local.GetUsingKey(),
		CoreAddress: addr,
	}
}

// Address extensions are of the form `BASE_CORE_ADDRESS` + "/api/core" + `addrExt`.
// So, to make a call to http://mycore.net/api/core/info, you would simply pass "/info" to `addrExt“
func (r *requester) coreRequest(method string, addrExt string, body any, baseOverride ...string) (
	*http.Response, error,
) {
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

// func (r *requester) AttachToCore(i types.Instance) (types.Instance, error) {
// 	coreAddr, err := i.GetCoreAddress()
// 	if err != nil {
// 		return nil, err
// 	}
// 	r.CoreAddress = coreAddr
// 	r.ApiKey = i.GetUsingKey()
//
// 	body := newServerBody{Id: i.ServerId(), Role: types.Backup, Name: i.GetName(), UsingKey: i.GetUsingKey()}
// 	resp, err := r.coreRequest("POST", "/remote", body)
// 	if err != nil {
// 		return nil, types.WeblensErrorFromError(err)
// 	}
//
// 	// if resp.StatusCode == 201 {
// 	// 	return readRespBody[](resp)
// 	// } else {
// 	// 	return nil, types.WeblensErrorMsg("failed to attach to remote core")
// 	// }
// }

// func (r *requester) RequestCoreSnapshot() ([]types.FileJournalEntry, error) {
// 	latest, err := dataStore.GetLatestBackup()
// 	if err != nil {
// 		return nil, err
// 	}
// 	resp, err := r.coreRequest("GET", "/snapshot?since="+strconv.FormatInt(latest.UnixMilli(), 10), nil)
// 	if err != nil {
// 		return nil, err
// 	} else if resp.StatusCode != 200 {
// 		err = errors.New("bad status: " + resp.Status)
// 		return nil, err
// 	}
//
// 	j, err := readRespBody[dataStore.JournalResp](resp)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return util.SliceConvert[types.FileJournalEntry](j.Journal), nil
// }

func (r *requester) GetCoreUsers() (us []types.User, err error) {
	resp, err := r.coreRequest("GET", "/users", nil)
	if err != nil {
		return
	}

	body, err := readRespBody[[]*user.User](resp)
	if err != nil {
		return
	}

	return util.SliceConvert[types.User](body), nil

}

func (r *requester) GetCoreFileBin(f types.WeblensFile) ([][]byte, error) {
	resp, err := r.coreRequest("GET", "/file/"+string(f.ID())+"/content", nil)
	if err != nil {
		return nil, err
	}
	var bs [][]byte
	origFileBs, err := readRespBodyRaw(resp)
	if err != nil {
		return nil, err
	}
	bs = append(bs, origFileBs)

	if f.IsDisplayable() {
		m := types.SERV.MediaRepo.Get(f.GetContentId())

		if m == nil {
			return nil, types.ErrNoMedia
		}

		resp, err := r.coreRequest("GET", "/media/"+string(m.ID())+"/thumbnail", nil, "/api")
		if err != nil {
			return nil, err
		}
		thumbBs, err := readRespBodyRaw(resp)
		if err != nil {
			return nil, err
		}
		bs = append(bs, thumbBs)

		resp, err = r.coreRequest("GET", "/media/"+string(m.ID())+"/fullres", nil, "/api")
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
