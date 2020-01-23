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
	"testing"
)

func TestNewSliceDynamicArray(t *testing.T) {
	s := []interface{}{1, 2, 3}
	ptr := &s
	sda := NewSliceDynamicArray(ptr)
	if p := sda.p.Interface(); p != ptr {
		t.Errorf("sda.p: %v != &s: %p.", p, ptr)
	}
	if v := sda.v.Interface().([]interface{}); sliceUnequal(v, s) {
		t.Errorf("sda.v: %v != s: %v.", v, s)
	}
	itf := interface{}(s)
	sda = NewSliceDynamicArray(&itf)
	if p := sda.p.Interface(); p != &itf {
		t.Errorf("sda.p: %v != &s: %p.", p, &itf)
	}
	if v := sda.v.Interface().([]interface{}); sliceUnequal(v, s) {
		t.Errorf("sda.v: %v != s: %v.", v, s)
	}
	sda = NewSliceDynamicArray(&ptr)
	if p := sda.p.Interface(); p != &ptr {
		t.Errorf("sda.p: %v != &s: %p.", p, &ptr)
	}
	if v := sda.v.Interface().([]interface{}); sliceUnequal(v, s) {
		t.Errorf("sda.v: %v != s: %v.", v, s)
	}
}

func TestMakeSliceDynamicArray(t *testing.T) {
	var i int
	sda := MakeSliceDynamicArray(reflect.TypeOf(i), 2, 5)
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
	sda = NewSliceDynamicArray(&s)
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
	sda = NewSliceDynamicArray(&s)
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
	s := []interface{}{}
	sda = NewSliceDynamicArray(&s)
	if n := sda.Len(); n != len(s) {
		t.Errorf("sda.Len(): %d != len(s): %d.", n, len(s))
	}
	s = []interface{}{1}
	sda = NewSliceDynamicArray(&s)
	if n := sda.Len(); n != len(s) {
		t.Errorf("sda.Len(): %d != len(s): %d.", n, len(s))
	}
}

func TestSliceDynamicArray_Reverse(t *testing.T) {
	s := []interface{}{1, 1, 2, 3, 4}
	sda := NewSliceDynamicArray(&s)
	wanted := []interface{}{4, 3, 2, 1, 1}
	sda.Reverse()
	if sliceUnequal(s, wanted) {
		t.Errorf("After reverse: %v, wanted: %v.", s, wanted)
	}
}

func TestSliceDynamicArray_Scan(t *testing.T) {
	s := []interface{}{1, 1, 2, 3}
	sda := NewSliceDynamicArray(&s)
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
	sda := NewSliceDynamicArray(&s)
	for i, x := range s {
		item := sda.Get(i)
		if item != x {
			t.Errorf("sda.Get(%d): %v != s[%[1]d]: %[3]v.", i, item, x)
		}
	}
}

func TestSliceDynamicArray_Set(t *testing.T) {
	s := []interface{}{1, 1, 2, 3}
	sda := NewSliceDynamicArray(&s)
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
	sda := NewSliceDynamicArray(&s)
	wanted := []interface{}{2, 1, 3}
	sda.Swap(1, 0)
	if sliceUnequal(s, wanted) {
		t.Errorf("After swap: %v, wanted: %v.", s, wanted)
	}
}

func TestSliceDynamicArray_Slice(t *testing.T) {
	s := []interface{}{1, 2, 3, 4}
	sda := NewSliceDynamicArray(&s)
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
	sda = NewSliceDynamicArray(&s)
	if c := sda.Cap(); c != cap(s) {
		t.Errorf("sda.Cap(): %d != cap(s): %d.", c, cap(s))
	}
	s = make([]interface{}, 2, 10)
	sda = NewSliceDynamicArray(&s)
	if c := sda.Cap(); c != cap(s) {
		t.Errorf("sda.Cap(): %d != cap(s): %d.", c, cap(s))
	}
}
