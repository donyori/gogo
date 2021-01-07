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

import (
	"reflect"
	"testing"
)

func TestNewSliceDynamicArray(t *testing.T) {
	s := []interface{}{1, 2, 3}
	ptr := &s
	sda := WrapSlice(ptr)
	if p := sda.p.Interface(); p != ptr {
		t.Errorf("sda.p: %v != &s: %p.", p, ptr)
	}
	if v := sda.v.Interface().([]interface{}); sliceUnequal(v, s) {
		t.Errorf("sda.v: %v != s: %v.", v, s)
	}
	itf := interface{}(s)
	sda = WrapSlice(&itf)
	if p := sda.p.Interface(); p != &itf {
		t.Errorf("sda.p: %v != &s: %p.", p, &itf)
	}
	if v := sda.v.Interface().([]interface{}); sliceUnequal(v, s) {
		t.Errorf("sda.v: %v != s: %v.", v, s)
	}
	sda = WrapSlice(&ptr)
	if p := sda.p.Interface(); p != &ptr {
		t.Errorf("sda.p: %v != &s: %p.", p, &ptr)
	}
	if v := sda.v.Interface().([]interface{}); sliceUnequal(v, s) {
		t.Errorf("sda.v: %v != s: %v.", v, s)
	}
}

func TestMakeSliceDynamicArray(t *testing.T) {
	var i int
	sda := NewSliceDynamicArray(reflect.TypeOf(i), 2, 5)
	wanted := make([]int, 2, 5)
	if v := sda.v.Interface(); !reflect.DeepEqual(v, wanted) {
		t.Errorf("sda.v: %v, wanted: %v.", v, wanted)
	}
	if c := sda.v.Cap(); c != cap(wanted) {
		t.Errorf("sda.v.Cap(): %d != cap(wanted): %d.", c, cap(wanted))
	}
}

func TestSliceDynamicArray_RetrieveSlicePtr(t *testing.T) {
	var sda *SliceDynamicArray
	slicePtr := sda.RetrieveSlicePtr()
	if slicePtr != nil {
		t.Errorf("sda.RetrieveSlicePtr(): %v != nil when sda == nil.", slicePtr)
	}
	s := []interface{}{1, 2, 3}
	sda = WrapSlice(&s)
	slicePtr = sda.RetrieveSlicePtr()
	if slicePtr != &s {
		t.Errorf("sda.RetrieveSlicePtr(): %v != &s: %p.", slicePtr, &s)
	}
}

func TestSliceDynamicArray_RetrieveSlice(t *testing.T) {
	var sda *SliceDynamicArray
	slice := sda.RetrieveSlice()
	if slice != nil {
		t.Errorf("sda.RetrieveSlice(): %v != nil when sda == nil.", slice)
	}
	s := []interface{}{1, 2, 3}
	sda = WrapSlice(&s)
	slice = sda.RetrieveSlice()
	if sliceUnequal(slice.([]interface{}), s) {
		t.Errorf("sda.RetrieveSlice(): %v != s: %v.", slice, s)
	}
}

func TestSliceDynamicArray_Len(t *testing.T) {
	var sda *SliceDynamicArray
	if n := sda.Len(); n != 0 {
		t.Errorf("sda.Len(): %d != 0.", n)
	}
	var s []interface{}
	sda = WrapSlice(&s)
	if n := sda.Len(); n != len(s) {
		t.Errorf("sda.Len(): %d != len(s): %d.", n, len(s))
	}
	s = []interface{}{1}
	sda = WrapSlice(&s)
	if n := sda.Len(); n != len(s) {
		t.Errorf("sda.Len(): %d != len(s): %d.", n, len(s))
	}
}

func TestSliceDynamicArray_Reverse(t *testing.T) {
	s := []interface{}{1, 1, 2, 3, 4}
	sda := WrapSlice(&s)
	wanted := []interface{}{4, 3, 2, 1, 1}
	sda.Reverse()
	if sliceUnequal(s, wanted) {
		t.Errorf("After reverse: %v, wanted: %v.", s, wanted)
	}
}

