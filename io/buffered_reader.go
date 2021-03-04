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

package io

import (
	"bufio"
	stdio "io"

	"github.com/donyori/gogo/errors"
)

// BufferedReader is an interface for a reader with a buffer.
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

	// Size returns the size of the underlying buffer in bytes.
	Size() int

	// Buffered returns the number of bytes
	// that can be read from the current buffer.
	Buffered() int

	// Peek returns the next n bytes without advancing the reader.
	//
	// The bytes stop being valid at the next read call.
	// If it returns fewer than n bytes,
	// it also returns an error explaining why the read is short.
	// The error is bufio.ErrBufferFull if n is larger than its buffer size.
	// (To test whether err is bufio.ErrBufferFull, use function errors.Is.)
	//
	// Calling Peek prevents a UnreadByte or UnreadRune call from succeeding
	// until the next read operation.
	Peek(n int) (data []byte, err error)

	// Discard skips the next n bytes and returns the number of bytes discarded.
	//
	// If it skips fewer than n bytes, it also returns an error explaining why.
	//
	// If 0 <= n <= Buffered(),
	// it is guaranteed to succeed without reading from the underlying reader.
	Discard(n int) (discarded int, err error)
}

// ResettableBufferedReader is an interface
// combining BufferedReader and ReaderResetter.
type ResettableBufferedReader interface {
	BufferedReader
	ReaderResetter
}

// defaultBufferSize is the default buffer size used by
// function NewBufferedReader.
const defaultBufferSize = 4096

// resettableBufferedReader is an implementation of
// interface ResettableBufferedReader.
type resettableBufferedReader struct {
	br *bufio.Reader
}

// NewBufferedReader creates a BufferedReader on r,
// whose buffer has at least the default size (4096 bytes).
func NewBufferedReader(r stdio.Reader) ResettableBufferedReader {
	return NewBufferedReaderSize(r, defaultBufferSize)
}

// NewBufferedReaderSize creates a BufferedReader on r,
// whose buffer has at least the specified size.
//
// If r is a BufferedReader with a large enough buffer, it returns r directly.
func NewBufferedReaderSize(r stdio.Reader, size int) ResettableBufferedReader {
	if br, ok := r.(ResettableBufferedReader); ok && br.Size() >= size {
		return br
	}
	if br, ok := r.(*resettableBufferedReader); ok {
		br = &resettableBufferedReader{bufio.NewReaderSize(br.br, size)}
		return br
	}
	br, ok := r.(*bufio.Reader)
	if !ok || br.Size() < size {
		br = bufio.NewReaderSize(r, size)
	}
	return &resettableBufferedReader{br}
}

// Read reads data into p.
//
// It returns the number of bytes read into p.
// The bytes are taken from at most one Read on the underlying reader,
// hence n may be less than len(p).
//
// To read exactly len(p) bytes, use io.ReadFull(b, p).
// At EOF, the count will be zero and err will be io.EOF.
func (rbr *resettableBufferedReader) Read(p []byte) (n int, err error) {
	n, err = rbr.br.Read(p)
	return n, errors.AutoWrap(err)
}

// ReadByte reads and returns a single byte.
//
// If no byte is available, returns an error.
func (rbr *resettableBufferedReader) ReadByte() (byte, error) {
	c, err := rbr.br.ReadByte()
	return c, errors.AutoWrap(err)
}

// UnreadByte unreads the last byte.
// Only the most recently read byte can be unread.
//
// UnreadByte returns an error if the most recent method called on the
// reader was not a read operation.
// Notably, Peek is not considered a read operation.
func (rbr *resettableBufferedReader) UnreadByte() error {
	return errors.AutoWrap(rbr.br.UnreadByte())
}

// ReadRune reads a single UTF-8 encoded Unicode character and
// returns the rune and its size in bytes.
//
// If the encoded rune is invalid,
// it consumes one byte and returns unicode.ReplacementChar (U+FFFD)
// with a size of 1.
func (rbr *resettableBufferedReader) ReadRune() (r rune, size int, err error) {
	r, size, err = rbr.br.ReadRune()
	return r, size, errors.AutoWrap(err)
}

// UnreadRune unreads the last rune.
//
// If the most recent method called on the reader was not a ReadRune,
// UnreadRune returns an error.
// (In this regard it is stricter than UnreadByte,
// which will unread the last byte from any read operation.)
func (rbr *resettableBufferedReader) UnreadRune() error {
	return errors.AutoWrap(rbr.br.UnreadRune())
}

// WriteTo writes data to w until there's no more data to write or
// when an error occurs.
//
// The return value n is the number of bytes written.
// Any error encountered during the write is also returned.
//
// This may make multiple calls to the Read method of the underlying reader.
//
// If the underlying reader supports the WriteTo method,
// this calls the underlying WriteTo without buffering.
func (rbr *resettableBufferedReader) WriteTo(w stdio.Writer) (n int64, err error) {
	n, err = rbr.br.WriteTo(w)
	return n, errors.AutoWrap(err)
}

// ReadLine reads a line excluding the end-of-line bytes.
//
// If the line is too long for the buffer,
// then more is set and the beginning of the line is returned.
// The rest of the line will be returned from future calls.
// more will be false when returning the last fragment of the line.
//
// It either returns a non-nil line or it returns an error, never both.
//
// Caller should not keep the return value line,
// and line is only valid until the next call to the reader,
// including the method ReadLine and any other possible methods.
func (rbr *resettableBufferedReader) ReadLine() (line []byte, more bool, err error) {
	line, more, err = rbr.br.ReadLine()
	return line, more, errors.AutoWrap(err)
}

// WriteLineTo reads a line from its underlying reader and writes it to w.
//
// It stops writing data if an error occurs.
//
// It returns the number of bytes written to w and any error encountered.
func (rbr *resettableBufferedReader) WriteLineTo(w stdio.Writer) (n int64, err error) {
	var line []byte
	var written int
	var errList errors.ErrorList
	more := true
	for more {
		line, more, err = rbr.ReadLine()
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
			return n, errors.AutoWrap(err)
		}
	}
	return
}

// Size returns the size of the underlying buffer in bytes.
func (rbr *resettableBufferedReader) Size() int {
	return rbr.br.Size()
}

// Buffered returns the number of bytes
// that can be read from the current buffer.
func (rbr *resettableBufferedReader) Buffered() int {
	return rbr.br.Buffered()
}

// Peek returns the next n bytes without advancing the reader.
//
// The bytes stop being valid at the next read call.
// If it returns fewer than n bytes,
// it also returns an error explaining why the read is short.
// The error is bufio.ErrBufferFull if n is larger than its buffer size.
// (To test whether err is bufio.ErrBufferFull, use function errors.Is.)
//
// Calling Peek prevents a UnreadByte or UnreadRune call from succeeding
// until the next read operation.
func (rbr *resettableBufferedReader) Peek(n int) (data []byte, err error) {
	data, err = rbr.br.Peek(n)
	return data, errors.AutoWrap(err)
}

// Discard skips the next n bytes and returns the number of bytes discarded.
//
// If it skips fewer than n bytes, it also returns an error explaining why.
//
// If 0 <= n <= Buffered(),
// it is guaranteed to succeed without reading from the underlying reader.
func (rbr *resettableBufferedReader) Discard(n int) (discarded int, err error) {
	discarded, err = rbr.br.Discard(n)
	return discarded, errors.AutoWrap(err)
}

// Reset resets all states and switches to read from r.
func (rbr *resettableBufferedReader) Reset(r stdio.Reader) {
	rbr.br.Reset(r)
}
