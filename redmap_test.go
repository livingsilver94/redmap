package redmap_test

import (
	"encoding"
	"fmt"
	"reflect"
	"testing"

	"github.com/livingsilver94/redmap"
)

const (
	stringerOut      = "stub"
	textMarshalerOut = "stubtext"
)

type Stub struct{}

func (s Stub) String() string { return stringerOut }

type StubTextMarshaler struct{}

func (s StubTextMarshaler) MarshalText() ([]byte, error) { return []byte(textMarshalerOut), nil }

func TestValidType(t *testing.T) {
	var (
		stub Stub         = Stub{}
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
		stub *Stub        = nil
		ifac fmt.Stringer = stub
	)
	nils := []interface{}{nil, stub, ifac}
	for _, n := range nils {
		if res, err := redmap.Marshal(n); res != nil || err != nil {
			t.Fatalf("Marshal() with nil value of type %T must return nil", n)
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

		// Marshal interfaces by passing the real value.
		{In: struct{ V Stub }{Stub{}}, Out: map[string]string{"V": stringerOut}},
		{In: struct{ V StubTextMarshaler }{StubTextMarshaler{}}, Out: map[string]string{"V": textMarshalerOut}},

		// Marshal interfaces by interfaces.
		{In: struct{ V fmt.Stringer }{Stub{}}, Out: map[string]string{"V": stringerOut}},
		{In: struct{ V encoding.TextMarshaler }{StubTextMarshaler{}}, Out: map[string]string{"V": textMarshalerOut}},
	}
	for i, test := range tests {
		t.Log(i)
		out, err := redmap.Marshal(test.In)
		if err != nil {
			t.Fatalf("Marshal returned unexpected error %q", err)
		}
		if !reflect.DeepEqual(out, test.Out) {
			t.Fatalf("Marshal's output doesn't match the expected value\n\tIn: %v\n\tExpected: %v\n\tOut: %v", test.In, test.Out, out)
		}
	}
}

func TestUnexported(t *testing.T) {
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

func TestStructTags(t *testing.T) {
	stru := struct {
		DefaultName string
		Renamed     string `redmap:"customname"`
		Ignored     string `redmap:"-"`
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
