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

import (
	"reflect"

	"github.com/donyori/gogo/errors"
)

// SliceDynamicArray is an adapter for: go slice -> DynamicArray.
// It is based on package reflect. If you concern the performance,
// consider to use GeneralDynamicArray, IntDynamicArray,
// Float64DynamicArray or StringDynamicArray if possible.
type SliceDynamicArray struct {
	p reflect.Value
	v reflect.Value
}

// WrapSlice makes a SliceDynamicArray by given slicePtr: a pointer to a slice.
// It panics if slicePtr isn't a pointer to a slice.
func WrapSlice(slicePtr interface{}) *SliceDynamicArray {
	if slicePtr == nil {
		panic(errors.AutoMsg("slicePtr is nil"))
	}
	sda := new(SliceDynamicArray)
	sda.p = reflect.ValueOf(slicePtr)
	if sda.p.Kind() != reflect.Ptr {
		panic(errors.AutoMsg("slicePtr is NOT a pointer"))
	}
	sda.v = sda.p.Elem()
	for sda.v.Kind() == reflect.Ptr || sda.v.Kind() == reflect.Interface {
		sda.v = sda.v.Elem()
	}
	if sda.v.Kind() != reflect.Slice {
		panic(errors.AutoMsg("slicePtr does NOT point to a slice"))
	}
	return sda
}

// NewSliceDynamicArray makes a SliceDynamicArray with given
// itemType, length and capacity.
// The underlying slice will be:
//  make([]itemType, length, capacity)
// It panics if itemType is nil, or length or capacity is illegal.
func NewSliceDynamicArray(itemType reflect.Type, length int, capacity int) *SliceDynamicArray {
	if itemType == nil {
		panic(errors.AutoMsg("itemType is nil"))
	}
	s := reflect.MakeSlice(reflect.SliceOf(itemType), length, capacity).Interface()
	sda := new(SliceDynamicArray)
	sda.p = reflect.ValueOf(&s)
	sda.v = sda.p.Elem().Elem()
	return sda
}

// RetrieveSlicePtr returns the pointer to the slice.
// It returns nil if the receiver is nil.
func (sda *SliceDynamicArray) RetrieveSlicePtr() interface{} {
	if sda == nil {
		return nil
	}
	return sda.p.Interface()
}

// RetrieveSlice returns the slice.
// It returns nil if the receiver is nil.
func (sda *SliceDynamicArray) RetrieveSlice() interface{} {
	if sda == nil {
		return nil
	}
	return sda.v.Interface()
}

// Len returns the number of items in the array.
func (sda *SliceDynamicArray) Len() int {
	if sda == nil {
		return 0
	}
	return sda.v.Len()
}

// Front returns the first item of the array.
// It panics if the array is nil or empty.
func (sda *SliceDynamicArray) Front() interface{} {
	return sda.v.Index(0).Interface()
}

// SetFront sets the first item to x.
// It panics if the array is nil or empty.
// If x is nil, a zero value of the array item type will be used.
func (sda *SliceDynamicArray) SetFront(x interface{}) {
	sda.v.Index(0).Set(sda.valueOf(x))
}

// Back returns the last item of the array.
// It panics if the array is nil or empty.
func (sda *SliceDynamicArray) Back() interface{} {
	return sda.v.Index(sda.v.Len() - 1).Interface()
}

// SetBack sets the last item to x.
// It panics if the array is nil or empty.
// If x is nil, a zero value of the array item type will be used.
func (sda *SliceDynamicArray) SetBack(x interface{}) {
	sda.v.Index(sda.v.Len() - 1).Set(sda.valueOf(x))
}

// Reverse turns the other way round items of the array.
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

// Scan browses the items in the array from the first to the last.
//
// Its argument handler is a function to deal with the item x in the
// array and report whether to continue to check the next item or not.
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

// Get returns the i-th item of the array.
// It panics if i is out of range.
func (sda *SliceDynamicArray) Get(i int) interface{} {
	return sda.v.Index(i).Interface()
}

