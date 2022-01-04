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

package concurrency

import "sync"

// OnceIndicator is an object that performs exactly one action, like sync.Once.
// Moreover, it can indicate whether the action has been performed or not,
// and enable the client on another goroutine to wait for the action to finish.
type OnceIndicator interface {
	// Do performs the same as the method Do of sync.Once, and indicate whether
	// the function f is called in this invocation.
	//
	// In detail, it calls the function f and returns true if and only if
	// the method Do is being called for the first time for this instance
	// of OnceIndicator.
	// Otherwise, it does nothing but waits for the first call of f to finish,
	// and then returns false.
	//
	// If the client wants to do nothing but trigger this indicator,
	// just set f to nil (no panic will happen).
	Do(f func()) bool

	// C returns a channel that will be closed after calling the method Do
	// for the first time for this instance of OnceIndicator.
	C() <-chan struct{}

	// Wait waits for the first call of the method Do for this instance
	// on another goroutine to finish.
	Wait()

	// Test reports whether the method Do for this instance is called or not.
	//
	// It returns true if and only if the first call of the method Do
	// for this instance has finished.
	Test() bool
}

// NewOnceIndicator creates a new instance of OnceIndicator.
func NewOnceIndicator() OnceIndicator {
	return &onceIndicator{c: make(chan struct{})}
}

// onceIndicator is an implementation of interface OnceIndicator.
type onceIndicator struct {
	once sync.Once     // Once object.
	c    chan struct{} // Channel to broadcast the finish signal.
}

// Do performs the same as the method Do of sync.Once, and indicate whether
// the function f is called in this invocation.
//
// In detail, it calls the function f and returns true if and only if
// the method Do is being called for the first time for this instance
// of OnceIndicator.
// Otherwise, it does nothing but waits for the first call of f to finish,
// and then returns false.
//
// If the client wants to do nothing but trigger this indicator,
// just set f to nil (no panic will happen).
func (oi *onceIndicator) Do(f func()) bool {
	var r bool
	oi.once.Do(func() {
		r = true
		defer close(oi.c)
		if f != nil {
			f()
		}
	})
	return r
}

// C returns a channel that will be closed after calling the method Do
// for the first time for this instance of OnceIndicator.
func (oi *onceIndicator) C() <-chan struct{} {
	return oi.c
}

// Wait waits for the first call of the method Do for this instance
// on another goroutine to finish.
func (oi *onceIndicator) Wait() {
	<-oi.c
}

// Test reports whether the method Do for this instance is called or not.
//
// It returns true if and only if the first call of the method Do
// for this instance has finished.
func (oi *onceIndicator) Test() bool {
	select {
	case <-oi.c:
		return true
	default:
		return false
	}
}
