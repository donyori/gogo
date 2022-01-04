// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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

package compare

import (
	"reflect"

	"github.com/donyori/gogo/errors"
)

// EqualFunc is a function to test whether a == b.
type EqualFunc func(a, b interface{}) bool

// Not returns a negative function to test whether !(a == b).
func (ef EqualFunc) Not() EqualFunc {
	return func(a, b interface{}) bool {
		return !ef(a, b)
	}
}

// Equal is a prefab EqualFunc performing as follows:
//
// If any input variable is nil (the nil interface{}),
// it returns true if and only if the other input variable is also nil or
// is the zero value of its type.
//
// Otherwise (two input variables are both non-nil interface{}),
// it returns true if and only if the two input variables satisfies
// the following three conditions:
//  1. they have identical dynamic types;
//  2. values of their type are comparable;
//  3. they have equal dynamic values.
//
// If any input variable is not comparable,
// it returns false rather than panicking.
//
// For more information about identical types,
// see <https://golang.org/ref/spec#Type_identity>.
//
// For more information about comparable types,
// see <https://golang.org/ref/spec#Comparison_operators>.
var Equal EqualFunc = equal

// equal is an implementation of function Equal.
func equal(a, b interface{}) bool {
	if a == nil {
		return b == nil || reflect.ValueOf(b).IsZero()
	}
	if b == nil {
		return reflect.ValueOf(a).IsZero()
	}
	// It's sufficient to just test whether a is comparable.
	return reflect.TypeOf(a).Comparable() && a == b
}

// BytesEqual is a prefab EqualFunc for []byte and string.
//
// It returns true if and only if the input slices have the same length
// and the same content.
// The nil interface{}, nil []byte, empty []byte, and empty string are
// considered equal for convenience.
//
// It panics if any non-nil input variable is neither []byte nor string.
//
// Note that []uint8 is equivalent to []byte as byte is an alias for uint8.
var BytesEqual EqualFunc = bytesEqual

// bytesEqualBytesString is used by function bytesEqual.
//
// It returns true if and only if b and s have the same length
// and the same content.
func bytesEqualBytesString(b []byte, s string) bool {
	if len(b) != len(s) {
		return false
	}
	for i := range b {
		if b[i] != s[i] {
			return false
		}
	}
	return true
}

// bytesEqual is an implementation of function bytesEqual.
func bytesEqual(a, b interface{}) bool {
	if a == nil {
		a = ""
	}
	if b == nil {
		b = ""
	}
	if sa, ok := a.(string); ok {
		if sb, ok := b.(string); ok {
			return sa == sb
		}
		if bb, ok := b.([]byte); ok {
			return bytesEqualBytesString(bb, sa)
		}
	} else if ba, ok := a.([]byte); ok {
		if sb, ok := b.(string); ok {
			return bytesEqualBytesString(ba, sb)
		}
		if bb, ok := b.([]byte); ok {
			if len(ba) != len(bb) {
				return false
			}
			for i := range ba {
				if ba[i] != bb[i] {
					return false
				}
			}
			return true
		}
	} else {
		panic(errors.AutoMsg("a is neither []byte nor string"))
	}
	panic(errors.AutoMsg("b is neither []byte nor string"))
}

// IntsEqual is a prefab EqualFunc for []int.
//
// It returns true if and only if the input slices have the same length
// and the same content.
// The nil interface{}, nil []int, and empty []int are considered equal
// for convenience.
//
// It panics if any non-nil input variable is not []int.
var IntsEqual EqualFunc = intsEqual

// intsEqual is an implementation of function IntsEqual.
func intsEqual(a, b interface{}) bool {
	var ia, ib []int
	var ok bool
	if a != nil {
		ia, ok = a.([]int)
		if !ok {
			panic(errors.AutoMsg("a is not []int"))
		}
	}
	if b != nil {
		ib, ok = b.([]int)
		if !ok {
			panic(errors.AutoMsg("b is not []int"))
		}
	}
	if len(ia) != len(ib) {
		return false
	}
	for i := range ia {
		// ia and ib will never be nil here.
		if ia[i] != ib[i] {
			return false
		}
	}
	return true
}

// Float64sEqual is a prefab EqualFunc for []float64.
//
// It returns true if and only if the input slices have the same length
// and the same content.
// The nil interface{}, nil []float64, and empty []float64 are considered equal
// for convenience.
//
// It panics if any non-nil input variable is not []float64.
var Float64sEqual EqualFunc = float64sEqual

