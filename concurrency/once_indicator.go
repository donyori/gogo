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

package concurrency

import "sync"

// An object that performs exactly one action, just like sync.Once.
// Moreover, it can indicate whether the action has been performed or not,
// and enable the client on another goroutine to wait for the action to finish.
type OnceIndicator interface {
	// Perform the same as the method Do of sync.Once, and indicate whether
	// the function f is called in this invocation.
	//
	// In detail, it calls the function f and returns true iff the method Do
	// is being called for the first time for this instance of OnceIndicator.
	// Otherwise, it does nothing but waits for the first call of f to finish,
	// and then returns false.
	//
	// If the client wants to do nothing but trigger this indicator,
	// just set f to nil (no panic will happen).
	Do(f func()) bool

	// Return a channel that will be closed after calling the method Do
	// for the first time for this instance of OnceIndicator.
	C() <-chan struct{}

	// Wait for the first call of the method Do for this instance
	// on another goroutine to finish.
	Wait()

	// Test whether the method Do for this instance is called or not.
	//
	// It returns true iff the first call of the method Do
	// for this instance has finished.
	Test() bool
}

// Create a new instance of OnceIndicator.
func NewOnceIndicator() OnceIndicator {
	return &onceIndicator{c: make(chan struct{})}
}

// An implementation of interface OnceIndicator.
type onceIndicator struct {
	once sync.Once     // Once object.
	c    chan struct{} // Channel to broadcast the finish signal.
}

func (oi *onceIndicator) Do(f func()) bool {
	r := false
	oi.once.Do(func() {
		r = true
		defer close(oi.c)
		if f != nil {
			f()
		}
	})
	return r
}

func (oi *onceIndicator) C() <-chan struct{} {
	return oi.c
}

func (oi *onceIndicator) Wait() {
	<-oi.c
}

func (oi *onceIndicator) Test() bool {
	select {
	case <-oi.c:
		return true
	default:
		return false
	}
}
