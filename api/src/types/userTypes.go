package types

type Username string

func (u Username) String() string {
	return string(u)
}

type User interface {
	GetUsername() Username
	IsAdmin() bool
	IsActive() bool
	IsOwner() bool
	IsSystemUser() bool
	GetToken() string
	GetHomeFolder() WeblensFile
	SetHomeFolder(WeblensFile) error
	GetTrashFolder() WeblensFile
	SetTrashFolder(WeblensFile) error

	CheckLogin(password string) bool

	FormatArchive() (map[string]any, error)
	UnmarshalJSON(data []byte) error
}

type UserService interface {
	WeblensService[Username, User, UserStore]
	GetAll() ([]User, error)
	GetPublicUser() User
	SearchByUsername(searchString string) ([]User, error)
	SetUserAdmin(User, bool) error
	ActivateUser(User) error
	UpdateUserPassword(username Username, oldPassword, newPassword string, allowEmptyOld bool) error
}

var ErrBadPassword = NewWeblensError("password provided does not authenticate user")
var ErrUserAlreadyExists = NewWeblensError("cannot create two users with the same username")
