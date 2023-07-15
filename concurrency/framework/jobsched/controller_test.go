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
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/donyori/gogo/concurrency/framework"
	"github.com/donyori/gogo/concurrency/framework/jobsched"
)

func TestNew_FeedbackChanBufSize(t *testing.T) {
	const NumJob = 16
	for bufSize := -2; bufSize <= NumJob+1; bufSize++ {
		t.Run(fmt.Sprintf("bufSize=%d", bufSize), func(t *testing.T) {
			var x atomic.Int32
			ctrl := jobsched.New(func(quitDevice framework.QuitDevice, rank, job int) (
				newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
				x.Add(1)
				return nil, 1
			}, &jobsched.Options[int, jobsched.NoProperty, int]{
				NumWorker:           4,
				FeedbackChanBufSize: bufSize,
			}, make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...)
			feedbackRecvDoneC := make(chan struct{})
			defer func() {
				<-feedbackRecvDoneC
			}()
			go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, NumJob)
			ctrl.Run()
			if gotX := x.Load(); gotX != NumJob {
				t.Errorf("got x %d; want %d", gotX, NumJob)
			}
			if prs := ctrl.PanicRecords(); len(prs) > 0 {
				t.Errorf("panic %q", prs)
			}
		})
	}
}

func TestRun_NilFeedbackHandler(t *testing.T) {
	const NumWorker = 3
	const NumJob = NumWorker * 2
	var x atomic.Int32
	prs := jobsched.Run(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		x.Add(1)
		return nil, 1
	}, nil, &jobsched.Options[int, jobsched.NoProperty, int]{
		NumWorker: NumWorker,
	}, make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...)
	if gotX := x.Load(); gotX != NumJob {
		t.Errorf("got x %d; want %d", gotX, NumJob)
	}
	if len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestRun_FeedbackHandlerPanic(t *testing.T) {
	const PanicMsg = "test panic"
	const NumWorker = 3
	prs := jobsched.Run(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		return // do nothing
	}, func(feedbackChan <-chan int) {
		panic(PanicMsg)
	}, &jobsched.Options[int, jobsched.NoProperty, int]{
		NumWorker: NumWorker,
	}, make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker*2)...)
	switch {
	case len(prs) != 1:
		t.Errorf("got len(prs) %d; want 1", len(prs))
	case prs[0].Name != "feedback handler":
		t.Error(prs[0])
	default:
		msg, ok := prs[0].Content.(string)
		if !ok || msg != PanicMsg {
			t.Error(prs[0])
		}
	}
}

func TestRun_Panic(t *testing.T) {
	const PanicMsg = "test panic"
	const NumWorker = 3
	var x atomic.Int32
	prs := jobsched.Run(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		// Job handler should not be called
		// since the feedback handler is called before it
		// and the feedback handler panics.
		x.Add(1)
		panic(PanicMsg)
	}, func(feedbackChan <-chan int) {
		panic(PanicMsg)
	}, &jobsched.Options[int, jobsched.NoProperty, int]{
		NumWorker: NumWorker,
	}, make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker*2)...)
	if gotX := x.Load(); gotX != 0 {
		t.Errorf("got x %d; want 0", gotX)
	}
	switch {
	case len(prs) != 1:
		t.Errorf("got len(prs) %d; want 1", len(prs))
	case prs[0].Name != "feedback handler":
		t.Error(prs[0])
	default:
		msg, ok := prs[0].Content.(string)
		if !ok || msg != PanicMsg {
			t.Error(prs[0])
		}
	}
}

func TestRunWithoutFeedback_Panic(t *testing.T) {
	const PanicMsg = "test panic"
	const NumWorker = 3
	var x atomic.Int32
	var wg sync.WaitGroup
	wg.Add(NumWorker)
	prs := jobsched.RunWithoutFeedback(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback jobsched.NoFeedback) {
		x.Add(1)
		wg.Done()
		wg.Wait() // block the worker to ensure that each worker is ready to panic
		panic(PanicMsg)
	}, &jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
		NumWorker: NumWorker,
	}, make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker)...)
	if gotX := x.Load(); gotX != NumWorker {
		t.Errorf("got x %d; want %d", gotX, NumWorker)
	}
	if len(prs) != NumWorker {
		t.Errorf("got len(prs) %d; want %d", len(prs), NumWorker)
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
	ctrl := jobsched.New(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		return // do nothing
	}, nil)
	if got := ctrl.Wait(); got != -1 {
		t.Errorf("got %d; want -1", got)
	}
}

