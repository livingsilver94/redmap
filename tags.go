package redmap

import (
	"reflect"
	"strings"
)

const (
	tagKeyword   = "redmap"
	tagSeparator = ","

	tagIgnore = "-"
	tagInline = "inline"
)

type structTags struct {
	name    string
	ignored bool
	inline  bool
}

func redmapTags(t reflect.StructTag) (structTags, bool) {
	str, has := t.Lookup(tagKeyword)
	if !has || str == "" {
		return structTags{}, false
	}

	if str == tagIgnore {
		return structTags{ignored: true}, true
	}
	tokens := strings.Split(str, tagSeparator)
	tags := structTags{}
	for _, t := range tokens {
		switch t {
		case tagInline:
			tags.inline = true
		default:
			tags.name = t
		}
	}
	return tags, true
}
