package redmap

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
)

const inlineSep = "."

// Marshal returns the map[string]string representation of v, which must be a struct,
// by reading every exported field and translating it into a (key, value) pair to be
// added to the resulting map. Interfaces or pointers to struct are also accepted.
//
// Marshal converts all non-reference built-in types except arrays, plus
// structs implementing encoding.TextMarshaler or fmt.Stringer, checked in this exact order.
//
// The encoding of each struct field can be customized by the format string stored under the "redmap"
// key in the struct field's tag. The format string gives the name of the field, possibly followed by
// a comma-separated list of options. The name may be empty in order to specify options without
// overriding the default field name. If the format string is equal to "-", the struct field
// is excluded from marshaling.
//
// Examples of struct field tags and their meanings:
//
//   // Field appears in the map as key "customName".
//   Field int `redmap:"customName"`
//
//   // Field appears in the map as key "customName" and
//   // the field is omitted from the map if its value
//   // is empty as defined by the Go language.
//   Field int `redmap:"customName,omitempty"`
//
//   // Field appears in the map as key "Field" (the default), but
//   // the field is skipped if empty. Note the leading comma.
//   Field int `redmap:",omitempty"`
//
//   // Field is ignored by this package.
//   Field int `redmap:"-"`
//
//   // Field appears in the map as key "-".
//   Field int `redmap:"-,"`
//
//   // Field must be a struct. Field is flattened and its fields
//   // are added to the map as (key, value) pairs, where the keys
//   // are constructed in the "customName.subFieldName" format.
//   Field int `redmap:"customName,inline"`
func Marshal(v interface{}) (map[string]string, error) {
	val, err := structValue(v)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]string)
	return ret, marshalRecursive(ret, "", val)
}

func structValue(v interface{}) (reflect.Value, error) {
	val := reflect.ValueOf(v)
	kin := val.Kind()
	for kin == reflect.Interface || kin == reflect.Ptr {
		val = val.Elem()
		kin = val.Kind()
	}
	switch kin {
	case reflect.Struct:
		return val, nil
	case reflect.Invalid:
		return reflect.Value{}, ErrNilValue
	default:
		return reflect.Value{}, errIs(val.Type(), ErrNotStruct)
	}
}

// marshalRecursive marshal a struct represented by val into a map[string]string.
// Given its recursive nature, it needs to remember the intermediate results:
// mp is the temporary marshal result; prefix is the prefix applied to a field
// name in case of an inlined inner struct.
func marshalRecursive(mp map[string]string, prefix string, stru reflect.Value) error {
	typ := stru.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			// We don't want to marshal unexported fields. PkgPath is empty for exported fields.
			// TODO: In Go 1.17, use field.IsExported().
			continue
		}
		tags := redmapTags(field.Tag)
		value := stru.Field(i)
		if tags.ignored || (tags.omitempty && value.IsZero()) {
			continue
		}
		if tags.name == "" {
			tags.name = field.Name
		}

		for value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		if tags.inline {
			if kind := value.Kind(); kind != reflect.Struct {
				return fmt.Errorf("cannot inline: %w", errIs(value.Type(), ErrNotStruct))
			}
			err := marshalRecursive(mp, prefix+tags.name+inlineSep, value)
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
	typ := val.Type()
	if typ.Implements(textMarshalerType) {
		str, err := val.Interface().(encoding.TextMarshaler).MarshalText()
		return string(str), err
	}
	if typ.Implements(stringerType) {
		return val.Interface().(fmt.Stringer).String(), nil
	}

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
	}
	return "", fmt.Errorf("%s doesn't implement TextMarshaler or Stringer", val.Type())
}

var (
	textMarshalerType = reflect.TypeOf(new(encoding.TextMarshaler)).Elem()
	stringerType      = reflect.TypeOf(new(fmt.Stringer)).Elem()
)
