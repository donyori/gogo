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

type BufferedWriter interface {
	stdio.Writer
	stdio.ByteWriter
	stdio.StringWriter
	stdio.ReaderFrom
	Flusher

	// Return the size of the underlying buffer in bytes.
	Size() int

	// Return the number of bytes that have been written into the current buffer.
	Buffered() int

	// Return the number of bytes unused in the current buffer.
	Available() int

	// Write a single Unicode code point.
	// It returns the number of bytes written and any encountered error.
	WriteRune(r rune) (size int, err error)
}

type ResettableBufferedWriter interface {
	BufferedWriter
	WriterResetter
}
