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

package sequence

import (
	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/function/compare"
)

// BinarySearchInterface represents an integer-indexed sequence
// used in the binary search algorithm.
//
// Its type parameter Item represents the type of item in the sequence.
type BinarySearchInterface[Item any] interface {
	// SetGoal sets the search goal.
	//
	// It will be called once at the beginning of the search functions.
	//
	// Its implementation can do initialization for each search in this method.
	SetGoal(goal Item)

	// Len returns the number of items in the sequence.
	Len() int

	// Compare compares the item at index i and the search goal.
	//
	// It returns an integer and a boolean indicator,
	// where the integer represents the item at index i is less than,
	// equal to, or greater than the search goal,
	// and the indicator reports whether the item at index i
	// is the search goal (only valid when the returned integer is 0).
	//
	// The returned integer is
	//
	//	-1 if the item at index i is less than the search goal,
	//	 0 if the item at index i is equal to the search goal,
	//	+1 if the item at index i is greater than the search goal.
	//
	// It panics if i is out of range.
	Compare(i int) (lessEqualOrGreater int, isGoal bool)
}

// BinarySearch finds goal in itf using binary search algorithm,
// and returns its index.
//
// itf must be sorted in ascending order!
// (If itf is in descending order,
// you can change the behavior of your itf.Compare such that
// it returns +1 if the item is less than the search goal,
// and returns -1 if the item is greater than the search goal.)
// This function won't check whether itf is sorted.
// You must sort itf before calling this function,
// otherwise, goal may not be found as expected.
//
// If itf.Compare returns (0, true) for multiple items,
// it returns the index of one of them.
//
// It returns -1 if goal is not found.
//
// goal is only used to call the method SetGoal of itf.
// It's OK to handle goal in your implementation of BinarySearchInterface,
// and set goal to an arbitrary value, such as a zero value.
//
// Time complexity: O(log n + m), where n = itf.Len(),
// m is the number of items that let itf.Compare return (0, false).
func BinarySearch[Item any](itf BinarySearchInterface[Item], goal Item) int {
	itf.SetGoal(goal)
	// Denote the first return value of itf.Compare(i) as f(i).
	// Define f(-1) < 0 and f(itf.Len()) > 0.
	// Invariant: f(low-1) < 0, f(high) > 0.
	low, high := 0, itf.Len()
	for low < high {
		mid := avg(low, high)
		cmp, isGoal := itf.Compare(mid)
		if cmp < 0 {
			low = mid + 1 // preserve f(low-1) < 0
		} else if cmp > 0 {
			high = mid // preserve f(high) > 0
		} else {
			if isGoal {
				return mid
			}
			for i := mid - 1; i >= low; i-- {
				cmp, isGoal = itf.Compare(i)
				// check cmp before isGoal as isGoal is only valid when cmp == 0
				if cmp != 0 {
					break
				}
				if isGoal {
					return i
				}
			}
			for i := mid + 1; i < high; i++ {
				cmp, isGoal = itf.Compare(i)
				// check cmp before isGoal as isGoal is only valid when cmp == 0
				if cmp != 0 {
					break
				}
				if isGoal {
					return i
				}
			}
			return -1
		}
	}
	return -1
}

// BinarySearchMaxLess finds the maximum item less than goal in itf
// using binary search algorithm, and returns its index.
//
// itf must be sorted in ascending order!
// (If itf is in descending order,
// you can change the behavior of your itf.Compare such that
// it returns +1 if the item is less than the search goal,
// and returns -1 if the item is greater than the search goal,
// and then use function BinarySearchMinGreater instead of this function.)
// This function won't check whether itf is sorted.
// You must sort itf before calling this function,
// otherwise, the item may not be found as expected.
//
// If multiple items satisfy the condition,
// it returns the index of the last one of them.
//
// It returns -1 if there is no such item in itf.
//
// goal is only used to call the method SetGoal of itf.
// It's OK to handle goal in your implementation of BinarySearchInterface,
// and set goal to an arbitrary value, such as a zero value.
//
// Time complexity: O(log n), where n = itf.Len().
func BinarySearchMaxLess[Item any](
	itf BinarySearchInterface[Item],
	goal Item,
) int {
	itf.SetGoal(goal)
	// Denote the first return value of itf.Compare(i) as f(i).
	// Define f(-1) < 0 and f(itf.Len()) >= 0.
	// Invariant: f(low-1) < 0, f(high) >= 0.
	low, high := 0, itf.Len()
	for low < high {
		mid := avg(low, high)
		if cmp, _ := itf.Compare(mid); cmp < 0 {
			low = mid + 1 // preserve f(low-1) < 0
		} else {
			high = mid // preserve f(high) >= 0
		}
	}
	return low - 1
}

