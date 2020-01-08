// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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

package sequence

import (
	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/function"
)

// A type standing for an integer-indexed sequence used in BinarySearch.
type BinarySearchInterface interface {
	// Return the number of items in the sequence.
	// It returns 0 if the sequence is nil.
	Len() int

	// Test whether i-th item of the sequence equals to x.
	// It panics if i is out of range.
	Equal(i int, x interface{}) bool

	// Test whether i-th item is less than x.
	// It panics if i is out of range.
	Less(i int, x interface{}) bool

	// Test whether i-th item is greater than x.
	// It panics if i is out of range.
	Greater(i int, x interface{}) bool
}

// Find target in data using binary search algorithm.
// data must be sorted in ascending order!
// (If data is in descending order, you can exchange the behavior
// of Less and Greater methods of BinarySearchInterface.)
// This function won't check whether data is sorted.
// You must sort data before calling this function,
// otherwise, target may not be found as expected.
// If multiple items equal to target, it returns the index of one of them.
// It returns -1 if target is not found.
// Time complexity: O(log n + m), where n = data.Len(),
// m = the number of items satisfy: !Equal() && !Less() && !Greater().
func BinarySearch(target interface{}, data BinarySearchInterface) int {
	if data.Len() == 0 {
		return -1
	}
	low, high := 0, data.Len()
	mid := (low + high) / 2
	for low != mid {
		if data.Less(mid, target) {
			low = mid
		} else if data.Greater(mid, target) {
			high = mid
		} else {
			break
		}
		mid = (low + high) / 2
	}
	if data.Equal(mid, target) {
		return mid
	}
	for i := mid - 1; i >= low && !data.Less(i, target); i-- {
		if data.Equal(i, target) {
			return i
		}
	}
	for i := mid + 1; i < high && !data.Greater(i, target); i++ {
		if data.Equal(i, target) {
			return i
		}
	}
	return -1
}

// An adapter for: Array + EqualFunc + LessFunc -> BinarySearchInterface.
type BinarySearchArrayAdapter struct {
	Data    sequence.Array
	EqualFn function.EqualFunc
	LessFn  function.LessFunc
}

func (bsad *BinarySearchArrayAdapter) Len() int {
	if bsad == nil || bsad.Data == nil {
		return 0
	}
	return bsad.Data.Len()
}

func (bsad *BinarySearchArrayAdapter) Equal(i int, x interface{}) bool {
	return bsad.EqualFn(bsad.Data.Get(i), x)
}

func (bsad *BinarySearchArrayAdapter) Less(i int, x interface{}) bool {
	return bsad.LessFn(bsad.Data.Get(i), x)
}

func (bsad *BinarySearchArrayAdapter) Greater(i int, x interface{}) bool {
	return bsad.LessFn(x, bsad.Data.Get(i))
}
