// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2025  Yuan Gao
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

package concurrency

import "runtime"

// Once is an object that calls a specified function only once,
// based on Go channel.
//
// Moreover, it can indicate whether the function has been called,
// and enable the client on another goroutine to
// wait for the function to finish (use <-Once.C()).
type Once interface {
	// Do calls the specified function if and only if
	// the method Do or DoRecover is being called for
	// the first time for this instance of Once.
	// Otherwise, it does nothing but waits for
	// the first call of the specified function to finish.
	//
	// Because no call to Do or DoRecover returns until
	// the one call to the specified function returns,
	// if the specified function causes Do or DoRecover to be called,
	// it deadlocks.
	//
	// If the specified function panics with a non-nil value,
	// Do panics with the same value on every call.
	//
	// If the specified function panics with nil,
	// or interrupts (e.g., calls runtime.Goexit),
	// subsequent calls to Do call runtime.Goexit
	// to interrupt their goroutines.
	Do()

	// DoRecover is like Do,
	// but it recovers the goroutine if the specified function panics.
	//
	// It returns an indicator to report
	// whether the specified function is called in this call to DoRecover,
	// and returns the panic value of the specified function (nil if no panic).
	//
	// In particular, if the specified function panics with nil,
	// or interrupts (e.g., calls runtime.Goexit), like Do,
	// subsequent calls to DoRecover call runtime.Goexit
	// to interrupt their goroutines.
	DoRecover() (called bool, panicValue any)

	// C returns a channel that will be closed after
	// the first call to the specified function returns.
	C() <-chan struct{}

	// Done reports whether the first call to
	// the specified function has been finished.
	Done() bool

	// PanicValue returns two values.
	//
	// The first one is an indicator to report whether
	// the specified function has been called but interrupted
	// (such as panicking, calling runtime.Goexit, and so on).
	//
	// The second one is the panic value of the specified function.
	// It is nil if the function does not panic.
	//
	// In particular, if the specified function has not been called or finished,
	// PanicValue returns (false, nil).
	//
	// If PanicValue returns (true, nil),
	// subsequent calls to Do or DoRecover will call runtime.Goexit
	// to interrupt their goroutines.
	PanicValue() (interrupted bool, panicValue any)
}

// NewOnce creates a new Once that calls the specified function f only once.
//
// f can be nil, which is considered a function that does nothing.
func NewOnce(f func()) Once {
	firstC := make(chan struct{}, 1)
	firstC <- struct{}{}
	close(firstC)
	return &once{
		doneC:       make(chan struct{}),
		firstC:      firstC,
		f:           f,
		interrupted: true,
	}
}

// once is an implementation of interface Once.
type once struct {
	// Channel to broadcast the finish signal.
	doneC chan struct{}

	// Channel to determine whether the current call is
	// the first call to the method Do or DoRecover.
	//
	// Usage:
	//
	//	// in the initialization of the struct once
	//	c := make(chan struct{}, 1)
	//	c <- struct{}{}
	//	close(c)
	//	o.firstC = c
	//
	//	// in the body of methods Do and DoRecover
	//	_, ok := <-o.firstC
	//	// ok indicates whether the current call is the first
	firstC <-chan struct{}

	// The specified function that is called only once.
	f func()

	// An indicator to report whether the specified function is interrupted,
	// such as panicking, calling runtime.Goexit, and so on.
	//
	// It is set to true before calling the specified function.
	interrupted bool

	// Panic value of the specified function (nil if no panic).
	panicValue any
}

func (o *once) Do() {
	_, ok := <-o.firstC
	if ok {
		defer close(o.doneC)
		defer func() {
			o.panicValue = recover()
			if o.panicValue != nil {
				// Re-panic immediately so on the first call
				// the client gets a complete stack trace into o.f.
				panic(o.panicValue)
			}
		}()
		if o.f != nil {
			o.f()
		}
		o.interrupted = false
	} else {
		<-o.doneC // wait for the first call to finish
		if o.panicValue != nil {
			panic(o.panicValue)
		} else if o.interrupted {
			runtime.Goexit()
		}
	}
}

func (o *once) DoRecover() (called bool, panicValue any) {
	_, called = <-o.firstC
	if called {
		defer close(o.doneC)
		defer func() {
			o.panicValue = recover()
			panicValue = o.panicValue
		}()
		if o.f != nil {
			o.f()
		}
		o.interrupted = false
	} else {
		<-o.doneC // wait for the first call to finish
		panicValue = o.panicValue
		if o.panicValue == nil && o.interrupted {
			runtime.Goexit()
		}
	}
	return
}

func (o *once) C() <-chan struct{} {
	return o.doneC
}

func (o *once) Done() bool {
	select {
	case <-o.doneC:
		return true
	default:
		return false
	}
}

func (o *once) PanicValue() (interrupted bool, panicValue any) {
	select {
	case <-o.doneC:
		interrupted, panicValue = o.interrupted, o.panicValue
	default:
	}
	return
}
