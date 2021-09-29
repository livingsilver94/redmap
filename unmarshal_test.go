package redmap_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/livingsilver94/redmap"
)

type StubTextUnmarshaler struct{ S string }

func (s *StubTextUnmarshaler) UnmarshalText(text []byte) error {
	s.S = string(text)
	return nil
}

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

func TestUnmarshalScalars(t *testing.T) {
	tests := []struct {
		In  map[string]string
		Out interface{}
	}{
		{In: map[string]string{"V": "true"}, Out: struct{ V bool }{true}},
		{In: map[string]string{"V": "100"}, Out: struct{ V int }{100}},
		{In: map[string]string{"V": "100"}, Out: struct{ V int8 }{100}},
		{In: map[string]string{"V": "100"}, Out: struct{ V int16 }{100}},
		{In: map[string]string{"V": "100"}, Out: struct{ V int32 }{100}},
		{In: map[string]string{"V": "100"}, Out: struct{ V int64 }{100}},
		{In: map[string]string{"V": "100"}, Out: struct{ V uint }{100}},
		{In: map[string]string{"V": "100"}, Out: struct{ V uint8 }{100}},
		{In: map[string]string{"V": "100"}, Out: struct{ V uint16 }{100}},
		{In: map[string]string{"V": "100"}, Out: struct{ V uint32 }{100}},
		{In: map[string]string{"V": "100"}, Out: struct{ V uint64 }{100}},
		{In: map[string]string{"V": "100.1"}, Out: struct{ V float32 }{100.1}},
		{In: map[string]string{"V": "100.1"}, Out: struct{ V float64 }{100.1}},
		{In: map[string]string{"V": "(100.1+80.1i)"}, Out: struct{ V complex64 }{100.1 + 80.1i}},
		{In: map[string]string{"V": "(100.1+80.1i)"}, Out: struct{ V complex128 }{100.1 + 80.1i}},
		{In: map[string]string{"V": "str"}, Out: struct{ V string }{"str"}},
		{In: map[string]string{"V": "a test"}, Out: struct{ V StubTextUnmarshaler }{StubTextUnmarshaler{S: "a test"}}},
	}
	for _, test := range tests {
		zero := reflect.New(reflect.TypeOf(test.Out))
		err := redmap.Unmarshal(test.In, zero.Interface())
		if err != nil {
			t.Fatalf("Unmarshal returned unexpected error %q", err)
		}
		if !reflect.DeepEqual(zero.Elem().Interface(), test.Out) {
			t.Fatalf("Unmarshal's output doesn't match the expected value\n\tIn: %v\n\tExpected: %v\n\tOut: %v", test.In, test.Out, zero)
		}
	}
}
