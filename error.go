package redmap

import (
	"errors"
	"fmt"
)

var (
	ErrNilValue   = errors.New("nil")
	ErrNotPointer = errors.New("not a pointer")
	ErrNotStruct  = errors.New("not a struct type")
)

func errIs(something interface{}, err error) error {
	return fmt.Errorf("%s is %w", something, err)
}
