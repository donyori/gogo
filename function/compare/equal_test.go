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
	"fmt"
	"reflect"
	"testing"

	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/function/compare"
	"github.com/donyori/gogo/internal/floats"
)

func TestAnyEqual(t *testing.T) {
	// Comparable cases:
	pairs := [][2]any{
		{nil, nil},
		{1, nil},
		{nil, 1},
		{1, 1},
		{1, 0},
		{0, 1},
		{1., 1.},
		{1, 1.},
		{1., 1},
		{0., floats.NegZero64},
	}
	for _, pair := range pairs {
		t.Run(
			fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", pair[0], pair[1]),
			func(t *testing.T) {
				got := compare.AnyEqual(pair[0], pair[1])
				if got != (pair[0] == pair[1]) {
					t.Errorf("got %t", got)
				}
			},
		)
	}

	// Incomparable cases:
	s1, s2 := []int{}, []int{1, 2}
	pairs = [][2]any{
		{nil, s1},
		{s1, nil},
		{s1, s1},
		{s1, 1},
		{1, s1},
		{nil, s2},
		{s2, nil},
		{s2, s2},
		{s2, 1},
		{1, s2},
		{s1, s2},
	}
	for _, pair := range pairs {
		t.Run(
			fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", pair[0], pair[1]),
			func(t *testing.T) {
				got := compare.AnyEqual(pair[0], pair[1])
				if got {
					t.Errorf("got %t", got)
				}
			},
		)
	}
	// Incomparable type nil value, and nil any:
	var s []int
	pairs = [][2]any{
		{nil, s},
		{s, nil},
	}
	for _, pair := range pairs {
		t.Run(
			fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", pair[0], pair[1]),
			func(t *testing.T) {
				got := compare.AnyEqual(pair[0], pair[1])
				if !got {
					t.Errorf("got %t", got)
				}
			},
		)
	}
}

func TestEqualFunc_Not_AnyEqual(t *testing.T) {
	neq := compare.AnyEqual.Not()
	if neq == nil {
		t.Fatal("got nil EqualFunc")
	}
	// Comparable cases:
	pairs := [][2]any{
		{nil, nil},
		{1, nil},
		{nil, 1},
		{1, 1},
		{1, 0},
		{0, 1},
		{1., 1.},
		{1, 1.},
		{1., 1},
		{0., floats.NegZero64},
	}
	for _, pair := range pairs {
		t.Run(
			fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", pair[0], pair[1]),
			func(t *testing.T) {
				got := neq(pair[0], pair[1])
				if got != (pair[0] != pair[1]) {
					t.Errorf("got %t", got)
				}
			},
		)
	}

	// Incomparable cases:
	s1, s2 := []int{}, []int{1, 2}
	pairs = [][2]any{
		{nil, s1},
		{s1, nil},
		{s1, s1},
		{s1, 1},
		{1, s1},
		{nil, s2},
		{s2, nil},
		{s2, s2},
		{s2, 1},
		{1, s2},
		{s1, s2},
	}
	for _, pair := range pairs {
		t.Run(
			fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", pair[0], pair[1]),
			func(t *testing.T) {
				got := neq(pair[0], pair[1])
				if !got {
					t.Errorf("got %t", got)
				}
			},
		)
	}
	// Incomparable type nil value, and nil any:
	var s []int
	pairs = [][2]any{
		{nil, s},
		{s, nil},
	}
	for _, pair := range pairs {
		t.Run(
			fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", pair[0], pair[1]),
			func(t *testing.T) {
				got := neq(pair[0], pair[1])
				if got {
					t.Errorf("got %t", got)
				}
			},
		)
	}
}

func TestEqualFunc_Not_Nil(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()
	neq := compare.EqualFunc[int](nil).Not()
	if neq != nil {
		t.Error("got non-nil EqualFunc")
	}
}

func TestEqualFunc_Reflexive_AnyEqual(t *testing.T) {
	req := compare.AnyEqual.Reflexive()
	if req == nil {
		t.Fatal("got nil EqualFunc")
	}
	// Comparable cases:
	pairs := [][2]any{
		{nil, nil},
		{1, nil},
		{nil, 1},
		{1, 1},
		{1, 0},
		{0, 1},
		{1., 1.},
		{1, 1.},
		{1., 1},
		{0., floats.NegZero64},
		{0., floats.NaN64A},
		{floats.NaN64A, floats.NaN64B},
	}
	for _, pair := range pairs {
		t.Run(
			fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", pair[0], pair[1]),
			func(t *testing.T) {
				got := req(pair[0], pair[1])
				want := (pair[0] == pair[1]) ||
					(pair[0] != pair[0]) && (pair[1] != pair[1])
				if got != want {
					t.Errorf("got %t", got)
				}
			},
		)
	}

	// Incomparable cases:
	s1, s2 := []int{}, []int{1, 2}
	// Only one non-nil incomparable value:
	pairs = [][2]any{
		{nil, s1},
		{s1, nil},
		{s1, 1},
		{1, s1},
		{nil, s2},
		{s2, nil},
		{s2, 1},
		{1, s2},
	}
	for _, pair := range pairs {
		t.Run(
			fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", pair[0], pair[1]),
			func(t *testing.T) {
				got := req(pair[0], pair[1])
				if got {
					t.Errorf("got %t", got)
				}
			},
		)
	}
	// Two non-nil incomparable values:
	pairs = [][2]any{
		{s1, s1},
		{s2, s2},
		{s1, s2},
	}
	for _, pair := range pairs {
		t.Run(
			fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", pair[0], pair[1]),
			func(t *testing.T) {
				got := req(pair[0], pair[1])
				if !got {
					t.Errorf("got %t", got)
				}
			},
		)
	}
	// Incomparable type nil value, and nil any:
	var s []int
	pairs = [][2]any{
		{nil, s},
		{s, nil},
	}
	for _, pair := range pairs {
		t.Run(
			fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", pair[0], pair[1]),
			func(t *testing.T) {
				got := req(pair[0], pair[1])
				if !got {
					t.Errorf("got %t", got)
				}
			},
		)
	}
}

