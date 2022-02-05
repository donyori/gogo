// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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

// Float64DynamicArray is a prefab DynamicArray for float64.
type Float64DynamicArray []float64

// NewFloat64DynamicArray makes a new Float64DynamicArray
// with specified capacity.
// It panics if capacity < 0.
func NewFloat64DynamicArray(capacity int) Float64DynamicArray {
	return make(Float64DynamicArray, 0, capacity)
}

// Len returns the number of items in the array.
func (fda Float64DynamicArray) Len() int {
	return len(fda)
}

// Front returns the first item.
//
// It panics if the array is nil or empty.
func (fda Float64DynamicArray) Front() interface{} {
	return fda[0]
}

// SetFront sets the first item to x.
//
// It panics if the array is nil or empty.
func (fda Float64DynamicArray) SetFront(x interface{}) {
	fda[0] = x.(float64)
}

// Back returns the last item.
//
// It panics if the array is nil or empty.
func (fda Float64DynamicArray) Back() interface{} {
	return fda[len(fda)-1]
}

// SetBack sets the last item to x.
//
// It panics if the array is nil or empty.
func (fda Float64DynamicArray) SetBack(x interface{}) {
	fda[len(fda)-1] = x.(float64)
}

// Reverse turns the other way round items in the array.
func (fda Float64DynamicArray) Reverse() {
	for i, k := 0, len(fda)-1; i < k; i, k = i+1, k-1 {
		fda[i], fda[k] = fda[k], fda[i]
	}
}

// Range browses the items in the array from the first to the last.
//
// Its argument handler is a function to deal with the item x in the
// array and report whether to continue to check the next item.
func (fda Float64DynamicArray) Range(handler func(x interface{}) (cont bool)) {
	for _, x := range fda {
		if !handler(x) {
			return
		}
	}
}

// Get returns the item with index i.
//
// It panics if i is out of range.
func (fda Float64DynamicArray) Get(i int) interface{} {
	return fda[i]
}

// Set sets the item with index i to x.
//
// It panics if i is out of range.
func (fda Float64DynamicArray) Set(i int, x interface{}) {
	fda[i] = x.(float64)
}

// Swap exchanges the items with indexes i and j.
//
// It panics if i or j is out of range.
func (fda Float64DynamicArray) Swap(i, j int) {
	fda[i], fda[j] = fda[j], fda[i]
}

// Slice returns a slice from argument begin (inclusive) to
// argument end (exclusive) of the array, as an Array.
//
// It panics if begin or end is out of range, or begin > end.
func (fda Float64DynamicArray) Slice(begin, end int) Array {
	return fda[begin:end:end]
}

// Less reports whether the item with index i must sort before
// the item with index j.
//
// It implements a transitive ordering:
//  - if both Less(i, j) and Less(j, k) are true, then Less(i, k) must be true as well.
//  - if both Less(i, j) and Less(j, k) are false, then Less(i, k) must be false as well.
// It treats NaN values as less than any others.
//
// It panics if i or j is out of range.
func (fda Float64DynamicArray) Less(i, j int) bool {
	return fda[i] < fda[j] || (isNaN(fda[i]) && !isNaN(fda[j]))
}

// Cap returns the current capacity of the dynamic array.
func (fda Float64DynamicArray) Cap() int {
	return cap(fda)
}

// Push adds x to the back of the dynamic array.
func (fda *Float64DynamicArray) Push(x interface{}) {
	*fda = append(*fda, x.(float64))
}

// Pop removes and returns the last item.
//
// It panics if the dynamic array is nil or empty.
func (fda *Float64DynamicArray) Pop() interface{} {
	back := len(*fda) - 1
	x := (*fda)[back]
	*fda = (*fda)[:back]
	return x
}

// Append adds s to the back of the dynamic array.
//
// s shouldn't be modified during calling this method,
// otherwise, unknown error may occur.
func (fda *Float64DynamicArray) Append(s Sequence) {
	if s == nil {
		return
	}
	n := s.Len()
	if n == 0 {
		return
	}
	if slice, ok := s.(Float64DynamicArray); ok {
		*fda = append(*fda, slice...)
		return
	}
	i := len(*fda)
	*fda = append(*fda, make([]float64, n)...)
	s.Range(func(x interface{}) (cont bool) {
		(*fda)[i] = x.(float64)
		i++
		return true
	})
}

// Truncate removes the item with index i and all subsequent items.
//
// It does nothing if i is out of range.
func (fda *Float64DynamicArray) Truncate(i int) {
	if i < 0 || i >= len(*fda) {
		return
	}
	*fda = (*fda)[:i]
}

// Insert adds x as the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len().
func (fda *Float64DynamicArray) Insert(i int, x interface{}) {
	if i == len(*fda) {
		fda.Push(x)
		return
	}
	_ = (*fda)[i] // ensure i is valid
	*fda = append(*fda, 0)
	copy((*fda)[i+1:], (*fda)[i:])
	(*fda)[i] = x.(float64)
}

