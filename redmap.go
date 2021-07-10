package redmap

import (
	"fmt"
	"reflect"
)

func Marshal(v interface{}) (map[string]string, error) {
	val := reflect.ValueOf(v)
	if !isValidType(val) {
		return nil, fmt.Errorf("cannot marshal type %T", v)
	}
	if isNil(val) {
		return nil, nil
	}
	return marshal(val), nil
}

func isValidType(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Func, reflect.Chan:
		return false
	default:
		return true
	}
}

func isNil(val reflect.Value) bool {
	if !val.IsValid() {
		// This is a nil value.
		return true
	}
	switch val.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map:
		return val.IsNil()
	default:
		return false
	}
}

func marshal(val reflect.Value) map[string]string {
	switch val.Kind() {
	case reflect.Ptr, reflect.Interface:
		return marshal(val.Elem())
	case reflect.Array, reflect.Slice:
	case reflect.Struct:
	default:
		return map[string]string{"1": scalarToString(val)}
	}
	return nil
}
