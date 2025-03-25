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

package array_test

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/function/compare"
	"github.com/donyori/gogo/internal/floats"
)

func TestWrapSliceFunctionsNilSlicePtr(t *testing.T) {
	t.Run(
		"WrapSlice-int,compare.OrderedLess,compare.OrderedCompare",
		func(t *testing.T) {
			testWrapSliceFunctionsNilSlicePtr(
				t,
				func() array.OrderedDynamicArray[int] {
					return array.WrapSlice[int](
						nil,
						compare.OrderedLess,
						compare.OrderedCompare,
					)
				},
			)
		},
	)

	t.Run(
		"WrapStrictWeakOrderedSlice-int",
		func(t *testing.T) {
			testWrapSliceFunctionsNilSlicePtr(
				t,
				func() array.OrderedDynamicArray[int] {
					return array.WrapStrictWeakOrderedSlice[int](nil)
				},
			)
		},
	)

	t.Run(
		"WrapFloatSlice-float64",
		func(t *testing.T) {
			testWrapSliceFunctionsNilSlicePtr(
				t,
				func() array.OrderedDynamicArray[float64] {
					return array.WrapFloatSlice[float64](nil)
				},
			)
		},
	)
}

func testWrapSliceFunctionsNilSlicePtr[Item any](
	t *testing.T,
	getOrderedDynamicArray func() array.OrderedDynamicArray[Item],
) {
	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()
	oda := getOrderedDynamicArray()
	if oda == nil {
		t.Error("got nil OrderedDynamicArray")
	} else if oda.Len() != 0 {
		t.Error("got nonempty OrderedDynamicArray")
	}
}

func TestWrapSlice_NilLessFn(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()
	slice := []int{1, 2, 2, -1}
	n := len(slice)
	want := make([][]bool, n)
	for i := range n {
		want[i] = make([]bool, n)
		for j := range n {
			want[i][j] = slice[i] < slice[j]
		}
	}

	oda := array.WrapSlice(&slice, nil, compare.OrderedCompare)
	if oda == nil {
		t.Fatal("got nil OrderedDynamicArray")
	} else if got := oda.Len(); got != n {
		t.Fatalf("got oda.Len %d; want %d", got, n)
	}
	for i := range slice {
		for j := range slice {
			got := oda.Less(i, j)
			if got != want[i][j] {
				t.Errorf("got oda.Less(%d, %d) %t; want %t",
					i, j, got, want[i][j])
			}
		}
	}
}

func TestWrapSlice_NilCmpFn(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()
	slice := []int{1, 2, 2, -1}
	n := len(slice)
	want := make([][]int, n)
	for i := range n {
		want[i] = make([]int, n)
		for j := range n {
			if slice[i] < slice[j] {
				want[i][j] = -1
			} else if slice[i] > slice[j] {
				want[i][j] = 1
			}
		}
	}

	oda := array.WrapSlice(&slice, compare.OrderedLess, nil)
	if oda == nil {
		t.Fatal("got nil OrderedDynamicArray")
	} else if got := oda.Len(); got != n {
		t.Fatalf("got oda.Len %d; want %d", got, n)
	}
	for i := range slice {
		for j := range slice {
			got := oda.Compare(i, j)
			if got != want[i][j] {
				t.Errorf("got oda.Compare(%d, %d) %d; want %d",
					i, j, got, want[i][j])
			}
		}
	}
}

func TestWrapSlice_NilLessFnAndCmpFn(t *testing.T) {
	defer func() {
		e := recover()
		if e == nil {
			t.Error("want panic but not")
			return
		}
		s, ok := e.(string)
		if !ok || !strings.HasSuffix(s, "both lessFn and cmpFn are nil") {
			t.Error("panic -", e)
		}
	}()
	slice := []int{1, 2, 2, -1}
	array.WrapSlice(&slice, nil, nil) // want panic here
}

func TestAffectProvidedSlice(t *testing.T) {
	t.Run(
		"WrapSlice-int,compare.OrderedLess,compare.OrderedCompare",
		func(t *testing.T) {
			testAffectProvidedSlice(
				t,
				func(slicePtr *[]int) array.OrderedDynamicArray[int] {
					return array.WrapSlice(
						slicePtr,
						compare.OrderedLess,
						compare.OrderedCompare,
					)
				},
			)
		},
	)

	t.Run(
		"WrapStrictWeakOrderedSlice-int",
		func(t *testing.T) {
			testAffectProvidedSlice(
				t,
				func(slicePtr *[]int) array.OrderedDynamicArray[int] {
					return array.WrapStrictWeakOrderedSlice(slicePtr)
				},
			)
		},
	)

	t.Run(
		"WrapFloatSlice-float64",
		func(t *testing.T) {
			testAffectProvidedSlice(
				t,
				func(slicePtr *[]float64) array.OrderedDynamicArray[float64] {
					return array.WrapFloatSlice(slicePtr)
				},
			)
		},
	)
}

func testAffectProvidedSlice[Item constraints.Real](
	t *testing.T,
	getOrderedDynamicArray func(
		slicePtr *[]Item,
	) array.OrderedDynamicArray[Item],
) {
	slice := []Item{1, 2, 3, 4}
	afterCut := []Item{1, 4}
	afterSetFront := []Item{5, 4}
	afterPush := []Item{1, 2, 3}
	oda := getOrderedDynamicArray(&slice)

	// Operations on oda should affect slice.
	oda.Cut(1, 3)
	if !slices.Equal(slice, afterCut) {
		t.Fatalf("after cutting, got %v; want %v", slice, afterCut)
	}
	oda.SetFront(5)
	if !slices.Equal(slice, afterSetFront) {
		t.Fatalf("after setting front, got %v; want %v", slice, afterSetFront)
	}
	oda.Clear()
	if slice != nil {
		t.Fatalf("after clearing, got %v; want <nil>", slice)
	}

	// Operations on slice should also affect oda.
	slice = []Item{1, 2}
	oda.Push(3)
	if !slices.Equal(slice, afterPush) {
		t.Errorf("after reassigning slice and pushing, got %v; want %v",
			slice, afterPush)
	}
}

type sliceOrderedDynamicArrayTestCase[Item, Want any] struct {
	sliceName string
	slice     []Item
	want      Want
}