func TestEqualFunc_Reflexive_Nil(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()
	req := compare.EqualFunc[float64](nil).Reflexive()
	if req != nil {
		t.Error("got non-nil EqualFunc")
	}
}

func TestEqual(t *testing.T) {
	subtestEqual(t, "type=int", []int{1, 2, 3})
	subtestEqual(t, "type=float64", []float64{
		1., 2., 3.,
		0., floats.SmallestNonzeroFloat64, floats.MaxFloat64, floats.Inf64,
		-floats.SmallestNonzeroFloat64, -floats.MaxFloat64, floats.NegInf64,
	})
	subtestEqual(t, "type=string", []string{"1", "2", "3"})

	subtestPairs(
		t,
		"type=float64&NaN&-0.0",
		compare.Equal,
		[][2]float64{
			{0., floats.NegZero64},
		},
		[][2]float64{
			{floats.NaN64A, floats.NaN64A},
			{floats.NaN64A, floats.NaN64B},
			{floats.NaN64B, floats.NaN64B},
			{floats.NaN64A, 0.},
			{floats.NaN64A, floats.NegZero64},
			{floats.NaN64A, floats.Inf64},
			{floats.NaN64A, floats.NegInf64},
		},
	)
}

func subtestEqual[T comparable](t *testing.T, name string, data []T) {
	eqGroups := make([][]T, len(data))
	for i := range eqGroups {
		eqGroups[i] = []T{data[i]}
	}
	eqPairs, neqPairs := mkEqNeqPairs(eqGroups, 0, 0)
	subtestPairs(t, name, compare.Equal, eqPairs, neqPairs)
}

func TestReflexiveEqual(t *testing.T) {
	subtestReflexiveEqual(t, "type=int", [][]int{{1}, {2}, {3}})
	subtestReflexiveEqual(t, "type=string", [][]string{{"1"}, {"2"}, {"3"}})
	subtestReflexiveEqual(t, "type=float32", [][]float32{
		{1.}, {2.},
		{floats.NaN32A, floats.NaN32B},
		{0., floats.NegZero32},
		{floats.SmallestNonzeroFloat32}, {floats.MaxFloat32},
		{floats.Inf32},
		{-floats.SmallestNonzeroFloat32}, {-floats.MaxFloat32},
		{floats.NegInf32},
	})
	subtestReflexiveEqual(t, "type=float64", [][]float64{
		{1.}, {2.},
		{floats.NaN64A, floats.NaN64B},
		{0., floats.NegZero64},
		{floats.SmallestNonzeroFloat64}, {floats.MaxFloat64},
		{floats.Inf64},
		{-floats.SmallestNonzeroFloat64}, {-floats.MaxFloat64},
		{floats.NegInf64},
	})
}

func subtestReflexiveEqual[T comparable](
	t *testing.T,
	name string,
	eqGroups [][]T,
) {
	eqPairs, neqPairs := mkEqNeqPairs(eqGroups, 0, 0)
	subtestPairs(t, name, compare.ReflexiveEqual, eqPairs, neqPairs)
}

func TestFloatEqual(t *testing.T) {
	subtestFloatEqual(t, "type=float32", [][]float32{
		{1.}, {2.},
		{floats.NaN32A, floats.NaN32B},
		{0., floats.NegZero32},
		{floats.SmallestNonzeroFloat32}, {floats.MaxFloat32},
		{floats.Inf32},
		{-floats.SmallestNonzeroFloat32}, {-floats.MaxFloat32},
		{floats.NegInf32},
	})
	subtestFloatEqual(t, "type=float64", [][]float64{
		{1.}, {2.},
		{floats.NaN64A, floats.NaN64B},
		{0., floats.NegZero64},
		{floats.SmallestNonzeroFloat64}, {floats.MaxFloat64},
		{floats.Inf64},
		{-floats.SmallestNonzeroFloat64}, {-floats.MaxFloat64},
		{floats.NegInf64},
	})
}

func subtestFloatEqual[T constraints.Float](
	t *testing.T,
	name string,
	eqGroups [][]T,
) {
	eqPairs, neqPairs := mkEqNeqPairs(eqGroups, 0, 0)
	subtestPairs(t, name, compare.FloatEqual, eqPairs, neqPairs)
}

