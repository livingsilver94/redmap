package redmap

import (
	"errors"
	"fmt"
)

var (
	// ErrNilValue is returned when an argument passed is nil but it should not be.
	ErrNilValue = errors.New("nil")
	// ErrNotPointer is retuned when an argument passed is not a pointer but it should be.
	ErrNotPointer = errors.New("not a pointer")
	// ErrNotStruct is returned when an argument passed is not a struct but it should be.
	ErrNotStruct = errors.New("not a struct type")
)

func errIs(something interface{}, err error) error {
	return fmt.Errorf("%s is %w", something, err)
}
