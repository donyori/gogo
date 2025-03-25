// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2025  Yuan Gao
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

package queue_test

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/donyori/gogo/container/queue"
	"github.com/donyori/gogo/fmtcoll"
)

func TestGetBufferSize(t *testing.T) {
	for _, capacity := range []int{
		1, 2, 3, 4,
		6, 7, 8, 9, 10,
		14, 15, 16, 17, 18,
		math.MaxInt >> 1, math.MaxInt>>1 + 1,
	} {
		want := 1
		for want > 0 && want < capacity {
			want <<= 1
		}

		t.Run(fmt.Sprintf("cap=%d", capacity), func(t *testing.T) {
			got := queue.GetBufferSize(capacity)
			if got != want {
				t.Errorf("got %#x; want %#x", got, want)
			}
		})
	}
}

func TestGetBufferSize_Overflow(t *testing.T) {
	wantSizeInPanicMsg := uint(math.MaxInt + 1)

	for _, capacity := range []int{math.MaxInt>>1 + 2, math.MaxInt} {
		t.Run(fmt.Sprintf("cap=%d", capacity), func(t *testing.T) {
			defer func() {
				if e := recover(); e != nil {
					s, ok := e.(string)
					if !ok || !strings.HasSuffix(
						s, fmt.Sprintf("buffer size overflows; "+
							"specified capacity %d (%#[1]x), "+
							"required buffer size %d (%#[2]x) "+
							"(buffer size must be a power of two)",
							capacity, wantSizeInPanicMsg)) {
						t.Error("panic -", e)
					}
				}
			}()

			got := queue.GetBufferSize(capacity) // want panic here
			t.Errorf("want panic but got %#x", got)
		})
	}
}

func TestNew(t *testing.T) {
	for capacity := -1; capacity <= 33; capacity++ {
		wantInitCap := queue.DefaultCapacity
		if capacity > 0 {
			wantInitCap = queue.GetBufferSize(capacity)
		}
		t.Run(fmt.Sprintf("cap=%d", capacity), func(t *testing.T) {
			q := queue.New[int](capacity)
			if q == nil {
				t.Error("got nil queue")
			} else if c := q.Cap(); c != wantInitCap {
				t.Errorf("got initial capacity %d; want %d", c, wantInitCap)
			}
		})
	}
}

