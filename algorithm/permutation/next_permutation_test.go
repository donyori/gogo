// gogo. A Golang toolbox.
// Copyright (C) 2019 Yuan Gao
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

package permutation

import (
	"sort"
	"testing"

	"github.com/donyori/gogo/function"
)

func TestNextPermutationSlice(t *testing.T) {
	ps := [][4]interface{}{
		{0, 1, 1, 3},
		{0, 1, 3, 1},
		{0, 3, 1, 1},
		{1, 0, 1, 3},
		{1, 0, 3, 1},
		{1, 1, 0, 3},
		{1, 1, 3, 0},
		{1, 3, 0, 1},
		{1, 3, 1, 0},
		{3, 0, 1, 1},
		{3, 1, 0, 1},
		{3, 1, 1, 0},
	}
	p := ps[0]
	s := p[:]
	for i, n := 1, len(ps); i < n; i++ {
		if r := NextPermutationSlice(s, function.IntLess); !r {
			t.Fatalf("No %d. NextPermutationSlice returns false before exhausted.", i)
		}
		if p != ps[i] {
			t.Errorf("No %d. Wrong permutation, want %v, got %v.", i, ps[i], p)
		}
	}
	if r := NextPermutationSlice(s, function.IntLess); r {
		t.Error("NextPermutationSlice returns true after exhausted.")
	}
	if r := NextPermutationSlice(nil, function.IntLess); r {
		t.Error("NextPermutationSlice returns true for nil slice.")
	}
	if r := NextPermutationSlice([]interface{}{}, function.IntLess); r {
		t.Error("NextPermutationSlice returns true for empty slice.")
	}
	if r := NextPermutationSlice([]interface{}{1}, function.IntLess); r {
		t.Error("NextPermutationSlice returns true for one-element slice.")
	}
}

func TestNextPermutationSortItf(t *testing.T) {
	ps := [][4]int{
		{0, 1, 1, 3},
		{0, 1, 3, 1},
		{0, 3, 1, 1},
		{1, 0, 1, 3},
		{1, 0, 3, 1},
		{1, 1, 0, 3},
		{1, 1, 3, 0},
		{1, 3, 0, 1},
		{1, 3, 1, 0},
		{3, 0, 1, 1},
		{3, 1, 0, 1},
		{3, 1, 1, 0},
	}
	p := ps[0]
	itf := sort.IntSlice(p[:])
	for i, n := 1, len(ps); i < n; i++ {
		if r := NextPermutationSortItf(itf); !r {
			t.Fatalf("No %d. NextPermutationSlice returns false before exhausted.", i)
		}
		if p != ps[i] {
			t.Errorf("No %d. Wrong permutation, want %v, got %v.", i, ps[i], p)
		}
	}
	if r := NextPermutationSortItf(itf); r {
		t.Error("NextPermutationSlice returns true after exhausted.")
	}
	if r := NextPermutationSortItf(nil); r {
		t.Error("NextPermutationSlice returns true for nil slice.")
	}
	if r := NextPermutationSortItf(sort.IntSlice{}); r {
		t.Error("NextPermutationSlice returns true for empty slice.")
	}
	if r := NextPermutationSortItf(sort.IntSlice{1}); r {
		t.Error("NextPermutationSlice returns true for one-element slice.")
	}
}
