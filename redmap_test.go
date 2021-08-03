package redmap_test

import (
	"fmt"
	"testing"

	"github.com/livingsilver94/redmap"
)

type Stub struct{}

func (s Stub) String() string { return "stub" }

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
