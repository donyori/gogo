// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2025  Yuan Gao
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
//
// It returns nil if this EqualFunc is nil.
func (ef EqualFunc[T]) Not() EqualFunc[T] {
	if ef == nil {
		return nil
	}
	return func(a, b T) bool {
		return !ef(a, b)
	}
}

// Reflexive returns an EqualFunc that guarantees
// the reflexivity of the equality relation.
// The returned function considers that
// all elements that are not equal to themselves
// (such as not-a-number (NaN) values of floating-point numbers) are equal.
//
// It returns nil if this EqualFunc is nil.
func (ef EqualFunc[T]) Reflexive() EqualFunc[T] {
	if ef == nil {
		return nil
	}
	return func(a, b T) bool {
		return ef(a, b) || !ef(a, a) && !ef(b, b)
	}
}

// Equal is a generic function to test whether a == b.
//
// The client can instantiate it to get an EqualFunc.
//
// For floating-point numbers, to consider NaN values equal to each other,
// use function FloatEqual.
func Equal[T comparable](a, b T) bool {
	return a == b
}

// ReflexiveEqual is a generic function to test whether a equals b.
// In particular, it considers that
// all elements that are not equal to themselves
// (such as not-a-number (NaN) values of floating-point numbers) are equal.
// Therefore, this equality relation is reflexive.
//
// The client can instantiate it to get an EqualFunc.
//
// To just test whether a == b, use function Equal.
func ReflexiveEqual[T comparable](a, b T) bool {
	return a == b || a != a && b != b
}

