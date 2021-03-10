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

// LessFunc is a function to test whether a < b.
type LessFunc func(a, b interface{}) bool

// Not returns a negative function to test whether !(a < b).
func (lf LessFunc) Not() LessFunc {
	return func(a, b interface{}) bool {
		return !lf(a, b)
	}
}

// Reverse returns a reverse function to test whether b < a.
func (lf LessFunc) Reverse() LessFunc {
	return func(a, b interface{}) bool {
		return lf(b, a)
	}
}

// ToEqual returns an EqualFunc to test whether a == b.
// The return function reports true if and only if
//  !(less(a, b) || less(b, a))
func (lf LessFunc) ToEqual() EqualFunc {
	return func(a, b interface{}) bool {
		return !(lf(a, b) || lf(b, a))
	}
}

// IntLess is a prefab LessFunc for int.
var IntLess LessFunc = func(a, b interface{}) bool {
	return a.(int) < b.(int)
}

// Float64Less is a prefab LessFunc for float64.
var Float64Less LessFunc = func(a, b interface{}) bool {
	return a.(float64) < b.(float64)
}

// StringLess is a prefab LessFunc for string.
var StringLess LessFunc = func(a, b interface{}) bool {
	return a.(string) < b.(string)
}
