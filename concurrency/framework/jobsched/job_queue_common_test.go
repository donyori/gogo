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
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/donyori/gogo/concurrency/framework/jobsched"
)

// numMetaJob is the number of jobs with their meta information
// in the global variable metaJobs.
const numMetaJob int = 600

// metaJobs is a list of jobs with their meta information
// for testing each job queue.
var metaJobs []*jobsched.MetaJob[int, jobsched.NoProperty] // it is set in function init

var enqueueFnTestCases = []struct {
	name      string
	enqueueFn func(jq jobsched.JobQueue[int, jobsched.NoProperty])
}{
	{
		"enqueue_one_by_one",
		func(jq jobsched.JobQueue[int, jobsched.NoProperty]) {
			for _, mj := range metaJobs {
				jq.Enqueue(mj)
			}
		},
	},
	{
		"enqueue_one_time",
		func(jq jobsched.JobQueue[int, jobsched.NoProperty]) {
			jq.Enqueue(metaJobs...)
		},
	},
	{
		"enqueue_three_parts",
		func(jq jobsched.JobQueue[int, jobsched.NoProperty]) {
			i := numMetaJob / 3
			j := i << 1
			jq.Enqueue(metaJobs[:i]...)
			jq.Enqueue(metaJobs[i:j]...)
			jq.Enqueue(metaJobs[j:]...)
		},
	},
	{
		"enqueue_a_half_then_one_by_one",
		func(jq jobsched.JobQueue[int, jobsched.NoProperty]) {
			i := numMetaJob >> 1
			jq.Enqueue(metaJobs[:i]...)

			for _, mj := range metaJobs[i:] {
				jq.Enqueue(mj)
			}
		},
	},
}

func init() {
	random := rand.New(rand.NewChaCha8( //gosec:disable G404 -- math/rand/v2 is reproducible
		[32]byte([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ123456")),
	))

	priorities := make([]uint, numMetaJob)
	for i := range priorities {
		priorities[i] = uint(i) >> 2
	}

	random.Shuffle(numMetaJob, func(i, j int) {
		priorities[i], priorities[j] = priorities[j], priorities[i]
	})

	times := make([]time.Time, numMetaJob)
	for i := range times {
		times[i] = time.Date(2000, time.January, 1, 0, 0, i%60, 0, time.UTC)
	}

	random.Shuffle(numMetaJob, func(i, j int) {
		times[i], times[j] = times[j], times[i]
	})

	metaJobs = make([]*jobsched.MetaJob[int, jobsched.NoProperty], numMetaJob)
	for i := range metaJobs {
		metaJobs[i] = &jobsched.MetaJob[int, jobsched.NoProperty]{
			Meta: jobsched.Meta[jobsched.NoProperty]{
				Priority:     priorities[i],
				CreationTime: times[i],
			},
			Job: i,
		}
	}
}

// makeWantForTestJobQueueFunc uses metaJobs as the input jobs to generate
// the argument for the parameter want of the function testJobQueueFunc.
//
// cmp compares two jobs for sorting. It returns
//
//	-1 if job a must be dequeued before job b,
//	 0 if job a and job b can be dequeued in any order,
//	+1 if job a must be dequeued after job b.
//
// Its return value is a sequence of groups of jobs.
// The less the index of the group, the earlier dequeued.
// The jobs in the same group can be dequeued in any order.
func makeWantForTestJobQueueFunc(
	cmp func(a, b *jobsched.MetaJob[int, jobsched.NoProperty]) int,
) [][]int {
	mjs := make([]*jobsched.MetaJob[int, jobsched.NoProperty], numMetaJob)
	copy(mjs, metaJobs)
	slices.SortFunc(mjs, cmp)

	want := make([][]int, 0, numMetaJob)

	for i := range mjs {
		if i > 0 && cmp(mjs[i-1], mjs[i]) == 0 { //gosec:disable G602 -- slice index i-1 is valid as i > 0
			want[len(want)-1] = append(want[len(want)-1], mjs[i].Job)
		} else {
			want = append(want, []int{mjs[i].Job})
		}
	}

	return want
}

// testJobQueueFunc uses enqueueFnTestCases to test the job queue maker m.
//
// want is a sequence of groups of jobs.
// The less the index of the group, the earlier dequeued.
// The jobs in the same group can be dequeued in any order.
// want can be obtained by the function gotWrongForTestJobQueueFunc.
func testJobQueueFunc( //nolint:thelper // this is the main body, not a helper
	t *testing.T,
	m jobsched.JobQueueMaker[int, jobsched.NoProperty],
	want [][]int,
) {
	var wantN int
	for _, group := range want {
		wantN += len(group)
	}

	if wantN != numMetaJob {
		t.Errorf("warning: wantN (%d) != numMetaJob (%d)", wantN, numMetaJob)
	}

	for _, tc := range enqueueFnTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jq := m.New()
			tc.enqueueFn(jq)

			got := make([]int, 0, numMetaJob)
			for jq.Len() > 0 {
				got = append(got, jq.Dequeue())
			}

			if len(got) != wantN || gotWrongForTestJobQueueFunc(got, want) {
				t.Errorf("got (len=%d) %v;\nwant (len=%d) %v",
					len(got), got, wantN, want)

				if len(got) != wantN {
					return
				}
			}

			defer func() {
				if e := recover(); !isEmptyQueuePanicMessage(e) {
					t.Error(e)
				}
			}()

			job := jq.Dequeue() // want panic here
			t.Errorf("dequeued more than %d items, got %d", wantN, job)
		})
	}
}

// gotWrongForTestJobQueueFunc reports whether got violates want
// for the function testJobQueueFunc.
//
// want is a sequence of groups of jobs.
// The less the index of the group, the earlier dequeued.
// The jobs in the same group can be dequeued in any order.
func gotWrongForTestJobQueueFunc(got []int, want [][]int) bool {
	var gotIdx int

	for _, group := range want {
		switch {
		case gotIdx >= len(got):
			return true
		case len(group) == 1:
			if got[gotIdx] != group[0] {
				return true
			}

			gotIdx++
		case len(group) > 0:
			m := make(map[int]int, len(group))
			for _, x := range group {
				m[x]++
			}

			for range group {
				if gotIdx >= len(got) {
					return true
				}

				x := got[gotIdx]

				m[x]--
				if m[x] < 0 {
					return true
				}

				gotIdx++
			}
		}
	}

	return gotIdx != len(got)
}

func isEmptyQueuePanicMessage(panicValue any) bool {
	if panicValue == nil {
		return false
	}

	msg, ok := panicValue.(string)

	return ok && strings.HasSuffix(msg, jobsched.EmptyQueuePanicMessage)
}
