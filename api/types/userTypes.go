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
	GetToken() string
	GetHomeFolder() WeblensFile
	SetHomeFolder(WeblensFile) error
	CreateHomeFolder() (WeblensFile, error)
	GetTrashFolder() WeblensFile
	SetTrashFolder(WeblensFile) error

	CheckLogin(password string) bool
	UpdatePassword(oldPass, newPass string) error
	SetAdmin(isAdmin bool) error
	Activate() error

	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

type UserService interface {
	BaseService[Username, User]
	GetAll() ([]User, error)
}

var ErrUserNotAuthenticated = NewWeblensError("user credentials are invalid")
var ErrBadPassword = NewWeblensError("password provided does not authenticate user")