func TestOrderedDynamicArray_Min(t *testing.T) {
	t.Run("WrapSlice-[2]int", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[[2]int, [2]int]{
			{
				"[{1,0}]",
				[][2]int{{1, 0}},
				[2]int{1, 0},
			},
			{
				"[{1,0},{1,1}]",
				[][2]int{{1, 0}, {1, 1}},
				[2]int{1, 0},
			},
			{
				"[{1,0},{2,1}]",
				[][2]int{{1, 0}, {2, 1}},
				[2]int{1, 0},
			},
			{
				"[{2,0},{1,1}]",
				[][2]int{{2, 0}, {1, 1}},
				[2]int{1, 1},
			},
			{
				"[{2,0},{1,1},{1,2}]",
				[][2]int{{2, 0}, {1, 1}, {1, 2}},
				[2]int{1, 1},
			},
			{
				"[{2,0},{1,1},{2,2}]",
				[][2]int{{2, 0}, {1, 1}, {2, 2}},
				[2]int{1, 1},
			},
			{
				"[{2,0},{2,1},{1,2}]",
				[][2]int{{2, 0}, {2, 1}, {1, 2}},
				[2]int{1, 2},
			},
			{
				"[{3,0},{1,1},{2,2}]",
				[][2]int{{3, 0}, {1, 1}, {2, 2}},
				[2]int{1, 1},
			},
			{
				"[{3,0},{2,1},{1,2}]",
				[][2]int{{3, 0}, {2, 1}, {1, 2}},
				[2]int{1, 2},
			},
		}
		testOrderedDynamicArrayMinMax(
			t,
			func(slicePtr *[][2]int) array.OrderedDynamicArray[[2]int] {
				return array.WrapSlice(
					slicePtr,
					func(a, b [2]int) bool {
						return a[0] < b[0]
					},
					func(a, b [2]int) int {
						if a[0] < b[0] {
							return -1
						} else if a[0] > b[0] {
							return 1
						}
						return 0
					},
				)
			},
			func(oda array.OrderedDynamicArray[[2]int]) [2]int {
				return oda.Min()
			},
			testCases,
		)
	})

	t.Run("WrapStrictWeakOrderedSlice-int", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[int, int]{
			{"[1]", []int{1}, 1},
			{"[1,1]", []int{1, 1}, 1},
			{"[1,2]", []int{1, 2}, 1},
			{"[2,1]", []int{2, 1}, 1},
			{"[2,1,1]", []int{2, 1, 1}, 1},
			{"[2,1,2]", []int{2, 1, 2}, 1},
			{"[2,2,1]", []int{2, 2, 1}, 1},
			{"[3,1,2]", []int{3, 1, 2}, 1},
			{"[3,2,1]", []int{3, 2, 1}, 1},
		}
		testOrderedDynamicArrayMinMax(
			t,
			func(slicePtr *[]int) array.OrderedDynamicArray[int] {
				return array.WrapStrictWeakOrderedSlice(slicePtr)
			},
			func(oda array.OrderedDynamicArray[int]) int {
				return oda.Min()
			},
			testCases,
		)
	})

	t.Run("WrapFloatSlice-float64", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[float64, float64]{
			{"[1]", []float64{1.}, 1.},
			{"[1,1]", []float64{1., 1.}, 1.},
			{"[1,2]", []float64{1., 2.}, 1.},
			{"[2,1]", []float64{2., 1.}, 1.},
			{"[2,1,1]", []float64{2., 1., 1.}, 1.},
			{"[2,1,2]", []float64{2., 1., 2.}, 1.},
			{"[2,1,1]", []float64{2., 2., 1.}, 1.},
			{"[3,1,2]", []float64{3., 1., 2.}, 1.},
			{"[3,2,1]", []float64{3., 2., 1.}, 1.},
			{"[NaN]", []float64{floats.NaN64A}, floats.NaN64A},
			{"[NaN,NaN]", []float64{floats.NaN64A, floats.NaN64A}, floats.NaN64A},
			{"[NaN,1]", []float64{floats.NaN64A, 1.}, floats.NaN64A},
			{"[1,NaN]", []float64{1., floats.NaN64A}, floats.NaN64A},
			{"[1,NaN,NaN]", []float64{1., floats.NaN64A, floats.NaN64A}, floats.NaN64A},
			{"[2,NaN,1]", []float64{2., floats.NaN64A, 1.}, floats.NaN64A},
			{"[2,1,NaN]", []float64{2., 1., floats.NaN64A}, floats.NaN64A},
		}
		testOrderedDynamicArrayMinMax(
			t,
			func(slicePtr *[]float64) array.OrderedDynamicArray[float64] {
				return array.WrapFloatSlice(slicePtr)
			},
			func(oda array.OrderedDynamicArray[float64]) float64 {
				return oda.Min()
			},
			testCases,
		)
	})
}

func TestOrderedDynamicArray_Max(t *testing.T) {
	t.Run("WrapSlice-[2]int", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[[2]int, [2]int]{
			{
				"[{1,0}]",
				[][2]int{{1, 0}},
				[2]int{1, 0},
			},
			{
				"[{1,0},{1,1}]",
				[][2]int{{1, 0}, {1, 1}},
				[2]int{1, 0},
			},
			{
				"[{1,0},{2,1}]",
				[][2]int{{1, 0}, {2, 1}},
				[2]int{2, 1},
			},
			{
				"[{2,0},{1,1}]",
				[][2]int{{2, 0}, {1, 1}},
				[2]int{2, 0},
			},
			{
				"[{1,0},{1,1},{2,2}]",
				[][2]int{{1, 0}, {1, 1}, {2, 2}},
				[2]int{2, 2},
			},
			{
				"[{1,0},{2,1},{1,2}]",
				[][2]int{{1, 0}, {2, 1}, {1, 2}},
				[2]int{2, 1},
			},
			{
				"[{1,0},{2,1},{2,2}]",
				[][2]int{{1, 0}, {2, 1}, {2, 2}},
				[2]int{2, 1},
			},
			{
				"[{1,0},{2,1},{3,2}]",
				[][2]int{{1, 0}, {2, 1}, {3, 2}},
				[2]int{3, 2},
			},
			{
				"[{1,0},{3,1},{2,2}]",
				[][2]int{{1, 0}, {3, 1}, {2, 2}},
				[2]int{3, 1},
			},
		}
		testOrderedDynamicArrayMinMax(
			t,
			func(slicePtr *[][2]int) array.OrderedDynamicArray[[2]int] {
				return array.WrapSlice(
					slicePtr,
					func(a, b [2]int) bool {
						return a[0] < b[0]
					},
					func(a, b [2]int) int {
						if a[0] < b[0] {
							return -1
						} else if a[0] > b[0] {
							return 1
						}
						return 0
					},
				)
			},
			func(oda array.OrderedDynamicArray[[2]int]) [2]int {
				return oda.Max()
			},
			testCases,
		)
	})

	t.Run("WrapStrictWeakOrderedSlice-int", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[int, int]{
			{"[1]", []int{1}, 1},
			{"[1,1]", []int{1, 1}, 1},
			{"[1,2]", []int{1, 2}, 2},
			{"[2,1]", []int{2, 1}, 2},
			{"[1,1,2]", []int{1, 1, 2}, 2},
			{"[1,2,1]", []int{1, 2, 1}, 2},
			{"[1,2,2]", []int{1, 2, 2}, 2},
			{"[1,2,3]", []int{1, 2, 3}, 3},
			{"[1,3,2]", []int{1, 3, 2}, 3},
		}
		testOrderedDynamicArrayMinMax(
			t,
			func(slicePtr *[]int) array.OrderedDynamicArray[int] {
				return array.WrapStrictWeakOrderedSlice(slicePtr)
			},
			func(oda array.OrderedDynamicArray[int]) int {
				return oda.Max()
			},
			testCases,
		)
	})

	t.Run("WrapFloatSlice-float64", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[float64, float64]{
			{"[1]", []float64{1.}, 1.},
			{"[1,1]", []float64{1., 1.}, 1.},
			{"[1,2]", []float64{1., 2.}, 2.},
			{"[2,1]", []float64{2., 1.}, 2.},
			{"[1,1,2]", []float64{1., 1., 2.}, 2.},
			{"[1,2,1]", []float64{1., 2., 1.}, 2.},
			{"[1,2,2]", []float64{1., 2., 2.}, 2.},
			{"[1,2,3]", []float64{1., 2., 3.}, 3.},
			{"[1,3,2]", []float64{1., 3., 2.}, 3.},
			{"[NaN]", []float64{floats.NaN64A}, floats.NaN64A},
			{"[NaN,NaN]", []float64{floats.NaN64A, floats.NaN64A}, floats.NaN64A},
			{"[NaN,1]", []float64{floats.NaN64A, 1.}, 1.},
			{"[1,NaN]", []float64{1., floats.NaN64A}, 1.},
			{"[1,NaN,NaN]", []float64{1., floats.NaN64A, floats.NaN64A}, 1.},
			{"[1,NaN,2]", []float64{1., floats.NaN64A, 2.}, 2.},
			{"[1,2,NaN]", []float64{1., 2., floats.NaN64A}, 2.},
		}
		testOrderedDynamicArrayMinMax(
			t,
			func(slicePtr *[]float64) array.OrderedDynamicArray[float64] {
				return array.WrapFloatSlice(slicePtr)
			},
			func(oda array.OrderedDynamicArray[float64]) float64 {
				return oda.Max()
			},
			testCases,
		)
	})
}

