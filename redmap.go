package redmap

import (
	"fmt"
	"reflect"
)

func Marshal(v interface{}) (map[string]string, error) {
	typ := reflect.TypeOf(v)
	val := reflect.ValueOf(v)

	if !isValidType(typ) {
		return nil, fmt.Errorf("cannot marshal type %T", v)
	}
	if isNil(typ, val) {
		return nil, nil
	}
	return nil, nil
}

func isValidType(typ reflect.Type) bool {
	if typ == nil {
		return true
	}
	switch typ.Kind() {
	case reflect.Func, reflect.Chan:
		return false
	default:
		return true
	}
}

func isNil(typ reflect.Type, val reflect.Value) bool {
	if typ == nil {
		return true
	}
	switch typ.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map:
		return val.IsNil()
	default:
		return false
	}
}