func TestController_Input_BeforeLaunch(t *testing.T) {
	const NumJob = 3
	var x atomic.Int32
	ctrl := jobsched.New(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		x.Add(1)
		return nil, 1
	}, nil)
	feedbackRecvDoneC := make(chan struct{})
	defer func() {
		<-feedbackRecvDoneC
	}()
	go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, NumJob)
	gotInput := ctrl.Input(make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...)
	if gotInput != NumJob {
		t.Fatalf("before calling Launch, got %d; want %d", gotInput, NumJob)
	}
	ctrl.Run()
	if gotX := x.Load(); gotX != NumJob {
		t.Errorf("got x %d; want %d", gotX, NumJob)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_BeforeLaunch_Concurrency(t *testing.T) {
	const MaxJob = 8
	const NumJobPerInput = 10
	const WantX = NumJobPerInput * MaxJob * (1 + MaxJob) / 2
	const WantFeedbackSum = NumJobPerInput * MaxJob
	var x atomic.Int32
	ctrl := jobsched.New(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		x.Add(int32(job))
		return nil, 1
	}, &jobsched.Options[int, jobsched.NoProperty, int]{
		NumWorker: 4,
	})
	feedbackRecvDoneC := make(chan struct{})
	defer func() {
		<-feedbackRecvDoneC
	}()
	go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, WantFeedbackSum)
	var wg sync.WaitGroup
	wg.Add(MaxJob)
	for job := 1; job <= MaxJob; job++ {
		go func(job int) {
			defer wg.Done()
			mjs := make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJobPerInput)
			for i := range mjs {
				mjs[i] = &jobsched.MetaJob[int, jobsched.NoProperty]{Job: job}
			}
			gotInput := ctrl.Input(mjs...)
			if gotInput != NumJobPerInput {
				t.Errorf("job %d - got %d; want %d",
					job, gotInput, NumJobPerInput)
			}
		}(job)
	}
	wg.Wait()
	ctrl.Run()
	if gotX := x.Load(); gotX != WantX {
		t.Errorf("got x %d; want %d", gotX, WantX)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_DuringLaunch(t *testing.T) {
	const NumJob = 3
	var x atomic.Int32
	ctrl := jobsched.New(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		x.Add(1)
		return nil, 1
	}, nil)
	feedbackRecvDoneC := make(chan struct{})
	defer func() {
		<-feedbackRecvDoneC
	}()
	go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, NumJob)
	lnchDoneC := make(chan struct{})
	go func() {
		ctrl.Launch()
		close(lnchDoneC)
	}()
	gotInput := ctrl.Input(make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...)
	if gotInput != NumJob {
		t.Fatalf("during calling Launch, got %d; want %d", gotInput, NumJob)
	}
	<-lnchDoneC
	ctrl.Wait()
	if gotX := x.Load(); gotX != NumJob {
		t.Errorf("got x %d; want %d", gotX, NumJob)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_DuringLaunch_Concurrency(t *testing.T) {
	const MaxJob = 8
	const NumJobPerInput = 10
	const WantX = NumJobPerInput * MaxJob * (1 + MaxJob) / 2
	const WantFeedbackSum = NumJobPerInput * MaxJob
	var x atomic.Int32
	ctrl := jobsched.New(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		x.Add(int32(job))
		return nil, 1
	}, &jobsched.Options[int, jobsched.NoProperty, int]{
		NumWorker: 4,
	})
	feedbackRecvDoneC := make(chan struct{})
	defer func() {
		<-feedbackRecvDoneC
	}()
	go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, WantFeedbackSum)
	lnchDoneC := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(MaxJob)
	go func() {
		ctrl.Launch()
		close(lnchDoneC)
	}()
	for job := 1; job <= MaxJob; job++ {
		go func(job int) {
			defer wg.Done()
			mjs := make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJobPerInput)
			for i := range mjs {
				mjs[i] = &jobsched.MetaJob[int, jobsched.NoProperty]{Job: job}
			}
			gotInput := ctrl.Input(mjs...)
			if gotInput != NumJobPerInput {
				t.Errorf("job %d - got %d; want %d",
					job, gotInput, NumJobPerInput)
			}
		}(job)
	}
	<-lnchDoneC
	wg.Wait()
	ctrl.Wait()
	if gotX := x.Load(); gotX != WantX {
		t.Errorf("got x %d; want %d", gotX, WantX)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_AfterLaunch(t *testing.T) {
	const NumJob = 3
	var x atomic.Int32
	ctrl := jobsched.New(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		x.Add(1)
		return nil, 1
	}, nil)
	feedbackRecvDoneC := make(chan struct{})
	defer func() {
		<-feedbackRecvDoneC
	}()
	go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, NumJob)
	ctrl.Launch()
	gotInput := ctrl.Input(make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...)
	if gotInput != NumJob {
		t.Fatalf("after calling Launch, got %d; want %d", gotInput, NumJob)
	}
	ctrl.Wait()
	if gotX := x.Load(); gotX != NumJob {
		t.Errorf("got x %d; want %d", gotX, NumJob)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_AfterLaunch_Concurrency(t *testing.T) {
	const MaxJob = 8
	const NumJobPerInput = 10
	const WantX = NumJobPerInput * MaxJob * (1 + MaxJob) / 2
	const WantFeedbackSum = NumJobPerInput * MaxJob
	var x atomic.Int32
	ctrl := jobsched.New(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		x.Add(int32(job))
		return nil, 1
	}, &jobsched.Options[int, jobsched.NoProperty, int]{
		NumWorker: 4,
	})
	feedbackRecvDoneC := make(chan struct{})
	defer func() {
		<-feedbackRecvDoneC
	}()
	go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, WantFeedbackSum)
	ctrl.Launch()
	var wg sync.WaitGroup
	wg.Add(MaxJob)
	for job := 1; job <= MaxJob; job++ {
		go func(job int) {
			defer wg.Done()
			mjs := make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJobPerInput)
			for i := range mjs {
				mjs[i] = &jobsched.MetaJob[int, jobsched.NoProperty]{Job: job}
			}
			gotInput := ctrl.Input(mjs...)
			if gotInput != NumJobPerInput {
				t.Errorf("job %d - got %d; want %d",
					job, gotInput, NumJobPerInput)
			}
		}(job)
	}
	wg.Wait()
	ctrl.Wait()
	if gotX := x.Load(); gotX != WantX {
		t.Errorf("got x %d; want %d", gotX, WantX)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_DuringWaiting(t *testing.T) {
	var x atomic.Int32
	handlerPauseC := make(chan struct{})
	ctrl := jobsched.New(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		x.Add(1)
		<-handlerPauseC
		return nil, 1
	}, nil)
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
	time.Sleep(time.Millisecond) // sleep to wait for starting ctrl.Wait
	gotInput := ctrl.Input(make([]*jobsched.MetaJob[int, jobsched.NoProperty], 3)...)
	if gotInput != 0 {
		t.Fatalf("during waiting, got %d; want 0", gotInput)
	}
	close(handlerPauseC)
	if gotX := x.Load(); gotX != 0 {
		t.Errorf("got x %d; want 0", gotX)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_AfterWait(t *testing.T) {
	var x atomic.Int32
	ctrl := jobsched.New(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		x.Add(1)
		return nil, 1
	}, nil)
	feedbackRecvDoneC := make(chan struct{})
	defer func() {
		<-feedbackRecvDoneC
	}()
	go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, 0)
	ctrl.Run()
	gotInput := ctrl.Input(make([]*jobsched.MetaJob[int, jobsched.NoProperty], 3)...)
	if gotInput != 0 {
		t.Fatalf("after calling Wait, got %d; want 0", gotInput)
	}
	if gotX := x.Load(); gotX != 0 {
		t.Errorf("got x %d; want 0", gotX)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_AfterIneffectiveWait(t *testing.T) {
	const NumJob = 3
	var x atomic.Int32
	ctrl := jobsched.New(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback int) {
		x.Add(1)
		return nil, 1
	}, nil)
	feedbackRecvDoneC := make(chan struct{})
	defer func() {
		<-feedbackRecvDoneC
	}()
	go receiveIntFeedbackAndTestSum(t, ctrl, feedbackRecvDoneC, NumJob)
	if gotWait := ctrl.Wait(); gotWait != -1 {
		t.Errorf("got %d on ineffective call to Wait; want -1", gotWait)
	}
	gotInput := ctrl.Input(make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...)
	if gotInput != NumJob {
		t.Errorf("after calling Wait ineffectively, got %d; want %d",
			gotInput, NumJob)
	}
	ctrl.Run()
	if gotX := x.Load(); gotX != NumJob {
		t.Errorf("got x %d; want %d", gotX, NumJob)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_NoFeedback(t *testing.T) {
	const NumJob = 6
	var x atomic.Int32
	ctrl := jobsched.New(
		func(quitDevice framework.QuitDevice, rank, job int) (
			newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback jobsched.NoFeedback) {
			x.Add(1)
			return
		},
		&jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
			NumWorker: 2,
		},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)..., // 2 workers and 6 jobs to test blocking
	)
	if gotFC := ctrl.FeedbackChan(); gotFC != nil {
		t.Error("got non-nil feedback channel; want nil")
	}
	ctrl.Run()
	if gotX := x.Load(); gotX != NumJob {
		t.Errorf("got x %d; want %d", gotX, NumJob)
	}
	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Setup(t *testing.T) {
	const NumWorker = 3
	var setupCounter [NumWorker]atomic.Int32
	prs := jobsched.RunWithoutFeedback(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback jobsched.NoFeedback) {
		return // do nothing
	}, &jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
		NumWorker: NumWorker,
		Setup: func(ctrl jobsched.Controller[int, jobsched.NoProperty, jobsched.NoFeedback], rank int) {
			setupCounter[rank].Add(1)
		},
	}, make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker*2)...)
	if len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
	for rank := range setupCounter {
		if gotCtr := setupCounter[rank].Load(); gotCtr != 1 {
			t.Errorf("got setupCounter[%d] %d; want 1", rank, gotCtr)
		}
	}
}

