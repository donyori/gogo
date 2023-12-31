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

package inout

import "io"

// Resetter is an interface that wraps method Reset,
// which resets all states of its instance.
type Resetter interface {
	// Reset resets all states.
	Reset()
}

// ReaderResetter is an interface that wraps method Reset,
// which resets all states of its instance and
// switches to read from the reader r.
type ReaderResetter interface {
	// Reset resets all states and switches to read from r.
	Reset(r io.Reader)
}

// WriterResetter is an interface that wraps method Reset,
// which discards any unflushed data, resets all states of its instance,
// and switches to write to the writer w.
type WriterResetter interface {
	// Reset discards any unflushed data, resets all states,
	// and switches to write to w.
	Reset(w io.Writer)
}
