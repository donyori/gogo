// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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

package spmd

import (
	"runtime"
	"sync"

	"github.com/donyori/gogo/errors"
)

// A controller to start, quit and wait the job,
// and also the context of the job.
type World interface {
	// Start the job.
	//
	// This method will NOT wait until the job ends.
	// Use method Wait if you want to wait for that.
	//
	// Note that Start can affect only once.
	// To do the same job again, create a new World with the same parameters.
	Start()

	// Quit the job.
	//
	// This method will NOT wait until the job ends.
	// Use method Wait if you want to wait for that.
	Quit()

	// Wait for the job to finish or quit.
	// It returns the number of panic goroutines.
	Wait() int

	// Start the job and wait for it.
	// It returns the number of panic goroutines.
	Run() int

	// Return the number of goroutines to process this job.
	NumGoroutine() int

	// Return the panic records.
	PanicRecords() []PanicRec
}

// Create a World for a new job.
//
// n is the number of goroutines to process the job.
// If n is non-positive, runtime.NumCPU() will be used instead.
//
// biz is the business handler for the job.
// It will panic if biz is nil.
func New(n int, biz func(comm Communicator)) World {
	if biz == nil {
		panic(errors.AutoMsg("biz is nil"))
	}
	if n <= 0 {
		n = runtime.NumCPU()
	}
	w := &world{
		QuitChan: make(chan struct{}),
		Comms:    make([]*communicator, n),
		n:        n,
		biz:      biz,
		pr:       new(panicRecords),
	}
	for i := 0; i < n; i++ {
		w.Comms[i] = newCommunicator(w, i)
	}
	return w
}

// Create a World with n and biz, and then run it.
// It returns the panic records of the World.
//
// The arguments n and biz are the same as those of function New.
func Run(n int, biz func(comm Communicator)) []PanicRec {
	w := New(n, biz)
	w.Run()
	return w.PanicRecords()
}

// An implementation of interface World.
type world struct {
	QuitChan chan struct{}   // A channel to broadcast the quit signal.
	Comms    []*communicator // List of communicators.

	n         int                     // The number of goroutines to process this job.
	biz       func(comm Communicator) // Business function.
	pr        *panicRecords           // Panic records.
	wg        sync.WaitGroup          // A wait group for the main process.
	startOnce sync.Once               // For starting the job.
	quitOnce  sync.Once               // For closing QuitChan.
}

func (w *world) Start() {
	w.startOnce.Do(func() {
		w.wg.Add(w.n)
		for i := 0; i < w.n; i++ {
			go func(rank int) {
				defer func() {
					if r := recover(); r != nil {
						w.Quit()
						w.pr.Append(PanicRec{
							Rank:    rank,
							Content: r,
						})
					}
					w.wg.Done()
				}()
				w.biz(w.Comms[rank])
			}(i)
		}
	})
}

func (w *world) Quit() {
	w.quitOnce.Do(func() {
		close(w.QuitChan)
	})
}

func (w *world) Wait() int {
	w.wg.Wait()
	return w.pr.Len()
}

func (w *world) Run() int {
	w.Start()
	return w.Wait()
}

func (w *world) NumGoroutine() int {
	return w.n
}

func (w *world) PanicRecords() []PanicRec {
	return w.pr.List()
}