func TestSliceDynamicArray_Scan(t *testing.T) {
	s := []interface{}{1, 1, 2, 3}
	sda := WrapSlice(&s)
	s2 := make([]interface{}, 0, len(s))
	sda.Scan(func(x interface{}) (cont bool) {
		s2 = append(s2, x)
		return len(s2) < 3
	})
	if sliceUnequal(s2, s[:3]) {
		t.Errorf("s2: %v, wanted: %v.", s2, s[:3])
	}
}

func TestSliceDynamicArray_Get(t *testing.T) {
	s := []interface{}{1, 1, 2, 3}
	sda := WrapSlice(&s)
	for i, x := range s {
		item := sda.Get(i)
		if item != x {
			t.Errorf("sda.Get(%d): %v != s[%[1]d]: %[3]v.", i, item, x)
		}
	}
}

func TestSliceDynamicArray_Set(t *testing.T) {
	s := []interface{}{1, 1, 2, 3}
	sda := WrapSlice(&s)
	wanted := []interface{}{1, 8, nil, 7.5}
	for i := range s {
		sda.Set(i, wanted[i])
		if s[i] != wanted[i] {
			t.Errorf("s[%d]: %v != wanted[%[1]d]: %[3]v.", i, s[i], wanted[i])
		}
	}
}

func TestSliceDynamicArray_Swap(t *testing.T) {
	s := []interface{}{1, 2, 3}
	sda := WrapSlice(&s)
	wanted := []interface{}{2, 1, 3}
	sda.Swap(1, 0)
	if sliceUnequal(s, wanted) {
		t.Errorf("After swap: %v, wanted: %v.", s, wanted)
	}
}

func TestSliceDynamicArray_Slice(t *testing.T) {
	s := []interface{}{1, 2, 3, 4}
	sda := WrapSlice(&s)
	slice := sda.Slice(1, 3)
	if n := slice.Len(); n != 2 {
		t.Errorf("slice.Len(): %d != 2.", n)
	}
	i := 1
	count := 0
	slice.Scan(func(x interface{}) (cont bool) {
		if i < 3 {
			if x != s[i] {
				t.Errorf("slice[%d]: %v != s[1:3][%[1]d]: %[3]v.", count, x, s[i])
			}
		} else {
			t.Errorf("slice[%d]: %v (beyond s[1:3]).", count, x)
		}
		i++
		count++
		return true
	})
	// To test underlying arrays are the same, modify one item and check again:
	s[2] = nil
	i = 1
	count = 0
	slice.Scan(func(x interface{}) (cont bool) {
		if i < 3 {
			if x != s[i] {
				t.Errorf("slice[%d]: %v != s[1:3][%[1]d]: %[3]v.", count, x, s[i])
			}
		} else {
			t.Errorf("slice[%d]: %v (beyond s[1:3]).", count, x)
		}
		i++
		count++
		return true
	})
}

func TestSliceDynamicArray_Cap(t *testing.T) {
	var sda *SliceDynamicArray
	if c := sda.Cap(); c != 0 {
		t.Errorf("sda.Cap(): %d != 0.", c)
	}
	s := make([]interface{}, 0)
	sda = WrapSlice(&s)
	if c := sda.Cap(); c != cap(s) {
		t.Errorf("sda.Cap(): %d != cap(s): %d.", c, cap(s))
	}
	s = make([]interface{}, 2, 10)
	sda = WrapSlice(&s)
	if c := sda.Cap(); c != cap(s) {
		t.Errorf("sda.Cap(): %d != cap(s): %d.", c, cap(s))
	}
}

func TestSliceDynamicArray_Push(t *testing.T) {
	var s []interface{}
	sda := WrapSlice(&s)
	wanted := []interface{}{1}
	sda.Push(1)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 1st push: %v, wanted: %v.", s, wanted)
	}
	wanted = append(wanted, 2)
	sda.Push(2)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 2nd push: %v, wanted: %v.", s, wanted)
	}
	wanted = append(wanted, nil)
	sda.Push(nil)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 3rd push: %v, wanted: %v.", s, wanted)
	}
}

func TestSliceDynamicArray_Pop(t *testing.T) {
	data := []interface{}{1, 2}
	s := data
	sda := WrapSlice(&s)
	wanted := []interface{}{1}
	x := sda.Pop()
	if sliceUnequal(s, wanted) || x.(int) != 2 {
		t.Errorf("After 1st pop: %v, wanted: %v, x = %v.", s, wanted, x)
	}
	wanted = wanted[:0]
	x = sda.Pop()
	if sliceUnequal(s, wanted) || x.(int) != 1 {
		t.Errorf("After 2nd pop: %v, wanted %v, x = %v.", s, wanted, x)
	}
	if testNonNilItem(data) {
		t.Errorf("After pop all, underlying data have non-nil item: %v.", data)
	}
}

