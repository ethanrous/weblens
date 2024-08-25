package proxy

//
// import (

// 	"github.com/ethrousseau/weblens/api/internal"
// 	"github.com/ethrousseau/weblens/api/internal/werror"
// 	"github.com/ethrousseau/weblens/api/types"
// )
//
// func (p *ProxyStoreImpl) GetAllUsers() ([]*weblens.User, error) {
// 	ret, err := p.CallHome("GET", "/api/core/users", nil)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	us, err := ReadResponseBody[[]*weblens.User](ret)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return internal.SliceConvert[*weblens.User](us), nil
// }
//
// func (p *ProxyStoreImpl) CreateUser(user *weblens.User) error {
// 	return p.db.CreateUser(user)
// }
//
// func (p *ProxyStoreImpl) UpdatePasswordByUsername(username weblens.Username, newPasswordHash string) error {
// 	return werror.NotImplemented("UpdatePasswordByUsername proxy")
// }
//
// func (p *ProxyStoreImpl) SetAdminByUsername(username weblens.Username, isAdmin bool) error {
// 	return werror.NotImplemented("SetAdminByUsername proxy")
// }
//
// func (p *ProxyStoreImpl) ActivateUser(username weblens.Username) error {
// 	return werror.NotImplemented("ActivateUser proxy")
// }
//
// func (p *ProxyStoreImpl) AddTokenToUser(username weblens.Username, token string) error {
// 	return werror.NotImplemented("AddTokenToUser proxy")
// }
//
// func (p *ProxyStoreImpl) SearchUsers(search string) ([]weblens.Username, error) {
// 	return nil, werror.NotImplemented("SearchUsers proxy")
// }
//
// func (p *ProxyStoreImpl) DeleteUserByUsername(username weblens.Username) error {
// 	return werror.NotImplemented("DeleteUserByUsername proxy")
// }
//
// func (p *ProxyStoreImpl) DeleteAllUsers() error {
// 	return werror.NotImplemented("DeleteAllUsers proxy")
// }
