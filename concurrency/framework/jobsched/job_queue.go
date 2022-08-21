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

// JobQueue is a queue for job scheduling.
// The client can customize a job scheduling algorithm
// by implementing this interface.
//
// The first type parameter Job is the type of jobs.
// The second type parameter Properties is the type of custom properties
// in the meta information of jobs.
//
// A JobQueue will be used in only one goroutine.
// Its implementation does not need to consider concurrency issues.
type JobQueue[Job, Properties any] interface {
	// Len returns the number of jobs in the queue.
	Len() int

	// Enqueue adds metaJob into the queue.
	//
	// The framework guarantees that all items in metaJob are never nil
	// and have a non-zero creation time in their meta information.
	Enqueue(metaJob ...*MetaJob[Job, Properties])

	// Dequeue removes and returns a job in the queue.
	//
	// It panics if the queue is nil or empty.
	Dequeue() Job
}

// JobQueueMaker is a maker for creating a job queue.
//
// The first type parameter Job is the type of jobs.
// The second type parameter Properties is the type of custom properties
// in the meta information of jobs.
//
// It has a method New, with no parameter.
// The client should set any parameters required for creating a job queue
// in the instance of this interface.
type JobQueueMaker[Job, Properties any] interface {
	// New creates a new job queue.
	New() JobQueue[Job, Properties]
}
