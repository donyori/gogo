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
	"fmt"
	"slices"

	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/errors"
)

// SliceDynamicArray is a dynamic array wrapped on Go slice.
// *SliceDynamicArray implements the interface DynamicArray.
//
// The client can convert a Go slice to SliceDynamicArray by type conversion,
// e.g.:
//
//	SliceDynamicArray[int]([]int{1, 2, 3})
//
// Or allocate a new SliceDynamicArray by the slice literal or
// the built-in function make, e.g.:
//
//	SliceDynamicArray[int]{1, 2, 3}
//	make(SliceDynamicArray[int], 2, 10)
type SliceDynamicArray[Item any] []Item

var _ DynamicArray[any] = (*SliceDynamicArray[any])(nil)

// Len returns the number of items in the slice.
//
// It returns 0 if the slice is nil.
func (sda *SliceDynamicArray[Item]) Len() int {
	var n int
	if sda != nil {
		n = len(*sda)
	}
	return n
}

// Range accesses the items in the slice from first to last.
// Each item is accessed once.
//
// Its parameter handler is a function to deal with the item x in the
// slice and report whether to continue to access the next item.
func (sda *SliceDynamicArray[Item]) Range(handler func(x Item) (cont bool)) {
	if sda != nil {
		for _, x := range *sda {
			if !handler(x) {
				return
			}
		}
	}
}

// Front returns the first item.
//
// It panics if the slice is nil or empty.
func (sda *SliceDynamicArray[Item]) Front() Item {
	sda.checkNonEmpty()
	return (*sda)[0]
}

// SetFront sets the first item to x.
//
// It panics if the slice is nil or empty.
func (sda *SliceDynamicArray[Item]) SetFront(x Item) {
	sda.checkNonEmpty()
	(*sda)[0] = x
}

// Back returns the last item.
//
// It panics if the slice is nil or empty.
func (sda *SliceDynamicArray[Item]) Back() Item {
	sda.checkNonEmpty()
	return (*sda)[len(*sda)-1]
}

// SetBack sets the last item to x.
//
// It panics if the slice is nil or empty.
func (sda *SliceDynamicArray[Item]) SetBack(x Item) {
	sda.checkNonEmpty()
	(*sda)[len(*sda)-1] = x
}

// Reverse turns items in the slice the other way round.
func (sda *SliceDynamicArray[Item]) Reverse() {
	if sda != nil {
		slices.Reverse(*sda)
	}
}

// Get returns the item with index i.
//
// It panics if i is out of range.
func (sda *SliceDynamicArray[Item]) Get(i int) Item {
	sda.checkNonEmpty()
	return (*sda)[i]
}

// Set sets the item with index i to x.
//
// It panics if i is out of range.
func (sda *SliceDynamicArray[Item]) Set(i int, x Item) {
	sda.checkNonEmpty()
	(*sda)[i] = x
}

// Swap exchanges the items with indexes i and j.
//
// It panics if i or j is out of range.
func (sda *SliceDynamicArray[Item]) Swap(i, j int) {
	sda.checkNonEmpty()
	(*sda)[i], (*sda)[j] = (*sda)[j], (*sda)[i]
}

// Slice returns a slice from argument begin (inclusive) to
// argument end (exclusive) of the slice, as an Array.
//
// It panics if begin or end is out of range, or begin > end.
func (sda *SliceDynamicArray[Item]) Slice(begin, end int) Array[Item] {
	sda.checkNonEmpty()
	s := (*sda)[begin:end:end]
	return &s
}

// Filter refines items in the slice (in-place).
//
// Its parameter filter is a function to report whether to keep the item x.
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
	clear((*sda)[n:]) // avoid memory leak
	*sda = (*sda)[:n]
}

// Cap returns the current capacity of the slice.
//
// It returns 0 if the slice is nil.
func (sda *SliceDynamicArray[Item]) Cap() int {
	var c int
	if sda != nil {
		c = cap(*sda)
	}
	return c
}

// Push adds x to the back of the slice.
func (sda *SliceDynamicArray[Item]) Push(x Item) {
	if sda == nil {
		panic(errors.AutoMsg(nilSliceDynamicArrayPointerPanicMessage))
	}
	*sda = append(*sda, x)
}

// Pop removes and returns the last item.
//
// It panics if the slice is nil or empty.
func (sda *SliceDynamicArray[Item]) Pop() Item {
	sda.checkNonEmpty()
	back := len(*sda) - 1
	x := (*sda)[back]
	clear((*sda)[back:]) // avoid memory leak
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
	} else if sda == nil {
		panic(errors.AutoMsg(nilSliceDynamicArrayPointerPanicMessage))
	} else if t, ok := s.(*SliceDynamicArray[Item]); ok {
		*sda = append(*sda, *t...)
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
	if sda == nil || i < 0 || i >= len(*sda) {
		return
	}
	clear((*sda)[i:]) // avoid memory leak
	*sda = (*sda)[:i]
}

// Insert adds x as the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len().
func (sda *SliceDynamicArray[Item]) Insert(i int, x Item) {
	if sda == nil {
		panic(errors.AutoMsg(nilSliceDynamicArrayPointerPanicMessage))
	} else if i == len(*sda) {
		sda.Push(x)
		return
	}
	_ = (*sda)[i] // ensure i is valid
	*sda = slices.Insert(*sda, i, x)
}

