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

package jobsched

import (
	"math/rand"
	"testing"
	"time"
)

func TestCtJobQueue(t *testing.T) {
	random := rand.New(rand.NewSource(10))
	now := time.Now()
	const N = 10
	jobs := make([]*Job, N)
	for i := 0; i < 3; i++ {
		jobs[i] = &Job{
			Data: i,
			Pri:  uint(N - i),
			Ct:   now,
		}
	}
	jobs[3] = &Job{
		Data: 3,
		Pri:  0,
		Ct:   now.Add(1),
	}
	for i := 4; i < 8; i++ {
		jobs[i] = &Job{
			Data: 11 - i,
			Pri:  uint(i),
			Ct:   now.Add(2),
		}
	}
	for i := 8; i < N; i++ {
		jobs[i] = &Job{
			Data: N + 7 - i,
			Pri:  uint(i),
			Ct:   now.Add(3),
		}
	}
	random.Shuffle(N, func(i, j int) {
		jobs[i], jobs[j] = jobs[j], jobs[i]
	})
	var wanted, dqs [N]interface{}
	for i := range wanted {
		wanted[i] = i
	}

	cjq := new(CtJobQueueMaker).New()
	cjq.Enqueue(jobs[0])
	cjq.Enqueue(jobs[1:6]...)
	cjq.Enqueue(jobs[6:]...)

	for i := range dqs {
		dqs[i] = cjq.Dequeue()
	}
	if n := cjq.Len(); n > 0 {
		t.Errorf("cjq.Len(): %d != 0.", n)
	}
	if dqs != wanted {
		t.Errorf("Dequeued: %v, wanted: %v.", dqs, wanted)
	}
}
