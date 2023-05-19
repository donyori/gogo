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
	const PanicMsg = "test panic"
	var x int32
	var wg sync.WaitGroup
	wg.Add(3)
	prs := jobsched.Run(3, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback jobsched.NoFeedback) {
		atomic.AddInt32(&x, 1)
		wg.Done()
		wg.Wait()
		panic(PanicMsg)
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty](), 0, nil, nil, nil)
	if gotX := atomic.LoadInt32(&x); gotX != 3 {
		t.Errorf("got x %d; want 3", gotX)
	}
	if len(prs) != 3 {
		t.Errorf("got len(prs) %d; want 3", len(prs))
	}
	for _, pr := range prs {
		if !strings.HasPrefix(pr.Name, "worker ") {
			t.Error(pr)
		} else {
			msg, ok := pr.Content.(string)
			if !ok || msg != PanicMsg {
				t.Error(pr)
			}
		}
	}
}

func TestController_Wait_BeforeLaunch(t *testing.T) {
	ctrl := jobsched.New(0, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		return // do nothing
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty](), 0)
	if got := ctrl.Wait(); got != -1 {
		t.Errorf("got %d; want -1", got)
	}
}

func TestController_Input_BeforeLaunch(t *testing.T) {
	var x int32
	ctrl := jobsched.New(0, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		atomic.AddInt32(&x, 1)
		return nil, 1
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty](), 0)
	feedbackRecvDoneC := make(chan struct{})
	defer func() {
		<-feedbackRecvDoneC
	}()
	go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, 3)
	if gotInput := ctrl.Input(nil, nil, nil); gotInput != 3 {
		t.Fatalf("before calling Launch, got %d; want 3", gotInput)
	}
	ctrl.Run()
	if gotX := atomic.LoadInt32(&x); gotX != 3 {
		t.Errorf("got x %d; want 3", gotX)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_DuringLaunch(t *testing.T) {
	var x int32
	ctrl := jobsched.New(0, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		atomic.AddInt32(&x, 1)
		return nil, 1
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty](), 0)
	feedbackRecvDoneC := make(chan struct{})
	defer func() {
		<-feedbackRecvDoneC
	}()
	go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, 3)
	lnchDoneC := make(chan struct{})
	go func() {
		ctrl.Launch()
		close(lnchDoneC)
	}()
	if gotInput := ctrl.Input(nil, nil, nil); gotInput != 3 {
		t.Fatalf("during calling Launch, got %d; want 3", gotInput)
	}
	<-lnchDoneC
	ctrl.Wait()
	if gotX := atomic.LoadInt32(&x); gotX != 3 {
		t.Errorf("got x %d; want 3", gotX)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_AfterLaunch(t *testing.T) {
	var x int32
	ctrl := jobsched.New(0, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		atomic.AddInt32(&x, 1)
		return nil, 1
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty](), 0)
	feedbackRecvDoneC := make(chan struct{})
	defer func() {
		<-feedbackRecvDoneC
	}()
	go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, 3)
	ctrl.Launch()
	if gotInput := ctrl.Input(nil, nil, nil); gotInput != 3 {
		t.Fatalf("after calling Launch, got %d; want 3", gotInput)
	}
	ctrl.Wait()
	if gotX := atomic.LoadInt32(&x); gotX != 3 {
		t.Errorf("got x %d; want 3", gotX)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_DuringWaiting(t *testing.T) {
	var x int32
	handlerPauseC := make(chan struct{})
	ctrl := jobsched.New(0, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		atomic.AddInt32(&x, 1)
		<-handlerPauseC
		return nil, 1
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty](), 0)
	feedbackRecvDoneC := make(chan struct{})
	defer func() {
		<-feedbackRecvDoneC
	}()
	go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, 0)
	ctrl.Launch()
	waitStartC := make(chan struct{})
	go func() {
		close(waitStartC)
		ctrl.Wait()
	}()
	<-waitStartC
	if gotInput := ctrl.Input(nil, nil, nil); gotInput != 0 {
		t.Fatalf("during waiting, got %d; want 0", gotInput)
	}
	close(handlerPauseC)
	if gotX := atomic.LoadInt32(&x); gotX != 0 {
		t.Errorf("got x %d; want 0", gotX)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_AfterWait(t *testing.T) {
	var x int32
	ctrl := jobsched.New(0, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		atomic.AddInt32(&x, 1)
		return nil, 1
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty](), 0)
	feedbackRecvDoneC := make(chan struct{})
	defer func() {
		<-feedbackRecvDoneC
	}()
	go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, 0)
	ctrl.Run()
	if gotInput := ctrl.Input(nil, nil, nil); gotInput != 0 {
		t.Fatalf("after calling Wait, got %d; want 0", gotInput)
	}
	if gotX := atomic.LoadInt32(&x); gotX != 0 {
		t.Errorf("got x %d; want 0", gotX)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_AfterIneffectiveWait(t *testing.T) {
	var x int32
	ctrl := jobsched.New(0, func(job int, quitDevice framework.QuitDevice) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		atomic.AddInt32(&x, 1)
		return nil, 1
	}, queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty](), 0)
	feedbackRecvDoneC := make(chan struct{})
	defer func() {
		<-feedbackRecvDoneC
	}()
	go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, 3)
	if gotWait := ctrl.Wait(); gotWait != -1 {
		t.Errorf("got %d on ineffective call to Wait; want -1", gotWait)
	}
	if gotInput := ctrl.Input(nil, nil, nil); gotInput != 3 {
		t.Errorf("after calling Wait ineffectively, got %d; want 3", gotInput)
	}
	ctrl.Run()
	if gotX := atomic.LoadInt32(&x); gotX != 3 {
		t.Errorf("got x %d; want 3", gotX)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_NoFeedback(t *testing.T) {
	var x int32
	ctrl := jobsched.New(
		2,
		func(job int, quitDevice framework.QuitDevice) (
			newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback jobsched.NoFeedback) {
			atomic.AddInt32(&x, 1)
			return
		},
		queue.NewFCFSJobQueueMaker[int, jobsched.NoProperty](),
		0,
		nil, nil, nil, nil, nil, nil, // 2 workers and 6 jobs to test blocking
	)
	if gotFbC := ctrl.FeedbackChan(); gotFbC != nil {
		t.Error("got non-nil feedback channel; want nil")
	}
	ctrl.Run()
	if gotX := atomic.LoadInt32(&x); gotX != 6 {
		t.Errorf("got x %d; want 6", gotX)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

// receiveIntFeedbackAndTestSum receives feedback (of type int)
// from ctrl.FeedbackChan(), calculates the sum,
// and compares the result with wantSum.
// If ctrl.FeedbackChan returns nil or the sum is not equal to wantSum,
// it reports the error message using t.Error and t.Errorf.
func receiveIntFeedbackAndTestSum[Job, Properties any](
	t *testing.T,
	ctrl jobsched.Controller[Job, Properties, int],
	doneChan chan<- struct{},
	wantSum int,
) {
	defer close(doneChan)
	fbc := ctrl.FeedbackChan()
	if fbc == nil {
		t.Error("got nil feedback channel")
		return
	}
	var sum int
	for feedback := range fbc {
		sum += feedback
	}
	if sum != wantSum {
		t.Errorf("got sum of feedback %d; want %d", sum, wantSum)
	}
}
