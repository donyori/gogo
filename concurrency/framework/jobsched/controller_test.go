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
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/donyori/gogo/concurrency"
	"github.com/donyori/gogo/concurrency/framework"
	"github.com/donyori/gogo/concurrency/framework/jobsched"
)

func TestNew_FeedbackChanBufSize(t *testing.T) {
	t.Parallel()

	const (
		NumWorker int = 4
		NumJob    int = 16
	)

	for bufSize := -2; bufSize <= NumJob+1; bufSize++ {
		t.Run(fmt.Sprintf("bufSize=%d", bufSize), func(t *testing.T) {
			t.Parallel()

			var (
				x           atomic.Int32
				feedbackSum int
			)

			ctrl := jobsched.New(
				getJobHandlerAddOne(&x),
				getSumFeedbackHandler(&feedbackSum),
				&jobsched.Options[int, jobsched.NoProperty, int]{
					NumWorker:           NumWorker,
					FeedbackChanBufSize: bufSize,
				},
				make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...,
			)
			ctrl.Run()
			checkXFeedbackSumAndPanicRecords(
				t,
				&x,
				feedbackSum,
				ctrl.PanicRecords(),
				int32(NumJob),
				NumJob,
			)
		})
	}
}

func TestRun_NilFeedbackHandler(t *testing.T) {
	t.Parallel()

	const (
		NumWorker int = 3
		NumJob    int = NumWorker << 1
	)

	var x atomic.Int32

	prs := jobsched.Run(
		getJobHandlerAddOne(&x),
		nil,
		&jobsched.Options[int, jobsched.NoProperty, int]{NumWorker: NumWorker},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...,
	)
	checkXAndPanicRecords(t, &x, prs, int32(NumJob))
}

func TestRun_FeedbackHandlerPanic(t *testing.T) {
	t.Parallel()

	const (
		PanicMsg      = "test panic"
		NumWorker int = 3
	)

	prs := jobsched.Run(
		getJobHandlerNoOp[int](),
		func(concurrency.Canceler, int) {
			panic(PanicMsg)
		},
		&jobsched.Options[int, jobsched.NoProperty, int]{NumWorker: NumWorker},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker<<1)...,
	)
	switch {
	case len(prs) != 1:
		t.Errorf("got len(prs) %d; want 1", len(prs))
	case prs[0].GoroutineName() != "feedback handler":
		t.Error(prs[0])
	default:
		msg, ok := prs[0].PanicValue().(string)
		if !ok || msg != PanicMsg {
			t.Error(prs[0])
		}
	}
}

func TestRun_Panic(t *testing.T) {
	t.Parallel()

	const (
		PanicMsg      = "test panic"
		NumWorker int = 3
	)

	var (
		x  atomic.Int32
		wg sync.WaitGroup
	)

	wg.Add(NumWorker)
	prs := jobsched.Run(
		func(concurrency.Canceler, int, int) (
			newJobs []*jobsched.MetaJob[int, jobsched.NoProperty],
			feedback int,
		) {
			x.Add(1)
			wg.Done()
			wg.Wait() // block the worker to ensure that each worker is ready to panic
			panic(PanicMsg)
		},
		func(concurrency.Canceler, int) {
			// Feedback handler should not be called
			// since the job handlers panic and return no feedback.
			panic(PanicMsg)
		},
		&jobsched.Options[int, jobsched.NoProperty, int]{NumWorker: NumWorker},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker<<1)...,
	)

	if gotX := x.Load(); gotX != int32(NumWorker) {
		t.Errorf("got x %d; want %d", gotX, NumWorker)
	}

	if len(prs) != NumWorker {
		t.Errorf("got len(prs) %d; want %d", len(prs), NumWorker)
	}

	for _, pr := range prs {
		if !strings.HasPrefix(pr.GoroutineName(), "worker ") {
			t.Error(pr)
		} else {
			msg, ok := pr.PanicValue().(string)
			if !ok || msg != PanicMsg {
				t.Error(pr)
			}
		}
	}
}

