// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2023  Yuan Gao
//
// This file is part of gogo.
//
// gogo is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package fmtcoll

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/container/mapping"
	"github.com/donyori/gogo/errors"
)

// FormatMapToString formats the map m into a string
// with the specified format options.
//
// It returns the result string and any error encountered.
//
// If format is nil, it uses default format options
// as returned by NewDefaultMapFormat instead.
func FormatMapToString[M constraints.Map[Key, Value], Key comparable, Value any](
	m M, format *MapFormat[Key, Value]) (result string, err error) {
	result, err = formatMapToString(
		format,
		reflect.TypeOf(m).String(),
		m == nil,
		len(m),
		func(handler func(x mapping.Entry[Key, Value]) (cont bool)) {
			for k, v := range m {
				if !handler(mapping.Entry[Key, Value]{Key: k, Value: v}) {
					return
				}
			}
		},
	)
	return result, errors.AutoWrap(err)
}

// MustFormatMapToString is like FormatMapToString
// but panics when encountering an error.
func MustFormatMapToString[M constraints.Map[Key, Value], Key comparable, Value any](
	m M, format *MapFormat[Key, Value]) string {
	result, err := FormatMapToString(m, format)
	if err != nil {
		panic(errors.AutoWrap(err))
	}
	return result
}

// FormatGogoMapToString formats the map m
// (of type github.com/donyori/gogo/container/mapping.Map)
// into a string with the specified format options.
//
// It returns the result string and any error encountered.
//
// If format is nil, it uses default format options
// as returned by NewDefaultMapFormat instead.
func FormatGogoMapToString[Key, Value any](
	m mapping.Map[Key, Value], format *MapFormat[Key, Value],
) (result string, err error) {
	var size int
	var rangeFn func(handler func(x mapping.Entry[Key, Value]) (cont bool))
	if m != nil {
		size, rangeFn = m.Len(), m.Range
	}
	result, err = formatMapToString(
		format,
		reflect.TypeOf(&m).Elem().String(), // reflect.TypeOf(m) returns nil if m is nil, so use &m here
		m == nil,
		size,
		rangeFn,
	)
	return result, errors.AutoWrap(err)
}

// MustFormatGogoMapToString is like FormatGogoMapToString
// but panics when encountering an error.
func MustFormatGogoMapToString[Key, Value any](
	m mapping.Map[Key, Value], format *MapFormat[Key, Value]) string {
	result, err := FormatGogoMapToString(m, format)
	if err != nil {
		panic(errors.AutoWrap(err))
	}
	return result
}

// formatMapToString is the main body to format a Go map or
// a github.com/donyori/gogo/container/mapping.Map.
//
// Caller should guarantee that size is 0 if isNil is true,
// and rangeFn is not nil if size is greater than 0.
func formatMapToString[Key, Value any](
	format *MapFormat[Key, Value],
	typeStr string,
	isNil bool,
	size int,
	rangeFn func(handler func(x mapping.Entry[Key, Value]) (cont bool)),
) (result string, err error) {
	if format == nil {
		format = NewDefaultMapFormat[Key, Value]()
	}
	var prefix string
	if format.PrependType {
		if format.PrependSize {
			prefix = fmt.Sprintf("(%s,%d)", typeStr, size)
		} else {
			prefix = fmt.Sprintf("(%s)", typeStr)
		}
	} else if format.PrependSize {
		prefix = fmt.Sprintf("(%d)", size)
	}
	if isNil {
		return prefix + "<nil>", nil
	}

	var b strings.Builder
	b.WriteString(prefix)
	b.WriteByte('{')
	if size > 0 {
		if format.FormatKeyFn != nil || format.FormatValueFn != nil {
			entries := make([]mapping.Entry[Key, Value], 0, size)
			rangeFn(func(x mapping.Entry[Key, Value]) (cont bool) {
				entries = append(entries, x)
				return true
			})
			if format.KeyValueLess != nil {
				sort.Slice(entries, func(i, j int) bool {
					return format.KeyValueLess(entries[i].Key, entries[j].Key,
						entries[i].Value, entries[j].Value)
				})
			}

			b.WriteString(format.Prefix)
			for i := range entries {
				if i > 0 {
					b.WriteString(format.Separator)
				}
				if format.FormatKeyFn != nil {
					err = format.FormatKeyFn(&b, entries[i].Key)
					if err != nil {
						return "", errors.AutoWrap(err)
					} else if format.FormatValueFn != nil {
						b.WriteByte(':')
					}
				}
				if format.FormatValueFn != nil {
					err = format.FormatValueFn(&b, entries[i].Value)
					if err != nil {
						return "", errors.AutoWrap(err)
					}
				}
			}
			b.WriteString(format.Suffix)
		} else {
			b.WriteString("...")
		}
	}
	b.WriteByte('}')
	return b.String(), nil
}
