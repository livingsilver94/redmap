package redmap_test

import (
	"reflect"
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

func TestMarshalScalar(t *testing.T) {
	tests := []struct {
		In  interface{}
		Out map[string]string
	}{
		{In: int(127), Out: map[string]string{"1": "127"}},
		{In: int8(127), Out: map[string]string{"1": "127"}},
		{In: int16(127), Out: map[string]string{"1": "127"}},
		{In: int32(127), Out: map[string]string{"1": "127"}},
		{In: int64(127), Out: map[string]string{"1": "127"}},
		{In: uint(255), Out: map[string]string{"1": "255"}},
		{In: uint8(255), Out: map[string]string{"1": "255"}},
		{In: uint16(255), Out: map[string]string{"1": "255"}},
		{In: uint32(255), Out: map[string]string{"1": "255"}},
		{In: uint64(255), Out: map[string]string{"1": "255"}},
		{In: uintptr(255), Out: map[string]string{"1": "255"}},
		{In: float32(123.456), Out: map[string]string{"1": "123.456"}},
		{In: float64(123.456), Out: map[string]string{"1": "123.456"}},
		{In: complex64(123.456 + 789.012i), Out: map[string]string{"1": "(123.456+789.012i)"}},
		{In: complex128(123.456 + 789.012i), Out: map[string]string{"1": "(123.456+789.012i)"}},
	}
	for repeat := 0; repeat < 2; repeat++ {
		for i, ts := range tests {
			out, err := redmap.Marshal(ts.In)
			if err != nil {
				t.Fatalf("Marshal() with scalar input %T %v returned error \"%v\"", ts.In, ts.In, err)
			}
			if !reflect.DeepEqual(out, ts.Out) {
				t.Fatalf("Marshal() with scalar input %T %v returned %v instead of %v", ts.In, ts.In, out, ts.Out)
			}

			// Repeat test with pointers.
			p := reflect.New(reflect.TypeOf(tests[i].In))
			p.Elem().Set(reflect.ValueOf(tests[i].In))
			tests[i].In = p.Interface()
		}
	}
}
