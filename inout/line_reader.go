// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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

// LineReader is an interface that wraps method ReadLine.
//
// It may be useful to read long lines
// that are hard to be loaded in a buffer once.
type LineReader interface {
	// ReadLine reads a line excluding the end-of-line bytes.
	//
	// If the line is too long for the buffer,
	// then more is set and the beginning of the line is returned.
	// The rest of the line will be returned from future calls.
	// more will be false when returning the last fragment of the line.
	//
	// It either returns a non-nil line or it returns an error, never both.
	// If an error (including io.EOF) occurs after reading some content,
	// it returns the content as a line and a nil error.
	// The error encountered will be reported on future read calls.
	//
	// No indication or error is given if the input ends
	// without a final line end.
	// Even if the input ends without end-of-line bytes,
	// the content before EOF is treated as a line.
	//
	// Caller should not keep the return value line,
	// and line is only valid until the next call to the reader,
	// including the method ReadLine and any other possible methods.
	ReadLine() (line []byte, more bool, err error)
}

// EntireLineReader is an interface that wraps method ReadEntireLine.
type EntireLineReader interface {
	// ReadEntireLine reads an entire line excluding the end-of-line bytes.
	//
	// It either returns a non-nil line or it returns an error, never both.
	// If an error (including io.EOF) occurs after reading some content,
	// it returns the content as a line and a nil error.
	// The error encountered will be reported on future read calls.
	//
	// No indication or error is given if the input ends
	// without a final line end.
	// Even if the input ends without end-of-line bytes,
	// the content before EOF is treated as a line.
	//
	// Unlike the method ReadLine of interface LineReader,
	// the returned line is always valid.
	// Caller can keep the returned line safely.
	//
	// If the line is too long to be stored in a []byte
	// (hardly happens in text files), it may panic or report an error.
	ReadEntireLine() (line []byte, err error)
}

// LineWriterTo is an interface that wraps method WriteLineTo.
type LineWriterTo interface {
	// WriteLineTo reads a line excluding the end-of-line bytes
	// from its underlying reader and writes it to w.
	//
	// It stops writing data if an error occurs.
	//
	// It returns the number of bytes written to w and any error encountered.
	//
	// If an error (including io.EOF) occurs while reading from
	// the underlying reader, but some content has already been read,
	// it writes the content as a line and returns a nil error.
	// The error encountered will be reported on future read calls.
	//
	// No indication or error is given if the input ends
	// without a final line end.
	// Even if the input ends without end-of-line bytes,
	// the content before EOF is treated as a line.
	WriteLineTo(w io.Writer) (n int64, err error)
}
