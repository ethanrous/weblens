package proxy

import (
	"iter"
	"net/http"
	"slices"

	"github.com/ethrousseau/weblens/models"
)

var _ models.UserService = (*ProxyUserService)(nil)

type ProxyUserService struct {
	Core *models.Instance
}

func (pus *ProxyUserService) Init() error {
	// TODO implement me
	panic("implement me")
}

func (pus *ProxyUserService) Size() int {
	// TODO implement me
	panic("implement me")
}

func (pus *ProxyUserService) Get(id models.Username) *models.User {
	// TODO implement me
	panic("implement me")
}

func (pus *ProxyUserService) Add(user *models.User) error {
	// TODO implement me
	panic("implement me")
}

func (pus *ProxyUserService) Del(id models.Username) error {
	// TODO implement me
	panic("implement me")
}

func (pus *ProxyUserService) GetAll() (iter.Seq[*models.User], error) {
	users, err := CallHomeStruct[[]*models.User](pus.Core, http.MethodGet, "/users", nil)

	if err != nil {
		return nil, err
	}

	return slices.Values(users), nil
}

func (pus *ProxyUserService) GetPublicUser() *models.User {
	// TODO implement me
	panic("implement me")
}

func (pus *ProxyUserService) SearchByUsername(searchString string) (iter.Seq[*models.User], error) {
	// TODO implement me
	panic("implement me")
}

func (pus *ProxyUserService) SetUserAdmin(user *models.User, b bool) error {
	// TODO implement me
	panic("implement me")
}

func (pus *ProxyUserService) ActivateUser(user *models.User) error {
	// TODO implement me
	panic("implement me")
}

func (pus *ProxyUserService) GetRootUser() *models.User {
	// TODO implement me
	panic("implement me")
}

func (pus *ProxyUserService) GenerateToken(user *models.User) (string, error) {
	// TODO implement me
	panic("implement me")
}

func (pus *ProxyUserService) UpdateUserPassword(
	username models.Username, oldPassword, newPassword string, allowEmptyOld bool,
) error {
	// TODO implement me
	panic("implement me")
}
