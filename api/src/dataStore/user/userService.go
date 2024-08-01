package user

import (
	"sync"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type userService struct {
	userMap    map[types.Username]*User
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
	users, err := us.db.GetAllUsers()
	if err != nil {
		return err
	}

	for _, u := range users {
		if u == nil {
			util.ShowErr(types.NewWeblensError("nil user in user service init"))
			continue
		}
		realU := u.(*User)
		realU.tokensLock = &sync.RWMutex{}
		us.userMap[u.GetUsername()] = realU
	}

	publicUser := &User{
		Username:     "PUBLIC",
		Activated:    true,
		isSystemUser: true,
		tokensLock:   &sync.RWMutex{},
	}

	us.publicUser = publicUser
	us.userMap["PUBLIC"] = publicUser
	us.userMap["WEBLENS"] = &User{
		tokensLock:   &sync.RWMutex{},
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
	err := types.SERV.StoreService.CreateUser(user)
	if err != nil {
		return err
	}

	us.userMap[user.GetUsername()] = user.(*User)
	return nil
}

func (us *userService) Del(un types.Username) error {
	return types.ErrNotImplemented("delete user")
}

func (us *userService) GetAll() ([]types.User, error) {
	users := util.MapToValues(us.userMap)
	return util.FilterMap(
		users, func(t *User) (types.User, bool) {
			return t, !t.isSystemUser
		},
	), nil
}

func (us *userService) Get(username types.Username) types.User {
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
