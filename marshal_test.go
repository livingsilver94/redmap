package redmap_test

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/livingsilver94/redmap"
)

const (
	stringerOut      = "stub"     // stringerOut is the output of fmt.Stringer implementations.
	textMarshalerOut = "stubtext" // textMarshalerOut  is the output of encoding.TextMarshaler implementations.
)

var (
	// mapMarshalerOut is the output of redmap.StringMapMarshaler implementations.
	mapMarshalerOut = map[string]string{"field1": "value1", "field2": "value2"}
)

// stubStringer implements the fmt.Stringer interface.
type stubStringer struct{}

func (s stubStringer) String() string { return stringerOut }

// stubIntStringer is an int that implements fmt.Stringer,
// so that we can test if a non-struct type is correctly handled as an interface.
type stubIntStringer int

func (s stubIntStringer) String() string { return stringerOut }

// stubTextMarshaler implements the encoding.TextMarshaler interface.
type stubTextMarshaler struct{}

func (s stubTextMarshaler) MarshalText() ([]byte, error) { return []byte(textMarshalerOut), nil }

// stubMapMarshaler implements the redmap.StringMapMarshaler interface.
type stubMapMarshaler struct{}

func (s stubMapMarshaler) MarshalStringMap() (map[string]string, error) {
	return mapMarshalerOut, nil
}

// stubIntMapMarshaler is an int that implements redmap.StringMapMarshaler,
// so that we can test if a non-struct type is correctly handled as an interface.
type stubIntMapMarshaler int

func (s stubIntMapMarshaler) MarshalStringMap() (map[string]string, error) {
	return mapMarshalerOut, nil
}

func TestMarshalValidType(t *testing.T) {
	var (
		stub stubStringer = stubStringer{}
		ifac fmt.Stringer = stub
	)
	types := []interface{}{
		stub,  // Struct.
		ifac,  // Interface.
		&stub, // Pointer to struct.
		&ifac, // Pointer to interface.
	}
	for _, typ := range types {
		if _, err := redmap.Marshal(typ); err != nil {
			t.Fatalf("Marshal() of valid type %T must not return error", typ)
		}
	}
}

func TestMarshalNil(t *testing.T) {
	var (
		stub *stubStringer = nil
		ifac fmt.Stringer  = stub
	)
	nils := []interface{}{nil, stub, ifac}
	for _, n := range nils {
		if _, err := redmap.Marshal(n); !errors.Is(err, redmap.ErrNilValue) {
			t.Fatalf("Marshal() with nil value of type %T did not return error", n)
		}
	}
}

func TestMarshalInvalidType(t *testing.T) {
	noStruct := 45
	tests := []interface{}{noStruct, &noStruct}
	for _, test := range tests {
		_, err := redmap.Marshal(test)
		if !errors.Is(err, redmap.ErrNotStruct) {
			t.Fatalf("Unmarshal returned error %q but %q was expected", err, redmap.ErrNotStruct)
		}
	}
}

func TestMarshalScalars(t *testing.T) {
	tests := []struct {
		In  interface{}
		Out map[string]string
	}{
		{In: struct{ V bool }{true}, Out: map[string]string{"V": "true"}},
		{In: struct{ V int }{100}, Out: map[string]string{"V": "100"}},
		{In: struct{ V uint }{100}, Out: map[string]string{"V": "100"}},
		{In: struct{ V float32 }{100.1}, Out: map[string]string{"V": "100.1"}},
		{In: struct{ V float64 }{100.1}, Out: map[string]string{"V": "100.1"}},
		{In: struct{ V complex64 }{100.1 + 80.1i}, Out: map[string]string{"V": "(100.1+80.1i)"}},
		{In: struct{ V complex128 }{100.1 + 80.1i}, Out: map[string]string{"V": "(100.1+80.1i)"}},
		{In: struct{ V string }{"str"}, Out: map[string]string{"V": "str"}},

		// // Marshal interfaces by passing the real value.
		{In: struct{ V stubStringer }{stubStringer{}}, Out: map[string]string{"V": stringerOut}},
		{In: struct{ V stubIntStringer }{stubIntStringer(100)}, Out: map[string]string{"V": stringerOut}},
		{In: struct{ V stubTextMarshaler }{stubTextMarshaler{}}, Out: map[string]string{"V": textMarshalerOut}},

		// Marshal interfaces by interfaces.
		{In: struct{ V fmt.Stringer }{stubStringer{}}, Out: map[string]string{"V": stringerOut}},
		{In: struct{ V fmt.Stringer }{stubIntStringer(100)}, Out: map[string]string{"V": stringerOut}},
		{In: struct{ V encoding.TextMarshaler }{stubTextMarshaler{}}, Out: map[string]string{"V": textMarshalerOut}},
	}
	for _, test := range tests {
		out, err := redmap.Marshal(test.In)
		if err != nil {
			t.Fatalf("Marshal returned unexpected error %q", err)
		}
		if !reflect.DeepEqual(out, test.Out) {
			t.Fatalf("Marshal's output doesn't match the expected value\n\tIn: %v\n\tExpected: %v\n\tOut: %v", test.In, test.Out, out)
		}
	}
}

