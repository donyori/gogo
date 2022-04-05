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

package jobsched

import (
	"math"
	"runtime"

	"github.com/donyori/gogo/container/pqueue"
)

// threshold is used by the default job queue to calculate the priority of jobs.
// It is slightly smaller than the positive solution of the equation:
//  1.025^x - x - 1 = 0
const threshold float64 = 218.302152071367373

// DefaultJobQueueMaker is a maker for creating a job queue with
// the default job scheduling algorithm, which is starvation-free.
type DefaultJobQueueMaker struct {
	// The number of goroutines to process jobs.
	// If non-positive, runtime.NumCPU() will be used instead.
	N int
}

// New creates a new job queue with the default job scheduling algorithm.
func (m *DefaultJobQueueMaker) New() JobQueue {
	n := m.N
	if n <= 0 {
		n = runtime.NumCPU()
	}
	jq := &defaultJobQueue{rn: 1 / float64(n)}
	jq.pq = pqueue.NewPriorityQueueMini(jq.jobLess)
	return jq
}

// defaultJobQueue is a implementation of interface JobQueue
// with the default job scheduling algorithm.
type defaultJobQueue struct {
	// Priority queue to manage jobs.
	pq pqueue.PriorityQueueMini

	// A parameter for calculating the priority of jobs,
	// equals rn by the product of the number of calls to the method Dequeue.
	t float64

	// Reciprocal of the number of goroutines to process jobs.
	rn float64
}

// Len returns the number of jobs in the queue.
func (djq *defaultJobQueue) Len() int {
	return djq.pq.Len()
}

// Enqueue adds jobs into the job queue.
//
// The framework guarantees that all items in jobs are never nil and
// have a non-zero Ct field.
func (djq *defaultJobQueue) Enqueue(jobs ...*Job) {
	if len(jobs) == 0 {
		return
	}
	a := make([]interface{}, len(jobs))
	for i := range jobs {
		jobs[i].CustAttr = djq.t
		a[i] = jobs[i]
	}
	djq.pq.Enqueue(a...)
}

// Dequeue pops a job in the queue and returns its data
// (i.e., the Data field of Job).
// It panics if the queue is nil or empty.
func (djq *defaultJobQueue) Dequeue() interface{} {
	job := djq.pq.Dequeue().(*Job)
	djq.t += djq.rn
	return job.Data
}

// jobLess is a function for the priority queue djq.pq.
// A job with a higher priority is "less" than a job with a lower priority.
func (djq *defaultJobQueue) jobLess(a, b interface{}) bool {
	ja, jb := a.(*Job), b.(*Job)
	pa, pb := djq.calculatePriority(ja), djq.calculatePriority(jb)
	if math.Abs(pa-pb) > 1e-3 {
		return pa > pb
	}
	return ja.Ct.Before(jb.Ct)
}

// calculatePriority calculates the ultimate priority of the job.
//
// The ultimate priority provides a chance for low-priority jobs
// to avoid the starvation problem.
//
// The ultimate priority (p_u) is calculated as follows:
//  p_u = (p+1) * min(1+t/n, 1.025^(t/n))
// where p is the priority specified by the user,
// t is a measure of the waiting time of the job, equal to the number of
// calls to the method Dequeue since the job was added to this queue,
// and n is the number of goroutines to process jobs.
func (djq *defaultJobQueue) calculatePriority(job *Job) float64 {
	x := djq.t - job.CustAttr.(float64) // = t/n.
	if x < threshold {
		x = math.Pow(1.025, x) // = 1.025^(t/n).
	} else {
		x++ // = 1+t/n.
	}
	return (float64(job.Pri) + 1) * x
}