func TestQueue_Range(t *testing.T) {
	data := []int{0, 1, 2, 3, 3, 4, 0, 1, 2, 3, 3, 4}
	want := []int{0, 1, 2, 3, 3, 4}

	q := queue.New[int](0)
	for _, x := range data {
		q.Enqueue(x)
	}
	got := make([]int, 0, len(data))
	q.Range(func(x int) (cont bool) {
		got = append(got, x)
		return len(got) < len(data)>>1
	})
	if !slices.Equal(got, want) {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestQueue_Range_Empty(t *testing.T) {
	q := queue.New[int](0)
	q.Range(func(x int) (cont bool) {
		t.Error("handler was called, x:", x)
		return true
	})
}

func TestQueue_Range_NilHandler(t *testing.T) {
	data := []int{0, 1, 2, 3, 3, 4, 0, 1, 2, 3, 3, 4}
	q := queue.New[int](0)
	for _, x := range data {
		q.Enqueue(x)
	}
	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()
	q.Range(nil)
}

func TestQueue_IterItems(t *testing.T) {
	data := []int{0, 1, 2, 3, 3, 4, 5, 6, 7, 8, 8, 9}
	want := []int{0, 1, 2, 3, 3, 4}

	q := queue.New[int](0)
	for _, x := range data {
		q.Enqueue(x)
	}
	seq := q.IterItems()
	if seq == nil {
		t.Fatal("got nil iterator")
	}
	gotData := make([]int, 0, len(data))
	for x := range seq {
		gotData = append(gotData, x)
		if len(gotData) >= len(data)>>1 {
			break
		}
	}
	if !slices.Equal(gotData, want) {
		t.Errorf("got %v; want %v", gotData, want)
	}
	// Rewind the iterator and test it again.
	gotData = gotData[:0]
	for x := range seq {
		gotData = append(gotData, x)
		if len(gotData) >= len(data)>>1 {
			break
		}
	}
	if !slices.Equal(gotData, want) {
		t.Errorf("rewind - got %v; want %v", gotData, want)
	}
}

func TestQueue_IterItems_Empty(t *testing.T) {
	q := queue.New[int](0)
	seq := q.IterItems()
	if seq == nil {
		t.Fatal("got nil iterator")
	}
	for x := range seq {
		t.Error("yielded", x)
	}
}

func TestQueue_Reserve(t *testing.T) {
	dataList := [][]int{
		nil,
		{},
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3, 4, 5, 6},
		{0, 1, 2, 3, 4, 5, 6, 7},
		{0, 1, 2, 3, 4, 5, 6, 7, 8},
	}
	capList := []int{-1, 0, 1, 2, 3, 4, 7, 8, 9}

	testCases := make([]struct {
		data          []int
		capacity      int
		wantRangeData []int
		wantCap       int
	}, len(dataList)*len(capList))
	var idx int
	for _, data := range dataList {
		for _, capacity := range capList {
			testCases[idx].data = data
			testCases[idx].capacity = capacity
			testCases[idx].wantRangeData = slices.Clone(data)
			initCap := len(data)
			if initCap == 0 {
				initCap = queue.DefaultCapacity
			}
			testCases[idx].wantCap = capacity
			if testCases[idx].wantCap <= 0 {
				testCases[idx].wantCap = queue.DefaultCapacity
			}
			if testCases[idx].wantCap < initCap {
				testCases[idx].wantCap = initCap
			}
			testCases[idx].wantCap = queue.GetBufferSize(testCases[idx].wantCap)
			idx++
		}
	}

	for _, tc := range testCases {
		dataName := fmtcoll.MustFormatSliceToString(
			tc.data,
			&fmtcoll.SequenceFormat[int]{
				CommonFormat: fmtcoll.CommonFormat{
					Separator: ",",
				},
				FormatItemFn: fmtcoll.FprintfToFormatFunc[int]("%d"),
			},
		)
		t.Run(
			fmt.Sprintf("cap=%d&data=%s", tc.capacity, dataName),
			func(t *testing.T) {
				q := queue.New[int](len(tc.data))
				for _, x := range tc.data {
					q.Enqueue(x)
				}
				q.Reserve(tc.capacity)
				if c := q.Cap(); c != tc.wantCap {
					t.Errorf("got capacity %d; want %d", c, tc.wantCap)
				}
				rangeData := make([]int, 0, q.Len())
				q.Range(func(x int) (cont bool) {
					rangeData = append(rangeData, x)
					return true
				})
				if !slices.Equal(rangeData, tc.wantRangeData) {
					t.Errorf("got data by q.Range %v; want %v",
						rangeData, tc.wantRangeData)
				}
			},
		)
	}
}

func TestQueue_EnqueueNAndDequeueN(t *testing.T) {
	ns := make([]int, 33, 36)
	for i := range ns {
		ns[i] = i + 1
	}
	ns = append(ns, 63, 4096, 524288)

	for _, n := range ns {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			q := queue.New[int](0)
			if qn := q.Len(); qn != 0 {
				t.Fatalf("initial - q.Len() %d; want 0", qn)
			}

			testQueueEnqueueNAndDequeueNEnqueueStage(t, n, q)
			if t.Failed() {
				return
			}

			testQueueEnqueueNAndDequeueNDequeueStage(t, n, q)
			if t.Failed() {
				return
			}

			finalCap := q.Cap()
			q.RemoveAll()
			if qn := q.Len(); qn != 0 {
				t.Errorf("after q.RemoveAll() - got q.Len() %d; want 0", qn)
			}
			if c := q.Cap(); c != finalCap {
				t.Errorf("after q.RemoveAll() - got q.Cap() %d; want %d",
					c, finalCap)
			}

			q.Clear()
			if qn := q.Len(); qn != 0 {
				t.Errorf("after q.Clear() - got q.Len() %d; want 0", qn)
			}
			if c := q.Cap(); c != 0 {
				t.Errorf("after q.Clear() - got q.Cap() %d; want 0", c)
			}

			q.RemoveAll()
			if qn := q.Len(); qn != 0 {
				t.Errorf("after q.Clear() then q.RemoveAll() - got q.Len() %d; want 0",
					qn)
			}
			if c := q.Cap(); c != 0 {
				t.Errorf("after q.Clear() then q.RemoveAll() - got q.Cap() %d; want 0",
					c)
			}
		})
	}
}

