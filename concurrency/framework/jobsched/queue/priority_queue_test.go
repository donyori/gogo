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

package queue_test

import (
	"testing"

	"github.com/donyori/gogo/concurrency/framework/jobsched"
	"github.com/donyori/gogo/concurrency/framework/jobsched/queue"
)

func TestPriorityFirstJobQueue(t *testing.T) {
	testJobQueueFunc(
		t,
		queue.NewPriorityFirstJobQueueMaker[int, jobsched.NoProperty](),
		makeWant(func(a, b *jobsched.MetaJob[int, jobsched.NoProperty]) bool {
			if a.Meta.Priority == b.Meta.Priority {
				return a.Meta.CreationTime.Before(b.Meta.CreationTime)
			}
			return a.Meta.Priority > b.Meta.Priority
		}),
	)
}

func TestCreationTimeFirstJobQueue(t *testing.T) {
	testJobQueueFunc(
		t,
		queue.NewCreationTimeFirstJobQueueMaker[int, jobsched.NoProperty](),
		makeWant(func(a, b *jobsched.MetaJob[int, jobsched.NoProperty]) bool {
			if a.Meta.CreationTime.Equal(b.Meta.CreationTime) {
				return a.Meta.Priority > b.Meta.Priority
			}
			return a.Meta.CreationTime.Before(b.Meta.CreationTime)
		}),
	)
}

func TestJobPriorityQueue(t *testing.T) {
	// Define a special priority rule:
	// A job has a higher priority if its Meta.Priority is closer to 100.
	// For two jobs with the same |Meta.Priority-100|,
	// the earlier the Meta.CreationTime, the higher the priority.
	lessFn := func(a, b *jobsched.MetaJob[int, jobsched.NoProperty]) bool {
		pa, pb := int(a.Meta.Priority)-100, int(b.Meta.Priority)-100
		if pa < 0 {
			pa = -pa
		}
		if pb < 0 {
			pb = -pb
		}
		if pa == pb {
			return a.Meta.CreationTime.Before(b.Meta.CreationTime)
		}
		return pa < pb
	}
	testJobQueueFunc(
		t,
		queue.NewJobPriorityQueueMaker[int, jobsched.NoProperty](lessFn),
		makeWant(lessFn),
	)
}