func TestRunWithoutFeedback_Panic(t *testing.T) {
	t.Parallel()

	const (
		PanicMsg      = "test panic"
		NumWorker int = 3
	)

	var (
		x  atomic.Int32
		wg sync.WaitGroup
	)

	wg.Add(NumWorker)
	prs := jobsched.RunWithoutFeedback(
		func(concurrency.Canceler, int, int) (
			newJobs []*jobsched.MetaJob[int, jobsched.NoProperty],
			feedback jobsched.NoFeedback,
		) {
			x.Add(1)
			wg.Done()
			wg.Wait() // block the worker to ensure that each worker is ready to panic
			panic(PanicMsg)
		},
		&jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
			NumWorker: NumWorker,
		},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker<<1)...,
	)

	if gotX := x.Load(); gotX != int32(NumWorker) {
		t.Errorf("got x %d; want %d", gotX, NumWorker)
	}

	if len(prs) != NumWorker {
		t.Errorf("got len(prs) %d; want %d", len(prs), NumWorker)
	}

	for _, pr := range prs {
		if !strings.HasPrefix(pr.GoroutineName(), "worker ") {
			t.Error(pr)
		} else {
			msg, ok := pr.PanicValue().(string)
			if !ok || msg != PanicMsg {
				t.Error(pr)
			}
		}
	}
}

func TestController_Wait_BeforeLaunch(t *testing.T) {
	t.Parallel()

	ctrl := jobsched.New(getJobHandlerNoOp[int](), nil, nil)

	got := ctrl.Wait()
	if got != -1 {
		t.Errorf("got %d; want -1", got)
	}
}

func TestController_NoFeedback(t *testing.T) {
	t.Parallel()

	// 2 workers and 6 jobs to test blocking.
	const (
		NumWorker int = 2
		NumJob    int = 6
	)

	var x atomic.Int32

	ctrl := jobsched.NewWithoutFeedback(
		getJobHandlerNoFeedbackAddOne(&x),
		&jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
			NumWorker: NumWorker,
		},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...,
	)
	ctrl.Run()
	checkXAndPanicRecords(t, &x, ctrl.PanicRecords(), int32(NumJob))
}

func TestController_Setup(t *testing.T) {
	t.Parallel()

	const NumWorker int = 3

	var setupCounter [NumWorker]atomic.Int32

	prs := jobsched.RunWithoutFeedback(
		getJobHandlerNoOp[jobsched.NoFeedback](),
		&jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
			NumWorker: NumWorker,
			Setup: func(
				_ jobsched.Controller[
					int,
					jobsched.NoProperty,
					jobsched.NoFeedback,
				],
				rank int,
			) {
				setupCounter[rank].Add(1)
			},
		},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker<<1)...,
	)
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
	t.Parallel()

	const (
		PanicMsg      = "test panic"
		NumWorker int = 3
	)

	var (
		setupCounter [NumWorker]atomic.Int32
		wg           sync.WaitGroup
	)

	wg.Add(NumWorker)
	prs := jobsched.RunWithoutFeedback(
		func(concurrency.Canceler, int, int) (
			newJobs []*jobsched.MetaJob[int, jobsched.NoProperty],
			feedback jobsched.NoFeedback,
		) {
			wg.Done()
			wg.Wait() // block the worker to ensure that each worker is ready to panic
			panic(PanicMsg)
		},
		&jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
			NumWorker: NumWorker,
			Setup: func(
				_ jobsched.Controller[
					int,
					jobsched.NoProperty,
					jobsched.NoFeedback,
				],
				rank int,
			) {
				setupCounter[rank].Add(1)
			},
		},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker<<1)...,
	)

	for rank := range setupCounter {
		if gotCtr := setupCounter[rank].Load(); gotCtr != 1 {
			t.Errorf("got setupCounter[%d] %d; want 1", rank, gotCtr)
		}
	}

	if len(prs) != NumWorker {
		t.Errorf("got len(prs) %d; want %d", len(prs), NumWorker)
	}

	for _, pr := range prs {
		if !strings.HasPrefix(pr.GoroutineName(), "worker ") {
			t.Error(pr)
		} else {
			msg, ok := pr.PanicValue().(string)
			if !ok || msg != PanicMsg {
				t.Error(pr)
			}
		}
	}
}

