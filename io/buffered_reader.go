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

import (
	"bufio"
	stdio "io"

	"github.com/donyori/gogo/errors"
)

// An interface for buffered reader.
//
// Note that bufio.Reader implements all methods of this interface
// except WriteLineTo of LineWriterTo.
type BufferedReader interface {
	stdio.Reader
	stdio.ByteScanner
	stdio.RuneScanner
	stdio.WriterTo
	LineReader
	LineWriterTo

	// Return the size of the underlying buffer in bytes.
	Size() int

	// Return the number of bytes that can be read from the current buffer.
	Buffered() int

	// Return the next n bytes without advancing the reader. The bytes stop
	// being valid at the next read call. If it returns fewer than n bytes, it
	// also returns an error explaining why the read is short. The error is
	// bufio.ErrBufferFull if n is larger than its buffer size.
	//
	// Calling Peek prevents a UnreadByte or UnreadRune call from succeeding
	// until the next read operation.
	Peek(n int) (data []byte, err error)

	// Discard the next n bytes and return the number of bytes discarded. If it
	// discards fewer than n bytes, it also returns an error explaining why.
	Discard(n int) (discarded int, err error)
}

// An interface combining BufferedReader and ReaderResetter.
type ResettableBufferedReader interface {
	BufferedReader
	ReaderResetter
}

const defaultBufferSize = 4096

type resettableBufferedReader struct {
	*bufio.Reader
}

// Create a BufferedReader on r, whose buffer has at least the default size.
func NewBufferedReader(r stdio.Reader) ResettableBufferedReader {
	return NewBufferedReaderSize(r, defaultBufferSize)
}

// Create a BufferedReader on r, whose buffer has at least given size.
// If r is a BufferedReader with large enough buffer, it returns r directly.
func NewBufferedReaderSize(r stdio.Reader, size int) ResettableBufferedReader {
	if br, ok := r.(ResettableBufferedReader); ok && br.Size() >= size {
		return br
	}
	if br, ok := r.(*resettableBufferedReader); ok {
		br = &resettableBufferedReader{Reader: bufio.NewReaderSize(br.Reader, size)}
		return br
	}
	br, ok := r.(*bufio.Reader)
	if !ok || br.Size() < size {
		br = bufio.NewReaderSize(r, size)
	}
	return &resettableBufferedReader{Reader: br}
}

func (r *resettableBufferedReader) WriteLineTo(w stdio.Writer) (n int64, err error) {
	var line []byte
	var written int
	var errList errors.ErrorList
	more := true
	for more {
		line, more, err = r.ReadLine()
		if err != nil {
			errList.Append(err)
		}
		if len(line) > 0 {
			written, err = w.Write(line)
			n += int64(written)
			if err != nil {
				errList.Append(err)
			}
		}
		if err = errList.ToError(); err != nil {
			// Don't wrap err to keep consistent with other methods of bufio.Reader.
			return
		}
	}
	return
}
