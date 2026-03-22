// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2026  Yuan Gao
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

// Interface represents a zero-based integer-indexed permutation
// with method Less.
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

	const SecondToLastOffset int = -2

	i := itf.Len() + SecondToLastOffset
	for i >= 0 && !itf.Less(i, i+1) {
		i--
	}

	if i < 0 {
		return false
	}

	bsi := &nextPermutationBinarySearchInterface{
		p:     itf,
		begin: i + 1,
	}
	j := bsi.begin + sequence.BinarySearchMaxLess(bsi, i) // find the last item greater than p[i]
	itf.Swap(i, j)

	for i, j = i+1, itf.Len()-1; i < j; i, j = i+1, j-1 {
		itf.Swap(i, j)
	}

	return true
}

// nextPermutationBinarySearchInterface is an implementation of interface
// github.com/donyori/gogo/algorithm/search/sequence.BinarySearchInterface.
//
// In this implementation, target is the index of the search target
// in the permutation.
//
// As this implementation is designed to find the last item
// greater than the search target in a DESCENDING sequence,
// if the item is greater than the search target,
// it is treated as "less" than the target in the binary search algorithm.
// Otherwise, it is treated as "greater" than the target.
type nextPermutationBinarySearchInterface struct {
	p     Interface
	begin int
}

// Len returns the number of items in the sequence,
// where the sequence is a slice of the permutation.
func (bsi *nextPermutationBinarySearchInterface) Len() int {
	if bsi.p == nil {
		return 0
	}

	return bsi.p.Len() - bsi.begin
}

// CompareWithTarget compares the item at index i in the sequence
// with the search target, where the sequence is a slice of the permutation.
//
// It returns (-1, false) if the item is greater than the target
// (in this case, the item is treated as "less" than
// the target in the binary search algorithm).
//
// Otherwise, it returns (1, false) (corresponding to the case where
// the item is "greater" than the target in the binary search algorithm).
//
// It panics if i is out of range.
func (bsi *nextPermutationBinarySearchInterface) CompareWithTarget(
	i int,
	target int,
) (cmpResult int, isTarget bool) {
	if bsi.p.Less(target, i+bsi.begin) {
		return -1, false
	}

	return 1, false
}
