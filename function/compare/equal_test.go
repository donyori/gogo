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
	"reflect"
	"testing"

	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/function/compare"
)

func TestAnyEqual(t *testing.T) {
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
	}
	for _, pair := range pairs {
		t.Run(fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", pair[0], pair[1]), func(t *testing.T) {
			if got := compare.AnyEqual(pair[0], pair[1]); got != (pair[0] == pair[1]) {
				t.Errorf("got %t", got)
			}
		})
	}

	s := []int{1, 2} // []int is not comparable; compare.AnyEqual(s, s) should be false
	t.Run(fmt.Sprintf("a=%v(%[1]T)&b=%[1]v(%[1]T)", s), func(t *testing.T) {
		if got := compare.AnyEqual(s, s); got {
			t.Error("got true; want false")
		}
	})
}

func TestEqualFunc_Not_AnyEqual(t *testing.T) {
	neq := compare.AnyEqual.Not()
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
	}
	for _, pair := range pairs {
		t.Run(fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", pair[0], pair[1]), func(t *testing.T) {
			if got := neq(pair[0], pair[1]); got != !compare.AnyEqual(pair[0], pair[1]) {
				t.Errorf("got %t", got)
			}
		})
	}

	s := []int{1, 2} // []int is not comparable; neq(s, s) should be true
	t.Run(fmt.Sprintf("a=%v(%[1]T)&b=%[1]v(%[1]T)", s), func(t *testing.T) {
		if got := neq(s, s); !got {
			t.Error("got false; want true")
		}
	})
}

func TestComparableEqual(t *testing.T) {
	subtestComparableEqual(t, "type=int", []int{1, 2, 3})
	subtestComparableEqual(t, "type=float64", []float64{
		1., 2., 3.,
		0., math.SmallestNonzeroFloat64, math.MaxFloat64, math.Inf(1),
		-math.SmallestNonzeroFloat64, -math.MaxFloat64, math.Inf(-1),
	})
	subtestComparableEqual(t, "type=string", []string{"1", "2", "3"})

	subtestPairs(t, "type=float64&NaN", compare.ComparableEqual[float64], nil, [][2]float64{
		{math.NaN(), math.NaN()},
		{0., math.NaN()},
		{math.NaN(), 0.},
		{math.Inf(1), math.NaN()},
		{math.NaN(), math.Inf(1)},
		{math.Inf(-1), math.NaN()},
		{math.NaN(), math.Inf(-1)},
	})
}

func subtestComparableEqual[T comparable](t *testing.T, name string, data []T) {
	eqGroups := make([][]T, len(data))
	for i := range eqGroups {
		eqGroups[i] = []T{data[i]}
	}
	eqPairs, neqPairs := mkEqNeqPairs(eqGroups, 0, 0)
	subtestPairs(t, name, compare.ComparableEqual[T], eqPairs, neqPairs)
}

func TestFloatEqual(t *testing.T) {
	subtestFloatEqual(t, "type=float32", []float32{
		1., 2., float32(math.NaN()),
		0., math.SmallestNonzeroFloat32, math.MaxFloat32, float32(math.Inf(1)),
		-math.SmallestNonzeroFloat32, -math.MaxFloat32, float32(math.Inf(-1)),
	})
	subtestFloatEqual(t, "type=float64", []float64{
		1., 2., math.NaN(),
		0., math.SmallestNonzeroFloat64, math.MaxFloat64, math.Inf(1),
		-math.SmallestNonzeroFloat64, -math.MaxFloat64, math.Inf(-1),
	})
}

func subtestFloatEqual[T constraints.Float](t *testing.T, name string, data []T) {
	eqGroups := make([][]T, len(data))
	for i := range eqGroups {
		eqGroups[i] = []T{data[i]}
	}
	eqPairs, neqPairs := mkEqNeqPairs(eqGroups, 0, 0)
	subtestPairs(t, name, compare.FloatEqual[T], eqPairs, neqPairs)
}

var (
	intsEqPairs, intsNeqPairs                       [][2][]int
	float64sEqPairs, float64sNeqPairs               [][2][]float64
	float64sWithNaNEqPairs, float64sWithNaNNeqPairs [][2][]float64
	stringsEqPairs, stringsNeqPairs                 [][2][]string
	anySliceEqPairs, anySliceNeqPairs               [][2][]any
)

