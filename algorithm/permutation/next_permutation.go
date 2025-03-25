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

package permutation

import "github.com/donyori/gogo/algorithm/search/sequence"

// Interface represents an integer-indexed permutation with method Less.
//
// It is consistent with the interface sort.Interface.
type Interface interface {
	// Len returns the number of items in the permutation.
	Len() int

	// Less reports whether the item at index i is less than that at index j.
	//
	// Less must describe a strict weak ordering.
	// See <https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings>
	// for details.
	//
	// Note that floating-point comparison
	// (the < operator on float32 or float64 values)
	// is not a strict weak ordering
	// when not-a-number (NaN) values are involved.
	//
	// It panics if i or j is out of range.
	Less(i, j int) bool

	// Swap exchanges the items at index i and index j.
	//
	// It panics if i or j is out of range.
	Swap(i, j int)
}

// NextPermutation transforms itf to its next permutation in lexical order.
// It returns false if itf.Len() == 0 or the permutations are exhausted,
// and true otherwise.
//
// Time complexity: O(n), where n = itf.Len().
func NextPermutation(itf Interface) bool {
	if itf == nil {
		return false
	}
	i := itf.Len() - 2
	for i >= 0 && !itf.Less(i, i+1) {
		i--
	}
	if i < 0 {
		return false
	}
	npbsi := &nextPermutationBinarySearchInterface{
		data:  itf,
		begin: i + 1,
	}
	j := npbsi.begin + sequence.BinarySearchMaxLess(npbsi, i) // find the last item greater than data[i]
	itf.Swap(i, j)
	for i, j = i+1, itf.Len()-1; i < j; i, j = i+1, j-1 {
		itf.Swap(i, j)
	}
	return true
}

// nextPermutationBinarySearchInterface is an implementation of interface
// github.com/donyori/gogo/algorithm/search/sequence.BinarySearchInterface.
//
// As this implementation is designed to find the last item
// greater than the search goal in a DESCENDING sequence,
// if the item is greater than the search goal,
// it is treated as "less" than the goal in the binary search algorithm.
// Otherwise, it is treated as "greater" than the goal.
type nextPermutationBinarySearchInterface struct {
	data  Interface
	begin int
	goal  int
}

// SetGoal sets the search goal.
//
// It will be called once at the beginning of the search functions.
//
// In this implementation, goal is the index of the search goal
// in the permutation.
func (npbsi *nextPermutationBinarySearchInterface) SetGoal(goal int) {
	npbsi.goal = goal
}

// Len returns the number of items in the sequence.
//
// The sequence is a slice of the permutation.
func (npbsi *nextPermutationBinarySearchInterface) Len() int {
	if npbsi.data == nil {
		return 0
	}
	return npbsi.data.Len() - npbsi.begin
}

// Compare compares the item at index i in the sequence and the search goal.
//
// It returns (-1, false) if the item is greater than the goal
// (in this case, the item is treated as "less" than
// the goal in the binary search algorithm).
//
// Otherwise, it returns (1, false) (corresponding to the case where
// the item is "greater" than the goal in the binary search algorithm).
//
// It panics if i is out of range.
func (npbsi *nextPermutationBinarySearchInterface) Compare(i int) (
	lessEqualOrGreater int, isGoal bool) {
	if npbsi.data.Less(npbsi.goal, i+npbsi.begin) {
		return -1, false
	}
	return 1, false
}
