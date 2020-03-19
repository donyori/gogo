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

package io

import stdio "io"

// An interface that wraps method ReadLine. It may be useful to read long lines
// that are hard to be loaded in a buffer once.
type LineReader interface {
	// Read a line excluding the end-of-line bytes. If the line was too long
	// for the buffer then more is set and the beginning of the line is returned.
	// The rest of the line will be returned from future calls. more will be
	// false when returning the last fragment of the line. It either returns a
	// non-nil line or it returns an error, never both.
	//
	// Caller shouldn't keep the return value line, and line is only valid
	// until the next call to the reader, including the method ReadLine and
	// any other possible methods.
	ReadLine() (line []byte, more bool, err error)
}

// An interface that wraps method WriteLineTo.
type LineWriterTo interface {
	// Write a line to w. It stops to write data if an error occurs. It returns
	// the number of bytes written to w and any encountered error.
	WriteLineTo(w stdio.Writer) (n int64, err error)
}