func testOrderedDynamicArrayMinMax[Item comparable](
	t *testing.T,
	getOrderedDynamicArray func(
		slicePtr *[]Item,
	) array.OrderedDynamicArray[Item],
	minOrMaxMethod func(oda array.OrderedDynamicArray[Item]) Item,
	testCases []sliceOrderedDynamicArrayTestCase[Item, Item],
) {
	for _, tc := range testCases {
		t.Run("s="+tc.sliceName, func(t *testing.T) {
			s := slices.Clone(tc.slice)
			got := minOrMaxMethod(getOrderedDynamicArray(&s))
			if !compare.ReflexiveEqual(got, tc.want) {
				t.Errorf("got %#v; want %#v", got, tc.want)
			}
		})
	}
}

func TestOrderedDynamicArray_IsSorted(t *testing.T) {
	t.Run("WrapSlice-[2]int", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[[2]int, bool]{
			{
				"<nil>",
				nil,
				true,
			},
			{
				"[]",
				[][2]int{},
				true,
			},
			{
				"[{1,0}]",
				[][2]int{{1, 0}},
				true,
			},
			{
				"[{1,0},{1,1}]",
				[][2]int{{1, 0}, {1, 1}},
				true,
			},
			{
				"[{1,0},{2,1}]",
				[][2]int{{1, 0}, {2, 1}},
				true,
			},
			{
				"[{2,0},{1,1}]",
				[][2]int{{2, 0}, {1, 1}},
				false,
			},
			{
				"[{1,0},{1,1},{1,2}]",
				[][2]int{{1, 0}, {1, 1}, {1, 2}},
				true,
			},
			{
				"[{1,0},{1,1},{2,2}]",
				[][2]int{{1, 0}, {1, 1}, {2, 2}},
				true,
			},
			{
				"[{1,0},{2,1},{1,2}]",
				[][2]int{{1, 0}, {2, 1}, {1, 2}},
				false,
			},
			{
				"[{1,0},{2,1},{2,2}]",
				[][2]int{{1, 0}, {2, 1}, {2, 2}},
				true,
			},
			{
				"[{1,0},{2,1},{3,2}]",
				[][2]int{{1, 0}, {2, 1}, {3, 2}},
				true,
			},
			{
				"[{1,0},{3,1},{2,2}]",
				[][2]int{{1, 0}, {3, 1}, {2, 2}},
				false,
			},
			{
				"[{2,0},{1,1},{1,2}]",
				[][2]int{{2, 0}, {1, 1}, {1, 2}},
				false,
			},
			{
				"[{2,0},{1,1},{2,2}]",
				[][2]int{{2, 0}, {1, 1}, {2, 2}},
				false,
			},
			{
				"[{2,0},{1,1},{3,2}]",
				[][2]int{{2, 0}, {1, 1}, {3, 2}},
				false,
			},
			{
				"[{2,0},{2,1},{1,2}]",
				[][2]int{{2, 0}, {2, 1}, {1, 2}},
				false,
			},
			{
				"[{2,0},{3,1},{1,2}]",
				[][2]int{{2, 0}, {3, 1}, {1, 2}},
				false,
			},
			{
				"[{3,0},{1,1},{2,2}]",
				[][2]int{{3, 0}, {1, 1}, {2, 2}},
				false,
			},
			{
				"[{3,0},{2,1},{1,2}]",
				[][2]int{{3, 0}, {2, 1}, {1, 2}},
				false,
			},
		}
		testOrderedDynamicArrayIsSorted(
			t,
			func(slicePtr *[][2]int) array.OrderedDynamicArray[[2]int] {
				return array.WrapSlice(
					slicePtr,
					func(a, b [2]int) bool {
						return a[0] < b[0]
					},
					func(a, b [2]int) int {
						if a[0] < b[0] {
							return -1
						} else if a[0] > b[0] {
							return 1
						}
						return 0
					},
				)
			},
			testCases,
		)
	})

	t.Run("WrapStrictWeakOrderedSlice-int", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[int, bool]{
			{"<nil>", nil, true},
			{"[]", []int{}, true},
			{"[1]", []int{1}, true},
			{"[1,1]", []int{1, 1}, true},
			{"[1,2]", []int{1, 2}, true},
			{"[2,1]", []int{2, 1}, false},
			{"[1,1,1]", []int{1, 1, 1}, true},
			{"[1,1,2]", []int{1, 1, 2}, true},
			{"[1,2,1]", []int{1, 2, 1}, false},
			{"[1,2,2]", []int{1, 2, 2}, true},
			{"[1,2,3]", []int{1, 2, 3}, true},
			{"[1,3,2]", []int{1, 3, 2}, false},
			{"[2,1,1]", []int{2, 1, 1}, false},
			{"[2,1,2]", []int{2, 1, 2}, false},
			{"[2,1,3]", []int{2, 1, 3}, false},
			{"[2,2,1]", []int{2, 2, 1}, false},
			{"[2,3,1]", []int{2, 3, 1}, false},
			{"[3,1,2]", []int{3, 1, 2}, false},
			{"[3,2,1]", []int{3, 2, 1}, false},
		}
		testOrderedDynamicArrayIsSorted(
			t,
			func(slicePtr *[]int) array.OrderedDynamicArray[int] {
				return array.WrapStrictWeakOrderedSlice(slicePtr)
			},
			testCases,
		)
	})

	t.Run("WrapFloatSlice-float64", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[float64, bool]{
			{"<nil>", nil, true},
			{"[]", []float64{}, true},
			{"[1]", []float64{1.}, true},
			{"[1,1]", []float64{1., 1.}, true},
			{"[1,2]", []float64{1., 2.}, true},
			{"[2,1]", []float64{2., 1.}, false},
			{"[1,1,1]", []float64{1., 1., 1.}, true},
			{"[1,1,2]", []float64{1., 1., 2.}, true},
			{"[1,2,1]", []float64{1., 2., 1.}, false},
			{"[1,2,2]", []float64{1., 2., 2.}, true},
			{"[1,2,3]", []float64{1., 2., 3.}, true},
			{"[1,3,2]", []float64{1., 3., 2.}, false},
			{"[2,1,1]", []float64{2., 1., 1.}, false},
			{"[2,1,2]", []float64{2., 1., 2.}, false},
			{"[2,1,3]", []float64{2., 1., 3.}, false},
			{"[2,2,1]", []float64{2., 2., 1.}, false},
			{"[2,3,1]", []float64{2., 3., 1.}, false},
			{"[3,1,2]", []float64{3., 1., 2.}, false},
			{"[3,2,1]", []float64{3., 2., 1.}, false},
			{"[NaN]", []float64{floats.NaN64A}, true},
			{"[NaN,NaN]", []float64{floats.NaN64A, floats.NaN64B}, true},
			{"[NaN,1]", []float64{floats.NaN64A, 1.}, true},
			{"[1,NaN]", []float64{1., floats.NaN64A}, false},
			{
				"[NaN,-Inf,-MaxFloat64,-3.14,-1,-0.1,-SmallestNonzeroFloat64," +
					"-0,0,SmallestNonzeroFloat64,0.1,1,3.14,MaxFloat64,+Inf]",
				[]float64{
					floats.NaN64A, floats.NegInf64, -floats.MaxFloat64,
					-3.14, -1., -.1, -floats.SmallestNonzeroFloat64,
					floats.NegZero64, 0., floats.SmallestNonzeroFloat64,
					.1, 1., 3.14, floats.MaxFloat64, floats.Inf64,
				},
				true,
			},
			{
				"[-Inf,-MaxFloat64,-3.14,-1,-0.1,-SmallestNonzeroFloat64,-0," +
					"0,SmallestNonzeroFloat64,0.1,1,3.14,MaxFloat64,+Inf,NaN]",
				[]float64{
					floats.NegInf64, -floats.MaxFloat64, -3.14, -1., -.1,
					-floats.SmallestNonzeroFloat64, floats.NegZero64,
					0., floats.SmallestNonzeroFloat64, .1, 1., 3.14,
					floats.MaxFloat64, floats.Inf64, floats.NaN64A,
				},
				false,
			},
		}
		testOrderedDynamicArrayIsSorted(
			t,
			func(slicePtr *[]float64) array.OrderedDynamicArray[float64] {
				return array.WrapFloatSlice(slicePtr)
			},
			testCases,
		)
	})
}

