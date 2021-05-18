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

// StringDynamicArray is a prefab DynamicArray for string.
type StringDynamicArray []string

// NewStringDynamicArray makes a new StringDynamicArray
// with specified capacity.
// It panics if capacity < 0.
func NewStringDynamicArray(capacity int) StringDynamicArray {
	return make(StringDynamicArray, 0, capacity)
}

// Len returns the number of items in the array.
func (sda StringDynamicArray) Len() int {
	return len(sda)
}

// Front returns the first item.
//
// It panics if the array is nil or empty.
func (sda StringDynamicArray) Front() interface{} {
	return sda[0]
}

// SetFront sets the first item to x.
//
// It panics if the array is nil or empty.
func (sda StringDynamicArray) SetFront(x interface{}) {
	sda[0] = x.(string)
}

// Back returns the last item.
//
// It panics if the array is nil or empty.
func (sda StringDynamicArray) Back() interface{} {
	return sda[len(sda)-1]
}

// SetBack sets the last item to x.
//
// It panics if the array is nil or empty.
func (sda StringDynamicArray) SetBack(x interface{}) {
	sda[len(sda)-1] = x.(string)
}

// Reverse turns the other way round items in the array.
func (sda StringDynamicArray) Reverse() {
	for i, k := 0, len(sda)-1; i < k; i, k = i+1, k-1 {
		sda[i], sda[k] = sda[k], sda[i]
	}
}

// Range browses the items in the array from the first to the last.
//
// Its argument handler is a function to deal with the item x in the
// array and report whether to continue to check the next item or not.
func (sda StringDynamicArray) Range(handler func(x interface{}) (cont bool)) {
	for _, x := range sda {
		if !handler(x) {
			return
		}
	}
}

// Get returns the item with index i.
//
// It panics if i is out of range.
func (sda StringDynamicArray) Get(i int) interface{} {
	return sda[i]
}

// Set sets the item with index i to x.
//
// It panics if i is out of range.
func (sda StringDynamicArray) Set(i int, x interface{}) {
	sda[i] = x.(string)
}

// Swap exchanges the items with indexes i and j.
//
// It panics if i or j is out of range.
func (sda StringDynamicArray) Swap(i, j int) {
	sda[i], sda[j] = sda[j], sda[i]
}

// Slice returns a slice from argument begin (inclusive) to
// argument end (exclusive) of the array, as an Array.
//
// It panics if begin or end is out of range, or begin > end.
func (sda StringDynamicArray) Slice(begin, end int) Array {
	return sda[begin:end:end]
}

// Less reports whether the item with index i must sort before
// the item with index j.
//
// It panics if i or j is out of range.
func (sda StringDynamicArray) Less(i, j int) bool {
	return sda[i] < sda[j]
}

// Cap returns the current capacity of the dynamic array.
func (sda StringDynamicArray) Cap() int {
	return cap(sda)
}

// Push adds x to the back of the dynamic array.
func (sda *StringDynamicArray) Push(x interface{}) {
	*sda = append(*sda, x.(string))
}

// Pop removes and returns the last item.
//
// It panics if the dynamic array is nil or empty.
func (sda *StringDynamicArray) Pop() interface{} {
	back := len(*sda) - 1
	x := (*sda)[back]
	(*sda)[back] = "" // avoid memory leak
	*sda = (*sda)[:back]
	return x
}

// Append adds s to the back of the dynamic array.
//
// s shouldn't be modified during calling this method,
// otherwise, unknown error may occur.
func (sda *StringDynamicArray) Append(s Sequence) {
	if s == nil {
		return
	}
	n := s.Len()
	if n == 0 {
		return
	}
	if slice, ok := s.(StringDynamicArray); ok {
		*sda = append(*sda, slice...)
		return
	}
	i := len(*sda)
	*sda = append(*sda, make([]string, n)...)
	s.Range(func(x interface{}) (cont bool) {
		(*sda)[i] = x.(string)
		i++
		return true
	})
}

// Truncate removes the item with index i and all subsequent items.
//
// It does nothing if i is out of range.
func (sda *StringDynamicArray) Truncate(i int) {
	if i < 0 || i >= len(*sda) {
		return
	}
	for k := i; k < len(*sda); k++ {
		(*sda)[k] = "" // avoid memory leak
	}
	*sda = (*sda)[:i]
}

// Insert adds x as the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len().
func (sda *StringDynamicArray) Insert(i int, x interface{}) {
	if i == len(*sda) {
		sda.Push(x)
		return
	}
	_ = (*sda)[i] // ensure i is valid
	*sda = append(*sda, "")
	copy((*sda)[i+1:], (*sda)[i:])
	(*sda)[i] = x.(string)
}

