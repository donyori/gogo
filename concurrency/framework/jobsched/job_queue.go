// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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
	"time"

	"github.com/donyori/gogo/container/pqueue"
)

// A unit representing a job.
type Job struct {
	// Data for the job handler.
	// Note that only Data will be passed to the job handler.
	// Other fields will only be used for the job scheduling.
	Data interface{}

	Pri uint      // Priority, a non-negative integer.
	Ct  time.Time // Creation time.

	// Custom attribute used to customize job scheduling algorithm.
	// It is available in the method Enqueue of JobQueue,
	// but won't be passed to the job handler.
	CustAttr interface{}
}

// A job queue for the job scheduling.
// The client can customize the job scheduling algorithm
// by implementing this interface.
//
// The job queue is used by only one goroutine.
// Don't worry about multiple access problems such as race condition
// when implementing this interface.
type JobQueue interface {
	// Return the number of jobs in the queue.
	Len() int

	// Add jobs into the job queue.
	Enqueue(jobs ...*Job)

	// Pop a job in the queue and return its data (i.e., the Data field of Job).
	// It panics if the queue is nil or empty.
	Dequeue() interface{}
}

// A function to create a new job queue.
//
// n is the number of goroutines to process jobs, passed by the function New.
type JobQueueMaker func(n int) JobQueue

// Threshold used by the default job queue to calculate the priority of jobs.
// It is slightly smaller than the positive solution of the equation:
//  1.2^(x/4) = x + 1
const threshold float64 = 101.596383011308744

// Default job queue maker.
func defaultJobQueueMaker(n int) JobQueue {
	jq := &jobQueue{rn: 1 / float64(n)}
	jq.pq = pqueue.NewPriorityQueueMini(jq.jobLess)
	return jq
}

// A default implementation of interface JobQueue.
type jobQueue struct {
	// Priority queue to manage jobs.
	pq pqueue.PriorityQueueMini

	// A parameter for calculating the priority of jobs,
	// equals rn by the product of the number of calls to the method Dequeue.
	t float64

	// Reciprocal of the number of goroutines to process jobs.
	rn float64
}

func (jq *jobQueue) Len() int {
	return jq.pq.Len()
}

func (jq *jobQueue) Enqueue(jobs ...*Job) {
	if len(jobs) == 0 {
		return
	}
	var now time.Time
	a := make([]interface{}, 0, len(jobs))
	for _, job := range jobs {
		if job == nil {
			continue
		}
		if job.Ct.IsZero() {
			if now.IsZero() {
				now = time.Now()
			}
			job.Ct = now
		}
		job.CustAttr = jq.t
		a = append(a, job)
	}
	jq.pq.Enqueue(a...)
}

func (jq *jobQueue) Dequeue() interface{} {
	job := jq.pq.Dequeue().(*Job)
	jq.t += jq.rn
	return job.Data
}

// Less function for the priority queue jq.pq.
// A job with a higher priority is "less" than a job with a lower priority.
func (jq *jobQueue) jobLess(a, b interface{}) bool {
	ja := a.(*Job)
	jb := b.(*Job)
	pa := jq.calcPriority(ja)
	pb := jq.calcPriority(jb)
	if math.Abs(pa-pb) > 1e-3 {
		return pa > pb
	}
	return ja.Ct.Before(jb.Ct)
}

// Calculate the ultimate priority of the job.
//
// The ultimate priority provides a chance for low-priority jobs
// to avoid the starvation problem.
//
// The ultimate priority (p_u) is calculated as follows:
//  p_u = (p+1) * min(1+t/n, 1.2^(t/4n))
// where p is the priority given by the user,
// t is a measure of the waiting time of the job, equal to the number of
// calls to the method Dequeue since the job was added to this queue,
// and n is the number of goroutines to process jobs.
func (jq *jobQueue) calcPriority(job *Job) float64 {
	x := jq.t - job.CustAttr.(float64) // = t/n.
	if x < threshold {
		x = math.Pow(1.2, x/4) // = 1.2^(t/4n).
	} else {
		x++ // = 1+t/n.
	}
	return (float64(job.Pri) + 1) * x
}
