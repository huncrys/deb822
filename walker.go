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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// marshalStruct renders a struct into a Stanza.
//
// Fields are visited in declaration order, so a stanza comes out in the order
// the struct spells its fields out. A field is left out of the stanza when:
//
//   - it renders to the empty string. A "Field:" line with no value carries no
//     information in deb822, so emptiness is decided on the rendered text, not
//     on the Go value, and it applies whether or not the field asked for
//     omitempty;
//   - it is a nil pointer, or sits behind one;
//   - it carries the omitempty option and holds the zero value of its type.
//     This is what keeps a numeric field whose zero renders as "0" out of the
//     stanza.
func marshalStruct(data reflect.Value) (*Stanza, error) {
	if data.Kind() != reflect.Struct {
		return nil, errors.New("can only Encode a Struct")
	}

	info := cachedStructInfo(data.Type())

	stanza := &Stanza{}

	for i := range info.fields {
		field := &info.fields[i]

		value, ok := fieldByIndex(data, field.index)
		if !ok {
			continue
		}

		if field.omitEmpty && isEmptyValue(value) {
			continue
		}

		text, ok, err := marshalFieldText(value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal field %q: %w", field.name, err)
		}

		if !ok || text == "" {
			continue
		}

		// Multi line values are handed over as they are; Stanza.WriteTo turns
		// the embedded newlines into continuation lines.
		stanza.Set(field.name, text)
	}

	return stanza, nil
}

// unmarshalStanza fills a struct from a Stanza.
//
// Field names are matched case insensitively, as Debian Policy 5.1 requires.
// Stanza fields the struct has no home for are ignored, and struct fields the
// stanza does not mention are left untouched. Fields present with an empty
// value are treated as absent.
func unmarshalStanza(stanza Stanza, into reflect.Value) error {
	if into.Kind() != reflect.Struct {
		return errors.New("can only Decode a Struct")
	}

	info := cachedStructInfo(into.Type())

	for _, key := range stanza.Order {
		text := stanza.Values[key]
		if text == "" {
			continue
		}

		i, found := info.byName[strings.ToLower(key)]
		if !found {
			continue
		}

		field := &info.fields[i]

		value, err := fieldByIndexAlloc(into, field.index)
		if err != nil {
			return fmt.Errorf("failed to unmarshal field %q: %w", key, err)
		}

		if err := unmarshalFieldText(value, text); err != nil {
			return fmt.Errorf("failed to unmarshal field %q: %w", key, err)
		}
	}

	return nil
}

// fieldByIndex walks an index path down to a field. It reports false when the
// path runs through a nil pointer, in which case there is no value to render.
func fieldByIndex(value reflect.Value, index []int) (reflect.Value, bool) {
	for _, i := range index {
		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return reflect.Value{}, false
			}

			value = value.Elem()
		}

		value = value.Field(i)
	}

	return value, true
}

// fieldByIndexAlloc walks an index path down to a field, allocating the
// intermediate pointers it finds unset along the way.
func fieldByIndexAlloc(value reflect.Value, index []int) (reflect.Value, error) {
	for _, i := range index {
		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				if !value.CanSet() {
					return reflect.Value{}, fmt.Errorf("cannot set embedded pointer of type %s", value.Type())
				}

				value.Set(reflect.New(value.Type().Elem()))
			}

			value = value.Elem()
		}

		value = value.Field(i)
	}

	return value, nil
}

// marshalFieldText renders a single value as deb822 field text. The boolean
// result reports whether there is a value to write at all; it is false for nil
// pointers and nil interfaces.
func marshalFieldText(value reflect.Value) (string, bool, error) {
	if kind := value.Kind(); kind == reflect.Pointer || kind == reflect.Interface {
		if value.IsNil() {
			return "", false, nil
		}
	}

	if marshaler, ok := asTextMarshaler(value); ok {
		text, err := marshaler.MarshalText()
		if err != nil {
			return "", false, err
		}

		return string(text), true, nil
	}

	switch value.Kind() {
	case reflect.Pointer, reflect.Interface:
		return marshalFieldText(value.Elem())
	case reflect.String:
		return value.String(), true, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10), true, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(value.Uint(), 10), true, nil
	case reflect.Bool:
		return strconv.FormatBool(value.Bool()), true, nil
	case reflect.Float32:
		return strconv.FormatFloat(value.Float(), 'g', -1, 32), true, nil
	case reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'g', -1, 64), true, nil
	}

	return "", false, fmt.Errorf("unsupported type: %s", value.Type())
}

// unmarshalFieldText parses deb822 field text into a single value.
func unmarshalFieldText(value reflect.Value, text string) error {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			if !value.CanSet() {
				return fmt.Errorf("cannot set value of type %s", value.Type())
			}

			value.Set(reflect.New(value.Type().Elem()))
		}

		if unmarshaler, ok := value.Interface().(encoding.TextUnmarshaler); ok {
			return unmarshaler.UnmarshalText([]byte(text))
		}

		return unmarshalFieldText(value.Elem(), text)
	}

	if value.CanAddr() {
		if unmarshaler, ok := value.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return unmarshaler.UnmarshalText([]byte(text))
		}
	}

	if !value.CanSet() {
		return fmt.Errorf("cannot set value of type %s", value.Type())
	}

	switch value.Kind() {
	case reflect.String:
		value.SetString(text)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(text, 10, value.Type().Bits())
		if err != nil {
			return err
		}

		value.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		parsed, err := strconv.ParseUint(text, 10, value.Type().Bits())
		if err != nil {
			return err
		}

		value.SetUint(parsed)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(text)
		if err != nil {
			return err
		}

		value.SetBool(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(text, value.Type().Bits())
		if err != nil {
			return err
		}

		value.SetFloat(parsed)
	default:
		return fmt.Errorf("unsupported type: %s", value.Type())
	}

	return nil
}

// isEmptyValue reports whether a value counts as empty for the omitempty
// option: the zero value of its type, or an empty string, slice, array or map.
func isEmptyValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return value.Len() == 0
	}

	return value.IsZero()
}

// asTextMarshaler returns the encoding.TextMarshaler of a value, preferring
// the value itself and falling back to its address when that is available.
func asTextMarshaler(value reflect.Value) (encoding.TextMarshaler, bool) {
	if !value.IsValid() {
		return nil, false
	}

	if value.Type().Implements(textMarshalerType) {
		marshaler, ok := value.Interface().(encoding.TextMarshaler)

		return marshaler, ok
	}

	if value.CanAddr() && reflect.PointerTo(value.Type()).Implements(textMarshalerType) {
		marshaler, ok := value.Addr().Interface().(encoding.TextMarshaler)

		return marshaler, ok
	}

	return nil, false
}
