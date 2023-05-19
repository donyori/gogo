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

package jobsched

import (
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/donyori/gogo/concurrency"
	"github.com/donyori/gogo/concurrency/framework"
	"github.com/donyori/gogo/concurrency/framework/internal/quitdevice"
	"github.com/donyori/gogo/errors"
)

// Controller is a controller for this job scheduling framework.
//
// It is used to launch, quit, and wait for the job.
// Also, it is used to input new jobs.
//
// The first type parameter Job is the type of jobs.
// The second type parameter Properties is the type of custom properties
// in the meta information of jobs.
// The third type parameter Feedback is the type of feedback on the jobs,
// which is collected through a dedicated channel returned by FeedbackChan.
type Controller[Job, Properties, Feedback any] interface {
	framework.Controller

	// FeedbackChan returns the channel for feedback on jobs.
	//
	// The channel is closed when all jobs are finished or quit.
	// The client is responsible for receiving feedback from the channel
	// in time to avoid blocking the framework.
	//
	// FeedbackChan returns nil if and only if
	// the type of feedback is NoFeedback.
	// In this case, the framework skips sending feedback
	// and the client should not listen to the channel.
	FeedbackChan() <-chan Feedback

	// Input enables the client to input new jobs.
	//
	// If an item in metaJob is nil, it is treated as a zero-value job
	// (with the field Job to its zero value, Meta.Priority to 0,
	// Meta.CreationTime to time.Now(), and Meta.Custom to its zero value).
	// If an item in metaJob has the field Meta.CreationTime with a zero value,
	// this field is set to time.Now() by the framework.
	//
	// It returns the number of jobs input successfully.
	//
	// The client can input new jobs before the first effective call to
	// the method Wait (i.e., the call after invoking the method Launch).
	// After calling the method Wait, Input does nothing and returns 0.
	// Note that the method Run calls Wait inside.
	Input(metaJob ...*MetaJob[Job, Properties]) int
}

// NoFeedback is a special case of feedback type
// to indicate that the job handler has no feedback.
// In this case, the method FeedbackChan of Controller returns nil,
// and the framework skips sending feedback.
type NoFeedback struct{}

// noFeedbackType is the reflect.Type of NoFeedback.
var noFeedbackType = reflect.TypeOf(NoFeedback{})

// JobHandler is a function to process a job.
//
// The first type parameter Job is the type of jobs.
// The second type parameter Properties is the type of custom properties
// in the meta information of jobs.
// The third type parameter Feedback is the type of feedback on the jobs,
// which is collected through a dedicated channel.
//
// The first parameter is the job to be processed.
//
// The second parameter is a device to acquire the channel for the quit signal,
// detect the quit signal, and broadcast a quit signal to interrupt
// all job processors.
//
// It returns new jobs generated during or after processing
// the current job, called newJobs.
// If an item in newJobs is nil, it is treated as a zero-value job
// (with the field Job to its zero value, Meta.Priority to 0,
// Meta.CreationTime to time.Now(), and Meta.Custom to its zero value).
// If an item in newJobs has the field Meta.CreationTime with a zero value,
// this field is set to time.Now() by the framework.
//
// In addition, it also returns feedback on this job,
// which is sent to a dedicated channel by the framework.
// The client is responsible for receiving feedback from that channel
// in time to avoid blocking the framework.
// In particular, if the type of feedback is NoFeedback,
// the framework skips sending feedback and sets that channel to nil,
// and the client should not listen to that channel.
type JobHandler[Job, Properties, Feedback any] func(
	job Job, quitDevice framework.QuitDevice,
) (newJobs []*MetaJob[Job, Properties], feedback Feedback)

// New creates a new Controller.
//
// The first type parameter Job is the type of jobs.
// The second type parameter Properties is the type of custom properties
// in the meta information of jobs.
// The third type parameter Feedback is the type of feedback on the jobs,
// which is collected through a dedicated channel.
//
// n is the number of goroutines to process jobs.
// If n is non-positive, runtime.NumCPU() is used instead.
//
// handler is the job handler.
// It panics if handler is nil.
//
// jobQueueMaker is a maker to create a new job queue.
// It enables the client to make custom JobQueue.
// It panics if jobQueueMaker is nil.
//
// feedbackChanBufSize is the buffer size of the feedback channel.
// If feedbackChanBufSize is not positive, the feedback channel is unbuffered.
// It only takes effect when the type of feedback is not NoFeedback.
//
// metaJob is initial jobs to process.
// If an item in metaJob is nil, it is treated as a zero-value job
// (with the field Job to its zero value, Meta.Priority to 0,
// Meta.CreationTime to time.Now(), and Meta.Custom to its zero value).
// If an item in metaJob has the field Meta.CreationTime with a zero value,
// this field is set to time.Now() by the framework.
func New[Job, Properties, Feedback any](
	n int,
	handler JobHandler[Job, Properties, Feedback],
	jobQueueMaker JobQueueMaker[Job, Properties],
	feedbackChanBufSize int,
	metaJob ...*MetaJob[Job, Properties],
) Controller[Job, Properties, Feedback] {
	switch {
	case handler == nil:
		panic(errors.AutoMsg("handler is nil"))
	case jobQueueMaker == nil:
		panic(errors.AutoMsg("job queue maker is nil"))
	case n <= 0:
		n = runtime.NumCPU()
	}
	jq := jobQueueMaker.New()
	if len(metaJob) > 0 {
		jq.Enqueue(copyMetaJobs(metaJob)...)
	}
	ctrl := &controller[Job, Properties, Feedback]{
		n:    n,
		h:    handler,
		qd:   quitdevice.NewQuitDevice(),
		jq:   jq,
		ic:   make(chan []*MetaJob[Job, Properties]),
		eqc:  make(chan []*MetaJob[Job, Properties]),
		dqc:  make(chan Job),
		loi:  concurrency.NewOnceIndicator(),
		wsoi: concurrency.NewOnceIndicator(),
	}
	// Use reflect.TypeOf((*Feedback)(nil)).Elem() to work with interface types.
	if reflect.TypeOf((*Feedback)(nil)).Elem() != noFeedbackType {
		if feedbackChanBufSize < 0 {
			feedbackChanBufSize = 0
		}
		ctrl.fc = make(chan Feedback, feedbackChanBufSize)
	}
	return ctrl
}

