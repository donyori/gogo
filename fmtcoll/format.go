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
	"io"
)

// FormatFunc is a function to format x and write to w.
//
// It returns any error encountered.
type FormatFunc[T any] func(w io.Writer, x T) error

// FprintfToFormatFunc returns a FormatFunc that uses fmt.Fprintf to format x.
//
// The parameter format is the format specifier passed to fmt.Fprintf.
// It must have exactly one verb for x.
func FprintfToFormatFunc[T any](format string) FormatFunc[T] {
	return func(w io.Writer, x T) error {
		_, err := fmt.Fprintf(w, format, x)
		return err
	}
}

// CommonFormat contains the format options common to sequences and maps.
type CommonFormat struct {
	// Separator is the string inserted between every two items.
	//
	// Prefix is the string inserted before the first item.
	//
	// Suffix is the string inserted after the last item.
	//
	// Prefix and Suffix are used only if the collection is non-empty
	// and the collection content is not omitted.
	//
	// Separator is used only if the collection has at least two items
	// and the collection content is not omitted.
	Separator, Prefix, Suffix string

	// PrependType indicates whether to add
	// the collection type name before the content.
	//
	// PrependSize indicates whether to add
	// the collection size before the content.
	//
	// If PrependType is true and PrependSize is false,
	// the formatting result begins with the collection type name
	// wrapped in parentheses (e.g., "([]int)", "(map[string]int)").
	//
	// If PrependType is false and PrependSize is true,
	// the formatting result begins with the collection size
	// wrapped in parentheses (e.g., "(3)").
	//
	// If both PrependType and PrependSize are true,
	// the formatting result begins with the collection type name and size,
	// separated by a comma (','), and wrapped in parentheses
	// (e.g., "([]int,3)", "(map[string]int,3)").
	PrependType, PrependSize bool
}

// SequenceFormat contains the format options for sequences
// whose items are of type Item (e.g., []Item,
// github.com/donyori/gogo/container/sequence.Sequence[Item]).
//
// The formatting result is as follows:
//   - <TYPE-AND-SIZE> "<nil>", if the sequence is nil.
//   - <TYPE-AND-SIZE> "[" <CONTENT> "]", if the sequence is non-nil (can be empty).
//
// where <TYPE-AND-SIZE> is as follows:
//   - "(" <TYPE> "," <SIZE> ")", if both PrependType and PrependSize are true.
//   - "(" <TYPE> ")", if PrependType is true and PrependSize is false.
//   - "(" <SIZE> ")", if PrependType is false and PrependSize is true.
//   - "" (empty), if both PrependType and PrependSize are false.
//
// and <CONTENT> is as follows:
//   - <PREFIX> <ITEM-1> <SEPARATOR> <ITEM-2> <SEPARATOR> ... <SEPARATOR> <ITEM-N> <SUFFIX>,
//     if the sequence has at least two items and FormatItemFn is non-nil.
//   - <PREFIX> <ITEM> <SUFFIX>,
//     if the sequence has only one item and FormatItemFn is non-nil.
//   - "...", if the sequence is non-empty and FormatItemFn is nil.
//   - "" (empty), if the sequence is empty.
type SequenceFormat[Item any] struct {
	CommonFormat

	// FormatItemFn is a function to format the item x and write to w.
	// It returns any error encountered.
	//
	// If FormatItemFn is nil, the sequence content is omitted.
	// In this case, if the sequence is non-empty,
	// the content is printed as "...".
	FormatItemFn FormatFunc[Item]
}

// NewDefaultSequenceFormat creates a new SequenceFormat
// with the default options as follows:
//   - Separator: ","
//   - Prefix: ""
//   - Suffix: ""
//   - PrependType: true
//   - PrependSize: true
//   - FormatItemFn: FprintfToFormatFunc[Item]("%v")
func NewDefaultSequenceFormat[Item any]() *SequenceFormat[Item] {
	return &SequenceFormat[Item]{
		CommonFormat: CommonFormat{
			Separator:   ",",
			PrependType: true,
			PrependSize: true,
		},
		FormatItemFn: FprintfToFormatFunc[Item]("%v"),
	}
}

