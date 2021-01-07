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

package sorta

import (
	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/function"
)

// Array is an adapter for:
// sequence.Array + function.LessFunc -> sort.Interface.
type Array struct {
	Data   sequence.Array
	LessFn function.LessFunc
}

// Len returns the number of items in the array.
func (a *Array) Len() int {
	if a == nil || a.Data == nil {
		return 0
	}
	return a.Data.Len()
}

// Less reports whether the item with
// index i should sort before the item with index j.
func (a *Array) Less(i, j int) bool {
	return a.LessFn(a.Data.Get(i), a.Data.Get(j))
}

// Swap swaps the items with indexes i and j.
func (a *Array) Swap(i, j int) {
	a.Data.Swap(i, j)
}
