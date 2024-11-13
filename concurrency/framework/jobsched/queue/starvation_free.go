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
	"math"
	"runtime"

	"github.com/donyori/gogo/concurrency/framework/jobsched"
	"github.com/donyori/gogo/container/heap/pqueue"
	"github.com/donyori/gogo/errors"
)

// exponentialJobQueueMaker is a maker for creating job queues with
// a starvation-free job scheduling algorithm.
//
// The priority of jobs increases exponentially.
type exponentialJobQueueMaker[Job, Properties any] struct {
	// The number of goroutines to process jobs.
	n int

	// The base of the exponent.
	//
	// The greater b is, the faster the priority of jobs increases.
	b float64
}

// NewExponentialJobQueueMaker creates a job queue maker for creating job queues
// with a starvation-free job scheduling algorithm.
//
// The priority of jobs increases exponentially.
//
// n is the number of goroutines to process jobs.
// If n is nonpositive, runtime.NumCPU() is used instead.
//
// b is the base of the exponent.
// The greater b is, the faster the priority of jobs increases.
// If b is not greater than 1, 1.025 is used instead.
func NewExponentialJobQueueMaker[Job, Properties any](
	n int,
	b float64,
) jobsched.JobQueueMaker[Job, Properties] {
	m := &exponentialJobQueueMaker[Job, Properties]{n: n, b: b}
	if m.n <= 0 {
		m.n = runtime.NumCPU()
	}
	if m.b <= 1. {
		m.b = 1.025
	}
	return m
}

func (m *exponentialJobQueueMaker[Job, Properties]) New() jobsched.JobQueue[Job, Properties] {
	jq := &exponentialJobQueue[Job, Properties]{
		b:  m.b,
		rn: 1. / float64(m.n),
	}
	jq.pq = pqueue.New(jq.jobLess, 0)
	return jq
}

// exponentialJobQueue is a job queue implementing
// a starvation-free job scheduling algorithm.
//
// The priority of jobs increases exponentially.
type exponentialJobQueue[Job, Properties any] struct {
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

func (jq *exponentialJobQueue[Job, Properties]) Len() int {
	return jq.pq.Len()
}

func (jq *exponentialJobQueue[Job, Properties]) Enqueue(
	metaJob ...*jobsched.MetaJob[Job, Properties]) {
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

func (jq *exponentialJobQueue[Job, Properties]) Dequeue() Job {
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
func (jq *exponentialJobQueue[Job, Properties]) jobLess(
	a, b *jobsched.MetaJob[Job, float64]) bool {
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
func (jq *exponentialJobQueue[Job, Properties]) calculatePriority(
	mj *jobsched.MetaJob[Job, float64]) float64 {
	return (float64(mj.Meta.Priority) + 1) * math.Pow(jq.b, jq.t-mj.Meta.Custom)
}
