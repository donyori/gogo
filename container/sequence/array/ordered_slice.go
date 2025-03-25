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

package array

import (
	"slices"

	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/function/compare"
)

// sliceOrderedDynamicArray combines SliceDynamicArray,
// github.com/donyori/gogo/function/compare.LessFunc,
// and github.com/donyori/gogo/function/compare.CompareFunc.
//
// It implements the interface OrderedDynamicArray.
type sliceOrderedDynamicArray[Item any] struct {
	*SliceDynamicArray[Item]
	lessFn compare.LessFunc[Item]
	cmpFn  compare.CompareFunc[Item]
}

var _ OrderedDynamicArray[any] = (*sliceOrderedDynamicArray[any])(nil)

// WrapSlice wraps a pointer to a Go slice with
// github.com/donyori/gogo/function/compare.LessFunc and
// github.com/donyori/gogo/function/compare.CompareFunc
// to an OrderedDynamicArray.
//
// The specified LessFunc and CompareFunc must be consistent,
// and describe a strict weak ordering.
// See <https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings>
// for details.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a strict weak ordering when not-a-number (NaN) values are involved.
//
// Operations on the returned OrderedDynamicArray will
// affect the provided Go slice.
// Operations on the Go slice will also affect
// the returned OrderedDynamicArray.
//
// If the slice pointer is nil,
// WrapSlice creates an empty slice and uses the pointer to it.
//
// If one of the LessFunc and CompareFunc is nil,
// WrapSlice generates it from the other
// (via LessFunc.ToCompare and CompareFunc.ToLess).
// If both of them are nil, WrapSlice panics.
func WrapSlice[Item any](
	slicePtr *[]Item,
	lessFn compare.LessFunc[Item],
	cmpFn compare.CompareFunc[Item],
) OrderedDynamicArray[Item] {
	if lessFn == nil {
		if cmpFn == nil {
			panic(errors.AutoMsg("both lessFn and cmpFn are nil"))
		}
		lessFn = cmpFn.ToLess()
	} else if cmpFn == nil {
		cmpFn = lessFn.ToCompare()
	}
	if slicePtr == nil {
		slicePtr = new([]Item)
	}
	return &sliceOrderedDynamicArray[Item]{
		SliceDynamicArray: (*SliceDynamicArray[Item])(slicePtr),
		lessFn:            lessFn,
		cmpFn:             cmpFn,
	}
}

// WrapStrictWeakOrderedSlice wraps a pointer to Go slice
// to an OrderedDynamicArray.
//
// It requires that the items of the slice must be strict weak ordered.
// See github.com/donyori/gogo/constraints.StrictWeakOrdered for details.
//
// Operations on the returned OrderedDynamicArray will
// affect the provided Go slice.
// Operations on the Go slice will also affect
// the returned OrderedDynamicArray.
//
// If the slice pointer is nil,
// WrapStrictWeakOrderedSlice creates an empty slice and uses the pointer to it.
func WrapStrictWeakOrderedSlice[Item constraints.StrictWeakOrdered](
	slicePtr *[]Item,
) OrderedDynamicArray[Item] {
	return WrapSlice(slicePtr, compare.OrderedLess, compare.OrderedCompare)
}

// WrapFloatSlice wraps a pointer to Go slice to an OrderedDynamicArray.
//
// It requires that the items of the slice must be floating-point numbers.
// See github.com/donyori/gogo/constraints.Float for details.
//
// The returned OrderedDynamicArray treats NaN values as less than any others.
//
// Operations on the returned OrderedDynamicArray will
// affect the provided Go slice.
// Operations on the Go slice will also affect
// the returned OrderedDynamicArray.
//
// If the slice pointer is nil,
// WrapFloatSlice creates an empty slice and uses the pointer to it.
func WrapFloatSlice[Item constraints.Float](
	slicePtr *[]Item,
) OrderedDynamicArray[Item] {
	return WrapSlice(slicePtr, compare.FloatLess, compare.FloatCompare)
}

func (soda *sliceOrderedDynamicArray[Item]) Min() Item {
	soda.checkNonempty()
	m := (*soda.SliceDynamicArray)[0]
	for i := 1; i < len(*soda.SliceDynamicArray); i++ {
		if soda.lessFn((*soda.SliceDynamicArray)[i], m) {
			m = (*soda.SliceDynamicArray)[i]
		}
	}
	return m
}

func (soda *sliceOrderedDynamicArray[Item]) Max() Item {
	soda.checkNonempty()
	m := (*soda.SliceDynamicArray)[0]
	for i := 1; i < len(*soda.SliceDynamicArray); i++ {
		if soda.lessFn(m, (*soda.SliceDynamicArray)[i]) {
			m = (*soda.SliceDynamicArray)[i]
		}
	}
	return m
}

func (soda *sliceOrderedDynamicArray[Item]) IsSorted() bool {
	for i := len(*soda.SliceDynamicArray) - 1; i > 0; i-- {
		if soda.lessFn(
			(*soda.SliceDynamicArray)[i],
			(*soda.SliceDynamicArray)[i-1],
		) {
			return false
		}
	}
	return true
}

func (soda *sliceOrderedDynamicArray[Item]) Sort() {
	slices.SortFunc(*soda.SliceDynamicArray, soda.cmpFn)
}

func (soda *sliceOrderedDynamicArray[Item]) SortStable() {
	slices.SortStableFunc(*soda.SliceDynamicArray, soda.cmpFn)
}

func (soda *sliceOrderedDynamicArray[Item]) Less(i, j int) bool {
	soda.checkNonempty()
	return soda.lessFn(
		(*soda.SliceDynamicArray)[i],
		(*soda.SliceDynamicArray)[j],
	)
}

func (soda *sliceOrderedDynamicArray[Item]) Compare(i, j int) int {
	soda.checkNonempty()
	return soda.cmpFn(
		(*soda.SliceDynamicArray)[i],
		(*soda.SliceDynamicArray)[j],
	)
}

// checkNonempty panics if the wrapped slice is nil or empty.
func (soda *sliceOrderedDynamicArray[Item]) checkNonempty() {
	if len(*soda.SliceDynamicArray) == 0 {
		panic(errors.AutoMsgCustom("OrderedDynamicArray[...] is empty", -1, 1))
	}
}
