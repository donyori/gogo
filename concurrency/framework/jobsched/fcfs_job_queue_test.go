// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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

import "testing"

func TestFcfsJobQueue(t *testing.T) {
	const N = 10
	jobs := make([]*Job, N)
	var wanted, dqs [N]interface{}
	for i := range jobs {
		jobs[i] = &Job{Data: i}
		wanted[i] = i
	}
	fjq := new(FcfsJobQueueMaker).New()
	fjq.Enqueue(jobs[0])
	fjq.Enqueue(jobs[1:5]...)
	fjq.Enqueue(jobs[5:]...)
	for i := range dqs {
		dqs[i] = fjq.Dequeue()
	}
	if n := fjq.Len(); n > 0 {
		t.Errorf("fjq.Len(): %d != 0.", n)
	}
	if dqs != wanted {
		t.Errorf("Dequeued: %v, wanted: %v.", dqs, wanted)
	}
}
