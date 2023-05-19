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

package concurrency

// OnceIndicator is an object that performs exactly one action,
// similar to sync.Once, but based on Go channel.
// Moreover, it can indicate whether the action has been performed,
// and enable the client on another goroutine to wait for the action to finish.
type OnceIndicator interface {
	// Do performs similarly to the method Do of sync.Once.
	// Moreover, it indicates whether the function f is called
	// in this invocation.
	//
	// In detail, it calls the function f and returns true if and only if
	// the method Do is being called for the first time for this instance
	// of OnceIndicator.
	// Otherwise, it does nothing but waits for the first call of f to finish,
	// and then returns false.
	//
	// If the client wants to do nothing but trigger this indicator,
	// just set f to nil (no panic happens).
	Do(f func()) bool

	// C returns a channel that will be closed after calling the method Do
	// for the first time for this instance of OnceIndicator.
	C() <-chan struct{}

	// Wait waits for the first call of the method Do for this instance
	// on another goroutine to finish.
	Wait()

	// Test reports whether the method Do for this instance is called.
	//
	// It returns true if and only if the first call of the method Do
	// for this instance has finished.
	Test() bool
}

// NewOnceIndicator creates a new instance of OnceIndicator.
func NewOnceIndicator() OnceIndicator {
	firstC := make(chan struct{}, 1)
	firstC <- struct{}{}
	close(firstC)
	return &onceIndicator{
		firstC: firstC,
		doneC:  make(chan struct{}),
	}
}

// onceIndicator is an implementation of interface OnceIndicator.
type onceIndicator struct {
	// Channel to determine whether the current call is
	// the first call to the method Do.
	//
	// Usage:
	//
	//	// in the initialization of the onceIndicator
	//	c := make(chan struct{}, 1)
	//	c <- struct{}{}
	//	close(c)
	//	oi.firstC = c
	//
	//	// in the body of method Do
	//	_, ok := <-oi.firstC
	//	// ok indicates whether the current call is the first
	firstC <-chan struct{}

	// Channel to broadcast the finish signal.
	doneC chan struct{}
}

func (oi *onceIndicator) Do(f func()) bool {
	_, ok := <-oi.firstC
	if ok {
		defer close(oi.doneC)
		if f != nil {
			f()
		}
	} else {
		<-oi.doneC // wait for the first call to finish
	}
	return ok
}

func (oi *onceIndicator) C() <-chan struct{} {
	return oi.doneC
}

func (oi *onceIndicator) Wait() {
	<-oi.doneC
}

func (oi *onceIndicator) Test() bool {
	select {
	case <-oi.doneC:
		return true
	default:
		return false
	}
}
