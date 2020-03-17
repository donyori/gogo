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

package permutation

import (
	"github.com/donyori/gogo/algorithm/search/sequence"
	"github.com/donyori/gogo/errors"
)

// Transform data to its next permutation in lexical order.
// It returns false if data.Len() == 0 or the permutations are exhausted,
// and true otherwise.
// Time complexity: O(n), where n = data.Len().
func NextPermutation(data Interface) bool {
	if data == nil {
		return false
	}
	i := data.Len() - 2
	for i >= 0 && !data.Less(i, i+1) {
		i--
	}
	if i < 0 {
		return false
	}
	npbsi := &nextPermutationBinarySearchInterface{
		Data:   data,
		Target: i,
		Begin:  i + 1,
		End:    data.Len(),
	}
	j := npbsi.Begin + sequence.BinarySearchMaxLess(npbsi, nil) // target is set in npbsi.
	data.Swap(i, j)
	for i, j = i+1, data.Len()-1; i < j; i, j = i+1, j-1 {
		data.Swap(i, j)
	}
	return true
}

type nextPermutationBinarySearchInterface struct {
	Data   Interface
	Target int
	Begin  int
	End    int
}

func (npbsi *nextPermutationBinarySearchInterface) Len() int {
	if npbsi == nil || npbsi.Data == nil {
		return 0
	}
	return npbsi.End - npbsi.Begin
}

func (npbsi *nextPermutationBinarySearchInterface) Equal(i int, x interface{}) bool {
	panic(errors.AutoMsg("method Equal not implement"))
}

// Here, x is just a dummy argument.
// This method should act as a Greater() because the Data is in descending order.
func (npbsi *nextPermutationBinarySearchInterface) Less(i int, x interface{}) bool {
	return npbsi.Data.Less(npbsi.Target, i+npbsi.Begin)
}

func (npbsi *nextPermutationBinarySearchInterface) Greater(i int, x interface{}) bool {
	panic(errors.AutoMsg("method Greater not implement"))
}
