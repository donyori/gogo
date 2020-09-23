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

import "testing"

type testJobQueueMonitor struct {
	jq           JobQueue
	tb           testing.TB
	numEq, numDq int64
}

func (tjq *testJobQueueMonitor) Len() int {
	return tjq.jq.Len()
}

func (tjq *testJobQueueMonitor) Enqueue(jobs ...*Job) {
	tjq.jq.Enqueue(jobs...)
	jobDataList := make([]interface{}, len(jobs))
	for i := range jobs {
		jobDataList[i] = jobs[i].Data
	}
	tjq.tb.Logf("++ Enqueue %d: %v.", tjq.numEq, jobDataList)
	tjq.numEq++
}

func (tjq *testJobQueueMonitor) Dequeue() interface{} {
	jobData := tjq.jq.Dequeue()
	tjq.tb.Logf("-- Dequeue %d: %v.", tjq.numDq, jobData)
	tjq.numDq++
	return jobData
}

type testJobQueueMonitorMaker struct {
	Maker JobQueueMaker
	Tb    testing.TB
}

func (m *testJobQueueMonitorMaker) New() JobQueue {
	return &testJobQueueMonitor{
		jq: m.Maker.New(),
		tb: m.Tb,
	}
}