var (
	float64sNonemptyEqGroups        [][][]float64
	float64sWithNaNNonemptyEqGroups [][][]float64

	intsEqPairs, intsNeqPairs                       [][2][]int
	float64sEqPairs, float64sNeqPairs               [][2][]float64
	float64sWithNaNEqPairs, float64sWithNaNNeqPairs [][2][]float64
	stringsEqPairs, stringsNeqPairs                 [][2][]string
	anySliceEqPairs, anySliceNeqPairs               [][2][]any
)

func init() {
	float64sNonemptyEqGroups = [][][]float64{
		{{1., floats.Inf64, 3., 4.}},     // even length - 1
		{{1., floats.Inf64, 2., 4.}},     // even length - 2
		{{1., floats.Inf64, 3., 4., 5.}}, // odd length - 1
		{{1., floats.Inf64, 2., 4., 5.}}, // odd length - 2
	}
	float64sWithNaNNonemptyEqGroups = [][][]float64{
		{{1., floats.Inf64, 3., 4.}},                                                  // even length - 1
		{{1., floats.Inf64, 2., 4.}},                                                  // even length - 2
		{{1., floats.Inf64, 3., floats.NaN64A}},                                       // even length - 3
		{{floats.NaN64A, floats.NaN64A, floats.NaN64A, floats.NaN64A}},                // even length - 4
		{{1., floats.Inf64, 3., 4., 5.}},                                              // odd length - 1
		{{1., floats.Inf64, 2., 4., 5.}},                                              // odd length - 2
		{{1., floats.Inf64, 3., floats.NaN64A, 5.}},                                   // odd length - 3
		{{floats.NaN64A, floats.NaN64A, floats.NaN64A, floats.NaN64A, floats.NaN64A}}, // odd length - 4
	}

	intsEqPairs, intsNeqPairs = mkEqNeqPairs([][][]int{
		{nil},             // nil
		{{}},              // empty
		{{1, 2, 3, 4}},    // even length - 1
		{{1, 2, 2, 4}},    // even length - 2
		{{1, 2, 3, 4, 5}}, // odd length - 1
		{{1, 2, 2, 4, 5}}, // odd length - 2
	}, 0, 0)

	float64sEqPairs, float64sNeqPairs = mkEqNeqPairs(append([][][]float64{
		{nil}, // nil
		{{}},  // empty
	}, float64sNonemptyEqGroups...), 0, 0)
	float64sWithNaNEqPairs, float64sWithNaNNeqPairs = mkEqNeqPairs(
		append([][][]float64{
			{nil}, // nil
			{{}},  // empty
		}, float64sWithNaNNonemptyEqGroups...),
		0,
		0,
	)

	stringsEqPairs, stringsNeqPairs = mkEqNeqPairs([][][]string{
		{nil},                       // nil
		{{}},                        // empty
		{{"1", "2", "3", "4"}},      // even length - 1
		{{"1", "2", "2", "4"}},      // even length - 2
		{{"1", "2", "3", "4", "5"}}, // odd length - 1
		{{"1", "2", "2", "4", "5"}}, // odd length - 2
	}, 0, 0)

	anySliceEqPairs, anySliceNeqPairs = mkEqNeqPairs([][][]any{
		{nil},                  // nil
		{{}},                   // empty
		{{1, 2, 3, 4}},         // even length - 1
		{{1., 2., 3., 4.}},     // even length - 2
		{{1, 2, 2, 4}},         // even length - 3
		{{1., 2., 2., 4.}},     // even length - 4
		{{1, 2, 3, 4, 5}},      // odd length - 1
		{{1., 2., 3., 4., 5.}}, // odd length - 2
		{{1, 2, 2, 4, 5}},      // odd length - 3
		{{1., 2., 2., 4., 5.}}, // odd length - 4
	}, 1, 3)
	anySliceEqPairs = append(anySliceEqPairs, [2][]any{
		{1, 2., '3', byte('4')},
		{1, 2., '3', byte('4')},
	})
	anySliceNeqPairs = append(anySliceNeqPairs, [2][]any{
		// It should regard as unequal since []int is not comparable.
		{[]int{1, 2, 3}},
		{[]int{1, 2, 3}},
	}, [2][]any{
		// 2. (type: float64) != 2 (type: int).
		{1, 2., '3', byte('4')},
		{1, 2, '3', byte('4')},
	}, [2][]any{
		// NaN != NaN.
		{floats.NaN64A},
		{floats.NaN64A},
	})
}

func TestSliceEqual(t *testing.T) {
	subtestSliceEqual(t, "type=[]int", intsEqPairs, intsNeqPairs)
	subtestSliceEqual(t, "type=[]float64", float64sEqPairs, float64sNeqPairs)
	subtestSliceEqual(t, "type=[]string", stringsEqPairs, stringsNeqPairs)

	subtestSliceEqual(t, "type=[]float64&NaN", nil, [][2][]float64{
		{{floats.NaN64A}, {floats.NaN64A}},
		{{floats.NaN64A, floats.NaN64A}, {floats.NaN64A, floats.NaN64A}},
		{{1., 2., floats.NaN64A}, {1., 2., floats.NaN64A}},
		{{floats.NaN64A, 1.}, {floats.NaN64A, 1.}},
		{{floats.Inf64, floats.NaN64A}, {floats.Inf64, floats.NaN64A}},
	})
}

