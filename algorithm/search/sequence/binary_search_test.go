// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2023  Yuan Gao
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

package sequence_test

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/donyori/gogo/algorithm/search/sequence"
	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/internal/testaux"
)

type idValue struct {
	id    string
	value int
}

func (iv *idValue) String() string {
	if iv == nil {
		return "<nil>"
	}
	return iv.id
}

func idValueLess(a, b *idValue) bool {
	if a == nil {
		return b != nil
	}
	return b != nil && a.value < b.value
}

func idValueEqual(a, b *idValue) bool {
	if a == nil {
		return b == nil
	}
	return b != nil && a.id == b.id
}

func idValueSortLess(a, b *idValue) bool {
	if a == nil {
		return b != nil
	}
	if b == nil {
		return false
	}
	if a.value != b.value {
		return a.value < b.value
	}
	return a.id < b.id
}

const MaxValue int = 6                // The range of values in dataList is {0, 1, 2, 3, 4, 5, 6}.
const BaseLength = (MaxValue+1)*3 + 1 // Each value is repeated 3 times, and finally, a nil *idValue is appended.
const MaxCopy int = 3

// These variables are set in function init.
var (
	valueCounter map[int]int
	dataList     [][]*idValue
)

var acceptNotFound = map[int]bool{-1: true}

func init() {
	valueCounter = make(map[int]int, MaxValue+1)
	base := make([]*idValue, BaseLength)
	for i := 0; i < BaseLength-1; i++ {
		v := i % (MaxValue + 1)
		ctr := valueCounter[v]
		valueCounter[v] = ctr + 1
		base[i] = &idValue{id: fmt.Sprintf("%d-%d", v, ctr), value: v}
	}
	// base[BaseLength-1] is nil
	dataList = make([][]*idValue, BaseLength*MaxCopy+2)
	// dataList[0] is nil
	dataList[1] = []*idValue{}
	idx := 2
	for length := 1; length <= BaseLength; length++ {
		for numCopy := 1; numCopy <= MaxCopy; numCopy++ {
			data := make([]*idValue, length*numCopy)
			var copied int
			for copied < len(data) {
				copied += copy(data[copied:], base[:length])
			}
			if len(data) > 1 {
				sort.Slice(data, func(i, j int) bool {
					return idValueSortLess(data[i], data[j])
				})
			}
			dataList[idx], idx = data, idx+1
		}
	}
}

type testCase[AcceptType int | map[int]bool] struct {
	data   []*idValue
	goal   *idValue
	accept AcceptType
}

func TestBinarySearch(t *testing.T) {
	var testCases []testCase[map[int]bool]
	for _, data := range dataList {
		for _, goal := range data {
			accept := make(map[int]bool, 3)
			for i, x := range data {
				if idValueEqual(x, goal) {
					accept[i] = true
				}
			}
			testCases = append(testCases, testCase[map[int]bool]{
				data:   data,
				goal:   goal,
				accept: accept,
			})
		}
		for v := -2; v <= MaxValue+2; v++ {
			testCases = append(testCases, testCase[map[int]bool]{
				data: data,
				goal: &idValue{
					id:    fmt.Sprintf("%d-%d", v, valueCounter[v]),
					value: v,
				},
				accept: acceptNotFound,
			})
		}
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?data=%s&goal=%v",
			i, dataToName(tc.data), tc.goal), func(t *testing.T) {
			itf := sequence.WrapArrayLessEqual[*idValue](
				array.SliceDynamicArray[*idValue](tc.data), idValueLess, idValueEqual)
			if idx := sequence.BinarySearch(itf, tc.goal); !tc.accept[idx] {
				t.Errorf("got %d; accept %s", idx, acceptSetString(tc.accept))
			}
		})
	}
}

func TestBinarySearchMaxLess(t *testing.T) {
	var testCases []testCase[int]
	for _, data := range dataList {
		for _, goal := range data {
			want := -1
			for i := 0; i < len(data) && idValueLess(data[i], goal); i++ {
				want = i
			}
			testCases = append(testCases, testCase[int]{
				data:   data,
				goal:   goal,
				accept: want,
			})
		}
		for v := -2; v <= MaxValue+2; v++ {
			goal := &idValue{
				id:    fmt.Sprintf("%d-%d", v, valueCounter[v]),
				value: v,
			}
			want := -1
			for i := 0; i < len(data) && idValueLess(data[i], goal); i++ {
				want = i
			}
			testCases = append(testCases, testCase[int]{
				data:   data,
				goal:   goal,
				accept: want,
			})
		}
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?data=%s&goal=%v",
			i, dataToName(tc.data), tc.goal), func(t *testing.T) {
			itf := sequence.WrapArrayLessEqual[*idValue](
				array.SliceDynamicArray[*idValue](tc.data), idValueLess, idValueEqual)
			if idx := sequence.BinarySearchMaxLess(itf, tc.goal); idx != tc.accept {
				t.Errorf("got %d; want %d", idx, tc.accept)
			}
		})
	}
}

func TestBinarySearchMinGreater(t *testing.T) {
	var testCases []testCase[int]
	for _, data := range dataList {
		for _, goal := range data {
			want := -1
			for i := len(data) - 1; i >= 0 && idValueLess(goal, data[i]); i-- {
				want = i
			}
			testCases = append(testCases, testCase[int]{
				data:   data,
				goal:   goal,
				accept: want,
			})
		}
		for v := -2; v <= MaxValue+2; v++ {
			goal := &idValue{
				id:    fmt.Sprintf("%d-%d", v, valueCounter[v]),
				value: v,
			}
			want := -1
			for i := len(data) - 1; i >= 0 && idValueLess(goal, data[i]); i-- {
				want = i
			}
			testCases = append(testCases, testCase[int]{
				data:   data,
				goal:   goal,
				accept: want,
			})
		}
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?data=%s&goal=%v",
			i, dataToName(tc.data), tc.goal), func(t *testing.T) {
			itf := sequence.WrapArrayLessEqual[*idValue](
				array.SliceDynamicArray[*idValue](tc.data), idValueLess, idValueEqual)
			if idx := sequence.BinarySearchMinGreater(itf, tc.goal); idx != tc.accept {
				t.Errorf("got %d; want %d", idx, tc.accept)
			}
		})
	}
}

func acceptSetString(acceptSet map[int]bool) string {
	if acceptSet == nil {
		return "<nil>"
	}
	vs := make([]int, len(acceptSet))
	var i int
	for v := range acceptSet {
		vs[i], i = v, i+1
	}
	sort.Ints(vs)
	var b strings.Builder
	b.Grow(len(vs)*3 + 2)
	b.WriteByte('[')
	for i, v := range vs {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(v))
	}
	b.WriteByte(']')
	return b.String()
}

func dataToName(data []*idValue) string {
	return testaux.SliceToName(data, ",", "%v", false)
}
