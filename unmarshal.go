package redmap

import (
	"reflect"
)

func Unmarshal(data map[string]string, v interface{}) error {
	if data == nil {
		return errIs("map passed", ErrNilValue)
	}
	val, err := ptrStructValue(v)
	if err != nil {
		return err
	}
	return unmarshalRecursive(data, val)
}

func ptrStructValue(v interface{}) (reflect.Value, error) {
	val := reflect.ValueOf(v)
	kin := val.Kind()

	switch kin {
	case reflect.Ptr:
	case reflect.Invalid:
		return reflect.Value{}, errIs("argument provided", ErrNilValue)
	default:
		return reflect.Value{}, errIs(val.Type(), ErrNotPointer)
	}

	for kin == reflect.Ptr {
		val = val.Elem()
		kin = val.Kind()
	}

	switch kin {
	case reflect.Struct:
		return val, nil
	case reflect.Invalid:
		return reflect.Value{}, errIs(reflect.TypeOf(v), ErrNilValue)
	default:
		return reflect.Value{}, errIs(val.Type(), ErrNotStruct)
	}
}

func unmarshalRecursive(data map[string]string, val reflect.Value) error {
	return nil
}