func subtestSliceEqual[T comparable](
	t *testing.T, name string, eqPairs, neqPairs [][2][]T) {
	subtestPairs(t, name, compare.SliceEqual, eqPairs, neqPairs)
}

func TestFloatSliceEqual(t *testing.T) {
	f32EqPairs := make([][2][]float32, len(float64sWithNaNEqPairs))
	for i := range float64sWithNaNEqPairs {
		for j := range float64sWithNaNEqPairs[i] {
			if float64sWithNaNEqPairs[i][j] == nil {
				continue
			}
			f32EqPairs[i][j] = make(
				[]float32, len(float64sWithNaNEqPairs[i][j]))
			for k := range float64sWithNaNEqPairs[i][j] {
				f32EqPairs[i][j][k] = float32(float64sWithNaNEqPairs[i][j][k])
			}
		}
	}
	f32NeqPairs := make([][2][]float32, len(float64sWithNaNNeqPairs))
	for i := range float64sWithNaNNeqPairs {
		for j := range float64sWithNaNNeqPairs[i] {
			if float64sWithNaNNeqPairs[i][j] == nil {
				continue
			}
			f32NeqPairs[i][j] = make(
				[]float32, len(float64sWithNaNNeqPairs[i][j]))
			for k := range float64sWithNaNNeqPairs[i][j] {
				f32NeqPairs[i][j][k] = float32(float64sWithNaNNeqPairs[i][j][k])
			}
		}
	}
	subtestFloatSliceEqual(t, "type=[]float32", f32EqPairs, f32NeqPairs)
	subtestFloatSliceEqual(
		t, "type=[]float64", float64sWithNaNEqPairs, float64sWithNaNNeqPairs)
}

func subtestFloatSliceEqual[T constraints.Float](
	t *testing.T, name string, eqPairs, neqPairs [][2][]T) {
	subtestPairs(t, name, compare.FloatSliceEqual, eqPairs, neqPairs)
}

func TestAnySliceEqual(t *testing.T) {
	subtestAnySliceEqual(t, "type=[]int", intsEqPairs, intsNeqPairs)
	subtestAnySliceEqual(t, "type=[]float64", float64sEqPairs, float64sNeqPairs)
	subtestAnySliceEqual(t, "type=[]string", stringsEqPairs, stringsNeqPairs)
	subtestAnySliceEqual(t, "type=[]any", anySliceEqPairs, anySliceNeqPairs)
}

func subtestAnySliceEqual[T any](
	t *testing.T,
	name string,
	eqPairs [][2][]T,
	neqPairs [][2][]T,
) {
	subtestPairs(t, name, compare.AnySliceEqual, eqPairs, neqPairs)
}

func TestEqualToSliceEqual_FloatEqual_Float64(t *testing.T) {
	for _, nilEqToEmpty := range []bool{false, true} {
		toSlice := compare.EqualToSliceEqual[[]float64](
			compare.FloatEqual, nilEqToEmpty)
		var eqPairs, neqPairs [][2][]float64
		if nilEqToEmpty {
			eqPairs, neqPairs = mkEqNeqPairs(append([][][]float64{
				{nil, {}}, // nil and empty
			}, float64sWithNaNNonemptyEqGroups...), 0, 0)
		} else {
			eqPairs, neqPairs = float64sWithNaNEqPairs, float64sWithNaNNeqPairs
		}
		subtestPairs(t, fmt.Sprintf("nilEqualsEmpty=%t", nilEqToEmpty),
			toSlice, eqPairs, neqPairs)
	}
}

func TestEqualToSliceEqual_NilEf_Float64(t *testing.T) {
	for _, nilEqToEmpty := range []bool{false, true} {
		toSlice := compare.EqualToSliceEqual[[]float64](nil, nilEqToEmpty)
		var eqPairs, neqPairs [][2][]float64
		if nilEqToEmpty {
			eqPairs, neqPairs = mkEqNeqPairs(append([][][]float64{
				{nil, {}}, // nil and empty
			}, float64sNonemptyEqGroups...), 0, 0)
		} else {
			eqPairs, neqPairs = float64sEqPairs, float64sNeqPairs
		}
		subtestPairs(t, fmt.Sprintf("nilEqualsEmpty=%t", nilEqToEmpty),
			toSlice, eqPairs, neqPairs)
	}
}

