// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2026  Yuan Gao
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

import "github.com/donyori/gogo/constraints"

// Func is a function that returns
//
//	-1 if a is less than b,
//	 0 if a equals b,
//	+1 if a is greater than b.
//
// To use Func in cases where a strict weak ordering is required,
// such as sorting, it must implement a strict weak ordering.
// For strict weak ordering,
// see <https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings>.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a strict weak ordering when not-a-number (NaN) values are involved.
type Func[T any] func(a, b T) int

// Reverse returns a reverse function that returns
//
//	-1 if a is greater than b,
//	 0 if a equals b,
//	+1 if a is less than b.
//
// Reverse returns nil if this Func is nil.
func (f Func[T]) Reverse() Func[T] {
	if f == nil {
		return nil
	}

	return func(a, b T) int {
		return -f(a, b)
	}
}

// ToEqual returns an EqualFunc to test whether a == b.
// The returned function reports true if and only if
//
//	compare(a, b) == 0
//
// ToEqual returns nil if this Func is nil.
func (f Func[T]) ToEqual() EqualFunc[T] {
	if f == nil {
		return nil
	}

	return func(a, b T) bool {
		return f(a, b) == 0
	}
}

// ToLess returns a LessFunc to test whether a < b.
// The returned function reports true if and only if
//
//	compare(a, b) < 0
//
// ToLess returns nil if this Func is nil.
func (f Func[T]) ToLess() LessFunc[T] {
	if f == nil {
		return nil
	}

	return func(a, b T) bool {
		return f(a, b) < 0
	}
}

// FuncToReverse is equivalent to
//
//	f.Reverse()
func FuncToReverse[T any](f Func[T]) Func[T] {
	return f.Reverse()
}

// FuncToEqual is equivalent to
//
//	f.ToEqual()
func FuncToEqual[T any](f Func[T]) EqualFunc[T] {
	return f.ToEqual()
}

// FuncToLess is equivalent to
//
//	f.ToLess()
func FuncToLess[T any](f Func[T]) LessFunc[T] {
	return f.ToLess()
}

// Ordered is a generic function that returns
//
//	-1 if a < b,
//	+1 if a > b,
//	 0 otherwise.
//
// The client can instantiate it to get a Func.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a strict weak ordering when not-a-number (NaN) values are involved.
// If a strict weak ordering is required (such as sorting),
// use the function Float for floating-point numbers.
func Ordered[T constraints.Ordered](a, b T) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}

	return 0
}

// Float is a generic function that returns
//
//	-1 if a < b, or a is a NaN and b is not a NaN,
//	+1 if a > b, or a is not a NaN and b is a NaN,
//	 0 otherwise (a == b or both a and b are NaN).
//
// It implements a strict weak ordering.
// See <https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings>
// for details.
//
// It treats NaN values as less than any others.
// A NaN is treated as equal to a NaN, and -0.0 is equal to 0.0.
//
// The client can instantiate it to get a Func.
func Float[T constraints.Float](a, b T) int {
	// "x != x" means that x is a NaN.
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	case a != a:
		if b == b {
			return -1
		}
	case b != b:
		return 1
	}

	return 0
}
