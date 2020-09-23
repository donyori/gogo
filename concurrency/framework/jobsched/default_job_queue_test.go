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
	"testing"
	"time"

	"github.com/donyori/gogo/concurrency/framework"
)

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
	}, &testJobQueueMonitorMaker{
		Maker: &DefaultJobQueueMaker{4},
		Tb:    t,
	}, initJobs...)
	if len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
}
