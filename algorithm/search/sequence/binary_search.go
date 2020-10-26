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

// BinarySearchInterface represents an integer-indexed sequence
// used in the binary search algorithm.
type BinarySearchInterface interface {
	// Len returns the number of items in the sequence.
	Len() int

	// Equal reports whether i-th item of the sequence equals to x.
	// It panics if i is out of range.
	Equal(i int, x interface{}) bool

	// Less reports whether i-th item is less than x.
	// It panics if i is out of range.
	Less(i int, x interface{}) bool

	// Greater reports whether i-th item is greater than x.
	// It panics if i is out of range.
	Greater(i int, x interface{}) bool
}

// BinarySearch finds target in data using binary search algorithm,
// and returns its index.
//
// data must be sorted in ascending order!
// (If data is in descending order, you can exchange the behavior
// of Less and Greater methods of BinarySearchInterface.)
// This function won't check whether data is sorted.
// You must sort data before calling this function,
// otherwise, target may not be found as expected.
//
// If multiple items equal to target, it returns the index of one of them.
// It returns -1 if target is not found.
//
// target is only used to call the methods of data (BinarySearchInterface).
// It's OK to handle target in your implementation of BinarySearchInterface,
// and set target to an arbitrary value, such as nil.
//
// Time complexity: O(log n + m), where n = data.Len(),
// m = the number of items that satisfy: !Equal() && !Less() && !Greater().
func BinarySearch(data BinarySearchInterface, target interface{}) int {
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

// BinarySearchMaxLess finds the maximum item less than target in data
// using binary search algorithm, and returns its index.
//
// data must be sorted in ascending order!
// (If data is in descending order, you can exchange the behavior
// of Less and Greater methods of BinarySearchInterface, and then
// use BinarySearchMinGreater instead of this function.)
// This function won't check whether data is sorted.
// You must sort data before calling this function,
// otherwise, the item may not be found as expected.
//
// If multiple items satisfy the condition,
// it returns the index of the last one of them.
// It returns -1 if no such item in data.
//
// Only Len() and Less() are used in this function,
// and it's OK to leave Equal() and Greater() as empty methods.
//
// target is only used to call the method Less() of data (BinarySearchInterface).
// It's OK to handle target in your implementation of BinarySearchInterface,
// and set target to an arbitrary value, such as nil.
//
// Time complexity: O(log n), where n = data.Len().
func BinarySearchMaxLess(data BinarySearchInterface, target interface{}) int {
	n := data.Len()
	if n == 0 {
		return -1
	}
	if !data.Less(0, target) {
		return -1
	}
	if data.Less(n-1, target) {
		return n - 1
	}
	low, high := 0, n
	mid := (low + high) / 2
	for low != mid {
		if data.Less(mid, target) {
			low = mid
		} else {
			high = mid
		}
		mid = (low + high) / 2
	}
	return low
}

// BinarySearchMinGreater finds the minimum item greater than target in data
// using binary search algorithm, and returns its index.
//
// data must be sorted in ascending order!
// (If data is in descending order, you can exchange the behavior
// of Less and Greater methods of BinarySearchInterface, and then
// use BinarySearchMaxLess instead of this function.)
// This function won't check whether data is sorted.
// You must sort data before calling this function,
// otherwise, the item may not be found as expected.
//
// If multiple items satisfy the condition,
// it returns the index of the first one of them.
// It returns -1 if no such item in data.
//
// Only Len() and Greater() are used in this function,
// and it's OK to leave Equal() and Less() as empty methods.
//
// target is only used to call the method Greater() of data (BinarySearchInterface).
// It's OK to handle target in your implementation of BinarySearchInterface,
// and set target to an arbitrary value, such as nil.
//
// Time complexity: O(log n), where n = data.Len().
func BinarySearchMinGreater(data BinarySearchInterface, target interface{}) int {
	n := data.Len()
	if n == 0 {
		return -1
	}
	if data.Greater(0, target) {
		return 0
	}
	if !data.Greater(n-1, target) {
		return -1
	}
	low, high := 0, n
	mid := (low + high) / 2
	for low != mid {
		if data.Greater(mid, target) {
			high = mid
		} else {
			low = mid
		}
		mid = (low + high) / 2
	}
	return high
}

// BinarySearchArrayAdapter is an adapter for:
// sequence.Array + function.EqualFunc + function.LessFunc -> BinarySearchInterface.
type BinarySearchArrayAdapter struct {
	Data    sequence.Array
	EqualFn function.EqualFunc
	LessFn  function.LessFunc
}

// Len returns the number of items in the sequence.
func (bsad *BinarySearchArrayAdapter) Len() int {
	if bsad == nil || bsad.Data == nil {
		return 0
	}
	return bsad.Data.Len()
}

// Equal reports whether i-th item of the sequence equals to x.
// It panics if i is out of range.
func (bsad *BinarySearchArrayAdapter) Equal(i int, x interface{}) bool {
	if bsad.EqualFn == nil && bsad.LessFn != nil {
		bsad.EqualFn = function.GenerateEqualViaLess(bsad.LessFn)
	}
	return bsad.EqualFn(bsad.Data.Get(i), x)
}

// Less reports whether i-th item is less than x.
// It panics if i is out of range.
func (bsad *BinarySearchArrayAdapter) Less(i int, x interface{}) bool {
	return bsad.LessFn(bsad.Data.Get(i), x)
}

// Greater reports whether i-th item is greater than x.
// It panics if i is out of range.
func (bsad *BinarySearchArrayAdapter) Greater(i int, x interface{}) bool {
	return bsad.LessFn(x, bsad.Data.Get(i))
}
