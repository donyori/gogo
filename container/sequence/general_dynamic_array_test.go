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

import "testing"

func TestNewGeneralDynamicArray(t *testing.T) {
	gda := NewGeneralDynamicArray(3)
	if n, c := gda.Len(), gda.Cap(); n != 0 || c != 3 {
		t.Errorf("NewGeneralDynamicArray(3) - Len(): %d, Cap(): %d.", n, c)
	}
}

func TestGeneralDynamicArray_Len(t *testing.T) {
	var gda GeneralDynamicArray
	if n := gda.Len(); n != 0 {
		t.Errorf("gda.Len(): %d != 0.", n)
	}
	gda = GeneralDynamicArray{}
	if n := gda.Len(); n != 0 {
		t.Errorf("gda.Len(): %d != 0.", n)
	}
	gda = GeneralDynamicArray{1}
	if n := gda.Len(); n != 1 {
		t.Errorf("gda.Len(): %d != 1.", n)
	}
}

func TestGeneralDynamicArray_Reverse(t *testing.T) {
	gda := GeneralDynamicArray{1, 1, 2, 3, 4}
	wanted := []interface{}{4, 3, 2, 1, 1}
	gda.Reverse()
	if sliceUnequal(gda, wanted) {
		t.Errorf("After reverse: %v, wanted: %v.", gda, wanted)
	}
}

func TestGeneralDynamicArray_Swap(t *testing.T) {
	gda := GeneralDynamicArray{1, 2, 3}
	wanted := []interface{}{2, 1, 3}
	gda.Swap(1, 0)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After swap: %v, wanted: %v.", gda, wanted)
	}
}

func TestGeneralDynamicArray_Cap(t *testing.T) {
	var gda GeneralDynamicArray
	if c := gda.Cap(); c != 0 {
		t.Errorf("gda.Cap(): %d != 0.", c)
	}
	gda = make(GeneralDynamicArray, 0)
	if c := gda.Cap(); c != 0 {
		t.Errorf("gda.Cap(): %d != 0.", c)
	}
	gda = make(GeneralDynamicArray, 2, 10)
	if c := gda.Cap(); c != 10 {
		t.Errorf("gda.Cap(): %d != 10.", c)
	}
}

func TestGeneralDynamicArray_Push(t *testing.T) {
	var gda GeneralDynamicArray
	wanted := []interface{}{1}
	gda.Push(1)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 1st push: %v, wanted: %v.", gda, wanted)
	}
	wanted = append(wanted, 2)
	gda.Push(2)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 2nd push: %v, wanted: %v.", gda, wanted)
	}
	wanted = append(wanted, nil)
	gda.Push(nil)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 3rd push: %v, wanted: %v.", gda, wanted)
	}
}

func TestGeneralDynamicArray_Pop(t *testing.T) {
	data := []interface{}{1, 2}
	gda := GeneralDynamicArray(data)
	wanted := []interface{}{1}
	x := gda.Pop()
	if sliceUnequal(gda, wanted) || x.(int) != 2 {
		t.Errorf("After 1st pop: %v, wanted: %v, x = %v.", gda, wanted, x)
	}
	wanted = wanted[:0]
	x = gda.Pop()
	if sliceUnequal(gda, wanted) || x.(int) != 1 {
		t.Errorf("After 2nd pop: %v, wanted %v, x = %v.", gda, wanted, x)
	}
	if testNonNilItem(data) {
		t.Errorf("After pop all, underlying data have non-nil item: %v.", data)
	}
}

func TestGeneralDynamicArray_Append(t *testing.T) {
	var gda GeneralDynamicArray
	seq := GeneralDynamicArray{1, 2, 3}
	wanted := []interface{}{1, 2, 3}
	gda.Append(seq)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 1st append: %v, wanted: %v.", gda, wanted)
	}
	wanted = append(wanted, seq...)
	gda.Append(seq)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 2nd append: %v, wanted: %v.", gda, wanted)
	}
	wanted = append(wanted, 1)
	gda.Append(GeneralDynamicArray{1})
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 3rd append: %v, wanted: %v.", gda, wanted)
	}
	gda.Append(nil)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 4th append: %v, wanted: %v.", gda, wanted)
	}
}

