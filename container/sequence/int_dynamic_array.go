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

package sequence

// IntDynamicArray is a prefab DynamicArray for int.
type IntDynamicArray []int

// NewIntDynamicArray makes a new IntDynamicArray with specified capacity.
// It panics if capacity < 0.
func NewIntDynamicArray(capacity int) IntDynamicArray {
	return make(IntDynamicArray, 0, capacity)
}

// Len returns the number of items in the array.
func (ida IntDynamicArray) Len() int {
	return len(ida)
}

// Front returns the first item.
//
// It panics if the array is nil or empty.
func (ida IntDynamicArray) Front() interface{} {
	return ida[0]
}

// SetFront sets the first item to x.
//
// It panics if the array is nil or empty.
func (ida IntDynamicArray) SetFront(x interface{}) {
	ida[0] = x.(int)
}

// Back returns the last item.
//
// It panics if the array is nil or empty.
func (ida IntDynamicArray) Back() interface{} {
	return ida[len(ida)-1]
}

// SetBack sets the last item to x.
//
// It panics if the array is nil or empty.
func (ida IntDynamicArray) SetBack(x interface{}) {
	ida[len(ida)-1] = x.(int)
}

// Reverse turns the other way round items in the array.
func (ida IntDynamicArray) Reverse() {
	for i, k := 0, len(ida)-1; i < k; i, k = i+1, k-1 {
		ida[i], ida[k] = ida[k], ida[i]
	}
}

// Range browses the items in the array from the first to the last.
//
// Its argument handler is a function to deal with the item x in the
// array and report whether to continue to check the next item or not.
func (ida IntDynamicArray) Range(handler func(x interface{}) (cont bool)) {
	for _, x := range ida {
		if !handler(x) {
			return
		}
	}
}

// Get returns the item with index i.
//
// It panics if i is out of range.
func (ida IntDynamicArray) Get(i int) interface{} {
	return ida[i]
}

// Set sets the item with index i to x.
//
// It panics if i is out of range.
func (ida IntDynamicArray) Set(i int, x interface{}) {
	ida[i] = x.(int)
}

// Swap exchanges the items with indexes i and j.
//
// It panics if i or j is out of range.
func (ida IntDynamicArray) Swap(i, j int) {
	ida[i], ida[j] = ida[j], ida[i]
}

// Slice returns a slice from argument begin (inclusive) to
// argument end (exclusive) of the array, as an Array.
//
// It panics if begin or end is out of range, or begin > end.
func (ida IntDynamicArray) Slice(begin, end int) Array {
	return ida[begin:end:end]
}

// Less reports whether the item with index i must sort before
// the item with index j.
//
// It panics if i or j is out of range.
func (ida IntDynamicArray) Less(i, j int) bool {
	return ida[i] < ida[j]
}

// Cap returns the current capacity of the dynamic array.
func (ida IntDynamicArray) Cap() int {
	return cap(ida)
}

// Push adds x to the back of the dynamic array.
func (ida *IntDynamicArray) Push(x interface{}) {
	*ida = append(*ida, x.(int))
}

// Pop removes and returns the last item.
//
// It panics if the dynamic array is nil or empty.
func (ida *IntDynamicArray) Pop() interface{} {
	back := len(*ida) - 1
	x := (*ida)[back]
	*ida = (*ida)[:back]
	return x
}

// Append adds s to the back of the dynamic array.
//
// s shouldn't be modified during calling this method,
// otherwise, unknown error may occur.
func (ida *IntDynamicArray) Append(s Sequence) {
	if s == nil {
		return
	}
	n := s.Len()
	if n == 0 {
		return
	}
	if slice, ok := s.(IntDynamicArray); ok {
		*ida = append(*ida, slice...)
		return
	}
	i := len(*ida)
	*ida = append(*ida, make([]int, n)...)
	s.Range(func(x interface{}) (cont bool) {
		(*ida)[i] = x.(int)
		i++
		return true
	})
}

// Truncate removes the item with index i and all subsequent items.
//
// It does nothing if i is out of range.
func (ida *IntDynamicArray) Truncate(i int) {
	if i < 0 || i >= len(*ida) {
		return
	}
	*ida = (*ida)[:i]
}

// Insert adds x as the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len().
func (ida *IntDynamicArray) Insert(i int, x interface{}) {
	if i == len(*ida) {
		ida.Push(x)
		return
	}
	_ = (*ida)[i] // ensure i is valid
	*ida = append(*ida, 0)
	copy((*ida)[i+1:], (*ida)[i:])
	(*ida)[i] = x.(int)
}