func testOrderedDynamicArrayIsSorted[Item any](
	t *testing.T,
	getOrderedDynamicArray func(
		slicePtr *[]Item,
	) array.OrderedDynamicArray[Item],
	testCases []sliceOrderedDynamicArrayTestCase[Item, bool],
) {
	for _, tc := range testCases {
		t.Run("s="+tc.sliceName, func(t *testing.T) {
			s := slices.Clone(tc.slice)
			got := getOrderedDynamicArray(&s).IsSorted()
			if got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
		})
	}
}

func TestOrderedDynamicArray_Sort(t *testing.T) {
	t.Run("WrapSlice-[2]int", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[[2]int, [][][2]int]{
			{
				"<nil>",
				nil,
				[][][2]int{nil},
			},
			{
				"[]",
				[][2]int{},
				[][][2]int{{}},
			},
			{
				"[{1,0}]",
				[][2]int{{1, 0}},
				[][][2]int{{{1, 0}}},
			},
			{
				"[{1,0},{1,1}]",
				[][2]int{{1, 0}, {1, 1}},
				[][][2]int{
					{{1, 0}, {1, 1}},
					{{1, 1}, {1, 0}},
				},
			},
			{
				"[{1,0},{2,1}]",
				[][2]int{{1, 0}, {2, 1}},
				[][][2]int{{{1, 0}, {2, 1}}},
			},
			{
				"[{2,0},{1,1}]",
				[][2]int{{2, 0}, {1, 1}},
				[][][2]int{{{1, 1}, {2, 0}}},
			},
			{
				"[{1,0},{1,1},{1,2}]",
				[][2]int{{1, 0}, {1, 1}, {1, 2}},
				[][][2]int{
					{{1, 0}, {1, 1}, {1, 2}},
					{{1, 0}, {1, 2}, {1, 1}},
					{{1, 1}, {1, 0}, {1, 2}},
					{{1, 1}, {1, 2}, {1, 0}},
					{{1, 2}, {1, 0}, {1, 1}},
					{{1, 2}, {1, 1}, {1, 0}},
				},
			},
			{
				"[{1,0},{1,1},{2,2}]",
				[][2]int{{1, 0}, {1, 1}, {2, 2}},
				[][][2]int{
					{{1, 0}, {1, 1}, {2, 2}},
					{{1, 1}, {1, 0}, {2, 2}},
				},
			},
			{
				"[{1,0},{2,1},{1,2}]",
				[][2]int{{1, 0}, {2, 1}, {1, 2}},
				[][][2]int{
					{{1, 0}, {1, 2}, {2, 1}},
					{{1, 2}, {1, 0}, {2, 1}},
				},
			},
			{
				"[{1,0},{2,1},{2,2}]",
				[][2]int{{1, 0}, {2, 1}, {2, 2}},
				[][][2]int{
					{{1, 0}, {2, 1}, {2, 2}},
					{{1, 0}, {2, 2}, {2, 1}},
				},
			},
			{
				"[{1,0},{2,1},{3,2}]",
				[][2]int{{1, 0}, {2, 1}, {3, 2}},
				[][][2]int{{{1, 0}, {2, 1}, {3, 2}}},
			},
			{
				"[{1,0},{3,1},{2,2}]",
				[][2]int{{1, 0}, {3, 1}, {2, 2}},
				[][][2]int{{{1, 0}, {2, 2}, {3, 1}}},
			},
			{
				"[{2,0},{1,1},{1,2}]",
				[][2]int{{2, 0}, {1, 1}, {1, 2}},
				[][][2]int{
					{{1, 1}, {1, 2}, {2, 0}},
					{{1, 2}, {1, 1}, {2, 0}},
				},
			},
			{
				"[{2,0},{1,1},{2,2}]",
				[][2]int{{2, 0}, {1, 1}, {2, 2}},
				[][][2]int{
					{{1, 1}, {2, 0}, {2, 2}},
					{{1, 1}, {2, 2}, {2, 0}},
				},
			},
			{
				"[{2,0},{1,1},{3,2}]",
				[][2]int{{2, 0}, {1, 1}, {3, 2}},
				[][][2]int{{{1, 1}, {2, 0}, {3, 2}}},
			},
			{
				"[{2,0},{2,1},{1,2}]",
				[][2]int{{2, 0}, {2, 1}, {1, 2}},
				[][][2]int{
					{{1, 2}, {2, 0}, {2, 1}},
					{{1, 2}, {2, 1}, {2, 0}},
				},
			},
			{
				"[{2,0},{3,1},{1,2}]",
				[][2]int{{2, 0}, {3, 1}, {1, 2}},
				[][][2]int{{{1, 2}, {2, 0}, {3, 1}}},
			},
			{
				"[{3,0},{1,1},{2,2}]",
				[][2]int{{3, 0}, {1, 1}, {2, 2}},
				[][][2]int{{{1, 1}, {2, 2}, {3, 0}}},
			},
			{
				"[{3,0},{2,1},{1,2}]",
				[][2]int{{3, 0}, {2, 1}, {1, 2}},
				[][][2]int{{{1, 2}, {2, 1}, {3, 0}}},
			},
		}
		testOrderedDynamicArraySort(
			t,
			func(slicePtr *[][2]int) array.OrderedDynamicArray[[2]int] {
				return array.WrapSlice(
					slicePtr,
					func(a, b [2]int) bool {
						return a[0] < b[0]
					},
					func(a, b [2]int) int {
						if a[0] < b[0] {
							return -1
						} else if a[0] > b[0] {
							return 1
						}
						return 0
					},
				)
			},
			testCases,
		)
	})

	t.Run("WrapStrictWeakOrderedSlice-int", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[int, [][]int]{
			{"<nil>", nil, [][]int{nil}},
			{"[]", []int{}, [][]int{{}}},
			{"[1]", []int{1}, [][]int{{1}}},
			{"[1,1]", []int{1, 1}, [][]int{{1, 1}}},
			{"[1,2]", []int{1, 2}, [][]int{{1, 2}}},
			{"[2,1]", []int{2, 1}, [][]int{{1, 2}}},
			{"[1,1,1]", []int{1, 1, 1}, [][]int{{1, 1, 1}}},
			{"[1,1,2]", []int{1, 1, 2}, [][]int{{1, 1, 2}}},
			{"[1,2,1]", []int{1, 2, 1}, [][]int{{1, 1, 2}}},
			{"[1,2,2]", []int{1, 2, 2}, [][]int{{1, 2, 2}}},
			{"[1,2,3]", []int{1, 2, 3}, [][]int{{1, 2, 3}}},
			{"[1,3,2]", []int{1, 3, 2}, [][]int{{1, 2, 3}}},
			{"[2,1,1]", []int{2, 1, 1}, [][]int{{1, 1, 2}}},
			{"[2,1,2]", []int{2, 1, 2}, [][]int{{1, 2, 2}}},
			{"[2,1,3]", []int{2, 1, 3}, [][]int{{1, 2, 3}}},
			{"[2,2,1]", []int{2, 2, 1}, [][]int{{1, 2, 2}}},
			{"[2,3,1]", []int{2, 3, 1}, [][]int{{1, 2, 3}}},
			{"[3,1,2]", []int{3, 1, 2}, [][]int{{1, 2, 3}}},
			{"[3,2,1]", []int{3, 2, 1}, [][]int{{1, 2, 3}}},
		}
		testOrderedDynamicArraySort(
			t,
			func(slicePtr *[]int) array.OrderedDynamicArray[int] {
				return array.WrapStrictWeakOrderedSlice(slicePtr)
			},
			testCases,
		)
	})

	t.Run("WrapFloatSlice-float64", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[float64, [][]float64]{
			{"<nil>", nil, [][]float64{nil}},
			{"[]", []float64{}, [][]float64{{}}},
			{"[1]", []float64{1.}, [][]float64{{1.}}},
			{"[1,1]", []float64{1., 1.}, [][]float64{{1., 1.}}},
			{"[1,2]", []float64{1., 2.}, [][]float64{{1., 2.}}},
			{"[2,1]", []float64{2., 1.}, [][]float64{{1., 2.}}},
			{"[1,1,1]", []float64{1., 1., 1.}, [][]float64{{1., 1., 1.}}},
			{"[1,1,2]", []float64{1., 1., 2.}, [][]float64{{1., 1., 2.}}},
			{"[1,2,1]", []float64{1., 2., 1.}, [][]float64{{1., 1., 2.}}},
			{"[1,2,2]", []float64{1., 2., 2.}, [][]float64{{1., 2., 2.}}},
			{"[1,2,3]", []float64{1., 2., 3.}, [][]float64{{1., 2., 3.}}},
			{"[1,3,2]", []float64{1., 3., 2.}, [][]float64{{1., 2., 3.}}},
			{"[2,1,1]", []float64{2., 1., 1.}, [][]float64{{1., 1., 2.}}},
			{"[2,1,2]", []float64{2., 1., 2.}, [][]float64{{1., 2., 2.}}},
			{"[2,1,3]", []float64{2., 1., 3.}, [][]float64{{1., 2., 3.}}},
			{"[2,2,1]", []float64{2., 2., 1.}, [][]float64{{1., 2., 2.}}},
			{"[2,3,1]", []float64{2., 3., 1.}, [][]float64{{1., 2., 3.}}},
			{"[3,1,2]", []float64{3., 1., 2.}, [][]float64{{1., 2., 3.}}},
			{"[3,2,1]", []float64{3., 2., 1.}, [][]float64{{1., 2., 3.}}},
			{"[NaN]", []float64{floats.NaN64A}, [][]float64{{floats.NaN64A}}},
			{"[NaN,NaN]", []float64{floats.NaN64A, floats.NaN64B}, [][]float64{
				{floats.NaN64A, floats.NaN64B},
				{floats.NaN64B, floats.NaN64A},
			}},
			{"[NaN,1]", []float64{floats.NaN64A, 1.}, [][]float64{{floats.NaN64A, 1.}}},
			{"[1,NaN]", []float64{1., floats.NaN64A}, [][]float64{{floats.NaN64A, 1.}}},
			{
				"[NaN,-Inf,-MaxFloat64,-3.14,-1,-0.1,-SmallestNonzeroFloat64," +
					"-0,0,SmallestNonzeroFloat64,0.1,1,3.14,MaxFloat64,+Inf]",
				[]float64{
					floats.NaN64A, floats.NegInf64, -floats.MaxFloat64,
					-3.14, -1., -.1, -floats.SmallestNonzeroFloat64,
					floats.NegZero64, 0., floats.SmallestNonzeroFloat64,
					.1, 1., 3.14, floats.MaxFloat64, floats.Inf64,
				},
				[][]float64{
					{
						floats.NaN64A, floats.NegInf64, -floats.MaxFloat64,
						-3.14, -1., -.1, -floats.SmallestNonzeroFloat64,
						floats.NegZero64, 0., floats.SmallestNonzeroFloat64,
						.1, 1., 3.14, floats.MaxFloat64, floats.Inf64,
					},
					{
						floats.NaN64A, floats.NegInf64, -floats.MaxFloat64,
						-3.14, -1., -.1, -floats.SmallestNonzeroFloat64,
						0., floats.NegZero64, floats.SmallestNonzeroFloat64,
						.1, 1., 3.14, floats.MaxFloat64, floats.Inf64,
					},
				},
			},
			{
				"[-Inf,-MaxFloat64,-3.14,-1,-0.1,-SmallestNonzeroFloat64,-0," +
					"0,SmallestNonzeroFloat64,0.1,1,3.14,MaxFloat64,+Inf,NaN]",
				[]float64{
					floats.NegInf64, -floats.MaxFloat64, -3.14, -1., -.1,
					-floats.SmallestNonzeroFloat64, floats.NegZero64,
					0., floats.SmallestNonzeroFloat64, .1, 1., 3.14,
					floats.MaxFloat64, floats.Inf64, floats.NaN64A,
				},
				[][]float64{
					{
						floats.NaN64A, floats.NegInf64, -floats.MaxFloat64,
						-3.14, -1., -.1, -floats.SmallestNonzeroFloat64,
						floats.NegZero64, 0., floats.SmallestNonzeroFloat64,
						.1, 1., 3.14, floats.MaxFloat64, floats.Inf64,
					},
					{
						floats.NaN64A, floats.NegInf64, -floats.MaxFloat64,
						-3.14, -1., -.1, -floats.SmallestNonzeroFloat64,
						0., floats.NegZero64, floats.SmallestNonzeroFloat64,
						.1, 1., 3.14, floats.MaxFloat64, floats.Inf64,
					},
				},
			},
		}
		testOrderedDynamicArraySort(
			t,
			func(slicePtr *[]float64) array.OrderedDynamicArray[float64] {
				return array.WrapFloatSlice(slicePtr)
			},
			testCases,
		)
	})
}

