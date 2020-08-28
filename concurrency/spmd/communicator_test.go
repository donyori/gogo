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
