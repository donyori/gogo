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
	"sync/atomic"
	"testing"
	"time"

	"github.com/donyori/gogo/concurrency/framework"
)

func TestDefaultJobQueue(t *testing.T) {
	initJobs := make([]*Job, 10)
	now := time.Now()
	for i := range initJobs {
		initJobs[i] = &Job{
			Data: string('0' + rune(i)),
			Ct:   now.Add(time.Duration(i)),
		}
	}

	for n := 1; n <= 6; n++ {
		for i := 0; i < 10; i++ { // Repeat 10 times.
			var dqCntr uint32
			numDqs := make([]uint32, 10)
			prs := Run(n, func(jobData interface{}, quitDevice framework.QuitDevice) (newJobs []*Job) {
				numDq := atomic.AddUint32(&dqCntr, 1)
				data := jobData.(string)
				if len(data) >= 3 {
					return
				}
				if len(data) == 1 {
					numDqs[data[0]-'0'] = numDq
				}
				pri := uint(len(data)) * 3
				now := time.Now()
				newJobs = make([]*Job, 10)
				for k := range newJobs {
					newJobs[k] = &Job{
						Data: data + string('0'+rune(k)),
						Pri:  pri,
						Ct:   now.Add(time.Duration(k)),
					}
				}
				return
			}, nil, initJobs...)
			if len(prs) > 0 {
				t.Errorf("Panic: %q.", prs)
			}

			lastTenth := dqCntr * 4 / 5
			for k, num := range numDqs {
				if num > lastTenth {
					t.Errorf("%d is %d-th dequeued, in the last tenth. (Total: %d jobs, %d workers)", k, num, dqCntr, n)
				}
			}
		}
	}
}

func _TestShowDefaultJobQueue(t *testing.T) {
	initJobs := make([]*Job, 10)
	now := time.Now()
	for i := range initJobs {
		initJobs[i] = &Job{
			Data: string('0' + rune(i)),
			Ct:   now.Add(time.Duration(i)),
		}
	}

	prs := Run(4, func(jobData interface{}, quitDevice framework.QuitDevice) (newJobs []*Job) {
		data := jobData.(string)
		if len(data) >= 3 {
			return
		}
		pri := uint(len(data)) * 3
		now := time.Now()
		newJobs = make([]*Job, 10)
		for i := range newJobs {
			newJobs[i] = &Job{
				Data: data + string('0'+rune(i)),
				Pri:  pri,
				Ct:   now.Add(time.Duration(i)),
			}
		}
		return
	}, func(n int) JobQueue {
		return newTestJobQueueMonitor(DefaultJobQueueMaker(n), t)
	}, initJobs...)
	if len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
}
