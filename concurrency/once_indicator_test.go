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
)

type testOne int

func (o *testOne) Increase() {
	*o++
}

func TestOnceIndicator_Do(t *testing.T) {
	o := new(testOne)
	oi := NewOnceIndicator()
	const N = 10
	rs := make([]bool, N)
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(rank int) {
			defer wg.Done()
			rs[rank] = oi.Do(func() {
				o.Increase()
			})
			if v := *o; v != 1 {
				t.Errorf("Once failed: %d != 1.", v)
			}
		}(i)
	}
	wg.Wait()
	if *o != 1 {
		t.Errorf("Once failed: %d != 1.", *o)
	}
	cntr := 0
	for _, r := range rs {
		if r {
			cntr++
		}
	}
	if cntr != 1 {
		t.Errorf("Not only one call of Do return true. #true: %d.", cntr)
	}
}

func TestOnceIndicator_Do_Panic(t *testing.T) {
	oi := NewOnceIndicator()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("OnceIndicator.Do did NOT panic.")
			}
		}()
		oi.Do(func() {
			panic("first panic")
		})
	}()

	oi.Do(func() {
		t.Fatal("OnceIndicator.Do called twice.")
	})

	select {
	case <-oi.C():
	default:
		t.Error("OnceIndicator.Do did NOT trigger the channel after calling Do.")
	}
}

func TestOnceIndicator_Do_NilF(t *testing.T) {
	oi := NewOnceIndicator()
	if !oi.Do(nil) {
		t.Error("OnceIndicator.Do returns false on the first call.")
	}

	select {
	case <-oi.C():
	default:
		t.Error("OnceIndicator.Do did NOT trigger the channel after calling Do with f set to nil.")
	}
}
