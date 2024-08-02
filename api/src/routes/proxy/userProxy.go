package proxy

import (
	"github.com/ethrousseau/weblens/api/dataStore/user"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func (p *ProxyStore) GetAllUsers() ([]types.User, error) {
	ret, err := p.CallHome("GET", "/api/core/users", nil)
	if err != nil {
		return nil, err
	}

	us, err := ReadResponseBody[[]*user.User](ret)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[types.User](us), nil
}

func (p *ProxyStore) CreateUser(user types.User) error {
	return types.ErrNotImplemented("CreateUser proxy")
}

func (p *ProxyStore) UpdatePasswordByUsername(username types.Username, newPasswordHash string) error {
	return types.ErrNotImplemented("UpdatePasswordByUsername proxy")
}

func (p *ProxyStore) SetAdminByUsername(username types.Username, isAdmin bool) error {
	return types.ErrNotImplemented("SetAdminByUsername proxy")
}

func (p *ProxyStore) ActivateUser(username types.Username) error {
	return types.ErrNotImplemented("ActivateUser proxy")
}

func (p *ProxyStore) AddTokenToUser(username types.Username, token string) error {
	return types.ErrNotImplemented("AddTokenToUser proxy")
}

func (p *ProxyStore) SearchUsers(search string) ([]types.Username, error) {
	return nil, types.ErrNotImplemented("SearchUsers proxy")
}

func (p *ProxyStore) DeleteUserByUsername(username types.Username) error {
	return types.ErrNotImplemented("DeleteUserByUsername proxy")
}

func (p *ProxyStore) DeleteAllUsers() error {
	return types.ErrNotImplemented("DeleteAllUsers proxy")
}
