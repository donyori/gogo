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

package testaux

import (
	"fmt"
	"strings"
)

// SliceToName formats s into a string for subtest names.
func SliceToName[T any](s []T, sep, format string, prependType bool) string {
	var typeStr string
	if prependType {
		typeStr = fmt.Sprintf("(%T)", s)
	}
	if s == nil {
		return typeStr + "<nil>"
	}
	var b strings.Builder
	b.WriteString(typeStr)
	b.WriteByte('[')
	for i := range s {
		if i > 0 {
			b.WriteString(sep)
		}
		_, _ = fmt.Fprintf(&b, format, s[i]) // ignore error as error is always nil
	}
	b.WriteByte(']')
	return b.String()
}
