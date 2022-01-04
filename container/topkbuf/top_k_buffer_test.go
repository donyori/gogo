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

package topkbuf

import (
	"sort"
	"testing"

	"github.com/donyori/gogo/function/compare"
)

func TestNewTopKBuffer(t *testing.T) {
	k := 5
	tkb := NewTopKBuffer(k, compare.IntLess, 3, 2, 5)
	y := []int{2, 3, 5}
	x := tkb.Drain()
	if len(x) == len(y) {
		for i, n := 0, len(x); i < n; i++ {
			if x[i] != y[i] {
				t.Errorf("test1: x[%d] = %v != y[%d] = %d.", i, x[i], i, y[i])
			}
		}
	} else {
		t.Errorf("test1: len(x): %d != len(y): %d", len(x), len(y))
	}
	tkb = NewTopKBuffer(k, compare.IntLess, 3, 2, 5, 1, 0, 7, 9, 3)
	y = []int{0, 1, 2, 3, 3}
	x = tkb.Drain()
	if len(x) == len(y) {
		for i, n := 0, len(x); i < n; i++ {
			if x[i] != y[i] {
				t.Errorf("test2: x[%d] = %v != y[%d] = %d.", i, x[i], i, y[i])
			}
		}
	} else {
		t.Errorf("test2: len(x): %d != len(y): %d", len(x), len(y))
	}
}

func TestTopKBuffer(t *testing.T) {
	k := 5
	samples := []interface{}{3, 2, 5, 1, 0, 7, 9, 0, 3}
	tkb := NewTopKBuffer(k, compare.IntLess)
	if x := tkb.K(); x != k {
		t.Errorf("tkb.K(): %d != %d", x, k)
	}
	if n := tkb.Len(); n != 0 {
		t.Errorf("After create an empty TopKBuffer: tkb.Len(): %d != 0", n)
	}
	tkb.Add(samples...)
	if n := tkb.Len(); n > k {
		t.Errorf("After add samples: tkb.Len(): %d > k: %d.", n, k)
	}
	output := tkb.Drain()
	if n := len(output); n == k {
		sortedSamples := make([]int, len(samples))
		for i := range samples {
			sortedSamples[i] = samples[i].(int)
		}
		sort.Ints(sortedSamples)
		for i := 0; i < n; i++ {
			if output[i] != sortedSamples[i] {
				t.Errorf("output[%d] = %v != %d", i, output[i], sortedSamples[i])
			}
		}
	} else {
		t.Errorf("len(output): %d != %d", n, k)
	}
	if n := tkb.Len(); n != 0 {
		t.Errorf("After drain: tkb.Len(): %d != 0", n)
	}
	tkb.Add(samples...)
	if n := tkb.Len(); n != k {
		t.Errorf("After add samples again: tkb.Len(): %d != %d", n, k)
	}
	tkb.Clear()
	if n := tkb.Len(); n != 0 {
		t.Errorf("After clear: tkb.Len(): %d != 0", n)
	}
}
