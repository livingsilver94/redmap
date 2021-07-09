package redmap_test

import (
	"testing"

	"github.com/livingsilver94/redmap"
)

func TestInvalidType(t *testing.T) {
	types := []interface{}{
		*new(func(int)),
		*new(chan int),
	}
	for _, typ := range types {
		if _, err := redmap.Marshal(typ); err == nil {
			t.Fatalf("Marshal() of invalid type %T must return error", typ)
		}
	}
}

func TestMarshalNil(t *testing.T) {
	nils := []interface{}{
		nil,
		*new(*int),
		*new([]int),
		*new(map[int]int),
	}
	for _, n := range nils {
		if res, err := redmap.Marshal(n); res != nil || err != nil {
			t.Fatalf("Marshal() with nil value of type %T must return nil", n)
		}
	}
}
