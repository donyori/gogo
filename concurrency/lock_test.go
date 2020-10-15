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

func TestLock_Lock(t *testing.T) {
	k := NewLock()

	start := time.Now()
	go func() {
		k.Lock()
		defer k.Unlock()
		time.Sleep(time.Millisecond)
	}()
	time.Sleep(time.Microsecond)
	k.Lock()
	if time.Since(start) < time.Millisecond {
		t.Error("k.Lock() is not working.")
	}
	k.Unlock()

	start = time.Now()
	go func() {
		<-k.C()
		defer k.Unlock()
		time.Sleep(time.Millisecond)
	}()
	time.Sleep(time.Microsecond)
	k.Lock()
	if time.Since(start) < time.Millisecond {
		t.Error("<-k.C() is not working.")
	}
}

func TestLock_Locked(t *testing.T) {
	k := NewLock()
	if k.Locked() {
		t.Error("k.Locked() = true on a new lock.")
	}
	k.Lock()
	if !k.Locked() {
		t.Error("k.Locked() = false after calling Lock.")
	}
	k.Unlock()
	if k.Locked() {
		t.Error("k.Locked() = true after releasing the lock.")
	}
	<-k.C()
	if !k.Locked() {
		t.Error("k.Locked() = false after receiving on k.C().")
	}
}

func TestLock_UnlockOfUnlockedLock(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("No panic when calling Unlock of an unlocked lock.")
		}
	}()
	NewLock().Unlock()
}

func TestLock_UnlockTwice(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("No panic when calling Unlock twice.")
		}
	}()
	k := NewLock()
	k.Lock()
	k.Unlock()
	k.Unlock()
}
