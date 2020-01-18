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
		return i < len(*gda)
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
	if i == len(*gda)-1 {
		return gda.Pop()
	}
	x := (*gda)[i]
	back := len(*gda) - 1
	if i < back {
		copy((*gda)[i:], (*gda)[i+1:])
	}
	(*gda)[back] = nil // avoid memory leak
	*gda = (*gda)[:back]
	return x
}

func (gda *GeneralDynamicArray) RemoveWithoutOrder(i int) interface{} {
	x := (*gda)[i]
	back := len(*gda) - 1
	(*gda)[i] = (*gda)[back]
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
		return k < i+n
	})
}

func (gda *GeneralDynamicArray) Cut(begin, end int) {
	if end == len(*gda) {
		gda.Truncate(begin)
		return
	}
	_ = (*gda)[begin:end] // ensure begin and end are valid
	copy((*gda)[begin:], (*gda)[end:])
	for i := len(*gda) - end + begin; i < len(*gda); i++ {
		(*gda)[i] = nil // avoid memory leak
	}
	*gda = (*gda)[:len(*gda)-end+begin]
}

func (gda *GeneralDynamicArray) CutWithoutOrder(begin, end int) {
	if end == len(*gda) {
		gda.Truncate(begin)
		return
	}
	_ = (*gda)[begin:end] // ensure begin and end are valid
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
	if capacity <= cap(*gda) {
		return
	}
	s := make(GeneralDynamicArray, len(*gda), capacity)
	copy(s, *gda)
	*gda = s
}

func (gda *GeneralDynamicArray) Shrink() {
	if len(*gda) == cap(*gda) {
		return
	}
	s := make(GeneralDynamicArray, len(*gda))
	copy(s, *gda)
	*gda = s
}

func (gda *GeneralDynamicArray) Clear() {
	*gda = nil
}

func (gda *GeneralDynamicArray) Filter(filter func(x interface{}) (keep bool)) {
	if len(*gda) == 0 {
		return
	}
	n := 0
	for _, x := range *gda {
		if filter(x) {
			(*gda)[n] = x
			n++
		}
	}
	for i := n; i < len(*gda); i++ {
		(*gda)[i] = nil // avoid memory leak
	}
	*gda = (*gda)[:n]
}

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
		return i < len(*ida)
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
	if i == len(*ida)-1 {
		return ida.Pop()
	}
	x := (*ida)[i]
	back := len(*ida) - 1
	if i < back {
		copy((*ida)[i:], (*ida)[i+1:])
	}
	*ida = (*ida)[:back]
	return x
}

func (ida *IntDynamicArray) RemoveWithoutOrder(i int) interface{} {
	x := (*ida)[i]
	back := len(*ida) - 1
	(*ida)[i] = (*ida)[back]
	*ida = (*ida)[:back]
	return x
}

func (ida *IntDynamicArray) InsertSequence(i int, s Sequence) {
	if i == len(*ida) {
		ida.Append(s)
		return
	}
	_ = (*ida)[i] // ensure i is valid
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
		return k < i+n
	})
}

func (ida *IntDynamicArray) Cut(begin, end int) {
	if end == len(*ida) {
		ida.Truncate(begin)
		return
	}
	_ = (*ida)[begin:end] // ensure begin and end are valid
	copy((*ida)[begin:], (*ida)[end:])
	*ida = (*ida)[:len(*ida)-end+begin]
}

func (ida *IntDynamicArray) CutWithoutOrder(begin, end int) {
	if end == len(*ida) {
		ida.Truncate(begin)
		return
	}
	_ = (*ida)[begin:end] // ensure begin and end are valid
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
	if capacity <= cap(*ida) {
		return
	}
	s := make(IntDynamicArray, len(*ida), capacity)
	copy(s, *ida)
	*ida = s
}

func (ida *IntDynamicArray) Shrink() {
	if len(*ida) == cap(*ida) {
		return
	}
	s := make(IntDynamicArray, len(*ida))
	copy(s, *ida)
	*ida = s
}

func (ida *IntDynamicArray) Clear() {
	*ida = nil
}

