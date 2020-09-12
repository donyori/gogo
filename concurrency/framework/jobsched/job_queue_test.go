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
	"strconv"
	"testing"
	"time"

	"github.com/donyori/gogo/concurrency/framework"
)

type testJobQueue struct {
	jq           JobQueue
	tb           testing.TB
	numEq, numDq int64
}

func (tjq *testJobQueue) Len() int {
	return tjq.jq.Len()
}

func (tjq *testJobQueue) Enqueue(jobs ...*Job) {
	tjq.jq.Enqueue(jobs...)
	jobDataList := make([]interface{}, len(jobs))
	for i := range jobs {
		jobDataList[i] = jobs[i].Data
	}
	tjq.tb.Logf("++ Enqueue %d: %v.", tjq.numEq, jobDataList)
	tjq.numEq++
}

func (tjq *testJobQueue) Dequeue() interface{} {
	jobData := tjq.jq.Dequeue()
	tjq.tb.Logf("-- Dequeue %d: %v.", tjq.numDq, jobData)
	tjq.numDq++
	return jobData
}

func _TestShowDefaultJobQueue(t *testing.T) {
	initJobs := make([]*Job, 10)
	now := time.Now()
	for i := range initJobs {
		initJobs[i] = &Job{
			Data: strconv.Itoa(i),
			Pri:  0,
			Ct:   now.Add(time.Duration(i)),
		}
	}
	prs := Run(4, func(jobData interface{}, quitDevice framework.QuitDevice) (newJobs []*Job) {
		data := jobData.(string)
		if len(data) >= 3 {
			return nil
		}
		newJobs = make([]*Job, 10)
		now := time.Now()
		for i := range newJobs {
			newJobs[i] = &Job{
				Data: data + strconv.Itoa(i),
				Pri:  uint(len(data)) * 3,
				Ct:   now.Add(time.Duration(i)),
			}
		}
		return
	}, func(n int) JobQueue {
		return &testJobQueue{
			jq: defaultJobQueueMaker(n),
			tb: t,
		}
	}, initJobs...)
	if len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
}
