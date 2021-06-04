// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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
	"github.com/donyori/gogo/function/compare"
)

// EqualButNotTarget is an integer used as a return value of
// the method Cmp of interface BinarySearchInterface.
//
// It stands for that the item is equal to the search target but not the target.
const EqualButNotTarget int = 7

// BinarySearchInterface represents an integer-indexed sequence
// used in the binary search algorithm.
type BinarySearchInterface interface {
	// Len returns the number of items in the sequence.
	Len() int

	// SetTarget sets the search target.
	//
	// It will be called once at the beginning of the search function.
	SetTarget(target interface{})

	// Cmp compares the item with index i and the search target.
	//
	// It returns 0 if the item with index i is the search target.
	//
	// It returns a positive integer except EqualButNotTarget (value: 7)
	// if the item with index i is greater than the search target.
	//
	// It returns a negative integer
	// if the item with index i is less than the search target.
	//
	// It returns EqualButNotTarget (value: 7) if the item with index i
	// is equal to the search target but is not the target.
	//
	// It panics if i is out of range.
	Cmp(i int) int
}

// BinarySearch finds target in itf using binary search algorithm,
// and returns its index.
//
// itf must be sorted in ascending order!
// (If itf is in descending order,
// you can change the behavior of method Cmp of BinarySearchInterface such that
// it returns a positive integer except EqualButNotTarget
// if the item is less than the search target, and returns a negative integer
// if the item is greater than the search target.)
// This function won't check whether itf is sorted.
// You must sort itf before calling this function,
// otherwise, target may not be found as expected.
//
// If Cmp returns 0 for multiple items, it returns the index of one of them.
//
// It returns -1 if target is not found.
//
// target is only used to call the method SetTarget of itf.
// It's OK to handle target in your implementation of BinarySearchInterface,
// and set target to an arbitrary value, such as nil.
//
// Time complexity: O(log n + m), where n = itf.Len(),
// m is the number of items that let the method Cmp return EqualButNotTarget.
func BinarySearch(itf BinarySearchInterface, target interface{}) int {
	itf.SetTarget(target)
	// Define: itf.Cmp(-1) < 0,
	//         itf.Cmp(itf.Len()) > 0 && itf.Cmp(itf.Len()) != EqualButNotTarget
	// Invariant: itf.Cmp(low-1) < 0,
	//            itf.Cmp(high) > 0 && itf.Cmp(high) != EqualButNotTarget
	low, high := 0, itf.Len()
	for low < high {
		mid := avg(low, high)
		cmp := itf.Cmp(mid)
		if cmp < 0 {
			low = mid + 1 // Preserve: itf.Cmp(low-1) < 0
		} else if cmp > 0 {
			if cmp == EqualButNotTarget {
				for i := mid - 1; i >= low && cmp == EqualButNotTarget; i-- {
					cmp = itf.Cmp(i)
					if cmp == 0 {
						return i
					}
				}
				cmp = EqualButNotTarget // Restore cmp to itf.Cmp(mid).
				for i := mid + 1; i < high && cmp == EqualButNotTarget; i++ {
					cmp = itf.Cmp(i)
					if cmp == 0 {
						return i
					}
				}
				return -1
			}
			high = mid // Preserve: itf.Cmp(high) > 0 && itf.Cmp(high) != EqualButNotTarget
		} else {
			return mid
		}
	}
	return -1
}

// BinarySearchMaxLess finds the maximum item less than target in itf
// using binary search algorithm, and returns its index.
//
// itf must be sorted in ascending order!
// (If itf is in descending order,
// you can change the behavior of method Cmp of BinarySearchInterface such that
// it returns a positive integer except EqualButNotTarget
// if the item is less than the search target, and returns a negative integer
// if the item is greater than the search target,
// and then use function BinarySearchMinGreater instead of this function.)
// This function won't check whether itf is sorted.
// You must sort itf before calling this function,
// otherwise, the item may not be found as expected.
//
// If multiple items satisfy the condition,
// it returns the index of the last one of them.
//
// It returns -1 if no such item in itf.
//
// target is only used to call the method SetTarget of itf.
// It's OK to handle target in your implementation of BinarySearchInterface,
// and set target to an arbitrary value, such as nil.
//
// Time complexity: O(log n), where n = itf.Len().
func BinarySearchMaxLess(itf BinarySearchInterface, target interface{}) int {
	itf.SetTarget(target)
	// Define: itf.Cmp(-1) < 0,
	//         itf.Cmp(itf.Len()) >= 0
	// Invariant: itf.Cmp(low-1) < 0,
	//            itf.Cmp(high) >= 0
	low, high := 0, itf.Len()
	for low < high {
		mid := avg(low, high)
		if itf.Cmp(mid) < 0 {
			low = mid + 1 // Preserve: itf.Cmp(low-1) < 0
		} else {
			high = mid // Preserve: itf.Cmp(high) >= 0
		}
	}
	return low - 1
}

