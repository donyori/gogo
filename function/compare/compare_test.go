// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2026  Yuan Gao
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

func TestOrdered(t *testing.T) {
	t.Parallel()

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

	subtestOrdered(t, "type=int", intPairs)
	subtestOrdered(t, "type=float64", float64Pairs)
	subtestOrdered(t, "type=string", stringPairs)
}

func subtestOrdered[T constraints.Ordered]( //nolint:thelper // this is the main body, not a helper
	t *testing.T,
	name string,
	data [][2]T,
) {
	t.Run(name, func(t *testing.T) {
		t.Parallel()

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
					t.Parallel()

					got := compare.Ordered(a, b)
					if got != want {
						t.Errorf("got %d; want %d", got, want)
					}
				},
			)
		}
	})
}

func TestFloat(t *testing.T) {
	t.Parallel()

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

	subtestFloat(t, "type=float32", float32Pairs)
	subtestFloat(t, "type=float64", float64Pairs)
}

func subtestFloat[T constraints.Float]( //nolint:thelper // this is the main body, not a helper
	t *testing.T,
	name string,
	data [][2]T,
) {
	t.Run(name, func(t *testing.T) {
		t.Parallel()

		for _, pair := range data {
			a, b := pair[0], pair[1]
			want := cmp.Compare(a, b)
			t.Run(fmt.Sprintf("a=%.1f&b=%.1f", a, b), func(t *testing.T) {
				t.Parallel()

				got := compare.Float(a, b)
				if got != want {
					t.Errorf("got %d; want %d", got, want)
				}
			})
		}
	})
}

func TestFunc_Reverse(t *testing.T) {
	t.Parallel()

	rf := compare.Func[int](compare.Ordered[int]).Reverse()
	testFuncToReverse(t, rf)
}

func TestFunc_Reverse_Nil(t *testing.T) {
	t.Parallel()

	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()

	rf := compare.Func[int](nil).Reverse()
	if rf != nil {
		t.Error("got non-nil Func")
	}
}

func TestFunc_ToEqual(t *testing.T) {
	t.Parallel()

	eq := compare.Func[int](compare.Ordered[int]).ToEqual()
	testFuncToEqual(t, eq)
}

func TestFunc_ToEqual_Nil(t *testing.T) {
	t.Parallel()

	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()

	eq := compare.Func[int](nil).ToEqual()
	if eq != nil {
		t.Error("got non-nil EqualFunc")
	}
}

func TestFunc_ToLess(t *testing.T) {
	t.Parallel()

	less := compare.Func[int](compare.Ordered[int]).ToLess()
	testFuncToLess(t, less)
}

func TestFunc_ToLess_Nil(t *testing.T) {
	t.Parallel()

	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()

	less := compare.Func[int](nil).ToLess()
	if less != nil {
		t.Error("got non-nil LessFunc")
	}
}

func TestFuncToReverse(t *testing.T) {
	t.Parallel()

	rf := compare.FuncToReverse(compare.Ordered[int])
	testFuncToReverse(t, rf)
}

func TestFuncToReverse_Nil(t *testing.T) {
	t.Parallel()

	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()

	rf := compare.FuncToReverse[int](nil)
	if rf != nil {
		t.Error("got non-nil Func")
	}
}

func TestFuncToEqual(t *testing.T) {
	t.Parallel()

	eq := compare.FuncToEqual(compare.Ordered[int])
	testFuncToEqual(t, eq)
}

func TestFuncToEqual_Nil(t *testing.T) {
	t.Parallel()

	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()

	eq := compare.FuncToEqual[int](nil)
	if eq != nil {
		t.Error("got non-nil EqualFunc")
	}
}

func TestFuncToLess(t *testing.T) {
	t.Parallel()

	less := compare.FuncToLess(compare.Ordered[int])
	testFuncToLess(t, less)
}

func TestFuncToLess_Nil(t *testing.T) {
	t.Parallel()

	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()

	less := compare.FuncToLess[int](nil)
	if less != nil {
		t.Error("got non-nil LessFunc")
	}
}

func testFuncToReverse( //nolint:thelper // this is the main body, not a helper
	t *testing.T,
	rf compare.Func[int],
) {
	if rf == nil {
		t.Error("got nil Func")
		return
	}

	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		a, b := pair[0], pair[1]

		var want int
		if a < b {
			want = 1
		} else if a > b {
			want = -1
		}

		t.Run(fmt.Sprintf("a=%d&b=%d", a, b), func(t *testing.T) {
			t.Parallel()

			got := rf(a, b)
			if got != want {
				t.Errorf("got %d; want %d", got, want)
			}
		})
	}
}

func testFuncToEqual( //nolint:thelper // this is the main body, not a helper
	t *testing.T,
	eq compare.EqualFunc[int],
) {
	if eq == nil {
		t.Error("got nil EqualFunc")
		return
	}

	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		a, b := pair[0], pair[1]
		t.Run(fmt.Sprintf("a=%d&b=%d", a, b), func(t *testing.T) {
			t.Parallel()

			got := eq(a, b)
			if got != (a == b) {
				t.Errorf("got %t", got)
			}
		})
	}
}

func testFuncToLess( //nolint:thelper // this is the main body, not a helper
	t *testing.T,
	less compare.LessFunc[int],
) {
	if less == nil {
		t.Error("got nil LessFunc")
		return
	}

	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		a, b := pair[0], pair[1]
		t.Run(fmt.Sprintf("a=%d&b=%d", a, b), func(t *testing.T) {
			t.Parallel()

			got := less(a, b)
			if got != (a < b) {
				t.Errorf("got %t", got)
			}
		})
	}
}
