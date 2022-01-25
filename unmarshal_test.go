package redmap_test

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/livingsilver94/redmap"
)

// stubTextUnmarshaler implements the encoding.TextUnmarshaler interface.
type stubTextUnmarshaler struct{ S string }

func (s *stubTextUnmarshaler) UnmarshalText(text []byte) error {
	s.S = string(text)
	return nil
}

// stubIntTextUnmarshaler is an int that implements encoding.TextUnmarshaler,
// so that we can test if a non-struct type is correctly handled as an interface.
type stubIntTextUnmarshaler int

func (s *stubIntTextUnmarshaler) UnmarshalText(text []byte) error {
	v, err := strconv.Atoi(string(text))
	if err != nil {
		return err
	}
	*s = stubIntTextUnmarshaler(v)
	return nil
}

// stubMapMarshaler implements the redmap.StringMapUnmarshaler interface.
type stubMapUnmarshaler struct {
	Field1 string
	Field2 string
}

func (s *stubMapUnmarshaler) UnmarshalStringMap(mp map[string]string) error {
	s.Field1 = mp["Field1"]
	s.Field2 = mp["Field2"]
	return nil
}

// StubIntUnmarshaler is an int that implements redmap.StringMapUnmarshaler,
// so that we can test if a non-struct type is correctly handled as an interface.
type stubIntMapUnmarshaler int

func (s *stubIntMapUnmarshaler) UnmarshalStringMap(mp map[string]string) error {
	v, err := strconv.Atoi(mp["Field1"])
	if err != nil {
		return err
	}
	*s = stubIntMapUnmarshaler(v)
	return nil
}

// emptyMap is a non-nil, zero-length map.
var emptyMap = make(map[string]string)

