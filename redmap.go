package redmap

import (
	"fmt"
	"reflect"
)

func Marshal(v interface{}) (map[string]string, error) {
	valid, val := isValidType(reflect.ValueOf(v))
	if !valid {
		return nil, fmt.Errorf("cannot marshal type %T", v)
	}
	if !val.IsValid() {
		// v is nil, so return a nil map.
		return nil, nil
	}
	return marshal(val)
}

// isValidType returns whether val is a valid struct to unmarshal.
// The returned values are a boolean flag telling whether the value is valid,
// and the actual valid value in the case the original argument was buried under layers of interfaces or pointers.
func isValidType(val reflect.Value) (bool, reflect.Value) {
	k := val.Kind()
	if k == reflect.Interface || k == reflect.Ptr {
		return isValidType(val.Elem())
	}
	// reflect.Invalid is a valid type because it's the zero value of interface{}.
	return k == reflect.Struct || k == reflect.Invalid, val
}

func marshal(val reflect.Value) (map[string]string, error) {
	if k := val.Kind(); k == reflect.Ptr || k == reflect.Interface {
		return marshal(val.Elem())
	}
	typ := val.Type()
	ret := make(map[string]string, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tags, ok := redmapTags(field.Tag)
		if tags.ignored {
			continue
		}
		if !ok {
			tags.name = field.Name
		}
		str, err := scalartoString(val.Field(i))
		if err != nil {
			return ret, err
		}
		ret[tags.name] = str
	}
	return ret, nil
}
