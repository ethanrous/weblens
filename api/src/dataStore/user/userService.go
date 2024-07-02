package user

import (
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

type userService struct {
	repo       map[types.Username]types.User
	publicUser types.User
	db         types.UserDB
}

func NewService() types.UserService {
	return &userService{
		repo: make(map[types.Username]types.User),
	}
}

func (us *userService) Init(db types.DatabaseService) error {
	us.db = db
	users, err := db.GetAllUsers()
	if err != nil {
		return err
	}

	for _, u := range users {
		if u == nil {
			util.ShowErr(types.NewWeblensError("nil user in user service init"))
			continue
		}
		us.repo[u.GetUsername()] = u
	}

	us.publicUser = &User{
		Username:  "PUBLIC",
		Activated: true,
	}

	return nil
}

func (us *userService) Size() int {
	return len(us.repo)
}

func (us *userService) GetPublicUser() types.User {
	return us.publicUser
}

func (us *userService) Add(user types.User) error {
	err := types.SERV.Database.CreateUser(user)
	if err != nil {
		return err
	}

	us.repo[user.GetUsername()] = user
	return nil
}

func (us *userService) Del(un types.Username) error {
	return types.ErrNotImplemented("delete user")
}

func (us *userService) GetAll() ([]types.User, error) {
	return util.MapToSlicePure(us.repo), nil
}

func (us *userService) Get(username types.Username) types.User {
	return us.repo[username]
}

func (us *userService) SearchByUsername(searchString string) ([]types.User, error) {
	usernames, err := types.SERV.Database.SearchUsers(searchString)
	if err != nil {
		return nil, err
	}

	return util.Map(
		usernames, func(un types.Username) types.User {
			return us.Get(un)
		},
	), nil
}