func TestSliceDynamicArray_Append(t *testing.T) {
	var s []interface{}
	sda := WrapSlice(&s)
	seq := GeneralDynamicArray{1, 2, 3}
	wanted := []interface{}{1, 2, 3}
	sda.Append(seq)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 1st append: %v, wanted: %v.", s, wanted)
	}
	wanted = append(wanted, seq...)
	sda.Append(seq)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 2nd append: %v, wanted: %v.", s, wanted)
	}
	wanted = append(wanted, 1)
	sda.Append(GeneralDynamicArray{1})
	if sliceUnequal(s, wanted) {
		t.Errorf("After 3rd append: %v, wanted: %v.", s, wanted)
	}
	sda.Append(nil)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 4th append: %v, wanted: %v.", s, wanted)
	}
}

func TestSliceDynamicArray_Truncate(t *testing.T) {
	data := []interface{}{1, 2, 2, 3, 4, 4}
	s := data
	sda := WrapSlice(&s)
	wanted := []interface{}{1, 2, 2, 3}
	sda.Truncate(4)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 1st truncate: %v, wanted: %v.", s, wanted)
	}
	wanted = wanted[:3]
	sda.Truncate(3)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 2nd truncate: %v, wanted: %v.", s, wanted)
	}
	wanted = wanted[:0]
	sda.Truncate(0)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 3rd truncate: %v, wanted: %v.", s, wanted)
	}
	if testNonNilItem(data) {
		t.Errorf("After truncate all, underlying data have non-nil item: %v.", data)
	}
}

func TestSliceDynamicArray_Insert(t *testing.T) {
	var s []interface{}
	sda := WrapSlice(&s)
	wanted := []interface{}{1}
	sda.Insert(0, 1)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 1st insert: %v, wanted: %v.", s, wanted)
	}
	wanted = append(wanted, 2)
	sda.Insert(1, 2)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 2nd insert: %v, wanted: %v.", s, wanted)
	}
	wanted = []interface{}{0, 1, 2}
	sda.Insert(0, 0)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 3rd insert: %v, wanted: %v.", s, wanted)
	}
	wanted = []interface{}{0, nil, 1, 2}
	sda.Insert(1, nil)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 4th insert: %v, wanted: %v.", s, wanted)
	}
}

func TestSliceDynamicArray_Remove(t *testing.T) {
	data := []interface{}{1, 2, 2, 4}
	s := data
	sda := WrapSlice(&s)
	wanted := []interface{}{1, 2, 4}
	sda.Remove(2)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 1st remove: %v, wanted: %v.", s, wanted)
	}
	wanted = wanted[:2]
	sda.Remove(2)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 2nd remove: %v, wanted: %v.", s, wanted)
	}
	wanted = wanted[1:]
	sda.Remove(0)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 3rd remove: %v, wanted: %v.", s, wanted)
	}
	sda.Remove(0)
	wanted = wanted[:0]
	if sliceUnequal(s, wanted) {
		t.Errorf("After 4th remove: %v, wanted: %v.", s, wanted)
	}
	if testNonNilItem(data) {
		t.Errorf("After remove all, underlying data have non-nil item: %v.", data)
	}
}

func TestSliceDynamicArray_RemoveWithoutOrder(t *testing.T) {
	data := []interface{}{1, 2, 2, 4}
	s := data
	sda := WrapSlice(&s)
	wanted := []interface{}{1, 2, 4}
	sda.RemoveWithoutOrder(2)
	if intItemSliceUnequalWithoutOrder(s, wanted) {
		t.Errorf("After 1st remove: %v, wanted: %v.", s, wanted)
	}
	wanted = wanted[:2]
	sda.RemoveWithoutOrder(2)
	if intItemSliceUnequalWithoutOrder(s, wanted) {
		t.Errorf("After 2nd remove: %v, wanted: %v.", s, wanted)
	}
	wanted = wanted[1:]
	sda.RemoveWithoutOrder(0)
	if intItemSliceUnequalWithoutOrder(s, wanted) {
		t.Errorf("After 3rd remove: %v, wanted: %v.", s, wanted)
	}
	sda.RemoveWithoutOrder(0)
	wanted = wanted[:0]
	if intItemSliceUnequalWithoutOrder(s, wanted) {
		t.Errorf("After 4th remove: %v, wanted: %v.", s, wanted)
	}
	if testNonNilItem(data) {
		t.Errorf("After remove all, underlying data have non-nil item: %v.", data)
	}
}

