// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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

package sequence

// Array, i.e., fixed-length direct-access sequence.
type Array interface {
	Sequence

	// Return the i-th item of the array.
	// It panics if i is out of range.
	Get(i int) interface{}

	// Set the i-th item to x.
	// It panics if i is out of range.
	Set(i int, x interface{})

	// Swap the i-th and j-th items.
	// It panics if i or j is out of range.
	Swap(i, j int)

	// Return a slice from argument begin (inclusive) to
	// argument end (exclusive) of the array, as an Array.
	// It panics if begin or end is out of range, or begin > end.
	Slice(begin, end int) Array
}
