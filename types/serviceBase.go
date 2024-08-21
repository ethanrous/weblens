package types

type WeblensService[ID comparable, Obj any, DB any] interface {
	Init(DB) error
	Size() int

	Get(ID) Obj
	Add(Obj) error
	Del(ID) error
}

