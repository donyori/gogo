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

package concurrency_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/donyori/gogo/concurrency"
)

func TestMutex_Lock(t *testing.T) {
	t.Run("m.Lock()", func(t *testing.T) {
		m := concurrency.NewMutex()
		c := make(chan struct{})
		var start time.Time
		go func() {
			m.Lock()
			defer m.Unlock()
			start = time.Now()
			c <- struct{}{}
			time.Sleep(time.Millisecond)
		}()
		<-c
		m.Lock()
		m.Unlock()
		if time.Since(start) < time.Millisecond {
			t.Error("not working")
		}
	})
	t.Run("<-m.C()", func(t *testing.T) {
		m := concurrency.NewMutex()
		c := make(chan struct{})
		var start time.Time
		go func() {
			<-m.C() // equivalent to m.Lock()
			defer m.Unlock()
			start = time.Now()
			c <- struct{}{}
			time.Sleep(time.Millisecond)
		}()
		<-c
		m.Lock()
		m.Unlock()
		if time.Since(start) < time.Millisecond {
			t.Error("not working")
		}
	})
}

func TestMutex_Locked(t *testing.T) {
	m := concurrency.NewMutex()
	if m.Locked() {
		t.Fatal("true on a new mutex")
	}
	m.Lock()
	if !m.Locked() {
		t.Fatal("false after calling Lock")
	}
	m.Unlock()
	if m.Locked() {
		t.Fatal("true after calling Unlock")
	}
	<-m.C()
	if !m.Locked() {
		t.Error("false after receiving on m.C()")
	}
}

func TestMutex_UnlockOfUnlockedMutex(t *testing.T) {
	defer func() {
		if e := recover(); !isUnlockPanicMessage(e) {
			t.Error(e)
		}
	}()
	concurrency.NewMutex().Unlock() // want panic here
	t.Error("no panic when calling Unlock of an unlocked mutex")
}

func TestMutex_UnlockTwice(t *testing.T) {
	defer func() {
		if e := recover(); !isUnlockPanicMessage(e) {
			t.Error(e)
		}
	}()
	m := concurrency.NewMutex()
	m.Lock()
	m.Unlock()
	m.Unlock() // want panic here
	t.Error("no panic when calling Unlock twice")
}

func TestMutex_Fairness(t *testing.T) {
	// This test refers to TestMutexFairness,
	// a test of sync.Mutex, in the file sync/mutex_test.go.
	m := concurrency.NewMutex()
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
				time.Sleep(time.Microsecond * 100)
				m.Unlock()
			}
		}
	}()
	doneC := make(chan struct{})
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			time.Sleep(time.Microsecond * 100)
			select {
			case <-stopC:
				return
			case <-m.C():
				m.Unlock()
			}
		}
		close(doneC)
	}()
	select {
	case <-doneC:
	case <-time.After(time.Second * 10):
		t.Error("cannot acquire the lock in 10 seconds")
	}
	close(sc)
	sc = nil
	wg.Wait()
}

func isUnlockPanicMessage(err any) bool {
	if err == nil {
		return false
	}
	msg, ok := err.(string)
	return ok && strings.HasSuffix(msg, "unlock of an unlocked mutex")
}
