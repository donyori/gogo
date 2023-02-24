// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2023  Yuan Gao
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

const emptyQueuePanicMessage = "job queue is empty"

// fcfsJobQueueMaker is a maker for creating job queues with
// FCFS (first come, first served) scheduling algorithm.
//
// The FCFS job queue implements a simple scheduling algorithm that
// queues jobs in the order that they arrive.
// All the properties (such as the priority and creation time)
// of jobs are ignored.
type fcfsJobQueueMaker[Job, Properties any] struct{}

// NewFCFSJobQueueMaker returns a job queue maker that creates job queues
// with FCFS (first come, first served) scheduling algorithm.
//
// The FCFS job queue implements a simple scheduling algorithm that
// queues jobs in the order that they arrive.
// All the properties (such as the priority and creation time)
// of jobs are ignored.
func NewFCFSJobQueueMaker[Job, Properties any]() jobsched.JobQueueMaker[Job, Properties] {
	return fcfsJobQueueMaker[Job, Properties]{}
}

func (m fcfsJobQueueMaker[Job, Properties]) New() jobsched.JobQueue[Job, Properties] {
	return new(fcfsJobQueue[Job, Properties])
}

// fcfsJobQueue is an FCFS (first come, first served) job queue.
type fcfsJobQueue[Job, Properties any] []Job

func (jq *fcfsJobQueue[Job, Properties]) Len() int {
	return len(*jq)
}

func (jq *fcfsJobQueue[Job, Properties]) Enqueue(metaJob ...*jobsched.MetaJob[Job, Properties]) {
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
