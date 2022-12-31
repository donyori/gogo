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

import (
	"sync"

	"github.com/donyori/gogo/errors"
)

// Mutex is a mutual exclusion lock based on Go channel.
//
// It can be used similarly to sync.Mutex.
// Moreover, it enables the client to acquire the lock while listening to
// other channels in a select statement, rather than just blocking.
//
// Like sync.Mutex, it permits the client to get the lock
// on one goroutine, and release it on another goroutine.
// It does not support reentry.
// And it will panic if the client calls the method Unlock
// when the lock has been released.
type Mutex interface {
	sync.Locker

	// C returns the channel for acquiring the lock.
	//
	// The client can acquire the lock by receiving a signal on this channel,
	// which has the same effect as calling the method Lock, i.e.,
	//  <-m.C()
	// is equivalent to
	//  m.Lock()
	C() <-chan struct{}

	// Locked reports whether the mutex is locked.
	Locked() bool
}

// NewMutex creates a new instance of Mutex.
func NewMutex() Mutex {
	m := &mutex{make(chan struct{}, 1)}
	m.c <- struct{}{}
	return m
}

// mutex is an implementation of interface Mutex.
type mutex struct {
	c chan struct{}
}

// Lock acquires the lock on the mutex.
// It blocks until the lock is gotten.
func (m *mutex) Lock() {
	<-m.c
}

// Unlock releases the lock on the mutex.
// It panics if the mutex is unlocked.
func (m *mutex) Unlock() {
	select {
	case m.c <- struct{}{}:
	default:
		panic(errors.AutoMsg("unlock of an unlocked mutex"))
	}
}

// C returns the channel for acquiring the lock.
//
// The client can acquire the lock by receiving a signal on this channel,
// which has the same effect as calling the method Lock, i.e.,
//
//	<-m.C()
//
// is equivalent to
//
//	m.Lock()
func (m *mutex) C() <-chan struct{} {
	return m.c
}

// Locked reports whether the mutex is locked.
func (m *mutex) Locked() bool {
	return len(m.c) == 0
}
