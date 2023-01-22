// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2023  Yuan Gao
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

package filesys

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/inout"
)

// WritableFile represents a single file that can be written to.
//
// It is similar to io/fs.File but has the method Write instead of Read.
type WritableFile interface {
	io.WriteCloser

	// Stat returns the io/fs.FileInfo structure describing file.
	Stat() (info fs.FileInfo, err error)
}

// WriteOptions are options for Write functions.
type WriteOptions struct {
	// True if not to compress the file with gzip and not to archive the file
	// with tar (i.e., tape archive) according to the file extension.
	Raw bool

	// The compression level for gzip.
	// Only take effect when Raw is false and the file extension
	// is either ".gz" or ".tgz".
	// The zero value (0) stands for no compression
	// other than the default value.
	// To use the default value, set it to compress/gzip.DefaultCompression.
	GzipLv int

	// Size of the buffer for writing the file at least.
	// Non-positive values for using default value.
	BufSize int
}

// defaultWriteOptions are default options for functions WriteTrunc,
// WriteAppend, and WriteExcl.
var defaultWriteOptions = &WriteOptions{GzipLv: gzip.DefaultCompression}

// Writer is a device to write data to a local file.
//
// Its method Close closes all closable objects opened by this writer
// (may include the file).
// After successfully closing this writer,
// its method Close will do nothing and return nil,
// and its write methods will report ErrFileWriterClosed.
// (To test whether the error is ErrFileWriterClosed, use function errors.Is.)
type Writer interface {
	inout.Closer
	inout.BufferedWriter

	// TarEnabled returns true if the file is archived by tar
	// (i.e., tape archive) and is not opened in raw mode.
	TarEnabled() bool

	// TarWriteHeader writes hdr and prepares to accept the content of
	// the next file.
	//
	// The tar.Header.Size determines how many bytes can be written for
	// the next file.
	// If the current file is not fully written, it will return an error.
	// It implicitly flushes any padding necessary before writing the header.
	//
	// If the file is not archived by tar, or the file is opened in raw mode,
	// it does nothing and reports ErrNotTar.
	// (To test whether the error is ErrNotTar, use function errors.Is.)
	TarWriteHeader(hdr *tar.Header) error

	// Options returns a copy of options used by this writer.
	Options() *WriteOptions

	// FileStat returns the io/fs.FileInfo structure describing file.
	FileStat() (info fs.FileInfo, err error)
}

// writer is an implementation of interface Writer.
//
// Use it with Write functions.
type writer struct {
	c    inout.Closer
	bw   inout.ResettableBufferedWriter
	uw   io.Writer // unbuffered writer
	opts WriteOptions
	f    WritableFile
	tw   *tar.Writer
}

// Write creates a writer on the specified file with options opts.
//
// If the file is a directory, Write reports ErrIsDir and returns a nil Writer.
// (To test whether err is ErrIsDir, use function errors.Is.)
//
// If opts are nil, the default options will be used.
// The default options are as follows:
//   - Raw: false,
//   - GzipLv: compress/gzip.DefaultCompression,
//   - BufSize: 0,
//
// To ensure that this function and the returned writer can work as expected,
// the specified file must not be operated by anyone else
// before closing the returned writer.
//
// closeFile indicates whether the writer should close the file
// when calling its method Close.
// If closeFile is false, the client is responsible for closing file
// after closing the writer.
// If closeFile is true, the client should not close the file,
// even if this function reports an error.
// In this case, the file will be closed during the method Close of the writer,
// and it will also be closed by this function when encountering an error.
//
// This function panics if file is nil.
func Write(file WritableFile, opts *WriteOptions, closeFile bool) (w Writer, err error) {
	if file == nil {
		panic(errors.AutoMsg("file is nil"))
	}
	info, err := file.Stat()
	if err != nil {
		return nil, errors.AutoWrap(err)
	} else if info.IsDir() {
		return nil, errors.AutoWrap(ErrIsDir)
	}

	if opts == nil {
		opts = defaultWriteOptions
	}

	el := errors.NewErrorList(true)
	defer func() {
		if el.Erroneous() {
			w, err = nil, errors.AutoWrapSkip(el.ToError(), 1) // skip = 1 to skip the inner function
		}
	}()

	closers := make([]io.Closer, 0, 3)
	if closeFile {
		closers = append(closers, file)
	}
	defer func() {
		if el.Erroneous() {
			for i := len(closers) - 1; i >= 0; i-- {
				el.Append(closers[i].Close())
			}
		}
	}()

	fw := &writer{
		uw:   file,
		opts: *opts,
		f:    file,
	}
	w = fw
	el.Append(writeSubRawClosersAndCreateBuffer(fw, info, &closers))
	return
}

