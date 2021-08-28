package redmap

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
)

const inlineSep = "."

func Marshal(stru interface{}) (map[string]string, error) {
	isStru, val := isStruct(reflect.ValueOf(stru))
	if !isStru {
		return nil, fmt.Errorf("cannot marshal type %T", stru)
	}
	if !val.IsValid() {
		// v is nil, so return a nil map.
		return nil, nil
	}

	ret := make(map[string]string)
	return ret, marshalRecurse(ret, "", val)
}

// isStruct returns whether val is a struct. Along with the boolean flag, isStruct also returns
// the actual struct value in the case the original argument was buried under layers of interfaces or pointers.
func isStruct(val reflect.Value) (bool, reflect.Value) {
	k := val.Kind()
	if k == reflect.Interface || k == reflect.Ptr {
		return isStruct(val.Elem())
	}
	// reflect.Invalid is a valid type because it's the zero value of interface{}.
	return k == reflect.Struct || k == reflect.Invalid, val
}

// marshalRecurse marshal a struct represented by val into a map[string]string.
// Given its recursive nature, it needs to remember the intermediate results:
// mp is the temporary marshal result; prefix is the prefix applied to a field
// name in case of an inlined inner struct.
func marshalRecurse(mp map[string]string, prefix string, stru reflect.Value) error {
	typ := stru.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			// We don't want to marshal unexported fields. PkgPath is empty for exported fields.
			// TODO: In Go 1.17, use field.IsExported().
			continue
		}
		tags, ok := redmapTags(field.Tag)
		value := stru.Field(i)
		if tags.ignored || (tags.omitempty && value.IsZero()) {
			continue
		}
		if !ok || tags.name == "" {
			tags.name = field.Name
		}

		for value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		if tags.inline {
			if kind := value.Kind(); kind != reflect.Struct {
				return fmt.Errorf("cannot inline %s. Only structs are allowed", kind)
			}
			err := marshalRecurse(mp, prefix+tags.name+inlineSep, value)
			if err != nil {
				return err
			}
		} else {
			str, err := fieldToString(value)
			if err != nil {
				return err
			}
			mp[prefix+tags.name] = str
		}
	}
	return nil
}

func fieldToString(val reflect.Value) (string, error) {
	switch val.Kind() {
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
