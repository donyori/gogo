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

package jobsched

// A maker for creating a job queue with
// FCFS (first come, first served) scheduling algorithm.
//
// The FCFS job queue implements a simplest scheduling algorithm that
// queues jobs in the order that they arrive.
// All the properties (such as the creation time and the priority) of jobs
// will be ignored.
type FcfsJobQueueMaker struct{}

func (m *FcfsJobQueueMaker) New() JobQueue {
	return new(fcfsJobQueue)
}

// FCFS (first come, first served) job queue.
type fcfsJobQueue []interface{}

func (fjq fcfsJobQueue) Len() int {
	return len(fjq)
}

func (fjq *fcfsJobQueue) Enqueue(jobs ...*Job) {
	if len(jobs) == 0 {
		return
	}
	data := make([]interface{}, len(jobs))
	for i := range jobs {
		data[i] = jobs[i].Data
	}
	*fjq = append(*fjq, data...)
}

func (fjq *fcfsJobQueue) Dequeue() interface{} {
	var r interface{}
	*fjq, r = (*fjq)[1:], (*fjq)[0]
	return r
}
