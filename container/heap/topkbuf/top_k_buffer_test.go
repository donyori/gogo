// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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

package topkbuf_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/donyori/gogo/container/heap/topkbuf"
	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/fmtcoll"
	"github.com/donyori/gogo/function/compare"
)

type IntSDAPtr = *array.SliceDynamicArray[int]

var IntLess = compare.OrderedLess[int]

var dataList = [][]int{
	nil, {},
	{0},
	{0, 0}, {0, 1}, {1, 0},
	{0, 0, 0}, {0, 0, 1}, {0, 1, 0}, {0, 1, 1}, {1, 0, 0}, {1, 0, 1}, {1, 1, 0},
	{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0},
	{0, 1, 2, 3, 4, 5, 6}, {0, 2, 4, 6, 1, 3, 5}, {4, 5, 6, 0, 1, 2, 3},
	{3, 2, 1, 0, 4, 5, 6}, {6, 5, 4, 3, 2, 1, 0},
}

var maxK int // it is set in function init

func init() {
	for _, data := range dataList {
		if maxK < len(data) {
			maxK = len(data)
		}
	}
}

func TestNew(t *testing.T) {
	testCases := make([]struct {
		data []int
		k    int
		want []int
	}, len(dataList)*maxK)
	var idx int
	for _, data := range dataList {
		for k := 1; k <= maxK; k++ {
			testCases[idx].data = data
			testCases[idx].k = k
			testCases[idx].want = kSuffixAndReverse(copyAndSort(data), k)
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&k=%d", sliceToName(tc.data), tc.k),
			func(t *testing.T) {
				tkb := topkbuf.New[int](tc.k, IntLess, IntSDAPtr(&tc.data))
				checkTopKBufferByDrain(t, tkb, tc.want)
			},
		)
	}
}

func TestTopKBuffer_Len(t *testing.T) {
	testCases := make([]struct {
		data    []int
		k, want int
	}, len(dataList)*maxK)
	var idx int
	for _, data := range dataList {
		for k := 1; k <= maxK; k++ {
			testCases[idx].data = data
			testCases[idx].k = k
			if len(data) < k {
				testCases[idx].want = len(data)
			} else {
				testCases[idx].want = k
			}
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&k=%d", sliceToName(tc.data), tc.k),
			func(t *testing.T) {
				tkb := topkbuf.New[int](tc.k, IntLess, IntSDAPtr(&tc.data))
				if n := tkb.Len(); n != tc.want {
					t.Errorf("got %d; want %d", n, tc.want)
				}
			},
		)
	}
}

func TestTopKBuffer_Range(t *testing.T) {
	testCases := make([]struct {
		data []int
		k    int
		want []int
	}, len(dataList)*maxK)
	var idx int
	for _, data := range dataList {
		for k := 1; k <= maxK; k++ {
			testCases[idx].data = data
			testCases[idx].k = k
			testCases[idx].want = kSuffixAndReverse(copyAndSort(data), k)
			idx++
		}
	}

	for _, tc := range testCases {
		counterMap := make(map[int]int, len(tc.want))
		for _, x := range tc.want {
			counterMap[x]++
		}
		t.Run(
			fmt.Sprintf("data=%s&k=%d", sliceToName(tc.data), tc.k),
			func(t *testing.T) {
				tkb := topkbuf.New[int](tc.k, IntLess, IntSDAPtr(&tc.data))
				tkb.Range(func(x int) (cont bool) {
					counterMap[x]--
					return true
				})
				for x, ctr := range counterMap {
					if ctr > 0 {
						t.Errorf("insufficient accesses to %d", x)
					} else if ctr < 0 {
						t.Errorf("too many accesses to %d", x)
					}
				}
			},
		)
	}
}

func TestTopKBuffer_K(t *testing.T) {
	for _, data := range dataList {
		for k := 1; k <= maxK; k++ {
			t.Run(
				fmt.Sprintf("data=%s&k=%d", sliceToName(data), k),
				func(t *testing.T) {
					tkb := topkbuf.New[int](k, IntLess, IntSDAPtr(&data))
					if got := tkb.K(); got != k {
						t.Errorf("got %d; want %d", got, k)
					}
				},
			)
		}
	}
}

func TestTopKBuffer_Add(t *testing.T) {
	xsList := [][]int{nil, {}, {-1}, {0}, {1}, {7}, {-1, 0, 1}, {0, 0, 7}}

	testCases := make([]struct {
		data     []int
		k        int
		xs, want []int
	}, len(dataList)*maxK*len(xsList))
	var idx int
	for _, data := range dataList {
		for k := 1; k <= maxK; k++ {
			for _, xs := range xsList {
				testCases[idx].data = data
				testCases[idx].k = k
				testCases[idx].xs = xs
				want := make([]int, len(data)+len(xs))
				copy(want[copy(want, data):], xs)
				slices.Sort(want)
				testCases[idx].want = kSuffixAndReverse(want, k)
				idx++
			}
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&k=%d&xs=%s",
				sliceToName(tc.data), tc.k, sliceToName(tc.xs)),
			func(t *testing.T) {
				tkb := topkbuf.New[int](tc.k, IntLess, IntSDAPtr(&tc.data))
				tkb.Add(tc.xs...)
				checkTopKBufferByDrain(t, tkb, tc.want)
			},
		)
	}
}

func TestTopKBuffer_Drain(t *testing.T) {
	testCases := make([]struct {
		data []int
		k    int
		want []int
	}, len(dataList)*maxK)
	var idx int
	for _, data := range dataList {
		for k := 1; k <= maxK; k++ {
			testCases[idx].data = data
			testCases[idx].k = k
			testCases[idx].want = kSuffixAndReverse(copyAndSort(data), k)
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&k=%d", sliceToName(tc.data), tc.k),
			func(t *testing.T) {
				tkb := topkbuf.New[int](tc.k, IntLess, IntSDAPtr(&tc.data))
				if topK := tkb.Drain(); !slices.Equal(topK, tc.want) {
					t.Errorf("got %v; want %v", topK, tc.want)
				}
			},
		)
	}
}

func TestTopKBuffer_Clear(t *testing.T) {
	for _, data := range dataList {
		for k := 1; k <= maxK; k++ {
			t.Run(
				fmt.Sprintf("data=%s&k=%d", sliceToName(data), k),
				func(t *testing.T) {
					tkb := topkbuf.New[int](k, IntLess, IntSDAPtr(&data))
					tkb.Clear()
					checkTopKBufferByDrain(t, tkb, nil)
				},
			)
		}
	}
}

func sliceToName[T any](s []T) string {
	return fmtcoll.MustFormatSliceToString(s, &fmtcoll.SequenceFormat[T]{
		CommonFormat: fmtcoll.CommonFormat{
			Separator:   ",",
			PrependType: true,
		},
		FormatItemFn: fmtcoll.FprintfToFormatFunc[T]("%v"),
	})
}

func copyAndSort(data []int) []int {
	if data == nil {
		return nil
	}
	sorted := make([]int, len(data))
	copy(sorted, data)
	slices.Sort(sorted)
	return sorted
}

// !!the last k items in s will be modified.
func kSuffixAndReverse[Item any](s []Item, k int) []Item {
	if s == nil {
		return nil
	} else if len(s) > k {
		s = s[len(s)-k:]
	}
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// !!tkb may be modified in this function.
func checkTopKBufferByDrain[Item comparable](
	t *testing.T, tkb topkbuf.TopKBuffer[Item], want []Item) {
	if topK := tkb.Drain(); !slices.Equal(topK, want) {
		t.Errorf("tkb contains %v; want %v", topK, want)
	}
}
