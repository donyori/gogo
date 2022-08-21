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
	"github.com/donyori/gogo/errors"
)

const emptyQueuePanicMessage string = "job queue is empty"

// FcfsJobQueueMaker is a maker for creating a job queue with
// FCFS (first come, first served) scheduling algorithm.
//
// The FCFS job queue implements a simple scheduling algorithm that
// queues jobs in the order that they arrive.
// All the properties (such as the priority and creation time)
// of jobs will be ignored.
type FcfsJobQueueMaker[Job any] struct{}

// New creates a new FCFS (first come, first served) job queue.
func (m FcfsJobQueueMaker[Job]) New() jobsched.JobQueue[Job, jobsched.NoProperty] {
	return new(fcfsJobQueue[Job])
}

// fcfsJobQueue is an FCFS (first come, first served) job queue.
type fcfsJobQueue[Job any] []Job

// Len returns the number of jobs in the queue.
func (jq *fcfsJobQueue[Job]) Len() int {
	return len(*jq)
}

// Enqueue adds metaJob into the queue.
//
// The framework guarantees that all items in metaJob are never nil
// and have a non-zero creation time in their meta information.
func (jq *fcfsJobQueue[Job]) Enqueue(metaJob ...*jobsched.MetaJob[Job, jobsched.NoProperty]) {
	if len(metaJob) == 0 {
		return
	}
	i := len(*jq)
	*jq = append(*jq, make([]Job, len(metaJob))...)
	for _, mj := range metaJob {
		(*jq)[i], i = mj.Job, i+1
	}
}

// Dequeue removes and returns a job in the queue.
//
// It panics if the queue is nil or empty.
func (jq *fcfsJobQueue[Job]) Dequeue() Job {
	if len(*jq) == 0 {
		panic(errors.AutoMsg(emptyQueuePanicMessage))
	}
	var job Job
	*jq, (*jq)[0], job = (*jq)[1:], job, (*jq)[0] // where (*jq)[0] = job is to avoid memory leak
	return job
}