// float64sEqual is an implementation of function Float64sEqual.
func float64sEqual(a, b interface{}) bool {
	var fa, fb []float64
	var ok bool
	if a != nil {
		fa, ok = a.([]float64)
		if !ok {
			panic(errors.AutoMsg("a is not []float64"))
		}
	}
	if b != nil {
		fb, ok = b.([]float64)
		if !ok {
			panic(errors.AutoMsg("b is not []float64"))
		}
	}
	if len(fa) != len(fb) {
		return false
	}
	for i := range fa {
		// fa and fb will never be nil here.
		if fa[i] != fb[i] {
			return false
		}
	}
	return true
}

// StringsEqual is a prefab EqualFunc for []string.
//
// It returns true if and only if the input slices have the same length
// and the same content.
// The nil interface{}, nil []string, and empty []string are considered equal
// for convenience.
//
// It panics if any non-nil input variable is not []string.
var StringsEqual EqualFunc = stringsEqual

// stringsEqual is an implementation of function StringsEqual.
func stringsEqual(a, b interface{}) bool {
	var sa, sb []string
	var ok bool
	if a != nil {
		sa, ok = a.([]string)
		if !ok {
			panic(errors.AutoMsg("a is not []string"))
		}
	}
	if b != nil {
		sb, ok = b.([]string)
		if !ok {
			panic(errors.AutoMsg("b is not []string"))
		}
	}
	if len(sa) != len(sb) {
		return false
	}
	for i := range sa {
		// sa and sb will never be nil here.
		if sa[i] != sb[i] {
			return false
		}
	}
	return true
}

// GeneralSliceEqual is a prefab EqualFunc for []interface{}.
//
// It returns true if and only if the input slices have the same length
// and the same content.
// The nil interface{}, nil []interface{}, and empty []interface{}
// are considered equal for convenience.
//
// It panics if any non-nil input variable is not []interface{}.
//
// It tests the elements of input slices through the prefab function Equal.
// See the document of Equal for details.
var GeneralSliceEqual EqualFunc = generalSliceEqual

// generalSliceEqual is an implementation of function GeneralSliceEqual.
func generalSliceEqual(a, b interface{}) bool {
	var ia, ib []interface{}
	var ok bool
	if a != nil {
		ia, ok = a.([]interface{})
		if !ok {
			panic(errors.AutoMsg("a is not []interface{}"))
		}
	}
	if b != nil {
		ib, ok = b.([]interface{})
		if !ok {
			panic(errors.AutoMsg("b is not []interface{}"))
		}
	}
	if len(ia) != len(ib) {
		return false
	}
	for i := range ia {
		// ia and ib will never be nil here.
		if !equal(ia[i], ib[i]) {
			return false
		}
	}
	return true
}

// SliceItemEqual is a prefab EqualFunc to test whether the items of
// a slice (i.e., []Type) or a string are correspondingly equal,
// where string is treated as []byte in this function.
//
// For better performance, if the slice type is []byte, []uint8, string, []int,
// []float64, []string, or []interface{},
// use the corresponding function BytesEqual, IntsEqual, Float64sEqual,
// StringsEqual, or GeneralSliceEqual instead.
//
// It returns true if and only if the input slices have the same length
// and the same content.
// It doesn't matter whether the two input slice types are identical.
// The nil interface{}, nil slice, and empty slice are considered equal
// for convenience.
//
// It panics if any non-nil input variable is neither slice nor string.
//
// It tests the elements of input variables through the prefab function Equal.
// See the document of Equal for details.
var SliceItemEqual EqualFunc = sliceItemEqual

// sliceItemEqual is an implementation of function SliceItemEqual.
func sliceItemEqual(a, b interface{}) bool {
	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
	var na, nb int
	switch va.Kind() {
	case reflect.Slice, reflect.String:
		na = va.Len()
	case reflect.Invalid:
	default:
		panic(errors.AutoMsg("a is neither slice nor string"))
	}
	switch vb.Kind() {
	case reflect.Slice, reflect.String:
		nb = vb.Len()
	case reflect.Invalid:
	default:
		panic(errors.AutoMsg("b is neither slice nor string"))
	}

	if na != nb {
		return false
	}
	for i := 0; i < na; i++ {
		if !equal(va.Index(i).Interface(), vb.Index(i).Interface()) {
			return false
		}
	}
	return true
}