func TestSliceEqualWithoutOrder(t *testing.T) {
	intsEqWithoutOrderPairs, intsNeqWithoutOrderPairs := mkEqNeqPairs(
		[][][]int{
			{nil},
			{{}},
			{{1, 1}},
			{{1, 2}, {2, 1}},
			{{1, 1, 1}},
			{{1, 1, 2}, {1, 2, 1}, {2, 1, 1}},
			{{1, 2, 2}, {2, 1, 2}, {2, 2, 1}},
		},
		0,
		0,
	)
	float64sEqWithoutOrderPairs, float64sNeqWithoutOrderPairs := mkEqNeqPairs(
		[][][]float64{
			{nil},
			{{}},
			{{1., 1.}},
			{{1., 2.}, {2., 1.}},
			{{1., 1., 1.}},
			{{1., 1., 2.}, {1., 2., 1.}, {2., 1., 1.}},
			{
				{1., floats.Inf64, floats.Inf64},
				{floats.Inf64, 1., floats.Inf64},
				{floats.Inf64, floats.Inf64, 1.},
			},
		},
		0,
		0,
	)
	stringsEqWithoutOrderPairs, stringsNeqWithoutOrderPairs := mkEqNeqPairs(
		[][][]string{
			{nil},
			{{}},
			{{"1", "1"}},
			{{"1", "2"}, {"2", "1"}},
			{{"1", "1", "1"}},
			{{"1", "1", "2"}, {"1", "2", "1"}, {"2", "1", "1"}},
			{{"1", "2", "2"}, {"2", "1", "2"}, {"2", "2", "1"}},
		},
		0,
		0,
	)

	subtestSliceEqualWithoutOrder(
		t,
		"type=[]int",
		intsEqWithoutOrderPairs,
		intsNeqWithoutOrderPairs,
	)
	subtestSliceEqualWithoutOrder(
		t,
		"type=[]float64",
		float64sEqWithoutOrderPairs,
		float64sNeqWithoutOrderPairs,
	)
	subtestSliceEqualWithoutOrder(
		t,
		"type=[]string",
		stringsEqWithoutOrderPairs,
		stringsNeqWithoutOrderPairs,
	)

	subtestSliceEqualWithoutOrder(t, "type=[]float64&NaN", nil, [][2][]float64{
		{{floats.NaN64A}, {floats.NaN64A}},
		{{floats.NaN64A, floats.NaN64A}, {floats.NaN64A, floats.NaN64A}},
		{{1., 1., floats.NaN64A}, {1., 1., floats.NaN64A}},
		{{1., 1., floats.NaN64A}, {1., floats.NaN64A, 1.}},
		{{floats.NaN64A, 1.}, {floats.NaN64A, 1.}},
		{{floats.NaN64A, 1.}, {1., floats.NaN64A}},
		{{floats.Inf64, floats.NaN64A}, {floats.Inf64, floats.NaN64A}},
		{{floats.Inf64, floats.NaN64A}, {floats.NaN64A, floats.Inf64}},
	})
}

func subtestSliceEqualWithoutOrder[T comparable](
	t *testing.T,
	name string,
	eqPairs [][2][]T,
	neqPairs [][2][]T,
) {
	subtestPairs(t, name, compare.SliceEqualWithoutOrder, eqPairs, neqPairs)
}

func TestFloatSliceEqualWithoutOrder(t *testing.T) {
	f64EqPairs, f64NeqPairs := mkEqNeqPairs([][][]float64{
		{nil},
		{{}},
		{{1.}},
		{{floats.Inf64}},
		{{floats.NaN64A}},
		{{1., 1.}},
		{{1., 2.}, {2., 1.}},
		{{1., floats.NaN64A}, {floats.NaN64A, 1.}},
		{{floats.NaN64A, floats.NaN64A}},
		{{1., 1., 1.}},
		{{1., 1., 2.}, {1., 2., 1.}, {2., 1., 1.}},
		{
			{1., floats.Inf64, floats.Inf64},
			{floats.Inf64, 1., floats.Inf64},
			{floats.Inf64, floats.Inf64, 1.},
		},
		{
			{floats.Inf64, floats.Inf64, floats.NaN64A},
			{floats.Inf64, floats.NaN64A, floats.Inf64},
			{floats.NaN64A, floats.Inf64, floats.Inf64},
		},
		{
			{1., floats.NaN64A, floats.NaN64A},
			{floats.NaN64A, 1., floats.NaN64A},
			{floats.NaN64A, floats.NaN64A, 1.},
		},
		{{floats.NaN64A, floats.NaN64A, floats.NaN64A}},
	}, 0, 0)
	f32EqPairs := make([][2][]float32, len(f64EqPairs))
	for i := range f64EqPairs {
		for j := range f64EqPairs[i] {
			if f64EqPairs[i][j] == nil {
				continue
			}
			f32EqPairs[i][j] = make([]float32, len(f64EqPairs[i][j]))
			for k := range f64EqPairs[i][j] {
				f32EqPairs[i][j][k] = float32(f64EqPairs[i][j][k])
			}
		}
	}
	f32NeqPairs := make([][2][]float32, len(f64NeqPairs))
	for i := range f64NeqPairs {
		for j := range f64NeqPairs[i] {
			if f64NeqPairs[i][j] == nil {
				continue
			}
			f32NeqPairs[i][j] = make([]float32, len(f64NeqPairs[i][j]))
			for k := range f64NeqPairs[i][j] {
				f32NeqPairs[i][j][k] = float32(f64NeqPairs[i][j][k])
			}
		}
	}
	subtestFloatSliceEqualWithoutOrder(
		t, "type=[]float32", f32EqPairs, f32NeqPairs)
	subtestFloatSliceEqualWithoutOrder(
		t, "type=[]float64", f64EqPairs, f64NeqPairs)
}