func testOrderedDynamicArraySort[Item comparable](
	t *testing.T,
	getOrderedDynamicArray func(
		slicePtr *[]Item,
	) array.OrderedDynamicArray[Item],
	testCases []sliceOrderedDynamicArrayTestCase[Item, [][]Item],
) {
	eq := compare.EqualToSliceEqual[[]Item](compare.ReflexiveEqual, false)
	for _, tc := range testCases {
		t.Run("s="+tc.sliceName, func(t *testing.T) {
			s := slices.Clone(tc.slice)
			getOrderedDynamicArray(&s).Sort()
			var ok bool
			for _, acceptable := range tc.want {
				if eq(s, acceptable) {
					ok = true
					break
				}
			}
			if !ok {
				t.Errorf("got unacceptable %#v", s)
			}
		})
	}
}

func TestOrderedDynamicArray_SortStable(t *testing.T) {
	t.Run("WrapSlice-[2]int", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[[2]int, [][2]int]{
			{
				"<nil>",
				nil,
				nil,
			},
			{
				"[]",
				[][2]int{},
				[][2]int{},
			},
			{
				"[{1,0}]",
				[][2]int{{1, 0}},
				[][2]int{{1, 0}},
			},
			{
				"[{1,0},{1,1}]",
				[][2]int{{1, 0}, {1, 1}},
				[][2]int{{1, 0}, {1, 1}},
			},
			{
				"[{1,0},{2,1}]",
				[][2]int{{1, 0}, {2, 1}},
				[][2]int{{1, 0}, {2, 1}},
			},
			{
				"[{2,0},{1,1}]",
				[][2]int{{2, 0}, {1, 1}},
				[][2]int{{1, 1}, {2, 0}},
			},
			{
				"[{1,0},{1,1},{1,2}]",
				[][2]int{{1, 0}, {1, 1}, {1, 2}},
				[][2]int{{1, 0}, {1, 1}, {1, 2}},
			},
			{
				"[{1,0},{1,1},{2,2}]",
				[][2]int{{1, 0}, {1, 1}, {2, 2}},
				[][2]int{{1, 0}, {1, 1}, {2, 2}},
			},
			{
				"[{1,0},{2,1},{1,2}]",
				[][2]int{{1, 0}, {2, 1}, {1, 2}},
				[][2]int{{1, 0}, {1, 2}, {2, 1}},
			},
			{
				"[{1,0},{2,1},{2,2}]",
				[][2]int{{1, 0}, {2, 1}, {2, 2}},
				[][2]int{{1, 0}, {2, 1}, {2, 2}},
			},
			{
				"[{1,0},{2,1},{3,2}]",
				[][2]int{{1, 0}, {2, 1}, {3, 2}},
				[][2]int{{1, 0}, {2, 1}, {3, 2}},
			},
			{
				"[{1,0},{3,1},{2,2}]",
				[][2]int{{1, 0}, {3, 1}, {2, 2}},
				[][2]int{{1, 0}, {2, 2}, {3, 1}},
			},
			{
				"[{2,0},{1,1},{1,2}]",
				[][2]int{{2, 0}, {1, 1}, {1, 2}},
				[][2]int{{1, 1}, {1, 2}, {2, 0}},
			},
			{
				"[{2,0},{1,1},{2,2}]",
				[][2]int{{2, 0}, {1, 1}, {2, 2}},
				[][2]int{{1, 1}, {2, 0}, {2, 2}},
			},
			{
				"[{2,0},{1,1},{3,2}]",
				[][2]int{{2, 0}, {1, 1}, {3, 2}},
				[][2]int{{1, 1}, {2, 0}, {3, 2}},
			},
			{
				"[{2,0},{2,1},{1,2}]",
				[][2]int{{2, 0}, {2, 1}, {1, 2}},
				[][2]int{{1, 2}, {2, 0}, {2, 1}},
			},
			{
				"[{2,0},{3,1},{1,2}]",
				[][2]int{{2, 0}, {3, 1}, {1, 2}},
				[][2]int{{1, 2}, {2, 0}, {3, 1}},
			},
			{
				"[{3,0},{1,1},{2,2}]",
				[][2]int{{3, 0}, {1, 1}, {2, 2}},
				[][2]int{{1, 1}, {2, 2}, {3, 0}},
			},
			{
				"[{3,0},{2,1},{1,2}]",
				[][2]int{{3, 0}, {2, 1}, {1, 2}},
				[][2]int{{1, 2}, {2, 1}, {3, 0}},
			},
		}
		testOrderedDynamicArraySortStable(
			t,
			func(slicePtr *[][2]int) array.OrderedDynamicArray[[2]int] {
				return array.WrapSlice(
					slicePtr,
					func(a, b [2]int) bool {
						return a[0] < b[0]
					},
					func(a, b [2]int) int {
						if a[0] < b[0] {
							return -1
						} else if a[0] > b[0] {
							return 1
						}
						return 0
					},
				)
			},
			testCases,
		)
	})

	t.Run("WrapStrictWeakOrderedSlice-int", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[int, []int]{
			{"<nil>", nil, nil},
			{"[]", []int{}, []int{}},
			{"[1]", []int{1}, []int{1}},
			{"[1,1]", []int{1, 1}, []int{1, 1}},
			{"[1,2]", []int{1, 2}, []int{1, 2}},
			{"[2,1]", []int{2, 1}, []int{1, 2}},
			{"[1,1,1]", []int{1, 1, 1}, []int{1, 1, 1}},
			{"[1,1,2]", []int{1, 1, 2}, []int{1, 1, 2}},
			{"[1,2,1]", []int{1, 2, 1}, []int{1, 1, 2}},
			{"[1,2,2]", []int{1, 2, 2}, []int{1, 2, 2}},
			{"[1,2,3]", []int{1, 2, 3}, []int{1, 2, 3}},
			{"[1,3,2]", []int{1, 3, 2}, []int{1, 2, 3}},
			{"[2,1,1]", []int{2, 1, 1}, []int{1, 1, 2}},
			{"[2,1,2]", []int{2, 1, 2}, []int{1, 2, 2}},
			{"[2,1,3]", []int{2, 1, 3}, []int{1, 2, 3}},
			{"[2,2,1]", []int{2, 2, 1}, []int{1, 2, 2}},
			{"[2,3,1]", []int{2, 3, 1}, []int{1, 2, 3}},
			{"[3,1,2]", []int{3, 1, 2}, []int{1, 2, 3}},
			{"[3,2,1]", []int{3, 2, 1}, []int{1, 2, 3}},
		}
		testOrderedDynamicArraySortStable(
			t,
			func(slicePtr *[]int) array.OrderedDynamicArray[int] {
				return array.WrapStrictWeakOrderedSlice(slicePtr)
			},
			testCases,
		)
	})

	t.Run("WrapFloatSlice-float64", func(t *testing.T) {
		testCases := []sliceOrderedDynamicArrayTestCase[float64, []float64]{
			{"<nil>", nil, nil},
			{"[]", []float64{}, []float64{}},
			{"[1]", []float64{1.}, []float64{1.}},
			{"[1,1]", []float64{1., 1.}, []float64{1., 1.}},
			{"[1,2]", []float64{1., 2.}, []float64{1., 2.}},
			{"[2,1]", []float64{2., 1.}, []float64{1., 2.}},
			{"[1,1,1]", []float64{1., 1., 1.}, []float64{1., 1., 1.}},
			{"[1,1,2]", []float64{1., 1., 2.}, []float64{1., 1., 2.}},
			{"[1,2,1]", []float64{1., 2., 1.}, []float64{1., 1., 2.}},
			{"[1,2,2]", []float64{1., 2., 2.}, []float64{1., 2., 2.}},
			{"[1,2,3]", []float64{1., 2., 3.}, []float64{1., 2., 3.}},
			{"[1,3,2]", []float64{1., 3., 2.}, []float64{1., 2., 3.}},
			{"[2,1,1]", []float64{2., 1., 1.}, []float64{1., 1., 2.}},
			{"[2,1,2]", []float64{2., 1., 2.}, []float64{1., 2., 2.}},
			{"[2,1,3]", []float64{2., 1., 3.}, []float64{1., 2., 3.}},
			{"[2,2,1]", []float64{2., 2., 1.}, []float64{1., 2., 2.}},
			{"[2,3,1]", []float64{2., 3., 1.}, []float64{1., 2., 3.}},
			{"[3,1,2]", []float64{3., 1., 2.}, []float64{1., 2., 3.}},
			{"[3,2,1]", []float64{3., 2., 1.}, []float64{1., 2., 3.}},
			{"[NaN]", []float64{floats.NaN64A}, []float64{floats.NaN64A}},
			{
				"[NaN,NaN]",
				[]float64{floats.NaN64A, floats.NaN64B},
				[]float64{floats.NaN64A, floats.NaN64B},
			},
			{"[NaN,1]", []float64{floats.NaN64A, 1.}, []float64{floats.NaN64A, 1.}},
			{"[1,NaN]", []float64{1., floats.NaN64A}, []float64{floats.NaN64A, 1.}},
			{
				"[NaN,-Inf,-MaxFloat64,-3.14,-1,-0.1,-SmallestNonzeroFloat64," +
					"-0,0,SmallestNonzeroFloat64,0.1,1,3.14,MaxFloat64,+Inf]",
				[]float64{
					floats.NaN64A, floats.NegInf64, -floats.MaxFloat64,
					-3.14, -1., -.1, -floats.SmallestNonzeroFloat64,
					floats.NegZero64, 0., floats.SmallestNonzeroFloat64,
					.1, 1., 3.14, floats.MaxFloat64, floats.Inf64,
				},
				[]float64{
					floats.NaN64A, floats.NegInf64, -floats.MaxFloat64,
					-3.14, -1., -.1, -floats.SmallestNonzeroFloat64,
					floats.NegZero64, 0., floats.SmallestNonzeroFloat64,
					.1, 1., 3.14, floats.MaxFloat64, floats.Inf64,
				},
			},
			{
				"[-Inf,-MaxFloat64,-3.14,-1,-0.1,-SmallestNonzeroFloat64,-0," +
					"0,SmallestNonzeroFloat64,0.1,1,3.14,MaxFloat64,+Inf,NaN]",
				[]float64{
					floats.NegInf64, -floats.MaxFloat64, -3.14, -1., -.1,
					-floats.SmallestNonzeroFloat64, floats.NegZero64,
					0., floats.SmallestNonzeroFloat64, .1, 1., 3.14,
					floats.MaxFloat64, floats.Inf64, floats.NaN64A,
				},
				[]float64{
					floats.NaN64A, floats.NegInf64, -floats.MaxFloat64,
					-3.14, -1., -.1, -floats.SmallestNonzeroFloat64,
					floats.NegZero64, 0., floats.SmallestNonzeroFloat64,
					.1, 1., 3.14, floats.MaxFloat64, floats.Inf64,
				},
			},
		}
		testOrderedDynamicArraySortStable(
			t,
			func(slicePtr *[]float64) array.OrderedDynamicArray[float64] {
				return array.WrapFloatSlice(slicePtr)
			},
			testCases,
		)
	})
}