// testQueueEnqueueNAndDequeueNEnqueueStage is a subprocess of
// TestQueue_EnqueueNAndDequeueN for the enqueuing stage.
func testQueueEnqueueNAndDequeueNEnqueueStage(
	t *testing.T,
	n int,
	q queue.Queue[int],
) {
	wantCap := q.Cap()
	for x := 1; !t.Failed() && x <= n; x++ {
		if wantCap < x {
			wantCap <<= 1
		}
		q.Enqueue(x)
		if qn := q.Len(); qn != x {
			t.Errorf("after q.Enqueue(%d) - got q.Len() %d; want %[1]d", x, qn)
		}
		if c := q.Cap(); c != wantCap {
			t.Errorf("after q.Enqueue(%d) - got q.Cap() %d; want %d",
				x, c, wantCap)
		}
		if front := q.Peek(); front != 1 {
			t.Errorf("after q.Enqueue(%d) - got q.Peek() %d; want 1", x, front)
		}
	}
}

// testQueueEnqueueNAndDequeueNDequeueStage is a subprocess of
// TestQueue_EnqueueNAndDequeueN for the dequeuing stage.
func testQueueEnqueueNAndDequeueNDequeueStage(
	t *testing.T,
	n int,
	q queue.Queue[int],
) {
	finalCap := q.Cap()

	for x := 1; !t.Failed() && x <= n; x++ {
		got := q.Dequeue()
		if got != x {
			t.Errorf("got No.%d q.Dequeue() %d; want %[1]d", x, got)
		}
		if qn := q.Len(); qn != n-x {
			t.Errorf("after No.%d q.Dequeue() - got q.Len() %d; want %d",
				x, qn, n-x)
		}
		if c := q.Cap(); c != finalCap {
			t.Errorf("after No.%d q.Dequeue() - got q.Cap() %d; want %d",
				x, c, finalCap)
		}
		if x < n {
			if front := q.Peek(); front != x+1 {
				t.Errorf("after No.%d q.Dequeue() - got q.Peek() %d; want %d",
					x, front, x+1)
			}
		}
	}
}

