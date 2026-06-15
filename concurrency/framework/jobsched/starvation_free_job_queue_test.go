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
	"time"

	"github.com/donyori/gogo/concurrency/framework/jobsched"
)

func TestExponentialJobQueue_Basic(t *testing.T) {
	t.Parallel()

	testJobQueueFunc(
		t,
		jobsched.NewExponentialJobQueueMaker[int, jobsched.NoProperty](0, 0.),
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

func TestExponentialJobQueue_StarvationFree(t *testing.T) {
	t.Parallel()

	const (
		N          int  = 2
		TargetJob  int  = 0
		NumDequeue uint = 1500
	)

	baseTime := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	jq := jobsched.NewExponentialJobQueueMaker[
		int,
		jobsched.NoProperty,
	](N, 0.).New()
	jq.Enqueue(&jobsched.MetaJob[int, jobsched.NoProperty]{
		Meta: jobsched.Meta[jobsched.NoProperty]{
			Priority:     0,
			CreationTime: baseTime,
		},
		Job: TargetJob,
	})
	nextJob := TargetJob + 1

	for i := range NumDequeue {
		p := 3 + i
		ct := baseTime.Add(time.Duration(i) * time.Millisecond)

		for range N {
			jq.Enqueue(&jobsched.MetaJob[int, jobsched.NoProperty]{
				Meta: jobsched.Meta[jobsched.NoProperty]{
					Priority:     p,
					CreationTime: ct,
				},
				Job: nextJob,
			})
			nextJob++
		}

		job := jq.Dequeue()
		if job == TargetJob {
			return
		}
	}

	t.Errorf("target job cannot be dequeued in %d calls to Dequeue", NumDequeue)
}