// Remove removes and returns the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i >= Len().
func (ida *IntDynamicArray) Remove(i int) interface{} {
	back := len(*ida) - 1
	if i == back {
		return ida.Pop()
	}
	x := (*ida)[i]
	copy((*ida)[i:], (*ida)[i+1:])
	*ida = (*ida)[:back]
	return x
}

// RemoveWithoutOrder removes and returns the item with index i,
// without preserving order.
//
// It panics if i is out of range, i.e., i < 0 or i >= Len().
func (ida *IntDynamicArray) RemoveWithoutOrder(i int) interface{} {
	x := (*ida)[i]
	back := len(*ida) - 1
	if i != back {
		(*ida)[i] = (*ida)[back]
	}
	*ida = (*ida)[:back]
	return x
}

// InsertSequence inserts s to the front of the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len().
//
// s shouldn't be modified during calling this method,
// otherwise, unknown error may occur.
func (ida *IntDynamicArray) InsertSequence(i int, s Sequence) {
	if i == len(*ida) {
		ida.Append(s)
		return
	}
	_ = (*ida)[i] // ensure i is valid
	if s == nil {
		return
	}
	n := s.Len()
	if n == 0 {
		return
	}
	*ida = append(*ida, make([]int, n)...)
	copy((*ida)[i+n:], (*ida)[i:])
	k := i
	s.Range(func(x interface{}) (cont bool) {
		(*ida)[k] = x.(int)
		k++
		return true
	})
}

// Cut removes items from argument begin (inclusive) to
// argument end (exclusive) of the dynamic array.
//
// It panics if begin or end is out of range, or begin > end.
func (ida *IntDynamicArray) Cut(begin, end int) {
	_ = (*ida)[begin:end] // ensure begin and end are valid
	if begin == end {
		return
	}
	if end == len(*ida) {
		ida.Truncate(begin)
		return
	}
	copy((*ida)[begin:], (*ida)[end:])
	*ida = (*ida)[:len(*ida)-end+begin]
}

// CutWithoutOrder removes items from argument begin (inclusive) to
// argument end (exclusive) of the dynamic array, without preserving order.
//
// It panics if begin or end is out of range, or begin > end.
func (ida *IntDynamicArray) CutWithoutOrder(begin, end int) {
	_ = (*ida)[begin:end] // ensure begin and end are valid
	if begin == end {
		return
	}
	if end == len(*ida) {
		ida.Truncate(begin)
		return
	}
	copyIdx := len(*ida) - end + begin
	if copyIdx < end {
		copyIdx = end
	}
	copy((*ida)[begin:], (*ida)[copyIdx:])
	*ida = (*ida)[:len(*ida)-end+begin]
}

// Extend adds n zero-value items to the back of the dynamic array.
//
// It panics if n < 0.
func (ida *IntDynamicArray) Extend(n int) {
	*ida = append(*ida, make([]int, n)...)
}

// Expand inserts n zero-value items to the front of the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len(), or n < 0.
func (ida *IntDynamicArray) Expand(i, n int) {
	if i == len(*ida) {
		ida.Extend(n)
		return
	}
	_ = (*ida)[i] // ensure i is valid
	*ida = append(*ida, make([]int, n)...)
	copy((*ida)[i+n:], (*ida)[i:])
	for k := i; k < i+n; k++ {
		(*ida)[k] = 0
	}
}

// Reserve requests that the capacity of the dynamic array
// is at least the specified capacity.
//
// It does nothing if capacity <= Cap().
func (ida *IntDynamicArray) Reserve(capacity int) {
	if capacity <= 0 || (ida != nil && capacity <= cap(*ida)) {
		return
	}
	s := make(IntDynamicArray, len(*ida), capacity)
	copy(s, *ida)
	*ida = s
}

// Shrink reduces the dynamic array to fit, i.e.,
// requests Cap() equals to Len().
//
// Note that it isn't equivalent to operations on slice
// like s[:len(s):len(s)],
// because it will allocate a new array and copy the content
// if Cap() > Len().
func (ida *IntDynamicArray) Shrink() {
	if ida == nil || len(*ida) == cap(*ida) {
		return
	}
	s := make(IntDynamicArray, len(*ida))
	copy(s, *ida)
	*ida = s
}

// Clear removes all items in the dynamic array and
// asks to release the memory.
func (ida *IntDynamicArray) Clear() {
	if ida != nil {
		*ida = nil
	}
}

// Filter refines items in the dynamic array (in place).
//
// Its argument filter is a function to report
// whether to keep the item x or not.
func (ida *IntDynamicArray) Filter(filter func(x interface{}) (keep bool)) {
	if ida == nil || len(*ida) == 0 {
		return
	}
	n := 0
	for _, x := range *ida {
		if filter(x) {
			(*ida)[n] = x
			n++
		}
	}
	*ida = (*ida)[:n]
}
