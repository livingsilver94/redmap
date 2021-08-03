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
	return nil, nil
}

func isValidType(val reflect.Value) bool {
	k := val.Kind()
	if k == reflect.Interface || k == reflect.Ptr {
		return isValidType(val.Elem())
	}
	// reflect.Invalid is a valid type because it's the zero value of interface{}.
	return k == reflect.Struct || k == reflect.Invalid
}

func isNil(val reflect.Value) bool {
	if !val.IsValid() {
		// This is the zero value of interface{}, so it's a nil.
		return true
	}
	switch val.Kind() {
	case reflect.Interface, reflect.Ptr:
		return val.IsNil()
	default:
		return false
	}
}
