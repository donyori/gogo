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

package permutation

import (
	"sort"

	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/function"
)

// Interface represents an integer-indexed permutation.
// It is the same as sort.Interface.
type Interface = sort.Interface

// ArrayAdapter is an adapter for:
// sequence.Array + function.LessFunc -> Interface.
type ArrayAdapter struct {
	Data   sequence.Array
	LessFn function.LessFunc
}

// Len returns the number of items in the array.
func (ad *ArrayAdapter) Len() int {
	if ad == nil || ad.Data == nil {
		return 0
	}
	return ad.Data.Len()
}

// Less reports whether the item with index i must sort before
// the item with index j.
func (ad *ArrayAdapter) Less(i, j int) bool {
	return ad.LessFn(ad.Data.Get(i), ad.Data.Get(j))
}

// Swap swaps the items with indexes i and j.
func (ad *ArrayAdapter) Swap(i, j int) {
	ad.Data.Swap(i, j)
}
