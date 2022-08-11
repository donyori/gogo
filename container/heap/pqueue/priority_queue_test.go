// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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
	"sort"
	"strings"
	"testing"

	"github.com/donyori/gogo/container/heap/pqueue"
	"github.com/donyori/gogo/function/compare"
)

var intLess = compare.OrderedLess[int]

func TestPriorityQueueVersionAssertion(t *testing.T) {
	t.Run("PriorityQueueBasic to PriorityQueue", func(t *testing.T) {
		pqb := pqueue.NewPriorityQueueBasic(intLess)
		if _, ok := pqb.(pqueue.PriorityQueue[int]); ok {
			t.Error("can convert; but should cannot")
		}
	})
	t.Run("PriorityQueue to PriorityQueueBasic", func(t *testing.T) {
		pb := pqueue.NewPriorityQueue(intLess)
		if _, ok := pb.(pqueue.PriorityQueueBasic[int]); !ok {
			t.Error("cannot convert; but should can")
		}
	})
}

var dataList = [][]int{
	nil, {},
	{0},
	{0, 0}, {0, 1}, {1, 0},
	{0, 0, 0}, {0, 0, 1}, {0, 1, 0}, {0, 1, 1}, {1, 0, 0}, {1, 0, 1}, {1, 1, 0},
	{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0},
	{0, 1, 2, 3, 4, 5, 6}, {0, 2, 4, 6, 1, 3, 5}, {4, 5, 6, 0, 1, 2, 3},
	{3, 2, 1, 0, 4, 5, 6}, {6, 5, 4, 3, 2, 1, 0},
}

func TestNewPriorityQueueBasic(t *testing.T) {
	for _, data := range dataList {
		sorted := copyAndSort(data)
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pqb := pqueue.NewPriorityQueueBasic(intLess, data...)
			checkPriorityQueueByDequeue(t, pqb, sorted)
		})
	}
}

func TestPriorityQueueBasic_Len(t *testing.T) {
	for _, data := range dataList {
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pqb := pqueue.NewPriorityQueueBasic(intLess, data...)
			if n := pqb.Len(); n != len(data) {
				t.Errorf("got %d; want %d", n, len(data))
			}
		})
	}
}

func TestPriorityQueueBasic_Enqueue(t *testing.T) {
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
			sort.Ints(want)
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("data=%s&xs=%s", sliceToName(tc.data), sliceToName(tc.xs)), func(t *testing.T) {
			pqb := pqueue.NewPriorityQueueBasic(intLess, tc.data...)
			pqb.Enqueue(tc.xs...)
			checkPriorityQueueByDequeue(t, pqb, tc.want)
		})
	}
}

func TestPriorityQueueBasic_Dequeue(t *testing.T) {
	for _, data := range dataList {
		if len(data) == 0 {
			continue
		}
		sorted := copyAndSort(data)
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pqb := pqueue.NewPriorityQueueBasic(intLess, data...)
			if x := pqb.Dequeue(); x != sorted[0] {
				t.Errorf("got %d; want %d", x, sorted[0])
			}
			checkPriorityQueueByDequeue(t, pqb, sorted[1:])
		})
	}
}

func TestNewPriorityQueue(t *testing.T) {
	for _, data := range dataList {
		sorted := copyAndSort(data)
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pq := pqueue.NewPriorityQueue(intLess, data...)
			checkPriorityQueueByDequeue[int](t, pq, sorted)
		})
	}
}

func TestPriorityQueue_Cap(t *testing.T) {
	for _, data := range dataList {
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pq := pqueue.NewPriorityQueue(intLess, data...)
			if c := pq.Cap(); c < len(data) {
				t.Errorf("got %d; want >= %d", c, len(data))
			}
		})
	}
}

func TestPriorityQueue_Top(t *testing.T) {
	for _, data := range dataList {
		if len(data) == 0 {
			continue
		}
		min := data[0]
		for i := 1; i < len(data); i++ {
			if min > data[i] {
				min = data[i]
			}
		}
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pq := pqueue.NewPriorityQueue(intLess, data...)
			if x := pq.Top(); x != min {
				t.Errorf("got %d; want %d", x, min)
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
			sort.Ints(want)
			testCases[idx].want = want
			idx++
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("data=%s&newX=%d", sliceToName(tc.data), tc.newX), func(t *testing.T) {
			pq := pqueue.NewPriorityQueue(intLess, tc.data...)
			if x := pq.ReplaceTop(tc.newX); x != tc.want[0] {
				t.Errorf("got %d; want %d", x, tc.want[0])
			}
			checkPriorityQueueByDequeue[int](t, pq, tc.want)
		})
	}
}

func TestPriorityQueue_Clear(t *testing.T) {
	for _, data := range dataList {
		t.Run("data="+sliceToName(data), func(t *testing.T) {
			pq := pqueue.NewPriorityQueue(intLess, data...)
			pq.Clear()
			checkPriorityQueueByDequeue[int](t, pq, nil)
		})
	}
}

func sliceToName[T any](s []T) string {
	typeStr := fmt.Sprintf("(%T)", s)
	if s == nil {
		return typeStr + "<nil>"
	}
	var b strings.Builder
	b.Grow(len(typeStr) + len(s)*3 + 2)
	b.WriteString(typeStr)
	b.WriteByte('[')
	for i, x := range s {
		if i > 0 {
			b.WriteByte(',')
		}
		_, _ = fmt.Fprintf(&b, "%v", x) // ignore error as error is always nil
	}
	b.WriteByte(']')
	return b.String()
}

func copyAndSort(data []int) []int {
	if data == nil {
		return nil
	}
	sorted := make([]int, len(data))
	copy(sorted, data)
	sort.Ints(sorted)
	return sorted
}

// !!pqb may be modified in this function.
//
// want must be sorted.
func checkPriorityQueueByDequeue[Item comparable](t *testing.T,
	pqb pqueue.PriorityQueueBasic[Item], want []Item) {
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
		if x := pqb.Dequeue(); x != want[i] {
			t.Errorf("i:%d, got %v; want %v", i, x, want[i])
			return
		}
	}
	defer func() {
		if e := recover(); !isDequeuePanicMessage(e) {
			t.Error(e)
		}
	}()
	x := pqb.Dequeue() // want to panic here
	t.Errorf("dequeued more than %d items, got %v", i, x)
}

func isDequeuePanicMessage(err any) bool {
	if err == nil {
		return false
	}
	msg, ok := err.(string)
	return ok && strings.HasSuffix(msg, "priority queue is empty")
}
