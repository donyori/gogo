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
func NewBufferedWriter(w io.Writer) ResettableBufferedWriter {
	return NewBufferedWriterSize(w, defaultBufferSize)
}

// NewBufferedWriterSize creates a ResettableBufferedWriter on w,
// whose buffer has at least the specified size.
//
// If size is non-positive, it will use the default size (4096) instead.
//
// If w is a ResettableBufferedWriter with a large enough buffer,
// it returns w directly.
func NewBufferedWriterSize(w io.Writer, size int) ResettableBufferedWriter {
	if size <= 0 {
		size = defaultBufferSize
	}
	if bw, ok := w.(ResettableBufferedWriter); ok && bw.Size() >= size {
		return bw
	}
	if bw, ok := w.(*resettableBufferedWriter); ok {
		return &resettableBufferedWriter{bufio.NewWriterSize(bw.bw, size)}
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

// MustWrite is like Write but panics when encountering an error.
//
// If it panics, the error value passed to the call of panic
// must be exactly of type *WritePanic.
func (rbw *resettableBufferedWriter) MustWrite(p []byte) (n int) {
	n, err := rbw.bw.Write(p)
	if err != nil {
		panic(NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

// WriteByte writes a single byte.
//
// It returns any write error encountered.
func (rbw *resettableBufferedWriter) WriteByte(c byte) error {
	return errors.AutoWrap(rbw.bw.WriteByte(c))
}

// MustWriteByte is like WriteByte but panics when encountering an error.
//
// If it panics, the error value passed to the call of panic
// must be exactly of type *WritePanic.
func (rbw *resettableBufferedWriter) MustWriteByte(c byte) {
	err := rbw.bw.WriteByte(c)
	if err != nil {
		panic(NewWritePanic(errors.AutoWrap(err)))
	}
}

// WriteRune writes a single Unicode code point.
//
// It returns the number of bytes written and any write error encountered.
func (rbw *resettableBufferedWriter) WriteRune(r rune) (size int, err error) {
	size, err = rbw.bw.WriteRune(r)
	return size, errors.AutoWrap(err)
}

// MustWriteRune is like WriteRune but panics when encountering an error.
//
// If it panics, the error value passed to the call of panic
// must be exactly of type *WritePanic.
func (rbw *resettableBufferedWriter) MustWriteRune(r rune) (size int) {
	size, err := rbw.bw.WriteRune(r)
	if err != nil {
		panic(NewWritePanic(errors.AutoWrap(err)))
	}
	return
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

// MustWriteString is like WriteString but panics when encountering an error.
//
// If it panics, the error value passed to the call of panic
// must be exactly of type *WritePanic.
func (rbw *resettableBufferedWriter) MustWriteString(s string) (n int) {
	n, err := rbw.bw.WriteString(s)
	if err != nil {
		panic(NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

// ReadFrom reads data from r until EOF or error.
//
// The return value n is the number of bytes read.
// Any error except EOF encountered during the read is also returned.
//
// If the underlying writer supports the ReadFrom method,
// and this buffered writer has no buffered data yet,
// it calls the underlying ReadFrom without buffering.
func (rbw *resettableBufferedWriter) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = rbw.bw.ReadFrom(r)
	return n, errors.AutoWrap(err)
}

// Printf formats arguments and writes to the buffer.
// Arguments are handled in the manner of fmt.Printf.
func (rbw *resettableBufferedWriter) Printf(format string, a ...any) (n int, err error) {
	n, err = fmt.Fprintf(rbw.bw, format, a...)
	return n, errors.AutoWrap(err)
}

// MustPrintf is like Printf but panics when encountering an error.
//
// If it panics, the error value passed to the call of panic
// must be exactly of type *WritePanic.
func (rbw *resettableBufferedWriter) MustPrintf(format string, a ...any) (n int) {
	n, err := fmt.Fprintf(rbw.bw, format, a...)
	if err != nil {
		panic(NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

// Print formats arguments and writes to the buffer.
// Arguments are handled in the manner of fmt.Print.
func (rbw *resettableBufferedWriter) Print(a ...any) (n int, err error) {
	n, err = fmt.Fprint(rbw.bw, a...)
	return n, errors.AutoWrap(err)
}

// MustPrint is like Print but panics when encountering an error.
//
// If it panics, the error value passed to the call of panic
// must be exactly of type *WritePanic.
func (rbw *resettableBufferedWriter) MustPrint(a ...any) (n int) {
	n, err := fmt.Fprint(rbw.bw, a...)
	if err != nil {
		panic(NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

// Println formats arguments and writes to the buffer.
// Arguments are handled in the manner of fmt.Println.
func (rbw *resettableBufferedWriter) Println(a ...any) (n int, err error) {
	n, err = fmt.Fprintln(rbw.bw, a...)
	return n, errors.AutoWrap(err)
}

// MustPrintln is like Println but panics when encountering an error.
//
// If it panics, the error value passed to the call of panic
// must be exactly of type *WritePanic.
func (rbw *resettableBufferedWriter) MustPrintln(a ...any) (n int) {
	n, err := fmt.Fprintln(rbw.bw, a...)
	if err != nil {
		panic(NewWritePanic(errors.AutoWrap(err)))
	}
	return
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

// Reset discards any unflushed data, resets all states,
// and switches to write to w.
func (rbw *resettableBufferedWriter) Reset(w io.Writer) {
	rbw.bw.Reset(w)
}
