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

package compare

import "github.com/donyori/gogo/constraints"

// LessFunc is a function to test whether a < b.
//
// To use LessFunc in cases where a transitive ordering is required,
// such as sorting, it must implement a transitive ordering.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
type LessFunc[T any] func(a, b T) bool

// Not returns a negative function to test whether !(a < b).
func (lf LessFunc[T]) Not() LessFunc[T] {
	return func(a, b T) bool {
		return !lf(a, b)
	}
}

// Reverse returns a reverse function to test whether b < a.
func (lf LessFunc[T]) Reverse() LessFunc[T] {
	return func(a, b T) bool {
		return lf(b, a)
	}
}

// ToEqual returns an EqualFunc to test whether a == b.
// The return function reports true if and only if
//
//	!(less(a, b) || less(b, a))
func (lf LessFunc[T]) ToEqual() EqualFunc[T] {
	return func(a, b T) bool {
		return !(lf(a, b) || lf(b, a))
	}
}

// OrderedLess is a generic function to test whether a < b.
//
// The client can instantiate it to get a LessFunc.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
// If a transitive ordering is required (such as sorting),
// use the function FloatLess for floating-point numbers.
func OrderedLess[T constraints.Ordered](a, b T) bool {
	return a < b
}

// FloatLess is a generic function to test whether a < b
// for floating-point numbers.
//
// It implements a transitive ordering:
//   - if both FloatLess(a, b) and FloatLess(b, c) are true, then FloatLess(a, c) must be true as well.
//   - if both FloatLess(a, b) and FloatLess(b, c) are false, then FloatLess(a, c) must be false as well.
//
// It treats NaN values as less than any others.
//
// The client can instantiate it to get a LessFunc.
func FloatLess[T constraints.Float](a, b T) bool {
	// "x == x" means that x is not a NaN.
	return a < b || (a != a && b == b)
}
