package redmap_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/livingsilver94/redmap"
)

var emptyMap = make(map[string]string)

func TestUnmarshalValid(t *testing.T) {
	tests := []interface{}{
		StubStringer{},
		&StubStringer{},
	}
	for _, test := range tests {
		val := reflect.New(reflect.TypeOf(test))
		val.Elem().Set(reflect.ValueOf(test))
		if err := redmap.Unmarshal(emptyMap, val.Interface()); err != nil {
			t.Fatalf("Unmarshal() of valid type must not return error. Returned %q", err)
		}
	}
}

func TestUnmarshalNil(t *testing.T) {
	err := redmap.Unmarshal(nil, &StubStringer{})
	if !errors.Is(err, redmap.ErrNilValue) {
		t.Fatal("Unmarshal() with a nil map did not return the specific error")
	}
	err = redmap.Unmarshal(emptyMap, nil)
	if !errors.Is(err, redmap.ErrNilValue) {
		t.Fatal("Unmarshal() with invalid target did not return the specific error")
	}
	var nilPtr *StubStringer
	err = redmap.Unmarshal(emptyMap, nilPtr)
	if !errors.Is(err, redmap.ErrNilValue) {
		t.Fatal("Unmarshal() with nil pointer did not return the specific error")
	}
}

func TestUnmarshalInvalid(t *testing.T) {
	tests := []struct {
		val    func() reflect.Value
		expErr error
	}{
		{
			// No pointer passed.
			val:    func() reflect.Value { return reflect.ValueOf(100) },
			expErr: redmap.ErrNotPointer},
		{
			// Int is not a struct.
			val: func() reflect.Value {
				v := 100
				return reflect.ValueOf(&v)
			},
			expErr: redmap.ErrNotStruct},
		{
			// Interface is not a struct.
			val: func() reflect.Value {
				v := fmt.Stringer(StubStringer{})
				return reflect.ValueOf(&v)
			},
			expErr: redmap.ErrNotStruct},
	}
	for _, test := range tests {
		err := redmap.Unmarshal(emptyMap, test.val().Interface())
		if !errors.Is(err, test.expErr) {
			t.Fatalf("Unmarshal returned %q but should have returned %q", err, test.expErr)
		}
	}
}
