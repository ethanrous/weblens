package proxy

import (
	"iter"
	"net/http"
	"slices"
	"sync"

	"github.com/ethanrous/weblens/models"
)

var _ models.UserService = (*ProxyUserService)(nil)

type ProxyUserService struct {
	Core *models.Instance

	publicUser  *models.User
	rootUser    *models.User
	userMap     map[string]*models.User
	userMapLock sync.RWMutex
}

func NewProxyUserService(core *models.Instance) *ProxyUserService {
	return &ProxyUserService{
		Core: core,
		publicUser: &models.User{
			Username:   "PUBLIC",
			Activated:  true,
			SystemUser: true,
		},
		rootUser: &models.User{
			Username:   "WEBLENS",
			SystemUser: true,
		},
	}
}

func (pus *ProxyUserService) Size() int {
	if pus.userMap == nil {
		return 0
	}

	return len(pus.userMap)
}

func (pus *ProxyUserService) Get(id models.Username) *models.User {
	pus.userMapLock.RLock()
	defer pus.userMapLock.RUnlock()
	if pus.userMap == nil {
		return nil
	}

	return pus.userMap[id]
}

func (pus *ProxyUserService) Add(user *models.User) error {
	pus.userMapLock.Lock()
	defer pus.userMapLock.Unlock()

	if pus.userMap == nil {
		pus.userMap = make(map[string]*models.User)
	}

	if _, ok := pus.userMap[user.Username]; !ok {
		pus.userMap[user.Username] = user
	}

	return nil
}

func (pus *ProxyUserService) CreateOwner(username, password, fullName string) (*models.User, error) {
	panic("implement me")
}

func (pus *ProxyUserService) UpdateFullName(u *models.User, newFullName string) error {
	panic("implement me")
}

func (pus *ProxyUserService) Del(id models.Username) error {
	panic("implement me")
}

func (pus *ProxyUserService) GetAll() (iter.Seq[*models.User], error) {
	r := NewCoreRequest(pus.Core, http.MethodGet, "/backup/users")
	users, err := CallHomeStruct[[]*models.User](r)

	if err != nil {
		return nil, err
	}

	return slices.Values(users), nil
}

func (pus *ProxyUserService) GetPublicUser() *models.User {
	return pus.publicUser
}

func (pus *ProxyUserService) SearchByUsername(searchString string) (iter.Seq[*models.User], error) {

	panic("implement me")
}

func (pus *ProxyUserService) SetUserAdmin(user *models.User, b bool) error {

	panic("implement me")
}

func (pus *ProxyUserService) ActivateUser(user *models.User, active bool) error {

	panic("implement me")
}

func (pus *ProxyUserService) GetRootUser() *models.User {
	return pus.rootUser
}

func (pus *ProxyUserService) GenerateToken(user *models.User) (string, error) {

	panic("implement me")
}

func (pus *ProxyUserService) UpdateUserPassword(
	username models.Username, oldPassword, newPassword string, allowEmptyOld bool,
) error {

	panic("implement me")
}

func (pus *ProxyUserService) UpdateUserHome(u *models.User) error {
	// TODO implement me
	panic("implement me")
}
