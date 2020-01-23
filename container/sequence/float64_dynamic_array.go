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

// A prefab DynamicArray for float64.
type Float64DynamicArray []float64

// Make a new Float64DynamicArray with given capacity.
// It panics if capacity < 0.
func NewFloat64DynamicArray(capacity int) Float64DynamicArray {
	return make(Float64DynamicArray, 0, capacity)
}

func (fda Float64DynamicArray) Len() int {
	return len(fda)
}

func (fda Float64DynamicArray) Front() interface{} {
	return fda[0]
}

func (fda Float64DynamicArray) SetFront(x interface{}) {
	fda[0] = x.(float64)
}

func (fda Float64DynamicArray) Back() interface{} {
	return fda[len(fda)-1]
}

func (fda Float64DynamicArray) SetBack(x interface{}) {
	fda[len(fda)-1] = x.(float64)
}

func (fda Float64DynamicArray) Reverse() {
	for i, k := 0, len(fda)-1; i < k; i, k = i+1, k-1 {
		fda[i], fda[k] = fda[k], fda[i]
	}
}

func (fda Float64DynamicArray) Scan(handler func(x interface{}) (cont bool)) {
	for _, x := range fda {
		if !handler(x) {
			return
		}
	}
}

func (fda Float64DynamicArray) Get(i int) interface{} {
	return fda[i]
}

func (fda Float64DynamicArray) Set(i int, x interface{}) {
	fda[i] = x.(float64)
}

func (fda Float64DynamicArray) Swap(i, j int) {
	fda[i], fda[j] = fda[j], fda[i]
}

func (fda Float64DynamicArray) Slice(begin, end int) Array {
	return fda[begin:end:end]
}

func (fda Float64DynamicArray) Cap() int {
	return cap(fda)
}

func (fda *Float64DynamicArray) Push(x interface{}) {
	*fda = append(*fda, x.(float64))
}

func (fda *Float64DynamicArray) Pop() interface{} {
	back := len(*fda) - 1
	x := (*fda)[back]
	*fda = (*fda)[:back]
	return x
}

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
	s.Scan(func(x interface{}) (cont bool) {
		(*fda)[i] = x.(float64)
		i++
		return true
	})
}

func (fda *Float64DynamicArray) Truncate(i int) {
	if i < 0 || i >= len(*fda) {
		return
	}
	*fda = (*fda)[:i]
}

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

func (fda *Float64DynamicArray) RemoveWithoutOrder(i int) interface{} {
	x := (*fda)[i]
	back := len(*fda) - 1
	if i != back {
		(*fda)[i] = (*fda)[back]
	}
	*fda = (*fda)[:back]
	return x
}

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
	s.Scan(func(x interface{}) (cont bool) {
		(*fda)[k] = x.(float64)
		k++
		return true
	})
}

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

func (fda *Float64DynamicArray) Extend(n int) {
	*fda = append(*fda, make([]float64, n)...)
}

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

func (fda *Float64DynamicArray) Reserve(capacity int) {
	if capacity <= 0 || (fda != nil && capacity <= cap(*fda)) {
		return
	}
	s := make(Float64DynamicArray, len(*fda), capacity)
	copy(s, *fda)
	*fda = s
}

func (fda *Float64DynamicArray) Shrink() {
	if fda == nil || len(*fda) == cap(*fda) {
		return
	}
	s := make(Float64DynamicArray, len(*fda))
	copy(s, *fda)
	*fda = s
}

func (fda *Float64DynamicArray) Clear() {
	if fda != nil {
		*fda = nil
	}
}

func (fda *Float64DynamicArray) Filter(filter func(x interface{}) (keep bool)) {
	if fda == nil || len(*fda) == 0 {
		return
	}
	n := 0
	for _, x := range *fda {
		if filter(x) {
			(*fda)[n] = x
			n++
		}
	}
	*fda = (*fda)[:n]
}

func (fda Float64DynamicArray) Less(i, j int) bool {
	return fda[i] < fda[j]
}