func TestGeneralDynamicArray_Truncate(t *testing.T) {
	data := []interface{}{1, 2, 2, 3, 4, 4}
	gda := GeneralDynamicArray(data)
	wanted := []interface{}{1, 2, 2, 3}
	gda.Truncate(4)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 1st truncate: %v, wanted: %v.", gda, wanted)
	}
	wanted = wanted[:3]
	gda.Truncate(3)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 2nd truncate: %v, wanted: %v.", gda, wanted)
	}
	wanted = wanted[:0]
	gda.Truncate(0)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 3rd truncate: %v, wanted: %v.", gda, wanted)
	}
	if testNonNilItem(data) {
		t.Errorf("After truncate all, underlying data have non-nil item: %v.", data)
	}
}

func TestGeneralDynamicArray_Insert(t *testing.T) {
	var gda GeneralDynamicArray
	wanted := []interface{}{1}
	gda.Insert(0, 1)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 1st insert: %v, wanted: %v.", gda, wanted)
	}
	wanted = append(wanted, 2)
	gda.Insert(1, 2)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 2nd insert: %v, wanted: %v.", gda, wanted)
	}
	wanted = []interface{}{0, 1, 2}
	gda.Insert(0, 0)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 3rd insert: %v, wanted: %v.", gda, wanted)
	}
	wanted = []interface{}{0, nil, 1, 2}
	gda.Insert(1, nil)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 4th insert: %v, wanted: %v.", gda, wanted)
	}
}

func TestGeneralDynamicArray_Remove(t *testing.T) {
	data := []interface{}{1, 2, 2, 4}
	gda := GeneralDynamicArray(data)
	wanted := []interface{}{1, 2, 4}
	gda.Remove(2)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 1st remove: %v, wanted: %v.", gda, wanted)
	}
	wanted = wanted[:2]
	gda.Remove(2)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 2nd remove: %v, wanted: %v.", gda, wanted)
	}
	wanted = wanted[1:]
	gda.Remove(0)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 3rd remove: %v, wanted: %v.", gda, wanted)
	}
	gda.Remove(0)
	wanted = wanted[:0]
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 4th remove: %v, wanted: %v.", gda, wanted)
	}
	if testNonNilItem(data) {
		t.Errorf("After remove all, underlying data have non-nil item: %v.", data)
	}
}

func TestGeneralDynamicArray_RemoveWithoutOrder(t *testing.T) {
	data := []interface{}{1, 2, 2, 4}
	gda := GeneralDynamicArray(data)
	wanted := []interface{}{1, 2, 4}
	gda.RemoveWithoutOrder(2)
	if intItemSliceUnequalWithoutOrder(gda, wanted) {
		t.Errorf("After 1st remove: %v, wanted: %v.", gda, wanted)
	}
	wanted = wanted[:2]
	gda.RemoveWithoutOrder(2)
	if intItemSliceUnequalWithoutOrder(gda, wanted) {
		t.Errorf("After 2nd remove: %v, wanted: %v.", gda, wanted)
	}
	wanted = wanted[1:]
	gda.RemoveWithoutOrder(0)
	if intItemSliceUnequalWithoutOrder(gda, wanted) {
		t.Errorf("After 3rd remove: %v, wanted: %v.", gda, wanted)
	}
	gda.RemoveWithoutOrder(0)
	wanted = wanted[:0]
	if intItemSliceUnequalWithoutOrder(gda, wanted) {
		t.Errorf("After 4th remove: %v, wanted: %v.", gda, wanted)
	}
	if testNonNilItem(data) {
		t.Errorf("After remove all, underlying data have non-nil item: %v.", data)
	}
}