func (ida *IntDynamicArray) Filter(filter func(x interface{}) (keep bool)) {
	if len(*ida) == 0 {
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
		return i < len(*fda)
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
	if i == len(*fda)-1 {
		return fda.Pop()
	}
	x := (*fda)[i]
	back := len(*fda) - 1
	if i < back {
		copy((*fda)[i:], (*fda)[i+1:])
	}
	*fda = (*fda)[:back]
	return x
}

func (fda *Float64DynamicArray) RemoveWithoutOrder(i int) interface{} {
	x := (*fda)[i]
	back := len(*fda) - 1
	(*fda)[i] = (*fda)[back]
	*fda = (*fda)[:back]
	return x
}

func (fda *Float64DynamicArray) InsertSequence(i int, s Sequence) {
	if i == len(*fda) {
		fda.Append(s)
		return
	}
	_ = (*fda)[i] // ensure i is valid
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
		return k < i+n
	})
}

func (fda *Float64DynamicArray) Cut(begin, end int) {
	if end == len(*fda) {
		fda.Truncate(begin)
		return
	}
	_ = (*fda)[begin:end] // ensure begin and end are valid
	copy((*fda)[begin:], (*fda)[end:])
	*fda = (*fda)[:len(*fda)-end+begin]
}

func (fda *Float64DynamicArray) CutWithoutOrder(begin, end int) {
	if end == len(*fda) {
		fda.Truncate(begin)
		return
	}
	_ = (*fda)[begin:end] // ensure begin and end are valid
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
	if capacity <= cap(*fda) {
		return
	}
	s := make(Float64DynamicArray, len(*fda), capacity)
	copy(s, *fda)
	*fda = s
}

func (fda *Float64DynamicArray) Shrink() {
	if len(*fda) == cap(*fda) {
		return
	}
	s := make(Float64DynamicArray, len(*fda))
	copy(s, *fda)
	*fda = s
}

func (fda *Float64DynamicArray) Clear() {
	*fda = nil
}

func (fda *Float64DynamicArray) Filter(filter func(x interface{}) (keep bool)) {
	if len(*fda) == 0 {
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
		return i < len(*sda)
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
	if i == len(*sda)-1 {
		return sda.Pop()
	}
	x := (*sda)[i]
	back := len(*sda) - 1
	if i < back {
		copy((*sda)[i:], (*sda)[i+1:])
	}
	(*sda)[back] = "" // avoid memory leak
	*sda = (*sda)[:back]
	return x
}

func (sda *StringDynamicArray) RemoveWithoutOrder(i int) interface{} {
	x := (*sda)[i]
	back := len(*sda) - 1
	(*sda)[i] = (*sda)[back]
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
		return k < i+n
	})
}

func (sda *StringDynamicArray) Cut(begin, end int) {
	if end == len(*sda) {
		sda.Truncate(begin)
		return
	}
	_ = (*sda)[begin:end] // ensure begin and end are valid
	copy((*sda)[begin:], (*sda)[end:])
	for i := len(*sda) - end + begin; i < len(*sda); i++ {
		(*sda)[i] = "" // avoid memory leak
	}
	*sda = (*sda)[:len(*sda)-end+begin]
}

func (sda *StringDynamicArray) CutWithoutOrder(begin, end int) {
	if end == len(*sda) {
		sda.Truncate(begin)
		return
	}
	_ = (*sda)[begin:end] // ensure begin and end are valid
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
	if capacity <= cap(*sda) {
		return
	}
	s := make(StringDynamicArray, len(*sda), capacity)
	copy(s, *sda)
	*sda = s
}

func (sda *StringDynamicArray) Shrink() {
	if len(*sda) == cap(*sda) {
		return
	}
	s := make(StringDynamicArray, len(*sda))
	copy(s, *sda)
	*sda = s
}

func (sda *StringDynamicArray) Clear() {
	*sda = nil
}

func (sda *StringDynamicArray) Filter(filter func(x interface{}) (keep bool)) {
	if len(*sda) == 0 {
		return
	}
	n := 0
	for _, x := range *sda {
		if filter(x) {
			(*sda)[n] = x
			n++
		}
	}
	for i := n; i < len(*sda); i++ {
		(*sda)[i] = "" // avoid memory leak
	}
	*sda = (*sda)[:n]
}
