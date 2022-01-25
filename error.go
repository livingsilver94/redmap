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
	// ErrNoCodec is returned when a type cannot be marshaled or unmarshaled,
	// e.g. it is neither a struct nor implements StringMap(Un)marshaler.
	ErrNoCodec = errors.New("not an encodable or decodable type")
)

func errIs(something interface{}, err error) error {
	return fmt.Errorf("%s is %w", something, err)
}
