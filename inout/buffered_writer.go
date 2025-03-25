// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2025  Yuan Gao
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
	"fmt"
	"io"

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
	Writer
	ByteWriter
	RuneWriter
	StringWriter
	io.ReaderFrom
	Printer
	Flusher

	// Size returns the size of the underlying buffer in bytes.
	Size() int

	// Buffered returns the number of bytes that
	// have been written into the current buffer.
	Buffered() int

	// Available returns the number of bytes unused in the current buffer.
	Available() int
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
//
// The writer w can be nil,
// in which case NewBufferedWriter only allocates the buffer,
// and the writer can be set later via the method Reset.
// Note that writing before setting up a valid writer may cause panic.
func NewBufferedWriter(w io.Writer) ResettableBufferedWriter {
	return NewBufferedWriterSize(w, defaultBufferSize)
}

// NewBufferedWriterSize creates a ResettableBufferedWriter on w,
// whose buffer has at least the specified size.
//
// If size is nonpositive, it uses the default size (4096) instead.
//
// The writer w can be nil,
// in which case NewBufferedWriterSize only allocates the buffer,
// and the writer can be set later via the method Reset.
// Note that writing before setting up a valid writer may cause panic.
//
// If w is a ResettableBufferedWriter with a large enough buffer,
// it returns w directly.
func NewBufferedWriterSize(w io.Writer, size int) ResettableBufferedWriter {
	if size <= 0 {
		size = defaultBufferSize
	}
	if bw, ok := w.(ResettableBufferedWriter); ok && bw.Size() >= size {
		return bw
	} else if bw, ok := w.(*resettableBufferedWriter); ok {
		return &resettableBufferedWriter{bufio.NewWriterSize(bw.bw, size)}
	}
	bw, ok := w.(*bufio.Writer)
	if !ok || bw == nil || bw.Size() < size {
		if ok && bw == nil {
			// If w is a nil *bufio.Writer, avoid panicking here.
			// Panic should occur when writing.
			w = nil
		}
		bw = bufio.NewWriterSize(w, size)
	}
	return &resettableBufferedWriter{bw}
}

func (rbw *resettableBufferedWriter) Write(p []byte) (n int, err error) {
	n, err = rbw.bw.Write(p)
	return n, errors.AutoWrap(err)
}

func (rbw *resettableBufferedWriter) MustWrite(p []byte) (n int) {
	n, err := rbw.bw.Write(p)
	if err != nil {
		panic(NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

func (rbw *resettableBufferedWriter) WriteByte(c byte) error {
	return errors.AutoWrap(rbw.bw.WriteByte(c))
}

func (rbw *resettableBufferedWriter) MustWriteByte(c byte) {
	err := rbw.bw.WriteByte(c)
	if err != nil {
		panic(NewWritePanic(errors.AutoWrap(err)))
	}
}

func (rbw *resettableBufferedWriter) WriteRune(r rune) (size int, err error) {
	size, err = rbw.bw.WriteRune(r)
	return size, errors.AutoWrap(err)
}

func (rbw *resettableBufferedWriter) MustWriteRune(r rune) (size int) {
	size, err := rbw.bw.WriteRune(r)
	if err != nil {
		panic(NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

func (rbw *resettableBufferedWriter) WriteString(s string) (n int, err error) {
	n, err = rbw.bw.WriteString(s)
	return n, errors.AutoWrap(err)
}

func (rbw *resettableBufferedWriter) MustWriteString(s string) (n int) {
	n, err := rbw.bw.WriteString(s)
	if err != nil {
		panic(NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

func (rbw *resettableBufferedWriter) ReadFrom(r io.Reader) (
	n int64, err error) {
	n, err = rbw.bw.ReadFrom(r)
	return n, errors.AutoWrap(err)
}

func (rbw *resettableBufferedWriter) Printf(format string, args ...any) (
	n int, err error) {
	n, err = fmt.Fprintf(rbw.bw, format, args...)
	return n, errors.AutoWrap(err)
}

func (rbw *resettableBufferedWriter) MustPrintf(format string, args ...any) (
	n int) {
	n, err := fmt.Fprintf(rbw.bw, format, args...)
	if err != nil {
		panic(NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

func (rbw *resettableBufferedWriter) Print(args ...any) (n int, err error) {
	n, err = fmt.Fprint(rbw.bw, args...)
	return n, errors.AutoWrap(err)
}

func (rbw *resettableBufferedWriter) MustPrint(args ...any) (n int) {
	n, err := fmt.Fprint(rbw.bw, args...)
	if err != nil {
		panic(NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

func (rbw *resettableBufferedWriter) Println(args ...any) (n int, err error) {
	n, err = fmt.Fprintln(rbw.bw, args...)
	return n, errors.AutoWrap(err)
}

func (rbw *resettableBufferedWriter) MustPrintln(args ...any) (n int) {
	n, err := fmt.Fprintln(rbw.bw, args...)
	if err != nil {
		panic(NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

func (rbw *resettableBufferedWriter) Flush() error {
	return errors.AutoWrap(rbw.bw.Flush())
}

func (rbw *resettableBufferedWriter) Size() int {
	return rbw.bw.Size()
}

func (rbw *resettableBufferedWriter) Buffered() int {
	return rbw.bw.Buffered()
}

func (rbw *resettableBufferedWriter) Available() int {
	return rbw.bw.Available()
}

func (rbw *resettableBufferedWriter) Reset(w io.Writer) {
	if rbw == w {
		return // do nothing if the resettableBufferedWriter is reset to itself
	}
	rbw.bw.Reset(w)
}
