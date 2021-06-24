// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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
	"sort"
	"testing"
)

func TestPriorityJobQueue(t *testing.T) {
	const N = 12
	wanted := [N]uint{0, 1, 1, 9, 2, 7, 8, 0, 3, 1, 9, 3}
	jobs := make([]*Job, len(wanted))
	for i := range jobs {
		jobs[i] = &Job{
			Data: wanted[i],
			Pri:  wanted[i],
		}
	}
	sort.Slice(wanted[:], func(i, j int) bool {
		return wanted[i] > wanted[j]
	})
	pjq := new(PriorityJobQueueMaker).New()
	for epoch := 0; epoch < 2; epoch++ {
		// Test twice on one job queue.
		var idx int
		for _, length := range []int{1, 1, 3, 3, 4} {
			if n := pjq.Len(); n != idx {
				t.Errorf("pjq.Len(): %d != %d.", n, idx)
			}
			pjq.Enqueue(jobs[idx : idx+length]...)
			idx += length
		}
		if n := pjq.Len(); n != idx {
			t.Errorf("pjq.Len(): %d != %d.", n, idx)
		}
		var dqs [N]uint
		for i := range dqs {
			dqs[i] = pjq.Dequeue().(uint)
		}
		if n := pjq.Len(); n > 0 {
			t.Errorf("pjq.Len(): %d != 0.", n)
		}
		if dqs != wanted {
			t.Errorf("Dequeued: %v, wanted: %v.", dqs, wanted)
		}
	}
}
