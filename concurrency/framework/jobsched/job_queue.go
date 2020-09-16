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

import "time"

// A unit representing a job.
type Job struct {
	// Data for the job handler.
	// Note that only Data will be passed to the job handler.
	// Other fields will only be used for the job scheduling.
	Data interface{}

	// Priority, a non-negative integer.
	//
	// The bigger the value, the higher the priority.
	// The default Pri (0) corresponds to the lowest priority.
	Pri uint

	// Creation time.
	//
	// A zero-value Ct field will be set to time.Now() before adding to
	// the job queue by the framework.
	Ct time.Time

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
	//
	// The framework guarantees that all items in jobs are never nil and
	// have a non-zero Ct field.
	Enqueue(jobs ...*Job)

	// Pop a job in the queue and return its data (i.e., the Data field of Job).
	// It panics if the queue is nil or empty.
	Dequeue() interface{}
}

// A function to create a new job queue.
//
// n is the number of goroutines to process jobs, passed by the function New.
type JobQueueMaker func(n int) JobQueue
