// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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

func TestController_Wait_BeforeLaunch(t *testing.T) {
	ctrl := New(0, func(jobData interface{}, quitDevice framework.QuitDevice) (newJobs []*Job) {
		return // Do nothing.
	}, nil)
	if r := ctrl.Wait(); r != -1 {
		t.Errorf("ctrl.Wait returns %d (not -1) before calling Launch.", r)
	}
}

func TestController_Input_BeforeLaunch(t *testing.T) {
	var x int32
	ctrl := New(0, func(jobData interface{}, quitDevice framework.QuitDevice) (newJobs []*Job) {
		atomic.AddInt32(&x, 1)
		return
	}, nil)
	if !ctrl.Input(nil, nil, nil) {
		t.Fatal("Input returns false before calling Launch.")
	}
	ctrl.Run()
	if v := atomic.LoadInt32(&x); v != 3 {
		t.Errorf("x: %d != 3.", v)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
}

func TestController_Input_DuringLaunch(t *testing.T) {
	var x int32
	c := make(chan struct{})
	ctrl := New(0, func(jobData interface{}, quitDevice framework.QuitDevice) (newJobs []*Job) {
		atomic.AddInt32(&x, 1)
		return
	}, nil)
	go func() {
		ctrl.Launch()
		close(c)
	}()
	if !ctrl.Input(nil, nil, nil) {
		t.Fatal("Input returns false during calling Launch.")
	}
	<-c
	ctrl.Wait()
	if v := atomic.LoadInt32(&x); v != 3 {
		t.Errorf("x: %d != 3.", v)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
}

func TestController_Input_AfterLaunch(t *testing.T) {
	var x int32
	ctrl := New(0, func(jobData interface{}, quitDevice framework.QuitDevice) (newJobs []*Job) {
		atomic.AddInt32(&x, 1)
		return
	}, nil)
	ctrl.Launch()
	if !ctrl.Input(nil, nil, nil) {
		t.Fatal("Input returns false after calling Launch.")
	}
	ctrl.Wait()
	if v := atomic.LoadInt32(&x); v != 3 {
		t.Errorf("x: %d != 3.", v)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
}

func TestController_Input_DuringWait(t *testing.T) {
	var x int32
	c := make(chan struct{})
	ctrl := New(0, func(jobData interface{}, quitDevice framework.QuitDevice) (newJobs []*Job) {
		atomic.AddInt32(&x, 1)
		<-c
		return
	}, nil)
	ctrl.Launch()
	go ctrl.Wait()
	time.Sleep(time.Millisecond)
	if ctrl.Input(nil, nil, nil) {
		t.Fatal("Input returns true during waiting.")
	}
	close(c)
	if v := atomic.LoadInt32(&x); v != 0 {
		t.Errorf("x: %d != 0.", v)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
}

func TestController_Input_AfterWait(t *testing.T) {
	var x int32
	ctrl := New(0, func(jobData interface{}, quitDevice framework.QuitDevice) (newJobs []*Job) {
		atomic.AddInt32(&x, 1)
		return
	}, nil)
	ctrl.Run()
	if ctrl.Input(nil, nil, nil) {
		t.Fatal("Input returns true after Wait.")
	}
	if v := atomic.LoadInt32(&x); v != 0 {
		t.Errorf("x: %d != 0.", v)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
}

func TestController_Input_AfterIneffectiveWait(t *testing.T) {
	var x int32
	ctrl := New(0, func(jobData interface{}, quitDevice framework.QuitDevice) (newJobs []*Job) {
		atomic.AddInt32(&x, 1)
		return
	}, nil)
	ctrl.Wait()
	if !ctrl.Input(nil, nil, nil) {
		t.Fatal("Input returns false after calling Wait ineffectively.")
	}
	ctrl.Run()
	if v := atomic.LoadInt32(&x); v != 3 {
		t.Errorf("x: %d != 3.", v)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("Panic: %q.", prs)
	}
}