func TestUnmarshalValidType(t *testing.T) {
	tests := []interface{}{
		stubStringer{},
		&stubStringer{},
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
	err := redmap.Unmarshal(nil, &stubStringer{})
	if !errors.Is(err, redmap.ErrNilValue) {
		t.Fatal("Unmarshal() with a nil map did not return the specific error")
	}
	err = redmap.Unmarshal(emptyMap, nil)
	if !errors.Is(err, redmap.ErrNilValue) {
		t.Fatal("Unmarshal() with invalid target did not return the specific error")
	}
	var nilPtr *stubStringer
	err = redmap.Unmarshal(emptyMap, nilPtr)
	if !errors.Is(err, redmap.ErrNilValue) {
		t.Fatal("Unmarshal() with nil pointer did not return the specific error")
	}
}

func TestUnmarshalInvalidType(t *testing.T) {
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
				v := fmt.Stringer(stubStringer{})
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
		{In: map[string]string{"V": "a test"}, Out: struct{ V stubTextUnmarshaler }{stubTextUnmarshaler{S: "a test"}}},
		{In: map[string]string{"V": "100"}, Out: struct{ V stubIntTextUnmarshaler }{stubIntTextUnmarshaler(100)}},
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

func TestUnmarshalInnerStructs(t *testing.T) {
	type (
		Inner1Level = struct {
			String string
		}
		Inner2Level = struct {
			Inner Inner1Level `redmap:",inline"`
		}

		Root1Level = struct {
			Inner Inner1Level `redmap:",inline"`
		}
		Root2Level = struct {
			Inner Inner2Level `redmap:",inline"`
		}
		RootWithPointer = struct {
			Inner *Inner1Level `redmap:",inline"`
		}
	)
	tests := []struct {
		In  map[string]string
		Out interface{}
	}{
		{In: map[string]string{"Inner.String": "oneLevel"}, Out: Root1Level{Inner: Inner1Level{String: "oneLevel"}}},
		{In: map[string]string{"Inner.Inner.String": "twoLevel"}, Out: Root2Level{Inner: Inner2Level{Inner: Inner1Level{String: "twoLevel"}}}},
		{In: map[string]string{"Inner.String": "oneLevel"}, Out: RootWithPointer{Inner: &Inner1Level{String: "oneLevel"}}},
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

func TestUnarshalUnexported(t *testing.T) {
	mp := map[string]string{"Exp": "atest"}
	tests := []struct {
		Out interface{}
	}{
		{Out: struct {
			Exp   string
			unexp string
		}{Exp: "atest"}},
		{Out: struct {
			Exp   string
			unexp *string
		}{Exp: "atest"}},
		{Out: struct {
			Exp   string
			unexp stubTextUnmarshaler
		}{Exp: "atest"}},
	}
	for _, test := range tests {
		zero := reflect.New(reflect.TypeOf(test.Out))
		err := redmap.Unmarshal(mp, zero.Interface())
		if err != nil {
			t.Fatalf("Unmarshal returned unexpected error %q", err)
		}
		if !reflect.DeepEqual(zero.Elem().Interface(), test.Out) {
			t.Fatalf("Unmarshal's output doesn't match the expected value\n\tIn: %v\n\tExpected: %v\n\tOut: %v", mp, test.Out, zero)
		}
	}
}

func TestUnmarshalWithTags(t *testing.T) {
	expected := struct {
		DefaultName   string
		Renamed       string `redmap:"customname"`
		Ignored       string `redmap:"-"`
		OmittedString string `redmap:",omitempty"`
	}{
		DefaultName:   "defaultname",
		Renamed:       "renamed",
		Ignored:       "ignored",
		OmittedString: "should not change",
	}
	mp := map[string]string{
		"DefaultName": "defaultname",
		"customname":  "renamed",
		"Ignored":     "should be ignored",
	}
	copy := expected
	err := redmap.Unmarshal(mp, &copy)
	if err != nil {
		t.Fatalf("Unmarshal returned unexpected error %q", err)
	}
	if !reflect.DeepEqual(copy, expected) {
		t.Fatalf("Unmarshal's output doesn't match the expected value\n\tIn: %v\n\tExpected: %v\n\tOut: %v", mp, expected, copy)
	}
}

func TestMapUnmarshaler(t *testing.T) {
	intUn := stubIntMapUnmarshaler(666)
	tests := []struct {
		In  map[string]string
		Out redmap.StringMapUnmarshaler
	}{
		{In: map[string]string{"Field1": "value1", "Field2": "value2"}, Out: &stubMapUnmarshaler{Field1: "value1", Field2: "value2"}},
		{In: map[string]string{"Field1": "666"}, Out: &intUn},
	}
	for _, test := range tests {
		actual := reflect.New(reflect.TypeOf(test.Out).Elem())
		err := redmap.Unmarshal(test.In, actual.Interface())
		if err != nil {
			t.Fatalf("Unmarshal returned unexpected error %q", err)
		}
		if !reflect.DeepEqual(actual.Interface(), test.Out) {
			t.Fatalf("Unmarshal's output doesn't match the expected value\n\tIn: %v\n\tExpected: %v\n\tOut: %v", test.In, test.Out, actual)
		}
	}
}

func TestInnerMapUnmarshaler(t *testing.T) {
	expected := struct {
		RegularField string
		Struct       stubMapUnmarshaler `redmap:",inline"`
	}{
		RegularField: "regular",
		Struct:       stubMapUnmarshaler{Field1: "value1", Field2: "value2"},
	}
	mp := map[string]string{
		"RegularField":  "regular",
		"Struct.Field1": "value1",
		"Struct.Field2": "value2",
	}

	actual := reflect.New(reflect.TypeOf(expected))
	err := redmap.Unmarshal(mp, actual.Interface())
	if err != nil {
		t.Fatalf("Unmarshal returned unexpected error %q", err)
	}
	if !reflect.DeepEqual(actual.Elem().Interface(), expected) {
		t.Fatalf("Unmarshal's output doesn't match the expected value\n\tIn: %v\n\tExpected: %v\n\tOut: %v", mp, expected, actual)
	}
}
