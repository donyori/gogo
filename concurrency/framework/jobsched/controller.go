// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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
	"github.com/donyori/gogo/errors"
)

// Controller is a controller for this job scheduling framework.
//
// It is used to launch, cancel, and wait for the job.
// Also, it is used to input new jobs.
//
// The first type parameter Job is the type of jobs.
// The second type parameter Properties is the type of custom properties
// in the meta information of jobs.
// The third type parameter Feedback is the type of feedback on the jobs,
// which is collected and handled in a dedicated goroutine.
type Controller[Job, Properties, Feedback any] interface {
	framework.Controller

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
	// It is safe for concurrent use by multiple goroutines.
	//
	// The client can input new jobs before the first effective call to
	// the method Wait (i.e., the call after invoking the method Launch).
	// After calling the method Wait, Input does nothing and returns 0.
	// Note that the method Run calls Wait inside.
	Input(metaJob ...*MetaJob[Job, Properties]) int
}

// NoFeedback is a special case of feedback type,
// indicating that the job handler has no feedback.
// In this case, the framework skips the feedback handler.
type NoFeedback struct{}

// noFeedbackType is the reflect.Type of NoFeedback.
var noFeedbackType = reflect.TypeFor[NoFeedback]()

// JobHandler is a function to process a job.
//
// The first type parameter Job is the type of jobs.
// The second type parameter Properties is the type of custom properties
// in the meta information of jobs.
// The third type parameter Feedback is the type of feedback on the jobs,
// which is collected and handled in a dedicated goroutine.
//
// The first parameter is a canceler to interrupt all job processors.
// The second parameter is the rank of the worker goroutine
// (from 0 to ctrl.NumGoroutine()-1) to identify the goroutine uniquely.
// The third parameter is the job to be processed.
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
// which is collected and handled in a dedicated goroutine.
// In particular, if the type of feedback is NoFeedback,
// the framework skips the feedback handler.
type JobHandler[Job, Properties, Feedback any] func(
	canceler concurrency.Canceler,
	rank int,
	job Job,
) (newJobs []*MetaJob[Job, Properties], feedback Feedback)

// FeedbackHandler is a function to process feedback.
//
// The first parameter is a canceler to interrupt all job processors.
// The second parameter is the feedback to be processed.
type FeedbackHandler[Feedback any] func(
	canceler concurrency.Canceler,
	feedback Feedback,
)

// Options are options for creating Controller.
type Options[Job, Properties, Feedback any] struct {
	// The number of worker goroutines to process jobs.
	// Nonpositive values for using max(1, runtime.NumCPU()-2).
	NumWorker int

	// The maker to create a new job queue.
	// It enables the client to make custom JobQueue.
	// If it is nil, an FCFS (first come, first served) job queue is used.
	JobQueueMaker JobQueueMaker[Job, Properties]

	// The buffer size of the feedback channel.
	// Nonpositive values for no buffer.
	// Only take effect when the type of feedback is not NoFeedback.
	FeedbackChanBufSize int

	// The setup function called by each worker goroutine that processes jobs.
	//
	// If it is not nil, each worker goroutine calls it
	// when the goroutine starts.
	//
	// Its first parameter is the controller.
	// Its second parameter is the rank of the worker goroutine
	// (from 0 to ctrl.NumGoroutine()-1) to identify the goroutine uniquely.
	//
	// The client is responsible for guaranteeing that
	// this function is safe for concurrency.
	Setup func(ctrl Controller[Job, Properties, Feedback], rank int)

	// The cleanup function called by each worker goroutine that processes jobs.
	//
	// If this cleanup function is not nil,
	// and the goroutine has successfully executed the setup function
	// (if the setup function is not nil),
	// then each worker goroutine calls this cleanup function
	// before the goroutine ends (even if the goroutine panics).
	//
	// Its first parameter is the controller.
	// Its second parameter is the rank of the worker goroutine
	// (from 0 to ctrl.NumGoroutine()-1) to identify the goroutine uniquely.
	//
	// The client is responsible for guaranteeing that
	// this function is safe for concurrency.
	Cleanup func(ctrl Controller[Job, Properties, Feedback], rank int)
}

