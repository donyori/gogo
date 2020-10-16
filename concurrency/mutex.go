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

import (
	"sync"

	"github.com/donyori/gogo/errors"
)

// A lock based on Golang channel.
//
// It can be used similarly to sync.Mutex.
// Moreover, it enables the client to acquire the lock while listening to
// other channels in a select statement, rather than just blocking.
//
// Like sync.Mutex, it allows the client to get the lock
// on one goroutine, and release it on another goroutine.
// It does not support reentry.
// And it will panic if the client calls the method Unlock
// when the lock has been released.
type Mutex interface {
	sync.Locker

	// Return the channel for acquiring the lock.
	//
	// The client can acquire the lock by receiving a signal on this channel,
	// which has the same effect as calling the method Lock, i.e.,
	//  <-m.C()
	// is equivalent to
	//  m.Lock()
	C() <-chan struct{}

	// Return true if the mutex is locked, otherwise, false.
	Locked() bool
}

// Create a new instance of Mutex.
func NewMutex() Mutex {
	m := make(mutex, 1)
	m <- struct{}{}
	return m
}

// An implementation of interface Mutex.
type mutex chan struct{}

func (m mutex) Lock() {
	<-m
}

func (m mutex) Unlock() {
	if len(m) > 0 {
		panic(errors.AutoMsg("unlock of an unlocked mutex"))
	}
	m <- struct{}{}
}

func (m mutex) C() <-chan struct{} {
	return m
}

func (m mutex) Locked() bool {
	return len(m) == 0
}
