package proxy

import (
	"github.com/ethrousseau/weblens/api/dataStore/instance"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util/wlog"
)

type newServerBody struct {
	Id       types.InstanceId    `json:"serverId"`
	Role     types.ServerRole    `json:"role"`
	Name     string              `json:"name"`
	UsingKey types.WeblensApiKey `json:"usingKey"`
}

func (p *ProxyStore) AttachToCore(this types.Instance, core types.Instance) (types.Instance, error) {
	coreAddr, err := core.GetAddress()
	if err != nil {
		return nil, err
	}

	p.coreAddress = coreAddr
	p.apiKey = core.GetUsingKey()

	body := newServerBody{Id: this.ServerId(), Role: types.Backup, Name: this.GetName(), UsingKey: core.GetUsingKey()}
	resp, err := p.CallHome("POST", "/api/core/remote", body)

	if resp.StatusCode == 201 {
		newCore, err := ReadResponseBody[*instance.WeblensInstance](resp)
		if err != nil {
			return nil, err
		}

		newCore.SetUsingKey(core.GetUsingKey())
		err = newCore.SetAddress(p.coreAddress)
		if err != nil {
			return nil, err
		}

		return newCore, nil
	} else {
		return nil, types.WeblensErrorMsg("failed to attach to remote core")
	}
}

func (p *ProxyStore) GetAllServers() ([]types.Instance, error) {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) NewServer(instance types.Instance) error {
	return p.db.NewServer(instance)
}

func (p *ProxyStore) CreateApiKey(info types.ApiKeyInfo) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) GetApiKeys() ([]types.ApiKeyInfo, error) {
	wlog.Debug.Println("implement me")
	return []types.ApiKeyInfo{}, nil
}

func (p *ProxyStore) DeleteApiKey(key types.WeblensApiKey) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) SetRemoteUsing(key types.WeblensApiKey, remoteId types.InstanceId) error {
	return p.db.SetRemoteUsing(key, remoteId)
}

func (p *ProxyStore) DeleteServer(id types.InstanceId) error {
	return p.db.DeleteServer(id)
}