func TestGeneralDynamicArray_InsertSequence(t *testing.T) {
	var gda GeneralDynamicArray
	seq := GeneralDynamicArray{1, 2, 3}
	wanted := []interface{}{1, 2, 3}
	gda.InsertSequence(0, seq)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 1st insert sequence: %v, wanted: %v.", gda, wanted)
	}
	wanted = append(wanted, seq...)
	gda.InsertSequence(3, seq)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 2nd insert sequence: %v, wanted: %v.", gda, wanted)
	}
	wanted = []interface{}{1, 2, 1, 2, 3, 3, 1, 2, 3}
	gda.InsertSequence(2, seq)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 3rd insert sequence: %v, wanted: %v.", gda, wanted)
	}
	wanted = append(seq, wanted...)
	gda.InsertSequence(0, seq)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 4th insert sequence: %v, wanted: %v.", gda, wanted)
	}
	gda.InsertSequence(1, nil)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 5th insert sequence: %v, wanted: %v.", gda, wanted)
	}
}

func TestGeneralDynamicArray_Cut(t *testing.T) {
	data := []interface{}{1, 2, 3, 3, 4, 5, 5}
	gda := GeneralDynamicArray(data)
	wanted := []interface{}{1, 2, 3, 3, 4}
	gda.Cut(5, 7)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 1st cut: %v, wanted: %v.", gda, wanted)
	}
	wanted = []interface{}{1, 2, 3, 4}
	gda.Cut(3, 4)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 2nd cut: %v, wanted: %v.", gda, wanted)
	}
	wanted = wanted[2:4]
	gda.Cut(0, 2)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 3rd cut: %v, wanted: %v.", gda, wanted)
	}
	wanted = wanted[:0]
	gda.Cut(0, 2)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 4th cut: %v, wanted: %v.", gda, wanted)
	}
	if testNonNilItem(data) {
		t.Errorf("After cut all, underlying data have non-nil item: %v.", data)
	}
	data = []interface{}{1, 2, 3, 4, 5, 6, 7, 8}
	gda = data
	wanted = []interface{}{1, 2, 7, 8}
	gda.Cut(2, 6)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After another cut: %v, wanted: %v.", gda, wanted)
	}
	if testNonNilItem(data[4:]) {
		t.Errorf("After another cut, the tail of underlying data has non-nil item: %v.", data)
	}
}

func TestGeneralDynamicArray_CutWithoutOrder(t *testing.T) {
	data := []interface{}{1, 2, 3, 3, 4, 5, 5}
	gda := GeneralDynamicArray(data)
	wanted := []interface{}{1, 2, 3, 3, 4}
	gda.CutWithoutOrder(5, 7)
	if intItemSliceUnequalWithoutOrder(gda, wanted) {
		t.Errorf("After 1st cut: %v, wanted: %v.", gda, wanted)
	}
	wanted = []interface{}{1, 2, 3, 4}
	gda.CutWithoutOrder(3, 4)
	if intItemSliceUnequalWithoutOrder(gda, wanted) {
		t.Errorf("After 2nd cut: %v, wanted: %v.", gda, wanted)
	}
	wanted = wanted[2:4]
	gda.CutWithoutOrder(0, 2)
	if intItemSliceUnequalWithoutOrder(gda, wanted) {
		t.Errorf("After 3rd cut: %v, wanted: %v.", gda, wanted)
	}
	wanted = wanted[:0]
	gda.CutWithoutOrder(0, 2)
	if intItemSliceUnequalWithoutOrder(gda, wanted) {
		t.Errorf("After 4th cut: %v, wanted: %v.", gda, wanted)
	}
	if testNonNilItem(data) {
		t.Errorf("After cut all, underlying data have non-nil item: %v.", data)
	}
	data = []interface{}{1, 2, 3, 4, 5, 6, 7, 8}
	gda = data
	wanted = []interface{}{1, 2, 7, 8}
	gda.CutWithoutOrder(2, 6)
	if intItemSliceUnequalWithoutOrder(gda, wanted) {
		t.Errorf("After another cut: %v, wanted: %v.", gda, wanted)
	}
	if testNonNilItem(data[4:]) {
		t.Errorf("After another cut, the tail of underlying data has non-nil item: %v.", data)
	}
}

