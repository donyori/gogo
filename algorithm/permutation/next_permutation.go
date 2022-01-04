// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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
type Interface interface {
	// Len returns the number of items in the permutation.
	Len() int

	// Less reports whether the item with index i
	// is less than the item with index j.
	//
	// Less must describe a transitive ordering:
	//  - if both Less(i, j) and Less(j, k) are true, then Less(i, k) must be true as well.
	//  - if both Less(i, j) and Less(j, k) are false, then Less(i, k) must be false as well.
	//
	// Note that floating-point comparison
	// (the < operator on float32 or float64 values)
	// is not a transitive ordering when not-a-number (NaN) values are involved.
	//
	// It panics if i or j is out of range.
	Less(i, j int) bool

	// Swap exchanges the items with indexes i and j.
	//
	// It panics if i or j is out of range.
	Swap(i, j int)
}

// NextPermutation transforms itf to its next permutation in lexical order.
// It returns false if itf.Len() == 0 or the permutations are exhausted,
// and true otherwise.
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
		Data:  itf,
		Begin: i + 1,
	}
	j := npbsi.Begin + sequence.BinarySearchMaxLess(npbsi, i)
	itf.Swap(i, j)
	for i, j = i+1, itf.Len()-1; i < j; i, j = i+1, j-1 {
		itf.Swap(i, j)
	}
	return true
}

// nextPermutationBinarySearchInterface is an implementation of
// interface BinarySearchInterface.
//
// As this implementation is designed to find the last item
// that is greater than the search goal in a descending sequence,
// if the item is greater than the search goal,
// it is treated as "less" than the goal in the binary search algorithm.
// Otherwise, it is treated as "greater" than the goal.
type nextPermutationBinarySearchInterface struct {
	Data  Interface
	Begin int

	goal int
}

// Len returns the number of items in the sequence.
//
// The sequence is a slice of the permutation.
func (npbsi *nextPermutationBinarySearchInterface) Len() int {
	if npbsi == nil || npbsi.Data == nil {
		return 0
	}
	return npbsi.Data.Len() - npbsi.Begin
}

// SetGoal sets the search goal.
//
// It will be called once at the beginning of the search functions.
//
// In this implementation, goal is the index of the search goal
// in the permutation.
func (npbsi *nextPermutationBinarySearchInterface) SetGoal(goal interface{}) {
	npbsi.goal = goal.(int)
}

// Cmp compares the item with index i in the sequence and the search goal.
//
// It returns -1 if the item is greater than the goal
// (in this case, the item is treated as "less" than
// the goal in the binary search algorithm).
//
// Otherwise, it returns 1 (corresponding to the case where
// the item is "greater" than the goal in the binary search algorithm).
//
// It panics if i is out of range.
func (npbsi *nextPermutationBinarySearchInterface) Cmp(i int) int {
	if npbsi.Data.Less(npbsi.goal, i+npbsi.Begin) {
		return -1
	}
	return 1
}
