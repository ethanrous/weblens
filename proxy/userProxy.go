package proxy

import (
	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/types"
)

func (p *ProxyStoreImpl) GetAllUsers() ([]types.User, error) {
	ret, err := p.CallHome("GET", "/api/core/users", nil)
	if err != nil {
		return nil, err
	}

	us, err := ReadResponseBody[[]*weblens.User](ret)
	if err != nil {
		return nil, err
	}

	return internal.SliceConvert[types.User](us), nil
}

func (p *ProxyStoreImpl) CreateUser(user types.User) error {
	return p.db.CreateUser(user)
}

func (p *ProxyStoreImpl) UpdatePasswordByUsername(username types.Username, newPasswordHash string) error {
	return werror.NotImplemented("UpdatePasswordByUsername proxy")
}

func (p *ProxyStoreImpl) SetAdminByUsername(username types.Username, isAdmin bool) error {
	return werror.NotImplemented("SetAdminByUsername proxy")
}

func (p *ProxyStoreImpl) ActivateUser(username types.Username) error {
	return werror.NotImplemented("ActivateUser proxy")
}

func (p *ProxyStoreImpl) AddTokenToUser(username types.Username, token string) error {
	return werror.NotImplemented("AddTokenToUser proxy")
}

func (p *ProxyStoreImpl) SearchUsers(search string) ([]types.Username, error) {
	return nil, werror.NotImplemented("SearchUsers proxy")
}

func (p *ProxyStoreImpl) DeleteUserByUsername(username types.Username) error {
	return werror.NotImplemented("DeleteUserByUsername proxy")
}

func (p *ProxyStoreImpl) DeleteAllUsers() error {
	return werror.NotImplemented("DeleteAllUsers proxy")
}