func TestController_Cleanup(t *testing.T) {
	t.Parallel()

	const NumWorker int = 3

	var cleanupCounter [NumWorker]atomic.Int32

	prs := jobsched.RunWithoutFeedback(
		getJobHandlerNoOp[jobsched.NoFeedback](),
		&jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
			NumWorker: NumWorker,
			Cleanup: func(
				_ jobsched.Controller[
					int,
					jobsched.NoProperty,
					jobsched.NoFeedback,
				],
				rank int,
			) {
				cleanupCounter[rank].Add(1)
			},
		},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker<<1)...,
	)
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
	t.Parallel()

	const (
		PanicMsg      = "test panic"
		NumWorker int = 3
	)

	var (
		cleanupCounter [NumWorker]atomic.Int32
		wg             sync.WaitGroup
	)

	wg.Add(NumWorker)
	prs := jobsched.RunWithoutFeedback(
		func(concurrency.Canceler, int, int) (
			newJobs []*jobsched.MetaJob[int, jobsched.NoProperty],
			feedback jobsched.NoFeedback,
		) {
			wg.Done()
			wg.Wait() // block the worker to ensure that each worker is ready to panic
			panic(PanicMsg)
		},
		&jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
			NumWorker: NumWorker,
			Cleanup: func(
				_ jobsched.Controller[
					int,
					jobsched.NoProperty,
					jobsched.NoFeedback,
				],
				rank int,
			) {
				cleanupCounter[rank].Add(1)
			},
		},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker<<1)...,
	)

	for rank := range cleanupCounter {
		if gotCtr := cleanupCounter[rank].Load(); gotCtr != 1 {
			t.Errorf("got cleanupCounter[%d] %d; want 1", rank, gotCtr)
		}
	}

	if len(prs) != NumWorker {
		t.Errorf("got len(prs) %d; want %d", len(prs), NumWorker)
	}

	for _, pr := range prs {
		if !strings.HasPrefix(pr.GoroutineName(), "worker ") {
			t.Error(pr)
		} else {
			msg, ok := pr.PanicValue().(string)
			if !ok || msg != PanicMsg {
				t.Error(pr)
			}
		}
	}
}

func TestController_Cleanup_SetupPanic(t *testing.T) {
	t.Parallel()

	const (
		PanicMsg      = "test panic"
		NumWorker int = 3
	)

	var (
		cleanupCounter [NumWorker]atomic.Int32
		wg             sync.WaitGroup
	)

	wg.Add(NumWorker)
	prs := jobsched.RunWithoutFeedback(
		getJobHandlerNoOp[jobsched.NoFeedback](),
		&jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
			NumWorker: NumWorker,
			Setup: func(
				jobsched.Controller[
					int,
					jobsched.NoProperty,
					jobsched.NoFeedback,
				],
				int,
			) {
				wg.Done()
				wg.Wait() // block the worker to ensure that each worker is ready to panic
				panic(PanicMsg)
			},
			Cleanup: func(
				_ jobsched.Controller[
					int,
					jobsched.NoProperty,
					jobsched.NoFeedback,
				],
				rank int,
			) {
				cleanupCounter[rank].Add(1)
			},
		},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker<<1)...,
	)

	for rank := range cleanupCounter {
		if gotCtr := cleanupCounter[rank].Load(); gotCtr != 0 {
			t.Errorf("got cleanupCounter[%d] %d; want 0", rank, gotCtr)
		}
	}

	if len(prs) != NumWorker {
		t.Errorf("got len(prs) %d; want %d", len(prs), NumWorker)
	}

	for _, pr := range prs {
		if !strings.HasPrefix(pr.GoroutineName(), "worker ") {
			t.Error(pr)
		} else {
			msg, ok := pr.PanicValue().(string)
			if !ok || msg != PanicMsg {
				t.Error(pr)
			}
		}
	}
}

func TestController_SetupAndCleanup(t *testing.T) {
	t.Parallel()

	const NumWorker int = 3

	var setupCounter, cleanupCounter [NumWorker]atomic.Int32

	prs := jobsched.RunWithoutFeedback(
		getJobHandlerNoOp[jobsched.NoFeedback](),
		&jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
			NumWorker: NumWorker,
			Setup: func(
				_ jobsched.Controller[
					int,
					jobsched.NoProperty,
					jobsched.NoFeedback,
				],
				rank int,
			) {
				setupCounter[rank].Add(1)
			},
			Cleanup: func(
				_ jobsched.Controller[
					int,
					jobsched.NoProperty,
					jobsched.NoFeedback,
				],
				rank int,
			) {
				cleanupCounter[rank].Add(1)
			},
		},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker<<1)...,
	)
	if len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}

	for rank := range NumWorker {
		if gotCtr := setupCounter[rank].Load(); gotCtr != 1 {
			t.Errorf("got setupCounter[%d] %d; want 1", rank, gotCtr)
		}

		if gotCtr := cleanupCounter[rank].Load(); gotCtr != 1 {
			t.Errorf("got cleanupCounter[%d] %d; want 1", rank, gotCtr)
		}
	}
}