func TestSliceDynamicArray_InsertSequence(t *testing.T) {
	var s []interface{}
	sda := WrapSlice(&s)
	seq := GeneralDynamicArray{1, 2, 3}
	wanted := []interface{}{1, 2, 3}
	sda.InsertSequence(0, seq)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 1st insert sequence: %v, wanted: %v.", s, wanted)
	}
	wanted = append(wanted, seq...)
	sda.InsertSequence(3, seq)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 2nd insert sequence: %v, wanted: %v.", s, wanted)
	}
	wanted = []interface{}{1, 2, 1, 2, 3, 3, 1, 2, 3}
	sda.InsertSequence(2, seq)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 3rd insert sequence: %v, wanted: %v.", s, wanted)
	}
	wanted = append(seq, wanted...)
	sda.InsertSequence(0, seq)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 4th insert sequence: %v, wanted: %v.", s, wanted)
	}
	sda.InsertSequence(1, nil)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 5th insert sequence: %v, wanted: %v.", s, wanted)
	}
}

func TestSliceDynamicArray_Cut(t *testing.T) {
	data := []interface{}{1, 2, 3, 3, 4, 5, 5}
	s := data
	sda := WrapSlice(&s)
	wanted := []interface{}{1, 2, 3, 3, 4}
	sda.Cut(5, 7)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 1st cut: %v, wanted: %v.", s, wanted)
	}
	wanted = []interface{}{1, 2, 3, 4}
	sda.Cut(3, 4)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 2nd cut: %v, wanted: %v.", s, wanted)
	}
	wanted = wanted[2:4]
	sda.Cut(0, 2)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 3rd cut: %v, wanted: %v.", s, wanted)
	}
	wanted = wanted[:0]
	sda.Cut(0, 2)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 4th cut: %v, wanted: %v.", s, wanted)
	}
	if testNonNilItem(data) {
		t.Errorf("After cut all, underlying data have non-nil item: %v.", data)
	}
	data = []interface{}{1, 2, 3, 4, 5, 6, 7, 8}
	s = data
	sda = WrapSlice(&s)
	wanted = []interface{}{1, 2, 7, 8}
	sda.Cut(2, 6)
	if sliceUnequal(s, wanted) {
		t.Errorf("After another cut: %v, wanted: %v.", s, wanted)
	}
	if testNonNilItem(data[4:]) {
		t.Errorf("After another cut, the tail of underlying data has non-nil item: %v.", data)
	}
}

func TestSliceDynamicArray_CutWithoutOrder(t *testing.T) {
	data := []interface{}{1, 2, 3, 3, 4, 5, 5}
	s := data
	sda := WrapSlice(&s)
	wanted := []interface{}{1, 2, 3, 3, 4}
	sda.CutWithoutOrder(5, 7)
	if intItemSliceUnequalWithoutOrder(s, wanted) {
		t.Errorf("After 1st cut: %v, wanted: %v.", s, wanted)
	}
	wanted = []interface{}{1, 2, 3, 4}
	sda.CutWithoutOrder(3, 4)
	if intItemSliceUnequalWithoutOrder(s, wanted) {
		t.Errorf("After 2nd cut: %v, wanted: %v.", s, wanted)
	}
	wanted = wanted[2:4]
	sda.CutWithoutOrder(0, 2)
	if intItemSliceUnequalWithoutOrder(s, wanted) {
		t.Errorf("After 3rd cut: %v, wanted: %v.", s, wanted)
	}
	wanted = wanted[:0]
	sda.CutWithoutOrder(0, 2)
	if intItemSliceUnequalWithoutOrder(s, wanted) {
		t.Errorf("After 4th cut: %v, wanted: %v.", s, wanted)
	}
	if testNonNilItem(data) {
		t.Errorf("After cut all, underlying data have non-nil item: %v.", data)
	}
	data = []interface{}{1, 2, 3, 4, 5, 6, 7, 8}
	s = data
	sda = WrapSlice(&s)
	wanted = []interface{}{1, 2, 7, 8}
	sda.CutWithoutOrder(2, 6)
	if intItemSliceUnequalWithoutOrder(s, wanted) {
		t.Errorf("After another cut: %v, wanted: %v.", s, wanted)
	}
	if testNonNilItem(data[4:]) {
		t.Errorf("After another cut, the tail of underlying data has non-nil item: %v.", data)
	}
}