func testOrderedDynamicArraySortStable[Item comparable](
	t *testing.T,
	getOrderedDynamicArray func(
		slicePtr *[]Item,
	) array.OrderedDynamicArray[Item],
	testCases []sliceOrderedDynamicArrayTestCase[Item, []Item],
) {
	eq := compare.EqualToSliceEqual[[]Item](compare.ReflexiveEqual, false)
	for _, tc := range testCases {
		t.Run("s="+tc.sliceName, func(t *testing.T) {
			s := slices.Clone(tc.slice)
			getOrderedDynamicArray(&s).SortStable()
			if !eq(s, tc.want) {
				t.Errorf("got\n%#v\nwant\n%#v", s, tc.want)
			}
		})
	}
}

func TestOrderedDynamicArray_Less(t *testing.T) {
	t.Run("WrapSlice-[2]int", func(t *testing.T) {
		lessFn := func(a, b [2]int) bool {
			return a[0] < b[0]
		}
		data := [][2]int{{1, 0}, {2, 1}, {2, 2}, {3, 3}, {1, 4}, {3, 5}, {2, 6}}
		testOrderedDynamicArrayLessCompare(
			t,
			func(slicePtr *[][2]int) array.OrderedDynamicArray[[2]int] {
				return array.WrapSlice(
					slicePtr,
					lessFn,
					func(a, b [2]int) int {
						if a[0] < b[0] {
							return -1
						} else if a[0] > b[0] {
							return 1
						}
						return 0
					},
				)
			},
			func(oda array.OrderedDynamicArray[[2]int], i, j int) bool {
				return oda.Less(i, j)
			},
			lessFn,
			data,
		)
	})

	t.Run("WrapStrictWeakOrderedSlice-int", func(t *testing.T) {
		data := []int{1, 2, 2, 3, 1, 3, 2}
		testOrderedDynamicArrayLessCompare(
			t,
			func(slicePtr *[]int) array.OrderedDynamicArray[int] {
				return array.WrapStrictWeakOrderedSlice(slicePtr)
			},
			func(oda array.OrderedDynamicArray[int], i, j int) bool {
				return oda.Less(i, j)
			},
			cmp.Less,
			data,
		)
	})

	t.Run("WrapFloatSlice-float64", func(t *testing.T) {
		data := []float64{
			floats.NaN64A, floats.NegInf64, -floats.MaxFloat64, -3.14,
			-1., -.1, -floats.SmallestNonzeroFloat64, floats.NegZero64,
			0., floats.SmallestNonzeroFloat64, .1, 1., 3.14, floats.MaxFloat64,
			floats.Inf64, floats.NaN64B, 0., 0., floats.NaN64C, .1, 3.14,
			floats.NaN64D, floats.NegZero64, floats.Inf64, floats.NaN64A,
			floats.MaxFloat64, -1., -floats.SmallestNonzeroFloat64,
		}
		testOrderedDynamicArrayLessCompare(
			t,
			func(slicePtr *[]float64) array.OrderedDynamicArray[float64] {
				return array.WrapFloatSlice(slicePtr)
			},
			func(oda array.OrderedDynamicArray[float64], i, j int) bool {
				return oda.Less(i, j)
			},
			cmp.Less,
			data,
		)
	})
}

