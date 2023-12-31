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

import (
	"github.com/donyori/gogo/container"
	"github.com/donyori/gogo/container/sequence"
)

// DynamicArraySpecific is an interface that groups
// the dynamic-array-specific methods.
type DynamicArraySpecific[Item any] interface {
	container.Filter[Item]

	// Cap returns the current capacity of the dynamic array.
	Cap() int

	// Push adds x to the back of the dynamic array.
	Push(x Item)

	// Pop removes and returns the last item.
	//
	// It panics if the dynamic array is nil or empty.
	Pop() Item

	// Append adds s to the back of the dynamic array.
	//
	// s shouldn't be modified during calling this method,
	// otherwise, unknown error may occur.
	Append(s sequence.Sequence[Item])

	// Truncate removes the item with index i and all subsequent items.
	//
	// It does nothing if i is out of range.
	Truncate(i int)

	// Insert adds x as the item with index i.
	//
	// It panics if i is out of range, i.e., i < 0 or i > Len().
	Insert(i int, x Item)

	// Remove removes and returns the item with index i.
	//
	// It panics if i is out of range, i.e., i < 0 or i >= Len().
	Remove(i int) Item

	// RemoveWithoutOrder removes and returns the item with index i,
	// without preserving order.
	//
	// It panics if i is out of range, i.e., i < 0 or i >= Len().
	RemoveWithoutOrder(i int) Item

	// InsertSequence inserts s to the front of the item with index i.
	//
	// It panics if i is out of range, i.e., i < 0 or i > Len().
	//
	// s shouldn't be modified during calling this method,
	// otherwise, unknown error may occur.
	InsertSequence(i int, s sequence.Sequence[Item])

	// Cut removes items from argument begin (inclusive) to
	// argument end (exclusive) of the dynamic array.
	//
	// It panics if begin or end is out of range, or begin > end.
	Cut(begin, end int)

	// CutWithoutOrder removes items from argument begin (inclusive) to
	// argument end (exclusive) of the dynamic array, without preserving order.
	//
	// It panics if begin or end is out of range, or begin > end.
	CutWithoutOrder(begin, end int)

	// Extend adds n zero-value items to the back of the dynamic array.
	//
	// It panics if n < 0.
	Extend(n int)

	// Expand inserts n zero-value items to the front of the item with index i.
	//
	// It panics if i is out of range, i.e., i < 0 or i > Len(), or n < 0.
	Expand(i, n int)

	// Reserve requests that the capacity of the dynamic array
	// is at least the specified capacity.
	//
	// It does nothing if capacity <= Cap().
	Reserve(capacity int)

	// Shrink reduces the dynamic array to fit, i.e.,
	// requests Cap() equals to Len().
	//
	// Note that it isn't equivalent to operations on Go slice
	// like s[:len(s):len(s)],
	// because it allocates a new array and copies the content
	// if Cap() > Len().
	Shrink()

	// Clear removes all items in the dynamic array and
	// asks to release the memory.
	Clear()
}

// DynamicArray is an interface representing
// a dynamic-length direct-access sequence.
type DynamicArray[Item any] interface {
	Array[Item]
	DynamicArraySpecific[Item]
}

// OrderedDynamicArray is an interface representing a dynamic-length
// direct-access sequence that can be sorted by integer index.
//
// It conforms to interface sort.Interface.
type OrderedDynamicArray[Item any] interface {
	OrderedArray[Item]
	DynamicArraySpecific[Item]
}
