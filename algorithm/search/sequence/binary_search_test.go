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
	"testing"

	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/function/compare"
)

func TestBinarySearch(t *testing.T) {
	data1 := []int{0, 0, 1, 3, 3, 4, 5, 7, 7, 7, 9, 9}
	negativeSamples1 := []int{-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	testBinarySearch(t, data1, negativeSamples1)
	data2 := []int{1, 1, 1, 1}
	negativeSamples2 := []int{-1, 0, 1, 2, 3}
	testBinarySearch(t, data2, negativeSamples2)
}

func TestBinarySearchMaxLess(t *testing.T) {
	data := sequence.GeneralDynamicArray{1, 1, 1, 2, 2, 2, 4, 4, 4}
	itf := &BinarySearchArrayAdapter{
		Data:    data,
		EqualFn: compare.Equal,
		LessFn:  compare.IntLess,
	}
	cases := []struct {
		Target int
		Index  int
	}{
		{0, -1},
		{1, -1},
		{2, 2},
		{3, 5},
		{4, 5},
		{5, 8},
	}
	for _, c := range cases {
		idx := BinarySearchMaxLess(itf, c.Target)
		if idx != c.Index {
			t.Errorf("BinarySearchMaxLess(%v, %d): %d != %d.", data, c.Target, idx, c.Index)
		}
	}
}

func TestBinarySearchMinGreater(t *testing.T) {
	data := sequence.GeneralDynamicArray{1, 1, 1, 2, 2, 2, 4, 4, 4}
	itf := &BinarySearchArrayAdapter{
		Data:    data,
		EqualFn: compare.Equal,
		LessFn:  compare.IntLess,
	}
	cases := []struct {
		Target int
		Index  int
	}{
		{0, 0},
		{1, 3},
		{2, 6},
		{3, 6},
		{4, -1},
		{5, -1},
	}
	for _, c := range cases {
		idx := BinarySearchMinGreater(itf, c.Target)
		if idx != c.Index {
			t.Errorf("BinarySearchMinGreater(%v, %d): %d != %d.", data, c.Target, idx, c.Index)
		}
	}
}

func testBinarySearch(t *testing.T, data, negativeSamples []int) {
	s := make(sequence.GeneralDynamicArray, len(data))
	for i, n := 0, len(data); i < n; i++ {
		s[i] = &data[i]
	}
	var less compare.LessFunc = func(a, b interface{}) bool {
		return *(a.(*int)) < *(b.(*int))
	}
	itf1 := &BinarySearchArrayAdapter{
		Data:    s,
		EqualFn: compare.Equal,
		LessFn:  less,
	}
	itf2 := &BinarySearchArrayAdapter{
		Data:    s,
		EqualFn: less.ToEqual(),
		LessFn:  less,
	}
	for i, x := range s {
		idx := BinarySearch(itf1, x)
		if idx != i {
			t.Errorf("BinarySearch(%v [%d], ...): %d != %d.", x, *x.(*int), idx, i)
		}
		idx = BinarySearch(itf2, x)
		if *x.(*int) != data[idx] {
			t.Errorf("BinarySearch(%v [%d], ...): %d [%d].", x, *x.(*int), idx, data[idx])
		}
	}
	for i := range negativeSamples {
		idx := BinarySearch(itf1, &negativeSamples[i])
		if idx != -1 {
			t.Errorf("BinarySearch(%v [%d], ...): %d != -1.", &negativeSamples[i], negativeSamples[i], idx)
		}
		idx = BinarySearch(itf2, &negativeSamples[i])
		if idx == -1 {
			for j := range data {
				if data[j] == negativeSamples[i] {
					t.Errorf("BinarySearch(%v [%d], ...): -1 != %d.", &negativeSamples[i], negativeSamples[i], j)
				}
			}
		} else if negativeSamples[i] != data[idx] {
			t.Errorf("BinarySearch(%v [%d], ...): %d [%d].", &negativeSamples[i], negativeSamples[i], idx, negativeSamples[idx])
		}
	}
}
