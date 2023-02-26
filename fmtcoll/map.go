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
)

// mapEntry represents a key-value pair of a map.
type mapEntry[K, V any] struct {
	key   K
	value V
}

// FormatMap formats m into a string.
//
// sep is the separator between every two key-value pairs in m.
//
// keyFormat and valueFormat are the layouts passed to fmt.Fprintf
// to print one key and one value in m, respectively.
// Both of them require exactly one argument (e.g., "%v") for the key or value.
// If keyFormat is empty, the keys are not printed.
// If valueFormat is empty, the values are not printed.
// If both keyFormat and valueFormat are non-empty,
// a colon (':') is inserted between the key and value.
//
// prependType indicates whether to add the slice type before the content.
// prependLen indicates whether to add the length of m before the content.
// If prependType is true and prependLen is false,
// the result begins with the type name
// wrapped in parentheses (e.g., "(map[string]int)").
// If prependType is false and prependLen is true,
// the result begins with the length of m wrapped in parentheses (e.g., "(3)").
// If both prependType and prependLen are true,
// the result begins with the type name and the length of m,
// separated by a comma (','), and wrapped in parentheses
// (e.g., "(map[string]int,3)").
//
// entryLess is a function to report whether the key-value pair 1
// is less than the key-value pair 2.
// It is used to sort the key-value pairs in m.
// It must describe a transitive ordering.
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
// If entryLess is nil, the key-value pairs may be unsorted.
//
// The result is as follows:
//   - <type-and-length> "<nil>", if m is nil, e.g., "(map[string]int,0)<nil>"
//   - <type-and-length> "{}", if m is not nil but empty, e.g, "(map[string]int,0){}"
//   - <type-and-length> "{" <item> "}", if m has only one key-value pair, e.g., `(map[string]int,1){"A":1}`
//   - <type-and-length> "{" <item-1> <sep> <item-2> <sep> ... <sep> <item-n> "}",
//     otherwise, e.g., `(map[string]int,3){"A":1,"B":2,"C":3}`
//
// where the <type-and-length> is as follows:
//   - "(" <type> "," <length> ")", if both prependType and prependLen are true, e.g., "(map[string]int,3)"
//   - "(" <type> ")", if prependType is true and prependLen is false, e.g, "(map[string]int)"
//   - "(" <length> ")", if prependType is false and prependLen is true, e.g., "(3)"
//   - "", if both prependType and prependLen are false
//
// and the <item> is as follows:
//   - <key> ":" <value>, if both keyFormat and valueFormat are non-empty, e.g., `"A":1`
//   - <key>, if keyFormat is non-empty and valueFormat is empty, e.g., `"A"`
//   - <value>, if keyFormat is empty and valueFormat is non-empty, e.g., "1"
//   - "", if both keyFormat and valueFormat are empty
func FormatMap[K comparable, V any](
	m map[K]V,
	sep, keyFormat, valueFormat string,
	prependType, prependLen bool,
	entryLess func(key1 K, value1 V, key2 K, value2 V) bool,
) string {
	var prefix string
	if prependType {
		if prependLen {
			prefix = fmt.Sprintf("(%T,%d)", m, len(m))
		} else {
			prefix = fmt.Sprintf("(%T)", m)
		}
	} else if prependLen {
		prefix = fmt.Sprintf("(%d)", len(m))
	}
	if m == nil {
		return prefix + "<nil>"
	}

	entries := make([]mapEntry[K, V], 0, len(m))
	for k, v := range m {
		entries = append(entries, mapEntry[K, V]{key: k, value: v})
	}
	if entryLess != nil {
		sort.Slice(entries, func(i, j int) bool {
			return entryLess(entries[i].key, entries[i].value,
				entries[j].key, entries[j].value)
		})
	}

	var b strings.Builder
	b.WriteString(prefix)
	b.WriteByte('{')
	for i := range entries {
		if i > 0 {
			b.WriteString(sep)
		}
		if keyFormat != "" {
			_, _ = fmt.Fprintf(&b, keyFormat, entries[i].key) // ignore error as error is always nil
			if valueFormat != "" {
				b.WriteByte(':')
			}
		}
		if valueFormat != "" {
			_, _ = fmt.Fprintf(&b, valueFormat, entries[i].value) // ignore error as error is always nil
		}
	}
	b.WriteByte('}')
	return b.String()
}