func TestGeneralDynamicArray_Extend(t *testing.T) {
	var gda GeneralDynamicArray
	wanted := make([]interface{}, 3, 8)
	gda.Extend(3)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 1st extend: %v, wanted: %v.", gda, wanted)
	}
	wanted = append(wanted, nil)
	gda.Extend(1)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 2nd extend: %v, wanted: %v.", gda, wanted)
	}
	wanted = append(wanted, nil, nil, nil, nil)
	gda.Extend(4)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 3rd extend: %v, wanted: %v.", gda, wanted)
	}
	gda = GeneralDynamicArray{1, 2}
	wanted = []interface{}{1, 2, nil, nil, nil}
	gda.Extend(3)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After another extend: %v, wanted: %v.", gda, wanted)
	}
}

func TestGeneralDynamicArray_Expand(t *testing.T) {
	var gda GeneralDynamicArray
	wanted := make([]interface{}, 2, 5)
	gda.Expand(0, 2)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 1st expand: %v, wanted: %v.", gda, wanted)
	}
	wanted = append(wanted, nil)
	gda.Expand(0, 1)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 2nd expand: %v, wanted: %v.", gda, wanted)
	}
	wanted = append(wanted, nil, nil)
	gda.Expand(1, 2)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After 3rd expand: %v, wanted: %v.", gda, wanted)
	}
	gda = GeneralDynamicArray{1, 2, 3, 4, 5}
	wanted = []interface{}{1, 2, nil, nil, nil, 3, 4, 5}
	gda.Expand(2, 3)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After another expand: %v, wanted: %v.", gda, wanted)
	}
}

func TestGeneralDynamicArray_Reserve(t *testing.T) {
	gda := make(GeneralDynamicArray, 2, 5)
	gda.Reserve(1)
	if c := gda.Cap(); c < 1 {
		t.Errorf("After 1st reserve, Cap(): %d < 1.", c)
	}
	gda.Reserve(5)
	if c := gda.Cap(); c < 5 {
		t.Errorf("After 2nd reserve, Cap(): %d < 5.", c)
	}
	gda.Reserve(10)
	if c := gda.Cap(); c < 10 {
		t.Errorf("After 3rd reserve, Cap(): %d < 10.", c)
	}
}

func TestGeneralDynamicArray_Shrink(t *testing.T) {
	data := make([]interface{}, 3, 10)
	gda := GeneralDynamicArray(data)
	gda.Shrink()
	if n, c := gda.Len(), gda.Cap(); c != n {
		t.Errorf("After shrink, Len(): %d, Cap(): %d.", n, c)
	}
	gda.Set(0, 1)
	if data[0] != nil {
		t.Errorf("After shrink, underlying data didn't change.")
	}
}

func TestGeneralDynamicArray_Filter(t *testing.T) {
	data := []interface{}{1, 2, 0, -1, -4, 1, 3, -5, 0}
	gda := GeneralDynamicArray(data)
	wanted := []interface{}{1, 2, 1, 3}
	filter := func(x interface{}) (keep bool) {
		return x.(int) > 0
	}
	gda.Filter(filter)
	if sliceUnequal(gda, wanted) {
		t.Errorf("After filter: %v, wanted: %v.", gda, wanted)
	}
	if testNonNilItem(data[4:]) {
		t.Errorf("After filter, the tail of underlying data has non-nil item: %v.", data)
	}
}

func sliceUnequal(a, b []interface{}) bool {
	if len(a) != len(b) {
		return true
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return true
		}
	}
	return false
}

func intItemSliceUnequalWithoutOrder(a, b []interface{}) bool {
	if len(a) != len(b) {
		return true
	}
	counter := make(map[int]int)
	for _, x := range a {
		counter[x.(int)]++
	}
	for _, x := range b {
		counter[x.(int)]--
	}
	for _, c := range counter {
		if c != 0 {
			return true
		}
	}
	return false
}

func testNonNilItem(s []interface{}) bool {
	for _, x := range s {
		if x != nil {
			return true
		}
	}
	return false
}
