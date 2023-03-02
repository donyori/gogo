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

	"github.com/donyori/gogo/errors"
)

// FormatSliceToString formats the slice s into a string
// with the specified format options.
//
// It returns the result string and any error encountered.
//
// If format is nil, it uses default format options
// as returned by NewDefaultSequenceFormat instead.
func FormatSliceToString[Item any](s []Item, format *SequenceFormat[Item]) (
	result string, err error) {
	if format == nil {
		format = NewDefaultSequenceFormat[Item]()
	}
	var prefix string
	if format.PrependType {
		if format.PrependSize {
			prefix = fmt.Sprintf("(%T,%d)", s, len(s))
		} else {
			prefix = fmt.Sprintf("(%T)", s)
		}
	} else if format.PrependSize {
		prefix = fmt.Sprintf("(%d)", len(s))
	}
	if s == nil {
		return prefix + "<nil>", nil
	}

	var b strings.Builder
	b.WriteString(prefix)
	b.WriteByte('[')
	if len(s) > 0 {
		if format.FormatItemFn != nil {
			b.WriteString(format.Prefix)
			for i := range s {
				if i > 0 {
					b.WriteString(format.Separator)
				}
				err = format.FormatItemFn(&b, s[i])
				if err != nil {
					return "", errors.AutoWrap(err)
				}
			}
			b.WriteString(format.Suffix)
		} else {
			b.WriteString("...")
		}
	}
	b.WriteByte(']')
	return b.String(), nil
}

// MustFormatSliceToString is like FormatSliceToString
// but panics when encountering an error.
func MustFormatSliceToString[Item any](s []Item, format *SequenceFormat[Item]) string {
	result, err := FormatSliceToString(s, format)
	if err != nil {
		panic(errors.AutoWrap(err))
	}
	return result
}
