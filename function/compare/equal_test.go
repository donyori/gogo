// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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
	"testing"

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
			if r := compare.AnyEqual(pair[0], pair[1]); r != (pair[0] == pair[1]) {
				t.Errorf("got %t", r)
			}
		})
	}
}

func TestEqualFunc_Not(t *testing.T) {
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
	neq := compare.AnyEqual.Not()
	for _, pair := range pairs {
		t.Run(fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", pair[0], pair[1]), func(t *testing.T) {
			if r := neq(pair[0], pair[1]); r != !compare.AnyEqual(pair[0], pair[1]) {
				t.Errorf("got %t", r)
			}
		})
	}
}

func TestComparableEqual(t *testing.T) {
	subtestComparableEqual(t, "type=int", []int{1, 2, 3, 4, 5})
	subtestComparableEqual(t, "type=float64", []float64{1., 2., 3., 4., 5.})
	subtestComparableEqual(t, "type=string", []string{"1", "2", "3", "4", "5"})
}

func subtestComparableEqual[T comparable](t *testing.T, name string, data []T) {
	eqGroups := make([][]T, len(data))
	for i := range eqGroups {
		eqGroups[i] = []T{data[i]}
	}
	eqPairs, neqPairs := mkEqNeqPairs(eqGroups, 0, 0)
	subtestPairs(t, name, compare.ComparableEqual[T], eqPairs, neqPairs)
}

var (
	intsEqPairs, intsNeqPairs         [][2][]int
	float64sEqPairs, float64sNeqPairs [][2][]float64
	stringsEqPairs, stringsNeqPairs   [][2][]string
)

func init() {
	intsEqPairs, intsNeqPairs = mkEqNeqPairs([][][]int{
		{nil, {}},         // Empty.
		{{1, 2, 3, 4}},    // Even length - 1.
		{{1, 2, 2, 4}},    // Even length - 2.
		{{1, 2, 3, 4, 5}}, // Odd length - 1.
		{{1, 2, 2, 4, 5}}, // Odd length - 2.
	}, 0, 0)
	float64sEqPairs, float64sNeqPairs = mkEqNeqPairs([][][]float64{
		{nil, {}},              // Empty.
		{{1., 2., 3., 4.}},     // Even length - 1.
		{{1., 2., 2., 4.}},     // Even length - 2.
		{{1., 2., 3., 4., 5.}}, // Odd length - 1.
		{{1., 2., 2., 4., 5.}}, // Odd length - 2.
	}, 0, 0)
	stringsEqPairs, stringsNeqPairs = mkEqNeqPairs([][][]string{
		{nil, {}},                   // Empty.
		{{"1", "2", "3", "4"}},      // Even length - 1.
		{{"1", "2", "2", "4"}},      // Even length - 2.
		{{"1", "2", "3", "4", "5"}}, // Odd length - 1.
		{{"1", "2", "2", "4", "5"}}, // Odd length - 2.
	}, 0, 0)
}

func TestComparableSliceEqual(t *testing.T) {
	subtestComparableSliceEqual(t, "type=[]int", intsEqPairs, intsNeqPairs)
	subtestComparableSliceEqual(t, "type=[]float64", float64sEqPairs, float64sNeqPairs)
	subtestComparableSliceEqual(t, "type=[]string", stringsEqPairs, stringsNeqPairs)
}

func subtestComparableSliceEqual[T comparable](t *testing.T, name string, eqPairs, neqPairs [][2][]T) {
	subtestPairs(t, name, compare.ComparableSliceEqual[T], eqPairs, neqPairs)
}

func TestAnySliceEqual(t *testing.T) {
	anyEqPairs, anyNeqPairs := mkEqNeqPairs([][][]any{
		{nil, {}},              // Empty.
		{{1, 2, 3, 4}},         // Even length - 1.
		{{1., 2., 3., 4.}},     // Even length - 2.
		{{1, 2, 2, 4}},         // Even length - 3.
		{{1., 2., 2., 4.}},     // Even length - 4.
		{{1, 2, 3, 4, 5}},      // Odd length - 1.
		{{1., 2., 3., 4., 5.}}, // Odd length - 2.
		{{1, 2, 2, 4, 5}},      // Odd length - 3.
		{{1., 2., 2., 4., 5.}}, // Odd length - 4.
	}, 1, 2)
	anyEqPairs = append(anyEqPairs, [2][]any{
		{1, 2., '3', byte('4')},
		{1, 2., '3', byte('4')},
	})
	anyNeqPairs = append(anyNeqPairs, [2][]any{
		// It should regard as unequal since []int is not comparable.
		{[]int{1, 2, 3}},
		{[]int{1, 2, 3}},
	}, [2][]any{
		// 2. (type: float64) != 2 (type: int).
		{1, 2., '3', byte('4')},
		{1, 2, '3', byte('4')},
	})

	subtestAnySliceEqual(t, "type=[]int", intsEqPairs, intsNeqPairs)
	subtestAnySliceEqual(t, "type=[]float64", float64sEqPairs, float64sNeqPairs)
	subtestAnySliceEqual(t, "type=[]string", stringsEqPairs, stringsNeqPairs)
	subtestAnySliceEqual(t, "type=[]any", anyEqPairs, anyNeqPairs)
}

func subtestAnySliceEqual[T any](t *testing.T, name string, eqPairs, neqPairs [][2][]T) {
	subtestPairs(t, name, compare.AnySliceEqual[T], eqPairs, neqPairs)
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
		{{1., 2., 2.}, {2., 1., 2.}, {2., 2., 1.}},
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
}

func subtestComparableSliceEqualWithoutOrder[T comparable](t *testing.T, name string, eqPairs, neqPairs [][2][]T) {
	subtestPairs(t, name, compare.ComparableSliceEqualWithoutOrder[T], eqPairs, neqPairs)
}

func subtestPairs[T any](t *testing.T, name string, f compare.EqualFunc[T], eqPairs, neqPairs [][2]T) {
	t.Run(name, func(t *testing.T) {
		for _, eqPair := range eqPairs {
			a, b := eqPair[0], eqPair[1]
			name := fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", a, b)
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
			name := fmt.Sprintf("a=%v(%[1]T)&b=%v(%[2]T)", a, b)
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
