// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/donyori/gogo/concurrency/framework"
	"github.com/donyori/gogo/concurrency/framework/jobsched"
	"github.com/donyori/gogo/concurrency/framework/jobsched/queue"
)

func TestRun(t *testing.T) {
	const panicMsg string = "test panic"
	var x int32
	var wg sync.WaitGroup
	wg.Add(3)
	prs := jobsched.Run[int, jobsched.NoProperty](3, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty]) {
		atomic.AddInt32(&x, 1)
		wg.Done()
		wg.Wait()
		panic(panicMsg)
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty](), nil, nil, nil)
	if got := atomic.LoadInt32(&x); got != 3 {
		t.Errorf("got %d; want 3", got)
	}
	if len(prs) != 3 {
		t.Errorf("got len(prs) %d; want 3", len(prs))
	}
	for _, pr := range prs {
		if !strings.HasPrefix(pr.Name, "worker ") {
			t.Error(pr)
		} else {
			msg, ok := pr.Content.(string)
			if !ok || msg != panicMsg {
				t.Error(pr)
			}
		}
	}
}

func TestController_Wait_BeforeLaunch(t *testing.T) {
	ctrl := jobsched.New[int, jobsched.NoProperty](0, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty]) {
		return // do nothing
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty]())
	if got := ctrl.Wait(); got != -1 {
		t.Errorf("got %d; want -1", got)
	}
}

func TestController_Input_BeforeLaunch(t *testing.T) {
	var x int32
	ctrl := jobsched.New[int, jobsched.NoProperty](0, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty]) {
		atomic.AddInt32(&x, 1)
		return
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty]())
	if got := ctrl.Input(nil, nil, nil); got != 3 {
		t.Fatalf("before calling Launch, got %d; want 3", got)
	}
	ctrl.Run()
	if got := atomic.LoadInt32(&x); got != 3 {
		t.Errorf("got x %d; want 3", got)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_DuringLaunch(t *testing.T) {
	var x int32
	c := make(chan struct{})
	ctrl := jobsched.New[int, jobsched.NoProperty](0, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty]) {
		atomic.AddInt32(&x, 1)
		return
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty]())
	go func() {
		ctrl.Launch()
		close(c)
	}()
	if got := ctrl.Input(nil, nil, nil); got != 3 {
		t.Fatalf("during calling Launch, got %d; want 3", got)
	}
	<-c
	ctrl.Wait()
	if got := atomic.LoadInt32(&x); got != 3 {
		t.Errorf("got %d; want 3", got)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_AfterLaunch(t *testing.T) {
	var x int32
	ctrl := jobsched.New[int, jobsched.NoProperty](0, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty]) {
		atomic.AddInt32(&x, 1)
		return
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty]())
	ctrl.Launch()
	if got := ctrl.Input(nil, nil, nil); got != 3 {
		t.Fatalf("after calling Launch, got %d; want 3", got)
	}
	ctrl.Wait()
	if got := atomic.LoadInt32(&x); got != 3 {
		t.Errorf("got %d; want 3", got)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_DuringWaiting(t *testing.T) {
	var x int32
	c1, c2 := make(chan struct{}), make(chan struct{})
	ctrl := jobsched.New[int, jobsched.NoProperty](0, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty]) {
		atomic.AddInt32(&x, 1)
		<-c2
		return
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty]())
	ctrl.Launch()
	go func() {
		close(c1)
		ctrl.Wait()
	}()
	<-c1
	if got := ctrl.Input(nil, nil, nil); got != 0 {
		t.Fatalf("during waiting, got %d; want 0", got)
	}
	close(c2)
	if got := atomic.LoadInt32(&x); got != 0 {
		t.Errorf("got %d; want 0", got)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_AfterWait(t *testing.T) {
	var x int32
	ctrl := jobsched.New[int, jobsched.NoProperty](0, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty]) {
		atomic.AddInt32(&x, 1)
		return
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty]())
	ctrl.Run()
	if got := ctrl.Input(nil, nil, nil); got != 0 {
		t.Fatalf("after calling Wait, got %d; want 0", got)
	}
	if got := atomic.LoadInt32(&x); got != 0 {
		t.Errorf("got %d; want 0", got)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_AfterIneffectiveWait(t *testing.T) {
	var x int32
	ctrl := jobsched.New[int, jobsched.NoProperty](0, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty]) {
		atomic.AddInt32(&x, 1)
		return
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty]())
	if got := ctrl.Wait(); got != -1 {
		t.Errorf("got %d on ineffective call to Wait; want -1", got)
	}
	if got := ctrl.Input(nil, nil, nil); got != 3 {
		t.Errorf("after calling Wait ineffectively, got %d; want 3", got)
	}
	ctrl.Run()
	if got := atomic.LoadInt32(&x); got != 3 {
		t.Errorf("got %d; want 3", got)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}