// writeSubRawClosersAndCreateBuffer is a sub-process of function Write
// to deal with the options Raw, set fw.c, update closers, and create a buffer.
func writeSubRawClosersAndCreateBuffer(fw *writer, info fs.FileInfo, pClosers *[]io.Closer) error {
	if !fw.opts.Raw {
		name := strings.ToLower(path.Clean(filepath.ToSlash(info.Name())))
		ext := path.Ext(name)
		for {
			var endLoop bool
			switch ext {
			case ".gz", ".tgz":
				gw, err := gzip.NewWriterLevel(fw.uw, fw.opts.GzipLv)
				if err != nil {
					return err
				}
				*pClosers = append(*pClosers, gw)
				fw.uw = gw
				if ext == ".tgz" {
					ext = ".tar"
					continue
				}
			case ".tar":
				fw.tw = tar.NewWriter(fw.uw)
				*pClosers = append(*pClosers, fw.tw)
				fw.uw = fw.tw
				endLoop = true
			default:
				endLoop = true
			}
			if endLoop {
				break
			}
			name = name[:len(name)-len(ext)]
			ext = path.Ext(name)
		}
	}
	switch len(*pClosers) {
	case 1:
		fw.c = inout.WrapNoErrorCloser((*pClosers)[0])
	case 0:
		fw.c = inout.NewNoOpCloser()
	default:
		fw.c = inout.NewMultiCloser(true, true, *pClosers...)
	}
	fw.bw = inout.NewBufferedWriterSize(fw.uw, fw.opts.BufSize)
	return nil
}

// Close closes all closers used by this writer.
//
// After successfully closing this writer,
// Close will do nothing and return nil,
// and the write methods will report ErrFileWriterClosed.
// (To test whether the error is ErrFileWriterClosed, use function errors.Is.)
func (fw *writer) Close() error {
	if fw.c.Closed() {
		return nil
	}
	flushErr := fw.bw.Flush()
	closeErr := fw.c.Close()
	return errors.AutoWrap(errors.Combine(flushErr, closeErr))
}

// Closed reports whether this writer is closed successfully.
func (fw *writer) Closed() bool {
	return fw.c.Closed()
}

// Write writes the content of p into the file.
//
// It returns the number of bytes written and any write error encountered.
// If n < len(p), it also returns an error explaining why the write is short.
//
// It conforms to interface io.Writer.
func (fw *writer) Write(p []byte) (n int, err error) {
	if fw.c.Closed() {
		return 0, errors.AutoWrap(ErrFileWriterClosed)
	}
	n, err = fw.bw.Write(p)
	return n, errors.AutoWrap(err)
}

