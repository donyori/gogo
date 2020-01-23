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

package function

import (
	"math"
	"testing"
)

func TestIntLess(t *testing.T) {
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		if r := IntLess(pair[0], pair[1]); r != (pair[0] < pair[1]) {
			t.Errorf("IntLess(%d, %d): %t.", pair[0], pair[1], r)
		}
	}
}

func TestFloat64Less(t *testing.T) {
	floatPairs := [][2]float64{
		{0.0, 0.5}, {1.0, 0.5}, {0.5, 0.5},
		{math.NaN(), 0.0}, {0.0, math.NaN()}, {math.NaN(), math.NaN()},
		{math.Inf(1), 0.0}, {math.Inf(-1), 0.0},
		{0.0, math.Inf(1)}, {0.0, math.Inf(-1)},
		{math.Inf(1), math.Inf(1)}, {math.Inf(1), math.Inf(-1)},
		{math.Inf(-1), math.Inf(1)}, {math.Inf(-1), math.Inf(-1)},
		{math.NaN(), math.Inf(1)}, {math.Inf(1), math.NaN()},
		{math.NaN(), math.Inf(-1)}, {math.Inf(-1), math.NaN()},
	}
	for _, pair := range floatPairs {
		if r := Float64Less(pair[0], pair[1]); r != (pair[0] < pair[1]) {
			t.Errorf("Float64Less(%f, %f): %t.", pair[0], pair[1], r)
		}
	}
}

func TestStringLess(t *testing.T) {
	stringPairs := [][2]string{
		{"hello", "hell"}, {"hell", "hello"}, {"hello", "hello"},
	}
	for _, pair := range stringPairs {
		if r := StringLess(pair[0], pair[1]); r != (pair[0] < pair[1]) {
			t.Errorf("StringLess(%s, %s): %t.", pair[0], pair[1], r)
		}
	}
}

func TestLessFunc_Not(t *testing.T) {
	var less LessFunc = IntLess
	nLess := less.Not()
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		r1 := !less(pair[0], pair[1])
		r2 := nLess(pair[0], pair[1])
		if r1 != r2 {
			t.Errorf("nLess(%d, %d) != !less(%[1]d, %d).", pair[0], pair[1])
		}
	}
}

func TestLessFunc_Reverse(t *testing.T) {
	var less LessFunc = IntLess
	rLess := less.Reverse()
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		r1 := less(pair[1], pair[0])
		r2 := rLess(pair[0], pair[1])
		if r1 != r2 {
			t.Errorf("rLess(%d, %d) != less(%[2]d, %[1]d).", pair[0], pair[1])
		}
	}
}