func TestSliceDynamicArray_Extend(t *testing.T) {
	var s []interface{}
	sda := WrapSlice(&s)
	wanted := make([]interface{}, 3, 8)
	sda.Extend(3)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 1st extend: %v, wanted: %v.", s, wanted)
	}
	wanted = append(wanted, nil)
	sda.Extend(1)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 2nd extend: %v, wanted: %v.", s, wanted)
	}
	wanted = append(wanted, nil, nil, nil, nil)
	sda.Extend(4)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 3rd extend: %v, wanted: %v.", s, wanted)
	}
	s = []interface{}{1, 2}
	sda = WrapSlice(&s)
	wanted = []interface{}{1, 2, nil, nil, nil}
	sda.Extend(3)
	if sliceUnequal(s, wanted) {
		t.Errorf("After another extend: %v, wanted: %v.", s, wanted)
	}
}

func TestSliceDynamicArray_Expand(t *testing.T) {
	var s []interface{}
	sda := WrapSlice(&s)
	wanted := make([]interface{}, 2, 5)
	sda.Expand(0, 2)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 1st expand: %v, wanted: %v.", s, wanted)
	}
	wanted = append(wanted, nil)
	sda.Expand(0, 1)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 2nd expand: %v, wanted: %v.", s, wanted)
	}
	wanted = append(wanted, nil, nil)
	sda.Expand(1, 2)
	if sliceUnequal(s, wanted) {
		t.Errorf("After 3rd expand: %v, wanted: %v.", s, wanted)
	}
	s = []interface{}{1, 2, 3, 4, 5}
	sda = WrapSlice(&s)
	wanted = []interface{}{1, 2, nil, nil, nil, 3, 4, 5}
	sda.Expand(2, 3)
	if sliceUnequal(s, wanted) {
		t.Errorf("After another expand: %v, wanted: %v.", s, wanted)
	}
}

func TestSliceDynamicArray_Reserve(t *testing.T) {
	s := make([]interface{}, 2, 5)
	sda := WrapSlice(&s)
	sda.Reserve(1)
	if c := sda.Cap(); c < 1 {
		t.Errorf("After 1st reserve, Cap(): %d < 1.", c)
	}
	sda.Reserve(5)
	if c := sda.Cap(); c < 5 {
		t.Errorf("After 2nd reserve, Cap(): %d < 5.", c)
	}
	sda.Reserve(10)
	if c := sda.Cap(); c < 10 {
		t.Errorf("After 3rd reserve, Cap(): %d < 10.", c)
	}
}

func TestSliceDynamicArray_Shrink(t *testing.T) {
	data := make([]interface{}, 3, 10)
	s := data
	sda := WrapSlice(&s)
	sda.Shrink()
	if n, c := sda.Len(), sda.Cap(); c != n {
		t.Errorf("After shrink, Len(): %d, Cap(): %d.", n, c)
	}
	sda.Set(0, 1)
	if data[0] != nil {
		t.Errorf("After shrink, underlying data didn't change.")
	}
}

func TestSliceDynamicArray_Filter(t *testing.T) {
	data := []interface{}{1, 2, 0, -1, -4, 1, 3, -5, 0}
	s := data
	sda := WrapSlice(&s)
	wanted := []interface{}{1, 2, 1, 3}
	filter := func(x interface{}) (keep bool) {
		return x.(int) > 0
	}
	sda.Filter(filter)
	if sliceUnequal(s, wanted) {
		t.Errorf("After filter: %v, wanted: %v.", s, wanted)
	}
	if testNonNilItem(data[4:]) {
		t.Errorf("After filter, the tail of underlying data has non-nil item: %v.", data)
	}
}
