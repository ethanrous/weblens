package user

import (
	"sync"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type userService struct {
	userMap    map[types.Username]*User
	userLock sync.RWMutex
	publicUser types.User
	db         types.UserStore
}

func NewService() types.UserService {
	return &userService{
		userMap: make(map[types.Username]*User),
	}
}

func (us *userService) Init(db types.UserStore) error {
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
	us.userMap["WEBLENS"] = &User{

		Username:     "WEBLENS",
		isSystemUser: true,
	}

	return nil
}

func (us *userService) Size() int {
	return len(us.userMap) - 2
}

func (us *userService) GetPublicUser() types.User {
	return us.publicUser
}

func (us *userService) Add(user types.User) error {
	if _, ok := us.userMap[user.GetUsername()]; ok {
		return nil
		// return types.NewWeblensError("user already exists")
	}

	if user.(*User).Id == [12]uint8{0} {
		user.(*User).Id = primitive.NewObjectID()
	}
	err := types.SERV.StoreService.CreateUser(user)
	if err != nil {
		return err
	}

	us.userLock.Lock()
	defer us.userLock.Unlock()
	us.userMap[user.GetUsername()] = user.(*User)

	return nil
}

func (us *userService) Del(un types.Username) error {
	err := us.db.DeleteUserByUsername(un)
	if err != nil {
		return err
	}

	us.userLock.Lock()
	defer us.userLock.Unlock()
	delete(us.userMap, un)

	return nil
}

func (us *userService) GetAll() ([]types.User, error) {
	us.userLock.RLock()
	defer us.userLock.RUnlock()
	users := util.MapToValues(us.userMap)
	return util.FilterMap(
		users, func(t *User) (types.User, bool) {
			return t, !t.isSystemUser
		},
	), nil
}

func (us *userService) Get(username types.Username) types.User {
	us.userLock.RLock()
	defer us.userLock.RUnlock()
	u, ok := us.userMap[username]
	if !ok {
		return nil
	}
	return u
}

func (us *userService) SearchByUsername(searchString string) ([]types.User, error) {
	usernames, err := types.SERV.StoreService.SearchUsers(searchString)
	if err != nil {
		return nil, err
	}

	return util.Map(
		usernames, func(un types.Username) types.User {
			return us.Get(un)
		},
	), nil
}

func (us *userService) SetUserAdmin(u types.User, admin bool) error {
	err := us.db.SetAdminByUsername(u.GetUsername(), admin)
	if err != nil {
		return err
	}

	u.(*User).Admin = admin

	return nil
}
