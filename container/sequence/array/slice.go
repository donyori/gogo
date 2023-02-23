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

package array

import "github.com/donyori/gogo/container/sequence"

// SliceDynamicArray is a dynamic array wrapped on Go slice.
// It implements the interface DynamicArray.
//
// The client can convert a Go slice to SliceDynamicArray by type conversion,
// e.g.:
//
//	sda := SliceDynamicArray[int]([]int{1, 2, 3})
//
// Or allocate a new SliceDynamicArray by the built-in function make, e.g.:
//
//	sda := make(SliceDynamicArray[int], 2, 10)
type SliceDynamicArray[Item any] []Item

// Len returns the number of items in the slice.
//
// It returns 0 if the slice is nil.
func (sda SliceDynamicArray[Item]) Len() int {
	return len(sda)
}

// Range accesses the items in the slice from first to last.
// Each item will be accessed once.
//
// Its parameter handler is a function to deal with the item x in the
// slice and report whether to continue to access the next item.
func (sda SliceDynamicArray[Item]) Range(handler func(x Item) (cont bool)) {
	for _, x := range sda {
		if !handler(x) {
			return
		}
	}
}

// Front returns the first item.
//
// It panics if the slice is nil or empty.
func (sda SliceDynamicArray[Item]) Front() Item {
	return sda[0]
}

// SetFront sets the first item to x.
//
// It panics if the slice is nil or empty.
func (sda SliceDynamicArray[Item]) SetFront(x Item) {
	sda[0] = x
}

// Back returns the last item.
//
// It panics if the slice is nil or empty.
func (sda SliceDynamicArray[Item]) Back() Item {
	return sda[len(sda)-1]
}

// SetBack sets the last item to x.
//
// It panics if the slice is nil or empty.
func (sda SliceDynamicArray[Item]) SetBack(x Item) {
	sda[len(sda)-1] = x
}

// Reverse turns items in the slice the other way round.
func (sda SliceDynamicArray[Item]) Reverse() {
	for i, j := 0, len(sda)-1; i < j; i, j = i+1, j-1 {
		sda[i], sda[j] = sda[j], sda[i]
	}
}

// Get returns the item with index i.
//
// It panics if i is out of range.
func (sda SliceDynamicArray[Item]) Get(i int) Item {
	return sda[i]
}

// Set sets the item with index i to x.
//
// It panics if i is out of range.
func (sda SliceDynamicArray[Item]) Set(i int, x Item) {
	sda[i] = x
}

// Swap exchanges the items with indexes i and j.
//
// It panics if i or j is out of range.
func (sda SliceDynamicArray[Item]) Swap(i, j int) {
	sda[i], sda[j] = sda[j], sda[i]
}

// Slice returns a slice from argument begin (inclusive) to
// argument end (exclusive) of the slice, as an Array.
//
// It panics if begin or end is out of range, or begin > end.
func (sda SliceDynamicArray[Item]) Slice(begin, end int) Array[Item] {
	return sda[begin:end:end]
}

// Filter refines items in the slice (in-place).
//
// Its parameter filter is a function to report
// whether to keep the item x.
func (sda *SliceDynamicArray[Item]) Filter(filter func(x Item) (keep bool)) {
	if sda == nil || len(*sda) == 0 {
		return
	}
	var n int
	for _, x := range *sda {
		if filter(x) {
			(*sda)[n], n = x, n+1
		}
	}
	if n == len(*sda) {
		return
	}
	var zero Item
	for i := n; i < len(*sda); i++ {
		(*sda)[i] = zero // avoid memory leak
	}
	*sda = (*sda)[:n]
}

// Cap returns the current capacity of the slice.
//
// It returns 0 if the slice is nil.
func (sda SliceDynamicArray[Item]) Cap() int {
	return cap(sda)
}

// Push adds x to the back of the slice.
func (sda *SliceDynamicArray[Item]) Push(x Item) {
	*sda = append(*sda, x)
}

// Pop removes and returns the last item.
//
// It panics if the slice is nil or empty.
func (sda *SliceDynamicArray[Item]) Pop() Item {
	back := len(*sda) - 1
	x := (*sda)[back]
	var zero Item
	(*sda)[back] = zero // avoid memory leak
	*sda = (*sda)[:back]
	return x
}

// Append adds s to the back of the slice.
//
// s shouldn't be modified during calling this method,
// otherwise, unknown error may occur.
func (sda *SliceDynamicArray[Item]) Append(s sequence.Sequence[Item]) {
	if s == nil {
		return
	}
	n := s.Len()
	if n == 0 {
		return
	}
	if t, ok := s.(SliceDynamicArray[Item]); ok {
		*sda = append(*sda, t...)
		return
	}
	i := len(*sda)
	*sda = append(*sda, make([]Item, n)...)
	s.Range(func(x Item) (cont bool) {
		(*sda)[i], i = x, i+1
		return true
	})
}

// Truncate removes the item with index i and all subsequent items.
//
// It does nothing if i is out of range.
func (sda *SliceDynamicArray[Item]) Truncate(i int) {
	if i < 0 || i >= len(*sda) {
		return
	}
	var zero Item
	for j := i; j < len(*sda); j++ {
		(*sda)[j] = zero // avoid memory leak
	}
	*sda = (*sda)[:i]
}

