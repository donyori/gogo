// gogo. A Golang toolbox.
// Copyright (C) 2019 Yuan Gao
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
	"sort"

	"github.com/donyori/gogo/adapter/sortw"
	"github.com/donyori/gogo/function"
)

// Export github.com/donyori/gogo/function.LessFunc.
type LessFunc = function.LessFunc

// Transform data to its next permutation in lexical order.
// It returns false if len(data) == 0 or the permutations are exhausted.
func NextPermutationSlice(data []interface{}, lessFunc LessFunc) bool {
	itf := &sortw.Slice{
		Data:     data,
		LessFunc: lessFunc,
	}
	return NextPermutationSortItf(itf)
}

// Do the same thing as NextPermutationSlice,
// but work on sort.Interface.
func NextPermutationSortItf(data sort.Interface) bool {
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
	j := data.Len() - 1
	for !data.Less(i, j) {
		j--
	}
	data.Swap(i, j)
	for i, j = i+1, data.Len()-1; i < j; i, j = i+1, j-1 {
		data.Swap(i, j)
	}
	return true
}