func init() {
	intsEqPairs, intsNeqPairs = mkEqNeqPairs([][][]int{
		{nil, {}},         // empty
		{{1, 2, 3, 4}},    // even length - 1
		{{1, 2, 2, 4}},    // even length - 2
		{{1, 2, 3, 4, 5}}, // odd length - 1
		{{1, 2, 2, 4, 5}}, // odd length - 2
	}, 0, 0)
	float64sEqPairs, float64sNeqPairs = mkEqNeqPairs([][][]float64{
		{nil, {}},                       // empty
		{{1., math.Inf(1), 3., 4.}},     // even length - 1
		{{1., math.Inf(1), 2., 4.}},     // even length - 2
		{{1., math.Inf(1), 3., 4., 5.}}, // odd length - 1
		{{1., math.Inf(1), 2., 4., 5.}}, // odd length - 2
	}, 0, 0)
	float64sWithNaNEqPairs, float64sWithNaNNeqPairs = mkEqNeqPairs([][][]float64{
		{nil, {}},                                                      // empty
		{{1., math.Inf(1), 3., 4.}},                                    // even length - 1
		{{1., math.Inf(1), 2., 4.}},                                    // even length - 2
		{{1., math.Inf(1), 3., math.NaN()}},                            // even length - 3
		{{math.NaN(), math.NaN(), math.NaN(), math.NaN()}},             // even length - 4
		{{1., math.Inf(1), 3., 4., 5.}},                                // odd length - 1
		{{1., math.Inf(1), 2., 4., 5.}},                                // odd length - 2
		{{1., math.Inf(1), 3., math.NaN(), 5.}},                        // odd length - 3
		{{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}}, // odd length - 4
	}, 0, 0)
	stringsEqPairs, stringsNeqPairs = mkEqNeqPairs([][][]string{
		{nil, {}},                   // empty
		{{"1", "2", "3", "4"}},      // even length - 1
		{{"1", "2", "2", "4"}},      // even length - 2
		{{"1", "2", "3", "4", "5"}}, // odd length - 1
		{{"1", "2", "2", "4", "5"}}, // odd length - 2
	}, 0, 0)
	anySliceEqPairs, anySliceNeqPairs = mkEqNeqPairs([][][]any{
		{nil, {}},              // empty
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
		{math.NaN()},
		{math.NaN()},
	})
}

func TestComparableSliceEqual(t *testing.T) {
	subtestComparableSliceEqual(t, "type=[]int", intsEqPairs, intsNeqPairs)
	subtestComparableSliceEqual(t, "type=[]float64", float64sEqPairs, float64sNeqPairs)
	subtestComparableSliceEqual(t, "type=[]string", stringsEqPairs, stringsNeqPairs)

	subtestComparableSliceEqual(t, "type=[]float64&NaN", nil, [][2][]float64{
		{{math.NaN()}, {math.NaN()}},
		{{math.NaN(), math.NaN()}, {math.NaN(), math.NaN()}},
		{{1., 2., math.NaN()}, {1., 2., math.NaN()}},
		{{math.NaN(), 1.}, {math.NaN(), 1.}},
		{{math.Inf(1), math.NaN()}, {math.Inf(1), math.NaN()}},
	})
}

func subtestComparableSliceEqual[T comparable](t *testing.T, name string, eqPairs, neqPairs [][2][]T) {
	subtestPairs(t, name, compare.ComparableSliceEqual[T], eqPairs, neqPairs)
}

func TestFloatSliceEqual(t *testing.T) {
	f32EqPairs := make([][2][]float32, len(float64sWithNaNEqPairs))
	for i := range float64sWithNaNEqPairs {
		for j := range float64sWithNaNEqPairs[i] {
			f32EqPairs[i][j] = make([]float32, len(float64sWithNaNEqPairs[i][j]))
			for k := range float64sWithNaNEqPairs[i][j] {
				f32EqPairs[i][j][k] = float32(float64sWithNaNEqPairs[i][j][k])
			}
		}
	}
	f32NeqPairs := make([][2][]float32, len(float64sWithNaNNeqPairs))
	for i := range float64sWithNaNNeqPairs {
		for j := range float64sWithNaNNeqPairs[i] {
			f32NeqPairs[i][j] = make([]float32, len(float64sWithNaNNeqPairs[i][j]))
			for k := range float64sWithNaNNeqPairs[i][j] {
				f32NeqPairs[i][j][k] = float32(float64sWithNaNNeqPairs[i][j][k])
			}
		}
	}
	subtestFloatSliceEqual(t, "type=[]float32", f32EqPairs, f32NeqPairs)
	subtestFloatSliceEqual(t, "type=[]float64", float64sWithNaNEqPairs, float64sWithNaNNeqPairs)
}

func subtestFloatSliceEqual[T constraints.Float](t *testing.T, name string, eqPairs, neqPairs [][2][]T) {
	subtestPairs(t, name, compare.FloatSliceEqual[T], eqPairs, neqPairs)
}

func TestAnySliceEqual(t *testing.T) {
	subtestAnySliceEqual(t, "type=[]int", intsEqPairs, intsNeqPairs)
	subtestAnySliceEqual(t, "type=[]float64", float64sEqPairs, float64sNeqPairs)
	subtestAnySliceEqual(t, "type=[]string", stringsEqPairs, stringsNeqPairs)
	subtestAnySliceEqual(t, "type=[]any", anySliceEqPairs, anySliceNeqPairs)
}

