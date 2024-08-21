package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	weblens "github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/fileTree"
	error2 "github.com/ethrousseau/weblens/api/internal/werror"
)

type ProxyStoreImpl struct {
	coreAddress string
	apiKey weblens.WeblensApiKey
}

func NewProxyStore(coreAddress string, apiKey weblens.WeblensApiKey) *ProxyStoreImpl {
	return &ProxyStoreImpl{
		coreAddress: coreAddress,
		apiKey:      apiKey,
	}
}

func ReadResponseBody[T any](resp *http.Response) (T, error) {
	var target T

	if resp.StatusCode > 299 {
		return target, error2.WErrMsg(
			fmt.Sprintf(
				"Trying to read response body of call to %s with bad status code: %d",
				resp.Request.URL.String(), resp.StatusCode,
			),
		)
	}

	bs, err := io.ReadAll(resp.Body)

	if err != nil {
		return target, error2.Wrap(err)
	}

	err = resp.Body.Close()
	if err != nil {
		return target, error2.Wrap(err)
	}

	// If the requester just wants the bytes, skip unmarshaling
	switch any(target).(type) {
	case []byte:
		return any(bs).(T), nil
	}

	err = json.Unmarshal(bs, &target)
	if err != nil {
		return target, error2.Wrap(err)
	}

	return target, nil
}

func (p *ProxyStoreImpl) CallHome(method, endpoint string, body any) (*http.Response, error) {
	if p.coreAddress == "" {
		return nil, error2.WErrMsg("Trying to dial core with no address")
	}
	if p.apiKey == "" {
		return nil, error2.WErrMsg("Trying to dial core without api key")
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

type ProxyStore interface {
	WriteFileEvent(event fileTree.FileEvent) error
	GetAllLifetimes() ([]fileTree.Lifetime, error)
	GetLifetimesSince(time.Time) ([]fileTree.Lifetime, error)
	UpsertLifetime(l fileTree.Lifetime) error
	InsertManyLifetimes([]fileTree.Lifetime) error
	GetActionsByPath(fileTree.WeblensFilepath) ([]fileTree.FileAction, error)
	DeleteAllFileHistory() error
	GetLatestAction() (fileTree.FileAction, error)

	NewTrashEntry(te TrashEntry) error
	GetTrashEntry(fileId fileTree.FileId) (TrashEntry, error)
	DeleteTrashEntry(fileId fileTree.FileId) error
	GetAllFiles() ([]fileTree.WeblensFile, error)
	StatFile(fileTree.WeblensFile) (FileStat, error)
	ReadFile(fileTree.WeblensFile) ([]byte, error)
	ReadDir(fileTree.WeblensFile) ([]FileStat, error)
	TouchFile(fileTree.WeblensFile) error
	GetFile(fileTree.FileId) (fileTree.WeblensFile, error)
	StreamFile(fileTree.WeblensFile) (io.ReadCloser, error)

	GetAllUsers() ([]*weblens.User, error)
	UpdatePasswordByUsername(username weblens.Username, newPasswordHash string) error
	SetAdminByUsername(weblens.Username, bool) error
	CreateUser(*weblens.User) error
	ActivateUser(weblens.Username) error
	AddTokenToUser(username weblens.Username, token string) error
	SearchUsers(search string) ([]weblens.Username, error)

	DeleteUserByUsername(weblens.Username) error
	DeleteAllUsers() error

	GetAllServers() ([]*weblens.WeblensInstance, error)
	NewServer(*weblens.WeblensInstance) error
	DeleteServer(id weblens.InstanceId) error
	AttachToCore(this *weblens.WeblensInstance, core *weblens.WeblensInstance) (*weblens.WeblensInstance, error)
}
