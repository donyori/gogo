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

package function

import "testing"

func TestEqual(t *testing.T) {
	pairs := [][2]interface{}{
		{nil, nil},
		{1, nil},
		{nil, 1},
		{1, 1},
		{1, 0},
		{0, 1},
		{1., 1.},
		{1, 1.},
		{1., 1},
	}
	for _, pair := range pairs {
		if r := Equal(pair[0], pair[1]); r != (pair[0] == pair[1]) {
			t.Errorf("Equal(%v, %v) = %t.", pair[0], pair[1], r)
		}
	}
}

func TestEqualFunc_Not(t *testing.T) {
	pairs := [][2]interface{}{
		{nil, nil},
		{1, nil},
		{nil, 1},
		{1, 1},
		{1, 0},
		{0, 1},
		{1., 1.},
		{1, 1.},
		{1., 1},
	}
	var eq EqualFunc = Equal
	nEq := eq.Not()
	for _, pair := range pairs {
		r1 := !eq(pair[0], pair[1])
		r2 := nEq(pair[0], pair[1])
		if r1 != r2 {
			t.Errorf("nEq(%v, %v) != !eq(%[1]v, %v).", pair[0], pair[1])
		}
	}
}

func TestGenerateEqualViaLess(t *testing.T) {
	intPairs := [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}}
	eq := GenerateEqualViaLess(IntLess)
	for _, pair := range intPairs {
		if r := eq(pair[0], pair[1]); r != (pair[0] == pair[1]) {
			t.Errorf("eq(%d, %d) = %t.", pair[0], pair[1], r)
		}
	}
}
