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
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/donyori/gogo/concurrency"
	"github.com/donyori/gogo/concurrency/framework/jobsched"
)

func TestController_Input_BeforeLaunch(t *testing.T) {
	t.Parallel()

	const NumJob int = 3

	var (
		x           atomic.Int32
		feedbackSum int
	)

	ctrl := jobsched.New(
		getJobHandlerAddOne(&x),
		getSumFeedbackHandler(&feedbackSum),
		nil,
	)

	gotInput := ctrl.Input(
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...,
	)
	if gotInput != NumJob {
		t.Errorf("before calling Launch, got %d; want %d", gotInput, NumJob)
	}

	ctrl.Run()
	checkXFeedbackSumAndPanicRecords(
		t,
		&x,
		feedbackSum,
		ctrl.PanicRecords(),
		int32(NumJob),
		NumJob,
	)
}

func TestController_Input_BeforeLaunch_Concurrency(t *testing.T) {
	t.Parallel()

	const (
		NumWorker int = 4

		MaxJob         int = 8
		NumJobPerInput int = 10

		WantX int32 = int32(NumJobPerInput * MaxJob * (1 + MaxJob) >> 1)

		WantFeedbackSum int = NumJobPerInput * MaxJob
	)

	synctest.Test(t, func(t *testing.T) {
		var (
			x           atomic.Int32
			feedbackSum int
		)

		ctrl := jobsched.New(
			getJobHandlerAddJob(&x),
			getSumFeedbackHandler(&feedbackSum),
			&jobsched.Options[int, jobsched.NoProperty, int]{
				NumWorker: NumWorker,
			},
		)
		startC := make(chan struct{})

		for job := 1; job <= MaxJob; job++ {
			go func(t *testing.T, startC <-chan struct{}, job int) {
				t.Helper()

				mjs := make(
					[]*jobsched.MetaJob[int, jobsched.NoProperty],
					NumJobPerInput,
				)
				for i := range mjs {
					mjs[i] = &jobsched.MetaJob[int, jobsched.NoProperty]{
						Job: job,
					}
				}

				<-startC

				gotInput := ctrl.Input(mjs...)
				if gotInput != NumJobPerInput {
					t.Errorf("job %d, got %d; want %d",
						job, gotInput, NumJobPerInput)
				}
			}(t, startC, job)
		}

		synctest.Wait()
		close(startC)

		synctest.Wait()
		ctrl.Run()
		checkXFeedbackSumAndPanicRecords(
			t,
			&x,
			feedbackSum,
			ctrl.PanicRecords(),
			WantX,
			WantFeedbackSum,
		)
	})
}

