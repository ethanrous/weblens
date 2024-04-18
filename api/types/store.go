package types

type Store interface {
	UserStore
}

type UserStore interface {
	LoadUsers() (err error)
	GetUsers() []User
}
