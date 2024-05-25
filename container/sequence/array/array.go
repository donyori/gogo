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

package array

import "github.com/donyori/gogo/container/sequence"

// Array is an interface representing a direct-access sequence.
type Array[Item any] interface {
	sequence.Sequence[Item]

	// Get returns the item with index i.
	//
	// It panics if i is out of range.
	Get(i int) Item

	// Set sets the item with index i to x.
	//
	// It panics if i is out of range.
	Set(i int, x Item)

	// Swap exchanges the items with indexes i and j.
	//
	// It panics if i or j is out of range.
	Swap(i, j int)

	// Slice returns a slice from argument begin (inclusive) to
	// argument end (exclusive) of the array, as an Array.
	//
	// It panics if begin or end is out of range, or begin > end.
	Slice(begin, end int) Array[Item]
}

// OrderedArray is an interface representing a direct-access sequence
// that can be sorted by integer index.
//
// It conforms to interface sort.Interface.
type OrderedArray[Item any] interface {
	Array[Item]

	// Less reports whether the item with index i must sort before
	// the item with index j.
	//
	// Less must describe a strict weak ordering.
	// See <https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings>
	// for details.
	//
	// Note that floating-point comparison
	// (the < operator on float32 or float64 values)
	// is not a strict weak ordering
	// when not-a-number (NaN) values are involved.
	//
	// It panics if i or j is out of range.
	Less(i, j int) bool
}