func TestController_Input_DuringLaunch(t *testing.T) {
	t.Parallel()

	const NumJob int = 3

	synctest.Test(t, func(t *testing.T) {
		var (
			x           atomic.Int32
			feedbackSum int
		)

		ctrl := jobsched.New(
			getJobHandlerAddOne(&x),
			getSumFeedbackHandler(&feedbackSum),
			nil,
		)

		go func() {
			time.Sleep(time.Millisecond) // ensure that ctrl.Launch has been called first
			ctrl.Launch()
		}()

		time.Sleep(time.Millisecond * 2) // ensure that ctrl.Launch has been called first

		gotInput := ctrl.Input(
			make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...,
		)
		if gotInput != NumJob {
			t.Errorf("during calling Launch, got %d; want %d",
				gotInput, NumJob)
		}

		synctest.Wait()
		ctrl.Wait()
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

func TestController_Input_DuringLaunch_Concurrency(t *testing.T) {
	t.Parallel()

	const (
		NumWorker int = 4

		MaxJob         int = 8
		NumJobPerInput int = 10

		WantX int32 = int32(NumJobPerInput * MaxJob * (1 + MaxJob) >> 1)

		WantFeedbackSum int = NumJobPerInput * MaxJob
	)

	synctest.Test(t, func(t *testing.T) {
		var (
			x           atomic.Int32
			feedbackSum int
		)

		ctrl := jobsched.New(
			getJobHandlerAddJob(&x),
			getSumFeedbackHandler(&feedbackSum),
			&jobsched.Options[int, jobsched.NoProperty, int]{
				NumWorker: NumWorker,
			},
		)
		startC := make(chan struct{})

		go func(startC <-chan struct{}) {
			<-startC
			time.Sleep(time.Millisecond) // ensure that ctrl.Launch has been called first
			ctrl.Launch()
		}(startC)

		for job := 1; job <= MaxJob; job++ {
			go func(t *testing.T, startC <-chan struct{}, job int) {
				t.Helper()

				mjs := make(
					[]*jobsched.MetaJob[int, jobsched.NoProperty],
					NumJobPerInput,
				)
				for i := range mjs {
					mjs[i] = &jobsched.MetaJob[int, jobsched.NoProperty]{
						Job: job,
					}
				}

				<-startC
				time.Sleep(time.Millisecond * 2) // ensure that ctrl.Launch has been called first

				gotInput := ctrl.Input(mjs...)
				if gotInput != NumJobPerInput {
					t.Errorf("job %d, got %d; want %d",
						job, gotInput, NumJobPerInput)
				}
			}(t, startC, job)
		}

		synctest.Wait()
		close(startC)

		time.Sleep(time.Millisecond * 3) // ensure that ctrl.Launch has been called first
		synctest.Wait()
		ctrl.Wait()
		checkXFeedbackSumAndPanicRecords(
			t,
			&x,
			feedbackSum,
			ctrl.PanicRecords(),
			WantX,
			WantFeedbackSum,
		)
	})
}

func TestController_Input_AfterLaunch(t *testing.T) {
	t.Parallel()

	const NumJob int = 3

	var (
		x           atomic.Int32
		feedbackSum int
	)

	ctrl := jobsched.New(
		getJobHandlerAddOne(&x),
		getSumFeedbackHandler(&feedbackSum),
		nil,
	)
	ctrl.Launch()

	gotInput := ctrl.Input(
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...,
	)
	if gotInput != NumJob {
		t.Errorf("after calling Launch, got %d; want %d", gotInput, NumJob)
	}

	ctrl.Wait()
	checkXFeedbackSumAndPanicRecords(
		t,
		&x,
		feedbackSum,
		ctrl.PanicRecords(),
		int32(NumJob),
		NumJob,
	)
}

func TestController_Input_AfterLaunch_Concurrency(t *testing.T) {
	t.Parallel()

	const (
		NumWorker int = 4

		MaxJob         int = 8
		NumJobPerInput int = 10

		WantX int32 = int32(NumJobPerInput * MaxJob * (1 + MaxJob) >> 1)

		WantFeedbackSum int = NumJobPerInput * MaxJob
	)

	synctest.Test(t, func(t *testing.T) {
		var (
			x           atomic.Int32
			feedbackSum int
		)

		ctrl := jobsched.New(
			getJobHandlerAddJob(&x),
			getSumFeedbackHandler(&feedbackSum),
			&jobsched.Options[int, jobsched.NoProperty, int]{
				NumWorker: NumWorker,
			},
		)
		startC := make(chan struct{})

		ctrl.Launch()

		for job := 1; job <= MaxJob; job++ {
			go func(t *testing.T, startC <-chan struct{}, job int) {
				t.Helper()

				mjs := make(
					[]*jobsched.MetaJob[int, jobsched.NoProperty],
					NumJobPerInput,
				)
				for i := range mjs {
					mjs[i] = &jobsched.MetaJob[int, jobsched.NoProperty]{
						Job: job,
					}
				}

				<-startC

				gotInput := ctrl.Input(mjs...)
				if gotInput != NumJobPerInput {
					t.Errorf("job %d, got %d; want %d",
						job, gotInput, NumJobPerInput)
				}
			}(t, startC, job)
		}

		synctest.Wait()
		close(startC)

		synctest.Wait()
		ctrl.Wait()
		checkXFeedbackSumAndPanicRecords(
			t,
			&x,
			feedbackSum,
			ctrl.PanicRecords(),
			WantX,
			WantFeedbackSum,
		)
	})
}

func TestController_Input_DuringWaiting(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		var (
			x           atomic.Int32
			feedbackSum int
		)

		ctrl := jobsched.New(
			getJobHandlerAddOne(&x),
			getSumFeedbackHandler(&feedbackSum),
			nil,
		)
		ctrl.Launch()

		go func() {
			time.Sleep(time.Millisecond) // ensure that ctrl.Wait has been called first
			ctrl.Wait()
		}()

		time.Sleep(time.Millisecond * 2) // ensure that ctrl.Wait has been called first

		gotInput := ctrl.Input(
			make([]*jobsched.MetaJob[int, jobsched.NoProperty], 3)...,
		)
		if gotInput != 0 {
			t.Errorf("during waiting, got %d; want 0", gotInput)
		}

		synctest.Wait()

		if gotX := x.Load(); gotX != 0 {
			t.Errorf("got x %d; want 0", gotX)
		}

		if feedbackSum != 0 {
			t.Errorf("got sum of feedback %d; want 0", feedbackSum)
		}

		if prs := ctrl.PanicRecords(); len(prs) > 0 {
			t.Errorf("panic %q", prs)
		}
	})
}

func TestController_Input_AfterWait(t *testing.T) {
	t.Parallel()

	var (
		x           atomic.Int32
		feedbackSum int
	)

	ctrl := jobsched.New(
		getJobHandlerAddOne(&x),
		getSumFeedbackHandler(&feedbackSum),
		nil,
	)
	ctrl.Run()

	gotInput := ctrl.Input(
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], 3)...,
	)
	if gotInput != 0 {
		t.Errorf("after calling Wait, got %d; want 0", gotInput)
	}

	if gotX := x.Load(); gotX != 0 {
		t.Errorf("got x %d; want 0", gotX)
	}

	if feedbackSum != 0 {
		t.Errorf("got sum of feedback %d; want 0", feedbackSum)
	}

	if prs := ctrl.PanicRecords(); len(prs) > 0 {
		t.Errorf("panic %q", prs)
	}
}

