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

// Dynamic array.
type DynamicArray interface {
	Array

	// Return the capacity of the dynamic array.
	// It returns 0 if the dynamic array is nil.
	Cap() int

	// Add x to the back of the dynamic array.
	Push(x interface{})

	// Remove and return the last item of the dynamic array.
	// It panics if the dynamic array is nil or empty.
	Pop() interface{}

	// Append s to the back of the dynamic array.
	Append(s Sequence)

	// Remove the i-th and all subsequent items of the dynamic array.
	// It does nothing if i is out of range.
	Truncate(i int)

	// Insert x as the i-th item of the dynamic array.
	// It panics if i is out of range, i.e., i < 0 or i > Len().
	Insert(i int, x interface{})

	// Remove and return the i-th item of the dynamic array.
	// It panics if i is out of range, i.e., i < 0 or i >= Len().
	Remove(i int) interface{}

	// Remove and return the i-th item of the dynamic array, without preserving order.
	// It panics if i is out of range, i.e., i < 0 or i >= Len().
	RemoveWithoutOrder(i int) interface{}

	// Insert s to the front of the i-th item of the dynamic array.
	// It panics if i is out of range, i.e., i < 0 or i > Len().
	InsertSequence(i int, s Sequence)

	// Remove items from argument begin (inclusive) to
	// argument end (exclusive) of the dynamic array.
	// It panics if begin or end is out of range, or begin > end.
	Cut(begin, end int)

	// Remove items from argument begin (inclusive) to
	// argument end (exclusive) of the dynamic array, without preserving order.
	// It panics if begin or end is out of range, or begin > end.
	CutWithoutOrder(begin, end int)

	// Append n zero-value items to the back of the dynamic array.
	// It panics if n < 0.
	Extend(n int)

	// Insert n zero-value items to the front of the i-th item of the dynamic array.
	// It panics if i is out of range, i.e., i < 0 or i > Len(), or n < 0.
	Expand(i, n int)

	// Request the capacity of the dynamic array is at least cap.
	// It does nothing if capacity <= Cap().
	Reserve(capacity int)

	// Shrink the dynamic array to fit, i.e., request Cap() equals to Len().
	// Note that it isn't equivalent to operations on slice as s[:len(s):len(s)],
	// because it will allocate a new array and copy the content if Cap() > Len().
	Shrink()

	// Remove all items of the dynamic array and release the memory.
	Clear()

	// Filter items of the dynamic array (in place).
	Filter(filter func(x interface{}) (keep bool))
}
