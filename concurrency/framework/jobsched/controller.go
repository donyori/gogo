// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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
	"github.com/donyori/gogo/concurrency/framework/internal"
	"github.com/donyori/gogo/errors"
)

// Controller is a controller for this jobsched framework.
//
// It is used to launch, quit, and wait for the job.
// And it is also used to input new jobs.
type Controller interface {
	framework.Controller

	// Input inputs new jobs.
	//
	// If an item in jobs is nil, it will be treated as a Job with Data to nil,
	// Pri to 0, Ct to time.Now(), and CustAttr to nil.
	// If an item in jobs has the Ct field with a zero value, its Ct field will
	// be set to time.Now().
	//
	// The client can input new jobs before the effective first all of
	// the method Wait (i.e., the call after invoking the method Launch).
	// After calling the method Wait, Input will do nothing and return false.
	// Note that the method Run will call Wait inside it.
	//
	// It returns true if the input succeeds, otherwise (e.g., the job has quit,
	// or the method Wait has been called effectively), false.
	Input(jobs ...*Job) bool
}

// JobHandler is a function to process a job.
//
// The first argument jobData is the data of the job.
//
// The second argument quitDevice is to acquire the channel for the quit signal,
// detect the quit signal, and broadcast a quit signal to quit the job.
//
// It returns a list of new jobs generated during or after processing
// the current job, called newJobs.
// If an item in newJobs is nil, it will be treated as a Job with Data to nil,
// Pri to 0, Ct to time.Now(), and CustAttr to nil.
// If an item in newJobs has the Ct field with a zero value, its Ct field will
// be set to time.Now().
type JobHandler func(jobData interface{}, quitDevice framework.QuitDevice) (newJobs []*Job)

// New creates a new Controller.
//
// n is the number of goroutines to process jobs.
// If n is non-positive, runtime.NumCPU() will be used instead.
//
// handler is the job handler.
// It will panic if handler is nil.
//
// jobQueueMaker is a maker to create a new job queue.
// It enables the client to make custom JobQueue.
// If jobQueueMaker is nil, a default job queue maker will be used.
// The default job queue implements a starvation-free scheduling algorithm.
//
// The rest arguments are initial jobs to process.
func New(n int, handler JobHandler, jobQueueMaker JobQueueMaker, jobs ...*Job) Controller {
	if handler == nil {
		panic(errors.AutoMsg("handler is nil"))
	}
	if n <= 0 {
		n = runtime.NumCPU()
	}
	if jobQueueMaker == nil {
		jobQueueMaker = &DefaultJobQueueMaker{n}
	}
	jq := jobQueueMaker.New()
	jq.Enqueue(jobs...)
	return &controller{
		n:    n,
		h:    handler,
		qd:   internal.NewQuitDevice(),
		jq:   jq,
		ic:   make(chan []*Job),
		eqc:  make(chan []*Job),
		dqc:  make(chan interface{}),
		loi:  concurrency.NewOnceIndicator(),
		wsoi: concurrency.NewOnceIndicator(),
	}
}

// Run creates a Controller with specified parameters, and then run it.
// It returns the panic records of the Controller.
//
// The arguments are the same as those of function New.
func Run(n int, handler JobHandler, jobQueueMaker JobQueueMaker, jobs ...*Job) []framework.PanicRec {
	ctrl := New(n, handler, jobQueueMaker, jobs...)
	ctrl.Run()
	return ctrl.PanicRecords()
}

// controller is an implementation of interface Controller.
type controller struct {
	n int        // The number of goroutines to process jobs.
	h JobHandler // Job handler.

	qd framework.QuitDevice // Quit device.
	jq JobQueue             // Job queue.

	ic  chan []*Job      // Input channel, to input jobs from the client.
	eqc chan []*Job      // Enqueue channel, to input jobs from workers.
	dqc chan interface{} // Dequeue channel, to output job data to workers.

	pr   framework.PanicRecords    // Panic records.
	wg   sync.WaitGroup            // Wait group for the main process.
	loi  concurrency.OnceIndicator // For launching the job.
	wsoi concurrency.OnceIndicator // For indicating the start of the effective first call of the method Wait.
	m    sync.Mutex                // Lock to avoid the race condition on jq when calling Launch and Input at the same time.
	lsi  bool                      // An indicator to report whether the method Launch is started or not.
}

// QuitChan returns the channel for the quit signal.
// When the job is finished or quit, this channel will be closed
// to broadcast the quit signal.
func (ctrl *controller) QuitChan() <-chan struct{} {
	return ctrl.qd.QuitChan()
}

// IsQuit detects the quit signal on the quit channel.
// It returns true if a quit signal is detected, and false otherwise.
func (ctrl *controller) IsQuit() bool {
	return ctrl.qd.IsQuit()
}

// Quit broadcasts a quit signal to quit the job.
//
// This method will NOT wait until the job ends.
func (ctrl *controller) Quit() {
	ctrl.qd.Quit()
}

// Launch starts the job.
//
// This method will NOT wait until the job ends.
// Use method Wait if you want to wait for that.
//
// Note that Launch can take effect only once.
// To do the same job again, create a new Controller
// with the same parameters.
func (ctrl *controller) Launch() {
	ctrl.loi.Do(func() {
		ctrl.m.Lock()
		defer ctrl.m.Unlock()
		ctrl.lsi = true

		ctrl.wg.Add(ctrl.n + 1) // n workers and 1 allocator.
		for i := 0; i < ctrl.n; i++ {
			go func(rank int) {
				defer func() {
					if r := recover(); r != nil {
						ctrl.qd.Quit()
						ctrl.pr.Append(framework.PanicRec{
							Name:    strconv.Itoa(rank),
							Content: r,
						})
					}
					ctrl.wg.Done()
				}()
				ctrl.workerProc()
			}(i)
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					ctrl.qd.Quit()
					ctrl.pr.Append(framework.PanicRec{
						Name:    "allocator",
						Content: r,
					})
				}
				ctrl.wg.Done()
			}()
			ctrl.allocatorProc()
		}()
	})
}

