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

package queue

import (
	"github.com/donyori/gogo/concurrency/framework/jobsched"
	"github.com/donyori/gogo/container/heap/pqueue"
	"github.com/donyori/gogo/errors"
)

// PriorityFirstJobQueueMaker is a maker for creating a job queue that
// schedules jobs according to their priority, and then creation time.
//
// A job with a higher priority (the field Priority in its meta information)
// will be dequeued earlier.
// The jobs with the same priority will be scheduled according to their
// creation time (the field CreationTime in the meta information).
// The earlier the creation time, the earlier dequeued.
type PriorityFirstJobQueueMaker[Job any] struct{}

// New creates a new priority-first job queue.
func (m PriorityFirstJobQueueMaker[Job]) New() jobsched.JobQueue[Job, jobsched.NoProperty] {
	return &priorityJobQueue[Job, jobsched.NoProperty]{
		pq: pqueue.New(func(a, b *jobsched.MetaJob[Job, jobsched.NoProperty]) bool {
			if a.Meta.Priority == b.Meta.Priority {
				return a.Meta.CreationTime.Before(b.Meta.CreationTime)
			}
			return a.Meta.Priority > b.Meta.Priority
		}),
	}
}

// CreationTimeFirstJobQueueMaker is a maker for creating a job queue that
// schedules jobs according to their creation time, and then priority.
//
// A job with an earlier creation time (the field CreationTime
// in its meta information) will be dequeued earlier.
// The jobs with the same creation time will be scheduled according to
// their priority (the field Priority in the meta information).
// The higher the priority, the earlier dequeued.
type CreationTimeFirstJobQueueMaker[Job any] struct{}

// New creates a new creation-time-first job queue.
func (m CreationTimeFirstJobQueueMaker[Job]) New() jobsched.JobQueue[Job, jobsched.NoProperty] {
	return &priorityJobQueue[Job, jobsched.NoProperty]{
		pq: pqueue.New(func(a, b *jobsched.MetaJob[Job, jobsched.NoProperty]) bool {
			if a.Meta.CreationTime.Equal(b.Meta.CreationTime) {
				return a.Meta.Priority > b.Meta.Priority
			}
			return a.Meta.CreationTime.Before(b.Meta.CreationTime)
		}),
	}
}

// priorityJobQueue wraps
// github.com/donyori/gogo/container/heap/pqueue.PriorityQueue
// to a github.com/donyori/gogo/concurrency/framework/jobsched.JobQueue.
//
// Its type parameters are consistent with that of
// github.com/donyori/gogo/concurrency/framework/jobsched.JobQueue.
type priorityJobQueue[Job, Properties any] struct {
	pq pqueue.PriorityQueue[*jobsched.MetaJob[Job, Properties]]
}

// Len returns the number of jobs in the queue.
func (jq *priorityJobQueue[Job, Properties]) Len() int {
	return jq.pq.Len()
}

// Enqueue adds metaJob into the queue.
//
// The framework guarantees that all items in metaJob are never nil
// and have a non-zero creation time in their meta information.
func (jq *priorityJobQueue[Job, Properties]) Enqueue(metaJob ...*jobsched.MetaJob[Job, Properties]) {
	if len(metaJob) == 0 {
		return
	}
	jq.pq.Enqueue(metaJob...)
}

// Dequeue removes and returns a job in the queue.
//
// It panics if the queue is nil or empty.
func (jq *priorityJobQueue[Job, Properties]) Dequeue() Job {
	if jq.pq.Len() == 0 {
		panic(errors.AutoMsg(emptyQueuePanicMessage))
	}
	return jq.pq.Dequeue().Job
}