func TestController_Setup_WorkerPanic(t *testing.T) {
	const PanicMsg = "test panic"
	const NumWorker = 3
	var setupCounter [NumWorker]atomic.Int32
	var wg sync.WaitGroup
	wg.Add(NumWorker)
	prs := jobsched.RunWithoutFeedback(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback jobsched.NoFeedback) {
		wg.Done()
		wg.Wait() // block the worker to ensure that each worker is ready to panic
		panic(PanicMsg)
	}, &jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
		NumWorker: NumWorker,
		Setup: func(ctrl jobsched.Controller[int, jobsched.NoProperty, jobsched.NoFeedback], rank int) {
			setupCounter[rank].Add(1)
		},
	}, make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker*2)...)
	for rank := range setupCounter {
		if gotCtr := setupCounter[rank].Load(); gotCtr != 1 {
			t.Errorf("got setupCounter[%d] %d; want 1", rank, gotCtr)
		}
	}
	if len(prs) != NumWorker {
		t.Errorf("got len(prs) %d; want %d", len(prs), NumWorker)
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

func TestController_Cleanup(t *testing.T) {
	const NumWorker = 3
	var cleanupCounter [NumWorker]atomic.Int32
	prs := jobsched.RunWithoutFeedback(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback jobsched.NoFeedback) {
		return // do nothing
	}, &jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
		NumWorker: NumWorker,
		Cleanup: func(ctrl jobsched.Controller[int, jobsched.NoProperty, jobsched.NoFeedback], rank int) {
			cleanupCounter[rank].Add(1)
		},
	}, make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker*2)...)
	if len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
	for rank := range cleanupCounter {
		if gotCtr := cleanupCounter[rank].Load(); gotCtr != 1 {
			t.Errorf("got cleanupCounter[%d] %d; want 1", rank, gotCtr)
		}
	}
}

