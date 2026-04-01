// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2026  Yuan Gao
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

package stack_test

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/donyori/gogo/container/stack"
	"github.com/donyori/gogo/fmtcoll"
)

func TestNew(t *testing.T) {
	t.Parallel()

	for capacity := -1; capacity <= 33; capacity++ {
		wantInitCap := capacity
		if wantInitCap <= 0 {
			wantInitCap = stack.DefaultCapacity
		}

		t.Run(fmt.Sprintf("cap=%d", capacity), func(t *testing.T) {
			t.Parallel()

			s := stack.New[int](capacity)
			if s == nil {
				t.Error("got nil stack")
			} else if c := s.Cap(); c != wantInitCap {
				t.Errorf("got initial capacity %d; want %d", c, wantInitCap)
			}
		})
	}
}

func TestStack_Range(t *testing.T) {
	t.Parallel()

	data := []int{0, 1, 2, 3, 3, 4, 0, 1, 2, 3, 3, 4}
	want := []int{4, 3, 3, 2, 1, 0}

	s := stack.New[int](0)
	for _, x := range data {
		s.Push(x)
	}

	got := make([]int, 0, len(data))

	s.Range(func(x int) (cont bool) {
		got = append(got, x)
		return len(got) < len(data)>>1
	})

	if !slices.Equal(got, want) {
		t.Errorf("got %v; want %v", got, want)
	}
}

func TestStack_Range_Empty(t *testing.T) {
	t.Parallel()

	s := stack.New[int](0)
	s.Range(func(x int) (cont bool) {
		t.Error("handler was called, x:", x)
		return true
	})
}

func TestStack_Range_NilHandler(t *testing.T) {
	t.Parallel()

	data := []int{0, 1, 2, 3, 3, 4, 0, 1, 2, 3, 3, 4}

	s := stack.New[int](0)
	for _, x := range data {
		s.Push(x)
	}

	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()

	s.Range(nil)
}

func TestStack_IterItems(t *testing.T) {
	t.Parallel()

	data := []int{0, 1, 2, 3, 3, 4, 5, 6, 7, 8, 8, 9}
	want := []int{9, 8, 8, 7, 6, 5}

	s := stack.New[int](0)
	for _, x := range data {
		s.Push(x)
	}

	seq := s.IterItems()
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

func TestStack_IterItems_Empty(t *testing.T) {
	t.Parallel()

	s := stack.New[int](0)

	seq := s.IterItems()
	if seq == nil {
		t.Fatal("got nil iterator")
	}

	for x := range seq {
		t.Error("yielded", x)
	}
}

func TestStack_Reserve(t *testing.T) {
	t.Parallel()

	dataList := [][]int{nil, {}, {0}, {0, 1}, {0, 1, 2}}
	capList := []int{-1, 0, 1, 2, 3, 4}

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

			testCases[idx].wantRangeData = make([]int, len(data))
			for i := range data {
				testCases[idx].wantRangeData[i] = data[len(data)-1-i]
			}

			initCap := len(data)
			if initCap == 0 {
				initCap = stack.DefaultCapacity
			}

			testCases[idx].wantCap = capacity
			if testCases[idx].wantCap <= 0 {
				testCases[idx].wantCap = stack.DefaultCapacity
			}

			if testCases[idx].wantCap < initCap {
				testCases[idx].wantCap = initCap
			}

			idx++
		}
	}

	for _, tc := range testCases {
		dataName := fmtcoll.MustFormatSliceToString(
			tc.data,
			fmtcoll.NewDefaultSequenceFormat[int](),
		)
		t.Run(
			fmt.Sprintf("cap=%d&data=%s", tc.capacity, dataName),
			func(t *testing.T) {
				t.Parallel()

				testStackReserve(
					t,
					tc.data,
					tc.capacity,
					tc.wantRangeData,
					tc.wantCap,
				)
			},
		)
	}
}

// testStackReserve is the main process of TestStack_Reserve.
func testStackReserve(
	t *testing.T,
	data []int,
	capacity int,
	wantRangeData []int,
	wantCap int,
) {
	t.Helper()

	s := stack.New[int](len(data))
	for _, x := range data {
		s.Push(x)
	}

	s.Reserve(capacity)

	if c := s.Cap(); c != wantCap {
		t.Errorf("got capacity %d; want %d", c, wantCap)
	}

	rangeData := make([]int, 0, s.Len())
	s.Range(func(x int) (cont bool) {
		rangeData = append(rangeData, x)
		return true
	})

	if !slices.Equal(rangeData, wantRangeData) {
		t.Errorf("got data by s.Range %v; want %v", rangeData, wantRangeData)
	}
}

