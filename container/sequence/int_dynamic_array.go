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

// A prefab DynamicArray for int.
type IntDynamicArray []int

// Make a new IntDynamicArray with given capacity.
// It panics if capacity < 0.
func NewIntDynamicArray(capacity int) IntDynamicArray {
	return make(IntDynamicArray, 0, capacity)
}

func (ida IntDynamicArray) Len() int {
	return len(ida)
}

func (ida IntDynamicArray) Front() interface{} {
	return ida[0]
}

func (ida IntDynamicArray) SetFront(x interface{}) {
	ida[0] = x.(int)
}

func (ida IntDynamicArray) Back() interface{} {
	return ida[len(ida)-1]
}

func (ida IntDynamicArray) SetBack(x interface{}) {
	ida[len(ida)-1] = x.(int)
}

func (ida IntDynamicArray) Reverse() {
	for i, k := 0, len(ida)-1; i < k; i, k = i+1, k-1 {
		ida[i], ida[k] = ida[k], ida[i]
	}
}

func (ida IntDynamicArray) Scan(handler func(x interface{}) (cont bool)) {
	for _, x := range ida {
		if !handler(x) {
			return
		}
	}
}

func (ida IntDynamicArray) Get(i int) interface{} {
	return ida[i]
}

func (ida IntDynamicArray) Set(i int, x interface{}) {
	ida[i] = x.(int)
}

func (ida IntDynamicArray) Swap(i, j int) {
	ida[i], ida[j] = ida[j], ida[i]
}

func (ida IntDynamicArray) Slice(begin, end int) Array {
	return ida[begin:end:end]
}

func (ida IntDynamicArray) Cap() int {
	return cap(ida)
}

func (ida *IntDynamicArray) Push(x interface{}) {
	*ida = append(*ida, x.(int))
}

func (ida *IntDynamicArray) Pop() interface{} {
	back := len(*ida) - 1
	x := (*ida)[back]
	*ida = (*ida)[:back]
	return x
}

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
	s.Scan(func(x interface{}) (cont bool) {
		(*ida)[i] = x.(int)
		i++
		return true
	})
}

func (ida *IntDynamicArray) Truncate(i int) {
	if i < 0 || i >= len(*ida) {
		return
	}
	*ida = (*ida)[:i]
}

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

func (ida *IntDynamicArray) RemoveWithoutOrder(i int) interface{} {
	x := (*ida)[i]
	back := len(*ida) - 1
	if i != back {
		(*ida)[i] = (*ida)[back]
	}
	*ida = (*ida)[:back]
	return x
}

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
	s.Scan(func(x interface{}) (cont bool) {
		(*ida)[k] = x.(int)
		k++
		return true
	})
}

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

func (ida *IntDynamicArray) Extend(n int) {
	*ida = append(*ida, make([]int, n)...)
}

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

func (ida *IntDynamicArray) Reserve(capacity int) {
	if capacity <= 0 || (ida != nil && capacity <= cap(*ida)) {
		return
	}
	s := make(IntDynamicArray, len(*ida), capacity)
	copy(s, *ida)
	*ida = s
}

func (ida *IntDynamicArray) Shrink() {
	if ida == nil || len(*ida) == cap(*ida) {
		return
	}
	s := make(IntDynamicArray, len(*ida))
	copy(s, *ida)
	*ida = s
}

func (ida *IntDynamicArray) Clear() {
	if ida != nil {
		*ida = nil
	}
}

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

func (ida IntDynamicArray) Less(i, j int) bool {
	return ida[i] < ida[j]
}
