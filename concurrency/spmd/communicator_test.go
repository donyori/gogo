// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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
	dataFn := func(src, dest int) int {
		return src*10 + dest
	}

	// Point-to-point communication test.
	//
	// 1st round: 0 -> 2, 1 -> 3
	// 2nd round: 0 -> 3, 1 -> 2
	// 3rd round: 0 -> 1, 2 -> 3
	// 4th round: 1 -> 0, 3 -> 2
	// 5th round: 2 -> 0, 3 -> 1
	// 6th round: 3 -> 0, 2 -> 1
	prs := Run(4, func(world Communicator, commMap map[string]Communicator) {
		r := world.Rank()
		var src, dest int
		sendTest := func() {
			if !world.Send(dest, dataFn(r, dest)) {
				t.Errorf("Goroutine %d, dest %d: Send returns false.", r, dest)
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
			for _, dest = range []int{2, 3, 1} {
				sendTest()
			}
			for _, src = range []int{1, 2, 3} {
				recvTest()
			}
		case 1:
			for _, dest = range []int{3, 2} {
				sendTest()
			}
			src = 0
			recvTest()
			dest = 0
			sendTest()
			for _, src = range []int{3, 2} {
				recvTest()
			}
		case 2:
			for _, src = range []int{0, 1} {
				recvTest()
			}
			dest = 3
			sendTest()
			src = 3
			recvTest()
			for _, dest = range []int{0, 1} {
				sendTest()
			}
		case 3:
			for _, src = range []int{1, 0, 2} {
				recvTest()
			}
			for _, dest = range []int{2, 1, 0} {
				sendTest()
			}
		}
	}, nil)
	if len(prs) > 0 {
		t.Errorf("Panic: %v.", prs)
	}
}

func TestCommunicator_Send_Receive_Any(t *testing.T) {
	dataFn := func(src, dest int) int {
		return src*10 + dest
	}
	prs := Run(6, func(world Communicator, commMap map[string]Communicator) {
		r := world.Rank()
		switch r {
		case 0, 3:
			for _, d := range []int{2, 5} {
				dest := world.SendToAny(dataFn(r, d))
				if dest < 0 {
					t.Errorf("Goroutine %d, dest %d: An unexpected quit signal was detected.", r, d)
					return
				}
				if dest != d {
					t.Errorf("Goroutine %d: dest: %d != %d.", r, dest, d)
				}
				if d == 2 {
					world.Barrier()
				}
			}
		case 1, 4:
			if !world.Send(2, dataFn(r, 2)) {
				t.Errorf("Goroutine %d, dest 2: Send returns false.", r)
			}
			world.Barrier()
			if world.Send(5, dataFn(r, 5)) {
				t.Errorf("Goroutine %d, dest 5: Send should return false but got true.", r)
			}
		case 2:
			for i := 0; i < 4; i++ {
				src, msg := world.ReceiveFromAny()
				if src < 0 {
					t.Errorf("Goroutine %d: An unexpected quit signal was detected.", r)
					return
				}
				if wanted := dataFn(src, r); msg != wanted {
					t.Errorf("Goroutine %d, src %d: Receive msg: %v != %d.", r, src, msg, wanted)
				}
			}
			world.Barrier()
		case 5:
			world.Barrier()
			for i := 0; i < 2; i++ {
				src, msg := world.ReceiveOnlyAny()
				if src < 0 {
					t.Errorf("Goroutine %d: An unexpected quit signal was detected.", r)
					return
				}
				if wanted := dataFn(src, r); msg != wanted {
					t.Errorf("Goroutine %d, src %d: Receive msg: %v != %d.", r, src, msg, wanted)
				}
			}
			time.AfterFunc(time.Microsecond, func() {
				world.Quit() // Quit the job to let Goroutine 1, Goroutine 4, and Goroutine 5 exit.
			})
			src, msg := world.ReceiveOnlyAny()
			if src >= 0 {
				t.Errorf("Goroutine %d: ReceiveOnlyAny should get nothing but src: %d, msg: %v.", r, src, msg)
			}
		}
	}, nil)
	if len(prs) > 0 {
		t.Errorf("Panic: %v.", prs)
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
		t.Errorf("Panic: %v.", prs)
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
		t.Errorf("Panic: %v.", prs)
	}
	for i, m := range ctrl.World.ChanMaps {
		if n := len(m); n > 0 {
			t.Errorf("Channel map %d is NOT clean. %d element(s) remained.", i, n)
		}
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
		t.Errorf("Panic: %v.", prs)
	}
	for i, m := range ctrl.World.ChanMaps {
		if n := len(m); n > 0 {
			t.Errorf("Channel map %d is NOT clean. %d element(s) remained.", i, n)
		}
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
		t.Errorf("Panic: %v.", prs)
	}
	for i, m := range ctrl.World.ChanMaps {
		if n := len(m); n > 0 {
			t.Errorf("Channel map %d is NOT clean. %d element(s) remained.", i, n)
		}
	}
}