func TestController_JobHandlerRankUnique(t *testing.T) {
	t.Parallel()

	for n := -1; n <= 100; n++ {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			t.Parallel()

			numWorker := n
			if numWorker <= 0 {
				numWorker = max(1, runtime.NumCPU()-2)
			}

			counter := make([]atomic.Int32, numWorker)

			var wg sync.WaitGroup

			wg.Add(numWorker)

			prs := jobsched.RunWithoutFeedback(
				func(_ concurrency.Canceler, rank, _ int) (
					newJobs []*jobsched.MetaJob[int, jobsched.NoProperty],
					feedback jobsched.NoFeedback,
				) {
					counter[rank].Add(1)
					wg.Done()
					wg.Wait() // block the worker to ensure that each worker processes exactly one job

					return
				},
				&jobsched.Options[
					int,
					jobsched.NoProperty,
					jobsched.NoFeedback,
				]{
					NumWorker: n,
				},
				make(
					[]*jobsched.MetaJob[int, jobsched.NoProperty],
					numWorker,
				)...,
			)
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

func TestController_Cancel(t *testing.T) {
	t.Parallel()

	const NumWorker int = 4

	var (
		x           atomic.Int32
		feedbackSum int
	)

	ctrl := jobsched.New(
		getJobHandlerAddOneCopyJob(&x),
		getSumFeedbackHandler(&feedbackSum),
		&jobsched.Options[int, jobsched.NoProperty, int]{NumWorker: NumWorker},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker<<1)...,
	)
	if ctrl == nil {
		t.Fatal("got nil controller")
	}

	canceler := ctrl.Canceler()
	if canceler == nil {
		t.Fatal("got nil canceler")
	}

	ctrl.Launch()

	doneC := make(chan struct{})
	go func(
		ctrl jobsched.Controller[int, jobsched.NoProperty, int],
		doneC chan<- struct{},
	) {
		defer close(doneC)

		ctrl.Wait()
	}(ctrl, doneC)

	timer := time.NewTimer(time.Second)

	t.Cleanup(func() {
		timer.Stop()
	})

	time.Sleep(time.Millisecond) // wait for workers to handle some jobs
	canceler.Cancel()

	xAfterCancellation := x.Load()
	if xAfterCancellation == 0 {
		t.Error("got x after cancellation 0; want greater than 0")
	}

	select {
	case <-timer.C:
		t.Error("timeout after canceling")
	case <-doneC:
		checkXFeedbackSumAndPanicRecords(
			t,
			&x,
			feedbackSum,
			ctrl.PanicRecords(),
			xAfterCancellation,
			int(xAfterCancellation),
		)
	}
}

func TestController_Cancel_NoFeedback(t *testing.T) {
	t.Parallel()

	const NumWorker int = 4

	var x atomic.Int32

	ctrl := jobsched.NewWithoutFeedback(
		getJobHandlerNoFeedbackAddOneCopyJob(&x),
		&jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
			NumWorker: NumWorker,
		},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker<<1)...,
	)
	if ctrl == nil {
		t.Fatal("got nil controller")
	}

	canceler := ctrl.Canceler()
	if canceler == nil {
		t.Fatal("got nil canceler")
	}

	ctrl.Launch()

	doneC := make(chan struct{})
	go func(
		ctrl jobsched.Controller[int, jobsched.NoProperty, jobsched.NoFeedback],
		doneC chan<- struct{},
	) {
		defer close(doneC)

		ctrl.Wait()
	}(ctrl, doneC)

	timer := time.NewTimer(time.Second)

	t.Cleanup(func() {
		timer.Stop()
	})

	time.Sleep(time.Millisecond) // wait for workers to handle some jobs
	canceler.Cancel()

	xAfterCancellation := x.Load()
	if xAfterCancellation == 0 {
		t.Error("got x after cancellation 0; want greater than 0")
	}

	select {
	case <-timer.C:
		t.Error("timeout after canceling")
	case <-doneC:
		checkXAndPanicRecords(t, &x, ctrl.PanicRecords(), xAfterCancellation)
	}
}

// getJobHandlerNoOp returns a
// github.com/donyori/gogo/concurrency/framework/jobsched.JobHandler[
// int, jobsched.NoProperty, Feedback]
// that does nothing and returns zero values.
func getJobHandlerNoOp[Feedback any]() jobsched.JobHandler[
	int,
	jobsched.NoProperty,
	Feedback,
] {
	return func(concurrency.Canceler, int, int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty],
		feedback Feedback,
	) {
		return // do nothing
	}
}

