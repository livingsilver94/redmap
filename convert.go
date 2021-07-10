package redmap

import (
	"fmt"
	"reflect"
	"strconv"
)

func scalarToString(val reflect.Value) string {
	switch val.Kind() {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr:
		return strconv.FormatUint(val.Uint(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(val.Float(), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'f', -1, 64)
	case reflect.Complex64:
		return strconv.FormatComplex(val.Complex(), 'f', -1, 64)
	case reflect.Complex128:
		return strconv.FormatComplex(val.Complex(), 'f', -1, 128)
	case reflect.String:
		return val.String()
	case reflect.Bool:
		return strconv.FormatBool(val.Bool())
	}
	if val.Type().Implements(reflect.TypeOf(fmt.Stringer(nil))) {
		return val.Interface().(fmt.Stringer).String()
	}
	panic(fmt.Sprintf("%s is not a scalar value", val.Type().String()))
}
