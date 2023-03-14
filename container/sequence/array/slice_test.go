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

package array_test

import (
	"container/list"
	"fmt"
	"strings"
	"testing"

	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/fmtcoll"
	"github.com/donyori/gogo/function/compare"
)

type IntSDA = array.SliceDynamicArray[int]

func TestSliceDynamicArray_Len(t *testing.T) {
	testCases := []struct {
		sda  *IntSDA
		want int
	}{
		{nil, 0},
		{new(IntSDA), 0},
		{&IntSDA{}, 0},
		{&IntSDA{1}, 1},
	}

	for _, tc := range testCases {
		t.Run("s="+sdaPtrToName(tc.sda), func(t *testing.T) {
			if n := tc.sda.Len(); n != tc.want {
				t.Errorf("got %d; want %d", n, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Range(t *testing.T) {
	sda := IntSDA{0, 1, 2, 3, 4, 0, 1, 2, 3, 4}
	want := []int{0, 1, 2, 3, 4}
	s := make([]int, 0, len(sda))
	sda.Range(func(x int) (cont bool) {
		s = append(s, x)
		return len(s) < len(sda)/2
	})
	if sliceUnequal(s, want) {
		t.Errorf("got %v; want %v", s, want)
	}
}

func TestSliceDynamicArray_Range_NilAndEmpty(t *testing.T) {
	sdas := []*IntSDA{nil, new(IntSDA), {}}
	for _, sda := range sdas {
		t.Run("s="+sdaPtrToName(sda), func(t *testing.T) {
			sda.Range(func(x int) (cont bool) {
				t.Error("handler was called, x:", x)
				return true
			})
		})
	}
}

func TestSliceDynamicArray_Front(t *testing.T) {
	testCases := []struct {
		sda  IntSDA
		want int
	}{
		{IntSDA{0}, 0},
		{IntSDA{0, 1}, 0},
		{IntSDA{0, 1, 2}, 0},
		{IntSDA{1}, 1},
		{IntSDA{1, 2}, 1},
		{IntSDA{1, 2, 3}, 1},
	}

	for _, tc := range testCases {
		t.Run("s="+sliceToName(tc.sda), func(t *testing.T) {
			if x := tc.sda.Front(); x != tc.want {
				t.Errorf("got %d; want %d", x, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_SetFront(t *testing.T) {
	testCases := []struct {
		data []int
		x    int
		want []int
	}{
		{[]int{0}, 3, []int{3}},
		{[]int{0, 1}, 3, []int{3, 1}},
		{[]int{0, 1, 2}, 3, []int{3, 1, 2}},
		{[]int{1}, 4, []int{4}},
		{[]int{1, 2}, 4, []int{4, 2}},
		{[]int{1, 2, 3}, 4, []int{4, 2, 3}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("s=%s&x=%d", sliceToName(tc.data), tc.x), func(t *testing.T) {
			sda := copySda(tc.data)
			sda.SetFront(tc.x)
			if sliceUnequal(sda, tc.want) {
				t.Errorf("got %v; want %v", sda, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Back(t *testing.T) {
	testCases := []struct {
		sda  IntSDA
		want int
	}{
		{IntSDA{0}, 0},
		{IntSDA{0, 1}, 1},
		{IntSDA{0, 1, 2}, 2},
		{IntSDA{1}, 1},
		{IntSDA{1, 2}, 2},
		{IntSDA{1, 2, 3}, 3},
	}

	for _, tc := range testCases {
		t.Run("s="+sliceToName(tc.sda), func(t *testing.T) {
			if x := tc.sda.Back(); x != tc.want {
				t.Errorf("got %d; want %d", x, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_SetBack(t *testing.T) {
	testCases := []struct {
		data []int
		x    int
		want []int
	}{
		{[]int{0}, 3, []int{3}},
		{[]int{0, 1}, 3, []int{0, 3}},
		{[]int{0, 1, 2}, 3, []int{0, 1, 3}},
		{[]int{1}, 4, []int{4}},
		{[]int{1, 2}, 4, []int{1, 4}},
		{[]int{1, 2, 3}, 4, []int{1, 2, 4}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("s=%s&x=%d", sliceToName(tc.data), tc.x), func(t *testing.T) {
			sda := copySda(tc.data)
			sda.SetBack(tc.x)
			if sliceUnequal(sda, tc.want) {
				t.Errorf("got %v; want %v", sda, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Reverse(t *testing.T) {
	testCases := []struct {
		data, want []int
	}{
		{nil, nil},
		{[]int{}, []int{}},
		{[]int{0}, []int{0}},
		{[]int{0, 1}, []int{1, 0}},
		{[]int{0, 1, 2}, []int{2, 1, 0}},
		{[]int{0, 1, 2, 3}, []int{3, 2, 1, 0}},
	}

	for _, tc := range testCases {
		t.Run("s="+sliceToName(tc.data), func(t *testing.T) {
			sda := copySda(tc.data)
			sda.Reverse()
			if sliceUnequal(sda, tc.want) {
				t.Errorf("got %v; want %v", sda, tc.want)
			}
		})
	}

	var nilSDA *IntSDA
	t.Run("s="+sdaPtrToName(nilSDA), func(t *testing.T) {
		nilSDA.Reverse()
		if nilSDA != nil {
			t.Errorf("got %v; want <nil>", nilSDA)
		}
	})
}

func TestSliceDynamicArray_Get(t *testing.T) {
	sda := IntSDA{2, 3, 4, 5, 6}
	testCases := []struct {
		i, want int
	}{
		{0, 2},
		{1, 3},
		{2, 4},
		{3, 5},
		{4, 6},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("i=%d", tc.i), func(t *testing.T) {
			if x := sda.Get(tc.i); x != tc.want {
				t.Errorf("got %d; want %d", x, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Set(t *testing.T) {
	data := []int{0, 1, 2, 3, 4}
	testCases := []struct {
		i, x int
		want []int
	}{
		{0, 2, []int{2, 1, 2, 3, 4}},
		{1, 3, []int{0, 3, 2, 3, 4}},
		{2, 4, []int{0, 1, 4, 3, 4}},
		{3, 5, []int{0, 1, 2, 5, 4}},
		{4, 6, []int{0, 1, 2, 3, 6}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("i=%d&x=%d", tc.i, tc.x), func(t *testing.T) {
			sda := copySda(data)
			sda.Set(tc.i, tc.x)
			if sliceUnequal(sda, tc.want) {
				t.Errorf("got %v; want %v", sda, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Swap(t *testing.T) {
	data := []int{0, 1, 2, 3, 4}
	n := len(data)

	testCases := make([]struct {
		i, j int
		want []int
	}, n*n)
	var idx int
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			testCases[idx].i, testCases[idx].j = i, j
			want := make([]int, n)
			copy(want, data)
			want[i], want[j] = want[j], want[i]
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("i=%d&j=%d", tc.i, tc.j), func(t *testing.T) {
			sda := copySda(data)
			sda.Swap(tc.i, tc.j)
			if sliceUnequal(sda, tc.want) {
				t.Errorf("got %v; want %v", sda, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Slice(t *testing.T) {
	sda := IntSDA{0, 1, 2, 3, 4}
	n := len(sda)

	testCases := make([]struct {
		begin, end int
		want       []int
	}, n*(n+3)/2) // (n+1)+n+(n-1)+...+2 = n*((n+1)+2)/2 = n*(n+3)/2
	var idx int
	for begin := 0; begin < n; begin++ {
		for end := begin; end <= n; end++ {
			testCases[idx].begin, testCases[idx].end = begin, end
			want := make([]int, end-begin)
			copy(want, sda[begin:end])
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("begin=%d&end=%d", tc.begin, tc.end), func(t *testing.T) {
			slice := sda.Slice(tc.begin, tc.end)
			s := make([]int, slice.Len())
			var i int
			slice.Range(func(x int) (cont bool) {
				s[i], i = x, i+1
				return true
			})
			if sliceUnequal(s, tc.want) {
				t.Errorf("got %v; want %v", s, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Filter(t *testing.T) {
	dataList := [][]int{nil, {}, {0}, {-1, 0, 1}, {-2, -1, 0, 1, 2}}
	filterList := []func(x int) (keep bool){
		func(x int) (keep bool) {
			return x >= 0
		},
		func(x int) (keep bool) {
			return x%2 == 0
		},
	}

	testCases := make([]struct {
		data      []int
		filterIdx int
		want      []int
	}, len(dataList)*len(filterList))
	var idx int
	for _, data := range dataList {
		for filterIdx, filter := range filterList {
			testCases[idx].data = data
			testCases[idx].filterIdx = filterIdx
			if len(data) > 0 {
				want := make([]int, 0, len(data))
				for _, x := range data {
					if filter(x) {
						want = append(want, x)
					}
				}
				testCases[idx].want = want
			} else if data != nil {
				testCases[idx].want = data
			}
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("s=%s&filterIdx=%d", sliceToName(tc.data), tc.filterIdx), func(t *testing.T) {
			sda := copySda(tc.data)
			var ptrBefore, ptrAfter *[0]int
			ptrBefore = (*[0]int)(sda)
			sda.Filter(filterList[tc.filterIdx])
			ptrAfter = (*[0]int)(sda)
			if sliceUnequal(sda, tc.want) {
				t.Errorf("got %v; want %v", sda, tc.want)
			}
			if ptrAfter != ptrBefore {
				t.Error("allocated new array, not in-place")
			}
		})
	}
}

func TestSliceDynamicArray_Filter_NilAndEmpty(t *testing.T) {
	sdas := []*IntSDA{nil, new(IntSDA), {}}
	for _, sda := range sdas {
		t.Run("s="+sdaPtrToName(sda), func(t *testing.T) {
			sda.Filter(func(x int) (keep bool) {
				t.Error("handler was called, x:", x)
				return true
			})
		})
	}
}

func TestSliceDynamicArray_Cap(t *testing.T) {
	sda1 := make(IntSDA, 0, 3)
	sda2 := IntSDA{1, 2, 3}[:1]
	testCases := []struct {
		sda  *IntSDA
		want int
	}{
		{nil, 0},
		{new(IntSDA), 0},
		{&IntSDA{}, 0},
		{&sda1, 3},
		{&IntSDA{1}, 1},
		{&sda2, 3},
	}

	for _, tc := range testCases {
		var sdaCap int
		if tc.sda != nil {
			sdaCap = cap(*tc.sda)
		}
		t.Run(fmt.Sprintf("s=%s&cap=%d", sdaPtrToName(tc.sda), sdaCap), func(t *testing.T) {
			if c := tc.sda.Cap(); c != tc.want {
				t.Errorf("got %d; want %d", c, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Push(t *testing.T) {
	testCases := []struct {
		data []int
		x    int
		want []int
	}{
		{nil, 0, []int{0}},
		{[]int{}, 0, []int{0}},
		{[]int{0}, 1, []int{0, 1}},
		{[]int{0, 1}, 2, []int{0, 1, 2}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("s=%s&x=%d", sliceToName(tc.data), tc.x), func(t *testing.T) {
			sda := copySda(tc.data)
			sda.Push(tc.x)
			if sliceUnequal(sda, tc.want) {
				t.Errorf("got %v; want %v", sda, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Pop(t *testing.T) {
	testCases := []struct {
		data, wantSda []int
		wantX         int
	}{
		{[]int{0}, []int{}, 0},
		{[]int{1}, []int{}, 1},
		{[]int{0, 1}, []int{0}, 1},
		{[]int{0, 1, 2}, []int{0, 1}, 2},
	}

	for _, tc := range testCases {
		t.Run("s="+sliceToName(tc.data), func(t *testing.T) {
			sda := copySda(tc.data)
			x := sda.Pop()
			if sliceUnequal(sda, tc.wantSda) || x != tc.wantX {
				t.Errorf("got %v, %d; want %v, %d", sda, x, tc.wantSda, tc.wantX)
			}
		})
	}
}

func TestSliceDynamicArray_Append(t *testing.T) {
	dataList := [][]int{nil, {}, {0}, {0, 1}, {0, 1, 2}}
	seqList := []sequence.Sequence[int]{
		nil,
		&IntSDA{},
		&IntSDA{3},
		&IntSDA{3, 4},
		&IntSDA{3, 4, 5},
		newSequence([]int{}),
		newSequence([]int{3}),
		newSequence([]int{3, 4}),
		newSequence([]int{3, 4, 5}),
	}

	testCases := make([]struct {
		data []int
		seq  sequence.Sequence[int]
		want []int
	}, len(dataList)*len(seqList))
	var idx int
	for _, data := range dataList {
		for _, seq := range seqList {
			testCases[idx].data = data
			testCases[idx].seq = seq
			seqLen := sequenceLen(seq)
			if seqLen > 0 {
				want := make([]int, len(data)+seqLen)
				i := copy(want, data)
				seq.Range(func(x int) (cont bool) {
					want[i], i = x, i+1
					return true
				})
				testCases[idx].want = want
			} else {
				testCases[idx].want = data
			}
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("s=%s&seq=%s", sliceToName(tc.data), sequenceToName(tc.seq)), func(t *testing.T) {
			sda := copySda(tc.data)
			sda.Append(tc.seq)
			if sliceUnequal(sda, tc.want) {
				t.Errorf("got %v; want %v", sda, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Truncate(t *testing.T) {
	data := []int{0, 1, 2, 3, 4}
	n := len(data)

	testCases := make([]struct {
		i    int
		want []int
	}, n+2)
	testCases[0].i, testCases[0].want = -1, data
	idx := 1
	for i := 0; i < n; i++ {
		testCases[idx].i = i
		testCases[idx].want = data[:i]
		idx++
	}
	testCases[idx].i, testCases[idx].want = n, data

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("i=%d", tc.i), func(t *testing.T) {
			sda := copySda(data)
			sda.Truncate(tc.i)
			if sliceUnequal(sda, tc.want) {
				t.Errorf("got %v; want %v", sda, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Insert(t *testing.T) {
	dataList := [][]int{nil, {}, {0}, {0, 1}, {0, 1, 2}}
	x := 3

	var numTestCase int
	for _, data := range dataList {
		numTestCase += len(data) + 1
	}
	testCases := make([]struct {
		data []int
		i    int
		want []int
	}, numTestCase)
	var idx int
	for _, data := range dataList {
		for i := 0; i <= len(data); i++ {
			testCases[idx].data = data
			testCases[idx].i = i
			want := make([]int, len(data)+1)
			copy(want, data[:i])
			want[i] = x
			copy(want[i+1:], data[i:])
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("s=%s&i=%d", sliceToName(tc.data), tc.i), func(t *testing.T) {
			sda := copySda(tc.data)
			sda.Insert(tc.i, x)
			if sliceUnequal(sda, tc.want) {
				t.Errorf("got %v; want %v", sda, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Remove(t *testing.T) {
	data := []int{0, 1, 2, 3, 4}
	testCases := []struct {
		i       int
		wantSda []int
		wantX   int
	}{
		{0, []int{1, 2, 3, 4}, 0},
		{1, []int{0, 2, 3, 4}, 1},
		{2, []int{0, 1, 3, 4}, 2},
		{3, []int{0, 1, 2, 4}, 3},
		{4, []int{0, 1, 2, 3}, 4},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("i=%d", tc.i), func(t *testing.T) {
			sda := copySda(data)
			x := sda.Remove(tc.i)
			if sliceUnequal(sda, tc.wantSda) || x != tc.wantX {
				t.Errorf("got %v, %d; want %v, %d", sda, x, tc.wantSda, tc.wantX)
			}
		})
	}
}

func TestSliceDynamicArray_RemoveWithoutOrder(t *testing.T) {
	data := []int{0, 1, 2, 3, 4}
	testCases := []struct {
		i       int
		wantSda []int
		wantX   int
	}{
		{0, []int{1, 2, 3, 4}, 0},
		{1, []int{0, 2, 3, 4}, 1},
		{2, []int{0, 1, 3, 4}, 2},
		{3, []int{0, 1, 2, 4}, 3},
		{4, []int{0, 1, 2, 3}, 4},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("i=%d", tc.i), func(t *testing.T) {
			sda := copySda(data)
			x := sda.RemoveWithoutOrder(tc.i)
			if sliceUnequalWithoutOrder(sda, tc.wantSda) || x != tc.wantX {
				t.Errorf("got %v, %d; want %v, %d", sda, x, tc.wantSda, tc.wantX)
			}
		})
	}
}

func TestSliceDynamicArray_InsertSequence(t *testing.T) {
	dataList := [][]int{nil, {}, {0}, {0, 1}, {0, 1, 2}}
	seqList := []sequence.Sequence[int]{
		nil,
		&IntSDA{},
		&IntSDA{3},
		&IntSDA{3, 4},
		&IntSDA{3, 4, 5},
		newSequence([]int{}),
		newSequence([]int{3}),
		newSequence([]int{3, 4}),
		newSequence([]int{3, 4, 5}),
	}

	var numTestCase int
	for _, data := range dataList {
		numTestCase += len(data) + 1
	}
	numTestCase *= len(seqList)
	testCases := make([]struct {
		data []int
		i    int
		seq  sequence.Sequence[int]
		want []int
	}, numTestCase)
	var idx int
	for _, data := range dataList {
		for _, seq := range seqList {
			for i := 0; i <= len(data); i++ {
				testCases[idx].data = data
				testCases[idx].i = i
				testCases[idx].seq = seq
				seqLen := sequenceLen(seq)
				if seqLen > 0 {
					want := make([]int, len(data)+seqLen)
					j := copy(want, data[:i])
					seq.Range(func(x int) (cont bool) {
						want[j], j = x, j+1
						return true
					})
					copy(want[j:], data[i:])
					testCases[idx].want = want
				} else {
					testCases[idx].want = data
				}
				idx++
			}
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("s=%s&i=%d&seq=%s", sliceToName(tc.data), tc.i, sequenceToName(tc.seq)),
			func(t *testing.T) {
				sda := copySda(tc.data)
				sda.InsertSequence(tc.i, tc.seq)
				if sliceUnequal(sda, tc.want) {
					t.Errorf("got %v; want %v", sda, tc.want)
				}
			})
	}
}

func TestSliceDynamicArray_Cut(t *testing.T) {
	data := []int{0, 1, 2, 3, 4}
	n := len(data)

	testCases := make([]struct {
		begin, end int
		want       []int
	}, n*(n+3)/2) // (n+1)+n+(n-1)+...+2 = n*((n+1)+2)/2 = n*(n+3)/2
	var idx int
	for begin := 0; begin < n; begin++ {
		for end := begin; end <= n; end++ {
			testCases[idx].begin, testCases[idx].end = begin, end
			want := make([]int, n-end+begin)
			copy(want, data[:begin])
			copy(want[begin:], data[end:])
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("begin=%d&end=%d", tc.begin, tc.end), func(t *testing.T) {
			sda := copySda(data)
			sda.Cut(tc.begin, tc.end)
			if sliceUnequal(sda, tc.want) {
				t.Errorf("got %v; want %v", sda, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_CutWithoutOrder(t *testing.T) {
	data := []int{0, 1, 2, 3, 4}
	n := len(data)

	testCases := make([]struct {
		begin, end int
		want       []int
	}, n*(n+3)/2) // (n+1)+n+(n-1)+...+2 = n*((n+1)+2)/2 = n*(n+3)/2
	var idx int
	for begin := 0; begin < n; begin++ {
		for end := begin; end <= n; end++ {
			testCases[idx].begin, testCases[idx].end = begin, end
			want := make([]int, n-end+begin)
			copy(want, data[:begin])
			copy(want[begin:], data[end:])
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("begin=%d&end=%d", tc.begin, tc.end), func(t *testing.T) {
			sda := copySda(data)
			sda.CutWithoutOrder(tc.begin, tc.end)
			if sliceUnequalWithoutOrder(sda, tc.want) {
				t.Errorf("got %v; want %v", sda, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Extend(t *testing.T) {
	dataList := [][]int{nil, {}, {0}, {0, 1}, {0, 1, 2}}
	nList := []int{0, 1, 2, 3}

	testCases := make([]struct {
		data []int
		n    int
		want []int
	}, len(dataList)*len(nList))
	var idx int
	for _, data := range dataList {
		for _, n := range nList {
			testCases[idx].data = data
			testCases[idx].n = n
			if data != nil || n > 0 {
				want := make([]int, len(data)+n)
				copy(want, data)
				testCases[idx].want = want
			}
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("s=%s&n=%d", sliceToName(tc.data), tc.n), func(t *testing.T) {
			sda := copySda(tc.data)
			sda.Extend(tc.n)
			if sliceUnequal(sda, tc.want) {
				t.Errorf("got %v; want %v", sda, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Expand(t *testing.T) {
	dataList := [][]int{nil, {}, {0}, {0, 1}, {0, 1, 2}}
	nList := []int{0, 1, 2, 3}

	var numTestCase int
	for _, data := range dataList {
		numTestCase += len(data) + 1
	}
	numTestCase *= len(nList)
	testCases := make([]struct {
		data []int
		i, n int
		want []int
	}, numTestCase)
	var idx int
	for _, data := range dataList {
		for _, n := range nList {
			for i := 0; i <= len(data); i++ {
				testCases[idx].data = data
				testCases[idx].i, testCases[idx].n = i, n
				if data != nil || n > 0 {
					want := make([]int, len(data)+n)
					copy(want, data[:i])
					copy(want[i+n:], data[i:])
					testCases[idx].want = want
				}
				idx++
			}
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("s=%s&i=%d&n=%d", sliceToName(tc.data), tc.i, tc.n), func(t *testing.T) {
			sda := copySda(tc.data)
			sda.Expand(tc.i, tc.n)
			if sliceUnequal(sda, tc.want) {
				t.Errorf("got %v; want %v", sda, tc.want)
			}
		})
	}
}

func TestSliceDynamicArray_Reserve(t *testing.T) {
	dataList := [][]int{nil, {}, {0}, {0, 1}, {0, 1, 2}}
	capList := []int{-1, 0, 1, 2, 3, 4}

	testCases := make([]struct {
		data     []int
		capacity int
		wantSda  []int
		wantCap  int
	}, len(dataList)*len(capList))
	var idx int
	for _, data := range dataList {
		for _, capacity := range capList {
			testCases[idx].data = data
			testCases[idx].capacity = capacity
			if capacity <= cap(data) {
				testCases[idx].wantSda = data
				testCases[idx].wantCap = cap(data)
			} else {
				if data != nil {
					testCases[idx].wantSda = data
				} else {
					testCases[idx].wantSda = []int{}
				}
				testCases[idx].wantCap = capacity
			}
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("s=%s&cap=%d", sliceToName(tc.data), tc.capacity), func(t *testing.T) {
			sda := copySda(tc.data)
			sda.Reserve(tc.capacity)
			if c := cap(sda); c != tc.wantCap {
				t.Errorf("got capacity %d; want %d", c, tc.wantCap)
			}
			if sliceUnequal(sda, tc.wantSda) {
				t.Errorf("data changed; got %v; want %v", sda, tc.wantSda)
			}
		})
	}
}

func TestSliceDynamicArray_Shrink(t *testing.T) {
	dataList := [][]int{
		nil, {}, {0}, {0, 1}, {0, 1, 2},
		make([]int, 0, 1), make([]int, 0, 2),
		make([]int, 1, 2), make([]int, 1, 3),
		make([]int, 2, 3), make([]int, 2, 4),
	}

	testCases := make([]struct {
		data       []int
		wantCap    int
		isNewArray bool
	}, len(dataList))
	var idx int
	for _, data := range dataList {
		testCases[idx].data = data
		testCases[idx].wantCap = len(data)
		testCases[idx].isNewArray = len(data) < cap(data)
		idx++
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("s=%s(cap=%d)", sliceToName(tc.data), cap(tc.data)), func(t *testing.T) {
			sda := copySda(tc.data)
			var ptrBefore, ptrAfter *[0]int
			ptrBefore = (*[0]int)(sda)
			sda.Shrink()
			ptrAfter = (*[0]int)(sda)
			if c := cap(sda); c != tc.wantCap {
				t.Errorf("got capacity %d; want %d", c, tc.wantCap)
			}
			if sliceUnequal(sda, tc.data) {
				t.Errorf("data changed; got %v; want %v", sda, tc.data)
			}
			if tc.isNewArray {
				if ptrAfter == ptrBefore {
					t.Error("should allocate a new array")
				}
			} else if ptrAfter != ptrBefore {
				t.Error("should keep using the old array")
			}
		})
	}
}

func TestSliceDynamicArray_Clear(t *testing.T) {
	dataList := [][]int{nil, {}, {0}, {0, 1}, {0, 1, 2}}
	for _, data := range dataList {
		t.Run("s="+sliceToName(data), func(t *testing.T) {
			sda := copySda(data)
			sda.Clear()
			if sda != nil {
				t.Errorf("got %v; want <nil>", sda)
			}
		})
	}

	var nilSDA *IntSDA
	t.Run("s="+sdaPtrToName(nilSDA), func(t *testing.T) {
		nilSDA.Clear()
		if nilSDA != nil {
			t.Errorf("got %v; want <nil>", nilSDA)
		}
	})
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

func sdaPtrToName[T any](p *array.SliceDynamicArray[T]) string {
	if p == nil {
		return fmt.Sprintf("(%T)<nil>", p)
	}
	return sliceToName(*p)
}

func sequenceToName[Item any](s sequence.Sequence[Item]) string {
	typeStr := fmt.Sprintf("(%T)", s)
	if s == nil {
		return typeStr + "<nil>"
	}
	var b strings.Builder
	b.Grow(len(typeStr) + s.Len()*3 + 2)
	b.WriteString(typeStr)
	b.WriteByte('[')
	var notFirst bool
	s.Range(func(x Item) (cont bool) {
		if notFirst {
			b.WriteByte(',')
		} else {
			notFirst = true
		}
		_, _ = fmt.Fprintf(&b, "%v", x) // ignore error as error is always nil
		return true
	})
	b.WriteByte(']')
	return b.String()
}

func sequenceLen[Item any](s sequence.Sequence[Item]) int {
	if s == nil {
		return 0
	}
	return s.Len()
}

func copySda[Item any](data []Item) array.SliceDynamicArray[Item] {
	if data == nil {
		return nil
	}
	sda := make(array.SliceDynamicArray[Item], len(data), cap(data))
	copy(sda, data)
	return sda
}

func sliceUnequal[T comparable](a, b []T) bool {
	if a == nil {
		return b != nil
	}
	return b == nil || !compare.ComparableSliceEqual(a, b)
}

func sliceUnequalWithoutOrder[T comparable](a, b []T) bool {
	if a == nil {
		return b != nil
	}
	return b == nil || !compare.ComparableSliceEqualWithoutOrder(a, b)
}

// sequenceImpl is a linked list-based implementation of interface
// github.com/donyori/gogo/container/sequence.Sequence.
type sequenceImpl[Item any] struct {
	linkedList      list.List
	useFrontAndNext bool
}

func newSequence[Item any](data []Item) *sequenceImpl[Item] {
	seq := &sequenceImpl[Item]{useFrontAndNext: true}
	for _, x := range data {
		seq.linkedList.PushBack(x)
	}
	return seq
}

func (s *sequenceImpl[Item]) Len() int {
	if s == nil {
		return 0
	}
	return s.linkedList.Len()
}

func (s *sequenceImpl[Item]) Front() Item {
	if s.Len() == 0 {
		panic(errors.AutoMsg("sequence is nil or empty"))
	} else if s.useFrontAndNext {
		return s.linkedList.Front().Value.(Item)
	}
	return s.linkedList.Back().Value.(Item)
}

func (s *sequenceImpl[Item]) SetFront(x Item) {
	switch {
	case s.Len() == 0:
		panic(errors.AutoMsg("sequence is nil or empty"))
	case s.useFrontAndNext:
		s.linkedList.Front().Value = x
	default:
		s.linkedList.Back().Value = x
	}
}

func (s *sequenceImpl[Item]) Back() Item {
	if s.Len() == 0 {
		panic(errors.AutoMsg("sequence is nil or empty"))
	} else if s.useFrontAndNext {
		return s.linkedList.Back().Value.(Item)
	}
	return s.linkedList.Front().Value.(Item)
}

func (s *sequenceImpl[Item]) SetBack(x Item) {
	switch {
	case s.Len() == 0:
		panic(errors.AutoMsg("sequence is nil or empty"))
	case s.useFrontAndNext:
		s.linkedList.Back().Value = x
	default:
		s.linkedList.Front().Value = x
	}
}

func (s *sequenceImpl[Item]) Reverse() {
	if s.Len() > 0 {
		s.useFrontAndNext = !s.useFrontAndNext
	}
}

func (s *sequenceImpl[Item]) Range(handler func(x Item) (cont bool)) {
	switch {
	case s.Len() == 0:
		return
	case s.useFrontAndNext:
		elem := s.linkedList.Front()
		for elem != nil {
			if !handler(elem.Value.(Item)) {
				return
			}
			elem = elem.Next()
		}
	default:
		elem := s.linkedList.Back()
		for elem != nil {
			if !handler(elem.Value.(Item)) {
				return
			}
			elem = elem.Prev()
		}
	}
}