// BinarySearchMinGreater finds the minimum item greater than goal in itf
// using binary search algorithm, and returns its index.
//
// itf must be sorted in ascending order!
// (If itf is in descending order,
// you can change the behavior of your itf.Compare such that
// it returns +1 if the item is less than the search goal,
// and returns -1 if the item is greater than the search goal,
// and then use function BinarySearchMaxLess instead of this function.)
// This function won't check whether itf is sorted.
// You must sort itf before calling this function,
// otherwise, the item may not be found as expected.
//
// If multiple items satisfy the condition,
// it returns the index of the first one of them.
//
// It returns -1 if there is no such item in itf.
//
// goal is only used to call the method SetGoal of itf.
// It's OK to handle goal in your implementation of BinarySearchInterface,
// and set goal to an arbitrary value, such as a zero value.
//
// Time complexity: O(log n), where n = itf.Len().
func BinarySearchMinGreater[Item any](
	itf BinarySearchInterface[Item],
	goal Item,
) int {
	itf.SetGoal(goal)
	// Denote the first return value of itf.Compare(i) as f(i).
	// Define f(-1) <= 0 and f(itf.Len()) > 0.
	// Invariant: f(low-1) <= 0, f(high) > 0.
	low, high := 0, itf.Len()
	for low < high {
		mid := avg(low, high)
		if cmp, _ := itf.Compare(mid); cmp > 0 {
			high = mid // preserve f(high) > 0
		} else {
			low = mid + 1 // preserve f(low-1) <= 0
		}
	}
	if high < itf.Len() {
		return high
	}
	return -1
}

// arrayBinarySearchAdapter combines
// github.com/donyori/gogo/container/sequence/array.Array,
// github.com/donyori/gogo/function/compare.LessFunc,
// and github.com/donyori/gogo/function/compare.EqualFunc.
//
// It implements the interface BinarySearchInterface.
//
// The field data is the array in which to search.
//
// The field lessFn is used to compare items in data with the search goal.
// It cannot be nil.
// If both lessFn(item, goal) and lessFn(goal, item) return false,
// item and goal are considered equal,
// and the first return value of method Compare is 0.
//
// lessFn must describe a strict weak ordering.
// See <https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings>
// for details.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a strict weak ordering when not-a-number (NaN) values are involved.
//
// The field equalFn is an additional function to test
// whether an item is the search goal.
// If equalFn is nil, the item is considered the search goal
// when they are equal.
// equalFn will be called only when both lessFn(item, goal) and
// lessFn(goal, item) return false.
// When equalFn is non-nil, the second return value of method Compare is true
// if and only if lessFn(item, goal) returns false,
// lessFn(goal, item) returns false,
// and equalFn(item, goal) returns true.
type arrayBinarySearchAdapter[Item any] struct {
	data    array.Array[Item]
	lessFn  compare.LessFunc[Item]
	equalFn compare.EqualFunc[Item]
	goal    Item
}

// WrapArrayLessEqual wraps
// github.com/donyori/gogo/container/sequence/array.Array
// with github.com/donyori/gogo/function/compare.LessFunc
// and github.com/donyori/gogo/function/compare.EqualFunc
// to a BinarySearchInterface.
//
// data is the array in which to search.
//
// lessFn is used to compare items in data with the search goal.
// It cannot be nil.
// If both lessFn(item, goal) and lessFn(goal, item) return false,
// item and goal are considered equal,
// and the first return value of method Compare is 0.
//
// lessFn must describe a strict weak ordering.
// See <https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings>
// for details.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a strict weak ordering when not-a-number (NaN) values are involved.
//
// WrapArrayLessEqual panics if lessFn is nil.
//
// equalFn is an additional function to test
// whether an item is the search goal.
// If equalFn is nil, the item is considered the search goal
// when they are equal.
// equalFn will be called only when both lessFn(item, goal) and
// lessFn(goal, item) return false.
// When equalFn is non-nil, the second return value of method Compare is true
// if and only if lessFn(item, goal) returns false,
// lessFn(goal, item) returns false,
// and equalFn(item, goal) returns true.
func WrapArrayLessEqual[Item any](
	data array.Array[Item],
	lessFn compare.LessFunc[Item],
	equalFn compare.EqualFunc[Item],
) BinarySearchInterface[Item] {
	if lessFn == nil {
		panic(errors.AutoMsg("lessFn is nil"))
	}
	return &arrayBinarySearchAdapter[Item]{
		data:    data,
		lessFn:  lessFn,
		equalFn: equalFn,
	}
}

func (absa *arrayBinarySearchAdapter[Item]) SetGoal(goal Item) {
	absa.goal = goal
}

func (absa *arrayBinarySearchAdapter[Item]) Len() int {
	if absa == nil || absa.data == nil {
		return 0
	}
	return absa.data.Len()
}

func (absa *arrayBinarySearchAdapter[Item]) Compare(i int) (
	lessEqualOrGreater int, isGoal bool) {
	item := absa.data.Get(i)
	if absa.lessFn(item, absa.goal) {
		return -1, false
	} else if absa.lessFn(absa.goal, item) {
		return 1, false
	}
	return 0, absa.equalFn == nil || absa.equalFn(item, absa.goal)
}

// avg returns the average of two nonnegative integers a and b.
//
// It avoids overflow when computing the average.
//
// The return value (denoted by c) satisfies a <= c < b.
//
// Caller should guarantee that a and b are nonnegative.
func avg(a, b int) int {
	return int(uint(a+b) >> 1)
}
