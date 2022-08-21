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
	"math"
	"runtime"

	"github.com/donyori/gogo/concurrency/framework/jobsched"
	"github.com/donyori/gogo/container/heap/pqueue"
	"github.com/donyori/gogo/errors"
)

// ExponentialJobQueueMaker is a maker for creating a job queue with
// a starvation-free job scheduling algorithm.
//
// The priority of jobs increases exponentially.
type ExponentialJobQueueMaker[Job any] struct {
	// The number of goroutines to process jobs.
	//
	// If N is non-positive, runtime.NumCPU() will be used instead.
	N int

	// The base of the exponent.
	//
	// The greater B is, the faster the priority of jobs increases.
	//
	// If B is not greater than 1, 1.025 will be used instead.
	B float64
}

// New creates a new exponential job queue.
func (m *ExponentialJobQueueMaker[Job]) New() jobsched.JobQueue[Job, jobsched.NoProperty] {
	var n int
	var b float64
	if m != nil {
		n = m.N
		b = m.B
	}
	if n <= 0 {
		n = runtime.NumCPU()
	}
	if b <= 1. {
		b = 1.025
	}
	jq := &exponentialJobQueue[Job]{
		b:  b,
		rn: 1. / float64(n),
	}
	jq.pq = pqueue.NewPriorityQueue(jq.jobLess)
	return jq
}

// exponentialJobQueue is a job queue implementing
// a starvation-free job scheduling algorithm.
//
// The priority of jobs increases exponentially.
type exponentialJobQueue[Job any] struct {
	pq pqueue.PriorityQueue[*jobsched.MetaJob[Job, float64]]

	// The base of the exponent.
	b float64

	// Reciprocal of the number of goroutines to process jobs.
	rn float64

	// A parameter for calculating the priority of jobs,
	// equal to the product of rn and the number of
	// calls to its method Dequeue.
	t float64
}

// Len returns the number of jobs in the queue.
func (jq *exponentialJobQueue[Job]) Len() int {
	return jq.pq.Len()
}

// Enqueue adds metaJob into the queue.
//
// The framework guarantees that all items in metaJob are never nil
// and have a non-zero creation time in their meta information.
func (jq *exponentialJobQueue[Job]) Enqueue(metaJob ...*jobsched.MetaJob[Job, jobsched.NoProperty]) {
	if len(metaJob) == 0 {
		return
	}
	a := make([]*jobsched.MetaJob[Job, float64], len(metaJob))
	for i, mj := range metaJob {
		a[i] = &jobsched.MetaJob[Job, float64]{
			Meta: jobsched.Meta[float64]{
				Priority:     mj.Meta.Priority,
				CreationTime: mj.Meta.CreationTime,
				Custom:       jq.t,
			},
			Job: mj.Job,
		}
	}
	jq.pq.Enqueue(a...)
}

// Dequeue removes and returns a job in the queue.
//
// It panics if the queue is nil or empty.
func (jq *exponentialJobQueue[Job]) Dequeue() Job {
	n := jq.pq.Len()
	if n == 0 {
		panic(errors.AutoMsg(emptyQueuePanicMessage))
	}
	job := jq.pq.Dequeue().Job
	if n > 1 {
		jq.t += jq.rn
	} else {
		jq.t = 0. // the job queue is empty now; reset t to 0
	}
	return job
}

// jobLess is a github.com/donyori/gogo/function/compare.LessFunc
// for the priority queue jq.pq.
//
// The higher the priority, the "less" the job.
// If two jobs have nearly the same priority (difference less than 0.001),
// the earlier its creation time, the "less" the job.
func (jq *exponentialJobQueue[Job]) jobLess(a, b *jobsched.MetaJob[Job, float64]) bool {
	pa, pb := jq.calculatePriority(a), jq.calculatePriority(b)
	if math.Abs(pa-pb) < 1e-3 {
		return a.Meta.CreationTime.Before(b.Meta.CreationTime)
	}
	return pa > pb
}

// calculatePriority calculates the ultimate priority of a job.
//
// The ultimate priority provides a chance for low-priority jobs
// to avoid the starvation problem.
//
// The ultimate priority (u) is calculated as follows:
//
//	u = (p+1) * b^(t/n)
//
// where p is the priority specified by the client,
// b is a number greater than 1,
// t is a measure of the waiting time of the job, equal to the number of
// calls to the method Dequeue since the job was added to this queue,
// and n is the number of goroutines to process jobs.
func (jq *exponentialJobQueue[Job]) calculatePriority(mj *jobsched.MetaJob[Job, float64]) float64 {
	return (float64(mj.Meta.Priority) + 1) * math.Pow(jq.b, jq.t-mj.Meta.Custom)
}