// New creates a new Controller with options opts.
//
// The first type parameter Job is the type of jobs.
// The second type parameter Properties is the type of custom properties
// in the meta information of jobs.
// The third type parameter Feedback is the type of feedback on the jobs,
// which is collected and handled in a dedicated goroutine.
//
// jobHandler is the job handler.
// New panics if jobHandler is nil.
//
// feedbackHandler is the feedback handler.
// If feedbackHandler is nil, all feedback is discarded silently.
// If the type of feedback is NoFeedback,
// the framework never calls feedbackHandler.
//
// If opts are nil, a zero-value Options is used.
//
// metaJob is initial jobs to process.
// If an item in metaJob is nil, it is treated as a zero-value job
// (with the field Job to its zero value, Meta.Priority to 0,
// Meta.CreationTime to time.Now(), and Meta.Custom to its zero value).
// If an item in metaJob has the field Meta.CreationTime with a zero value,
// this field is set to time.Now() by the framework.
func New[Job, Properties, Feedback any](
	jobHandler JobHandler[Job, Properties, Feedback],
	feedbackHandler FeedbackHandler[Feedback],
	opts *Options[Job, Properties, Feedback],
	metaJob ...*MetaJob[Job, Properties],
) Controller[Job, Properties, Feedback] {
	if jobHandler == nil {
		panic(errors.AutoMsg("job handler is nil"))
	} else if opts == nil {
		opts = new(Options[Job, Properties, Feedback])
	}
	n := opts.NumWorker
	if n <= 0 {
		n = runtime.NumCPU() - 2
		if n < 1 {
			n = 1
		}
	}
	var jq JobQueue[Job, Properties]
	if opts.JobQueueMaker != nil {
		jq = opts.JobQueueMaker.New()
	} else {
		jq = new(fcfsJobQueue[Job, Properties])
	}
	if len(metaJob) > 0 {
		jq.Enqueue(copyMetaJobs(metaJob)...)
	}
	ctrl := &controller[Job, Properties, Feedback]{
		n:       n,
		jh:      jobHandler,
		fh:      feedbackHandler,
		c:       concurrency.NewCanceler(),
		jq:      jq,
		ic:      make(chan []*MetaJob[Job, Properties], 1),
		eqc:     make(chan []*MetaJob[Job, Properties], n),
		dqc:     make(chan Job, 1),
		pr:      concurrency.NewRecorder[framework.PanicRecord](0),
		wso:     concurrency.NewOnce(nil),
		setup:   opts.Setup,
		cleanup: opts.Cleanup,
	}
	ctrl.lo = concurrency.NewOnce(ctrl.launchProc)
	if reflect.TypeFor[Feedback]() != noFeedbackType {
		bufSize := opts.FeedbackChanBufSize
		if bufSize < 0 {
			bufSize = 0
		}
		ctrl.fc = make(chan Feedback, bufSize)
		ctrl.fhdc = make(chan struct{})
	}
	return ctrl
}

// NewWithoutFeedback is like New but sets the type of feedback to NoFeedback.
//
// The parameters are similar to those of function New,
// except that NewWithoutFeedback does not require feedbackHandler.
//
// It is equivalent to
// New[Job, Properties, NoFeedback](jobHandler, nil, opts, metaJob...).
func NewWithoutFeedback[Job, Properties any](
	jobHandler JobHandler[Job, Properties, NoFeedback],
	opts *Options[Job, Properties, NoFeedback],
	metaJob ...*MetaJob[Job, Properties],
) Controller[Job, Properties, NoFeedback] {
	return New(jobHandler, nil, opts, metaJob...)
}

// Run creates a Controller with specified arguments, and then runs it.
// It returns the panic records of the Controller.
//
// The parameters are the same as those of function New.
func Run[Job, Properties, Feedback any](
	jobHandler JobHandler[Job, Properties, Feedback],
	feedbackHandler FeedbackHandler[Feedback],
	opts *Options[Job, Properties, Feedback],
	metaJob ...*MetaJob[Job, Properties],
) []framework.PanicRecord {
	ctrl := New(jobHandler, feedbackHandler, opts, metaJob...)
	ctrl.Run()
	return ctrl.PanicRecords()
}

