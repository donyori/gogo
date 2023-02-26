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
	"strings"
)

// FormatSlice formats s into a string.
//
// sep is the separator between every two items in s.
//
// format is the layout passed to fmt.Fprintf to print one item in s.
// It requires exactly one argument (e.g., "%v") for the item.
// If format is empty, the items in s are not printed.
//
// prependType indicates whether to add the slice type before the content.
// prependLen indicates whether to add the length of s before the content.
// If prependType is true and prependLen is false,
// the result begins with the type name
// wrapped in parentheses (e.g., "([]int)").
// If prependType is false and prependLen is true,
// the result begins with the length of s wrapped in parentheses (e.g., "(3)").
// If both prependType and prependLen are true,
// the result begins with the type name and the length of s,
// separated by a comma (','), and wrapped in parentheses (e.g., "([]int,3)").
//
// The result is as follows:
//   - <type-and-length> "<nil>", if s is nil, e.g., "([]int,0)<nil>"
//   - <type-and-length> "[]", if s is not nil but empty, e.g, "([]int,0)[]"
//   - <type-and-length> "[" <item> "]", if s has only one item, e.g., "([]int,1)[1]"
//   - <type-and-length> "[" <item-1> <sep> <item-2> <sep> ... <sep> <item-n> "]",
//     otherwise, e.g., "([]int,3)[1,2,3]"
//
// where the <type-and-length> is as follows:
//   - "(" <type> "," <length> ")", if both prependType and prependLen are true, e.g., "([]int,3)"
//   - "(" <type> ")", if prependType is true and prependLen is false, e.g, "([]int)"
//   - "(" <length> ")", if prependType is false and prependLen is true, e.g, "(3)"
//   - "", if both prependType and prependLen are false
func FormatSlice[T any](s []T, sep, format string, prependType, prependLen bool) string {
	var prefix string
	if prependType {
		if prependLen {
			prefix = fmt.Sprintf("(%T,%d)", s, len(s))
		} else {
			prefix = fmt.Sprintf("(%T)", s)
		}
	} else if prependLen {
		prefix = fmt.Sprintf("(%d)", len(s))
	}
	if s == nil {
		return prefix + "<nil>"
	}

	var b strings.Builder
	b.WriteString(prefix)
	b.WriteByte('[')
	for i := range s {
		if i > 0 {
			b.WriteString(sep)
		}
		if format != "" {
			_, _ = fmt.Fprintf(&b, format, s[i]) // ignore error as error is always nil
		}
	}
	b.WriteByte(']')
	return b.String()
}
