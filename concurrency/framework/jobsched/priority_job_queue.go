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

// PriorityJobQueueMaker is a maker for creating a job queue
// that schedules jobs according to their priority.
//
// A job with a higher priority (bigger Pri field) will be dequeued earlier.
// The jobs with the same priority will be managed in one sub-queue.
// The client can set the field Maker in order to customize
// the performance of sub-queues.
// If Maker is nil, a new FCFS job queue maker will be used.
type PriorityJobQueueMaker struct {
	// Job queue maker to create new sub-queues.
	//
	// Default: FcfsJobQueueMaker
	Maker JobQueueMaker
}

// New creates a new priority job queue.
func (m *PriorityJobQueueMaker) New() JobQueue {
	jq := &priorityJobQueue{
		pq: pqueue.NewPriorityQueue(func(a, b interface{}) bool {
			return a.(uint) > b.(uint)
		}),
		sqm:   make(map[uint]JobQueue),
		maker: m.Maker,
	}
	if jq.maker == nil {
		jq.maker = new(FcfsJobQueueMaker)
	}
	return jq
}

// priorityJobQueue is a job queue scheduling jobs according to their priority.
type priorityJobQueue struct {
	numJob int                  // The number of jobs in the queue.
	pq     pqueue.PriorityQueue // Priority queue to manage sub-queues.
	sqm    map[uint]JobQueue    // Sub-queue map.
	maker  JobQueueMaker        // Job queue maker to create new sub-queues.
}

// Len returns the number of jobs in the queue.
func (pjq *priorityJobQueue) Len() int {
	return pjq.numJob
}

// Enqueue adds jobs into the job queue.
//
// The framework guarantees that all items in jobs are never nil and
// have a non-zero Ct field.
func (pjq *priorityJobQueue) Enqueue(jobs ...*Job) {
	for _, job := range jobs {
		q, ok := pjq.sqm[job.Pri]
		if !ok {
			q = pjq.maker.New()
			pjq.sqm[job.Pri] = q
		}
		if q.Len() == 0 {
			pjq.pq.Enqueue(job.Pri)
		}
		q.Enqueue(job)
		pjq.numJob++
	}
}

// Dequeue pops a job in the queue and returns its data
// (i.e., the Data field of Job).
// It panics if the queue is nil or empty.
func (pjq *priorityJobQueue) Dequeue() interface{} {
	top := pjq.sqm[pjq.pq.Top().(uint)]
	data := top.Dequeue()
	pjq.numJob--
	if top.Len() == 0 {
		// Remove the sub-queue from pjq.pq, but still keep it in pjq.sqm,
		// to avoid too much memory allocation.
		pjq.pq.Dequeue()
	}
	return data
}