// getJobHandlerAddOne returns a
// github.com/donyori/gogo/concurrency/framework/jobsched.JobHandler[
// int, jobsched.NoProperty, int]
// that adds one to p and returns (nil, 1).
//
// It returns nil if p is nil.
func getJobHandlerAddOne(
	p *atomic.Int32,
) jobsched.JobHandler[int, jobsched.NoProperty, int] {
	if p == nil {
		return nil
	}

	return func(concurrency.Canceler, int, int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty],
		feedback int,
	) {
		p.Add(1)
		return nil, 1
	}
}

// getJobHandlerNoFeedbackAddOne returns a
// github.com/donyori/gogo/concurrency/framework/jobsched.JobHandler[
// int, jobsched.NoProperty, jobsched.NoFeedback]
// that adds one to p and returns zero values.
//
// It returns nil if p is nil.
func getJobHandlerNoFeedbackAddOne(
	p *atomic.Int32,
) jobsched.JobHandler[int, jobsched.NoProperty, jobsched.NoFeedback] {
	if p == nil {
		return nil
	}

	return func(concurrency.Canceler, int, int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty],
		feedback jobsched.NoFeedback,
	) {
		p.Add(1)
		return
	}
}

// getJobHandlerAddOneCopyJob returns a
// github.com/donyori/gogo/concurrency/framework/jobsched.JobHandler[
// int, jobsched.NoProperty, int]
// that adds one to p and returns a copy of the job and feedback as one.
// If a cancellation signal is detected,
// the job handler does nothing and returns zero values.
//
// Such a job handler makes the framework run
// until a cancellation signal is detected.
//
// getJobHandlerAddOneCopyJob returns nil if p is nil.
func getJobHandlerAddOneCopyJob(
	p *atomic.Int32,
) jobsched.JobHandler[int, jobsched.NoProperty, int] {
	if p == nil {
		return nil
	}

	return func(canceler concurrency.Canceler, _, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty],
		feedback int,
	) {
		if canceler.Canceled() {
			return
		}

		p.Add(1)

		return []*jobsched.MetaJob[int, jobsched.NoProperty]{{Job: job}}, 1
	}
}

// getJobHandlerNoFeedbackAddOneCopyJob returns a
// github.com/donyori/gogo/concurrency/framework/jobsched.JobHandler[
// int, jobsched.NoProperty, jobsched.NoFeedback]
// that adds one to p and returns a copy of the job and zero-value feedback.
// If a cancellation signal is detected,
// the job handler does nothing and returns zero values.
//
// Such a job handler makes the framework run
// until a cancellation signal is detected.
//
// getJobHandlerNoFeedbackAddOneCopyJob returns nil if p is nil.
func getJobHandlerNoFeedbackAddOneCopyJob(
	p *atomic.Int32,
) jobsched.JobHandler[int, jobsched.NoProperty, jobsched.NoFeedback] {
	if p == nil {
		return nil
	}

	return func(canceler concurrency.Canceler, _, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty],
		feedback jobsched.NoFeedback,
	) {
		if canceler.Canceled() {
			return
		}

		p.Add(1)

		newJobs = []*jobsched.MetaJob[int, jobsched.NoProperty]{{Job: job}}

		return
	}
}

// getSumFeedbackHandler returns a
// github.com/donyori/gogo/concurrency/framework/jobsched.FeedbackHandler[int]
// that adds the feedback (of type int) to *ptr.
//
// It returns nil if ptr is nil.
func getSumFeedbackHandler(ptr *int) jobsched.FeedbackHandler[int] {
	if ptr == nil {
		return nil
	}

	return func(_ concurrency.Canceler, feedback int) {
		*ptr += feedback
	}
}

func checkXAndPanicRecords(
	t *testing.T,
	pX *atomic.Int32,
	prs []*framework.GoroutinePanicError,
	wantX int32,
) {
	t.Helper()

	var x int32
	if pX != nil {
		x = pX.Load()
	}

	if x != wantX {
		t.Errorf("got x %d; want %d", x, wantX)
	}

	if len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func checkXFeedbackSumAndPanicRecords(
	t *testing.T,
	pX *atomic.Int32,
	feedbackSum int,
	prs []*framework.GoroutinePanicError,
	wantX int32,
	wantFeedbackSum int,
) {
	t.Helper()

	var x int32
	if pX != nil {
		x = pX.Load()
	}

	if x != wantX {
		t.Errorf("got x %d; want %d", x, wantX)
	}

	if feedbackSum != wantFeedbackSum {
		t.Errorf("got sum of feedback %d; want %d",
			feedbackSum, wantFeedbackSum)
	}

	if len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}
