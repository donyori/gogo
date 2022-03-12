// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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
	"fmt"
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
	testCases := []struct {
		goal int
		want int
	}{
		{0, -1},
		{1, -1},
		{2, 2},
		{3, 5},
		{4, 5},
		{5, 8},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("goal=%d", tc.goal), func(t *testing.T) {
			if idx := BinarySearchMaxLess(itf, tc.goal); idx != tc.want {
				t.Errorf("got %d; want %d", idx, tc.want)
			}
		})
	}
}

func TestBinarySearchMinGreater(t *testing.T) {
	data := sequence.GeneralDynamicArray{1, 1, 1, 2, 2, 2, 4, 4, 4}
	itf := &BinarySearchArrayAdapter{
		Data:    data,
		EqualFn: compare.Equal,
		LessFn:  compare.IntLess,
	}
	testCases := []struct {
		goal int
		want int
	}{
		{0, 0},
		{1, 3},
		{2, 6},
		{3, 6},
		{4, -1},
		{5, -1},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("goal=%d", tc.goal), func(t *testing.T) {
			if idx := BinarySearchMinGreater(itf, tc.goal); idx != tc.want {
				t.Errorf("got %d; want %d", idx, tc.want)
			}
		})
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

	t.Run(fmt.Sprintf("data=%v", data), func(t *testing.T) {
		for i, goal := range s {
			value := *(goal.(*int))
			t.Run(fmt.Sprintf("goal=<index=%d&value=%d>", i, value), func(t *testing.T) {
				t.Run("equalFn=compare.Equal", func(t *testing.T) {
					if idx := BinarySearch(itf1, goal); idx != i {
						t.Errorf("got %d; want %d", idx, i)
					}
				})
				t.Run("equalFn=less.ToEqual()", func(t *testing.T) {
					if idx := BinarySearch(itf2, goal); value != data[idx] {
						t.Errorf("got %d (value %d)", idx, data[idx])
					}
				})
			})
		}
		for i, value := range negativeSamples {
			goal := &value
			t.Run(fmt.Sprintf("goal=<index=%d&value=%d&isNegativeSample>", i, value), func(t *testing.T) {
				t.Run("equalFn=compare.Equal", func(t *testing.T) {
					if idx := BinarySearch(itf1, goal); idx != -1 {
						t.Errorf("got %d; want -1", idx)
					}
				})
				t.Run("equalFn=less.ToEqual()", func(t *testing.T) {
					idx := BinarySearch(itf2, goal)
					if idx == -1 {
						var wantList []int
						for j := range data {
							if data[j] == value {
								wantList = append(wantList, j)
							}
						}
						if len(wantList) == 1 {
							t.Errorf("got -1; want %d", wantList[0])
						} else if len(wantList) > 1 {
							t.Errorf("got -1; want anyone of %v", wantList)
						}
					} else if value != data[idx] {
						t.Errorf("got %d (value %d)", idx, data[idx])
					}
				})
			})
		}
	})
}
