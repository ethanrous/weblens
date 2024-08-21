package proxy

import (
	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
)

func (p *ProxyStoreImpl) CreateShare(share weblens.Share) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStoreImpl) UpdateShare(share weblens.Share) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStoreImpl) GetAllShares() ([]weblens.Share, error) {
	wlog.Debug.Println("implement me")
	return []weblens.Share{}, nil
}

func (p *ProxyStoreImpl) SetShareEnabledById(sId types.ShareId, enabled bool) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStoreImpl) AddUsersToShare(share weblens.Share, users []types.Username) error {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStoreImpl) GetSharedWithUser(username types.Username) ([]weblens.Share, error) {
	// TODO implement me
	panic("implement me")
}

func (p *ProxyStoreImpl) DeleteShare(shareId types.ShareId) error {
	// TODO implement me
	panic("implement me")
}
