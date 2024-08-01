package proxy

import "github.com/ethrousseau/weblens/api/types"

func (p *ProxyStore) CreateShare(share types.Share) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) UpdateShare(share types.Share) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) GetAllShares() ([]types.Share, error) {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) SetShareEnabledById(sId types.ShareId, enabled bool) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) AddUsersToShare(share types.Share, users []types.Username) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) GetSharedWithUser(username types.Username) ([]types.Share, error) {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStore) DeleteShare(shareId types.ShareId) error {
	// TODO implement me
	panic("implement me")
}
