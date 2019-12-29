// gogo. A Golang toolbox.
// Copyright (C) 2019 Yuan Gao
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

// A prefab EqualFunc for comparable variables.
func Equal(a, b interface{}) bool {
	return a == b
}
