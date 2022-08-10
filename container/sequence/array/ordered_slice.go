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

package array

import (
	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/function/compare"
)

// sliceLess combines SliceDynamicArray and
// github.com/donyori/gogo/function/compare.LessFunc.
//
// It implements the interface OrderedDynamicArray.
type sliceLess[Item any] struct {
	*SliceDynamicArray[Item]
	lessFn compare.LessFunc[Item]
}

// WrapSliceLess wraps a pointer to a Go slice with
// github.com/donyori/gogo/function/compare.LessFunc to an OrderedDynamicArray.
//
// The specified LessFunc must describe a transitive ordering:
//   - if both lessFn(i, j) and lessFn(j, k) are true, then lessFn(i, k) must be true as well.
//   - if both lessFn(i, j) and lessFn(j, k) are false, then lessFn(i, k) must be false as well.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
//
// Operations on the returned OrderedDynamicArray will
// affect the provided Go slice.
// Operations on the Go slice will also affect
// the returned OrderedDynamicArray.
//
// It panics if slicePtr or lessFn is nil.
func WrapSliceLess[Item any](
	slicePtr *[]Item, lessFn compare.LessFunc[Item]) OrderedDynamicArray[Item] {
	if slicePtr == nil {
		panic(errors.AutoMsg("slicePtr is nil"))
	}
	if lessFn == nil {
		panic(errors.AutoMsg("lessFn is nil"))
	}
	return &sliceLess[Item]{
		SliceDynamicArray: (*SliceDynamicArray[Item])(slicePtr),
		lessFn:            lessFn,
	}
}

// Less reports whether the item with index i must sort before
// the item with index j.
//
// If the LessFunc specified by the client describes a transitive ordering,
// then Less describes a transitive ordering as well:
//   - if both Less(i, j) and Less(j, k) are true, then Less(i, k) must be true as well.
//   - if both Less(i, j) and Less(j, k) are false, then Less(i, k) must be false as well.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
//
// It panics if i or j is out of range.
func (sl *sliceLess[Item]) Less(i, j int) bool {
	return sl.lessFn((*sl.SliceDynamicArray)[i], (*sl.SliceDynamicArray)[j])
}

// transitiveOrderedSlice is a SliceDynamicArray that constraints its item type
// to transitive ordered types.
//
// It implements the interface OrderedDynamicArray.
type transitiveOrderedSlice[Item constraints.TransitiveOrdered] struct {
	*SliceDynamicArray[Item]
}

// WrapTransitiveOrderedSlice wraps a pointer to Go slice
// to an OrderedDynamicArray.
//
// It requires that the items of the slice must be transitive ordered.
// See github.com/donyori/gogo/constraints.TransitiveOrdered for details.
//
// Operations on the returned OrderedDynamicArray will
// affect the provided Go slice.
// Operations on the Go slice will also affect
// the returned OrderedDynamicArray.
//
// It panics if slicePtr is nil.
func WrapTransitiveOrderedSlice[Item constraints.TransitiveOrdered](
	slicePtr *[]Item) OrderedDynamicArray[Item] {
	if slicePtr == nil {
		panic(errors.AutoMsg("slicePtr is nil"))
	}
	return &transitiveOrderedSlice[Item]{
		SliceDynamicArray: (*SliceDynamicArray[Item])(slicePtr),
	}
}

// Less reports whether the item with index i must sort before
// the item with index j.
//
// Less describes a transitive ordering:
//   - if both Less(i, j) and Less(j, k) are true, then Less(i, k) must be true as well.
//   - if both Less(i, j) and Less(j, k) are false, then Less(i, k) must be false as well.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
//
// It panics if i or j is out of range.
func (tos *transitiveOrderedSlice[Item]) Less(i, j int) bool {
	return (*tos.SliceDynamicArray)[i] < (*tos.SliceDynamicArray)[j]
}

// floatSlice is a SliceDynamicArray that constraints its item type
// to floating-point numbers.
//
// It implements the interface OrderedDynamicArray.
type floatSlice[Item constraints.Float] struct {
	*SliceDynamicArray[Item]
}

// WrapFloatSlice wraps a pointer to Go slice to an OrderedDynamicArray.
//
// It requires that the items of the slice must be floating-point numbers.
// See github.com/donyori/gogo/constraints.Float for details.
//
// Operations on the returned OrderedDynamicArray will
// affect the provided Go slice.
// Operations on the Go slice will also affect
// the returned OrderedDynamicArray.
//
// It panics if slicePtr is nil.
func WrapFloatSlice[Item constraints.Float](slicePtr *[]Item) OrderedDynamicArray[Item] {
	if slicePtr == nil {
		panic(errors.AutoMsg("slicePtr is nil"))
	}
	return &floatSlice[Item]{
		SliceDynamicArray: (*SliceDynamicArray[Item])(slicePtr),
	}
}

// Less reports whether the item with index i must sort before
// the item with index j.
//
// Less describes a transitive ordering:
//   - if both Less(i, j) and Less(j, k) are true, then Less(i, k) must be true as well.
//   - if both Less(i, j) and Less(j, k) are false, then Less(i, k) must be false as well.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
//
// It panics if i or j is out of range.
func (fs *floatSlice[Item]) Less(i, j int) bool {
	return compare.FloatLess((*fs.SliceDynamicArray)[i], (*fs.SliceDynamicArray)[j])
}
