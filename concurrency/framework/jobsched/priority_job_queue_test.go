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

package jobsched_test

import (
	"testing"

	"github.com/donyori/gogo/concurrency/framework/jobsched"
)

func TestPriorityFirstJobQueue(t *testing.T) {
	t.Parallel()

	testJobQueueFunc(
		t,
		jobsched.NewPriorityFirstJobQueueMaker[int, jobsched.NoProperty](),
		makeWantForTestJobQueueFunc(
			func(a, b *jobsched.MetaJob[int, jobsched.NoProperty]) int {
				switch {
				// Sort by priority in descending order:
				case a.Meta.Priority != b.Meta.Priority:
					if a.Meta.Priority < b.Meta.Priority {
						return 1
					}
				// Sort by creation time in ascending order:
				case !a.Meta.CreationTime.Equal(b.Meta.CreationTime):
					if a.Meta.CreationTime.After(b.Meta.CreationTime) {
						return 1
					}
				default:
					return 0
				}

				return -1
			},
		),
	)
}

func TestCreationTimeFirstJobQueue(t *testing.T) {
	t.Parallel()

	testJobQueueFunc(
		t,
		jobsched.NewCreationTimeFirstJobQueueMaker[int, jobsched.NoProperty](),
		makeWantForTestJobQueueFunc(
			func(a, b *jobsched.MetaJob[int, jobsched.NoProperty]) int {
				switch {
				// Sort by creation time in ascending order:
				case !a.Meta.CreationTime.Equal(b.Meta.CreationTime):
					if a.Meta.CreationTime.After(b.Meta.CreationTime) {
						return 1
					}
				// Sort by priority in descending order:
				case a.Meta.Priority != b.Meta.Priority:
					if a.Meta.Priority < b.Meta.Priority {
						return 1
					}
				default:
					return 0
				}

				return -1
			},
		),
	)
}

func TestJobPriorityQueue(t *testing.T) {
	t.Parallel()

	// Define a special priority rule:
	// A job has a higher priority if its Meta.Priority is closer to 100.
	// For two jobs with the same |Meta.Priority-100|,
	// the earlier the Meta.CreationTime, the higher the priority.

	minus100Abs := func(priority uint) int {
		p := int(priority) - 100
		if p < 0 {
			p = -p
		}

		return p
	}

	testJobQueueFunc(
		t,
		jobsched.NewJobPriorityQueueMaker(
			func(a, b *jobsched.MetaJob[int, jobsched.NoProperty]) bool {
				pa := minus100Abs(a.Meta.Priority)
				pb := minus100Abs(b.Meta.Priority)

				if pa != pb {
					return pa < pb
				}

				return a.Meta.CreationTime.Before(b.Meta.CreationTime)
			},
		),
		makeWantForTestJobQueueFunc(
			func(a, b *jobsched.MetaJob[int, jobsched.NoProperty]) int {
				pa := minus100Abs(a.Meta.Priority)
				pb := minus100Abs(b.Meta.Priority)

				switch {
				// Sort by |Meta.Priority-100| in ascending order:
				case pa != pb:
					if pa > pb {
						return 1
					}
				// Sort by creation time in ascending order:
				case !a.Meta.CreationTime.Equal(b.Meta.CreationTime):
					if a.Meta.CreationTime.After(b.Meta.CreationTime) {
						return 1
					}
				default:
					return 0
				}

				return -1
			},
		),
	)
}