func TestQueue_RandomEnqueueAndDequeue(t *testing.T) {
	q := queue.New[int](0)
	if qn := q.Len(); qn != 0 {
		t.Fatalf("initial - q.Len() %d; want 0", qn)
	}

	random := rand.New(rand.NewChaCha8(
		[32]byte([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"))))
	var queueData []int
	var enqueueCtr, dequeueCtr int

	// Enqueue and dequeue a total of N items.
	// Each time randomly enqueue a portion of them
	// and then randomly dequeue a portion of items in the queue.

	const N int = 1 << 20
	n := N // the number of remaining items to be enqueued

	// When n >= 3, enqueue randomly 1 to (2/3)n items
	// and then randomly dequeue at least 1 item.
	for n >= 3 {
		enqueueN := 1 + random.IntN(n/3<<1)
		n -= enqueueN
		testQueueRandomEnqueueAndDequeueEnqueueStage(
			t, enqueueN, &enqueueCtr, &queueData, q)
		if t.Failed() {
			return
		}
		testQueueRandomEnqueueAndDequeueDequeueStage(
			t, 1+random.IntN(len(queueData)), &dequeueCtr, &queueData, q)
		if t.Failed() {
			return
		}
	}
	// When n < 3, enqueue all remaining items and then dequeue all items.
	testQueueRandomEnqueueAndDequeueEnqueueStage(
		t, n, &enqueueCtr, &queueData, q)
	if t.Failed() {
		return
	}
	testQueueRandomEnqueueAndDequeueDequeueStage(
		t, len(queueData), &dequeueCtr, &queueData, q)
	if t.Failed() {
		return
	}
	// An unnecessary test on enqueueCtr and dequeueCtr
	// to verify whether all the N items have been enqueued and dequeued:
	if enqueueCtr != N || dequeueCtr != N {
		t.Fatalf("got enqueueCtr %d, dequeueCtr %d; both want %d",
			enqueueCtr, dequeueCtr, N)
	}

	finalCap := q.Cap()
	q.RemoveAll()
	if qn := q.Len(); qn != 0 {
		t.Errorf("after q.RemoveAll() - got q.Len() %d; want 0", qn)
	}
	if c := q.Cap(); c != finalCap {
		t.Errorf("after q.RemoveAll() - got q.Cap() %d; want %d", c, finalCap)
	}

	q.Clear()
	if qn := q.Len(); qn != 0 {
		t.Errorf("after q.Clear() - got q.Len() %d; want 0", qn)
	}
	if c := q.Cap(); c != 0 {
		t.Errorf("after q.Clear() - got q.Cap() %d; want 0", c)
	}

	q.RemoveAll()
	if qn := q.Len(); qn != 0 {
		t.Errorf("after q.Clear() then q.RemoveAll() - got q.Len() %d; want 0",
			qn)
	}
	if c := q.Cap(); c != 0 {
		t.Errorf("after q.Clear() then q.RemoveAll() - got q.Cap() %d; want 0",
			c)
	}
}

// testQueueRandomEnqueueAndDequeueEnqueueStage is a subprocess of
// TestQueue_RandomEnqueueAndDequeue for the enqueuing stage.
func testQueueRandomEnqueueAndDequeueEnqueueStage(
	t *testing.T,
	n int,
	pEnqueueCtr *int,
	pQueueData *[]int,
	q queue.Queue[int],
) {
	wantCap := q.Cap()
	for i := 0; !t.Failed() && i < n; i++ {
		*pEnqueueCtr++
		*pQueueData = append(*pQueueData, *pEnqueueCtr)
		if wantCap < len(*pQueueData) {
			wantCap <<= 1
		}

		q.Enqueue(*pEnqueueCtr)
		if qn := q.Len(); qn != len(*pQueueData) {
			t.Errorf("after q.Enqueue(%d) - got q.Len() %d; want %d",
				*pEnqueueCtr, qn, len(*pQueueData))
		}
		if c := q.Cap(); c != wantCap {
			t.Errorf("after q.Enqueue(%d) - got q.Cap() %d; want %d",
				*pEnqueueCtr, c, wantCap)
		}
		if front := q.Peek(); front != (*pQueueData)[0] {
			t.Errorf("after q.Enqueue(%d) - got q.Peek() %d; want %d",
				*pEnqueueCtr, front, (*pQueueData)[0])
		}
	}
}

// testQueueRandomEnqueueAndDequeueDequeueStage is a subprocess of
// TestQueue_RandomEnqueueAndDequeue for the dequeuing stage.
func testQueueRandomEnqueueAndDequeueDequeueStage(
	t *testing.T,
	n int,
	pDequeueCtr *int,
	pQueueData *[]int,
	q queue.Queue[int],
) {
	wantCap := q.Cap()

	for i := 0; !t.Failed() && i < n; i++ {
		*pDequeueCtr++
		want := (*pQueueData)[0]
		*pQueueData = (*pQueueData)[1:]

		got := q.Dequeue()
		if got != want {
			t.Errorf("got No.%d q.Dequeue() %d; want %d",
				*pDequeueCtr, got, want)
		}
		if qn := q.Len(); qn != len(*pQueueData) {
			t.Errorf("after No.%d q.Dequeue() - got q.Len() %d; want %d",
				*pDequeueCtr, qn, len(*pQueueData))
		}
		if c := q.Cap(); c != wantCap {
			t.Errorf("after No.%d q.Dequeue() - got q.Cap() %d; want %d",
				*pDequeueCtr, c, wantCap)
		}
		if len(*pQueueData) > 0 {
			if front := q.Peek(); front != (*pQueueData)[0] {
				t.Errorf("after No.%d q.Dequeue() - got q.Peek() %d; want %d",
					*pDequeueCtr, front, (*pQueueData)[0])
			}
		}
	}
}

func TestQueue_EnqueueAfterClear(t *testing.T) {
	// This tests whether the queue is reusable after Clear().

	q := queue.New[int](0)
	const N int = 10
	for range N {
		q.Enqueue(1)
	}
	if n := q.Len(); n != N {
		t.Errorf("before q.Clear() - got q.Len() %d; want %d", n, N)
	}

	q.Clear()
	if n := q.Len(); n != 0 {
		t.Errorf("after q.Clear() - got q.Len() %d; want 0", n)
	}
	if c := q.Cap(); c != 0 {
		t.Errorf("after q.Clear() - got q.Cap() %d; want 0", c)
	}

	q.Enqueue(2)
	if n := q.Len(); n != 1 {
		t.Errorf("after q.Enqueue(2) - got q.Len() %d; want 1", n)
	}
	if c := q.Cap(); c != queue.DefaultCapacity {
		t.Errorf("after q.Enqueue(2) - got q.Cap() %d; want %d",
			c, queue.DefaultCapacity)
	}
	if front := q.Peek(); front != 2 {
		t.Errorf("after q.Enqueue(2) - got q.Peek() %d; want 2", front)
	}
}
