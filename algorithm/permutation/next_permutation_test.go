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
	"fmt"
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
	for i := 0; i < len(ps)-1; i++ {
		p := ps[i]
		t.Run(fmt.Sprintf("p=%v", p), func(t *testing.T) {
			itf := sort.IntSlice(p[:])
			if more := NextPermutation(itf); !more {
				t.Fatal("return false before exhausted")
			}
			if p != ps[i+1] {
				t.Errorf("got %v; want %v", p, ps[i+1])
			}
		})
	}

	final := ps[len(ps)-1]
	falseCases := []sort.IntSlice{final[:], nil, {}, {1}}
	for _, itf := range falseCases {
		var name string
		if itf != nil {
			name = fmt.Sprintf("false case?p=%v", itf)
		} else {
			name = "false case?p=<nil>"
		}
		t.Run(name, func(t *testing.T) {
			if more := NextPermutation(itf); more {
				t.Error("return true")
			}
		})
	}
}
