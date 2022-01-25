package redmap

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// StringMapUnmarshaler is the interface implemented by types that can unmarshal themselves
// from a map of strings. Implementations must copy the given map if they wish to modify it.
type StringMapUnmarshaler interface {
	UnmarshalStringMap(map[string]string) error
}

// Unmarshal sets v's fields according to its map representation contained by data.
// v must be a pointer to struct or an interface. Neither data nor v can be nil.
//
// Unmarshal uses the inverse of the encodings that Marshal uses, so all the types supported
// by it are also supported in Unmarshal, except the interfaces: only encoding.TextUnmarshaler
// can be unmarshaled.
//
// The decoding of each struct field can be customized by the format string documented in Marshal.
func Unmarshal(data map[string]string, v interface{}) error {
	if data == nil {
		return errIs("map passed", ErrNilValue)
	}
	val, err := ptrValidValue(v)
	if err != nil {
		return err
	}
	return unmarshalRecursive(data, "", val)
}

func ptrValidValue(v interface{}) (reflect.Value, error) {
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
	if kin == reflect.Invalid {
		return reflect.Value{}, errIs(reflect.TypeOf(v), ErrNilValue)
	}
	return val, nil
}

func unmarshalRecursive(mp map[string]string, prefix string, stru reflect.Value) error {
	if ptr := stru.Addr(); ptr.Type().Implements(mapUnmarshalerType) {
		return mapToStruct(mp, prefix, ptr)
	}
	if stru.Kind() != reflect.Struct {
		return errIs(stru.Type(), ErrNoCodec)
	}
	typ := stru.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			// We don't want to unmarshal unexported fields. PkgPath is empty for exported fields.
			// TODO: In Go 1.17, use field.IsExported().
			continue
		}
		tags := redmapTags(field.Tag)
		if tags.ignored {
			continue
		}
		value := stru.Field(i)
		if tags.name == "" {
			tags.name = field.Name
		}
		tags.name = prefix + tags.name

		for value.Kind() == reflect.Ptr {
			if value.IsNil() {
				if !value.CanSet() {
					return fmt.Errorf("cannot set embedded pointer to unexported type %s", value.Elem().Type())
				}
				value.Set(reflect.New(value.Type().Elem()))
			}
			value = value.Elem()
		}

		if tags.inline {
			err := unmarshalRecursive(mp, tags.name+inlineSep, value)
			if err != nil {
				return err
			}
		} else {
			str, ok := mp[tags.name]
			if !ok {
				continue
			}
			err := stringToField(str, value, tags.omitempty)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func mapToStruct(mp map[string]string, prefix string, stru reflect.Value) error {
	if prefix != "" {
		// FIXME: Creating a submap is O(n). Can we think of a better algorithm?
		subMP := make(map[string]string, len(mp))
		for k, v := range mp {
			if !strings.HasPrefix(k, prefix) {
				continue
			}
			subMP[k[len(prefix):]] = v
		}
		mp = subMP
	}
	return stru.Interface().(StringMapUnmarshaler).UnmarshalStringMap(mp)
}

func stringToField(str string, field reflect.Value, omitempty bool) error {
	addr := field.Addr() // Unmarshaling always requires a pointer receiver.
	if addr.Type().Implements(textUnmarshalerType) {
		return addr.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(str))
	}

	var (
		val reflect.Value
		err error
	)
	switch field.Kind() {
	case reflect.Bool:
		v, e := strconv.ParseBool(str)
		val, err = reflect.ValueOf(v), e
	case reflect.Int:
		v, e := strconv.ParseInt(str, 10, 0)
		val, err = reflect.ValueOf(int(v)), e
	case reflect.Int8:
		v, e := strconv.ParseInt(str, 10, 8)
		val, err = reflect.ValueOf(int8(v)), e
	case reflect.Int16:
		v, e := strconv.ParseInt(str, 10, 16)
		val, err = reflect.ValueOf(int16(v)), e
	case reflect.Int32:
		v, e := strconv.ParseInt(str, 10, 32)
		val, err = reflect.ValueOf(int32(v)), e
	case reflect.Int64:
		v, e := strconv.ParseInt(str, 10, 64)
		val, err = reflect.ValueOf(v), e
	case reflect.Uint:
		v, e := strconv.ParseUint(str, 10, 0)
		val, err = reflect.ValueOf(uint(v)), e
	case reflect.Uint8:
		v, e := strconv.ParseUint(str, 10, 8)
		val, err = reflect.ValueOf(uint8(v)), e
	case reflect.Uint16:
		v, e := strconv.ParseUint(str, 10, 16)
		val, err = reflect.ValueOf(uint16(v)), e
	case reflect.Uint32:
		v, e := strconv.ParseUint(str, 10, 32)
		val, err = reflect.ValueOf(uint32(v)), e
	case reflect.Uint64:
		v, e := strconv.ParseUint(str, 10, 64)
		val, err = reflect.ValueOf(v), e
	case reflect.Float32:
		v, e := strconv.ParseFloat(str, 32)
		val, err = reflect.ValueOf(float32(v)), e
	case reflect.Float64:
		v, e := strconv.ParseFloat(str, 64)
		val, err = reflect.ValueOf(v), e
	case reflect.Complex64:
		v, e := strconv.ParseComplex(str, 64)
		val, err = reflect.ValueOf(complex64(v)), e
	case reflect.Complex128:
		v, e := strconv.ParseComplex(str, 128)
		val, err = reflect.ValueOf(v), e
	case reflect.String:
		val, err = reflect.ValueOf(str), nil
	default:
		return fmt.Errorf("%s doesn't implement TextUnmarshaler", addr)
	}
	if err != nil {
		return err
	}
	if omitempty && val.IsZero() {
		return nil
	}
	field.Set(val)
	return nil
}

var (
	mapUnmarshalerType  = reflect.TypeOf(new(StringMapUnmarshaler)).Elem()
	textUnmarshalerType = reflect.TypeOf(new(encoding.TextUnmarshaler)).Elem()
)