func subtestFloatSliceEqualWithoutOrder[T constraints.Float](
	t *testing.T,
	name string,
	eqPairs [][2][]T,
	neqPairs [][2][]T,
) {
	subtestPairs(
		t, name, compare.FloatSliceEqualWithoutOrder, eqPairs, neqPairs)
}

var (
	stringToFloat64NonemptyEqGroups        [][]map[string]float64
	stringToFloat64WithNaNNonemptyEqGroups [][]map[string]float64

	stringToIntEqPairs, stringToIntNeqPairs                       [][2]map[string]int
	stringToFloat64EqPairs, stringToFloat64NeqPairs               [][2]map[string]float64
	stringToFloat64WithNaNEqPairs, stringToFloat64WithNaNNeqPairs [][2]map[string]float64
	stringToStringEqPairs, stringToStringNeqPairs                 [][2]map[string]string
	stringToAnyEqPairs, stringToAnyNeqPairs                       [][2]map[string]any
)

func init() {
	stringToFloat64NonemptyEqGroups = [][]map[string]float64{
		{{"": 0.}},
		{{"A": 1.}},
		{{"A": 2.}},
		{{"A": floats.Inf64}},
		{{"B": 1.}},
		{{"A": 1., "B": 1.}},
		{{"A": 1., "B": 2.}},
		{{"A": 1., "B": floats.Inf64}},
		{{"A": floats.Inf64, "B": floats.Inf64}},
		{{"A": 1., "B": 1., "C": 1.}},
		{{"A": 1., "B": 1., "C": floats.NegInf64}},
		{{"A": 1., "B": floats.Inf64, "C": 1.}},
		{{"A": 1., "B": floats.Inf64, "C": floats.NegInf64}},
	}
	stringToFloat64WithNaNNonemptyEqGroups = [][]map[string]float64{
		{{"": 0.}},
		{{"A": 1.}},
		{{"A": 2.}},
		{{"A": floats.Inf64}},
		{{"A": floats.NaN64A}},
		{{"B": 1.}},
		{{"A": 1., "B": 1.}},
		{{"A": 1., "B": 2.}},
		{{"A": 1., "B": floats.Inf64}},
		{{"A": floats.Inf64, "B": floats.Inf64}},
		{{"A": 1., "B": floats.NaN64A}},
		{{"A": floats.NaN64A, "B": floats.NaN64A}},
		{{"A": 1., "B": 1., "C": 1.}},
		{{"A": 1., "B": 1., "C": floats.NegInf64}},
		{{"A": 1., "B": 1., "C": floats.NaN64A}},
		{{"A": 1., "B": floats.Inf64, "C": 1.}},
		{{"A": 1., "B": floats.Inf64, "C": floats.NegInf64}},
		{{"A": 1., "B": floats.Inf64, "C": floats.NaN64A}},
		{{"A": 1., "B": floats.NaN64A, "C": 1.}},
		{{"A": 1., "B": floats.NaN64A, "C": floats.NegInf64}},
		{{"A": 1., "B": floats.NaN64A, "C": floats.NaN64A}},
		{{"A": floats.NaN64A, "B": floats.NaN64A, "C": floats.NaN64A}},
	}

	stringToIntEqPairs, stringToIntNeqPairs = mkEqNeqPairs([][]map[string]int{
		{nil},
		{{}},
		{{"": 0}},
		{{"A": 1}},
		{{"A": 2}},
		{{"B": 1}},
		{{"A": 1, "B": 1}},
		{{"A": 1, "B": 2}},
	}, 0, 0)

	stringToFloat64EqPairs, stringToFloat64NeqPairs = mkEqNeqPairs(
		append([][]map[string]float64{
			{nil}, {{}},
		}, stringToFloat64NonemptyEqGroups...),
		0,
		0,
	)
	stringToFloat64WithNaNEqPairs, stringToFloat64WithNaNNeqPairs = mkEqNeqPairs(
		append([][]map[string]float64{
			{nil}, {{}},
		}, stringToFloat64WithNaNNonemptyEqGroups...),
		0,
		0,
	)

	stringToStringEqPairs, stringToStringNeqPairs = mkEqNeqPairs(
		[][]map[string]string{
			{nil},
			{{}},
			{{"": ""}},
			{{"A": "1"}},
			{{"A": "2"}},
			{{"B": "1"}},
			{{"A": "1", "B": "1"}},
			{{"A": "1", "B": "2"}},
		},
		0,
		0,
	)

	stringToAnyEqPairs, stringToAnyNeqPairs = mkEqNeqPairs([][]map[string]any{
		{nil},
		{{}},
		{{"": nil}},
		{{"A": 1}},
		{{"A": 1.}},
		{{"A": 2}},
		{{"A": 2.}},
		{{"B": 1}},
		{{"B": 1.}},
		{{"A": 1, "B": 1}},
		{{"A": 1., "B": 1.}},
		{{"A": 1, "B": 2}},
		{{"A": 1., "B": 2.}},
	}, 1, 3)
	stringToAnyEqPairs = append(stringToAnyEqPairs, [2]map[string]any{
		{"A": 1, "B": 2., "C": '3', "D": byte('4')},
		{"A": 1, "B": 2., "C": '3', "D": byte('4')},
	})
	stringToAnyNeqPairs = append(stringToAnyNeqPairs, [2]map[string]any{
		// It should regard as unequal since []int is not comparable.
		{"A": []int{1, 2, 3}},
		{"A": []int{1, 2, 3}},
	}, [2]map[string]any{
		// 2. (type: float64) != 2 (type: int).
		{"A": 1, "B": 2., "C": '3', "D": byte('4')},
		{"A": 1, "B": 2, "C": '3', "D": byte('4')},
	}, [2]map[string]any{
		// NaN != NaN.
		{"A": floats.NaN64A},
		{"A": floats.NaN64A},
	})
}

