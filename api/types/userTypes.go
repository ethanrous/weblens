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
	GetTrashFolder() WeblensFile
}
