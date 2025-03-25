// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2025  Yuan Gao
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
	"github.com/donyori/gogo/internal/floats"
)

func TestOrderedLess(t *testing.T) {
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	float64Pairs := [][2]float64{
		{0.0, floats.NegZero64},
		{0.0, 0.5}, {1.0, 0.5}, {0.5, 0.5},
		{floats.NaN64A, 0.0}, {0.0, floats.NaN64A},
		{floats.NaN64A, floats.NaN64B},
		{floats.Inf64, 0.0}, {floats.NegInf64, 0.0},
		{0.0, floats.Inf64}, {0.0, floats.NegInf64},
		{floats.Inf64, floats.Inf64}, {floats.Inf64, floats.NegInf64},
		{floats.NegInf64, floats.Inf64}, {floats.NegInf64, floats.NegInf64},
		{floats.NaN64A, floats.Inf64}, {floats.Inf64, floats.NaN64A},
		{floats.NaN64A, floats.NegInf64}, {floats.NegInf64, floats.NaN64A},
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
					if got := compare.OrderedLess(a, b); got != (a < b) {
						t.Errorf("got %t", got)
					}
				},
			)
		}
	})
}

func TestFloatLess(t *testing.T) {
	float32Pairs := [][2]float32{
		{0.0, floats.NegZero32},
		{0.0, 0.5}, {1.0, 0.5}, {0.5, 0.5},
		{floats.NaN32A, 0.0}, {0.0, floats.NaN32A},
		{floats.NaN32A, floats.NaN32B},
		{floats.Inf32, 0.0}, {floats.NegInf32, 0.0},
		{0.0, floats.Inf32}, {0.0, floats.NegInf32},
		{floats.Inf32, floats.Inf32}, {floats.Inf32, floats.NegInf32},
		{floats.NegInf32, floats.Inf32}, {floats.NegInf32, floats.NegInf32},
		{floats.NaN32A, floats.Inf32}, {floats.Inf32, floats.NaN32A},
		{floats.NaN32A, floats.NegInf32}, {floats.NegInf32, floats.NaN32A},
	}
	float64Pairs := [][2]float64{
		{0.0, floats.NegZero64},
		{0.0, 0.5}, {1.0, 0.5}, {0.5, 0.5},
		{floats.NaN64A, 0.0}, {0.0, floats.NaN64A},
		{floats.NaN64A, floats.NaN64B},
		{floats.Inf64, 0.0}, {floats.NegInf64, 0.0},
		{0.0, floats.Inf64}, {0.0, floats.NegInf64},
		{floats.Inf64, floats.Inf64}, {floats.Inf64, floats.NegInf64},
		{floats.NegInf64, floats.Inf64}, {floats.NegInf64, floats.NegInf64},
		{floats.NaN64A, floats.Inf64}, {floats.Inf64, floats.NaN64A},
		{floats.NaN64A, floats.NegInf64}, {floats.NegInf64, floats.NaN64A},
	}
	subtestFloatLess(t, "type=float32", float32Pairs)
	subtestFloatLess(t, "type=float64", float64Pairs)
}

func subtestFloatLess[T constraints.Float](
	t *testing.T,
	name string,
	data [][2]T,
) {
	t.Run(name, func(t *testing.T) {
		for _, pair := range data {
			a, b := pair[0], pair[1]
			t.Run(fmt.Sprintf("a=%.1f&b=%.1f", a, b), func(t *testing.T) {
				if got := compare.FloatLess(a, b); got != cmp.Less(a, b) {
					t.Errorf("got %t", got)
				}
			})
		}
	})
}

func TestLessFunc_Not(t *testing.T) {
	less := compare.LessFunc[int](compare.OrderedLess[int])
	nLess := less.Not()
	if nLess == nil {
		t.Fatal("got nil LessFunc")
	}
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

func TestLessFunc_Not_Nil(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()
	nLess := compare.LessFunc[int](nil).Not()
	if nLess != nil {
		t.Error("got non-nil LessFunc")
	}
}

func TestLessFunc_Reverse(t *testing.T) {
	less := compare.LessFunc[int](compare.OrderedLess[int])
	rLess := less.Reverse()
	if rLess == nil {
		t.Fatal("got nil LessFunc")
	}
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

func TestLessFunc_Reverse_Nil(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()
	rLess := compare.LessFunc[int](nil).Reverse()
	if rLess != nil {
		t.Error("got non-nil LessFunc")
	}
}

func TestLessFunc_ToEqual(t *testing.T) {
	less := compare.LessFunc[int](compare.OrderedLess[int])
	eq := less.ToEqual()
	if eq == nil {
		t.Fatal("got nil EqualFunc")
	}
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

func TestLessFunc_ToEqual_Nil(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()
	eq := compare.LessFunc[int](nil).ToEqual()
	if eq != nil {
		t.Error("got non-nil EqualFunc")
	}
}

func TestLessFunc_ToCompare(t *testing.T) {
	less := compare.LessFunc[int](compare.OrderedLess[int])
	f := less.ToCompare()
	if f == nil {
		t.Fatal("got nil CompareFunc")
	}
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		a, b := pair[0], pair[1]
		want := cmp.Compare(a, b)
		t.Run(fmt.Sprintf("a=%d&b=%d", a, b), func(t *testing.T) {
			if got := f(a, b); got != want {
				t.Errorf("got %d; want %d", got, want)
			}
		})
	}
}

func TestLessFunc_ToCompare_Nil(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()
	f := compare.LessFunc[int](nil).ToCompare()
	if f != nil {
		t.Error("got non-nil CompareFunc")
	}
}
