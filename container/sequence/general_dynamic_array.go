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

// A prefab DynamicArray for interface{}.
type GeneralDynamicArray []interface{}

// Make a new GeneralDynamicArray with given capacity.
// It panics if capacity < 0.
func NewGeneralDynamicArray(capacity int) GeneralDynamicArray {
	return make(GeneralDynamicArray, 0, capacity)
}

func (gda GeneralDynamicArray) Len() int {
	return len(gda)
}

func (gda GeneralDynamicArray) Front() interface{} {
	return gda[0]
}

func (gda GeneralDynamicArray) SetFront(x interface{}) {
	gda[0] = x
}

func (gda GeneralDynamicArray) Back() interface{} {
	return gda[len(gda)-1]
}

func (gda GeneralDynamicArray) SetBack(x interface{}) {
	gda[len(gda)-1] = x
}

func (gda GeneralDynamicArray) Reverse() {
	for i, k := 0, len(gda)-1; i < k; i, k = i+1, k-1 {
		gda[i], gda[k] = gda[k], gda[i]
	}
}

func (gda GeneralDynamicArray) Scan(handler func(x interface{}) (cont bool)) {
	for _, x := range gda {
		if !handler(x) {
			return
		}
	}
}

func (gda GeneralDynamicArray) Get(i int) interface{} {
	return gda[i]
}

func (gda GeneralDynamicArray) Set(i int, x interface{}) {
	gda[i] = x
}

func (gda GeneralDynamicArray) Swap(i, j int) {
	gda[i], gda[j] = gda[j], gda[i]
}

func (gda GeneralDynamicArray) Slice(begin, end int) Array {
	return gda[begin:end:end]
}

func (gda GeneralDynamicArray) Cap() int {
	return cap(gda)
}

func (gda *GeneralDynamicArray) Push(x interface{}) {
	*gda = append(*gda, x)
}

func (gda *GeneralDynamicArray) Pop() interface{} {
	back := len(*gda) - 1
	x := (*gda)[back]
	(*gda)[back] = nil // avoid memory leak
	*gda = (*gda)[:back]
	return x
}

func (gda *GeneralDynamicArray) Append(s Sequence) {
	if s == nil {
		return
	}
	n := s.Len()
	if n == 0 {
		return
	}
	if slice, ok := s.(GeneralDynamicArray); ok {
		*gda = append(*gda, slice...)
		return
	}
	i := len(*gda)
	*gda = append(*gda, make([]interface{}, n)...)
	s.Scan(func(x interface{}) (cont bool) {
		(*gda)[i] = x
		i++
		return true
	})
}

func (gda *GeneralDynamicArray) Truncate(i int) {
	if i < 0 || i >= len(*gda) {
		return
	}
	for k := i; k < len(*gda); k++ {
		(*gda)[k] = nil // avoid memory leak
	}
	*gda = (*gda)[:i]
}

func (gda *GeneralDynamicArray) Insert(i int, x interface{}) {
	if i == len(*gda) {
		gda.Push(x)
		return
	}
	_ = (*gda)[i] // ensure i is valid
	*gda = append(*gda, nil)
	copy((*gda)[i+1:], (*gda)[i:])
	(*gda)[i] = x
}

func (gda *GeneralDynamicArray) Remove(i int) interface{} {
	back := len(*gda) - 1
	if i == back {
		return gda.Pop()
	}
	x := (*gda)[i]
	copy((*gda)[i:], (*gda)[i+1:])
	(*gda)[back] = nil // avoid memory leak
	*gda = (*gda)[:back]
	return x
}

func (gda *GeneralDynamicArray) RemoveWithoutOrder(i int) interface{} {
	x := (*gda)[i]
	back := len(*gda) - 1
	if i != back {
		(*gda)[i] = (*gda)[back]
	}
	(*gda)[back] = nil // avoid memory leak
	*gda = (*gda)[:back]
	return x
}

func (gda *GeneralDynamicArray) InsertSequence(i int, s Sequence) {
	if i == len(*gda) {
		gda.Append(s)
		return
	}
	_ = (*gda)[i] // ensure i is valid
	if s == nil {
		return
	}
	n := s.Len()
	if n == 0 {
		return
	}
	*gda = append(*gda, make([]interface{}, n)...)
	copy((*gda)[i+n:], (*gda)[i:])
	k := i
	s.Scan(func(x interface{}) (cont bool) {
		(*gda)[k] = x
		k++
		return true
	})
}

func (gda *GeneralDynamicArray) Cut(begin, end int) {
	_ = (*gda)[begin:end] // ensure begin and end are valid
	if begin == end {
		return
	}
	if end == len(*gda) {
		gda.Truncate(begin)
		return
	}
	copy((*gda)[begin:], (*gda)[end:])
	for i := len(*gda) - end + begin; i < len(*gda); i++ {
		(*gda)[i] = nil // avoid memory leak
	}
	*gda = (*gda)[:len(*gda)-end+begin]
}

func (gda *GeneralDynamicArray) CutWithoutOrder(begin, end int) {
	_ = (*gda)[begin:end] // ensure begin and end are valid
	if begin == end {
		return
	}
	if end == len(*gda) {
		gda.Truncate(begin)
		return
	}
	copyIdx := len(*gda) - end + begin
	if copyIdx < end {
		copyIdx = end
	}
	copy((*gda)[begin:], (*gda)[copyIdx:])
	for i := len(*gda) - end + begin; i < len(*gda); i++ {
		(*gda)[i] = nil // avoid memory leak
	}
	*gda = (*gda)[:len(*gda)-end+begin]
}

func (gda *GeneralDynamicArray) Extend(n int) {
	*gda = append(*gda, make([]interface{}, n)...)
}

func (gda *GeneralDynamicArray) Expand(i, n int) {
	if i == len(*gda) {
		gda.Extend(n)
		return
	}
	_ = (*gda)[i] // ensure i is valid
	*gda = append(*gda, make([]interface{}, n)...)
	copy((*gda)[i+n:], (*gda)[i:])
	for k := i; k < i+n; k++ {
		(*gda)[k] = nil
	}
}

func (gda *GeneralDynamicArray) Reserve(capacity int) {
	if capacity <= 0 || (gda != nil && capacity <= cap(*gda)) {
		return
	}
	s := make(GeneralDynamicArray, len(*gda), capacity)
	copy(s, *gda)
	*gda = s
}

func (gda *GeneralDynamicArray) Shrink() {
	if gda == nil || len(*gda) == cap(*gda) {
		return
	}
	s := make(GeneralDynamicArray, len(*gda))
	copy(s, *gda)
	*gda = s
}

func (gda *GeneralDynamicArray) Clear() {
	if gda != nil {
		*gda = nil
	}
}

func (gda *GeneralDynamicArray) Filter(filter func(x interface{}) (keep bool)) {
	if gda == nil || len(*gda) == 0 {
		return
	}
	n := 0
	for _, x := range *gda {
		if filter(x) {
			(*gda)[n] = x
			n++
		}
	}
	if n == len(*gda) {
		return
	}
	for i := n; i < len(*gda); i++ {
		(*gda)[i] = nil // avoid memory leak
	}
	*gda = (*gda)[:n]
}
