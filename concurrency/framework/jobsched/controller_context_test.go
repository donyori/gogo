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
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/donyori/gogo/concurrency/framework/jobsched"
)

func TestController_Context_NoCancellation(t *testing.T) {
	t.Parallel()

	const (
		NumWorker int = 4
		NumJob    int = 16
	)

	var (
		x           atomic.Int32
		feedbackSum int
	)

	prs := jobsched.RunWithContext(
		context.Background(),
		getJobHandlerAddOne(&x),
		getSumFeedbackHandler(&feedbackSum),
		&jobsched.Options[int, jobsched.NoProperty, int]{NumWorker: NumWorker},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...,
	)
	checkXFeedbackSumAndPanicRecords(
		t,
		&x,
		feedbackSum,
		prs,
		int32(NumJob),
		NumJob,
	)
}

func TestController_Context_NoCancellationNoFeedback(t *testing.T) {
	t.Parallel()

	const (
		NumWorker int = 4
		NumJob    int = 16
	)

	var x atomic.Int32

	prs := jobsched.RunWithContextWithoutFeedback(
		context.Background(),
		getJobHandlerNoFeedbackAddOne(&x),
		&jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
			NumWorker: NumWorker,
		},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...,
	)
	checkXAndPanicRecords(t, &x, prs, int32(NumJob))
}

func TestController_CancelContext_WithCancel(t *testing.T) {
	t.Parallel()

	testControllerCancelContextFunc(
		t,
		time.Second,
		func() (ctx context.Context, cancel context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
		func(_ context.Context, cancel context.CancelFunc) {
			time.Sleep(time.Millisecond) // wait for workers to handle some jobs
			cancel()
		},
	)
}

func TestController_CancelContext_WithCancelNoFeedback(t *testing.T) {
	t.Parallel()

	testControllerNoFeedbackCancelContextFunc(
		t,
		time.Second,
		func() (ctx context.Context, cancel context.CancelFunc) {
			return context.WithCancel(context.Background())
		},
		func(_ context.Context, cancel context.CancelFunc) {
			time.Sleep(time.Millisecond) // wait for workers to handle some jobs
			cancel()
		},
	)
}

func TestController_CancelContext_WithTimeout(t *testing.T) {
	t.Parallel()

	testControllerCancelContextFunc(
		t,
		time.Second,
		func() (ctx context.Context, cancel context.CancelFunc) {
			return context.WithTimeout(
				context.Background(),
				time.Millisecond*10,
			)
		},
		nil,
	)
}

func TestController_CancelContext_WithTimeoutNoFeedback(t *testing.T) {
	t.Parallel()

	testControllerNoFeedbackCancelContextFunc(
		t,
		time.Second,
		func() (ctx context.Context, cancel context.CancelFunc) {
			return context.WithTimeout(
				context.Background(),
				time.Millisecond*10,
			)
		},
		nil,
	)
}

// testControllerCancelContextFunc is the common framework
// for testing a controller with context by canceling the context.
func testControllerCancelContextFunc(
	t *testing.T,
	timeout time.Duration,
	getContext func() (ctx context.Context, cancel context.CancelFunc),
	manuallyCancel func(ctx context.Context, cancel context.CancelFunc),
) {
	t.Helper()

	const NumWorker int = 4

	var (
		x           atomic.Int32
		feedbackSum int
	)

	ctx, cancel := getContext()
	if ctx == nil {
		t.Error("got nil context")
		return
	} else if cancel == nil {
		t.Error("got nil cancel function")
		return
	}

	t.Cleanup(cancel)

	ctrl := jobsched.NewWithContext(
		ctx,
		getJobHandlerAddOneCopyJob(&x),
		getSumFeedbackHandler(&feedbackSum),
		&jobsched.Options[int, jobsched.NoProperty, int]{NumWorker: NumWorker},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker<<1)...,
	)
	if ctrl == nil {
		t.Error("got nil controller")
		return
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

	timer := time.NewTimer(timeout)

	t.Cleanup(func() {
		timer.Stop()
	})

	if manuallyCancel != nil {
		manuallyCancel(ctx, cancel)
	} else {
		select {
		case <-ctx.Done():
		case <-timer.C:
			t.Error("timeout before the context is done")
			return
		}
	}

	xAfterCancellation := x.Load()
	if xAfterCancellation == 0 {
		t.Error("got x after cancellation 0; want greater than 0")
	}

	select {
	case <-doneC:
		checkXFeedbackSumAndPanicRecords(
			t,
			&x,
			feedbackSum,
			ctrl.PanicRecords(),
			xAfterCancellation,
			int(xAfterCancellation),
		)
	case <-timer.C:
		t.Error("timeout after canceling")
	}
}

// testControllerNoFeedbackCancelContextFunc is the no feedback version
// common framework for testing a controller with context
// by canceling the context.
func testControllerNoFeedbackCancelContextFunc(
	t *testing.T,
	timeout time.Duration,
	getContext func() (ctx context.Context, cancel context.CancelFunc),
	manuallyCancel func(ctx context.Context, cancel context.CancelFunc),
) {
	t.Helper()

	const NumWorker int = 4

	var x atomic.Int32

	ctx, cancel := getContext()
	if ctx == nil {
		t.Error("got nil context")
		return
	} else if cancel == nil {
		t.Error("got nil cancel function")
		return
	}

	t.Cleanup(cancel)

	ctrl := jobsched.NewWithContextWithoutFeedback(
		ctx,
		getJobHandlerNoFeedbackAddOneCopyJob(&x),
		&jobsched.Options[int, jobsched.NoProperty, jobsched.NoFeedback]{
			NumWorker: NumWorker,
		},
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumWorker<<1)...,
	)
	if ctrl == nil {
		t.Error("got nil controller")
		return
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

	timer := time.NewTimer(timeout)

	t.Cleanup(func() {
		timer.Stop()
	})

	if manuallyCancel != nil {
		manuallyCancel(ctx, cancel)
	} else {
		select {
		case <-ctx.Done():
		case <-timer.C:
			t.Error("timeout before the context is done")
			return
		}
	}

	xAfterCancellation := x.Load()
	if xAfterCancellation == 0 {
		t.Error("got x after cancellation 0; want greater than 0")
	}

	select {
	case <-doneC:
		checkXAndPanicRecords(t, &x, ctrl.PanicRecords(), xAfterCancellation)
	case <-timer.C:
		t.Error("timeout after canceling")
	}
}
