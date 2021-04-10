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

package heapa

import (
	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/function"
)

// DynamicArray is an adapter for:
// sequence.DynamicArray + function.LessFunc -> container/heap.Interface.
type DynamicArray struct {
	Data   sequence.DynamicArray
	LessFn function.LessFunc
}

// Len returns the number of items in the dynamic array.
func (da *DynamicArray) Len() int {
	if da == nil || da.Data == nil {
		return 0
	}
	return da.Data.Len()
}

// Less reports whether the item with index i must sort before
// the item with index j.
func (da *DynamicArray) Less(i, j int) bool {
	return da.LessFn(da.Data.Get(i), da.Data.Get(j))
}

// Swap swaps the items with indexes i and j.
func (da *DynamicArray) Swap(i, j int) {
	da.Data.Swap(i, j)
}

// Push adds x as the last item of the dynamic array.
func (da *DynamicArray) Push(x interface{}) {
	da.Data.Push(x)
}

// Pop removes and returns the last item of the dynamic array.
func (da *DynamicArray) Pop() interface{} {
	return da.Data.Pop()
}