// MapFormat contains the format options for maps whose keys are of type Key
// and values are of type Value (e.g., map[Key]Value if Key is comparable,
// github.com/donyori/gogo/container/mapping.Map[Key, Value]).
//
// The formatting result is as follows:
//   - <TYPE-AND-SIZE> "<nil>", if the map is nil.
//   - <TYPE-AND-SIZE> "{" <CONTENT> "}", if the map is non-nil (can be empty).
//
// where <TYPE-AND-SIZE> is as follows:
//   - "(" <TYPE> "," <SIZE> ")", if both PrependType and PrependSize are true.
//   - "(" <TYPE> ")", if PrependType is true and PrependSize is false.
//   - "(" <SIZE> ")", if PrependType is false and PrependSize is true.
//   - "" (empty), if both PrependType and PrependSize are false.
//
// and <CONTENT> is as follows:
//   - <PREFIX> <ITEM-1> <SEPARATOR> <ITEM-2> <SEPARATOR> ... <SEPARATOR> <ITEM-N> <SUFFIX>,
//     if the map has at least two key-value pairs and either FormatKeyFn or FormatValueFn is non-nil.
//   - <PREFIX> <ITEM> <SUFFIX>,
//     if the map has only one key-value pair and either FormatKeyFn or FormatValueFn is non-nil.
//   - "...", if the map is non-empty and both FormatKeyFn and FormatValueFn are nil.
//   - "" (empty), if the map is empty.
//
// where <ITEM> (including <ITEM-1>, <ITEM-2>, ..., <ITEM-N>) is as follows:
//   - <KEY> ":" <VALUE>, if both FormatKeyFn and FormatValueFn are non-nil.
//   - <KEY>, if FormatKeyFn is non-nil and FormatValueFn is nil.
//   - <VALUE>, if FormatKeyFn is nil and FormatValueFn is non-nil.
type MapFormat[Key, Value any] struct {
	CommonFormat

	// FormatKeyFn is a function to format the key and write to w.
	// It returns any error encountered.
	//
	// If FormatKeyFn is nil, the key is omitted.
	// If both the key and value are omitted and the map is non-empty,
	// the content is printed as "...".
	FormatKeyFn FormatFunc[Key]

	// FormatValueFn is a function to format the value and write to w.
	// It returns any error encountered.
	//
	// If FormatValueFn is nil, the value is omitted.
	// If both the key and value are omitted and the map is non-empty,
	// the content is printed as "...".
	FormatValueFn FormatFunc[Value]

	// KeyValueLess is a function to report whether the key-value pair
	// (key1, value1) is less than (key2, value2).
	//
	// It is used to sort the key-value pairs in the map.
	//
	// It must describe a transitive ordering.
	// Note that floating-point comparison
	// (the < operator on float32 or float64 values)
	// is not a transitive ordering when not-a-number (NaN) values are involved.
	//
	// If KeyValueLess is nil, the key-value pairs may be in random order.
	KeyValueLess func(key1, key2 Key, value1, value2 Value) bool
}

// NewDefaultMapFormat creates a new MapFormat
// with the default options as follows:
//   - Separator: ","
//   - Prefix: ""
//   - Suffix: ""
//   - PrependType: true
//   - PrependSize: true
//   - FormatKeyFn: FprintfToFormatFunc[Key]("%v")
//   - FormatValueFn: FprintfToFormatFunc[Value]("%v")
//   - KeyValueLess: nil
func NewDefaultMapFormat[Key, Value any]() *MapFormat[Key, Value] {
	return &MapFormat[Key, Value]{
		CommonFormat: CommonFormat{
			Separator:   ",",
			PrependType: true,
			PrependSize: true,
		},
		FormatKeyFn:   FprintfToFormatFunc[Key]("%v"),
		FormatValueFn: FprintfToFormatFunc[Value]("%v"),
	}
}
