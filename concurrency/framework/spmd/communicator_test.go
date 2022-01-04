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

package spmd

import (
	"testing"
	"time"

	"github.com/donyori/gogo/container/sequence"
)

func TestCommunicator_Send_Receive(t *testing.T) {
	dataFn := func(src, dst int) int {
		return src*10 + dst
	}

	// Point-to-point communication test.
	//
	// Round 1: 0 -> 2, 1 -> 3
	// Round 2: 0 -> 3, 1 -> 2
	// Round 3: 0 -> 1, 2 -> 3
	// Round 4: 1 -> 0, 3 -> 2
	// Round 5: 2 -> 0, 3 -> 1
	// Round 6: 3 -> 0, 2 -> 1
	prs := Run(4, func(world Communicator, commMap map[string]Communicator) {
		r := world.Rank()
		var src, dst int
		sendTest := func() {
			if !world.Send(dst, dataFn(r, dst)) {
				t.Errorf("Goroutine %d, dst %d: Send returns false.", r, dst)
			}
		}
		recvTest := func() {
			msg, ok := world.Receive(src)
			if !ok {
				t.Errorf("Goroutine %d, src %d: Receive returns ok to false.", r, src)
			} else if wanted := dataFn(src, r); msg != wanted {
				t.Errorf("Goroutine %d, src %d: Receive msg: %v != %d.", r, src, msg, wanted)
			}
		}
		switch r {
		case 0:
			for _, dst = range []int{2, 3, 1} {
				sendTest()
			}
			for _, src = range []int{1, 2, 3} {
				recvTest()
			}
		case 1:
			for _, dst = range []int{3, 2} {
				sendTest()
			}
			src = 0
			recvTest()
			dst = 0
			sendTest()
			for _, src = range []int{3, 2} {
				recvTest()
			}
		case 2:
			for _, src = range []int{0, 1} {
				recvTest()
			}
			dst = 3
			sendTest()
			src = 3
			recvTest()
			for _, dst = range []int{0, 1} {
				sendTest()
			}
		case 3:
			for _, src = range []int{1, 0, 2} {
				recvTest()
			}
			for _, dst = range []int{2, 1, 0} {
				sendTest()
			}
		}
	}, nil)
	if len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
}

func TestCommunicator_Send_Receive_Any(t *testing.T) {
	dataFn := func(src, dst int) int {
		return src*10 + dst
	}
	prs := Run(7, func(world Communicator, commMap map[string]Communicator) {
		if world.Rank() == 6 {
			timer := time.AfterFunc(time.Second, func() {
				t.Error("Timeout!")
				world.Quit()
			})
			defer timer.Stop()
			// Goroutine 6 is a progress monitor.
			for i := 1; i <= 4; i++ {
				t.Logf("Part %d starts at %v.", i, time.Now())
				if i < 4 && !world.Barrier() {
					return
				}
			}
			return
		}
		comm := commMap["tester"]
		r := comm.Rank()
		var src, dst int
		var msg interface{}
		checkD := func(d int) {
			if d < 0 {
				t.Errorf("Goroutine %d, dst %d: An unexpected quit signal was detected.", r, dst)
			} else if d != dst {
				t.Errorf("Goroutine %d: dst: %d != %d.", r, d, dst)
			}
		}
		checkSrcMsg := func() {
			if src < 0 {
				t.Errorf("Goroutine %d: An unexpected quit signal was detected.", r)
			} else if wanted := dataFn(src, r); msg != wanted {
				t.Errorf("Goroutine %d, src %d: Receive msg: %v != %d.", r, src, msg, wanted)
			}
		}
		switch r {
		case 0, 4:
			for _, dst = range []int{2, 3, 3} {
				checkD(comm.SendAny(dataFn(r, dst)))
				world.Barrier()
			}
			if comm.Send(3, dataFn(r, 3)) {
				t.Errorf("Goroutine %d, dst 3: Send should return false but got true.", r)
			}
		case 1, 5:
			if !comm.Send(2, dataFn(r, 2)) {
				t.Errorf("Goroutine %d, dst 2: Send returns false.", r)
			}
			for _, dst = range []int{2, 3} {
				checkD(comm.SendPublic(dataFn(r, dst)))
				world.Barrier()
				if dst == 2 {
					world.Barrier()
				}
			}
			if comm.Send(3, dataFn(r, 3)) {
				t.Errorf("Goroutine %d, dst 3: Send should return false but got true.", r)
			}
		case 2:
			for i := 0; i < 6; i++ {
				src, msg = comm.ReceiveAny()
				checkSrcMsg()
				// t.Logf("Goroutine %d: ReceiveAny() - src: %d, msg: %v.", r, src, msg)
			}
			world.Barrier()
			world.Barrier()
			world.Barrier()
			if comm.Send(3, dataFn(r, 3)) {
				t.Errorf("Goroutine %d, dst 3: Send should return false but got true.", r)
			}
		case 3:
			world.Barrier()
			for _, src = range []int{0, 4} {
				msg, ok := comm.Receive(src)
				if !ok {
					t.Errorf("Goroutine %d: An unexpected quit signal was detected.", r)
				} else if wanted := dataFn(src, r); msg != wanted {
					t.Errorf("Goroutine %d, src %d: Receive msg: %v != %d.", r, src, msg, wanted)
				}
			}
			world.Barrier()
			for i := 0; i < 4; i++ {
				src, msg = comm.ReceivePublic()
				checkSrcMsg()
			}
			world.Barrier()
			time.AfterFunc(time.Microsecond, func() {
				comm.Quit() // Quit the job to let other goroutines exit.
			})
			src, msg = comm.ReceivePublic()
			if src >= 0 {
				t.Errorf("Goroutine %d: ReceivePublic should get nothing but src: %d, msg: %v.", r, src, msg)
			}
		}
	}, map[string][]int{"tester": {0, 1, 2, 3, 4, 5}})
	if len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
}