// RunWithoutFeedback creates a Controller[Job, Properties, NoFeedback]
// with specified arguments, and then runs it.
// It returns the panic records of the Controller.
//
// The parameters are the same as those of function NewWithoutFeedback.
//
// It is equivalent to
// Run[Job, Properties, NoFeedback](jobHandler, nil, opts, metaJob...).
func RunWithoutFeedback[Job, Properties any](
	jobHandler JobHandler[Job, Properties, NoFeedback],
	opts *Options[Job, Properties, NoFeedback],
	metaJob ...*MetaJob[Job, Properties],
) []framework.PanicRecord {
	return Run(jobHandler, nil, opts, metaJob...)
}

// controller is an implementation of interface Controller.
type controller[Job, Properties, Feedback any] struct {
	n  int                                   // The number of worker goroutines to process jobs.
	jh JobHandler[Job, Properties, Feedback] // Job handler.
	fh FeedbackHandler[Feedback]             // Feedback handler.

	c  concurrency.Canceler      // Canceler.
	jq JobQueue[Job, Properties] // Job queue.

	ic   chan []*MetaJob[Job, Properties] // Input channel, to input jobs from the client.
	eqc  chan []*MetaJob[Job, Properties] // Enqueue channel, to input jobs from workers.
	dqc  chan Job                         // Dequeue channel, to dispatch jobs to workers.
	fc   chan Feedback                    // Feedback channel, to collect feedback on jobs.
	fhdc chan struct{}                    // Feedback handler done channel, to broadcast a signal when the feedback handler is finished.

	pr  concurrency.Recorder[framework.PanicRecord] // Panic recorder.
	wg  sync.WaitGroup                              // Wait group for the workers and the job allocator, not for the feedback handler.
	lo  concurrency.Once                            // For launching the framework.
	wso concurrency.Once                            // For indicating the start of the first effective call to the method Wait.
	// Lock to avoid the race condition on jq
	// when calling Launch and Input at the same time
	// or calling Input simultaneously.
	m   sync.Mutex
	lsi bool // An indicator to report whether the method Launch is started.

	setup   func(ctrl Controller[Job, Properties, Feedback], rank int) // Worker setup function.
	cleanup func(ctrl Controller[Job, Properties, Feedback], rank int) // Worker cleanup function.
}

func (ctrl *controller[Job, Properties, Feedback]) Canceler() concurrency.Canceler {
	return ctrl.c
}

func (ctrl *controller[Job, Properties, Feedback]) Launch() {
	ctrl.lo.Do()
}

func (ctrl *controller[Job, Properties, Feedback]) Wait() int {
	if !ctrl.lo.Done() {
		return -1
	}
	defer ctrl.c.Cancel() // for cleanup possible daemon goroutines that wait for a cancellation signal to exit
	ctrl.wso.Do()
	ctrl.wg.Wait()
	if ctrl.fhdc != nil {
		<-ctrl.fhdc
	}
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
	return ctrl.pr.All()
}

func (ctrl *controller[Job, Properties, Feedback]) Input(
	metaJob ...*MetaJob[Job, Properties]) int {
	if ctrl.wso.Done() {
		return 0
	}
	mjs := copyMetaJobs(metaJob)
	if !ctrl.lo.Done() && ctrl.inputBeforeLaunch(mjs) {
		return len(mjs)
	}
	select {
	case <-ctrl.c.C():
		return 0
	case ctrl.ic <- mjs:
		return len(mjs)
	}
}

