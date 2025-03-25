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
	"iter"

	"github.com/donyori/gogo/container"
)

// Sequence is an interface representing a general sequence.
//
// Its method Range accesses the items from first to last.
type Sequence[Item any] interface {
	container.Container[Item]

	// RangeBackward is like Range,
	// but the order of access is from last to first.
	RangeBackward(handler func(x Item) (cont bool))

	// IterItemsBackward returns an iterator over all items in the sequence,
	// traversing it from last to first.
	//
	// The returned iterator is always non-nil.
	IterItemsBackward() iter.Seq[Item]

	// Front returns the first item.
	//
	// It panics if the sequence is nil or empty.
	Front() Item

	// SetFront sets the first item to x.
	//
	// It panics if the sequence is nil or empty.
	SetFront(x Item)

	// Back returns the last item.
	//
	// It panics if the sequence is nil or empty.
	Back() Item

	// SetBack sets the last item to x.
	//
	// It panics if the sequence is nil or empty.
	SetBack(x Item)

	// Reverse turns items in the sequence the other way round.
	Reverse()
}

// OrderedSequence is an interface representing a sequence that can be sorted.
type OrderedSequence[Item any] interface {
	Sequence[Item]

	// Min returns the first smallest item in the sequence.
	//
	// It panics if the sequence is nil or empty.
	Min() Item

	// Max returns the first largest item in the sequence.
	//
	// It panics if the sequence is nil or empty.
	Max() Item

	// IsSorted reports whether the sequence is sorted in ascending order.
	//
	// If the sequence is empty, IsSorted returns true.
	IsSorted() bool

	// Sort sorts the sequence in ascending order.
	// This sort is not guaranteed to be stable.
	Sort()

	// SortStable sorts the sequence in ascending order
	// while keeping the original order of equal elements.
	SortStable()
}
