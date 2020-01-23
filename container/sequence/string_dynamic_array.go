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

// A prefab DynamicArray for string.
type StringDynamicArray []string

// Make a new StringDynamicArray with given capacity.
// It panics if capacity < 0.
func NewStringDynamicArray(capacity int) StringDynamicArray {
	return make(StringDynamicArray, 0, capacity)
}

func (sda StringDynamicArray) Len() int {
	return len(sda)
}

func (sda StringDynamicArray) Front() interface{} {
	return sda[0]
}

func (sda StringDynamicArray) SetFront(x interface{}) {
	sda[0] = x.(string)
}

func (sda StringDynamicArray) Back() interface{} {
	return sda[len(sda)-1]
}

func (sda StringDynamicArray) SetBack(x interface{}) {
	sda[len(sda)-1] = x.(string)
}

func (sda StringDynamicArray) Reverse() {
	for i, k := 0, len(sda)-1; i < k; i, k = i+1, k-1 {
		sda[i], sda[k] = sda[k], sda[i]
	}
}

func (sda StringDynamicArray) Scan(handler func(x interface{}) (cont bool)) {
	for _, x := range sda {
		if !handler(x) {
			return
		}
	}
}

func (sda StringDynamicArray) Get(i int) interface{} {
	return sda[i]
}

func (sda StringDynamicArray) Set(i int, x interface{}) {
	sda[i] = x.(string)
}

func (sda StringDynamicArray) Swap(i, j int) {
	sda[i], sda[j] = sda[j], sda[i]
}

func (sda StringDynamicArray) Slice(begin, end int) Array {
	return sda[begin:end:end]
}

func (sda StringDynamicArray) Cap() int {
	return cap(sda)
}

func (sda *StringDynamicArray) Push(x interface{}) {
	*sda = append(*sda, x.(string))
}

func (sda *StringDynamicArray) Pop() interface{} {
	back := len(*sda) - 1
	x := (*sda)[back]
	(*sda)[back] = "" // avoid memory leak
	*sda = (*sda)[:back]
	return x
}

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
	s.Scan(func(x interface{}) (cont bool) {
		(*sda)[i] = x.(string)
		i++
		return true
	})
}

func (sda *StringDynamicArray) Truncate(i int) {
	if i < 0 || i >= len(*sda) {
		return
	}
	for k := i; k < len(*sda); k++ {
		(*sda)[k] = "" // avoid memory leak
	}
	*sda = (*sda)[:i]
}

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
	s.Scan(func(x interface{}) (cont bool) {
		(*sda)[k] = x.(string)
		k++
		return true
	})
}

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

func (sda *StringDynamicArray) Extend(n int) {
	*sda = append(*sda, make([]string, n)...)
}

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

func (sda *StringDynamicArray) Reserve(capacity int) {
	if capacity <= 0 || (sda != nil && capacity <= cap(*sda)) {
		return
	}
	s := make(StringDynamicArray, len(*sda), capacity)
	copy(s, *sda)
	*sda = s
}

func (sda *StringDynamicArray) Shrink() {
	if sda == nil || len(*sda) == cap(*sda) {
		return
	}
	s := make(StringDynamicArray, len(*sda))
	copy(s, *sda)
	*sda = s
}

func (sda *StringDynamicArray) Clear() {
	if sda != nil {
		*sda = nil
	}
}

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

func (sda StringDynamicArray) Less(i, j int) bool {
	return sda[i] < sda[j]
}
