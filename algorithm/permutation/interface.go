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
	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/function"
)

// A type standing for an integer-indexed permutation.
// It is the same as sort.Interface.
type Interface interface {
	// Return the number of items in the permutation.
	// It returns 0 if the permutation is nil.
	Len() int

	// Test whether the i-th item is less than j-th item of the permutation.
	// It panics if i or j is out of range.
	Less(i, j int) bool

	// Swap the i-th and j-th items.
	// It panics if i or j is out of range.
	Swap(i, j int)
}

// An adapter for: Array + LessFunc -> Interface.
type ArrayAdapter struct {
	Data   sequence.Array
	LessFn function.LessFunc
}

func (ad *ArrayAdapter) Len() int {
	if ad == nil || ad.Data == nil {
		return 0
	}
	return ad.Data.Len()
}

func (ad *ArrayAdapter) Less(i, j int) bool {
	return ad.LessFn(ad.Data.Get(i), ad.Data.Get(j))
}

func (ad *ArrayAdapter) Swap(i, j int) {
	ad.Data.Swap(i, j)
}