func TestStack_PushNAndPopN(t *testing.T) {
	t.Parallel()

	ns := make([]int, 33, 36)
	for i := range ns {
		ns[i] = i + 1
	}

	ns = append(ns, 63, 4096, 524288)

	for _, n := range ns {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			t.Parallel()

			s := stack.New[int](0)
			if sn := s.Len(); sn != 0 {
				t.Fatalf("initial - s.Len() %d; want 0", sn)
			}

			testStackPushNAndPopNPushStage(t, n, s)

			if t.Failed() {
				return
			}

			testStackPushNAndPopNPopStage(t, n, s)

			if t.Failed() {
				return
			}

			testStackPushAndPopTearDownStage(t, s)
		})
	}
}

// testStackPushNAndPopNPushStage is a subprocess of TestStack_PushNAndPopN
// for the pushing stage.
func testStackPushNAndPopNPushStage(t *testing.T, n int, s stack.Stack[int]) {
	t.Helper()

	for x := 1; !t.Failed() && x <= n; x++ {
		s.Push(x)

		if sn := s.Len(); sn != x {
			t.Errorf("after s.Push(%d) - got s.Len() %d; want %[1]d", x, sn)
		}

		if top := s.Peek(); top != x {
			t.Errorf("after s.Push(%d) - got s.Peek() %d; want %[1]d", x, top)
		}
	}
}

// testStackPushNAndPopNPopStage is a subprocess of TestStack_PushNAndPopN
// for the popping stage.
func testStackPushNAndPopNPopStage(t *testing.T, n int, s stack.Stack[int]) {
	t.Helper()

	finalCap := s.Cap()

	for x := n - 1; !t.Failed() && x >= 0; x-- {
		got := s.Pop()
		if got != x+1 {
			t.Errorf("got No.%d s.Pop() %d; want %d", n-x, got, x+1)
		}

		if sn := s.Len(); sn != x {
			t.Errorf("after No.%d s.Pop() - got s.Len() %d; want %d",
				n-x, sn, x)
		}

		if c := s.Cap(); c != finalCap {
			t.Errorf("after No.%d s.Pop() - got s.Cap() %d; want %d",
				n-x, c, finalCap)
		}

		if x > 0 {
			if top := s.Peek(); top != x {
				t.Errorf("after No.%d s.Pop() - got s.Peek() %d; want %d",
					n-x, top, x)
			}
		}
	}
}

