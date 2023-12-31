// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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

import "github.com/donyori/gogo/container"

// Sequence is an interface representing a general sequence.
//
// Its method Range accesses the items from first to last.
type Sequence[Item any] interface {
	container.Container[Item]

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
