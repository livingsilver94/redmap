package redmap

import (
	"reflect"
	"strings"
)

const (
	tagKeyword   = "redmap"
	tagSeparator = ","

	tagIgnore    = "-"
	tagInline    = "inline"
	tagOmitEmpty = "omitempty"
)

type structTags struct {
	name      string
	ignored   bool
	inline    bool
	omitempty bool
}

func redmapTags(t reflect.StructTag) structTags {
	str, has := t.Lookup(tagKeyword)
	if !has || str == "" {
		return structTags{}
	}

	if str == tagIgnore {
		return structTags{ignored: true}
	}

	toks := strings.Split(str, tagSeparator)
	tags := structTags{name: toks[0]}
	for _, t := range toks[1:] {
		switch t {
		case tagInline:
			tags.inline = true
		case tagOmitEmpty:
			tags.omitempty = true
		}
	}
	return tags
}
