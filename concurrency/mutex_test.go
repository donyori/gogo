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
	"testing"
	"time"
)

func TestMutex_Lock(t *testing.T) {
	m := NewMutex()

	start := time.Now()
	go func() {
		m.Lock()
		defer m.Unlock()
		time.Sleep(time.Millisecond)
	}()
	time.Sleep(time.Microsecond)
	m.Lock()
	if time.Since(start) < time.Millisecond {
		t.Error("m.Lock() is not working.")
	}
	m.Unlock()

	start = time.Now()
	go func() {
		<-m.C()
		defer m.Unlock()
		time.Sleep(time.Millisecond)
	}()
	time.Sleep(time.Microsecond)
	m.Lock()
	if time.Since(start) < time.Millisecond {
		t.Error("<-m.C() is not working.")
	}
}

func TestMutex_Locked(t *testing.T) {
	m := NewMutex()
	if m.Locked() {
		t.Error("m.Locked() = true on a new mutex.")
	}
	m.Lock()
	if !m.Locked() {
		t.Error("m.Locked() = false after calling Lock.")
	}
	m.Unlock()
	if m.Locked() {
		t.Error("m.Locked() = true after calling Unlock.")
	}
	<-m.C()
	if !m.Locked() {
		t.Error("m.Locked() = false after receiving on m.C().")
	}
}

func TestMutex_UnlockOfUnlockedMutex(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("No panic when calling Unlock of an unlocked mutex.")
		}
	}()
	NewMutex().Unlock()
}

func TestMutex_UnlockTwice(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("No panic when calling Unlock twice.")
		}
	}()
	m := NewMutex()
	m.Lock()
	m.Unlock()
	m.Unlock()
}

func TestMutex_Fairness(t *testing.T) {
	// This test refers to TestMutexFairness
	// (a test of sync.Mutex, in the file sync/mutex_test.go).
	m := NewMutex()
	stopC := make(chan struct{})
	sc := stopC // A copy of stopC, for closing stopC.
	defer func() {
		if sc != nil {
			close(sc)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopC:
				return
			case <-m.C():
				time.Sleep(time.Microsecond * 10)
				m.Unlock()
			}
		}
	}()
	doneC := make(chan struct{})
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			select {
			case <-stopC:
				return
			case <-m.C():
				time.Sleep(time.Microsecond * 10)
				m.Unlock()
			}
		}
		close(doneC)
	}()
	select {
	case <-doneC:
	case <-time.After(time.Second):
		t.Error("Cannot acquire the lock in 1 second.")
	}
	close(sc)
	sc = nil
	wg.Wait()
}