func subtestAnySliceEqual[T any](t *testing.T, name string, eqPairs, neqPairs [][2][]T) {
	subtestPairs(t, name, compare.AnySliceEqual[T], eqPairs, neqPairs)
}

func TestToSliceEqual_FloatEqual_Float64(t *testing.T) {
	for _, nilEqToEmpty := range []bool{false, true} {
		toSlice := compare.ToSliceEqual(compare.FloatEqual[float64], nilEqToEmpty)
		var eqPairs, neqPairs [][2][]float64
		if nilEqToEmpty {
			eqPairs, neqPairs = float64sWithNaNEqPairs, float64sWithNaNNeqPairs
		} else {
			eqPairs, neqPairs = mkEqNeqPairs([][][]float64{
				{nil},                               // nil
				{{}},                                // empty
				{{1., math.Inf(1), 3., 4.}},         // even length - 1
				{{1., math.Inf(1), 2., 4.}},         // even length - 2
				{{1., math.Inf(1), 3., math.NaN()}}, // even length - 3
				{{math.NaN(), math.NaN(), math.NaN(), math.NaN()}},             // even length - 4
				{{1., math.Inf(1), 3., 4., 5.}},                                // odd length - 1
				{{1., math.Inf(1), 2., 4., 5.}},                                // odd length - 2
				{{1., math.Inf(1), 3., math.NaN(), 5.}},                        // odd length - 3
				{{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}}, // odd length - 4
			}, 0, 0)
		}
		subtestPairs(t, fmt.Sprintf("nilEqualToEmpty=%t", nilEqToEmpty), toSlice, eqPairs, neqPairs)
	}
}

func TestToSliceEqual_NilEf_Float64(t *testing.T) {
	for _, nilEqToEmpty := range []bool{false, true} {
		toSlice := compare.ToSliceEqual[float64](nil, nilEqToEmpty)
		var eqPairs, neqPairs [][2][]float64
		if nilEqToEmpty {
			eqPairs, neqPairs = float64sEqPairs, float64sNeqPairs
		} else {
			eqPairs, neqPairs = mkEqNeqPairs([][][]float64{
				{nil},                           // nil
				{{}},                            // empty
				{{1., math.Inf(1), 3., 4.}},     // even length - 1
				{{1., math.Inf(1), 2., 4.}},     // even length - 2
				{{1., math.Inf(1), 3., 4., 5.}}, // odd length - 1
				{{1., math.Inf(1), 2., 4., 5.}}, // odd length - 2
			}, 0, 0)
		}
		subtestPairs(t, fmt.Sprintf("nilEqualToEmpty=%t", nilEqToEmpty), toSlice, eqPairs, neqPairs)
	}
}

func TestComparableSliceEqualWithoutOrder(t *testing.T) {
	intsEqWithoutOrderPairs, intsNeqWithoutOrderPairs := mkEqNeqPairs([][][]int{
		{nil, {}},
		{{1, 1}},
		{{1, 2}, {2, 1}},
		{{1, 1, 1}},
		{{1, 1, 2}, {1, 2, 1}, {2, 1, 1}},
		{{1, 2, 2}, {2, 1, 2}, {2, 2, 1}},
	}, 0, 0)
	float64sEqWithoutOrderPairs, float64sNeqWithoutOrderPairs := mkEqNeqPairs([][][]float64{
		{nil, {}},
		{{1., 1.}},
		{{1., 2.}, {2., 1.}},
		{{1., 1., 1.}},
		{{1., 1., 2.}, {1., 2., 1.}, {2., 1., 1.}},
		{
			{1., math.Inf(1), math.Inf(1)},
			{math.Inf(1), 1., math.Inf(1)},
			{math.Inf(1), math.Inf(1), 1.},
		},
	}, 0, 0)
	stringsEqWithoutOrderPairs, stringsNeqWithoutOrderPairs := mkEqNeqPairs([][][]string{
		{nil, {}},
		{{"1", "1"}},
		{{"1", "2"}, {"2", "1"}},
		{{"1", "1", "1"}},
		{{"1", "1", "2"}, {"1", "2", "1"}, {"2", "1", "1"}},
		{{"1", "2", "2"}, {"2", "1", "2"}, {"2", "2", "1"}},
	}, 0, 0)

	subtestComparableSliceEqualWithoutOrder(t, "type=[]int", intsEqWithoutOrderPairs, intsNeqWithoutOrderPairs)
	subtestComparableSliceEqualWithoutOrder(t, "type=[]float64", float64sEqWithoutOrderPairs, float64sNeqWithoutOrderPairs)
	subtestComparableSliceEqualWithoutOrder(t, "type=[]string", stringsEqWithoutOrderPairs, stringsNeqWithoutOrderPairs)

	subtestComparableSliceEqualWithoutOrder(t, "type=[]float64", nil, [][2][]float64{
		{{math.NaN()}, {math.NaN()}},
		{{math.NaN(), math.NaN()}, {math.NaN(), math.NaN()}},
		{{1., 1., math.NaN()}, {1., 1., math.NaN()}},
		{{1., 1., math.NaN()}, {1., math.NaN(), 1.}},
		{{math.NaN(), 1.}, {math.NaN(), 1.}},
		{{math.NaN(), 1.}, {1., math.NaN()}},
		{{math.Inf(1), math.NaN()}, {math.Inf(1), math.NaN()}},
		{{math.Inf(1), math.NaN()}, {math.NaN(), math.Inf(1)}},
	})
}