func TestMarshalInnerStructs(t *testing.T) {
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
		In  interface{}
		Out map[string]string
	}{
		{In: Root1Level{Inner: Inner1Level{String: "oneLevel"}}, Out: map[string]string{"Inner.String": "oneLevel"}},
		{In: Root2Level{Inner: Inner2Level{Inner: Inner1Level{String: "twoLevel"}}}, Out: map[string]string{"Inner.Inner.String": "twoLevel"}},
		{In: RootWithPointer{Inner: &Inner1Level{String: "oneLevel"}}, Out: map[string]string{"Inner.String": "oneLevel"}},
	}
	for _, test := range tests {
		out, err := redmap.Marshal(test.In)
		if err != nil {
			t.Fatalf("Marshal returned unexpected error %q", err)
		}
		if !reflect.DeepEqual(out, test.Out) {
			t.Fatalf("Marshal's output doesn't match the expected value\n\tIn: %v\n\tExpected: %v\n\tOut: %v", test.In, test.Out, out)
		}
	}
}

func TestMarshalUnexported(t *testing.T) {
	stru := struct {
		Exported   string
		unexported string
	}{
		Exported:   "exported",
		unexported: "should be invisible",
	}
	expected := map[string]string{
		"Exported": "exported",
	}
	out, err := redmap.Marshal(stru)
	if err != nil {
		t.Fatalf("Marshal returned unexpected error %q", err)
	}
	if !reflect.DeepEqual(out, expected) {
		t.Fatal("Marshal's output includes an unexported struct field")
	}
}

func TestMarshalWithTags(t *testing.T) {
	stru := struct {
		DefaultName      string
		Renamed          string      `redmap:"customname"`
		Ignored          string      `redmap:"-"`
		OmittedString    string      `redmap:",omitempty"`
		OmittedInterface interface{} `redmap:",omitempty"`
	}{
		DefaultName: "defaultname",
		Renamed:     "renamed",
		Ignored:     "ignored",
	}
	expected := map[string]string{
		"DefaultName": "defaultname",
		"customname":  "renamed",
	}
	out, err := redmap.Marshal(stru)
	if err != nil {
		t.Fatalf("Marshal returned unexpected error %q", err)
	}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("Marshal's output doesn't respect struct tags\n\tExpected: %v\n\tOut: %v", expected, out)
	}
}

func TestMapMarshaler(t *testing.T) {
	tests := []struct {
		In  redmap.StringMapMarshaler
		Out map[string]string
	}{
		{In: stubMapMarshaler{}, Out: mapMarshalerOut},
		{In: stubIntMapMarshaler(666), Out: mapMarshalerOut},
	}
	for _, test := range tests {
		out, err := redmap.Marshal(test.In)
		if err != nil {
			t.Fatalf("Marshal returned unexpected error %q", err)
		}
		if !reflect.DeepEqual(out, test.Out) {
			t.Fatalf("Marshal's output doesn't match the expected value\n\tIn: %v\n\tExpected: %v\n\tOut: %v", test.In, test.Out, out)
		}
	}
}

func TestInnerMapMarshaler(t *testing.T) {
	stru := struct {
		RegularField string
		Struct       stubMapMarshaler `redmap:",inline"`
	}{
		RegularField: "regular",
		Struct:       stubMapMarshaler{},
	}
	expected := map[string]string{"RegularField": "regular"}
	for k, v := range mapMarshalerOut {
		expected["Struct."+k] = v
	}

	out, err := redmap.Marshal(stru)
	if err != nil {
		t.Fatalf("Marshal returned unexpected error %q", err)
	}
	if !reflect.DeepEqual(out, expected) {
		t.Fatalf("Marshal's output doesn't respect struct tags\n\tExpected: %v\n\tOut: %v", expected, out)
	}
}