func TestMapEqual(t *testing.T) {
	subtestMapEqual(
		t,
		"type=map[string]int",
		stringToIntEqPairs,
		stringToIntNeqPairs,
	)
	subtestMapEqual(
		t,
		"type=map[string]float64",
		stringToFloat64EqPairs,
		stringToFloat64NeqPairs,
	)
	subtestMapEqual(
		t,
		"type=map[string]string",
		stringToStringEqPairs,
		stringToStringNeqPairs,
	)

	subtestMapEqual(
		t,
		"type=map[string]float64&NaN",
		nil,
		[][2]map[string]float64{
			{{"A": floats.NaN64A}, {"A": floats.NaN64A}},
			{{"A": floats.NaN64A, "B": floats.NaN64A}, {"A": floats.NaN64A, "B": floats.NaN64A}},
			{{"A": 1., "B": floats.NaN64A}, {"A": 1., "B": floats.NaN64A}},
			{{"A": floats.Inf64, "B": floats.NaN64A}, {"A": floats.Inf64, "B": floats.NaN64A}},
			{{"A": floats.NaN64A, "B": floats.Inf64}, {"A": floats.NaN64A, "B": floats.Inf64}},
		},
	)
}

func subtestMapEqual[K, V comparable](
	t *testing.T,
	name string,
	eqPairs [][2]map[K]V,
	neqPairs [][2]map[K]V,
) {
	subtestPairs(t, name, compare.MapEqual, eqPairs, neqPairs)
}

func TestFloatValueMapEqual(t *testing.T) {
	f32EqPairs := make([][2]map[string]float32,
		len(stringToFloat64WithNaNEqPairs))
	for i := range stringToFloat64WithNaNEqPairs {
		for j := range stringToFloat64WithNaNEqPairs[i] {
			if stringToFloat64WithNaNEqPairs[i][j] == nil {
				continue
			}
			f32EqPairs[i][j] = make(map[string]float32,
				len(stringToFloat64WithNaNEqPairs[i][j]))
			for k, v := range stringToFloat64WithNaNEqPairs[i][j] {
				f32EqPairs[i][j][k] = float32(v)
			}
		}
	}
	f32NeqPairs := make([][2]map[string]float32,
		len(stringToFloat64WithNaNNeqPairs))
	for i := range stringToFloat64WithNaNNeqPairs {
		for j := range stringToFloat64WithNaNNeqPairs[i] {
			if stringToFloat64WithNaNNeqPairs[i][j] == nil {
				continue
			}
			f32NeqPairs[i][j] = make(map[string]float32,
				len(stringToFloat64WithNaNNeqPairs[i][j]))
			for k, v := range stringToFloat64WithNaNNeqPairs[i][j] {
				f32NeqPairs[i][j][k] = float32(v)
			}
		}
	}
	subtestFloatValueMapEqual(t, "type=map[string]float32",
		f32EqPairs, f32NeqPairs)
	subtestFloatValueMapEqual(t, "type=map[string]float64",
		stringToFloat64WithNaNEqPairs, stringToFloat64WithNaNNeqPairs)
}

func subtestFloatValueMapEqual[K comparable, V constraints.Float](
	t *testing.T,
	name string,
	eqPairs [][2]map[K]V,
	neqPairs [][2]map[K]V,
) {
	subtestPairs(t, name, compare.FloatValueMapEqual, eqPairs, neqPairs)
}

func TestAnyValueMapEqual(t *testing.T) {
	subtestAnyValueMapEqual(
		t,
		"type=map[string]int",
		stringToIntEqPairs,
		stringToIntNeqPairs,
	)
	subtestAnyValueMapEqual(
		t,
		"type=map[string]float64",
		stringToFloat64EqPairs,
		stringToFloat64NeqPairs,
	)
	subtestAnyValueMapEqual(
		t,
		"type=map[string]string",
		stringToStringEqPairs,
		stringToStringNeqPairs,
	)
	subtestAnyValueMapEqual(
		t,
		"type=map[string]any",
		stringToAnyEqPairs,
		stringToAnyNeqPairs,
	)
}

func subtestAnyValueMapEqual[K comparable, V any](
	t *testing.T,
	name string,
	eqPairs [][2]map[K]V,
	neqPairs [][2]map[K]V,
) {
	subtestPairs(t, name, compare.AnyValueMapEqual, eqPairs, neqPairs)
}