func TestCommunicator_Barrier(t *testing.T) {
	times := make([]time.Time, 4)
	prs := Run(4, func(world Communicator, commMap map[string]Communicator) {
		r := world.Rank()
		time.Sleep(time.Millisecond * time.Duration(r))
		world.Barrier()
		times[r] = time.Now()
	}, nil)
	if len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
	for i := 1; i < len(times); i++ {
		diff := times[0].Sub(times[i])
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Microsecond {
			t.Errorf("Goroutine 0 and %d are %v apart.", i, diff)
		}
	}
}

func TestCommunicator_Broadcast(t *testing.T) {
	data := [][4]interface{}{
		{1},
		{2, 0.3, 4, 5},
		{nil, nil, "Hello"},
		{nil, nil, nil, complex(1, -1)},
		{},
	}
	ctrl := New(4, func(world Communicator, commMap map[string]Communicator) {
		r := world.Rank()
		for i, a := range data {
			msg, ok := world.Broadcast(i%4, a[r])
			if !ok {
				t.Errorf("Goroutine %d, root %d: An unexpected quit signal was detected.", r, i%4)
			}
			if msg != a[i%4] {
				t.Errorf("Goroutine %d, root %d: msg: %v != root msg: %v.", r, i%4, msg, a[i%4])
			}
		}
	}, nil).(*controller)
	ctrl.Run()
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
	if n := len(ctrl.World.BcastMap); n > 0 {
		t.Errorf("Broadcast channel map is NOT clean. %d element(s) remained.", n)
	}
}

func TestCommunicator_Scatter(t *testing.T) {
	array := sequence.IntDynamicArray{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	wanted := [4]sequence.IntDynamicArray{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8},
		{9, 10},
	}
	data := [][4]sequence.IntDynamicArray{
		{array},
		{nil, array},
		{nil, nil, array},
		{nil, nil, nil, array},
		{},
	}
	ctrl := New(4, func(world Communicator, commMap map[string]Communicator) {
		r := world.Rank()
		for i, a := range data {
			msg, ok := world.Scatter(i%4, a[r])
			if !ok {
				t.Errorf("Goroutine %d, root %d: An unexpected quit signal was detected.", r, i%4)
			}
			if msg == nil {
				if i < 4 {
					t.Errorf("Goroutine %d, root %d: msg is nil but should be %v.", r, i%4, wanted[r])
				}
				continue
			}
			ints := msg.(sequence.IntDynamicArray)
			if ints.Len() != wanted[r].Len() {
				t.Errorf("Goroutine %d, root %d: msg: %v != %v.", r, i%4, ints, wanted[r])
				continue
			}
			for k := range ints {
				if ints[k] != wanted[r][k] {
					t.Errorf("Goroutine %d, root %d: msg: %v != %v.", r, i%4, ints, wanted[r])
					break
				}
			}
		}
	}, nil).(*controller)
	ctrl.Run()
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
	if n := len(ctrl.World.ScatterMap); n > 0 {
		t.Errorf("Scatter channel map is NOT clean. %d element(s) remained.", n)
	}
}

func TestCommunicator_Gather(t *testing.T) {
	data := [][4]interface{}{
		{4, 3, 2, 1},
		{1, nil, nil, 4},
		{nil, 2, 3},
		{},
	}
	ctrl := New(4, func(world Communicator, commMap map[string]Communicator) {
		r := world.Rank()
		for _, a := range data {
			for root := 0; root < 4; root++ {
				x, ok := world.Gather(root, a[r])
				if !ok {
					t.Errorf("Goroutine %d, root %d: An unexpected quit signal was detected.", r, root)
				}
				if r == root {
					if len(x) != len(a) {
						t.Errorf("Goroutine %d, root %d: x: %v != %v.", r, root, x, a)
						continue
					}
					for k := range x {
						if x[k] != a[k] {
							t.Errorf("Goroutine %d, root %d: x: %v != %v.", r, root, x, a)
							break
						}
					}
				} else if x != nil {
					t.Errorf("Goroutine %d, root %d: x != nil.", r, root)
				}
			}
		}
	}, nil).(*controller)
	ctrl.Run()
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
	if n := len(ctrl.World.GatherMap); n > 0 {
		t.Errorf("Gather channel map is NOT clean. %d element(s) remained.", n)
	}
}
