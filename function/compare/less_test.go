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

package compare_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/function/compare"
)

func TestOrderedLess(t *testing.T) {
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	float64Pairs := [][2]float64{
		{0.0, 0.5}, {1.0, 0.5}, {0.5, 0.5},
		{math.NaN(), 0.0}, {0.0, math.NaN()}, {math.NaN(), math.NaN()},
		{math.Inf(1), 0.0}, {math.Inf(-1), 0.0},
		{0.0, math.Inf(1)}, {0.0, math.Inf(-1)},
		{math.Inf(1), math.Inf(1)}, {math.Inf(1), math.Inf(-1)},
		{math.Inf(-1), math.Inf(1)}, {math.Inf(-1), math.Inf(-1)},
		{math.NaN(), math.Inf(1)}, {math.Inf(1), math.NaN()},
		{math.NaN(), math.Inf(-1)}, {math.Inf(-1), math.NaN()},
	}
	stringPairs := [][2]string{
		{"hello", "hell"},
		{"hell", "hello"},
		{"hello", "hello"},
	}
	subtestOrderedLess(t, "int", intPairs)
	subtestOrderedLess(t, "float64", float64Pairs)
	subtestOrderedLess(t, "string", stringPairs)
}

func subtestOrderedLess[T constraints.Ordered](
	t *testing.T,
	name string,
	data [][2]T,
) {
	t.Run(name, func(t *testing.T) {
		for _, pair := range data {
			a, b := pair[0], pair[1]
			t.Run(
				fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", a, b),
				func(t *testing.T) {
					if got := compare.OrderedLess[T](a, b); got != (a < b) {
						t.Errorf("got %t", got)
					}
				},
			)
		}
	})
}

func TestFloatLess(t *testing.T) {
	float64Pairs := [][2]float64{
		{0.0, 0.5}, {1.0, 0.5}, {0.5, 0.5},
		{math.NaN(), 0.0}, {0.0, math.NaN()}, {math.NaN(), math.NaN()},
		{math.Inf(1), 0.0}, {math.Inf(-1), 0.0},
		{0.0, math.Inf(1)}, {0.0, math.Inf(-1)},
		{math.Inf(1), math.Inf(1)}, {math.Inf(1), math.Inf(-1)},
		{math.Inf(-1), math.Inf(1)}, {math.Inf(-1), math.Inf(-1)},
		{math.NaN(), math.Inf(1)}, {math.Inf(1), math.NaN()},
		{math.NaN(), math.Inf(-1)}, {math.Inf(-1), math.NaN()},
	}
	wants := make([]bool, len(float64Pairs))
	for i := range wants {
		a, b := float64Pairs[i][0], float64Pairs[i][1]
		wants[i] = a < b || (math.IsNaN(a) && !math.IsNaN(b))
	}
	t.Run("float64", func(t *testing.T) {
		for i := range float64Pairs {
			a, b := float64Pairs[i][0], float64Pairs[i][1]
			t.Run(fmt.Sprintf("a=%.1f&b=%.1f", a, b), func(t *testing.T) {
				if got := compare.FloatLess(a, b); got != wants[i] {
					t.Errorf("got %t", got)
				}
			})
		}
	})
	float32Pairs := make([][2]float32, len(float64Pairs))
	for i := range float32Pairs {
		float32Pairs[i][0] = float32(float64Pairs[i][0])
		float32Pairs[i][1] = float32(float64Pairs[i][1])
	}
	t.Run("float32", func(t *testing.T) {
		for i := range float32Pairs {
			a, b := float32Pairs[i][0], float32Pairs[i][1]
			t.Run(fmt.Sprintf("a=%.1f&b=%.1f", a, b), func(t *testing.T) {
				if got := compare.FloatLess(a, b); got != wants[i] {
					t.Errorf("got %t", got)
				}
			})
		}
	})
}

func TestLessFunc_Not(t *testing.T) {
	less := compare.LessFunc[int](compare.OrderedLess[int])
	nLess := less.Not()
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		a, b := pair[0], pair[1]
		t.Run(fmt.Sprintf("a=%d&b=%d", a, b), func(t *testing.T) {
			if got := nLess(a, b); got != !(a < b) {
				t.Errorf("got %t", got)
			}
		})
	}
}

func TestLessFunc_Reverse(t *testing.T) {
	less := compare.LessFunc[int](compare.OrderedLess[int])
	rLess := less.Reverse()
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		a, b := pair[0], pair[1]
		t.Run(fmt.Sprintf("a=%d&b=%d", a, b), func(t *testing.T) {
			if got := rLess(a, b); got != (b < a) {
				t.Errorf("got %t", got)
			}
		})
	}
}

func TestLessFunc_ToEqual(t *testing.T) {
	less := compare.LessFunc[int](compare.OrderedLess[int])
	eq := less.ToEqual()
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		a, b := pair[0], pair[1]
		t.Run(fmt.Sprintf("a=%d&b=%d", a, b), func(t *testing.T) {
			if got := eq(a, b); got != !(a < b || b < a) {
				t.Errorf("got %t", got)
			}
		})
	}
}