// Set sets the i-th item to x.
// It panics if i is out of range.
// If x is nil, a zero value of the array item type will be used.
func (sda *SliceDynamicArray) Set(i int, x interface{}) {
	sda.v.Index(i).Set(sda.valueOf(x))
}

// Swap exchanges the i-th and j-th items.
// It panics if i or j is out of range.
func (sda *SliceDynamicArray) Swap(i, j int) {
	reflect.Swapper(sda.v.Interface())(i, j)
}

// Slice returns a slice from argument begin (inclusive) to
// argument end (exclusive) of the array, as an Array.
// It panics if begin or end is out of range, or begin > end.
func (sda *SliceDynamicArray) Slice(begin, end int) Array {
	p := reflect.New(sda.v.Type())
	p.Elem().Set(sda.v.Slice3(begin, end, end))
	return &SliceDynamicArray{p: p, v: p.Elem()}
}

// Cap returns the current capacity of the dynamic array.
func (sda *SliceDynamicArray) Cap() int {
	if sda == nil {
		return 0
	}
	return sda.v.Cap()
}

// Push adds x to the back of the dynamic array.
// If x is nil, a zero value of the array item type will be used.
func (sda *SliceDynamicArray) Push(x interface{}) {
	sda.v.Set(reflect.Append(sda.v, sda.valueOf(x)))
}

// Pop removes and returns the last item of the dynamic array.
// It panics if the dynamic array is nil or empty.
func (sda *SliceDynamicArray) Pop() interface{} {
	back := sda.v.Len() - 1
	backV := sda.v.Index(back)
	x := backV.Interface()
	backV.Set(reflect.Zero(backV.Type())) // avoid memory leak
	sda.v.Set(sda.v.Slice(0, back))
	return x
}

// Append adds s to the back of the dynamic array.
// s shouldn't be modified during calling this method,
// otherwise, unknown error may occur.
func (sda *SliceDynamicArray) Append(s Sequence) {
	if s == nil {
		return
	}
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
		sda.v.Index(i).Set(sda.valueOf(x))
		i++
		return true
	})
}

// Truncate removes the i-th and all subsequent items in the dynamic array.
// It does nothing if i is out of range.
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

// Insert adds x as the i-th item in the dynamic array.
// It panics if i is out of range, i.e., i < 0 or i > Len().
// If x is nil, a zero value of the array item type will be used.
func (sda *SliceDynamicArray) Insert(i int, x interface{}) {
	if i == sda.v.Len() {
		sda.Push(x)
		return
	}
	sda.v.Index(i) // ensure i is valid
	sda.v.Set(reflect.Append(sda.v, reflect.Zero(sda.v.Type().Elem())))
	end := sda.v.Len()
	reflect.Copy(sda.v.Slice(i+1, end), sda.v.Slice(i, end))
	sda.v.Index(i).Set(sda.valueOf(x))
}

// Remove removes and returns the i-th item in the dynamic array.
// It panics if i is out of range, i.e., i < 0 or i >= Len().
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

// RemoveWithoutOrder removes and returns the i-th item in
// the dynamic array, without preserving order.
// It panics if i is out of range, i.e., i < 0 or i >= Len().
func (sda *SliceDynamicArray) RemoveWithoutOrder(i int) interface{} {
	iV := sda.v.Index(i)
	x := iV.Interface()
	back := sda.v.Len() - 1
	backV := sda.v.Index(back)
	if i != back {
		iV.Set(backV)
	}
	backV.Set(reflect.Zero(backV.Type()))
	sda.v.Set(sda.v.Slice(0, back))
	return x
}

// InsertSequence inserts s to the front of the i-th item
// in the dynamic array.
// It panics if i is out of range, i.e., i < 0 or i > Len().
// s shouldn't be modified during calling this method,
// otherwise, unknown error may occur.
func (sda *SliceDynamicArray) InsertSequence(i int, s Sequence) {
	if i == sda.v.Len() {
		sda.Append(s)
		return
	}
	sda.v.Index(i) // ensure i is valid
	if s == nil {
		return
	}
	n := s.Len()
	if n == 0 {
		return
	}
	sda.v.Set(reflect.AppendSlice(sda.v, reflect.MakeSlice(sda.v.Type(), n, n)))
	end := sda.v.Len()
	reflect.Copy(sda.v.Slice(i+n, end), sda.v.Slice(i, end))
	k := i
	s.Scan(func(x interface{}) (cont bool) {
		sda.v.Index(k).Set(sda.valueOf(x))
		k++
		return true
	})
}

