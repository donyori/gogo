// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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

package function

import (
	"reflect"

	"github.com/donyori/gogo/errors"
)

// A function to test whether a == b.
type EqualFunc func(a, b interface{}) bool

// Generate EqualFunc via LessFunc.
// equal(a, b) = !(less(a, b) || less(b, a)).
func GenerateEqualViaLess(less LessFunc) EqualFunc {
	return func(a, b interface{}) bool {
		return !(less(a, b) || less(b, a))
	}
}

// Negative the function, i.e., to test whether !(a == b).
func (ef EqualFunc) Not() EqualFunc {
	return func(a, b interface{}) bool {
		return !ef(a, b)
	}
}

// A prefab EqualFunc for comparable variables
// (i.e., variables that can be operands of equality operators == and !=).
//
// For more information about comparable types,
// see <https://golang.org/ref/spec#Comparison_operators>.
func Equal(a, b interface{}) bool {
	return a == b
}

// A prefab EqualFunc for []int.
//
// It returns true iff a and b have the same length and the same content,
// or both a and b are nil.
// A nil slice and an empty slice are considered unequal.
func IntsEqual(a, b interface{}) bool {
	if a == nil || a.([]int) == nil {
		return b == nil || b.([]int) == nil
	} else if b == nil || b.([]int) == nil {
		return false
	}
	ia, ib := a.([]int), b.([]int)
	if len(ia) != len(ib) {
		return false
	}
	for i, k := 0, len(ia)-1; i <= k; i, k = i+1, k-1 {
		if ia[i] != ib[i] || ia[k] != ib[k] {
			return false
		}
	}
	return true
}

// A prefab EqualFunc for []float64.
//
// It returns true iff a and b have the same length and the same content,
// or both a and b are nil.
// A nil slice and an empty slice are considered unequal.
func Float64sEqual(a, b interface{}) bool {
	if a == nil || a.([]float64) == nil {
		return b == nil || b.([]float64) == nil
	} else if b == nil || b.([]float64) == nil {
		return false
	}
	fa, fb := a.([]float64), b.([]float64)
	if len(fa) != len(fb) {
		return false
	}
	for i, k := 0, len(fa)-1; i <= k; i, k = i+1, k-1 {
		if fa[i] != fb[i] || fa[k] != fb[k] {
			return false
		}
	}
	return true
}

// A prefab EqualFunc for []string.
//
// It returns true iff a and b have the same length and the same content,
// or both a and b are nil.
// A nil slice and an empty slice are considered unequal.
func StringsEqual(a, b interface{}) bool {
	if a == nil || a.([]string) == nil {
		return b == nil || b.([]string) == nil
	} else if b == nil || b.([]string) == nil {
		return false
	}
	sa, sb := a.([]string), b.([]string)
	if len(sa) != len(sb) {
		return false
	}
	for i, k := 0, len(sa)-1; i <= k; i, k = i+1, k-1 {
		if sa[i] != sb[i] || sa[k] != sb[k] {
			return false
		}
	}
	return true
}

// A prefab EqualFunc for []interface{}.
//
// It returns true iff a and b have the same length and the same content,
// or both a and b are nil.
// A nil slice and an empty slice are considered unequal.
//
// It tests the elements of a and b through the "not equal" operator (i.e., !=)
// rather than something like reflect.DeepEqual.
// It will panic if the element type is not comparable
// (i.e., cannot use "==" and "!=" on it).
// For more information about comparable types,
// see <https://golang.org/ref/spec#Comparison_operators>.
func GeneralSliceEqual(a, b interface{}) bool {
	if a == nil || a.([]interface{}) == nil {
		return b == nil || b.([]interface{}) == nil
	} else if b == nil || b.([]interface{}) == nil {
		return false
	}
	ia, ib := a.([]interface{}), b.([]interface{})
	if len(ia) != len(ib) {
		return false
	}
	for i, k := 0, len(ia)-1; i <= k; i, k = i+1, k-1 {
		if ia[i] != ib[i] || ia[k] != ib[k] {
			return false
		}
	}
	return true
}

// A prefab EqualFunc for slice (i.e., []Type).
//
// For better performance,
// if the slice type is []int, []float64, []string, or []interface{},
// use IntsEqual, Float64sEqual, StringsEqual, or GeneralSliceEqual instead.
//
// It returns true iff a and b have the same type, the same length,
// and the same content, or both a and b are nil.
// A nil slice and an empty slice are considered unequal.
//
// It tests the elements of a and b through the "not equal" operator (i.e., !=)
// rather than something like reflect.DeepEqual.
// It will panic if the element type is not comparable
// (i.e., cannot use "==" and "!=" on it).
// For more information about comparable types,
// see <https://golang.org/ref/spec#Comparison_operators>.
//
// It will panic if the type of a or b is not a slice.
// However, if the type of the elements of a is not the same as that of b,
// it will return false rather than panic.
func SliceEqual(a, b interface{}) bool {
	if a == nil {
		if b == nil {
			return true
		}
		vb := reflect.ValueOf(b)
		if vb.Kind() != reflect.Slice {
			panic(errors.AutoMsg("b is NOT a slice"))
		}
		return vb.IsNil()
	} else if b == nil {
		va := reflect.ValueOf(a)
		if va.Kind() != reflect.Slice {
			panic(errors.AutoMsg("a is NOT a slice"))
		}
		return va.IsNil()
	}
	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
	if va.Kind() != reflect.Slice {
		panic(errors.AutoMsg("a is NOT a slice"))
	}
	if vb.Kind() != reflect.Slice {
		panic(errors.AutoMsg("b is NOT a slice"))
	}
	if va.Type().Elem() != vb.Type().Elem() {
		return false
	}
	if va.IsNil() {
		return vb.IsNil()
	} else if vb.IsNil() {
		return false
	}
	if va.Len() != vb.Len() {
		return false
	}
	for i, k := 0, va.Len()-1; i <= k; i, k = i+1, k-1 {
		if va.Index(i).Interface() != vb.Index(i).Interface() ||
			va.Index(k).Interface() != vb.Index(k).Interface() {
			return false
		}
	}
	return true
}
