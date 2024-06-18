package types

import "fmt"

type BaseService[ID comparable, Obj any] interface {
	Init(DatabaseService) error
	Size() int

	Get(ID) Obj
	Add(Obj) error
	Del(ID) error
}

var ErrNotImplemented = func(note string) WeblensError { return NewWeblensError(fmt.Sprint("not implemented: ", note)) }
