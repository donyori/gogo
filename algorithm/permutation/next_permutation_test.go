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

package permutation_test

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/donyori/gogo/algorithm/permutation"
	"github.com/donyori/gogo/function/compare"
)

var dataList = [][][]int{
	{
		nil,
	},
	{
		{},
	},
	{
		{0},
	},
	{
		{0, 0},
	},
	{
		{0, 1},
		{1, 0},
	},
	{
		{0, 0, 0},
	},
	{
		{0, 0, 1},
		{0, 1, 0},
		{1, 0, 0},
	},
	{
		{0, 1, 1},
		{1, 0, 1},
		{1, 1, 0},
	},
	{
		{0, 1, 2},
		{0, 2, 1},
		{1, 0, 2},
		{1, 2, 0},
		{2, 0, 1},
		{2, 1, 0},
	},
	{
		{0, 0, 0, 0},
	},
	{
		{0, 0, 0, 1},
		{0, 0, 1, 0},
		{0, 1, 0, 0},
		{1, 0, 0, 0},
	},
	{
		{0, 0, 1, 1},
		{0, 1, 0, 1},
		{0, 1, 1, 0},
		{1, 0, 0, 1},
		{1, 0, 1, 0},
		{1, 1, 0, 0},
	},
	{
		{0, 1, 1, 1},
		{1, 0, 1, 1},
		{1, 1, 0, 1},
		{1, 1, 1, 0},
	},
	{
		{0, 0, 1, 2},
		{0, 0, 2, 1},
		{0, 1, 0, 2},
		{0, 1, 2, 0},
		{0, 2, 0, 1},
		{0, 2, 1, 0},
		{1, 0, 0, 2},
		{1, 0, 2, 0},
		{1, 2, 0, 0},
		{2, 0, 0, 1},
		{2, 0, 1, 0},
		{2, 1, 0, 0},
	},
	{
		{0, 1, 1, 2},
		{0, 1, 2, 1},
		{0, 2, 1, 1},
		{1, 0, 1, 2},
		{1, 0, 2, 1},
		{1, 1, 0, 2},
		{1, 1, 2, 0},
		{1, 2, 0, 1},
		{1, 2, 1, 0},
		{2, 0, 1, 1},
		{2, 1, 0, 1},
		{2, 1, 1, 0},
	},
	{
		{0, 1, 2, 2},
		{0, 2, 1, 2},
		{0, 2, 2, 1},
		{1, 0, 2, 2},
		{1, 2, 0, 2},
		{1, 2, 2, 0},
		{2, 0, 1, 2},
		{2, 0, 2, 1},
		{2, 1, 0, 2},
		{2, 1, 2, 0},
		{2, 2, 0, 1},
		{2, 2, 1, 0},
	},
	{
		{0, 1, 2, 3},
		{0, 1, 3, 2},
		{0, 2, 1, 3},
		{0, 2, 3, 1},
		{0, 3, 1, 2},
		{0, 3, 2, 1},
		{1, 0, 2, 3},
		{1, 0, 3, 2},
		{1, 2, 0, 3},
		{1, 2, 3, 0},
		{1, 3, 0, 2},
		{1, 3, 2, 0},
		{2, 0, 1, 3},
		{2, 0, 3, 1},
		{2, 1, 0, 3},
		{2, 1, 3, 0},
		{2, 3, 0, 1},
		{2, 3, 1, 0},
		{3, 0, 1, 2},
		{3, 0, 2, 1},
		{3, 1, 0, 2},
		{3, 1, 2, 0},
		{3, 2, 0, 1},
		{3, 2, 1, 0},
	},
}

// MaxItem is the maximum of items in dataList.
const MaxItem int = 3

var alphabet = [MaxItem + 1]string{"0", "1", "2", "3"}

func init() {
	// Check that MaxItem and alphabet are valid.
	var max int
	for _, ps := range dataList {
		for _, data := range ps {
			for _, x := range data {
				if x > max {
					max = x
				}
			}
		}
	}
	if max != MaxItem {
		panic("MaxItem needs to be updated")
	}
	for i := range alphabet {
		if s := strconv.Itoa(i); alphabet[i] != s {
			panic("alphabet[" + s + "] needs to be updated")
		}
	}
}

type testCase struct {
	data, wantData []int
	wantMore       bool
}

func TestNextPermutation(t *testing.T) {
	var testCases []testCase
	for _, ps := range dataList {
		for i := 0; i < len(ps)-1; i++ {
			testCases = append(testCases, testCase{
				data:     ps[i],
				wantData: ps[i+1],
				wantMore: true,
			})
		}
		testCases = append(testCases, testCase{
			data:     ps[len(ps)-1],
			wantData: ps[len(ps)-1],
		})
	}

	for _, tc := range testCases {
		t.Run("data="+dataToName(tc.data), func(t *testing.T) {
			var data []int
			if tc.data != nil {
				data = make([]int, len(tc.data))
				copy(data, tc.data)
			}
			itf := sort.IntSlice(data)
			if more := permutation.NextPermutation(itf); more != tc.wantMore {
				t.Errorf("return value: got %t; want %t", more, tc.wantMore)
			}
			if dataUnequal(data, tc.wantData) {
				t.Errorf("permutation: got %v; want %v", data, tc.wantData)
			}
		})
	}
}

func dataToName(data []int) string {
	if data == nil {
		return "<nil>"
	}
	var b strings.Builder
	b.Grow(len(data)*2 + 2)
	b.WriteByte('[')
	for i := range data {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(alphabet[data[i]])
	}
	b.WriteByte(']')
	return b.String()
}

func dataUnequal(a, b []int) bool {
	if a == nil {
		return b != nil
	}
	return b == nil || !compare.ComparableSliceEqual(a, b)
}
