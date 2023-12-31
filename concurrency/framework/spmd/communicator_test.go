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

package spmd_test

import (
	"testing"
	"time"

	"github.com/donyori/gogo/concurrency/framework/spmd"
	"github.com/donyori/gogo/container/sequence/array"
)

func TestCommunicator_Send_Receive(t *testing.T) {
	dataFn := func(src, dst int) int {
		return src*10 + dst
	}

	prs := spmd.Run(4, func(
		world spmd.Communicator[int],
		commMap map[string]spmd.Communicator[int],
	) {
		testCommunicatorSendReceiveProc(
			world.Rank(),
			func(rank int, dst int) {
				if !world.Send(dst, dataFn(rank, dst)) {
					t.Errorf("goroutine %d, dst %d, Send returns false",
						rank, dst)
				}
			},
			func(rank int, src int) {
				msg, ok := world.Receive(src)
				if !ok {
					t.Errorf("goroutine %d, src %d, Receive returns ok to false",
						rank, src)
				} else if want := dataFn(src, rank); msg != want {
					t.Errorf("goroutine %d, src %d, Receive got %d; want %d",
						rank, src, msg, want)
				}
			},
		)
	}, nil)
	if len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

// testCommunicatorSendReceiveProc is the main process of the business function
// for TestCommunicator_Send_Receive.
func testCommunicatorSendReceiveProc(
	rank int,
	sendTest func(rank, dst int),
	recvTest func(rank, src int),
) {
	// Point-to-point communication test.
	//
	// Round 1: 0 -> 2, 1 -> 3
	// Round 2: 0 -> 3, 1 -> 2
	// Round 3: 0 -> 1, 2 -> 3
	// Round 4: 1 -> 0, 3 -> 2
	// Round 5: 2 -> 0, 3 -> 1
	// Round 6: 3 -> 0, 2 -> 1
	switch rank {
	case 0:
		for _, dst := range []int{2, 3, 1} {
			sendTest(rank, dst)
		}
		for _, src := range []int{1, 2, 3} {
			recvTest(rank, src)
		}
	case 1:
		for _, dst := range []int{3, 2} {
			sendTest(rank, dst)
		}
		recvTest(rank, 0)
		sendTest(rank, 0)
		for _, src := range []int{3, 2} {
			recvTest(rank, src)
		}
	case 2:
		for _, src := range []int{0, 1} {
			recvTest(rank, src)
		}
		sendTest(rank, 3)
		recvTest(rank, 3)
		for _, dst := range []int{0, 1} {
			sendTest(rank, dst)
		}
	case 3:
		for _, src := range []int{1, 0, 2} {
			recvTest(rank, src)
		}
		for _, dst := range []int{2, 1, 0} {
			sendTest(rank, dst)
		}
	}
}

func TestCommunicator_Send_Receive_Any(t *testing.T) {
	dataFn := func(src, dst int) int {
		return src*10 + dst
	}
	prs := spmd.Run(7, func(
		world spmd.Communicator[int],
		commMap map[string]spmd.Communicator[int],
	) {
		if world.Rank() == 6 {
			timer := time.AfterFunc(time.Second, func() {
				t.Error("timeout")
				world.Canceler().Cancel()
			})
			defer timer.Stop()
			// Goroutine 6 is a progress monitor.
			for i := 1; i <= 4; i++ {
				// t.Logf("part %d starts at %v", i, time.Now())
				if i < 4 && !world.Barrier() {
					return
				}
			}
			return
		}
		comm := commMap["tester"]
		switch comm.Rank() {
		case 0, 1, 4, 5:
			testCommunicatorSendReceiveAny0145Proc(
				t,
				world,
				comm,
				dataFn,
				func(rank, dst, gotDst int) {
					if gotDst < 0 {
						t.Errorf("goroutine %d, dst %d, detected an unexpected cancellation signal",
							rank, dst)
					} else if gotDst != dst {
						t.Errorf("goroutine %d, got dst %d; want %d",
							rank, gotDst, dst)
					}
				},
			)
		case 2, 3:
			testCommunicatorSendReceiveAny23Proc(
				t,
				world,
				comm,
				dataFn,
				func(rank int, src int, msg any) {
					if src < 0 {
						t.Errorf("goroutine %d, detected an unexpected cancellation signal",
							rank)
					} else if want := dataFn(src, rank); msg != want {
						t.Errorf("goroutine %d, src %d, Receive got %d; want %d",
							rank, src, msg, want)
					}
				},
			)
		}
	}, map[string][]int{"tester": {0, 1, 2, 3, 4, 5}})
	if len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

// testCommunicatorSendReceiveAny0145Proc is the main process of
// the business function for goroutines 0, 1, 4, and 5
// for TestCommunicator_Send_Receive_Any.
func testCommunicatorSendReceiveAny0145Proc(
	t *testing.T,
	world spmd.Communicator[int],
	comm spmd.Communicator[int],
	dataFn func(src, dst int) int,
	checkDst func(rank, dst, gotDst int),
) {
	r := comm.Rank()
	switch r {
	case 0, 4:
		for _, dst := range []int{2, 3, 3} {
			checkDst(r, dst, comm.SendAny(dataFn(r, dst)))
			world.Barrier()
		}
		if comm.Send(3, dataFn(r, 3)) {
			t.Errorf("goroutine %d, dst 3, Send got true; want false", r)
		}
	case 1, 5:
		if !comm.Send(2, dataFn(r, 2)) {
			t.Errorf("goroutine %d, dst 2, Send got false; want true", r)
		}
		for _, dst := range []int{2, 3} {
			checkDst(r, dst, comm.SendPublic(dataFn(r, dst)))
			world.Barrier()
			if dst == 2 {
				world.Barrier()
			}
		}
		if comm.Send(3, dataFn(r, 3)) {
			t.Errorf("goroutine %d, dst 3, Send got true; want false", r)
		}
	}
}

// testCommunicatorSendReceiveAny23Proc is the main process of
// the business function for goroutines 2 and 3
// for TestCommunicator_Send_Receive_Any.
func testCommunicatorSendReceiveAny23Proc(
	t *testing.T,
	world spmd.Communicator[int],
	comm spmd.Communicator[int],
	dataFn func(src, dst int) int,
	checkSrcMsg func(rank int, src int, msg any),
) {
	r := comm.Rank()
	switch r {
	case 2:
		for i := 0; i < 6; i++ {
			src, msg := comm.ReceiveAny()
			checkSrcMsg(r, src, msg)
			// t.Logf("goroutine %d, ReceiveAny - src %d, msg %d", r, src, msg)
		}
		world.Barrier()
		world.Barrier()
		world.Barrier()
		if comm.Send(3, dataFn(r, 3)) {
			t.Errorf("goroutine %d, dst 3, Send got true; want false", r)
		}
	case 3:
		world.Barrier()
		for _, src := range []int{0, 4} {
			msg, ok := comm.Receive(src)
			if !ok {
				t.Errorf("goroutine %d, detected an unexpected cancellation signal",
					r)
			} else if want := dataFn(src, r); msg != want {
				t.Errorf("goroutine %d, src %d, Receive got %d; want %d",
					r, src, msg, want)
			}
		}
		world.Barrier()
		for i := 0; i < 4; i++ {
			src, msg := comm.ReceivePublic()
			checkSrcMsg(r, src, msg)
		}
		world.Barrier()
		time.AfterFunc(time.Microsecond, func() {
			comm.Canceler().Cancel() // cancel the job to let other goroutines exit
		})
		src, msg := comm.ReceivePublic()
		if src >= 0 {
			t.Errorf("goroutine %d, ReceivePublic got (%d, %d); want (-1, 0)",
				r, src, msg)
		}
	}
}

func TestCommunicator_Barrier(t *testing.T) {
	times := make([]time.Time, 4)
	prs := spmd.Run(4, func(
		world spmd.Communicator[int],
		commMap map[string]spmd.Communicator[int],
	) {
		r := world.Rank()
		time.Sleep(time.Millisecond * time.Duration(r))
		world.Barrier()
		times[r] = time.Now()
	}, nil)
	if len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
	for i := 1; i < len(times); i++ {
		diff := times[0].Sub(times[i])
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Microsecond {
			t.Errorf("goroutine 0 and %d are %v apart", i, diff)
		}
	}
}

func TestCommunicator_Broadcast(t *testing.T) {
	data := [][4]any{
		{1},
		{2, 0.3, 4, 5},
		{nil, nil, "Hello"},
		{nil, nil, nil, complex(1, -1)},
		{},
	}
	ctrl := spmd.New(4, func(
		world spmd.Communicator[any],
		commMap map[string]spmd.Communicator[any],
	) {
		r := world.Rank()
		for i, a := range data {
			msg, ok := world.Broadcast(i%4, a[r])
			if !ok {
				t.Errorf("goroutine %d, root %d, detected an unexpected cancellation signal",
					r, i%4)
			}
			if msg != a[i%4] {
				t.Errorf("goroutine %d, root %d, got %v; want %v",
					r, i%4, msg, a[i%4])
			}
		}
	}, nil)
	ctrl.Run()
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
	if n := spmd.WrapController[any](ctrl).GetWorldBcastMapLen(); n > 0 {
		t.Errorf("broadcast channel map is not clean: %d element(s) remained",
			n)
	}
}

func TestCommunicator_Scatter(t *testing.T) {
	a := array.SliceDynamicArray[int]{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	want := [4]array.SliceDynamicArray[int]{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8},
		{9, 10},
	}
	data := [][4]*array.SliceDynamicArray[int]{
		{&a},
		{nil, &a},
		{nil, nil, &a},
		{nil, nil, nil, &a},
		{},
	}
	ctrl := spmd.New(4, func(
		world spmd.Communicator[int],
		commMap map[string]spmd.Communicator[int],
	) {
		r := world.Rank()
		for i, a := range data {
			intArray, ok := world.Scatter(i%4, a[r])
			if !ok {
				t.Errorf("goroutine %d, root %d, detected an unexpected cancellation signal",
					r, i%4)
			}
			if intArray == nil {
				if i < 4 {
					t.Errorf("goroutine %d, root %d, intArray is nil; want %v",
						r, i%4, want[r])
				}
				continue
			}
			if n := intArray.Len(); n != want[r].Len() {
				t.Errorf("goroutine %d, root %d, got intArray.Len %d; want %d",
					r, i%4, n, want[r].Len())
				continue
			}
			var k int
			intArray.Range(func(x int) (cont bool) {
				if x != want[r][k] {
					t.Errorf("goroutine %d, root %d, got intArray[%d] %d; want %d",
						r, i%4, k, x, want[r][k])
					return false
				}
				k++
				return true
			})
		}
	}, nil)
	ctrl.Run()
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
	if n := spmd.WrapController[int](ctrl).GetWorldScatterMapLen(); n > 0 {
		t.Errorf("scatter channel map is not clean: %d element(s) remained", n)
	}
}

func TestCommunicator_Gather(t *testing.T) {
	data := [][4]any{
		{4, 3, 2, 1},
		{1, nil, nil, 4},
		{nil, 2, 3},
		{},
	}
	ctrl := spmd.New(4, func(
		world spmd.Communicator[any],
		commMap map[string]spmd.Communicator[any],
	) {
		r := world.Rank()
		for _, a := range data {
			for root := 0; root < 4; root++ {
				x, ok := world.Gather(root, a[r])
				if !ok {
					t.Errorf("goroutine %d, root %d, detected an unexpected cancellation signal",
						r, root)
				}
				if r == root {
					if len(x) != len(a) {
						t.Errorf("goroutine %d, root %d, got x %v; want %v",
							r, root, x, a)
						continue
					}
					for k := range x {
						if x[k] != a[k] {
							t.Errorf("goroutine %d, root %d, got x %v; want %v",
								r, root, x, a)
							break
						}
					}
				} else if x != nil {
					t.Errorf("goroutine %d, root %d, got x %v; want nil",
						r, root, x)
				}
			}
		}
	}, nil)
	ctrl.Run()
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
	if n := spmd.WrapController[any](ctrl).GetWorldGatherMapLen(); n > 0 {
		t.Errorf("gather channel map is not clean: %d element(s) remained", n)
	}
}