func TestController_Cleanup_WorkerPanic(t *testing.T) {
	const PanicMsg = "test panic"
	const NumWorker = 3
	var cleanupCounter [NumWorker]atomic.Int32
	var wg sync.WaitGroup
	wg.Add(NumWorker)
	prs := jobsched.RunWithoutFeedback(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback jobsched.NoFeedback) {
		wg.Done()
		wg.Wait() // block the worker to ensure that each worker is ready to panic
		panic(PanicMsg)
	}, &jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
		NumWorker: NumWorker,
		Cleanup: func(ctrl jobsched.Controller[int, jobsched.NoProperty, jobsched.NoFeedback], rank int) {
			cleanupCounter[rank].Add(1)
		},
	}, make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker*2)...)
	for rank := range cleanupCounter {
		if gotCtr := cleanupCounter[rank].Load(); gotCtr != 1 {
			t.Errorf("got cleanupCounter[%d] %d; want 1", rank, gotCtr)
		}
	}
	if len(prs) != NumWorker {
		t.Errorf("got len(prs) %d; want %d", len(prs), NumWorker)
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

func TestController_Cleanup_SetupPanic(t *testing.T) {
	const PanicMsg = "test panic"
	const NumWorker = 3
	var cleanupCounter [NumWorker]atomic.Int32
	var wg sync.WaitGroup
	wg.Add(NumWorker)
	prs := jobsched.RunWithoutFeedback(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback jobsched.NoFeedback) {
		return // do nothing
	}, &jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
		NumWorker: NumWorker,
		Setup: func(ctrl jobsched.Controller[int, jobsched.NoProperty, jobsched.NoFeedback], rank int) {
			wg.Done()
			wg.Wait() // block the worker to ensure that each worker is ready to panic
			panic(PanicMsg)
		},
		Cleanup: func(ctrl jobsched.Controller[int, jobsched.NoProperty, jobsched.NoFeedback], rank int) {
			cleanupCounter[rank].Add(1)
		},
	}, make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker*2)...)
	for rank := range cleanupCounter {
		if gotCtr := cleanupCounter[rank].Load(); gotCtr != 0 {
			t.Errorf("got cleanupCounter[%d] %d; want 0", rank, gotCtr)
		}
	}
	if len(prs) != NumWorker {
		t.Errorf("got len(prs) %d; want %d", len(prs), NumWorker)
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

func TestController_SetupAndCleanup(t *testing.T) {
	const NumWorker = 3
	var setupCounter, cleanupCounter [NumWorker]atomic.Int32
	prs := jobsched.RunWithoutFeedback(func(quitDevice framework.QuitDevice, rank, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback jobsched.NoFeedback) {
		return // do nothing
	}, &jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
		NumWorker: NumWorker,
		Setup: func(ctrl jobsched.Controller[int, jobsched.NoProperty, jobsched.NoFeedback], rank int) {
			setupCounter[rank].Add(1)
		},
		Cleanup: func(ctrl jobsched.Controller[int, jobsched.NoProperty, jobsched.NoFeedback], rank int) {
			cleanupCounter[rank].Add(1)
		},
	}, make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker*2)...)
	if len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
	for rank := 0; rank < NumWorker; rank++ {
		if gotCtr := setupCounter[rank].Load(); gotCtr != 1 {
			t.Errorf("got setupCounter[%d] %d; want 1", rank, gotCtr)
		}
		if gotCtr := cleanupCounter[rank].Load(); gotCtr != 1 {
			t.Errorf("got cleanupCounter[%d] %d; want 1", rank, gotCtr)
		}
	}
}

