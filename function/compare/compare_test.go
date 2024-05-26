// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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
	"cmp"
	"fmt"
	"testing"

	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/function/compare"
)

func TestOrderedCompare(t *testing.T) {
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	float64Pairs := [][2]float64{
		{0.0, NegZero64},
		{0.0, 0.5}, {1.0, 0.5}, {0.5, 0.5},
		{NaN64A, 0.0}, {0.0, NaN64A}, {NaN64A, NaN64B},
		{Inf64, 0.0}, {NegInf64, 0.0},
		{0.0, Inf64}, {0.0, NegInf64},
		{Inf64, Inf64}, {Inf64, NegInf64},
		{NegInf64, Inf64}, {NegInf64, NegInf64},
		{NaN64A, Inf64}, {Inf64, NaN64A},
		{NaN64A, NegInf64}, {NegInf64, NaN64A},
	}
	stringPairs := [][2]string{
		{"hello", "hell"},
		{"hell", "hello"},
		{"hello", "hello"},
	}
	subtestOrderedCompare(t, "int", intPairs)
	subtestOrderedCompare(t, "float64", float64Pairs)
	subtestOrderedCompare(t, "string", stringPairs)
}

func subtestOrderedCompare[T constraints.Ordered](
	t *testing.T,
	name string,
	data [][2]T,
) {
	t.Run(name, func(t *testing.T) {
		for _, pair := range data {
			a, b := pair[0], pair[1]
			var want int
			if a < b {
				want = -1
			} else if a > b {
				want = 1
			}
			t.Run(
				fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", a, b),
				func(t *testing.T) {
					if got := compare.OrderedCompare(a, b); got != want {
						t.Errorf("got %d; want %d", got, want)
					}
				},
			)
		}
	})
}

func TestFloatCompare(t *testing.T) {
	float32Pairs := [][2]float32{
		{0.0, NegZero32},
		{0.0, 0.5}, {1.0, 0.5}, {0.5, 0.5},
		{NaN32A, 0.0}, {0.0, NaN32A}, {NaN32A, NaN32B},
		{Inf32, 0.0}, {NegInf32, 0.0},
		{0.0, Inf32}, {0.0, NegInf32},
		{Inf32, Inf32}, {Inf32, NegInf32},
		{NegInf32, Inf32}, {NegInf32, NegInf32},
		{NaN32A, Inf32}, {Inf32, NaN32A},
		{NaN32A, NegInf32}, {NegInf32, NaN32A},
	}
	t.Run("float32", func(t *testing.T) {
		for i := range float32Pairs {
			a, b := float32Pairs[i][0], float32Pairs[i][1]
			want := cmp.Compare(a, b)
			t.Run(fmt.Sprintf("a=%.1f&b=%.1f", a, b), func(t *testing.T) {
				if got := compare.FloatCompare(a, b); got != want {
					t.Errorf("got %d; want %d", got, want)
				}
			})
		}
	})

	float64Pairs := [][2]float64{
		{0.0, NegZero64},
		{0.0, 0.5}, {1.0, 0.5}, {0.5, 0.5},
		{NaN64A, 0.0}, {0.0, NaN64A}, {NaN64A, NaN64B},
		{Inf64, 0.0}, {NegInf64, 0.0},
		{0.0, Inf64}, {0.0, NegInf64},
		{Inf64, Inf64}, {Inf64, NegInf64},
		{NegInf64, Inf64}, {NegInf64, NegInf64},
		{NaN64A, Inf64}, {Inf64, NaN64A},
		{NaN64A, NegInf64}, {NegInf64, NaN64A},
	}
	t.Run("float64", func(t *testing.T) {
		for i := range float64Pairs {
			a, b := float64Pairs[i][0], float64Pairs[i][1]
			want := cmp.Compare(a, b)
			t.Run(fmt.Sprintf("a=%.1f&b=%.1f", a, b), func(t *testing.T) {
				if got := compare.FloatCompare(a, b); got != want {
					t.Errorf("got %d; want %d", got, want)
				}
			})
		}
	})
}

func TestCompareFunc_Reverse(t *testing.T) {
	f := compare.CompareFunc[int](compare.OrderedCompare[int])
	rf := f.Reverse()
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		a, b := pair[0], pair[1]
		want := f(b, a)
		t.Run(fmt.Sprintf("a=%d&b=%d", a, b), func(t *testing.T) {
			if got := rf(a, b); got != want {
				t.Errorf("got %d; want %d", got, want)
			}
		})
	}
}

func TestCompareFunc_ToEqual(t *testing.T) {
	f := compare.CompareFunc[int](compare.OrderedCompare[int])
	eq := f.ToEqual()
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		a, b := pair[0], pair[1]
		t.Run(fmt.Sprintf("a=%d&b=%d", a, b), func(t *testing.T) {
			if got := eq(a, b); got != (a == b) {
				t.Errorf("got %t", got)
			}
		})
	}
}

func TestCompareFunc_ToLess(t *testing.T) {
	f := compare.CompareFunc[int](compare.OrderedCompare[int])
	less := f.ToLess()
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		a, b := pair[0], pair[1]
		t.Run(fmt.Sprintf("a=%d&b=%d", a, b), func(t *testing.T) {
			if got := less(a, b); got != (a < b) {
				t.Errorf("got %t", got)
			}
		})
	}
}
