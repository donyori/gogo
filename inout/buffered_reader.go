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

import (
	"bufio"
	"io"

	"github.com/donyori/gogo/errors"
)

// BufferedReader is an interface for a reader with a buffer.
//
// To get a BufferedReader, use function NewBufferedReader
// or NewBufferedReaderSize.
type BufferedReader interface {
	io.Reader
	io.ByteScanner
	ByteConsumer
	io.RuneScanner
	RuneConsumer
	io.WriterTo
	LineReader
	EntireLineReader
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
	// Calling Peek prevents an UnreadByte or UnreadRune call from succeeding
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
//
// To get a ResettableBufferedReader, use function NewBufferedReader
// or NewBufferedReaderSize.
type ResettableBufferedReader interface {
	BufferedReader
	ReaderResetter
}

// defaultBufferSize is the default buffer size used by functions
// NewBufferedReader, NewBufferedWriter, and NewBufferedWriterSize.
const defaultBufferSize int = 4096

// minReadBufferSize is the minimum buffer size of the reader
// used by function NewBufferedReaderSize.
const minReadBufferSize int = 16

// resettableBufferedReader is an implementation of
// interface ResettableBufferedReader based on bufio.Reader.
type resettableBufferedReader struct {
	br *bufio.Reader
}

// NewBufferedReader creates a ResettableBufferedReader on r,
// whose buffer has at least the default size (4096 bytes).
//
// The reader r can be nil,
// in which case NewBufferedReader only allocates the buffer,
// and the reader can be set later via the method Reset.
// Note that reading before setting up a valid reader may cause panic.
func NewBufferedReader(r io.Reader) ResettableBufferedReader {
	return NewBufferedReaderSize(r, defaultBufferSize)
}

// NewBufferedReaderSize creates a ResettableBufferedReader on r,
// whose buffer has at least the specified size.
//
// If size is less than 16, it uses 16 instead.
//
// The reader r can be nil,
// in which case NewBufferedReaderSize only allocates the buffer,
// and the reader can be set later via the method Reset.
// Note that reading before setting up a valid reader may cause panic.
//
// If r is a ResettableBufferedReader with a large enough buffer,
// it returns r directly.
func NewBufferedReaderSize(r io.Reader, size int) ResettableBufferedReader {
	if size < minReadBufferSize {
		size = minReadBufferSize
	}
	if br, ok := r.(ResettableBufferedReader); ok && br.Size() >= size {
		return br
	} else if br, ok := r.(*resettableBufferedReader); ok {
		return &resettableBufferedReader{bufio.NewReaderSize(br.br, size)}
	}
	br, ok := r.(*bufio.Reader)
	if !ok || br == nil || br.Size() < size {
		if ok && br == nil {
			// If r is a nil *bufio.Reader, avoid panicking here.
			// Panic should occur when reading.
			r = nil
		}
		br = bufio.NewReaderSize(r, size)
	}
	return &resettableBufferedReader{br}
}

func (rbr *resettableBufferedReader) Read(p []byte) (n int, err error) {
	n, err = rbr.br.Read(p)
	return n, errors.AutoWrap(err)
}

func (rbr *resettableBufferedReader) ReadByte() (byte, error) {
	c, err := rbr.br.ReadByte()
	return c, errors.AutoWrap(err)
}

func (rbr *resettableBufferedReader) UnreadByte() error {
	return errors.AutoWrap(rbr.br.UnreadByte())
}

func (rbr *resettableBufferedReader) ConsumeByte(target byte, n int64) (
	consumed int64, err error) {
	consumed, err = rbr.ConsumeByteFunc(func(c byte) bool {
		return c == target
	}, n)
	return consumed, errors.AutoWrap(err)
}

func (rbr *resettableBufferedReader) ConsumeByteFunc(
	f func(c byte) bool, n int64) (consumed int64, err error) {
	for n < 0 || consumed < n {
		c, err := rbr.br.ReadByte()
		if err != nil {
			return consumed, errors.AutoWrap(err)
		} else if !f(c) {
			err = rbr.br.UnreadByte()
			if err != nil {
				// Fail to put the byte back. The byte has been consumed.
				consumed++
			}
			return consumed, errors.AutoWrap(err)
		}
		consumed++
	}
	return
}

func (rbr *resettableBufferedReader) ReadRune() (r rune, size int, err error) {
	r, size, err = rbr.br.ReadRune()
	return r, size, errors.AutoWrap(err)
}

func (rbr *resettableBufferedReader) UnreadRune() error {
	return errors.AutoWrap(rbr.br.UnreadRune())
}

func (rbr *resettableBufferedReader) ConsumeRune(target rune, n int64) (
	consumed int64, err error) {
	consumed, err = rbr.ConsumeRuneFunc(func(r rune, size int) bool {
		return r == target
	}, n)
	return consumed, errors.AutoWrap(err)
}

func (rbr *resettableBufferedReader) ConsumeRuneFunc(
	f func(r rune, size int) bool, n int64) (consumed int64, err error) {
	for n < 0 || consumed < n {
		r, size, err := rbr.br.ReadRune()
		if err != nil {
			return consumed, errors.AutoWrap(err)
		} else if !f(r, size) {
			err = rbr.br.UnreadRune()
			if err != nil {
				// Fail to put the rune back. The rune has been consumed.
				consumed++
			}
			return consumed, errors.AutoWrap(err)
		}
		consumed++
	}
	return
}

func (rbr *resettableBufferedReader) WriteTo(w io.Writer) (
	n int64, err error) {
	n, err = rbr.br.WriteTo(w)
	return n, errors.AutoWrap(err)
}

func (rbr *resettableBufferedReader) ReadLine() (
	line []byte, more bool, err error) {
	line, more, err = rbr.br.ReadLine()
	return line, more, errors.AutoWrap(err)
}

func (rbr *resettableBufferedReader) ReadEntireLine() (
	line []byte, err error) {
	var parts [][]byte
	var n int
	more := true
	for more {
		var t []byte
		t, more, err = rbr.ReadLine()
		if len(t) > 0 {
			buf := make([]byte, len(t))
			n += copy(buf, t)
			parts = append(parts, buf)
		}
	}
	if n == 0 {
		return nil, errors.AutoWrap(err)
	}
	// If n > 0, set err to nil, as described in the document.
	if len(parts) == 1 {
		return parts[0], nil
	}
	line, err = make([]byte, n), nil
	n = 0
	for i := range parts {
		n += copy(line[n:], parts[i])
	}
	return
}

func (rbr *resettableBufferedReader) WriteLineTo(w io.Writer) (
	n int64, err error) {
	errList, more := errors.NewErrorList(true), true
	for more {
		var line []byte
		line, more, err = rbr.br.ReadLine()
		if err != nil {
			errList.Append(err)
		}
		if len(line) > 0 {
			written, err := w.Write(line)
			n += int64(written)
			if err != nil {
				errList.Append(err)
			}
		}
		if errList.Erroneous() {
			return n, errors.AutoWrap(errList.ToError())
		}
	}
	return // err must be nil
}

func (rbr *resettableBufferedReader) Size() int {
	return rbr.br.Size()
}

func (rbr *resettableBufferedReader) Buffered() int {
	return rbr.br.Buffered()
}

func (rbr *resettableBufferedReader) Peek(n int) (data []byte, err error) {
	data, err = rbr.br.Peek(n)
	return data, errors.AutoWrap(err)
}

func (rbr *resettableBufferedReader) Discard(n int) (
	discarded int, err error) {
	discarded, err = rbr.br.Discard(n)
	return discarded, errors.AutoWrap(err)
}

func (rbr *resettableBufferedReader) Reset(r io.Reader) {
	if rbr == r {
		return // do nothing if the resettableBufferedReader is reset to itself
	}
	rbr.br.Reset(r)
}
