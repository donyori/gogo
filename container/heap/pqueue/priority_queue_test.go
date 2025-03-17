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

package pqueue_test

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/donyori/gogo/container/heap/pqueue"
	"github.com/donyori/gogo/fmtcoll"
	"github.com/donyori/gogo/function/compare"
)

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

func TestNew(t *testing.T) {
	var n int
	for _, data := range dataList {
		n += len(data) + 3
	}

	testCases := make([]struct {
		data           []int
		capacity       int
		wantSortedData []int
		wantInitCap    int
	}, n)
	var idx int
	for _, data := range dataList {
		for c := -1; c <= len(data)+1; c++ {
			testCases[idx].data = data
			testCases[idx].capacity = c
			testCases[idx].wantSortedData = copyAndSort(data)
			if c > 0 {
				testCases[idx].wantInitCap = c
			} else {
				testCases[idx].wantInitCap = pqueue.DefaultCapacity
			}
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("cap=%d&data=%s", tc.capacity, sliceToName(tc.data)),
			func(t *testing.T) {
				pq := pqueue.New(IntLess, tc.capacity)
				if c := pq.Cap(); c != tc.wantInitCap {
					t.Errorf("got initial capacity %d; want %d",
						c, tc.wantInitCap)
				}
				pq.Enqueue(tc.data...)
				checkPriorityQueueByDequeue(t, pq, tc.wantSortedData)
			},
		)
	}
}

func TestPriorityQueue_Len(t *testing.T) {
	for _, data := range dataList {
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pq := newPriorityQueue(data)
			if n := pq.Len(); n != len(data) {
				t.Errorf("got %d; want %d", n, len(data))
			}
		})
	}
}

func TestPriorityQueue_Range(t *testing.T) {
	for _, data := range dataList {
		counterMap := make(map[int]int, len(data))
		for _, x := range data {
			counterMap[x]++
		}
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pq := newPriorityQueue(data)
			pq.Range(func(x int) (cont bool) {
				counterMap[x]--
				return true
			})
			for x, ctr := range counterMap {
				if ctr > 0 {
					t.Error("insufficient accesses to", x)
				} else if ctr < 0 {
					t.Error("too many accesses to", x)
				}
			}
		})
	}
}

func TestPriorityQueue_Range_NilHandler(t *testing.T) {
	pq := newPriorityQueue(dataList[len(dataList)-1])
	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()
	pq.Range(nil)
}

func TestPriorityQueue_IterItems(t *testing.T) {
	for _, data := range dataList {
		counterMap := make(map[int]int, len(data))
		for _, x := range data {
			counterMap[x]++
		}
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pq := newPriorityQueue(data)
			seq := pq.IterItems()
			if seq == nil {
				t.Fatal("got nil iterator")
			}
			counterMapCopy := maps.Clone(counterMap)
			for x := range seq {
				counterMap[x]--
			}
			for x, ctr := range counterMap {
				if ctr > 0 {
					t.Error("insufficient accesses to", x)
				} else if ctr < 0 {
					t.Error("too many accesses to", x)
				}
			}
			// Rewind the iterator and test it again.
			for x := range seq {
				counterMapCopy[x]--
			}
			for x, ctr := range counterMapCopy {
				if ctr > 0 {
					t.Error("rewind - insufficient accesses to", x)
				} else if ctr < 0 {
					t.Error("rewind - too many accesses to", x)
				}
			}
		})
	}
}

func TestPriorityQueue_Clear(t *testing.T) {
	for _, data := range dataList {
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pq := newPriorityQueue(data)
			pq.Clear()
			checkPriorityQueueByDequeue(t, pq, nil)
		})
	}
}

func TestPriorityQueue_RemoveAll(t *testing.T) {
	for _, data := range dataList {
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pq := newPriorityQueue(data)
			pq.RemoveAll()
			checkPriorityQueueByDequeue(t, pq, nil)
		})
	}
}

func TestPriorityQueue_Cap(t *testing.T) {
	for _, data := range dataList {
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pq := newPriorityQueue(data)
			if c := pq.Cap(); c < len(data) {
				t.Errorf("got %d; want >= %d", c, len(data))
			}
		})
	}
}

func TestPriorityQueue_Reserve(t *testing.T) {
	var n int
	for _, data := range dataList {
		n += len(data) + 3
	}

	testCases := make([]struct {
		data           []int
		capacity       int
		wantSortedData []int
		wantCap        int
	}, n)
	var idx int
	for _, data := range dataList {
		for c := -1; c <= len(data); c++ {
			testCases[idx].data = data
			testCases[idx].capacity = c
			testCases[idx].wantSortedData = copyAndSort(data)
			if c > 0 {
				testCases[idx].wantCap = len(data)
			} else {
				testCases[idx].wantCap = max(pqueue.DefaultCapacity, len(data))
			}
			idx++
		}
		c := max(len(data)<<2, 256)
		testCases[idx].data = data
		testCases[idx].capacity = c
		testCases[idx].wantSortedData = copyAndSort(data)
		testCases[idx].wantCap = c
		idx++
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("cap=%d&data=%s", tc.capacity, sliceToName(tc.data)),
			func(t *testing.T) {
				pq := newPriorityQueue(tc.data)
				pq.Reserve(tc.capacity)
				if c := pq.Cap(); c != tc.wantCap {
					t.Errorf("got capacity %d; want %d", c, tc.wantCap)
				}
				checkPriorityQueueByDequeue(t, pq, tc.wantSortedData)
			},
		)
	}
}

