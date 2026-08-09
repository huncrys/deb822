// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2026 Kristof Bach <crys@crys.hu>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package deb822

import (
	"encoding"
	"reflect"
	"strings"
	"sync"
)

// TagKey is the struct tag consulted when mapping struct fields onto deb822
// field names.
//
// The grammar is:
//
//	debian:"Field-Name[,omitempty][,inline]"
//	debian:"-"
//
// A tag of "-" skips the field in both directions. An empty name part keeps
// the fallback name (see lookupFieldTag). Unknown options are ignored, so new
// options can be added without breaking older structs.
const TagKey = "debian"

// jsonTagKey is the fallback struct tag. Only its name part is honoured on the
// deb822 path; encoding/json options such as omitempty, omitzero or string
// have no meaning for a stanza and are ignored. This keeps structs that were
// written against the old encoding/json based implementation working without
// having to be retagged.
const jsonTagKey = "json"

var (
	textMarshalerType   = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// tagOptions holds the recognised options of a debian struct tag.
type tagOptions struct {
	// omitEmpty drops the field from the generated stanza when the Go value is
	// the zero value of its type.
	omitEmpty bool
	// inline flattens a named struct field into the surrounding stanza, the
	// way an anonymous embedded struct is flattened.
	inline bool
}

// parseTag splits a struct tag into its name part and its options.
func parseTag(tag string) (string, tagOptions) {
	var opts tagOptions

	name, rest, _ := strings.Cut(tag, ",")

	for rest != "" {
		var opt string
		opt, rest, _ = strings.Cut(rest, ",")

		switch opt {
		case "omitempty":
			opts.omitEmpty = true
		case "inline":
			opts.inline = true
		}
	}

	return name, opts
}

// lookupFieldTag resolves the deb822 name of a struct field.
//
// The resolution order is:
//
//  1. the debian tag, if the field carries one,
//  2. otherwise the name part of the json tag, if the field carries one,
//  3. otherwise the Go field name.
//
// The returned name is empty when neither tag names the field; the caller
// decides whether that means "use the Go field name" or "flatten this
// anonymous field". keep is false when the field is tagged "-".
func lookupFieldTag(sf reflect.StructField) (name string, opts tagOptions, keep bool) {
	if tag, ok := sf.Tag.Lookup(TagKey); ok {
		if tag == "-" {
			return "", opts, false
		}

		name, opts = parseTag(tag)

		return name, opts, true
	}

	if tag, ok := sf.Tag.Lookup(jsonTagKey); ok {
		if tag == "-" {
			return "", opts, false
		}

		// Only the name part carries over; the json options are meaningless
		// on the stanza path.
		name, _ = parseTag(tag)

		return name, opts, true
	}

	return "", opts, true
}

// fieldInfo is a single struct field, resolved to the deb822 field name it is
// carried by.
type fieldInfo struct {
	// name is the field name as it appears in a stanza.
	name string
	// lowerName is name folded to lower case, for the case insensitive lookup
	// mandated by Debian Policy 5.1.
	lowerName string
	// index is the path from the outermost struct down to the field, as
	// accepted by reflect.Value.FieldByIndex. It has more than one element for
	// fields reached through an embedded or inlined struct.
	index []int
	// depth is len(index) - 1; used only to resolve name collisions.
	depth int
	// omitEmpty reports whether the field carried the omitempty option.
	omitEmpty bool
}

// structInfo is the resolved field table of a struct type.
type structInfo struct {
	// fields is in declaration order, which is the order stanzas are written
	// in. Fields of an embedded or inlined struct sit at the position of the
	// embedding field.
	fields []fieldInfo
	// byName maps a lower cased field name to an index into fields.
	byName map[string]int
}

// structInfoCache memoises the field table of every struct type walked, keyed
// by reflect.Type.
var structInfoCache sync.Map

// cachedStructInfo returns the field table of t, building it on first use.
func cachedStructInfo(t reflect.Type) *structInfo {
	if cached, ok := structInfoCache.Load(t); ok {
		return cached.(*structInfo)
	}

	info := newStructInfo(t)
	structInfoCache.Store(t, info)

	return info
}

// newStructInfo resolves every field of t (and of the structs embedded or
// inlined into it) to a deb822 field name.
//
// Name collisions are resolved the way Go resolves promoted fields: the
// shallowest field wins. Unlike encoding/json a collision at equal depth is
// not dropped, the first field in declaration order wins, so the result is
// always deterministic.
func newStructInfo(t reflect.Type) *structInfo {
	var collected []fieldInfo
	collectFields(t, nil, map[reflect.Type]bool{t: true}, &collected)

	winner := make(map[string]int, len(collected))
	for i := range collected {
		if j, found := winner[collected[i].lowerName]; found && collected[j].depth <= collected[i].depth {
			continue
		}

		winner[collected[i].lowerName] = i
	}

	info := &structInfo{byName: make(map[string]int, len(winner))}

	for i := range collected {
		if winner[collected[i].lowerName] != i {
			continue
		}

		info.byName[collected[i].lowerName] = len(info.fields)
		info.fields = append(info.fields, collected[i])
	}

	return info
}

// collectFields appends the fields of t to out, in declaration order,
// descending into embedded and inlined structs. visiting guards against
// recursive embedding.
func collectFields(t reflect.Type, index []int, visiting map[reflect.Type]bool, out *[]fieldInfo) {
	for i := range t.NumField() {
		sf := t.Field(i)

		elem := sf.Type
		if elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
		}

		if sf.Anonymous {
			// An embedded struct of an unexported type may still have exported
			// fields, and those are reachable and settable. Anything else that
			// is unexported is not.
			if !sf.IsExported() && elem.Kind() != reflect.Struct {
				continue
			}
		} else if !sf.IsExported() {
			continue
		}

		name, opts, keep := lookupFieldTag(sf)
		if !keep {
			continue
		}

		fieldIndex := make([]int, 0, len(index)+1)
		fieldIndex = append(fieldIndex, index...)
		fieldIndex = append(fieldIndex, i)

		// An anonymous field that no tag gives a name to is flattened, the way
		// encoding/json flattens it, unless it renders as text on its own. The
		// inline option asks for the same treatment on a named field.
		flatten := opts.inline || (sf.Anonymous && name == "" && !marshalsAsText(sf.Type))
		if flatten && elem.Kind() == reflect.Struct {
			if visiting[elem] {
				continue
			}

			visiting[elem] = true
			collectFields(elem, fieldIndex, visiting, out)
			delete(visiting, elem)

			continue
		}

		if name == "" {
			name = sf.Name
		}

		*out = append(*out, fieldInfo{
			name:      name,
			lowerName: strings.ToLower(name),
			index:     fieldIndex,
			depth:     len(fieldIndex) - 1,
			omitEmpty: opts.omitEmpty,
		})
	}
}

// marshalsAsText reports whether values of type t (or pointers to them) know
// how to render themselves as text, in which case an anonymous field of that
// type is a value in its own right rather than a bag of fields to flatten.
func marshalsAsText(t reflect.Type) bool {
	if t.Implements(textMarshalerType) || t.Implements(textUnmarshalerType) {
		return true
	}

	ptr := reflect.PointerTo(t)

	return ptr.Implements(textMarshalerType) || ptr.Implements(textUnmarshalerType)
}
