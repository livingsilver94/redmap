package redmap

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
)

func scalartoString(val reflect.Value) (string, error) {
	switch val.Kind() {
	case reflect.Ptr:
		return scalartoString(val.Elem())
	case reflect.Bool:
		return strconv.FormatBool(val.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(val.Uint(), 10), nil
	case reflect.Float32:
		return strconv.FormatFloat(val.Float(), 'f', -1, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'f', -1, 64), nil
	case reflect.Complex64:
		return strconv.FormatComplex(val.Complex(), 'f', -1, 64), nil
	case reflect.Complex128:
		return strconv.FormatComplex(val.Complex(), 'f', -1, 128), nil
	case reflect.String:
		return val.String(), nil
	case reflect.Interface, reflect.Struct:
		typ := val.Type()
		if typ.Implements(reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()) {
			str, err := val.Interface().(encoding.TextMarshaler).MarshalText()
			return string(str), err
		}
		if typ.Implements(reflect.TypeOf((*fmt.Stringer)(nil)).Elem()) {
			return val.Interface().(fmt.Stringer).String(), nil
		}
	}
	return "", fmt.Errorf("%s doesn't implement TextMarshaler or Stringer", val.Type())
}
