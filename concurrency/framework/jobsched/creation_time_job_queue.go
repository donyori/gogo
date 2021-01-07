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

import "github.com/donyori/gogo/container/pqueue"

// CtJobQueueMaker is a maker for creating a job queue that schedules jobs
// according to their creation time, and then priority.
//
// A job with an earlier creation time (corresponding to Ct field)
// will be dequeued earlier.
// The jobs with the same creation time will be scheduled according to
// their priority. The higher the priority, the earlier dequeued.
type CtJobQueueMaker struct{}

// New creates a new creation time job queue.
func (m *CtJobQueueMaker) New() JobQueue {
	return &ctJobQueue{pqueue.NewPriorityQueueMini(func(a, b interface{}) bool {
		ja, jb := a.(*Job), b.(*Job)
		if ja.Ct.Equal(jb.Ct) {
			return ja.Pri > jb.Pri
		}
		return ja.Ct.Before(jb.Ct)
	})}
}

// ctJobQueue is a job queue scheduling jobs according to their creation time,
// and then priority.
type ctJobQueue struct {
	pq pqueue.PriorityQueueMini // Priority queue to manage jobs.
}

// Len returns the number of jobs in the queue.
func (cjq *ctJobQueue) Len() int {
	return cjq.pq.Len()
}

// Enqueue adds jobs into the job queue.
//
// The framework guarantees that all items in jobs are never nil and
// have a non-zero Ct field.
func (cjq *ctJobQueue) Enqueue(jobs ...*Job) {
	if len(jobs) == 0 {
		return
	}
	a := make([]interface{}, len(jobs))
	for i := range jobs {
		a[i] = jobs[i]
	}
	cjq.pq.Enqueue(a...)
}

// Dequeue pops a job in the queue and returns its data
// (i.e., the Data field of Job).
// It panics if the queue is nil or empty.
func (cjq *ctJobQueue) Dequeue() interface{} {
	return cjq.pq.Dequeue().(*Job).Data
}
