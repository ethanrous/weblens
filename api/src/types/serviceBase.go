package types

import "fmt"

type WeblensService[ID comparable, Obj any, DB any] interface {
	Init(DB) error
	Size() int

	Get(ID) Obj
	Add(Obj) error
	Del(ID) error
}

var ErrNotImplemented = func(note string) WeblensError { return NewWeblensError(fmt.Sprint("not implemented: ", note)) }