func TestOrderedDynamicArray_Compare(t *testing.T) {
	t.Run("WrapSlice-[2]int", func(t *testing.T) {
		cmpFn := func(a, b [2]int) int {
			if a[0] < b[0] {
				return -1
			} else if a[0] > b[0] {
				return 1
			}
			return 0
		}
		data := [][2]int{{1, 0}, {2, 1}, {2, 2}, {3, 3}, {1, 4}, {3, 5}, {2, 6}}
		testOrderedDynamicArrayLessCompare(
			t,
			func(slicePtr *[][2]int) array.OrderedDynamicArray[[2]int] {
				return array.WrapSlice(
					slicePtr,
					func(a, b [2]int) bool {
						return a[0] < b[0]
					},
					cmpFn,
				)
			},
			func(oda array.OrderedDynamicArray[[2]int], i, j int) int {
				return oda.Compare(i, j)
			},
			cmpFn,
			data,
		)
	})

	t.Run("WrapStrictWeakOrderedSlice-int", func(t *testing.T) {
		data := []int{1, 2, 2, 3, 1, 3, 2}
		testOrderedDynamicArrayLessCompare(
			t,
			func(slicePtr *[]int) array.OrderedDynamicArray[int] {
				return array.WrapStrictWeakOrderedSlice(slicePtr)
			},
			func(oda array.OrderedDynamicArray[int], i, j int) int {
				return oda.Compare(i, j)
			},
			cmp.Compare,
			data,
		)
	})

	t.Run("WrapFloatSlice-float64", func(t *testing.T) {
		data := []float64{
			floats.NaN64A, floats.NegInf64, -floats.MaxFloat64, -3.14,
			-1., -.1, -floats.SmallestNonzeroFloat64, floats.NegZero64,
			0., floats.SmallestNonzeroFloat64, .1, 1., 3.14, floats.MaxFloat64,
			floats.Inf64, floats.NaN64B, 0., 0., floats.NaN64C, .1, 3.14,
			floats.NaN64D, floats.NegZero64, floats.Inf64, floats.NaN64A,
			floats.MaxFloat64, -1., -floats.SmallestNonzeroFloat64,
		}
		testOrderedDynamicArrayLessCompare(
			t,
			func(slicePtr *[]float64) array.OrderedDynamicArray[float64] {
				return array.WrapFloatSlice(slicePtr)
			},
			func(oda array.OrderedDynamicArray[float64], i, j int) int {
				return oda.Compare(i, j)
			},
			cmp.Compare,
			data,
		)
	})
}

func testOrderedDynamicArrayLessCompare[Item any, Result comparable](
	t *testing.T,
	getOrderedDynamicArray func(
		slicePtr *[]Item,
	) array.OrderedDynamicArray[Item],
	lessOrCompareMethod func(
		oda array.OrderedDynamicArray[Item],
		i int,
		j int,
	) Result,
	lessOrCompareWantFn func(a, b Item) Result,
	data []Item,
) {
	s := slices.Clone(data)
	oda := getOrderedDynamicArray(&s)
	for i := range data {
		for j := range data {
			t.Run(fmt.Sprintf("i=%d&j=%d", i, j), func(t *testing.T) {
				want := lessOrCompareWantFn(data[i], data[j])
				got := lessOrCompareMethod(oda, i, j)
				if got != want {
					t.Errorf("got %v; want %v", got, want)
				}
			})
		}
	}
}