func TestStack_RandomPushAndPop(t *testing.T) {
	t.Parallel()

	s := stack.New[int](0)
	if sn := s.Len(); sn != 0 {
		t.Fatalf("initial - s.Len() %d; want 0", sn)
	}

	random := rand.New(rand.NewChaCha8( //gosec:disable G404 -- math/rand/v2 is reproducible
		[32]byte([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ123456")),
	))

	var (
		stackData []int
		pushCtr   int
		popCtr    int
	)

	// Push and pop a total of N items.
	// Each time randomly push a portion of them
	// and then randomly pop a portion of items in the stack.

	const N int = 1 << 20

	n := N // the number of remaining items to be pushed

	// When n >= 3, push randomly 1 to (2/3)n items
	// and then randomly pop at least 1 item.
	for n >= 3 {
		pushN := 1 + random.IntN(n/3<<1)
		n -= pushN
		testStackRandomPushAndPopPushStage(t, pushN, &pushCtr, &stackData, s)

		if t.Failed() {
			return
		}

		testStackRandomPushAndPopPopStage(
			t,
			1+random.IntN(len(stackData)),
			&popCtr,
			&stackData,
			s,
		)

		if t.Failed() {
			return
		}
	}

	// When n < 3, push all remaining items and then pop all items.
	testStackRandomPushAndPopPushStage(t, n, &pushCtr, &stackData, s)

	if t.Failed() {
		return
	}

	testStackRandomPushAndPopPopStage(t, len(stackData), &popCtr, &stackData, s)

	if t.Failed() {
		return
	}

	// An unnecessary test on pushCtr and popCtr
	// to verify whether all the N items have been pushed and popped:
	if pushCtr != N || popCtr != N {
		t.Fatalf("got pushCtr %d, popCtr %d; both want %d", pushCtr, popCtr, N)
	}

	testStackPushAndPopTearDownStage(t, s)
}

// testStackRandomPushAndPopPushStage is a subprocess of
// TestStack_RandomPushAndPop for the pushing stage.
func testStackRandomPushAndPopPushStage(
	t *testing.T,
	n int,
	pPushCtr *int,
	pStackData *[]int,
	s stack.Stack[int],
) {
	t.Helper()

	for i := 0; !t.Failed() && i < n; i++ {
		*pPushCtr++
		*pStackData = append(*pStackData, *pPushCtr)

		s.Push(*pPushCtr)

		if sn := s.Len(); sn != len(*pStackData) {
			t.Errorf("after s.Push(%d) - got s.Len() %d; want %d",
				*pPushCtr, sn, len(*pStackData))
		}

		if top := s.Peek(); top != *pPushCtr {
			t.Errorf("after s.Push(%d) - got s.Peek() %d; want %[1]d",
				*pPushCtr, top)
		}
	}
}

// testStackRandomPushAndPopPopStage is a subprocess of
// TestStack_RandomPushAndPop for the popping stage.
func testStackRandomPushAndPopPopStage(
	t *testing.T,
	n int,
	pPopCtr *int,
	pStackData *[]int,
	s stack.Stack[int],
) {
	t.Helper()

	wantCap := s.Cap()

	for i := 0; !t.Failed() && i < n; i++ {
		*pPopCtr++
		want := (*pStackData)[len(*pStackData)-1]
		*pStackData = (*pStackData)[:len(*pStackData)-1]

		got := s.Pop()
		if got != want {
			t.Errorf("got No.%d s.Pop() %d; want %d", *pPopCtr, got, want)
		}

		if sn := s.Len(); sn != len(*pStackData) {
			t.Errorf("after No.%d s.Pop() - got s.Len() %d; want %d",
				*pPopCtr, sn, len(*pStackData))
		}

		if c := s.Cap(); c != wantCap {
			t.Errorf("after No.%d s.Pop() - got s.Cap() %d; want %d",
				*pPopCtr, c, wantCap)
		}

		if len(*pStackData) > 0 {
			if top := s.Peek(); top != (*pStackData)[len(*pStackData)-1] {
				t.Errorf("after No.%d s.Pop() - got s.Peek() %d; want %d",
					*pPopCtr, top, (*pStackData)[len(*pStackData)-1])
			}
		}
	}
}

// testStackPushAndPopTearDownStage is the common subprocess of
// TestStack_PushNAndPopN and TestStack_RandomPushAndPop
// for the tearing down stage.
func testStackPushAndPopTearDownStage(t *testing.T, s stack.Stack[int]) {
	t.Helper()

	finalCap := s.Cap()

	s.RemoveAll()

	if sn := s.Len(); sn != 0 {
		t.Errorf("after s.RemoveAll() - got s.Len() %d; want 0", sn)
	}

	if c := s.Cap(); c != finalCap {
		t.Errorf("after s.RemoveAll() - got s.Cap() %d; want %d", c, finalCap)
	}

	s.Clear()

	if sn := s.Len(); sn != 0 {
		t.Errorf("after s.Clear() - got s.Len() %d; want 0", sn)
	}

	if c := s.Cap(); c != 0 {
		t.Errorf("after s.Clear() - got s.Cap() %d; want 0", c)
	}

	s.RemoveAll()

	if sn := s.Len(); sn != 0 {
		t.Errorf("after s.Clear() then s.RemoveAll() - got s.Len() %d; want 0",
			sn)
	}

	if c := s.Cap(); c != 0 {
		t.Errorf("after s.Clear() then s.RemoveAll() - got s.Cap() %d; want 0",
			c)
	}
}

func TestStack_PushAfterClear(t *testing.T) {
	// This tests whether the stack is reusable after Clear().
	t.Parallel()

	s := stack.New[int](0)

	const N int = 10
	for range N {
		s.Push(1)
	}

	if n := s.Len(); n != N {
		t.Errorf("before s.Clear() - got s.Len() %d; want %d", n, N)
	}

	s.Clear()

	if n := s.Len(); n != 0 {
		t.Errorf("after s.Clear() - got s.Len() %d; want 0", n)
	}

	if c := s.Cap(); c != 0 {
		t.Errorf("after s.Clear() - got s.Cap() %d; want 0", c)
	}

	s.Push(2)

	if n := s.Len(); n != 1 {
		t.Errorf("after s.Push(2) - got s.Len() %d; want 1", n)
	}

	if top := s.Peek(); top != 2 {
		t.Errorf("after s.Push(2) - got s.Peek() %d; want 2", top)
	}
}
