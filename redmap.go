package redmap

import (
	"fmt"
	"reflect"
)

func Marshal(stru interface{}) (map[string]string, error) {
	isStru, val := isStruct(reflect.ValueOf(stru))
	if !isStru {
		return nil, fmt.Errorf("cannot marshal type %T", stru)
	}
	if !val.IsValid() {
		// v is nil, so return a nil map.
		return nil, nil
	}
	return marshal(val)
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

func marshal(val reflect.Value) (map[string]string, error) {
	typ := val.Type()
	ret := make(map[string]string, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			// We don't want to marshal unexported fields. PkgPath is empty for exported fields.
			// TODO: In Go 1.17, use field.IsExported().
			continue
		}
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
