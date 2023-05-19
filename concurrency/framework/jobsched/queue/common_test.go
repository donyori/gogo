// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2023  Yuan Gao
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
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/donyori/gogo/concurrency/framework/jobsched"
)

// The number of jobs in metaJobs.
const N int = 600

var metaJobs []*jobsched.MetaJob[int, jobsched.NoProperty] // it is set in function init

var enqueueFnTestCases = []struct {
	name      string
	enqueueFn func(jq jobsched.JobQueue[int, jobsched.NoProperty])
}{
	{"enqueue one by one", func(jq jobsched.JobQueue[int, jobsched.NoProperty]) {
		for _, mj := range metaJobs {
			jq.Enqueue(mj)
		}
	}},
	{"enqueue one time", func(jq jobsched.JobQueue[int, jobsched.NoProperty]) {
		jq.Enqueue(metaJobs...)
	}},
	{"enqueue 3 parts", func(jq jobsched.JobQueue[int, jobsched.NoProperty]) {
		i, j := N/3, N*2/3
		jq.Enqueue(metaJobs[:i]...)
		jq.Enqueue(metaJobs[i:j]...)
		jq.Enqueue(metaJobs[j:]...)
	}},
	{"enqueue half then one by one", func(jq jobsched.JobQueue[int, jobsched.NoProperty]) {
		i := N / 2
		jq.Enqueue(metaJobs[:i]...)
		for _, mj := range metaJobs[i:] {
			jq.Enqueue(mj)
		}
	}},
}

func init() {
	priorities := make([]uint, N)
	for i := range priorities {
		priorities[i] = uint(i) >> 2
	}
	times := make([]time.Time, N)
	for i := range times {
		times[i] = time.Date(2000, time.January, 1, 0, 0, i%60, 0, time.UTC)
	}
	random := rand.New(rand.NewSource(10))
	random.Shuffle(N, func(i, j int) {
		priorities[i], priorities[j] = priorities[j], priorities[i]
	})
	random.Shuffle(N, func(i, j int) {
		times[i], times[j] = times[j], times[i]
	})
	metaJobs = make([]*jobsched.MetaJob[int, jobsched.NoProperty], N)
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

// makeWant uses metaJobs as the input jobs to generate the argument for
// the parameter want of the function testJobQueueFunc.
//
// lessFn indicates whether job a must be dequeued before job b.
//
// Its return value is a sequence of groups of jobs.
// The less the index of the group, the earlier dequeued.
// The jobs in the same group can be dequeued in any order.
func makeWant(lessFn func(a, b *jobsched.MetaJob[int, jobsched.NoProperty]) bool) [][]int {
	mjs := make([]*jobsched.MetaJob[int, jobsched.NoProperty], N)
	copy(mjs, metaJobs)
	less := func(i, j int) bool {
		return lessFn(mjs[i], mjs[j])
	}
	sort.Slice(mjs, less)
	want := make([][]int, 0, N)
	for i, mj := range mjs {
		if i > 0 && !less(i-1, i) {
			want[len(want)-1] = append(want[len(want)-1], mj.Job)
		} else {
			want = append(want, []int{mj.Job})
		}
	}
	return want
}

// testJobQueueFunc uses enqueueFnTestCases to test the job queue maker m.
//
// want is a sequence of groups of jobs.
// The less the index of the group, the earlier dequeued.
// The jobs in the same group can be dequeued in any order.
func testJobQueueFunc(t *testing.T, m jobsched.JobQueueMaker[int, jobsched.NoProperty], want [][]int) {
	var wantN int
	for _, group := range want {
		wantN += len(group)
	}
	if wantN != N {
		t.Errorf("warning: wantN (%d) != N (%d)", wantN, N)
	}
	for _, tc := range enqueueFnTestCases {
		t.Run(tc.name, func(t *testing.T) {
			jq := m.New()
			tc.enqueueFn(jq)
			got := make([]int, 0, N)
			for jq.Len() > 0 {
				got = append(got, jq.Dequeue())
			}
			if len(got) != wantN || gotWrong(got, want) {
				t.Errorf("got (len=%d) %v;\nwant (len=%d) %v", len(got), got, wantN, want)
				if len(got) != wantN {
					return
				}
			}
			defer func() {
				if e := recover(); !isDequeuePanicMessage(e) {
					t.Error(e)
				}
			}()
			job := jq.Dequeue() // want panic here
			t.Errorf("dequeued more than %d items, got %d", wantN, job)
		})
	}
}

// gotWrong reports whether got violates want.
//
// want is a sequence of groups of jobs.
// The less the index of the group, the earlier dequeued.
// The jobs in the same group can be dequeued in any order.
func gotWrong(got []int, want [][]int) bool {
	var gotIdx int
	for _, group := range want {
		if gotIdx >= len(got) {
			return true
		}
		switch len(group) {
		case 0:
			// Do nothing here.
		case 1:
			if got[gotIdx] != group[0] {
				return true
			}
			gotIdx++
		default:
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

func isDequeuePanicMessage(err any) bool {
	msg, ok := err.(string)
	return ok && strings.HasSuffix(msg, "job queue is empty")
}
