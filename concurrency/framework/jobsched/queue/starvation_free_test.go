// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2025  Yuan Gao
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
	"time"

	"github.com/donyori/gogo/concurrency/framework/jobsched"
	"github.com/donyori/gogo/concurrency/framework/jobsched/queue"
)

func TestExponentialJobQueue_Basic(t *testing.T) {
	testJobQueueFunc(
		t,
		queue.NewExponentialJobQueueMaker[int, jobsched.NoProperty](0, 0.),
		makeWant(func(a, b *jobsched.MetaJob[int, jobsched.NoProperty]) int {
			switch {
			case a.Meta.Priority > b.Meta.Priority:
				return -1
			case a.Meta.Priority < b.Meta.Priority:
				return 1
			case a.Meta.CreationTime.Before(b.Meta.CreationTime):
				return -1
			case a.Meta.CreationTime.After(b.Meta.CreationTime):
				return 1
			}
			return 0
		}),
	)
}

func TestExponentialJobQueue_StarvationFree(t *testing.T) {
	const N int = 2
	baseTime := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	var target int
	jq := queue.NewExponentialJobQueueMaker[int, jobsched.NoProperty](N, 0.).New()
	jq.Enqueue(&jobsched.MetaJob[int, jobsched.NoProperty]{
		Meta: jobsched.Meta[jobsched.NoProperty]{
			Priority:     0,
			CreationTime: baseTime,
		},
		Job: target,
	})

	const NumDequeue uint = 1500
	nextJob := target + 1
	for i := uint(0); i < NumDequeue; i++ {
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
		if job == 0 {
			return
		}
	}
	t.Errorf("target job cannot be dequeued in %d calls to Dequeue", NumDequeue)
}
