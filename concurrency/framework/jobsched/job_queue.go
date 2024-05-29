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

package jobsched

import "github.com/donyori/gogo/errors"

// JobQueue is a queue for job scheduling.
// The client can customize a job scheduling algorithm
// by implementing this interface.
//
// The first type parameter Job is the type of jobs.
// The second type parameter Properties is the type of custom properties
// in the meta information of jobs.
//
// A JobQueue is used in only one goroutine.
// Its implementation does not need to consider concurrency issues.
type JobQueue[Job, Properties any] interface {
	// Len returns the number of jobs in the queue.
	Len() int

	// Enqueue adds metaJob into the queue.
	//
	// The framework guarantees that all items in metaJob are never nil
	// and have a nonzero creation time in their meta information.
	Enqueue(metaJob ...*MetaJob[Job, Properties])

	// Dequeue removes and returns a job in the queue.
	//
	// It panics if the queue is nil or empty.
	Dequeue() Job
}

// JobQueueMaker is a maker for creating job queues.
//
// The first type parameter Job is the type of jobs.
// The second type parameter Properties is the type of custom properties
// in the meta information of jobs.
//
// It has a method New, with no parameter.
// The client should set any parameters required for creating job queues
// in the instance of this interface.
type JobQueueMaker[Job, Properties any] interface {
	// New creates a new job queue.
	New() JobQueue[Job, Properties]
}

// The following code is copied from
// "github.com/donyori/gogo/concurrency/framework/jobsched/queue/fcfs.go"
// to avoid cycle import.

// emptyQueuePanicMessage is the panic message
// to indicate that the job queue is empty.
const emptyQueuePanicMessage = "job queue is empty"

// fcfsJobQueue is an FCFS (first come, first served) job queue.
type fcfsJobQueue[Job, Properties any] []Job

func (jq *fcfsJobQueue[Job, Properties]) Len() int {
	return len(*jq)
}

func (jq *fcfsJobQueue[Job, Properties]) Enqueue(
	metaJob ...*MetaJob[Job, Properties]) {
	if len(metaJob) == 0 {
		return
	}
	i := len(*jq)
	*jq = append(*jq, make([]Job, len(metaJob))...)
	for _, mj := range metaJob {
		(*jq)[i], i = mj.Job, i+1
	}
}

func (jq *fcfsJobQueue[Job, Properties]) Dequeue() Job {
	if len(*jq) == 0 {
		panic(errors.AutoMsg(emptyQueuePanicMessage))
	}
	var job Job
	*jq, (*jq)[0], job = (*jq)[1:], job, (*jq)[0] // where (*jq)[0] = job is to avoid memory leak
	return job
}
