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
	"sort"
	"strings"

	"github.com/donyori/gogo/errors"
)

// mapEntry represents a key-value pair of a map.
type mapEntry[Key, Value any] struct {
	key   Key
	value Value
}

// FormatMapToString formats the map m into a string
// with the specified format options.
//
// It returns the result string and any error encountered.
//
// If format is nil, it uses default format options
// as returned by NewDefaultMapFormat instead.
func FormatMapToString[Key comparable, Value any](
	m map[Key]Value, format *MapFormat[Key, Value],
) (result string, err error) {
	if format == nil {
		format = NewDefaultMapFormat[Key, Value]()
	}
	var prefix string
	if format.PrependType {
		if format.PrependSize {
			prefix = fmt.Sprintf("(%T,%d)", m, len(m))
		} else {
			prefix = fmt.Sprintf("(%T)", m)
		}
	} else if format.PrependSize {
		prefix = fmt.Sprintf("(%d)", len(m))
	}
	if m == nil {
		return prefix + "<nil>", nil
	}

	var b strings.Builder
	b.WriteString(prefix)
	b.WriteByte('{')
	if len(m) > 0 {
		if format.FormatKeyFn != nil || format.FormatValueFn != nil {
			entries := make([]mapEntry[Key, Value], 0, len(m))
			for k, v := range m {
				entries = append(entries, mapEntry[Key, Value]{key: k, value: v})
			}
			if format.KeyValueLess != nil {
				sort.Slice(entries, func(i, j int) bool {
					return format.KeyValueLess(entries[i].key, entries[j].key,
						entries[i].value, entries[j].value)
				})
			}

			b.WriteString(format.Prefix)
			for i := range entries {
				if i > 0 {
					b.WriteString(format.Separator)
				}
				if format.FormatKeyFn != nil {
					err = format.FormatKeyFn(&b, entries[i].key)
					if err != nil {
						return "", errors.AutoWrap(err)
					} else if format.FormatValueFn != nil {
						b.WriteByte(':')
					}
				}
				if format.FormatValueFn != nil {
					err = format.FormatValueFn(&b, entries[i].value)
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

// MustFormatMapToString is like FormatMapToString
// but panics when encountering an error.
func MustFormatMapToString[Key comparable, Value any](
	m map[Key]Value, format *MapFormat[Key, Value]) string {
	result, err := FormatMapToString(m, format)
	if err != nil {
		panic(errors.AutoWrap(err))
	}
	return result
}