// launchProc is the process of starting the job.
// It is invoked by ctrl.lo.Do.
func (ctrl *controller[Job, Properties, Feedback]) launchProc() {
	ctrl.m.Lock()
	defer ctrl.m.Unlock()
	ctrl.lsi = true

	ctrl.wg.Add(ctrl.n + 1) // n workers + 1 job allocator
	if ctrl.fc != nil {
		go func() { // goroutine for feedback handler
			defer close(ctrl.fhdc)
			defer func() {
				if e := recover(); e != nil {
					ctrl.c.Cancel()
					ctrl.pr.Record(framework.PanicRecord{
						Name:    "feedback handler",
						Content: e,
					})
				}
			}()
			ctrl.feedbackHandlerProc()
		}()
		go func() { // goroutine for closing feedback channel
			defer close(ctrl.fc)
			defer func() {
				if e := recover(); e != nil {
					ctrl.c.Cancel()
					ctrl.pr.Record(framework.PanicRecord{
						Name:    "feedback channel closer",
						Content: e,
					})
				}
			}()
			ctrl.wg.Wait()
		}()
	}
	for i := range ctrl.n {
		go func(rank int) { // goroutine for worker
			defer ctrl.wg.Done()
			defer func() {
				if e := recover(); e != nil {
					ctrl.c.Cancel()
					ctrl.pr.Record(framework.PanicRecord{
						Name:    "worker " + strconv.Itoa(rank),
						Content: e,
					})
				}
			}()
			ctrl.workerProc(rank)
		}(i)
	}
	go func() { // goroutine for job allocator
		defer ctrl.wg.Done()
		defer func() {
			if e := recover(); e != nil {
				ctrl.c.Cancel()
				ctrl.pr.Record(framework.PanicRecord{
					Name:    "job allocator",
					Content: e,
				})
			}
		}()
		ctrl.jobAllocatorProc()
	}()
}

// feedbackHandlerProc is the feedback handler main process,
// without panic checking and close(ctrl.fhdc).
func (ctrl *controller[Job, Properties, Feedback]) feedbackHandlerProc() {
	// The feedback handler must handle all feedback
	// returned by the job handlers.
	// Therefore, this function does not listen to ctrl.c.C().
	switch {
	case ctrl.fc == nil:
		// This should never happen, but will act as a safeguard for later.
		return
	case ctrl.fh != nil:
		for fb := range ctrl.fc {
			ctrl.fh(ctrl.c, fb)
		}
	default:
		for range ctrl.fc {
		}
	}
}

// workerProc is the worker main process,
// without panic checking and ctrl.wg.Done().
func (ctrl *controller[Job, Properties, Feedback]) workerProc(rank int) {
	if ctrl.setup != nil {
		ctrl.setup(ctrl, rank)
	}
	if ctrl.cleanup != nil {
		defer ctrl.cleanup(ctrl, rank)
	}
	cancelChan := ctrl.c.C()
	for {
		var mjs []*MetaJob[Job, Properties]
		var fb Feedback
		select {
		case <-cancelChan:
			return
		case job, ok := <-ctrl.dqc:
			if !ok {
				return
			}
			mjs, fb = ctrl.jh(ctrl.c, rank, job)
			mjs = copyMetaJobs(mjs)
		}
		if ctrl.fc != nil {
			// The feedback type is not NoFeedback.
			// Send feedback first.
			select {
			case <-cancelChan:
				return
			case ctrl.fc <- fb:
			}
		}
		// Always send new jobs to the job allocator,
		// regardless of whether jobs are empty.
		select {
		case <-cancelChan:
			return
		case ctrl.eqc <- mjs:
		}
	}
}

// jobAllocatorProc is the job allocator main process,
// without panic checking and ctrl.wg.Done().
func (ctrl *controller[Job, Properties, Feedback]) jobAllocatorProc() {
	defer close(ctrl.dqc)
	var dqc chan<- Job // disable dqc at the beginning
	var job Job
	if ctrl.jq.Len() > 0 {
		job = ctrl.jq.Dequeue()
		dqc = ctrl.dqc // enable dqc
	}
	ctr := 1 // counter for available input sources. 1 at the beginning stands for the client
	cancelChan, wsoC := ctrl.c.C(), ctrl.wso.C()
	for ctr > 0 || len(ctrl.ic) > 0 || dqc != nil {
		select {
		case <-cancelChan:
			return
		case <-wsoC:
			wsoC = nil // disable this channel to avoid receiving twice
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
func (ctrl *controller[Job, Properties, Feedback]) inputBeforeLaunch(
	metaJobs []*MetaJob[Job, Properties]) bool {
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
func copyMetaJobs[Job, Properties any](
	metaJobs []*MetaJob[Job, Properties]) []*MetaJob[Job, Properties] {
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