// Insert adds x as the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len().
func (sda *SliceDynamicArray[Item]) Insert(i int, x Item) {
	if i == len(*sda) {
		sda.Push(x)
		return
	}
	_ = (*sda)[i] // ensure i is valid
	var zero Item
	*sda = append(*sda, zero)
	copy((*sda)[i+1:], (*sda)[i:])
	(*sda)[i] = x
}

// Remove removes and returns the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i >= Len().
func (sda *SliceDynamicArray[Item]) Remove(i int) Item {
	back := len(*sda) - 1
	if i == back {
		return sda.Pop()
	}
	x := (*sda)[i]
	copy((*sda)[i:], (*sda)[i+1:])
	var zero Item
	(*sda)[back] = zero // avoid memory leak
	*sda = (*sda)[:back]
	return x
}

// RemoveWithoutOrder removes and returns the item with index i,
// without preserving order.
//
// It panics if i is out of range, i.e., i < 0 or i >= Len().
func (sda *SliceDynamicArray[Item]) RemoveWithoutOrder(i int) Item {
	x := (*sda)[i]
	back := len(*sda) - 1
	if i != back {
		(*sda)[i] = (*sda)[back]
	}
	var zero Item
	(*sda)[back] = zero // avoid memory leak
	*sda = (*sda)[:back]
	return x
}

// InsertSequence inserts s to the front of the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len().
//
// s shouldn't be modified during calling this method,
// otherwise, unknown error may occur.
func (sda *SliceDynamicArray[Item]) InsertSequence(i int, s sequence.Sequence[Item]) {
	if i == len(*sda) {
		sda.Append(s)
		return
	}
	_ = (*sda)[i] // ensure i is valid
	if s == nil {
		return
	}
	n := s.Len()
	if n == 0 {
		return
	}
	*sda = append(*sda, make([]Item, n)...)
	copy((*sda)[i+n:], (*sda)[i:])
	j := i
	s.Range(func(x Item) (cont bool) {
		(*sda)[j], j = x, j+1
		return true
	})
}

// Cut removes items from argument begin (inclusive) to
// argument end (exclusive) of the slice.
//
// It panics if begin or end is out of range, or begin > end.
func (sda *SliceDynamicArray[Item]) Cut(begin, end int) {
	_ = (*sda)[begin:end] // ensure begin and end are valid
	if begin == end {
		return
	}
	if end == len(*sda) {
		sda.Truncate(begin)
		return
	}
	copy((*sda)[begin:], (*sda)[end:])
	var zero Item
	for i := len(*sda) - end + begin; i < len(*sda); i++ {
		(*sda)[i] = zero // avoid memory leak
	}
	*sda = (*sda)[:len(*sda)-end+begin]
}

// CutWithoutOrder removes items from argument begin (inclusive) to
// argument end (exclusive) of the slice, without preserving order.
//
// It panics if begin or end is out of range, or begin > end.
func (sda *SliceDynamicArray[Item]) CutWithoutOrder(begin, end int) {
	_ = (*sda)[begin:end] // ensure begin and end are valid
	if begin == end {
		return
	}
	if end == len(*sda) {
		sda.Truncate(begin)
		return
	}
	copyIdx := len(*sda) - end + begin
	if copyIdx < end {
		copyIdx = end
	}
	copy((*sda)[begin:], (*sda)[copyIdx:])
	var zero Item
	for i := len(*sda) - end + begin; i < len(*sda); i++ {
		(*sda)[i] = zero // avoid memory leak
	}
	*sda = (*sda)[:len(*sda)-end+begin]
}

// Extend adds n zero-value items to the back of the slice.
//
// It panics if n < 0.
func (sda *SliceDynamicArray[Item]) Extend(n int) {
	*sda = append(*sda, make([]Item, n)...)
}

// Expand inserts n zero-value items to the front of the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len(), or n < 0.
func (sda *SliceDynamicArray[Item]) Expand(i, n int) {
	if i == len(*sda) {
		sda.Extend(n)
		return
	}
	_ = (*sda)[i] // ensure i is valid
	*sda = append(*sda, make([]Item, n)...)
	copy((*sda)[i+n:], (*sda)[i:])
	var zero Item
	for j := i; j < i+n; j++ {
		(*sda)[j] = zero
	}
}

// Reserve requests that the capacity of the slice
// is at least the specified capacity.
//
// It does nothing if capacity <= Cap().
func (sda *SliceDynamicArray[Item]) Reserve(capacity int) {
	if capacity <= 0 || (sda != nil && capacity <= cap(*sda)) {
		return
	}
	s := make(SliceDynamicArray[Item], len(*sda), capacity)
	copy(s, *sda)
	*sda = s
}

// Shrink reduces the slice to fit, i.e.,
// requests Cap() equals to Len().
//
// Note that it isn't equivalent to operations on Go slice
// like s[:len(s):len(s)],
// because it will allocate a new array and copy the content
// if Cap() > Len().
func (sda *SliceDynamicArray[Item]) Shrink() {
	if sda == nil || len(*sda) == cap(*sda) {
		return
	}
	s := make(SliceDynamicArray[Item], len(*sda))
	copy(s, *sda)
	*sda = s
}

// Clear sets the slice to nil.
func (sda *SliceDynamicArray[Item]) Clear() {
	if sda != nil {
		*sda = nil
	}
}