// Run creates a Controller[Job, Properties, NoFeedback]
// with specified arguments, and then runs it.
// It returns the panic records of the Controller.
//
// The parameters are the same as those of function New.
func Run[Job, Properties any](
	n int,
	handler JobHandler[Job, Properties, NoFeedback],
	jobQueueMaker JobQueueMaker[Job, Properties],
	feedbackChanBufSize int,
	metaJob ...*MetaJob[Job, Properties],
) []framework.PanicRecord {
	ctrl := New(n, handler, jobQueueMaker, feedbackChanBufSize, metaJob...)
	ctrl.Run()
	return ctrl.PanicRecords()
}

// controller is an implementation of interface Controller.
type controller[Job, Properties, Feedback any] struct {
	n int                                   // The number of goroutines to process jobs.
	h JobHandler[Job, Properties, Feedback] // Job handler.

	qd framework.QuitDevice      // Quit device.
	jq JobQueue[Job, Properties] // Job queue.

	ic  chan []*MetaJob[Job, Properties] // Input channel, to input jobs from the client.
	eqc chan []*MetaJob[Job, Properties] // Enqueue channel, to input jobs from workers.
	dqc chan Job                         // Dequeue channel, to dispatch jobs to workers.
	fc  chan Feedback                    // Feedback channel, to collect feedback on jobs.

	pr   framework.PanicRecords    // Panic records.
	wg   sync.WaitGroup            // Wait group for the main process.
	loi  concurrency.OnceIndicator // For launching the framework.
	wsoi concurrency.OnceIndicator // For indicating the start of the first effective call to the method Wait.
	m    sync.Mutex                // Lock to avoid the race condition on jq when calling Launch and Input at the same time.
	lsi  bool                      // An indicator to report whether the method Launch is started.
}

func (ctrl *controller[Job, Properties, Feedback]) QuitChan() <-chan struct{} {
	return ctrl.qd.QuitChan()
}

func (ctrl *controller[Job, Properties, Feedback]) IsQuit() bool {
	return ctrl.qd.IsQuit()
}

func (ctrl *controller[Job, Properties, Feedback]) Quit() {
	ctrl.qd.Quit()
}

func (ctrl *controller[Job, Properties, Feedback]) Launch() {
	ctrl.loi.Do(func() {
		ctrl.m.Lock()
		defer ctrl.m.Unlock()
		ctrl.lsi = true

		ctrl.wg.Add(ctrl.n + 1) // n workers + 1 allocator
		for i := 0; i < ctrl.n; i++ {
			go func(rank int) {
				defer func() {
					if e := recover(); e != nil {
						ctrl.qd.Quit()
						ctrl.pr.Append(framework.PanicRecord{
							Name:    "worker " + strconv.Itoa(rank),
							Content: e,
						})
					}
					ctrl.wg.Done()
				}()
				ctrl.workerProc()
			}(i)
		}
		go func() {
			defer func() {
				if e := recover(); e != nil {
					ctrl.qd.Quit()
					ctrl.pr.Append(framework.PanicRecord{
						Name:    "allocator",
						Content: e,
					})
				}
				ctrl.wg.Done()
			}()
			ctrl.allocatorProc()
		}()
	})
}

func (ctrl *controller[Job, Properties, Feedback]) Wait() int {
	if !ctrl.loi.Test() {
		return -1
	} else if ctrl.fc != nil {
		defer close(ctrl.fc) // close(ctrl.fc) should be after ctrl.qd.Quit(), so defer it first
	}
	defer ctrl.qd.Quit() // for cleanup possible daemon goroutines that wait for a quit signal to exit
	ctrl.wsoi.Do(nil)
	ctrl.wg.Wait()
	return ctrl.pr.Len()
}