func TestValueEqualToMapEqual_FloatEqual_Float64(t *testing.T) {
	for _, nilEqToEmpty := range []bool{false, true} {
		toMap := compare.ValueEqualToMapEqual[map[string]float64](
			compare.FloatEqual, nilEqToEmpty)
		var eqPairs, neqPairs [][2]map[string]float64
		if nilEqToEmpty {
			eqPairs, neqPairs = mkEqNeqPairs(append([][]map[string]float64{
				{nil, {}},
			}, stringToFloat64WithNaNNonemptyEqGroups...), 0, 0)
		} else {
			eqPairs = stringToFloat64WithNaNEqPairs
			neqPairs = stringToFloat64WithNaNNeqPairs
		}
		subtestPairs(t, fmt.Sprintf("nilEqualsEmpty=%t", nilEqToEmpty),
			toMap, eqPairs, neqPairs)
	}
}

func TestValueEqualToMapEqual_NilEf_Float64(t *testing.T) {
	for _, nilEqToEmpty := range []bool{false, true} {
		toMap := compare.ValueEqualToMapEqual[map[string]float64](
			nil, nilEqToEmpty)
		var eqPairs, neqPairs [][2]map[string]float64
		if nilEqToEmpty {
			eqPairs, neqPairs = mkEqNeqPairs(append([][]map[string]float64{
				{nil, {}},
			}, stringToFloat64NonemptyEqGroups...), 0, 0)
		} else {
			eqPairs, neqPairs = stringToFloat64EqPairs, stringToFloat64NeqPairs
		}
		subtestPairs(t, fmt.Sprintf("nilEqualsEmpty=%t", nilEqToEmpty),
			toMap, eqPairs, neqPairs)
	}
}

func subtestPairs[T any](
	t *testing.T,
	name string,
	f compare.EqualFunc[T],
	eqPairs [][2]T,
	neqPairs [][2]T,
) {
	t.Run(name, func(t *testing.T) {
		for _, eqPair := range eqPairs {
			a, b := eqPair[0], eqPair[1]
			name := pairsToName(a, b)
			t.Run(name, func(t *testing.T) {
				if !f(a, b) {
					t.Error("got false")
				}
			})
			t.Run(name+"&reverse", func(t *testing.T) {
				if !f(b, a) {
					t.Error("got false")
				}
			})
		}
		for _, neqPair := range neqPairs {
			a, b := neqPair[0], neqPair[1]
			name := pairsToName(a, b)
			t.Run(name, func(t *testing.T) {
				if f(a, b) {
					t.Error("got true")
				}
			})
			t.Run(name+"&reverse", func(t *testing.T) {
				if f(b, a) {
					t.Error("got true")
				}
			})
		}
	})
}

func pairsToName[T any](a, b T) string {
	var aName, bName string
	if reflectValueIsNil(reflect.ValueOf(a)) {
		aName = fmt.Sprintf("a=<nil>(%T)", a)
	} else {
		aName = fmt.Sprintf("a=%v(%[1]T)", a)
	}
	if reflectValueIsNil(reflect.ValueOf(b)) {
		bName = fmt.Sprintf("&b=<nil>(%T)", b)
	} else {
		bName = fmt.Sprintf("&b=%v(%[1]T)", b)
	}
	return aName + bName
}

// reflectValueIsNil reports whether v can be compared with nil and v is nil.
func reflectValueIsNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return v.IsNil()
	}
	return false
}

// mkEqNeqPairs generates eqPairs and neqPairs for
// testing the prefab functions for EqualFunc.
//
// An item in eqPairs consists of two equal elements.
// An item in neqPairs consists of two unequal elements.
//
// eqGroups consist of groups of equivalent elements.
// Items in the same group are equal to each other.
// Items in the different groups are unequal to each other.
//
// eqExCap is the additional capacity of eqPairs.
// neqExCap is the additional capacity of neqPairs.
// These two parameters are useful to avoid unnecessary memory allocation
// when the caller wants to append custom data to eqPairs and neqPairs.
func mkEqNeqPairs[T any](eqGroups [][]T, eqExCap, neqExCap int) (
	eqPairs, neqPairs [][2]T) {
	if eqExCap < 0 {
		eqExCap = 0
	}
	if neqExCap < 0 {
		neqExCap = 0
	}
	gn := len(eqGroups)
	if gn == 0 {
		return
	}
	var eqPairsLen, neqPairsLen, eqIdx, neqIdx int
	for i, group := range eqGroups {
		n := len(group)
		eqPairsLen += n * (n + 1) >> 1
		for j := i + 1; j < gn; j++ {
			neqPairsLen += n * len(eqGroups[j])
		}
	}
	eqPairs = make([][2]T, eqPairsLen, eqPairsLen+eqExCap)
	neqPairs = make([][2]T, neqPairsLen, neqPairsLen+neqExCap)
	for i, group := range eqGroups {
		for j := range group {
			for k := j; k < len(group); k++ {
				eqPairs[eqIdx][0], eqPairs[eqIdx][1] = group[j], group[k]
				eqIdx++
			}
		}
		for j := i + 1; j < gn; j++ {
			for _, a := range group {
				for _, b := range eqGroups[j] {
					neqPairs[neqIdx][0], neqPairs[neqIdx][1] = a, b
					neqIdx++
				}
			}
		}
	}
	return
}
