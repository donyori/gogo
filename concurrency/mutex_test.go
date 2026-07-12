// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2026  Yuan Gao
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
	t.Parallel()

	lockByMethodLock := func(m concurrency.Mutex) {
		m.Lock()
	}
	lockByMethodC := func(m concurrency.Mutex) {
		<-m.C() // equivalent to m.Lock()
	}
	testCases := []struct {
		name       string
		lockInMain func(m concurrency.Mutex)
		lockInGo   func(m concurrency.Mutex)
	}{
		{"lockInMain=Lock&lockInGo=Lock", lockByMethodLock, lockByMethodLock},
		{"lockInMain=Lock&lockInGo=C", lockByMethodLock, lockByMethodC},
		{"lockInMain=C&lockInGo=Lock", lockByMethodC, lockByMethodLock},
		{"lockInMain=C&lockInGo=C", lockByMethodC, lockByMethodC},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			const SleepDuration time.Duration = time.Millisecond

			m := concurrency.NewMutex()
			if m == nil {
				t.Fatal("got nil mutex")
			}

			finishLockTimeC := make(chan time.Time)
			go func(finishLockTimeC chan<- time.Time) {
				tc.lockInGo(m)
				defer m.Unlock()

				finishLockTimeC <- time.Now()

				time.Sleep(SleepDuration)
			}(finishLockTimeC)

			start := <-finishLockTimeC

			tc.lockInMain(m)

			if time.Since(start) < SleepDuration {
				t.Error("mutex does not work")
			}
		})
	}
}

func TestMutex_Locked(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	defer func() {
		if e := recover(); !isUnlockPanicMessage(e) {
			t.Error(e)
		}
	}()
	concurrency.NewMutex().Unlock() // want panic here
	t.Error("no panic when calling Unlock of an unlocked mutex")
}

func TestMutex_UnlockTwice(t *testing.T) {
	t.Parallel()

	defer func() {
		if e := recover(); !isUnlockPanicMessage(e) {
			t.Error(e)
		}
	}()

	m := concurrency.NewMutex()
	m.Lock()

	var _ struct{} // to suppress "SA2001: empty critical section (staticcheck)"

	m.Unlock()

	m.Unlock() // want panic here
	t.Error("no panic when calling Unlock twice")
}

func TestMutex_Fairness(t *testing.T) {
	t.Parallel()

	// This test refers to TestMutexFairness,
	// a test of sync.Mutex, in the file sync/mutex_test.go.

	m := concurrency.NewMutex()
	stopC := make(chan struct{})
	sc := stopC // a copy of stopC, for closing stopC

	defer func() {
		if sc != nil {
			close(sc)
		}
	}()

	var wg sync.WaitGroup

	wg.Go(func() {
		for {
			select {
			case <-stopC:
				return
			case <-m.C():
				time.Sleep(time.Microsecond * 100)
				m.Unlock()
			}
		}
	})

	doneC := make(chan struct{})

	wg.Go(func() {
		for range 10 {
			time.Sleep(time.Microsecond * 100)

			select {
			case <-stopC:
				return
			case <-m.C():
				m.Unlock()
			}
		}

		close(doneC)
	})

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
