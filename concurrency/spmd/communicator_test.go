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
	"math/rand"
	"testing"
	"time"
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
		{1, nil, nil, nil},
		{2, 0.3, 4, 5},
		{nil, nil, "Hello", nil},
		{nil, nil, nil, complex(1, -1)},
		{},
	}
	finishTimes := make([][4]time.Time, len(data))
	rand.Seed(time.Now().UnixNano())
	prs := Run(4, func(world Communicator, commMap map[string]Communicator) {
		for i, array := range data {
			r := world.Rank()
			time.Sleep(time.Microsecond * time.Duration(rand.Int63n(1000)+100)) // Random delay.
			msg, ok := world.Broadcast(i%4, array[r])
			finishTimes[i][r] = time.Now()
			if !ok {
				t.Error("An unexpected quit signal was detected.")
			}
			if msg != array[i%4] {
				t.Errorf("msg: %v != root msg: %v.", msg, array[i%4])
			}
		}
	}, nil)
	if len(prs) > 0 {
		t.Errorf("Panic: %v.", prs)
	}
	for _, times := range finishTimes {
		for i := 1; i < 4; i++ {
			diff := times[0].Sub(times[i])
			if diff < 0 {
				diff = -diff
			}
			if diff > time.Microsecond {
				return
			}
		}
	}
	t.Error("All broadcasts finished at the same time. Maybe an unexpected blocking exists.")
}