// Remove removes and returns the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i >= Len().
func (fda *Float64DynamicArray) Remove(i int) interface{} {
	back := len(*fda) - 1
	if i == back {
		return fda.Pop()
	}
	x := (*fda)[i]
	copy((*fda)[i:], (*fda)[i+1:])
	*fda = (*fda)[:back]
	return x
}

// RemoveWithoutOrder removes and returns the item with index i,
// without preserving order.
//
// It panics if i is out of range, i.e., i < 0 or i >= Len().
func (fda *Float64DynamicArray) RemoveWithoutOrder(i int) interface{} {
	x := (*fda)[i]
	back := len(*fda) - 1
	if i != back {
		(*fda)[i] = (*fda)[back]
	}
	*fda = (*fda)[:back]
	return x
}

// InsertSequence inserts s to the front of the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len().
//
// s shouldn't be modified during calling this method,
// otherwise, unknown error may occur.
func (fda *Float64DynamicArray) InsertSequence(i int, s Sequence) {
	if i == len(*fda) {
		fda.Append(s)
		return
	}
	_ = (*fda)[i] // ensure i is valid
	if s == nil {
		return
	}
	n := s.Len()
	if n == 0 {
		return
	}
	*fda = append(*fda, make([]float64, n)...)
	copy((*fda)[i+n:], (*fda)[i:])
	k := i
	s.Range(func(x interface{}) (cont bool) {
		(*fda)[k] = x.(float64)
		k++
		return true
	})
}

// Cut removes items from argument begin (inclusive) to
// argument end (exclusive) of the dynamic array.
//
// It panics if begin or end is out of range, or begin > end.
func (fda *Float64DynamicArray) Cut(begin, end int) {
	_ = (*fda)[begin:end] // ensure begin and end are valid
	if begin == end {
		return
	}
	if end == len(*fda) {
		fda.Truncate(begin)
		return
	}
	copy((*fda)[begin:], (*fda)[end:])
	*fda = (*fda)[:len(*fda)-end+begin]
}

// CutWithoutOrder removes items from argument begin (inclusive) to
// argument end (exclusive) of the dynamic array, without preserving order.
//
// It panics if begin or end is out of range, or begin > end.
func (fda *Float64DynamicArray) CutWithoutOrder(begin, end int) {
	_ = (*fda)[begin:end] // ensure begin and end are valid
	if begin == end {
		return
	}
	if end == len(*fda) {
		fda.Truncate(begin)
		return
	}
	copyIdx := len(*fda) - end + begin
	if copyIdx < end {
		copyIdx = end
	}
	copy((*fda)[begin:], (*fda)[copyIdx:])
	*fda = (*fda)[:len(*fda)-end+begin]
}

// Extend adds n zero-value items to the back of the dynamic array.
//
// It panics if n < 0.
func (fda *Float64DynamicArray) Extend(n int) {
	*fda = append(*fda, make([]float64, n)...)
}

// Expand inserts n zero-value items to the front of the item with index i.
//
// It panics if i is out of range, i.e., i < 0 or i > Len(), or n < 0.
func (fda *Float64DynamicArray) Expand(i, n int) {
	if i == len(*fda) {
		fda.Extend(n)
		return
	}
	_ = (*fda)[i] // ensure i is valid
	*fda = append(*fda, make([]float64, n)...)
	copy((*fda)[i+n:], (*fda)[i:])
	for k := i; k < i+n; k++ {
		(*fda)[k] = 0.
	}
}

// Reserve requests that the capacity of the dynamic array
// is at least the specified capacity.
//
// It does nothing if capacity <= Cap().
func (fda *Float64DynamicArray) Reserve(capacity int) {
	if capacity <= 0 || (fda != nil && capacity <= cap(*fda)) {
		return
	}
	s := make(Float64DynamicArray, len(*fda), capacity)
	copy(s, *fda)
	*fda = s
}

// Shrink reduces the dynamic array to fit, i.e.,
// requests Cap() equals to Len().
//
// Note that it isn't equivalent to operations on slice
// like s[:len(s):len(s)],
// because it will allocate a new array and copy the content
// if Cap() > Len().
func (fda *Float64DynamicArray) Shrink() {
	if fda == nil || len(*fda) == cap(*fda) {
		return
	}
	s := make(Float64DynamicArray, len(*fda))
	copy(s, *fda)
	*fda = s
}

// Clear removes all items in the dynamic array and
// asks to release the memory.
func (fda *Float64DynamicArray) Clear() {
	if fda != nil {
		*fda = nil
	}
}

// Filter refines items in the dynamic array (in place).
//
// Its argument filter is a function to report
// whether to keep the item x.
func (fda *Float64DynamicArray) Filter(filter func(x interface{}) (keep bool)) {
	if fda == nil || len(*fda) == 0 {
		return
	}
	var n int
	for _, x := range *fda {
		if filter(x) {
			(*fda)[n] = x
			n++
		}
	}
	*fda = (*fda)[:n]
}

// isNaN is a copy of math.IsNaN to avoid a dependency on the math package.
func isNaN(f float64) bool {
	return f != f
}
