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

package array_test

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/function/compare"
)

func TestAffectProvidedSlice(t *testing.T) {
	t.Run("sliceLess", func(t *testing.T) {
		testAffectProvidedSlice(
			t,
			func(slicePtr *[]int) array.OrderedDynamicArray[int] {
				return array.WrapSliceLess(slicePtr, compare.OrderedLess[int])
			},
		)
	})

	t.Run("transitiveOrderedSlice", func(t *testing.T) {
		testAffectProvidedSlice(
			t,
			func(slicePtr *[]int) array.OrderedDynamicArray[int] {
				return array.WrapTransitiveOrderedSlice(slicePtr)
			},
		)
	})

	t.Run("floatSlice", func(t *testing.T) {
		testAffectProvidedSlice(
			t,
			func(slicePtr *[]float64) array.OrderedDynamicArray[float64] {
				return array.WrapFloatSlice(slicePtr)
			},
		)
	})
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

func TestSliceLess_Less(t *testing.T) {
	testFloat64SliceLess(
		t,
		func(slicePtr *[]float64) array.OrderedDynamicArray[float64] {
			return array.WrapSliceLess(slicePtr, compare.FloatLess[float64])
		},
	)
}

func TestTransitiveOrderedSlice_Less(t *testing.T) {
	data := []string{"Alice", "Hello world!", "ball", "中文"} // data is sorted
	if !slices.IsSorted(data) {
		t.Fatal("unsorted data, please update")
	}
	oda := array.WrapTransitiveOrderedSlice(&data)
	for i := range data {
		for j := range data {
			t.Run(
				fmt.Sprintf("i=%d(s[i]=%s)&j=%d(s[j]=%s)",
					i, data[i], j, data[j]),
				func(t *testing.T) {
					got := oda.Less(i, j)
					if got != (i < j) {
						t.Errorf("got %t; want %t", got, i < j)
					}
				},
			)
		}
	}
}

func TestFloatSlice_Less(t *testing.T) {
	testFloat64SliceLess(
		t,
		func(slicePtr *[]float64) array.OrderedDynamicArray[float64] {
			return array.WrapFloatSlice(slicePtr)
		},
	)
}

func testFloat64SliceLess(
	t *testing.T,
	getOrderedDynamicArray func(
		slicePtr *[]float64,
	) array.OrderedDynamicArray[float64],
) {
	data := []float64{
		math.NaN(), math.Inf(-1), -1.1, -1., -.1, 0.,
		math.SmallestNonzeroFloat64, .1, 1., 1.1, math.MaxFloat64,
		math.Inf(1),
	} // data is sorted
	if !slices.IsSorted(data) {
		t.Fatal("unsorted data, please update")
	}
	dataValueStr := []string{
		"NaN", "-Inf", "-1.1", "-1.0", "-0.1", "0.0",
		"SmallestNonzero", "0.1", "1.0", "1.1", "Max",
		"Inf",
	}
	n := len(data)
	if n != len(dataValueStr) {
		t.Fatalf("len(data): %d, len(dataValueStr): %d, not equal, please check the test code",
			n, len(dataValueStr))
	}
	oda := getOrderedDynamicArray(&data)
	for i := range n {
		for j := range n {
			t.Run(
				fmt.Sprintf("i=%d(s[i]=%s)&j=%d(s[j]=%s)",
					i, dataValueStr[i], j, dataValueStr[j]),
				func(t *testing.T) {
					got := oda.Less(i, j)
					if got != (i < j) {
						t.Errorf("got %t; want %t", got, i < j)
					}
				},
			)
		}
	}
}
