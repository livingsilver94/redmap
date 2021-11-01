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
	stringerOut      = "stub"
	textMarshalerOut = "stubtext"
)

type StubStringer struct{}

func (s StubStringer) String() string { return stringerOut }

// StubIntStringer implements fmt.Stringer but doesn't rely on an
// underlying struct. Useful to test whether we can detect
// interfaces independently from their underlying type.
type StubIntStringer int

func (s StubIntStringer) String() string { return stringerOut }

type StubTextMarshaler struct{}

func (s StubTextMarshaler) MarshalText() ([]byte, error) { return []byte(textMarshalerOut), nil }

func TestMarshalValidType(t *testing.T) {
	var (
		stub StubStringer = StubStringer{}
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
		stub *StubStringer = nil
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
		{In: struct{ V StubStringer }{StubStringer{}}, Out: map[string]string{"V": stringerOut}},
		{In: struct{ V StubIntStringer }{StubIntStringer(100)}, Out: map[string]string{"V": stringerOut}},
		{In: struct{ V StubTextMarshaler }{StubTextMarshaler{}}, Out: map[string]string{"V": textMarshalerOut}},

		// Marshal interfaces by interfaces.
		{In: struct{ V fmt.Stringer }{StubStringer{}}, Out: map[string]string{"V": stringerOut}},
		{In: struct{ V fmt.Stringer }{StubIntStringer(100)}, Out: map[string]string{"V": stringerOut}},
		{In: struct{ V encoding.TextMarshaler }{StubTextMarshaler{}}, Out: map[string]string{"V": textMarshalerOut}},
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
