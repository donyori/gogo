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

package compare

import (
	"reflect"

	"github.com/donyori/gogo/constraints"
)

// EqualFunc is a function to test whether a == b.
type EqualFunc[T any] func(a, b T) bool

// Not returns a negative function to test whether !(a == b).
func (ef EqualFunc[T]) Not() EqualFunc[T] {
	return func(a, b T) bool {
		return !ef(a, b)
	}
}

// ComparableEqual is a generic function to test whether a == b.
//
// The client can instantiate it to get an EqualFunc.
//
// For floating-point numbers, to consider NaN values equal to each other,
// use function FloatEqual.
func ComparableEqual[T comparable](a, b T) bool {
	return a == b
}

// FloatEqual is a generic function that returns true
// if a == b or both a and b are NaN.
//
// The client can instantiate it to get an EqualFunc.
//
// To just test whether a == b, use function ComparableEqual.
func FloatEqual[T constraints.Float](a, b T) bool {
	return a == b || a != a && b != b // "x != x" means that x is a NaN
}

// AnyEqual is a prefab EqualFunc performing as follows:
//
// If any input variable is nil (the nil any (i.e., nil interface{})),
// it returns true if and only if the other input variable is also nil or
// is the zero value of its type.
//
// Otherwise (two input variables are both non-nil any
// (i.e., non-nil interface{})),
// it returns true if and only if the two input variables satisfies
// the following three conditions:
//  1. they have identical dynamic types;
//  2. values of their type are comparable;
//  3. they have equal dynamic values.
//
// If any input variable is not comparable,
// it returns false rather than panicking.
//
// Note that for floating-point numbers, NaN values are not equal to each other.
// To consider NaN values equal to each other, use function FloatEqual.
//
// For more information about identical types,
// see <https://go.dev/ref/spec#Type_identity>.
//
// For more information about comparable types,
// see <https://go.dev/ref/spec#Comparison_operators>.
var AnyEqual EqualFunc[any] = anyEqual

// anyEqual is an implementation of function AnyEqual.
func anyEqual(a, b any) bool {
	if a == nil {
		return b == nil || reflect.ValueOf(b).IsZero()
	} else if b == nil {
		return reflect.ValueOf(a).IsZero()
	}
	// It's sufficient to just test whether a is comparable.
	return reflect.TypeOf(a).Comparable() && a == b
}

// ComparableSliceEqual is a generic function to test whether
// the specified slices have the same length and the items
// with the same index are equal.
//
// It uses the not equal operator (!=) to test the equality
// of the slice items.
//
// The client can instantiate it to get an EqualFunc.
func ComparableSliceEqual[T comparable](a, b []T) bool {
	n := len(a)
	if n != len(b) {
		return false
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// FloatSliceEqual is a generic function to test whether
// the specified slices have the same length and the items
// with the same index are equal.
//
// Two items (floating-point numbers) a and b are considered equal
// if a == b or both a and b are NaN.
//
// The client can instantiate it to get an EqualFunc.
func FloatSliceEqual[T constraints.Float](a, b []T) bool {
	n := len(a)
	if n != len(b) {
		return false
	}
	for i := 0; i < n; i++ {
		// "x == x" means that x is not a NaN.
		if a[i] != b[i] && (a[i] == a[i] || b[i] == b[i]) {
			return false
		}
	}
	return true
}

// AnySliceEqual is a generic function to test whether
// the specified slices have the same length and the items
// with the same index are equal.
//
// It uses the function AnyEqual to test the equality
// of the slice items.
//
// The client can instantiate it to get an EqualFunc.
func AnySliceEqual[T any](a, b []T) bool {
	n := len(a)
	if n != len(b) {
		return false
	}
	for i := 0; i < n; i++ {
		if !AnyEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

// ToSliceEqual returns a function to test whether two slices of type []T,
// a and b, have the same length and whether their items
// with the same index are equal.
//
// It uses ef to test the equality of the slice items.
// If ef is nil, it uses AnyEqual instead.
//
// nilEqualToEmpty indicates whether to consider
// nil slices equal to non-nil empty slices.
func ToSliceEqual[T any](ef EqualFunc[T], nilEqualToEmpty bool) EqualFunc[[]T] {
	if ef == nil {
		ef = func(a, b T) bool {
			return AnyEqual(a, b)
		}
	}
	return func(a, b []T) bool {
		n := len(a)
		if n != len(b) {
			return false
		} else if n == 0 && !nilEqualToEmpty {
			return (a == nil) == (b == nil)
		}
		for i := 0; i < n; i++ {
			if !ef(a[i], b[i]) {
				return false
			}
		}
		return true
	}
}

// ComparableSliceEqualWithoutOrder is a generic function to test whether
// the specified slices have the same length and items.
// It compares the items of the slice regardless of their order.
// For example, the following slices are equal to each other for this function:
//
//	[]int{0, 0, 1, 2}, []int{0, 0, 2, 1}, []int{0, 1, 0, 2}, []int{0, 1, 2, 0},
//	[]int{0, 2, 0, 1}, []int{0, 2, 1, 0}, []int{1, 0, 0, 2}, []int{1, 0, 2, 0},
//	...
//
// because they all have two "0", one "1", and one "2".
//
// It is useful when slices are treated as sets or multisets
// rather than sequences.
//
// The client can instantiate it to get an EqualFunc.
func ComparableSliceEqualWithoutOrder[T comparable](a, b []T) bool {
	n := len(a)
	if n != len(b) {
		return false
	} else if n == 0 {
		return true
	}
	counter := make(map[T]int, n)
	for _, x := range a {
		counter[x]++
	}
	for _, x := range b {
		c := counter[x] - 1
		if c < 0 {
			return false
		}
		counter[x] = c
	}
	for _, c := range counter {
		if c > 0 {
			return false
		}
	}
	return true
}

// FloatSliceEqualWithoutOrder is like ComparableSliceEqualWithoutOrder,
// but it considers two items (floating-point numbers) a and b equal
// if a == b or both a and b are NaN.
//
// The client can instantiate it to get an EqualFunc.
func FloatSliceEqualWithoutOrder[T constraints.Float](a, b []T) bool {
	n := len(a)
	if n != len(b) {
		return false
	} else if n == 0 {
		return true
	}
	counter, numNaN := make(map[T]int, n), 0
	// "x == x" means that x is not a NaN.
	for _, x := range a {
		if x == x {
			counter[x]++
		} else {
			numNaN++
		}
	}
	for _, x := range b {
		if x == x {
			c := counter[x] - 1
			if c < 0 {
				return false
			}
			counter[x] = c
		} else {
			if numNaN <= 0 {
				return false
			}
			numNaN--
		}
	}
	for _, c := range counter {
		if c > 0 {
			return false
		}
	}
	// numNaN must be 0 here,
	// because len(a) == len(b) and they have the same number of non-NaN items.
	return true
}