func TestController_Input_AfterIneffectiveWait(t *testing.T) {
	t.Parallel()

	const NumJob int = 3

	var (
		x           atomic.Int32
		feedbackSum int
	)

	ctrl := jobsched.New(
		getJobHandlerAddOne(&x),
		getSumFeedbackHandler(&feedbackSum),
		nil,
	)

	gotIWait := ctrl.Wait()
	if gotIWait != -1 {
		t.Errorf("got %d on ineffective call to Wait; want -1", gotIWait)
	}

	gotInput := ctrl.Input(
		make([]*jobsched.MetaJob[int, jobsched.NoProperty], NumJob)...,
	)
	if gotInput != NumJob {
		t.Errorf("after calling Wait ineffectively, got %d; want %d",
			gotInput, NumJob)
	}

	ctrl.Run()
	checkXFeedbackSumAndPanicRecords(
		t,
		&x,
		feedbackSum,
		ctrl.PanicRecords(),
		int32(NumJob),
		NumJob,
	)
}

// getJobHandlerAddJob returns a
// github.com/donyori/gogo/concurrency/framework/jobsched.JobHandler[
// int, jobsched.NoProperty, int]
// that adds the specified "job" to p and returns (nil, 1).
//
// It returns nil if p is nil.
func getJobHandlerAddJob(
	p *atomic.Int32,
) jobsched.JobHandler[int, jobsched.NoProperty, int] {
	if p == nil {
		return nil
	}

	return func(_ concurrency.Canceler, _, job int) (
		newJobs []*jobsched.MetaJob[int, jobsched.NoProperty],
		feedback int,
	) {
		p.Add(int32(job)) //gosec:disable G115 -- job is small
		return nil, 1
	}
}