// Remove removes and returns the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i >= Len().
func (sda *SliceDynamicArray[Item]) Remove(i int) Item {
	sda.checkNonEmpty()
	back := len(*sda) - 1
	if i == back {
		return sda.Pop()
	}
	x := (*sda)[i]
	copy((*sda)[i:], (*sda)[i+1:])
	clear((*sda)[back:]) // avoid memory leak
	*sda = (*sda)[:back]
	return x
}

// RemoveWithoutOrder removes and returns the item with index i,
// without preserving order.
//
// It panics if i is out of range, i.e., i < 0 or i >= Len().
func (sda *SliceDynamicArray[Item]) RemoveWithoutOrder(i int) Item {
	sda.checkNonEmpty()
	x := (*sda)[i]
	back := len(*sda) - 1
	if i != back {
		(*sda)[i] = (*sda)[back]
	}
	clear((*sda)[back:]) // avoid memory leak
	*sda = (*sda)[:back]
	return x
}

// InsertSequence inserts s to the front of the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len().
//
// s shouldn't be modified during calling this method,
// otherwise, unknown error may occur.
func (sda *SliceDynamicArray[Item]) InsertSequence(
	i int, s sequence.Sequence[Item]) {
	if sda == nil {
		panic(errors.AutoMsg(nilSliceDynamicArrayPointerPanicMessage))
	} else if i == len(*sda) {
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
	} else if sda2, ok := s.(*SliceDynamicArray[Item]); ok {
		*sda = slices.Insert[SliceDynamicArray[Item], Item](*sda, i, *sda2...)
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
	sda.checkNonEmpty()
	_ = (*sda)[begin:end] // ensure begin and end are valid
	if begin == end {
		return
	} else if end == len(*sda) {
		sda.Truncate(begin)
		return
	}
	copy((*sda)[begin:], (*sda)[end:])
	clear((*sda)[len(*sda)-end+begin:]) // avoid memory leak
	*sda = (*sda)[:len(*sda)-end+begin]
}

// CutWithoutOrder removes items from argument begin (inclusive) to
// argument end (exclusive) of the slice, without preserving order.
//
// It panics if begin or end is out of range, or begin > end.
func (sda *SliceDynamicArray[Item]) CutWithoutOrder(begin, end int) {
	sda.checkNonEmpty()
	_ = (*sda)[begin:end] // ensure begin and end are valid
	if begin == end {
		return
	} else if end == len(*sda) {
		sda.Truncate(begin)
		return
	}
	copyIdx := len(*sda) - end + begin
	if copyIdx < end {
		copyIdx = end
	}
	copy((*sda)[begin:], (*sda)[copyIdx:])
	clear((*sda)[len(*sda)-end+begin:]) // avoid memory leak
	*sda = (*sda)[:len(*sda)-end+begin]
}

// Extend adds n zero-value items to the back of the slice.
//
// It panics if n < 0.
func (sda *SliceDynamicArray[Item]) Extend(n int) {
	if sda == nil {
		panic(errors.AutoMsg(nilSliceDynamicArrayPointerPanicMessage))
	} else if n < 0 {
		panic(errors.AutoMsg(fmt.Sprintf("n is %d < 0", n)))
	}
	i := len(*sda)
	if i+n > cap(*sda) {
		*sda = append(*sda, make([]Item, n)...)
		return
	}
	*sda = (*sda)[:i+n]
	clear((*sda)[i:])
}

// Expand inserts n zero-value items to the front of the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len(), or n < 0.
func (sda *SliceDynamicArray[Item]) Expand(i, n int) {
	switch {
	case sda == nil:
		panic(errors.AutoMsg(nilSliceDynamicArrayPointerPanicMessage))
	case n < 0:
		panic(errors.AutoMsg(fmt.Sprintf("n is %d < 0", n)))
	case i == len(*sda):
		sda.Extend(n)
		return
	}
	_ = (*sda)[i] // ensure i is valid
	if len(*sda)+n > cap(*sda) {
		*sda = append(*sda, make([]Item, n)...)
	} else {
		*sda = (*sda)[:len(*sda)+n]
	}
	copy((*sda)[i+n:], (*sda)[i:])
	clear((*sda)[i : i+n])
}

// Reserve requests that the capacity of the slice
// is at least the specified capacity.
//
// It does nothing if capacity <= Cap().
func (sda *SliceDynamicArray[Item]) Reserve(capacity int) {
	switch {
	case capacity <= 0:
		return
	case sda == nil:
		panic(errors.AutoMsg(nilSliceDynamicArrayPointerPanicMessage))
	case capacity <= cap(*sda):
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
// because it allocates a new array and copies the content
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

// checkNonEmpty panics if sda is nil, *sda is nil, or len(*sda) is 0.
func (sda *SliceDynamicArray[Item]) checkNonEmpty() {
	switch {
	case sda == nil:
		panic(errors.AutoMsgCustom(
			nilSliceDynamicArrayPointerPanicMessage, -1, 1))
	case *sda == nil:
		panic(errors.AutoMsgCustom(
			nilSliceDynamicArrayPanicMessage, -1, 1))
	case len(*sda) == 0:
		panic(errors.AutoMsgCustom(
			emptySliceDynamicArrayPanicMessage, -1, 1))
	}
}

const (
	nilSliceDynamicArrayPointerPanicMessage = "*SliceDynamicArray[...] is nil"
	nilSliceDynamicArrayPanicMessage        = "SliceDynamicArray[...] is nil"
	emptySliceDynamicArrayPanicMessage      = "SliceDynamicArray[...] is empty"
)
