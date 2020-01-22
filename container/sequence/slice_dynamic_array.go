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

import "reflect"

// An adapter for: go slice -> DynamicArray.
// It is based on package reflect. If you concern the performance,
// consider to use GeneralDynamicArray, IntDynamicArray,
// Float64DynamicArray or StringDynamicArray if possible.
type SliceDynamicArray struct {
	p reflect.Value
	v reflect.Value
}

// Make a SliceDynamicArray by given slicePtr: a pointer to a slice.
// It panics if slicePtr isn't a pointer to a slice.
func NewSliceDynamicArray(slicePtr interface{}) *SliceDynamicArray {
	if slicePtr == nil {
		panic("slicePtr is nil")
	}
	p := reflect.ValueOf(slicePtr)
	if p.Kind() != reflect.Ptr {
		panic("slicePtr is NOT a pointer")
	}
	v := p.Elem()
	for v.Kind() == reflect.Ptr {
		p = v
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice {
		panic("slicePtr does NOT point to a slice")
	}
	return &SliceDynamicArray{p: p, v: v}
}

// Return the pointer to the slice.
// It returns nil if the receiver is nil.
func (sda *SliceDynamicArray) RetrieveSlicePtr() interface{} {
	if sda == nil {
		return nil
	}
	return sda.p.Interface()
}

// Return the slice.
// It returns nil if the receiver is nil.
func (sda *SliceDynamicArray) RetrieveSlice() interface{} {
	if sda == nil {
		return nil
	}
	return sda.v.Interface()
}

func (sda *SliceDynamicArray) Len() int {
	if sda == nil {
		return 0
	}
	return sda.v.Len()
}

func (sda *SliceDynamicArray) Front() interface{} {
	return sda.v.Index(0).Interface()
}

func (sda *SliceDynamicArray) SetFront(x interface{}) {
	sda.v.Index(0).Set(reflect.ValueOf(x))
}

func (sda *SliceDynamicArray) Back() interface{} {
	return sda.v.Index(sda.v.Len() - 1).Interface()
}

func (sda *SliceDynamicArray) SetBack(x interface{}) {
	sda.v.Index(sda.v.Len() - 1).Set(reflect.ValueOf(x))
}

func (sda *SliceDynamicArray) Reverse() {
	if sda == nil {
		return
	}
	k := sda.v.Len() - 1
	if k <= 0 {
		return
	}
	swapper := reflect.Swapper(sda.v.Interface())
	for i := 0; i < k; i, k = i+1, k-1 {
		swapper(i, k)
	}
}

func (sda *SliceDynamicArray) Scan(handler func(x interface{}) (cont bool)) {
	if sda == nil {
		return
	}
	for i := 0; i < sda.v.Len(); i++ {
		if !handler(sda.v.Index(i).Interface()) {
			return
		}
	}
}

func (sda *SliceDynamicArray) Get(i int) interface{} {
	return sda.v.Index(i).Interface()
}

func (sda *SliceDynamicArray) Set(i int, x interface{}) {
	sda.v.Index(i).Set(reflect.ValueOf(x))
}

func (sda *SliceDynamicArray) Swap(i, j int) {
	reflect.Swapper(sda.v.Interface())(i, j)
}

func (sda *SliceDynamicArray) Slice(begin, end int) Array {
	v := sda.v.Slice3(begin, end, end)
	p := v.Addr()
	return &SliceDynamicArray{p: p, v: v}
}

func (sda *SliceDynamicArray) Cap() int {
	if sda == nil {
		return 0
	}
	return sda.v.Cap()
}

func (sda *SliceDynamicArray) Push(x interface{}) {
	sda.v.Set(reflect.Append(sda.v, reflect.ValueOf(x)))
}

func (sda *SliceDynamicArray) Pop() interface{} {
	back := sda.v.Len() - 1
	backV := sda.v.Index(back)
	x := backV.Interface()
	backV.Set(reflect.Zero(backV.Type())) // avoid memory leak
	sda.v.Set(sda.v.Slice(0, back))
	return x
}

func (sda *SliceDynamicArray) Append(s Sequence) {
	n := s.Len()
	if n == 0 {
		return
	}
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Slice {
		sda.v.Set(reflect.AppendSlice(sda.v, v))
		return
	}
	i := sda.v.Len()
	sda.v.Set(reflect.AppendSlice(sda.v, reflect.MakeSlice(sda.v.Type(), n, n)))
	s.Scan(func(x interface{}) (cont bool) {
		sda.v.Index(i).Set(reflect.ValueOf(x))
		i++
		return true
	})
}

func (sda *SliceDynamicArray) Truncate(i int) {
	if sda == nil {
		return
	}
	n := sda.v.Len()
	if i < 0 || i >= n {
		return
	}
	zero := reflect.Zero(sda.v.Type().Elem())
	for k := i; k < n; k++ {
		sda.v.Index(k).Set(zero) // avoid memory leak
	}
	sda.v.Set(sda.v.Slice(0, i))
}

func (sda *SliceDynamicArray) Insert(i int, x interface{}) {
	if i == sda.v.Len() {
		sda.Push(x)
		return
	}
	sda.v.Index(i) // ensure i is valid
	sda.v.Set(reflect.Append(sda.v, reflect.Zero(sda.v.Type().Elem())))
	end := sda.v.Len()
	reflect.Copy(sda.v.Slice(i+1, end), sda.v.Slice(i, end))
	sda.v.Index(i).Set(reflect.ValueOf(x))
}

func (sda *SliceDynamicArray) Remove(i int) interface{} {
	back := sda.v.Len() - 1
	if i == back {
		return sda.Pop()
	}
	x := sda.v.Index(i).Interface()
	reflect.Copy(sda.v.Slice(i, back), sda.v.Slice(i+1, back+1))
	backV := sda.v.Index(back)
	backV.Set(reflect.Zero(backV.Type())) // avoid memory leak
	sda.v.Set(sda.v.Slice(0, back))
	return x
}

func (sda *SliceDynamicArray) RemoveWithoutOrder(i int) interface{} {
	iV := sda.v.Index(i)
	x := iV.Interface()
	back := sda.v.Len()
	backV := sda.v.Index(back)
	if i != back {
		iV.Set(backV)
	}
	backV.Set(reflect.Zero(backV.Type()))
	sda.v.Set(sda.v.Slice(0, back))
	return x
}

func (sda *SliceDynamicArray) InsertSequence(i int, s Sequence) {
	if i == sda.v.Len() {
		sda.Append(s)
		return
	}
	sda.v.Index(i) // ensure i is valid
	n := s.Len()
	if n == 0 {
		return
	}
	sda.v.Set(reflect.AppendSlice(sda.v, reflect.MakeSlice(sda.v.Type(), n, n)))
	end := sda.v.Len()
	reflect.Copy(sda.v.Slice(i+n, end), sda.v.Slice(i, end))
	k := i
	s.Scan(func(x interface{}) (cont bool) {
		sda.v.Index(k).Set(reflect.ValueOf(x))
		k++
		return true
	})
}

func (sda *SliceDynamicArray) Cut(begin, end int) {
	sda.v.Slice(begin, end) // ensure begin and end are valid
	if begin == end {
		return
	}
	n := sda.v.Len()
	if end == n {
		sda.Truncate(begin)
		return
	}
	reflect.Copy(sda.v.Slice(begin, n), sda.v.Slice(end, n))
	zero := reflect.Zero(sda.v.Type().Elem())
	for i := n - end + begin; i < n; i++ {
		sda.v.Index(i).Set(zero) // avoid memory leak
	}
	sda.v.Set(sda.v.Slice(0, n-end+begin))
}

func (sda *SliceDynamicArray) CutWithoutOrder(begin, end int) {
	sda.v.Slice(begin, end) // ensure begin and end are valid
	if begin == end {
		return
	}
	n := sda.v.Len()
	if end == n {
		sda.Truncate(begin)
		return
	}
	copyIdx := n - end + begin
	if copyIdx < end {
		copyIdx = end
	}
	reflect.Copy(sda.v.Slice(begin, n), sda.v.Slice(copyIdx, n))
	zero := reflect.Zero(sda.v.Type().Elem())
	for i := n - end + begin; i < n; i++ {
		sda.v.Index(i).Set(zero)
	}
	sda.v.Set(sda.v.Slice(0, n-end+begin))
}

func (sda *SliceDynamicArray) Extend(n int) {
	sda.v.Set(reflect.AppendSlice(sda.v, reflect.MakeSlice(sda.v.Type(), n, n)))
}

func (sda *SliceDynamicArray) Expand(i, n int) {
	if i == sda.v.Len() {
		sda.Extend(n)
		return
	}
	sda.v.Index(i) // ensure i is valid
	sda.v.Set(reflect.AppendSlice(sda.v, reflect.MakeSlice(sda.v.Type(), n, n)))
	end := sda.v.Len()
	reflect.Copy(sda.v.Slice(i+n, end), sda.v.Slice(i, n))
	zero := reflect.Zero(sda.v.Type().Elem())
	for k := i; k < i+n; k++ {
		sda.v.Index(k).Set(zero)
	}
}

func (sda *SliceDynamicArray) Reserve(capacity int) {
	if capacity <= sda.Cap() {
		return
	}
	v := reflect.MakeSlice(sda.v.Type(), sda.v.Len(), capacity)
	reflect.Copy(v, sda.v)
	sda.v.Set(v)
}

func (sda *SliceDynamicArray) Shrink() {
	if sda.Len() == sda.Cap() {
		return
	}
	v := reflect.MakeSlice(sda.v.Type(), sda.v.Len(), sda.v.Len())
	reflect.Copy(v, sda.v)
	sda.v.Set(v)
}

func (sda *SliceDynamicArray) Clear() {
	if sda != nil {
		sda.v.Set(reflect.Zero(sda.v.Type()))
	}
}

func (sda *SliceDynamicArray) Filter(filter func(x interface{}) (keep bool)) {
	if sda.Len() == 0 {
		return
	}
	n := 0
	for i := 0; i < sda.v.Len(); i++ {
		xV := sda.v.Index(i)
		if filter(xV.Interface()) {
			sda.v.Index(n).Set(xV)
			n++
		}
	}
	end := sda.v.Len()
	if n == end {
		return
	}
	zero := reflect.Zero(sda.v.Type().Elem())
	for i := n; i < end; i++ {
		sda.v.Index(i).Set(zero) // avoid memory leak
	}
	sda.v.Set(sda.v.Slice(0, n))
}