// FloatEqual is a generic function that returns true
// if a == b or both a and b are NaN.
//
// The client can instantiate it to get an EqualFunc.
//
// To just test whether a == b, use function Equal.
func FloatEqual[T constraints.Float](a, b T) bool {
	return ReflexiveEqual(a, b) // "x != x" means that x is a NaN
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

// SliceEqual is a generic function to test whether
// the specified slices have the same length and the items
// with the same index are equal.
// In particular, a nil slice and a non-nil empty slice are considered unequal.
//
// It uses the not equal operator (!=) to test the equality of the slice items.
//
// The client can instantiate it to get an EqualFunc.
func SliceEqual[S constraints.Slice[T], T comparable](a, b S) bool {
	n := len(a)
	if n != len(b) {
		return false
	} else if n == 0 {
		return (a == nil) == (b == nil)
	}
	for i := range n {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// FloatSliceEqual is a generic function to test whether
// the specified slices have the same length and the items
// with the same index are equal.
// In particular, a nil slice and a non-nil empty slice are considered unequal.
//
// Two items (floating-point numbers) x and y are considered equal
// if x == y or both x and y are NaN.
//
// The client can instantiate it to get an EqualFunc.
func FloatSliceEqual[S constraints.Slice[T], T constraints.Float](a, b S) bool {
	n := len(a)
	if n != len(b) {
		return false
	} else if n == 0 {
		return (a == nil) == (b == nil)
	}
	for i := range n {
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
// In particular, a nil slice and a non-nil empty slice are considered unequal.
//
// It uses the function AnyEqual to test the equality of the slice items.
//
// The client can instantiate it to get an EqualFunc.
func AnySliceEqual[S constraints.Slice[T], T any](a, b S) bool {
	n := len(a)
	if n != len(b) {
		return false
	} else if n == 0 {
		return (a == nil) == (b == nil)
	}
	for i := range n {
		if !AnyEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

// EqualToSliceEqual returns a function to test whether two slices of type S
// (whose underlying type is []T), a and b, have the same length and
// whether their items with the same index are equal.
//
// It uses ef to test the equality of the slice items.
// If ef is nil, it uses AnyEqual instead.
//
// nilEqualsEmpty indicates whether to consider
// a nil slice equal to a non-nil empty slice.
func EqualToSliceEqual[S constraints.Slice[T], T any](
	ef EqualFunc[T],
	nilEqualsEmpty bool,
) EqualFunc[S] {
	if ef == nil {
		ef = func(a, b T) bool {
			return AnyEqual(a, b)
		}
	}
	return func(a, b S) bool {
		n := len(a)
		if n != len(b) {
			return false
		} else if n == 0 {
			return nilEqualsEmpty || (a == nil) == (b == nil)
		}
		for i := range n {
			if !ef(a[i], b[i]) {
				return false
			}
		}
		return true
	}
}

// SliceEqualWithoutOrder is a generic function to test whether
// the specified slices have the same length and items.
// It compares the items of the slice regardless of their order.
// For example, the following slices are equal to each other for this function:
//
//	[]int{0, 0, 1, 2}, []int{0, 0, 2, 1}, []int{0, 1, 0, 2}, []int{0, 1, 2, 0},
//	[]int{0, 2, 0, 1}, []int{0, 2, 1, 0}, []int{1, 0, 0, 2}, []int{1, 0, 2, 0},
//	...
//
// because they all have two "0", one "1", and one "2".
// In particular, a nil slice and a non-nil empty slice are considered unequal.
//
// It is useful when slices are treated as sets or multisets
// rather than sequences.
//
// The client can instantiate it to get an EqualFunc.
func SliceEqualWithoutOrder[S constraints.Slice[T], T comparable](a, b S) bool {
	n := len(a)
	if n != len(b) {
		return false
	} else if n == 0 {
		return (a == nil) == (b == nil)
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

// FloatSliceEqualWithoutOrder is like SliceEqualWithoutOrder,
// but it considers two items (floating-point numbers) x and y equal
// if x == y or both x and y are NaN.
//
// The client can instantiate it to get an EqualFunc.
func FloatSliceEqualWithoutOrder[S constraints.Slice[T], T constraints.Float](
	a S,
	b S,
) bool {
	n := len(a)
	if n != len(b) {
		return false
	} else if n == 0 {
		return (a == nil) == (b == nil)
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

// MapEqual is a generic function to test whether
// the specified maps have the same key-value pairs.
// In particular, a nil map and a non-nil empty map are considered unequal.
//
// It uses the not equal operator (!=) to test the equality
// of the map keys and values.
//
// The client can instantiate it to get an EqualFunc.
func MapEqual[M constraints.Map[K, V], K, V comparable](a, b M) bool {
	n := len(a)
	if n != len(b) {
		return false
	} else if n == 0 {
		return (a == nil) == (b == nil)
	}
	for k, v1 := range a {
		v2, ok := b[k]
		if !ok || v1 != v2 {
			return false
		}
	}
	return true
}

// FloatValueMapEqual is a generic function to test whether
// the specified maps have the same key-value pairs.
// In particular, a nil map and a non-nil empty map are considered unequal.
//
// Two map keys k1 and k2 are considered equal if k1 == k2.
// Two map values v1 and v2 are considered equal
// if v1 == v2 or both v1 and v2 are NaN.
//
// The client can instantiate it to get an EqualFunc.
func FloatValueMapEqual[M constraints.Map[K, V], K comparable, V constraints.Float](
	a M,
	b M,
) bool {
	n := len(a)
	if n != len(b) {
		return false
	} else if n == 0 {
		return (a == nil) == (b == nil)
	}
	for k, v1 := range a {
		v2, ok := b[k]
		// "x == x" means that x is not a NaN.
		if !ok || v1 != v2 && (v1 == v1 || v2 == v2) {
			return false
		}
	}
	return true
}

// AnyValueMapEqual is a generic function to test whether
// the specified maps have the same key-value pairs.
// In particular, a nil map and a non-nil empty map are considered unequal.
//
// It uses the equal operator (==) to test the equality of the map keys
// and the function AnyEqual to test the equality of the map values.
//
// The client can instantiate it to get an EqualFunc.
func AnyValueMapEqual[M constraints.Map[K, V], K comparable, V any](
	a M,
	b M,
) bool {
	n := len(a)
	if n != len(b) {
		return false
	} else if n == 0 {
		return (a == nil) == (b == nil)
	}
	for k, v1 := range a {
		v2, ok := b[k]
		if !ok || !AnyEqual(v1, v2) {
			return false
		}
	}
	return true
}

// ValueEqualToMapEqual returns a function to test whether two maps of type M
// (whose underlying type is map[K]V), a and b, have the same key-value pairs.
//
// It uses the equal operator (==) to test the equality of the map keys
// and ef to test the equality of the map values.
// If ef is nil, it uses AnyEqual instead.
//
// nilEqualsEmpty indicates whether to consider
// a nil map equal to a non-nil empty map.
func ValueEqualToMapEqual[M constraints.Map[K, V], K comparable, V any](
	ef EqualFunc[V],
	nilEqualsEmpty bool,
) EqualFunc[M] {
	if ef == nil {
		ef = func(a, b V) bool {
			return AnyEqual(a, b)
		}
	}
	return func(a, b M) bool {
		n := len(a)
		if n != len(b) {
			return false
		} else if n == 0 {
			return nilEqualsEmpty || (a == nil) == (b == nil)
		}
		for k, v1 := range a {
			v2, ok := b[k]
			if !ok || !ef(v1, v2) {
				return false
			}
		}
		return true
	}
}
