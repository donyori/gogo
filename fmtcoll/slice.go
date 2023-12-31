// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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
	"strings"

	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/errors"
)

// FormatSliceToString formats the slice s into a string
// with the specified format options.
//
// It returns the result string and any error encountered.
//
// If format is nil, it uses default format options
// as returned by NewDefaultSequenceFormat instead.
func FormatSliceToString[S constraints.Slice[Item], Item any](
	s S,
	format *SequenceFormat[Item],
) (result string, err error) {
	result, err = formatSequenceToString(
		format,
		reflect.TypeOf(s).String(),
		s == nil,
		len(s),
		func(handler func(x Item) (cont bool)) {
			for i := range s {
				if !handler(s[i]) {
					return
				}
			}
		},
	)
	return result, errors.AutoWrap(err)
}

// MustFormatSliceToString is like FormatSliceToString
// but panics when encountering an error.
func MustFormatSliceToString[S constraints.Slice[Item], Item any](
	s S,
	format *SequenceFormat[Item],
) string {
	result, err := FormatSliceToString(s, format)
	if err != nil {
		panic(errors.AutoWrap(err))
	}
	return result
}

// FormatSequenceToString formats the sequence s into a string
// with the specified format options.
//
// It returns the result string and any error encountered.
//
// If format is nil, it uses default format options
// as returned by NewDefaultSequenceFormat instead.
func FormatSequenceToString[Item any](
	s sequence.Sequence[Item],
	format *SequenceFormat[Item],
) (result string, err error) {
	var size int
	var rangeFn func(handler func(x Item) (cont bool))
	if s != nil {
		size, rangeFn = s.Len(), s.Range
	}
	result, err = formatSequenceToString(
		format,
		reflect.TypeOf(&s).Elem().String(), // reflect.TypeOf(s) returns nil if s is nil, so use &s here
		s == nil,
		size,
		rangeFn,
	)
	return result, errors.AutoWrap(err)
}

// MustFormatSequenceToString is like FormatSequenceToString
// but panics when encountering an error.
func MustFormatSequenceToString[Item any](
	s sequence.Sequence[Item],
	format *SequenceFormat[Item],
) string {
	result, err := FormatSequenceToString(s, format)
	if err != nil {
		panic(errors.AutoWrap(err))
	}
	return result
}

// formatSequenceToString is the main body to format a Go slice or
// a github.com/donyori/gogo/container/sequence.Sequence.
//
// Caller should guarantee that size is 0 if isNil is true,
// and rangeFn is not nil if size is greater than 0.
func formatSequenceToString[Item any](
	format *SequenceFormat[Item],
	typeStr string,
	isNil bool,
	size int,
	rangeFn func(handler func(x Item) (cont bool)),
) (result string, err error) {
	if format == nil {
		format = NewDefaultSequenceFormat[Item]()
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
	b.WriteByte('[')
	if size > 0 {
		if format.FormatItemFn != nil {
			b.WriteString(format.Prefix)
			var notFirst bool
			rangeFn(func(x Item) (cont bool) {
				if notFirst {
					b.WriteString(format.Separator)
				} else {
					notFirst = true
				}
				err = format.FormatItemFn(&b, x)
				return err == nil
			})
			if err != nil {
				return "", errors.AutoWrap(err)
			}
			b.WriteString(format.Suffix)
		} else {
			b.WriteString("...")
		}
	}
	b.WriteByte(']')
	return b.String(), nil
}
