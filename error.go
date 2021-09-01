package redmap

import "errors"

var (
	ErrNilValue  = errors.New("provided a nil value")
	ErrNonStruct = errors.New("not a struct type")
)