func TestPriorityQueue_Enqueue(t *testing.T) {
	xsList := [][]int{nil, {}, {-1}, {0}, {1}, {7}, {-1, 0, 1}, {0, 0, 7}}

	testCases := make([]struct {
		data, xs, want []int
	}, len(dataList)*len(xsList))
	var idx int
	for _, data := range dataList {
		for _, xs := range xsList {
			testCases[idx].data = data
			testCases[idx].xs = xs
			want := make([]int, len(data)+len(xs))
			copy(want[copy(want, data):], xs)
			slices.Sort(want)
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&xs=%s",
				sliceToName(tc.data), sliceToName(tc.xs)),
			func(t *testing.T) {
				pq := newPriorityQueue(tc.data)
				pq.Enqueue(tc.xs...)
				checkPriorityQueueByDequeue(t, pq, tc.want)
			},
		)
	}
}

func TestPriorityQueue_Dequeue(t *testing.T) {
	for _, data := range dataList {
		if len(data) == 0 {
			continue
		}
		sorted := copyAndSort(data)
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pq := newPriorityQueue(data)
			if x := pq.Dequeue(); x != sorted[0] {
				t.Errorf("got %d; want %d", x, sorted[0])
			}
			checkPriorityQueueByDequeue(t, pq, sorted[1:])
		})
	}
}

func TestPriorityQueue_Top(t *testing.T) {
	for _, data := range dataList {
		if len(data) == 0 {
			continue
		}
		minX := data[0]
		for i := 1; i < len(data); i++ {
			if minX > data[i] {
				minX = data[i]
			}
		}
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pq := newPriorityQueue(data)
			if x := pq.Top(); x != minX {
				t.Errorf("got %d; want %d", x, minX)
			}
		})
	}
}

func TestPriorityQueue_ReplaceTop(t *testing.T) {
	newXList := []int{-1, 0, 1, 2, 3, 4, 5, 6, 7}

	numTestCase := len(dataList)
	for _, data := range dataList {
		if len(data) == 0 {
			numTestCase--
		}
	}
	numTestCase *= len(newXList)
	testCases := make([]struct {
		data []int
		newX int
		want []int
	}, numTestCase)
	var idx int
	for _, data := range dataList {
		if len(data) == 0 {
			continue
		}
		for _, newX := range newXList {
			testCases[idx].data = data
			testCases[idx].newX = newX
			want := copyAndSort(data)
			want[0] = newX
			slices.Sort(want)
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("data=%s&newX=%d", sliceToName(tc.data), tc.newX),
			func(t *testing.T) {
				pq := newPriorityQueue(tc.data)
				if x := pq.ReplaceTop(tc.newX); x != tc.want[0] {
					t.Errorf("got %d; want %d", x, tc.want[0])
				}
				checkPriorityQueueByDequeue(t, pq, tc.want)
			},
		)
	}
}

func newPriorityQueue(data []int) pqueue.PriorityQueue[int] {
	pq := pqueue.New(IntLess, len(data))
	pq.Enqueue(data...)
	return pq
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

// !!pq may be modified in this function.
//
// want must be sorted.
func checkPriorityQueueByDequeue[Item comparable](
	t *testing.T,
	pq pqueue.PriorityQueue[Item],
	want []Item,
) {
	var i int
	defer func() {
		if e := recover(); e != nil {
			prefix := fmt.Sprintf("panic after dequeuing %d item:", i+1)
			if i != 0 {
				prefix += "s"
			}
			if isDequeuePanicMessage(e) {
				t.Error(prefix, "priority queue length <", len(want))
			} else {
				t.Error(prefix, e)
			}
		}
	}()
	for i = 0; i < len(want); i++ {
		if x := pq.Dequeue(); x != want[i] {
			t.Errorf("i:%d, got %v; want %v", i, x, want[i])
			return
		}
	}
	defer func() {
		if e := recover(); !isDequeuePanicMessage(e) {
			t.Error(e)
		}
	}()
	x := pq.Dequeue() // want panic here
	t.Errorf("dequeued more than %d items, got %v", i, x)
}

func isDequeuePanicMessage(err any) bool {
	if err == nil {
		return false
	}
	msg, ok := err.(string)
	return ok && strings.HasSuffix(msg, pqueue.EmptyQueuePanicMessage)
}