func (ctrl *controller[Job, Properties, Feedback]) Run() int {
	ctrl.Launch()
	return ctrl.Wait()
}

func (ctrl *controller[Job, Properties, Feedback]) NumGoroutine() int {
	return ctrl.n
}

func (ctrl *controller[Job, Properties, Feedback]) PanicRecords() []framework.PanicRecord {
	return ctrl.pr.List()
}

func (ctrl *controller[Job, Properties, Feedback]) FeedbackChan() <-chan Feedback {
	return ctrl.fc
}

func (ctrl *controller[Job, Properties, Feedback]) Input(metaJob ...*MetaJob[Job, Properties]) int {
	if ctrl.wsoi.Test() {
		return 0
	}
	mjs := copyMetaJobs(metaJob)
	if !ctrl.loi.Test() && ctrl.inputBeforeLaunch(mjs) {
		return len(mjs)
	}
	select {
	case <-ctrl.qd.QuitChan():
		return 0
	case ctrl.ic <- mjs:
		return len(mjs)
	}
}

// workerProc is the worker main process,
// without panic checking and wg.Done().
func (ctrl *controller[Job, Properties, Feedback]) workerProc() {
	quitChan := ctrl.qd.QuitChan()
	for {
		var mjs []*MetaJob[Job, Properties]
		var fb Feedback
		select {
		case <-quitChan:
			return
		case job, ok := <-ctrl.dqc:
			if !ok {
				return
			}
			mjs, fb = ctrl.h(job, ctrl.qd)
			mjs = copyMetaJobs(mjs)
		}
		if ctrl.fc != nil {
			// The feedback type is not NoFeedback.
			// Send feedback first.
			select {
			case <-quitChan:
				return
			case ctrl.fc <- fb:
			}
		}
		// Always send new jobs to the allocator,
		// regardless of whether jobs are empty.
		select {
		case <-quitChan:
			return
		case ctrl.eqc <- mjs:
		}
	}
}

// allocatorProc is the allocator main process,
// without panic checking and wg.Done().
func (ctrl *controller[Job, Properties, Feedback]) allocatorProc() {
	defer close(ctrl.dqc)
	var dqc chan<- Job // disable dqc at the beginning
	var job Job
	if ctrl.jq.Len() > 0 {
		job = ctrl.jq.Dequeue()
		dqc = ctrl.dqc // enable dqc
	}
	ctr := 1 // counter for available input sources. 1 at the beginning stands for the client
	quitChan, wsoiC := ctrl.qd.QuitChan(), ctrl.wsoi.C()
	for ctr > 0 || dqc != nil {
		select {
		case <-quitChan:
			return
		case <-wsoiC:
			wsoiC = nil // disable this channel to avoid receiving twice
			ctr--
		case mjs := <-ctrl.ic:
			if len(mjs) > 0 {
				ctrl.jq.Enqueue(mjs...)
			}
		case mjs := <-ctrl.eqc:
			ctr--
			if len(mjs) > 0 {
				ctrl.jq.Enqueue(mjs...)
			}
		case dqc <- job:
			ctr++
			if ctrl.jq.Len() > 0 {
				job = ctrl.jq.Dequeue()
				continue
			} else {
				dqc = nil // disable dqc
			}
		}
		if dqc == nil && ctrl.jq.Len() > 0 {
			job = ctrl.jq.Dequeue()
			dqc = ctrl.dqc // enable dqc
		}
	}
}

// inputBeforeLaunch inputs metaJobs before the first call to the method Launch.
//
// It returns true if metaJobs are put into the job queue successfully.
// When it returns false, the caller should then send metaJobs to ctrl.ic.
func (ctrl *controller[Job, Properties, Feedback]) inputBeforeLaunch(metaJobs []*MetaJob[Job, Properties]) bool {
	ctrl.m.Lock()
	defer ctrl.m.Unlock()
	if ctrl.lsi {
		return false
	}
	ctrl.jq.Enqueue(metaJobs...)
	return true
}

// copyMetaJobs copies metaJobs,
// replaces the nil items with zero-value items
// (with the field Job to its zero value, Meta.Priority to 0,
// Meta.CreationTime to time.Now(), and Meta.Custom to its zero value),
// and replaces the zero-value Meta.CreationTime field with time.Now().
func copyMetaJobs[Job, Properties any](metaJobs []*MetaJob[Job, Properties]) []*MetaJob[Job, Properties] {
	if metaJobs == nil {
		return nil
	}
	mjs := make([]*MetaJob[Job, Properties], 0, len(metaJobs))
	var now time.Time // lazy init
	for _, mj := range metaJobs {
		newMj := new(MetaJob[Job, Properties])
		if mj != nil {
			*newMj = *mj
		}
		if newMj.Meta.CreationTime.IsZero() {
			if now.IsZero() {
				now = time.Now()
			}
			newMj.Meta.CreationTime = now
		}
		mjs = append(mjs, newMj)
	}
	return mjs
}