// Cut removes items from argument begin (inclusive) to
// argument end (exclusive) of the dynamic array.
// It panics if begin or end is out of range, or begin > end.
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

// CutWithoutOrder removes items from argument begin (inclusive) to
// argument end (exclusive) of the dynamic array, without preserving order.
// It panics if begin or end is out of range, or begin > end.
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

// Extend adds n zero-value items to the back of the dynamic array.
// It panics if n < 0.
func (sda *SliceDynamicArray) Extend(n int) {
	sda.v.Set(reflect.AppendSlice(sda.v, reflect.MakeSlice(sda.v.Type(), n, n)))
}

// Expand inserts n zero-value items to the front of the i-th item
// in the dynamic array.
// It panics if i is out of range, i.e., i < 0 or i > Len(), or n < 0.
func (sda *SliceDynamicArray) Expand(i, n int) {
	if i == sda.v.Len() {
		sda.Extend(n)
		return
	}
	sda.v.Index(i) // ensure i is valid
	sda.v.Set(reflect.AppendSlice(sda.v, reflect.MakeSlice(sda.v.Type(), n, n)))
	end := sda.v.Len()
	reflect.Copy(sda.v.Slice(i+n, end), sda.v.Slice(i, end))
	zero := reflect.Zero(sda.v.Type().Elem())
	for k := i; k < i+n; k++ {
		sda.v.Index(k).Set(zero)
	}
}

// Reserve requests that the capacity of the dynamic array
// is at least the given capacity.
// It does nothing if capacity <= Cap().
func (sda *SliceDynamicArray) Reserve(capacity int) {
	if capacity <= sda.Cap() {
		return
	}
	v := reflect.MakeSlice(sda.v.Type(), sda.v.Len(), capacity)
	reflect.Copy(v, sda.v)
	sda.v.Set(v)
}

// Shrink reduces the dynamic array to fit, i.e.,
// requests Cap() equals to Len().
// Note that it isn't equivalent to operations on slice
// like s[:len(s):len(s)],
// because it will allocate a new array and copy the content
// if Cap() > Len().
func (sda *SliceDynamicArray) Shrink() {
	if sda == nil || sda.v.Len() == sda.v.Cap() {
		return
	}
	v := reflect.MakeSlice(sda.v.Type(), sda.v.Len(), sda.v.Len())
	reflect.Copy(v, sda.v)
	sda.v.Set(v)
}

// Clear removes all items in the dynamic array and
// asks to release the memory.
func (sda *SliceDynamicArray) Clear() {
	if sda != nil {
		sda.v.Set(reflect.Zero(sda.v.Type()))
	}
}

// Filter refines items in the dynamic array (in place).
//
// Its argument filter is a function to report
// whether to keep the item x or not.
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
	if n == sda.v.Len() {
		return
	}
	zero := reflect.Zero(sda.v.Type().Elem())
	for i := n; i < sda.v.Len(); i++ {
		sda.v.Index(i).Set(zero) // avoid memory leak
	}
	sda.v.Set(sda.v.Slice(0, n))
}

// valueOf returns the reflect.Value of x.
// If x is not nil, it returns reflect.ValueOf(x).
// Otherwise, it returns a zero value of the array item type.
// It panics if x is not assignable to the item of the dynamic array.
func (sda *SliceDynamicArray) valueOf(x interface{}) reflect.Value {
	itemType := sda.v.Type().Elem()
	if x == nil {
		return reflect.Zero(itemType)
	}
	xV := reflect.ValueOf(x)
	if !xV.Type().AssignableTo(itemType) {
		panic(errors.AutoMsg("x is not assignable to the item of the dynamic array"))
	}
	return xV
}