func subtestComparableSliceEqualWithoutOrder[T comparable](t *testing.T, name string, eqPairs, neqPairs [][2][]T) {
	subtestPairs(t, name, compare.ComparableSliceEqualWithoutOrder[T], eqPairs, neqPairs)
}

func TestFloatSliceEqualWithoutOrder(t *testing.T) {
	f64EqPairs, f64NeqPairs := mkEqNeqPairs([][][]float64{
		{nil, {}},
		{{1.}},
		{{math.Inf(1)}},
		{{math.NaN()}},
		{{1., 1.}},
		{{1., 2.}, {2., 1.}},
		{{1., math.NaN()}, {math.NaN(), 1.}},
		{{math.NaN(), math.NaN()}},
		{{1., 1., 1.}},
		{{1., 1., 2.}, {1., 2., 1.}, {2., 1., 1.}},
		{
			{1., math.Inf(1), math.Inf(1)},
			{math.Inf(1), 1., math.Inf(1)},
			{math.Inf(1), math.Inf(1), 1.},
		},
		{
			{math.Inf(1), math.Inf(1), math.NaN()},
			{math.Inf(1), math.NaN(), math.Inf(1)},
			{math.NaN(), math.Inf(1), math.Inf(1)},
		},
		{
			{1., math.NaN(), math.NaN()},
			{math.NaN(), 1., math.NaN()},
			{math.NaN(), math.NaN(), 1.},
		},
		{{math.NaN(), math.NaN(), math.NaN()}},
	}, 0, 0)
	f32EqPairs := make([][2][]float32, len(f64EqPairs))
	for i := range f64EqPairs {
		for j := range f64EqPairs[i] {
			f32EqPairs[i][j] = make([]float32, len(f64EqPairs[i][j]))
			for k := range f64EqPairs[i][j] {
				f32EqPairs[i][j][k] = float32(f64EqPairs[i][j][k])
			}
		}
	}
	f32NeqPairs := make([][2][]float32, len(f64NeqPairs))
	for i := range f64NeqPairs {
		for j := range f64NeqPairs[i] {
			f32NeqPairs[i][j] = make([]float32, len(f64NeqPairs[i][j]))
			for k := range f64NeqPairs[i][j] {
				f32NeqPairs[i][j][k] = float32(f64NeqPairs[i][j][k])
			}
		}
	}
	subtestFloatSliceEqualWithoutOrder(t, "type=[]float32", f32EqPairs, f32NeqPairs)
	subtestFloatSliceEqualWithoutOrder(t, "type=[]float64", f64EqPairs, f64NeqPairs)
}

func subtestFloatSliceEqualWithoutOrder[T constraints.Float](t *testing.T, name string, eqPairs, neqPairs [][2][]T) {
	subtestPairs(t, name, compare.FloatSliceEqualWithoutOrder[T], eqPairs, neqPairs)
}

func subtestPairs[T any](t *testing.T, name string, f compare.EqualFunc[T], eqPairs, neqPairs [][2]T) {
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
func mkEqNeqPairs[T any](eqGroups [][]T, eqExCap, neqExCap int) (eqPairs, neqPairs [][2]T) {
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
		eqPairsLen += n * (n + 1) / 2
		for j := i + 1; j < gn; j++ {
			neqPairsLen += n * len(eqGroups[j])
		}
	}
	eqPairs = make([][2]T, eqPairsLen, eqPairsLen+eqExCap)
	neqPairs = make([][2]T, neqPairsLen, neqPairsLen+neqExCap)
	for i, group := range eqGroups {
		for j := range group {
			for k := j; k < len(group); k++ {
				eqPairs[eqIdx][0], eqPairs[eqIdx][1], eqIdx = group[j], group[k], eqIdx+1
			}
		}
		for j := i + 1; j < gn; j++ {
			for _, a := range group {
				for _, b := range eqGroups[j] {
					neqPairs[neqIdx][0], neqPairs[neqIdx][1], neqIdx = a, b, neqIdx+1
				}
			}
		}
	}
	return
}
