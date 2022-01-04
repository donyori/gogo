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

package permutation

import (
	"sort"
	"testing"
)

func TestNextPermutation(t *testing.T) {
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
		if r := NextPermutation(itf); !r {
			t.Fatalf("No %d. NextPermutationArray returns false before exhausted.", i)
		}
		if p != ps[i] {
			t.Errorf("No %d. Wrong permutation, want %v, got %v.", i, ps[i], p)
		}
	}
	if r := NextPermutation(itf); r {
		t.Error("NextPermutationArray returns true after exhausted.")
	}
	if r := NextPermutation(nil); r {
		t.Error("NextPermutationArray returns true for nil Interface.")
	}
	if r := NextPermutation(sort.IntSlice{}); r {
		t.Error("NextPermutationArray returns true for empty Interface.")
	}
	if r := NextPermutation(sort.IntSlice{1}); r {
		t.Error("NextPermutationArray returns true for one-element Interface.")
	}
}