// Remove removes and returns the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i >= Len().
func (sda *StringDynamicArray) Remove(i int) interface{} {
	back := len(*sda) - 1
	if i == back {
		return sda.Pop()
	}
	x := (*sda)[i]
	copy((*sda)[i:], (*sda)[i+1:])
	(*sda)[back] = "" // avoid memory leak
	*sda = (*sda)[:back]
	return x
}

// RemoveWithoutOrder removes and returns the item with index i,
// without preserving order.
//
// It panics if i is out of range, i.e., i < 0 or i >= Len().
func (sda *StringDynamicArray) RemoveWithoutOrder(i int) interface{} {
	x := (*sda)[i]
	back := len(*sda) - 1
	if i != back {
		(*sda)[i] = (*sda)[back]
	}
	(*sda)[back] = "" // avoid memory leak
	*sda = (*sda)[:back]
	return x
}

// InsertSequence inserts s to the front of the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len().
//
// s shouldn't be modified during calling this method,
// otherwise, unknown error may occur.
func (sda *StringDynamicArray) InsertSequence(i int, s Sequence) {
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
	*sda = append(*sda, make([]string, n)...)
	copy((*sda)[i+n:], (*sda)[i:])
	k := i
	s.Range(func(x interface{}) (cont bool) {
		(*sda)[k] = x.(string)
		k++
		return true
	})
}

// Cut removes items from argument begin (inclusive) to
// argument end (exclusive) of the dynamic array.
//
// It panics if begin or end is out of range, or begin > end.
func (sda *StringDynamicArray) Cut(begin, end int) {
	_ = (*sda)[begin:end] // ensure begin and end are valid
	if begin == end {
		return
	}
	if end == len(*sda) {
		sda.Truncate(begin)
		return
	}
	copy((*sda)[begin:], (*sda)[end:])
	for i := len(*sda) - end + begin; i < len(*sda); i++ {
		(*sda)[i] = "" // avoid memory leak
	}
	*sda = (*sda)[:len(*sda)-end+begin]
}

// CutWithoutOrder removes items from argument begin (inclusive) to
// argument end (exclusive) of the dynamic array, without preserving order.
//
// It panics if begin or end is out of range, or begin > end.
func (sda *StringDynamicArray) CutWithoutOrder(begin, end int) {
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
	for i := len(*sda) - end + begin; i < len(*sda); i++ {
		(*sda)[i] = "" // avoid memory leak
	}
	*sda = (*sda)[:len(*sda)-end+begin]
}

// Extend adds n zero-value items to the back of the dynamic array.
//
// It panics if n < 0.
func (sda *StringDynamicArray) Extend(n int) {
	*sda = append(*sda, make([]string, n)...)
}

// Expand inserts n zero-value items to the front of the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len(), or n < 0.
func (sda *StringDynamicArray) Expand(i, n int) {
	if i == len(*sda) {
		sda.Extend(n)
		return
	}
	_ = (*sda)[i] // ensure i is valid
	*sda = append(*sda, make([]string, n)...)
	copy((*sda)[i+n:], (*sda)[i:])
	for k := i; k < i+n; k++ {
		(*sda)[k] = ""
	}
}

// Reserve requests that the capacity of the dynamic array
// is at least the specified capacity.
//
// It does nothing if capacity <= Cap().
func (sda *StringDynamicArray) Reserve(capacity int) {
	if capacity <= 0 || (sda != nil && capacity <= cap(*sda)) {
		return
	}
	s := make(StringDynamicArray, len(*sda), capacity)
	copy(s, *sda)
	*sda = s
}

// Shrink reduces the dynamic array to fit, i.e.,
// requests Cap() equals to Len().
//
// Note that it isn't equivalent to operations on slice
// like s[:len(s):len(s)],
// because it will allocate a new array and copy the content
// if Cap() > Len().
func (sda *StringDynamicArray) Shrink() {
	if sda == nil || len(*sda) == cap(*sda) {
		return
	}
	s := make(StringDynamicArray, len(*sda))
	copy(s, *sda)
	*sda = s
}

// Clear removes all items in the dynamic array and
// asks to release the memory.
func (sda *StringDynamicArray) Clear() {
	if sda != nil {
		*sda = nil
	}
}

// Filter refines items in the dynamic array (in place).
//
// Its argument filter is a function to report
// whether to keep the item x or not.
func (sda *StringDynamicArray) Filter(filter func(x interface{}) (keep bool)) {
	if sda == nil || len(*sda) == 0 {
		return
	}
	n := 0
	for _, x := range *sda {
		if filter(x) {
			(*sda)[n] = x
			n++
		}
	}
	if n == len(*sda) {
		return
	}
	for i := n; i < len(*sda); i++ {
		(*sda)[i] = "" // avoid memory leak
	}
	*sda = (*sda)[:n]
}
