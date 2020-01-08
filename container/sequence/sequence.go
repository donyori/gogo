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

package sequence

// General sequence interface.
type Sequence interface {
	// Return the number of items in the sequence.
	// It returns 0 if the sequence is nil.
	Len() int

	// Return the first item of the sequence.
	// It panics if the sequence is nil or empty.
	Front() interface{}

	// Set the first item to x.
	// It panics if the sequence is nil or empty.
	SetFront(x interface{})

	// Return the last item of the sequence.
	// It panics if the sequence is nil or empty.
	Back() interface{}

	// Set the last item to x.
	// It panics if the sequence is nil or empty.
	SetBack(x interface{})

	// Reverse items of the sequence.
	Reverse()

	// Scan the items in the sequence from the first to the last.
	Scan(handler func(x interface{}) (cont bool))
}