// Wait waits for the job to finish or quit.
// It returns the number of panic goroutines.
//
// If the job was not launched, it does nothing and returns -1.
func (ctrl *controller) Wait() int {
	if !ctrl.loi.Test() {
		return -1
	}
	defer ctrl.qd.Quit() // For cleanup possible daemon goroutines that wait for a quit signal to exit.
	ctrl.wsoi.Do(nil)
	ctrl.wg.Wait()
	return ctrl.pr.Len()
}

// Run launches the job and waits for it.
// It returns the number of panic goroutines.
func (ctrl *controller) Run() int {
	ctrl.Launch()
	return ctrl.Wait()
}

// NumGoroutine returns the number of goroutines to process this job.
//
// Note that it only includes the main goroutines to process the job.
// Any possible control goroutines, daemon goroutines, auxiliary goroutines,
// or the goroutines launched in the client's business functions
// are all excluded.
func (ctrl *controller) NumGoroutine() int {
	return ctrl.n
}

// PanicRecords returns the panic records.
func (ctrl *controller) PanicRecords() []framework.PanicRec {
	return ctrl.pr.List()
}

// Input inputs new jobs.
//
// If an item in jobs is nil, it will be treated as a Job with Data to nil,
// Pri to 0, Ct to time.Now(), and CustAttr to nil.
// If an item in jobs has the Ct field with a zero value, its Ct field will
// be set to time.Now().
//
// The client can input new jobs before the effective first all of
// the method Wait (i.e., the call after invoking the method Launch).
// After calling the method Wait, Input will do nothing and return false.
// Note that the method Run will call Wait inside it.
//
// It returns true if the input succeeds, otherwise (e.g., the job has quit,
// or the method Wait has been called effectively), false.
func (ctrl *controller) Input(jobs ...*Job) bool {
	if ctrl.wsoi.Test() {
		return false
	}
	var now time.Time
	for i, job := range jobs {
		if job == nil {
			job = new(Job)
			jobs[i] = job
		}
		if job.Ct.IsZero() {
			if now.IsZero() {
				now = time.Now()
			}
			job.Ct = now
		}
	}
	if !ctrl.loi.Test() && ctrl.inputBeforeLaunch(jobs) {
		return true
	}
	select {
	case <-ctrl.qd.QuitChan():
		return false
	case ctrl.ic <- jobs:
		return true
	}
}

// inputBeforeLaunch inputs new jobs before the first call of the method Launch.
//
// It returns true if jobs are put into the job queue successfully.
// When it returns false, the caller should then send jobs to ctrl.ic.
func (ctrl *controller) inputBeforeLaunch(jobs []*Job) bool {
	ctrl.m.Lock()
	defer ctrl.m.Unlock()
	if ctrl.lsi {
		return false
	}
	ctrl.jq.Enqueue(jobs...)
	return true
}

// allocatorProc is the allocator main process,
// without panic checking and wg.Done().
func (ctrl *controller) allocatorProc() {
	defer close(ctrl.dqc)
	var dqc chan<- interface{} // Disable dqc at the beginning.
	var jobData interface{}
	if ctrl.jq.Len() > 0 {
		jobData = ctrl.jq.Dequeue()
		dqc = ctrl.dqc // Enable dqc.
	}
	ctr := 1 // Counter for available input sources. 1 at the beginning stands for the client.
	quitChan, wsoiC := ctrl.qd.QuitChan(), ctrl.wsoi.C()
	for ctr > 0 || dqc != nil {
		select {
		case <-quitChan:
			return
		case <-wsoiC:
			wsoiC = nil // Disable this channel to avoid receiving twice.
			ctr--
		case jobs := <-ctrl.ic:
			if len(jobs) > 0 {
				ctrl.jq.Enqueue(jobs...)
			}
		case jobs := <-ctrl.eqc:
			ctr--
			if len(jobs) > 0 {
				ctrl.jq.Enqueue(jobs...)
			}
		case dqc <- jobData:
			ctr++
			if ctrl.jq.Len() > 0 {
				jobData = ctrl.jq.Dequeue()
				continue
			} else {
				dqc = nil // Disable dqc.
			}
		}
		if dqc == nil && ctrl.jq.Len() > 0 {
			jobData = ctrl.jq.Dequeue()
			dqc = ctrl.dqc // Enable dqc.
		}
	}
}

// workerProc is the worker main process, without panic checking and wg.Done().
func (ctrl *controller) workerProc() {
	quitChan := ctrl.qd.QuitChan()
	for {
		var jobs []*Job
		select {
		case <-quitChan:
			return
		case jobData, ok := <-ctrl.dqc:
			if !ok {
				return
			}
			jobs = ctrl.h(jobData, ctrl.qd)
			var now time.Time
			for i, job := range jobs {
				if job == nil {
					job = new(Job)
					jobs[i] = job
				}
				if job.Ct.IsZero() {
					if now.IsZero() {
						now = time.Now()
					}
					job.Ct = now
				}
			}
		}

		// Always send jobs to the allocator,
		// regardless of whether jobs is empty or not.
		select {
		case <-quitChan:
			return
		case ctrl.eqc <- jobs:
		}
	}
}