// MustWrite is like Write but panics when encountering an error.
//
// If it panics, the error value passed to the call of panic
// must be exactly of type *github.com/donyori/gogo/inout.WritePanic.
func (fw *writer) MustWrite(p []byte) (n int) {
	n, err := fw.Write(p)
	if err != nil {
		panic(inout.NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

// WriteByte writes a single byte.
//
// It returns any write error encountered.
//
// It conforms to interface io.ByteWriter.
func (fw *writer) WriteByte(c byte) error {
	if fw.c.Closed() {
		return errors.AutoWrap(ErrFileWriterClosed)
	}
	return errors.AutoWrap(fw.bw.WriteByte(c))
}

// MustWriteByte is like WriteByte but panics when encountering an error.
//
// If it panics, the error value passed to the call of panic
// must be exactly of type *github.com/donyori/gogo/inout.WritePanic.
func (fw *writer) MustWriteByte(c byte) {
	err := fw.WriteByte(c)
	if err != nil {
		panic(inout.NewWritePanic(errors.AutoWrap(err)))
	}
}

// WriteRune writes a single Unicode code point.
//
// It returns the number of bytes written and any write error encountered.
func (fw *writer) WriteRune(r rune) (size int, err error) {
	if fw.c.Closed() {
		return 0, errors.AutoWrap(ErrFileWriterClosed)
	}
	size, err = fw.bw.WriteRune(r)
	return size, errors.AutoWrap(err)
}

// MustWriteRune is like WriteRune but panics when encountering an error.
//
// If it panics, the error value passed to the call of panic
// must be exactly of type *github.com/donyori/gogo/inout.WritePanic.
func (fw *writer) MustWriteRune(r rune) (size int) {
	size, err := fw.WriteRune(r)
	if err != nil {
		panic(inout.NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

// WriteString writes a string.
//
// It returns the number of bytes written and any write error encountered.
// If n is less than len(s),
// it also returns an error explaining why the write is short.
//
// It conforms to interface io.StringWriter.
func (fw *writer) WriteString(s string) (n int, err error) {
	if fw.c.Closed() {
		return 0, errors.AutoWrap(ErrFileWriterClosed)
	}
	n, err = fw.bw.WriteString(s)
	return n, errors.AutoWrap(err)
}

// MustWriteString is like WriteString but panics when encountering an error.
//
// If it panics, the error value passed to the call of panic
// must be exactly of type *github.com/donyori/gogo/inout.WritePanic.
func (fw *writer) MustWriteString(s string) (n int) {
	n, err := fw.WriteString(s)
	if err != nil {
		panic(inout.NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

// ReadFrom reads data from r until EOF or error.
//
// The return value n is the number of bytes read.
// Any error except EOF encountered during the read is also returned.
//
// It conforms to interface io.ReaderFrom.
func (fw *writer) ReadFrom(r io.Reader) (n int64, err error) {
	if fw.c.Closed() {
		return 0, errors.AutoWrap(ErrFileWriterClosed)
	}
	n, err = fw.bw.ReadFrom(r)
	return n, errors.AutoWrap(err)
}

// Printf formats arguments and writes to the file.
// Arguments are handled in the manner of fmt.Printf.
func (fw *writer) Printf(format string, args ...any) (n int, err error) {
	if fw.c.Closed() {
		return 0, errors.AutoWrap(ErrFileWriterClosed)
	}
	n, err = fw.bw.Printf(format, args...)
	return n, errors.AutoWrap(err)
}

// MustPrintf is like Printf but panics when encountering an error.
//
// If it panics, the error value passed to the call of panic
// must be exactly of type *github.com/donyori/gogo/inout.WritePanic.
func (fw *writer) MustPrintf(format string, args ...any) (n int) {
	n, err := fw.Printf(format, args...)
	if err != nil {
		panic(inout.NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

// Print formats arguments and writes to the file.
// Arguments are handled in the manner of fmt.Print.
func (fw *writer) Print(args ...any) (n int, err error) {
	if fw.c.Closed() {
		return 0, errors.AutoWrap(ErrFileWriterClosed)
	}
	n, err = fw.bw.Print(args...)
	return n, errors.AutoWrap(err)
}

// MustPrint is like Print but panics when encountering an error.
//
// If it panics, the error value passed to the call of panic
// must be exactly of type *github.com/donyori/gogo/inout.WritePanic.
func (fw *writer) MustPrint(args ...any) (n int) {
	n, err := fw.Print(args...)
	if err != nil {
		panic(inout.NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

// Println formats arguments and writes to the file.
// Arguments are handled in the manner of fmt.Println.
func (fw *writer) Println(args ...any) (n int, err error) {
	if fw.c.Closed() {
		return 0, errors.AutoWrap(ErrFileWriterClosed)
	}
	n, err = fw.bw.Println(args...)
	return n, errors.AutoWrap(err)
}

// MustPrintln is like Println but panics when encountering an error.
//
// If it panics, the error value passed to the call of panic
// must be exactly of type *github.com/donyori/gogo/inout.WritePanic.
func (fw *writer) MustPrintln(args ...any) (n int) {
	n, err := fw.Println(args...)
	if err != nil {
		panic(inout.NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

// Flush writes any buffered data to the file.
//
// It returns any write error encountered.
func (fw *writer) Flush() error {
	if fw.c.Closed() {
		if fw.bw.Buffered() > 0 {
			// This should never happen, but will act as a safeguard for later,
			// as Flush is implicitly called in Close.
			return errors.AutoWrap(ErrFileWriterClosed)
		}
		return nil
	}
	return errors.AutoWrap(fw.bw.Flush())
}

// Size returns the size of the underlying buffer in bytes.
func (fw *writer) Size() int {
	return fw.bw.Size()
}

// Buffered returns the number of bytes that
// have been written into the current buffer.
func (fw *writer) Buffered() int {
	return fw.bw.Buffered()
}

// Available returns the number of bytes unused in the current buffer.
func (fw *writer) Available() int {
	return fw.bw.Available()
}

// TarEnabled returns true if the file is archived by tar
// (i.e., tape archive) and is not opened in raw mode.
func (fw *writer) TarEnabled() bool {
	return fw.tw != nil
}

// TarWriteHeader writes hdr and prepares to accept the content of
// the next file.
//
// The tar.Header.Size determines how many bytes can be written for
// the next file.
// If the current file is not fully written, it will return an error.
// It implicitly flushes any padding necessary before writing the header.
//
// If the file is not archived by tar, or the file is opened in raw mode,
// it does nothing and reports ErrNotTar.
// (To test whether the error is ErrNotTar, use function errors.Is.)
func (fw *writer) TarWriteHeader(hdr *tar.Header) error {
	if !fw.TarEnabled() {
		return errors.AutoWrap(ErrNotTar)
	}
	if fw.c.Closed() {
		return errors.AutoWrap(ErrFileWriterClosed)
	}
	err := fw.bw.Flush()
	if err != nil {
		return errors.AutoWrap(err)
	}
	err = errors.AutoWrap(fw.tw.WriteHeader(hdr))
	if err == nil {
		fw.bw.Reset(fw.uw) // reset the state of buffer
	}
	return err
}

// Options returns a copy of options used by this writer.
func (fw *writer) Options() *WriteOptions {
	opts := new(WriteOptions)
	*opts = fw.opts
	return opts
}

// FileStat returns the io/fs.FileInfo structure describing file.
func (fw *writer) FileStat() (info fs.FileInfo, err error) {
	return fw.f.Stat()
}