// BinarySearchMinGreater finds the minimum item greater than target in itf
// using binary search algorithm, and returns its index.
//
// itf must be sorted in ascending order!
// (If itf is in descending order,
// you can change the behavior of method Cmp of BinarySearchInterface such that
// it returns a positive integer except EqualButNotTarget
// if the item is less than the search target, and returns a negative integer
// if the item is greater than the search target,
// and then use function BinarySearchMaxLess instead of this function.)
// This function won't check whether itf is sorted.
// You must sort itf before calling this function,
// otherwise, the item may not be found as expected.
//
// If multiple items satisfy the condition,
// it returns the index of the first one of them.
//
// It returns -1 if no such item in itf.
//
// target is only used to call the method SetTarget of itf.
// It's OK to handle target in your implementation of BinarySearchInterface,
// and set target to an arbitrary value, such as nil.
//
// Time complexity: O(log n), where n = itf.Len().
func BinarySearchMinGreater(itf BinarySearchInterface, target interface{}) int {
	itf.SetTarget(target)
	// Define: itf.Cmp(-1) <= 0 || itf.Cmp(-1) == EqualButNotTarget,
	//         itf.Cmp(itf.Len()) > 0 && itf.Cmp(itf.Len()) != EqualButNotTarget
	// Invariant: itf.Cmp(low-1) <= 0 || itf.Cmp(low-1) == EqualButNotTarget,
	//            itf.Cmp(high) > 0 && itf.Cmp(high) != EqualButNotTarget
	low, high := 0, itf.Len()
	for low < high {
		mid := avg(low, high)
		cmp := itf.Cmp(mid)
		if cmp > 0 && cmp != EqualButNotTarget {
			high = mid // Preserve: itf.Cmp(high) > 0 && itf.Cmp(high) != EqualButNotTarget
		} else {
			low = mid + 1 // Preserve: itf.Cmp(low-1) <= 0 || itf.Cmp(low-1) == EqualButNotTarget
		}
	}
	if high < itf.Len() {
		return high
	}
	return -1
}

// BinarySearchArrayAdapter is an adapter for:
// sequence.Array + compare.EqualFunc + compare.LessFunc -> BinarySearchInterface.
//
// Note that EqualFn should return true if and only if
// the item is the search target.
// If the item is equal to the target but is not the target,
// EqualFn should return false.
type BinarySearchArrayAdapter struct {
	Data    sequence.Array
	EqualFn compare.EqualFunc
	LessFn  compare.LessFunc

	target interface{}
}

// Len returns the number of items in the sequence.
func (bsad *BinarySearchArrayAdapter) Len() int {
	if bsad == nil || bsad.Data == nil {
		return 0
	}
	return bsad.Data.Len()
}

// SetTarget sets the search target.
//
// It will be called once at the beginning of the search function.
func (bsad *BinarySearchArrayAdapter) SetTarget(target interface{}) {
	bsad.target = target
}

// Cmp compares the item with index i and the search target.
//
// It returns 0 if the item with index i is the search target.
//
// It returns a positive integer except EqualButNotTarget (value: 7)
// if the item with index i is greater than the search target.
//
// It returns a negative integer
// if the item with index i is less than the search target.
//
// It returns EqualButNotTarget (value: 7) if the item with index i
// is equal to the search target but is not the target.
//
// It panics if i is out of range.
func (bsad *BinarySearchArrayAdapter) Cmp(i int) int {
	item := bsad.Data.Get(i)
	if bsad.LessFn(item, bsad.target) {
		return -1
	}
	if bsad.LessFn(bsad.target, item) {
		return 1
	}
	if bsad.EqualFn(item, bsad.target) {
		return 0
	}
	return EqualButNotTarget
}

// avg returns the average of two non-negative integers a and b.
//
// It avoids overflow when computing the average.
//
// The return value (denoted by c) satisfies a <= c < b.
//
// Caller should guarantee that a and b are non-negative.
func avg(a, b int) int {
	return int(uint(a+b) >> 1)
}
