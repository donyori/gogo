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
type Controller[Job, Properties any] interface {
	framework.Controller

	// Input enables the client to input new jobs.
	//
	// If an item in metaJob is nil, it will be treated as a zero-value job
	// (with the field Job to its zero value, Meta.Priority to 0,
	// Meta.CreationTime to time.Now(), and Meta.Custom to its zero value).
	// If an item in metaJob has the field Meta.CreationTime with a zero value,
	// this field will be set to time.Now().
	//
	// It returns the number of jobs input successfully.
	//
	// The client can input new jobs before the first effective call to
	// the method Wait (i.e., the call after invoking the method Launch).
	// After calling the method Wait, Input will do nothing and return 0.
	// Note that the method Run will call Wait inside.
	Input(metaJob ...*MetaJob[Job, Properties]) int
}

// JobHandler is a function to process a job.
//
// The first type parameter Job is the type of jobs.
// The second type parameter Properties is the type of custom properties
// in the meta information of jobs.
//
// The first parameter is the job to be processed.
//
// The second parameter is a device to acquire the channel for the quit signal,
// detect the quit signal, and broadcast a quit signal to interrupt
// all job processors.
//
// It returns new jobs generated during or after processing
// the current job, called newJobs.
// If an item in newJobs is nil, it will be treated as a zero-value job
// (with the field Job to its zero value, Meta.Priority to 0,
// Meta.CreationTime to time.Now(), and Meta.Custom to its zero value).
// If an item in newJobs has the field Meta.CreationTime with a zero value,
// this field will be set to time.Now().
type JobHandler[Job, Properties any] func(job Job, quitDevice framework.QuitDevice) (newJobs []*MetaJob[Job, Properties])

// New creates a new Controller.
//
// The first type parameter Job is the type of jobs.
// The second type parameter Properties is the type of custom properties
// in the meta information of jobs.
//
// n is the number of goroutines to process jobs.
// If n is non-positive, runtime.NumCPU() will be used instead.
//
// handler is the job handler.
// It panics if handler is nil.
//
// jobQueueMaker is a maker to create a new job queue.
// It enables the client to make custom JobQueue.
// It panics if jobQueueMaker is nil.
//
// metaJob is initial jobs to process.
func New[Job, Properties any](n int, handler JobHandler[Job, Properties],
	jobQueueMaker JobQueueMaker[Job, Properties],
	metaJob ...*MetaJob[Job, Properties]) Controller[Job, Properties] {
	if handler == nil {
		panic(errors.AutoMsg("handler is nil"))
	}
	if jobQueueMaker == nil {
		panic(errors.AutoMsg("job queue maker is nil"))
	}
	if n <= 0 {
		n = runtime.NumCPU()
	}
	jq := jobQueueMaker.New()
	if len(metaJob) > 0 {
		jq.Enqueue(copyMetaJobs(metaJob)...)
	}
	return &controller[Job, Properties]{
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
}

// Run creates a Controller with specified arguments, and then run it.
// It returns the panic records of the Controller.
//
// The parameters are the same as those of function New.
func Run[Job, Properties any](n int, handler JobHandler[Job, Properties],
	jobQueueMaker JobQueueMaker[Job, Properties],
	metaJob ...*MetaJob[Job, Properties]) []framework.PanicRecord {
	ctrl := New(n, handler, jobQueueMaker, metaJob...)
	ctrl.Run()
	return ctrl.PanicRecords()
}

// controller is an implementation of interface Controller.
type controller[Job, Properties any] struct {
	n int                         // The number of goroutines to process jobs.
	h JobHandler[Job, Properties] // Job handler.

	qd framework.QuitDevice      // Quit device.
	jq JobQueue[Job, Properties] // Job queue.

	ic  chan []*MetaJob[Job, Properties] // Input channel, to input jobs from the client.
	eqc chan []*MetaJob[Job, Properties] // Enqueue channel, to input jobs from workers.
	dqc chan Job                         // Dequeue channel, to dispatch jobs to workers.

	pr   framework.PanicRecords    // Panic records.
	wg   sync.WaitGroup            // Wait group for the main process.
	loi  concurrency.OnceIndicator // For launching the framework.
	wsoi concurrency.OnceIndicator // For indicating the start of the first effective call to the method Wait.
	m    sync.Mutex                // Lock to avoid the race condition on jq when calling Launch and Input at the same time.
	lsi  bool                      // An indicator to report whether the method Launch is started.
}

func (ctrl *controller[Job, Properties]) QuitChan() <-chan struct{} {
	return ctrl.qd.QuitChan()
}

func (ctrl *controller[Job, Properties]) IsQuit() bool {
	return ctrl.qd.IsQuit()
}

func (ctrl *controller[Job, Properties]) Quit() {
	ctrl.qd.Quit()
}

func (ctrl *controller[Job, Properties]) Launch() {
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

func (ctrl *controller[Job, Properties]) Wait() int {
	if !ctrl.loi.Test() {
		return -1
	}
	defer ctrl.qd.Quit() // for cleanup possible daemon goroutines that wait for a quit signal to exit
	ctrl.wsoi.Do(nil)
	ctrl.wg.Wait()
	return ctrl.pr.Len()
}

func (ctrl *controller[Job, Properties]) Run() int {
	ctrl.Launch()
	return ctrl.Wait()
}

func (ctrl *controller[Job, Properties]) NumGoroutine() int {
	return ctrl.n
}

func (ctrl *controller[Job, Properties]) PanicRecords() []framework.PanicRecord {
	return ctrl.pr.List()
}

func (ctrl *controller[Job, Properties]) Input(metaJob ...*MetaJob[Job, Properties]) int {
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
func (ctrl *controller[Job, Properties]) workerProc() {
	quitChan := ctrl.qd.QuitChan()
	for {
		var mjs []*MetaJob[Job, Properties]
		select {
		case <-quitChan:
			return
		case job, ok := <-ctrl.dqc:
			if !ok {
				return
			}
			mjs = copyMetaJobs(ctrl.h(job, ctrl.qd))
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
func (ctrl *controller[Job, Properties]) allocatorProc() {
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
func (ctrl *controller[Job, Properties]) inputBeforeLaunch(metaJobs []*MetaJob[Job, Properties]) bool {
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
