// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2026  Yuan Gao
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
	"github.com/donyori/gogo/container/queue"
	"github.com/donyori/gogo/errors"
)

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
func NewFCFSJobQueueMaker[
	Job any,
	Properties any,
]() JobQueueMaker[Job, Properties] {
	return fcfsJobQueueMaker[Job, Properties]{}
}

func (m fcfsJobQueueMaker[Job, Properties]) New() JobQueue[Job, Properties] {
	return &fcfsJobQueue[Job, Properties]{q: queue.New[Job](0)}
}

// fcfsJobQueue is an FCFS (first come, first served) job queue.
type fcfsJobQueue[Job, Properties any] struct {
	q queue.Queue[Job]
}

func (jq *fcfsJobQueue[Job, Properties]) Len() int {
	return jq.q.Len()
}

func (jq *fcfsJobQueue[Job, Properties]) Enqueue(
	metaJob ...*MetaJob[Job, Properties],
) {
	if len(metaJob) == 0 {
		return
	}

	jq.q.Reserve(jq.q.Len() + len(metaJob))

	for _, mj := range metaJob {
		jq.q.Enqueue(mj.Job)
	}
}

func (jq *fcfsJobQueue[Job, Properties]) Dequeue() Job {
	if jq.q.Len() == 0 {
		panic(errors.AutoMsg(emptyQueuePanicMessage))
	}

	return jq.q.Dequeue()
}
