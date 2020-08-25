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
	finishTimes := make([][4]time.Time, len(data))
	ctrl := New(4, func(world Communicator, commMap map[string]Communicator) {
		for i, a := range data {
			world.Barrier()
			r := world.Rank()
			time.Sleep(time.Millisecond * 10 * time.Duration((r-i%4+4)%4)) // Make goroutines asynchronous, and let the sender go first.
			msg, ok := world.Broadcast(i%4, a[r])
			finishTimes[i][r] = time.Now()
			if !ok {
				t.Errorf("Goroutine %d: An unexpected quit signal was detected.", r)
			}
			if msg != a[i%4] {
				t.Errorf("Goroutine %d: msg: %v != root msg: %v.", r, msg, a[i%4])
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
	for _, times := range finishTimes {
		for i := 0; i < 3; i++ {
			for j := i + 1; j < 4; j++ {
				diff := times[j].Sub(times[i])
				if diff < 0 {
					diff = -diff
				}
				if diff < time.Microsecond {
					t.Error("Two broadcasts finished at the same time. Maybe an unexpected blocking exists.")
					return
				}
			}
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
	finishTimes := make([][4]time.Time, len(data))
	ctrl := New(4, func(world Communicator, commMap map[string]Communicator) {
		for i, a := range data {
			world.Barrier()
			r := world.Rank()
			time.Sleep(time.Millisecond * 10 * time.Duration((r-i%4+4)%4)) // Make goroutines asynchronous, and let the sender go first.
			msg, ok := world.Scatter(i%4, a[r])
			finishTimes[i][r] = time.Now()
			if !ok {
				t.Errorf("Goroutine %d: An unexpected quit signal was detected.", r)
			}
			if msg == nil {
				if i < 4 {
					t.Errorf("Goroutine %d: msg is nil but should be %v.", r, wanted[r])
				}
				continue
			}
			ints := msg.(sequence.IntDynamicArray)
			if ints.Len() != wanted[r].Len() {
				t.Errorf("Goroutine %d: msg: %v != %v.", r, ints, wanted[r])
				continue
			}
			for i := range ints {
				if ints[i] != wanted[r][i] {
					t.Errorf("Goroutine %d: msg: %v != %v.", r, ints, wanted[r])
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
	for _, times := range finishTimes {
		for i := 0; i < 3; i++ {
			for j := i + 1; j < 4; j++ {
				diff := times[j].Sub(times[i])
				if diff < 0 {
					diff = -diff
				}
				if diff < time.Microsecond {
					t.Error("Two broadcasts finished at the same time. Maybe an unexpected blocking exists.")
					return
				}
			}
		}
	}
}
