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
		t.Error("m.Mutex() is not working.")
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
		t.Error("m.Locked() = true on a new lock.")
	}
	m.Lock()
	if !m.Locked() {
		t.Error("m.Locked() = false after calling Mutex.")
	}
	m.Unlock()
	if m.Locked() {
		t.Error("m.Locked() = true after releasing the lock.")
	}
	<-m.C()
	if !m.Locked() {
		t.Error("m.Locked() = false after receiving on m.C().")
	}
}

func TestMutex_C_UnlockOfUnlockedLock(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("No panic when calling Unlock of an unlocked lock.")
		}
	}()
	NewMutex().Unlock()
}

func TestMutex_C_UnlockTwice(t *testing.T) {
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
