package types

type Store interface {
	UserStore
}

type UserStore interface {
	LoadUsers(FileTree) (err error)
	GetUsers() []User
}
