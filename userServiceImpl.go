package weblens

import (
	"sync"

	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceImpl struct {
	userMap    map[types.Username]*User
	userLock   sync.RWMutex
	publicUser *User
	rootUser   *User
	db         types.UserStore
}

func NewUserService() *UserServiceImpl {
	return &UserServiceImpl{
		userMap: make(map[types.Username]*User),
	}
}

func (us *UserServiceImpl) Init(db types.UserStore) error {
	us.db = db

	proxyStore, ok := us.db.(types.ProxyStore)
	if ok {
		users, err := proxyStore.GetLocalStore().GetAllUsers()
		if err != nil {
			return err
		}
		for _, user := range users {
			us.userMap[user.GetUsername()] = user.(*User)
		}

		users, err = proxyStore.GetAllUsers()
		if err != nil {
			return err
		}
		for _, user := range users {
			err = us.Add(user)
			if err != nil {
				return err
			}
		}
	} else {
		users, err := us.db.GetAllUsers()
		if err != nil {
			return err
		}
		for _, user := range users {
			us.userMap[user.GetUsername()] = user.(*User)
		}
	}

	publicUser := &User{
		Username:     "PUBLIC",
		Activated:    true,
		isSystemUser: true,
	}

	us.publicUser = publicUser
	us.userMap["PUBLIC"] = publicUser
	us.rootUser = &User{

		Username:     "WEBLENS",
		isSystemUser: true,
	}
	us.userMap["WEBLENS"] = us.rootUser

	return nil
}

func (us *UserServiceImpl) Size() int {
	return len(us.userMap) - 2
}

func (us *UserServiceImpl) GetPublicUser() *User {
	return us.publicUser
}

func (us *UserServiceImpl) GetRootUser() *User {
	return us.rootUser
}

func (us *UserServiceImpl) Add(user types.User) error {
	if _, ok := us.userMap[user.GetUsername()]; ok {
		return nil
		// return types.NewWeblensError("user already exists")
	}

	if user.(*User).Id == [12]uint8{0} {
		user.(*User).Id = primitive.NewObjectID()
	}
	err := us.db.CreateUser(user)
	if err != nil {
		return err
	}

	us.userLock.Lock()
	defer us.userLock.Unlock()
	us.userMap[user.GetUsername()] = user.(*User)

	return nil
}

func (us *UserServiceImpl) Del(un types.Username) error {
	err := us.db.DeleteUserByUsername(un)
	if err != nil {
		return err
	}

	us.userLock.Lock()
	defer us.userLock.Unlock()
	delete(us.userMap, un)

	return nil
}

func (us *UserServiceImpl) ActivateUser(u types.User) (err error) {
	realU := u.(*User)

	// _, err = u.CreateHomeFolder()
	// if err != nil {
	// 	return err
	// }

	err = us.db.ActivateUser(u.GetUsername())
	if err != nil {
		return err
	}

	realU.Activated = true

	return
}

func (us *UserServiceImpl) GetAll() ([]types.User, error) {
	us.userLock.RLock()
	defer us.userLock.RUnlock()
	users := internal.MapToValues(us.userMap)
	return internal.FilterMap(
		users, func(t *User) (types.User, bool) {
			return t, !t.isSystemUser
		},
	), nil
}

func (us *UserServiceImpl) Get(username types.Username) types.User {
	us.userLock.RLock()
	defer us.userLock.RUnlock()
	u, ok := us.userMap[username]
	if !ok {
		return nil
	}
	return u
}

func (us *UserServiceImpl) SearchByUsername(searchString string) ([]types.User, error) {
	usernames, err := us.db.SearchUsers(searchString)
	if err != nil {
		return nil, err
	}

	return internal.Map(
		usernames, func(un types.Username) types.User {
			return us.Get(un)
		},
	), nil
}

func (us *UserServiceImpl) SetUserAdmin(u types.User, admin bool) error {
	err := us.db.SetAdminByUsername(u.GetUsername(), admin)
	if err != nil {
		return err
	}

	u.(*User).Admin = admin

	return nil
}

func (us *UserServiceImpl) UpdateUserPassword(
	username types.Username, oldPassword, newPassword string,
	allowEmptyOld bool,
) error {
	usr := us.userMap[username]

	if !allowEmptyOld || oldPassword != "" {
		if auth := usr.CheckLogin(oldPassword); !auth {
			return types.ErrBadPassword
		}
	}

	passHashBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), 11)
	if err != nil {
		return err
	}

	passHashStr := string(passHashBytes)

	err = us.db.UpdatePasswordByUsername(username, passHashStr)
	if err != nil {
		return err
	}
	usr.Password = passHashStr

	return nil
}