func TestController_JobHandlerRankUnique(t *testing.T) {
	for n := -1; n <= 100; n++ {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			numWorker := n
			if numWorker <= 0 {
				numWorker = runtime.NumCPU()
			}
			counter := make([]atomic.Int32, numWorker)
			var wg sync.WaitGroup
			wg.Add(numWorker)
			prs := jobsched.RunWithoutFeedback(func(quitDevice framework.QuitDevice, rank, job int) (
				newJobs []*jobsched.MetaJob[int, jobsched.NoProperty], feedback jobsched.NoFeedback) {
				counter[rank].Add(1)
				wg.Done()
				wg.Wait() // block the worker to ensure that each worker processes exactly one job
				return
			}, &jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
				NumWorker: n,
			}, make([]*jobsched.MetaJob[int, jobsched.NoProperty], numWorker)...)
			if len(prs) > 0 {
				t.Errorf("panic %q", prs)
			}
			for rank := range counter {
				if gotCtr := counter[rank].Load(); gotCtr != 1 {
					t.Errorf("got counter[%d] %d; want 1", rank, gotCtr)
				}
			}
		})
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
	fc := ctrl.FeedbackChan()
	if fc == nil {
		t.Error("got nil feedback channel")
		return
	}
	var sum int
	for feedback := range fc {
		sum += feedback
	}
	if sum != wantSum {
		t.Errorf("got sum of feedback %d; want %d", sum, wantSum)
	}
}
