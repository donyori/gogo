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
	"sync"
	"sync/atomic"
	"testing"

	"github.com/donyori/gogo/concurrency"
)

type one struct {
	v atomic.Int32
}

func (o *one) Increase() {
	o.v.Add(1)
}

func (o *one) Load() int32 {
	return o.v.Load()
}

func TestOnceIndicator_Do_Once(t *testing.T) {
	o := new(one)
	oi := concurrency.NewOnceIndicator()
	const N int = 10
	rs := make([]bool, N)
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(rank int) {
			defer wg.Done()
			rs[rank] = oi.Do(func() {
				o.Increase()
			})
			if v := o.Load(); v != 1 {
				t.Errorf("goroutine %d, once failed: %d is not 1", rank, v)
			}
		}(i)
	}
	wg.Wait()
	if v := o.Load(); v != 1 {
		t.Errorf("main goroutine, once failed: %d is not 1", v)
	}
	var ctr int
	for _, r := range rs {
		if r {
			ctr++
		}
	}
	if ctr != 1 {
		t.Errorf("not only one call of Do return true, #true: %d", ctr)
	}
}

func TestOnceIndicator_Do_Panic(t *testing.T) {
	oi := concurrency.NewOnceIndicator()
	func() {
		defer func() {
			if e := recover(); e == nil {
				t.Fatal("OnceIndicator.Do did NOT panic")
			}
		}()
		oi.Do(func() {
			panic("panic")
		})
	}()

	oi.Do(func() {
		t.Fatal("OnceIndicator.Do called twice")
	})

	select {
	case <-oi.C():
	default:
		t.Error("OnceIndicator.Do did NOT trigger the channel after being called")
	}
}

func TestOnceIndicator_Do_NilF(t *testing.T) {
	oi := concurrency.NewOnceIndicator()
	if !oi.Do(nil) {
		t.Error("OnceIndicator.Do returned false on the first call")
	}

	select {
	case <-oi.C():
	default:
		t.Error("OnceIndicator.Do did NOT trigger the channel after being called with a nil f")
	}
}
