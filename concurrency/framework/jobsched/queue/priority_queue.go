// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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
	"github.com/donyori/gogo/function/compare"
)

// priorityFirstJobQueueMaker is a maker for creating job queues that
// schedule jobs according to their priority, and then creation time.
//
// A job with a higher priority (the field Priority in its meta information)
// is dequeued earlier.
// The jobs with the same priority are scheduled according to their
// creation time (the field CreationTime in the meta information).
// The earlier the creation time, the earlier dequeued.
type priorityFirstJobQueueMaker[Job, Properties any] struct{}

// NewPriorityFirstJobQueueMaker returns a job queue maker for
// creating job queues that schedule jobs according to their priority,
// and then creation time.
//
// A job with a higher priority (the field Priority in its meta information)
// is dequeued earlier.
// The jobs with the same priority are scheduled according to their
// creation time (the field CreationTime in the meta information).
// The earlier the creation time, the earlier dequeued.
func NewPriorityFirstJobQueueMaker[Job, Properties any]() jobsched.JobQueueMaker[Job, Properties] {
	return priorityFirstJobQueueMaker[Job, Properties]{}
}

func (m priorityFirstJobQueueMaker[Job, Properties]) New() jobsched.JobQueue[Job, Properties] {
	return &priorityJobQueue[Job, Properties]{
		pq: pqueue.New(func(a, b *jobsched.MetaJob[Job, Properties]) bool {
			if a.Meta.Priority == b.Meta.Priority {
				return a.Meta.CreationTime.Before(b.Meta.CreationTime)
			}
			return a.Meta.Priority > b.Meta.Priority
		}, nil),
	}
}

// creationTimeFirstJobQueueMaker is a maker for creating job queues that
// schedule jobs according to their creation time, and then priority.
//
// A job with an earlier creation time (the field CreationTime
// in its meta information) is dequeued earlier.
// The jobs with the same creation time are scheduled according to
// their priority (the field Priority in the meta information).
// The higher the priority, the earlier dequeued.
type creationTimeFirstJobQueueMaker[Job, Properties any] struct{}

// NewCreationTimeFirstJobQueueMaker returns a job queue maker for
// creating job queues that schedule jobs according to their creation time,
// and then priority.
//
// A job with a higher priority (the field Priority in its meta information)
// is dequeued earlier.
// The jobs with the same priority are scheduled according to their
// creation time (the field CreationTime in the meta information).
// The earlier the creation time, the earlier dequeued.
func NewCreationTimeFirstJobQueueMaker[Job, Properties any]() jobsched.JobQueueMaker[Job, Properties] {
	return creationTimeFirstJobQueueMaker[Job, Properties]{}
}

func (m creationTimeFirstJobQueueMaker[Job, Properties]) New() jobsched.JobQueue[Job, Properties] {
	return &priorityJobQueue[Job, Properties]{
		pq: pqueue.New(func(a, b *jobsched.MetaJob[Job, Properties]) bool {
			if a.Meta.CreationTime.Equal(b.Meta.CreationTime) {
				return a.Meta.Priority > b.Meta.Priority
			}
			return a.Meta.CreationTime.Before(b.Meta.CreationTime)
		}, nil),
	}
}

// jobPriorityQueueMaker is a maker for creating job queues based on
// a priority queue.
// The job queues schedule jobs according to their custom priority.
//
// The priority of jobs is determined by its lessFn.
// The "less" the job, the higher its priority, and the earlier dequeued.
type jobPriorityQueueMaker[Job, Properties any] struct {
	// A function to determine which job has higher priority.
	//
	// The "less" the job, the higher its priority, and the earlier dequeued.
	//
	// lessFn must describe a strict weak ordering.
	// See <https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings>
	// for details.
	//
	// Note that floating-point comparison
	// (the < operator on float32 or float64 values)
	// is not a strict weak ordering
	// when not-a-number (NaN) values are involved.
	//
	// The framework guarantees that arguments passed to lessFn are never nil
	// and have a nonzero creation time in their meta information.
	lessFn compare.LessFunc[*jobsched.MetaJob[Job, Properties]]
}

// NewJobPriorityQueueMaker creates a job queue maker for creating job queues
// based on a priority queue.
// The job queues schedule jobs according to their custom priority.
//
// The priority of jobs is determined by the function lessFn.
// The "less" the job, the higher its priority, and the earlier dequeued.
//
// lessFn must describe a strict weak ordering.
// See <https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings>
// for details.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a strict weak ordering when not-a-number (NaN) values are involved.
//
// The framework guarantees that arguments passed to lessFn are never nil
// and have a nonzero creation time in their meta information.
//
// NewJobPriorityQueueMaker panics if lessFn is nil.
func NewJobPriorityQueueMaker[Job, Properties any](
	lessFn compare.LessFunc[*jobsched.MetaJob[Job, Properties]],
) jobsched.JobQueueMaker[Job, Properties] {
	if lessFn == nil {
		panic(errors.AutoMsg("lessFn is nil"))
	}
	return &jobPriorityQueueMaker[Job, Properties]{lessFn: lessFn}
}

func (m *jobPriorityQueueMaker[Job, Properties]) New() jobsched.JobQueue[Job, Properties] {
	return &priorityJobQueue[Job, Properties]{
		pq: pqueue.New(m.lessFn, nil),
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

func (jq *priorityJobQueue[Job, Properties]) Len() int {
	return jq.pq.Len()
}

func (jq *priorityJobQueue[Job, Properties]) Enqueue(
	metaJob ...*jobsched.MetaJob[Job, Properties]) {
	if len(metaJob) == 0 {
		return
	}
	jq.pq.Enqueue(metaJob...)
}

func (jq *priorityJobQueue[Job, Properties]) Dequeue() Job {
	if jq.pq.Len() == 0 {
		panic(errors.AutoMsg(emptyQueuePanicMessage))
	}
	return jq.pq.Dequeue().Job
}
