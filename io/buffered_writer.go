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

// BufferedWriter is an interface for a writer with a buffer.
//
// Note that after all data has been written,
// the client should call the method Flush to guarantee that
// all data has been forwarded to the underlying writer.
//
// To get a BufferedWriter, use function NewBufferedWriter
// or NewBufferedWriterSize.
type BufferedWriter interface {
	stdio.Writer
	stdio.ByteWriter
	stdio.StringWriter
	stdio.ReaderFrom
	Flusher

	// Size returns the size of the underlying buffer in bytes.
	Size() int

	// Buffered returns the number of bytes that
	// have been written into the current buffer.
	Buffered() int

	// Available returns the number of bytes unused in the current buffer.
	Available() int

	// WriteRune writes a single Unicode code point.
	//
	// It returns the number of bytes written and any write error encountered.
	WriteRune(r rune) (size int, err error)
}

// ResettableBufferedWriter is an interface
// combining BufferedWriter and WriterResetter.
//
// To get a ResettableBufferedWriter, use function NewBufferedWriter
// or NewBufferedWriterSize.
type ResettableBufferedWriter interface {
	BufferedWriter
	WriterResetter
}

// resettableBufferedWriter is an implementation of
// interface ResettableBufferedWriter based on bufio.Writer.
type resettableBufferedWriter struct {
	bw *bufio.Writer
}

// NewBufferedWriter creates a ResettableBufferedWriter on w,
// whose buffer has at least the default size (4096 bytes).
func NewBufferedWriter(w stdio.Writer) ResettableBufferedWriter {
	return NewBufferedWriterSize(w, defaultBufferSize)
}

// NewBufferedWriterSize creates a ResettableBufferedWriter on w,
// whose buffer has at least the specified size.
//
// If size is non-positive, it will use the default size (4096) instead.
//
// If w is a ResettableBufferedWriter with a large enough buffer,
// it returns w directly.
func NewBufferedWriterSize(w stdio.Writer, size int) ResettableBufferedWriter {
	if size <= 0 {
		size = defaultBufferSize
	}
	if bw, ok := w.(ResettableBufferedWriter); ok && bw.Size() >= size {
		return bw
	}
	if bw, ok := w.(*resettableBufferedWriter); ok {
		bw = &resettableBufferedWriter{bufio.NewWriterSize(bw.bw, size)}
		return bw
	}
	bw, ok := w.(*bufio.Writer)
	if !ok || bw.Size() < size {
		bw = bufio.NewWriterSize(w, size)
	}
	return &resettableBufferedWriter{bw}
}

// Write writes the contents of p into the buffer.
//
// It returns the number of bytes written and any write error encountered.
// If n < len(p), it also returns an error explaining why the write is short.
func (rbw *resettableBufferedWriter) Write(p []byte) (n int, err error) {
	n, err = rbw.bw.Write(p)
	return n, errors.AutoWrap(err)
}

// WriteByte writes a single byte.
//
// It returns any write error encountered.
func (rbw *resettableBufferedWriter) WriteByte(c byte) error {
	return errors.AutoWrap(rbw.bw.WriteByte(c))
}

// WriteString writes a string.
//
// It returns the number of bytes written and any write error encountered.
// If n is less than len(s),
// it also returns an error explaining why the write is short.
func (rbw *resettableBufferedWriter) WriteString(s string) (n int, err error) {
	n, err = rbw.bw.WriteString(s)
	return n, errors.AutoWrap(err)
}

// ReadFrom reads data from r until EOF or error.
//
// The return value n is the number of bytes read.
// Any error except EOF encountered during the read is also returned.
//
// If the underlying writer supports the ReadFrom method,
// and this buffered writer has no buffered data yet,
// it calls the underlying ReadFrom without buffering.
func (rbw *resettableBufferedWriter) ReadFrom(r stdio.Reader) (n int64, err error) {
	n, err = rbw.bw.ReadFrom(r)
	return n, errors.AutoWrap(err)
}

// Flush writes any buffered data to the underlying writer.
//
// It returns any write error encountered.
func (rbw *resettableBufferedWriter) Flush() error {
	return errors.AutoWrap(rbw.bw.Flush())
}

// Size returns the size of the underlying buffer in bytes.
func (rbw *resettableBufferedWriter) Size() int {
	return rbw.bw.Size()
}

// Buffered returns the number of bytes that
// have been written into the current buffer.
func (rbw *resettableBufferedWriter) Buffered() int {
	return rbw.bw.Buffered()
}

// Available returns the number of bytes unused in the current buffer.
func (rbw *resettableBufferedWriter) Available() int {
	return rbw.bw.Available()
}

// WriteRune writes a single Unicode code point.
//
// It returns the number of bytes written and any write error encountered.
func (rbw *resettableBufferedWriter) WriteRune(r rune) (size int, err error) {
	size, err = rbw.bw.WriteRune(r)
	return size, errors.AutoWrap(err)
}

// Reset discards any unflushed data, resets all states,
// and switches to write to w.
func (rbw *resettableBufferedWriter) Reset(w stdio.Writer) {
	rbw.bw.Reset(w)
}
