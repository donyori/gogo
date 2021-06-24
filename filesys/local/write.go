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

package local

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/filesys"
	"github.com/donyori/gogo/inout"
)

// ErrVerificationFail is an error indicating that the file verification failed.
var ErrVerificationFail = errors.AutoNewWithStrategy("file verification failed",
	errors.PrefixFullPkgName, 0)

// WriteOptions are options for function Write.
type WriteOptions struct {
	// Only take effect when the target file already exists.
	// If true, it will append data to the file.
	// Otherwise, it will truncate the file.
	Append bool

	// True if not to compress the file with gzip and not to archive the file
	// with tar (i.e., tape archive) according to the filename.
	Raw bool

	// Size of the buffer for writing the file at least.
	// Non-positive values for using default value.
	BufSize int

	// If true, a buffer will be created when open the file. Otherwise,
	// a buffer won't be created until calling methods that need a buffer.
	BufOpen bool

	// Let the writer write data to a temporary file. After calling method
	// Close, the writer move the temporary file to the target file. If any
	// error occurs during writing, the temporary file will be discarded, and
	// the original target file won't be changed.
	Backup bool

	// Only take effect when the option Backup is disabled.
	// Let the writer preserve the written file when encountering an error.
	// If false (the default), the written file will be removed
	// by the method Close if any error occurs during writing.
	PreserveOnFail bool

	// Make parent directories before creating the file.
	MkDirs bool

	// A verify function to report whether the data written to the file is
	// correct. The function will be called in our writer's method Close.
	// If the function returns true, our writer will finish writing.
	// Otherwise, our writer will return ErrVerificationFail, and discard
	// the file if Backup is true.
	VerifyFn func() bool

	// The compression level for GNU zip. Note that the zero value (0) stands
	// for no compression other than the default value. Remember setting it to
	// gzip.DefaultCompression if you want to use the default.
	GzipLv int
}

// defaultWriteOptions are default options for function Write.
var defaultWriteOptions = &WriteOptions{
	Backup: true,
	MkDirs: true,
	GzipLv: gzip.DefaultCompression,
}

// Writer is a device to write data to a file.
//
// Its method Close closes all closable objects opened by this writer,
// including the file, and processes the temporary file used by this writer.
// After successfully closing this writer,
// its method Close will do nothing and return nil.
//
// The written file will be removed by its method Close if any error occurs
// during writing and the option PreserveOnFail is disabled.
type Writer interface {
	inout.Closer
	inout.BufferedWriter

	// TarEnabled returns true if the file is archived by tar
	// (i.e., tape archive) and is not opened in raw mode.
	TarEnabled() bool

	// TarWriteHeader writes hdr and prepares to accept the file's contents.
	//
	// The tar.Header.Size determines how many bytes can be written
	// for the next file.
	// If the current file is not fully written, it will return an error.
	// It implicitly flushes any padding necessary before writing the header.
	//
	// If the file is not archived by tar, or the file is opened in raw mode,
	// it does nothing and returns ErrNotTar.
	// (To test whether the error is ErrNotTar, use function errors.Is.)
	TarWriteHeader(hdr *tar.Header) error

	// Options returns a copy of options used by this writer.
	Options() *WriteOptions

	// Filename returns the filename as presented to function Write.
	Filename() string

	// TmpFilename returns the name of the temporary file created by
	// function Write if the option Backup enabled.
	// Otherwise, it returns an empty string.
	TmpFilename() string
}

// writer is an implementation of interface Writer.
//
// Use it with function Write.
type writer struct {
	err    error
	opts   WriteOptions
	bw     inout.ResettableBufferedWriter
	ubw    io.Writer // unbuffered writer
	tmp    string    // name of the temporary file
	name   string    // name of the destination file
	closed bool      // true if method Close has been called once and no error occurred during that call
	c      inout.Closer
	tw     *tar.Writer
}

// Write creates (if necessary) and opens a file
// with specified name for writing.
//
// If name is empty, it does nothing and returns an error.
// If opts is nil, it will use the default write options instead.
// The default write options are shown as follows:
//  Append: false,
//  Raw: false,
//  BufSize: 0,
//  BufOpen: false,
//  Backup: true,
//  MkDirs: true,
//  VerifyFn: nil,
//  GzipLv: gzip.DefaultCompression,
//
// Data ultimately written to the file will also be written to copies.
// (Due to possible compression, data written to the file may be
// different from data provided to the writer when the option Raw is disabled.)
// The client can use copies to monitor the data,
// such as calculating the checksum to verify the file.
//
// But note that:
// 1. The client is responsible for managing copies,
// including flushing or closing them after use.
// 2. If an error occurs when writing to copies,
// other writing will also stop and the writer will fall into the error state.
// 3. During one write operation, data will be written to the file first,
// and then to copies sequentially. If an error occurs when writing to copies,
// the data of current write operation has already been written to the file.
//
// As for the write options,
// notice that when options Append and Backup are both enabled
// and the specified file already exists,
// this function will copy the specified file to a temporary file,
// which may cost a lot of time and space resource.
// Data copied from the specified file won't be written to copies.
func Write(name string, perm filesys.FileMode, opts *WriteOptions, copies ...io.Writer) (w Writer, err error) {
	if name == "" {
		return nil, errors.AutoNew("name is empty")
	}
	if opts == nil {
		opts = defaultWriteOptions
	}
	fw := &writer{
		opts: *opts,
		name: name,
	}
	w = fw

	dir, base := filepath.Split(name)
	if opts.MkDirs {
		err = os.MkdirAll(dir, perm)
		if err != nil {
			return nil, errors.AutoWrap(err)
		}
	}

	el := errors.NewErrorList(true)
	defer func() {
		if el.Erroneous() {
			w, err = nil, errors.AutoWrapSkip(el.ToError(), 1) // skip = 1 to skip the inner function
		}
	}()

	var f *os.File
	if opts.Backup {
		f, err = Tmp(dir, base+".", ".tmp", perm)
		if err != nil {
			el.Append(err)
			return
		}
		fw.tmp = f.Name()
		defer func() {
			if el.Erroneous() {
				el.Append(f.Close())
				el.Append(os.Remove(fw.tmp))
			}
		}()
		if opts.Append {
			r, err1 := os.Open(name)
			if err1 == nil {
				// Don't use
				//  defer el.Append(r.Close())
				// because the argument (i.e., r.Close()) will be evaluated here
				// rather than when executing the deferred function.
				defer func() {
					el.Append(r.Close())
				}()
				_, err1 = io.Copy(f, r)
				if err1 != nil {
					el.Append(err1)
					return
				}
			} else if !errors.Is(err1, os.ErrNotExist) {
				el.Append(err1)
				return
			}
		}
	} else {
		flag := os.O_WRONLY | os.O_CREATE
		if opts.Append {
			flag |= os.O_APPEND
		} else {
			flag |= os.O_TRUNC
		}
		f, err = os.OpenFile(name, flag, perm)
		if err != nil {
			el.Append(err)
			return
		}
		defer func() {
			if el.Erroneous() {
				el.Append(f.Close())
			}
		}()
	}

	fw.ubw = f
	closers := make([]io.Closer, 1, 3)
	closers[0] = f
	defer func() {
		if el.Erroneous() {
			// Close all closers except fw.closers[0] (i.e., f),
			// which will be closed in previous defer function.
			for i := len(closers) - 1; i > 0; i-- {
				el.Append(closers[i].Close())
			}
		}
	}()

	if len(copies) > 0 {
		fw.ubw = io.MultiWriter(append([]io.Writer{fw.ubw}, copies...)...)
	}
	if !opts.Raw {
		base = strings.ToLower(base)
		ext := filepath.Ext(base)
		loop := true
		for loop {
			switch ext {
			case ".gz", ".tgz":
				gw, err1 := gzip.NewWriterLevel(fw.ubw, opts.GzipLv)
				if err1 != nil {
					el.Append(err1)
					return
				}
				closers = append(closers, gw)
				fw.ubw = gw
				if ext == ".tgz" {
					ext = ".tar"
					continue
				}
			case ".tar":
				fw.tw = tar.NewWriter(fw.ubw)
				closers = append(closers, fw.tw)
				fw.ubw = fw.tw
				loop = false
			default:
				loop = false
			}
			base = base[:len(base)-len(ext)]
			ext = filepath.Ext(base)
		}
	}

	if len(closers) > 1 {
		fw.c = inout.NewMultiCloser(true, true, closers...)
	} else {
		fw.c = inout.WrapNoErrorCloser(f)
	}
	if opts.BufOpen {
		fw.bw = inout.NewBufferedWriterSize(fw.ubw, fw.opts.BufSize)
	}
	return
}

// Close closes all closers used by this writer (including the file),
// verify the written file,
// and process the temporary file if the option Backup enabled.
//
// The written file will be removed if any error occurs during writing and
// the option PreserveOnFail is disabled.
func (fw *writer) Close() (err error) {
	if fw.closed {
		return
	}
	el := errors.NewErrorList(true) // el records the errors occurred during Close.
	var rmDone bool
	defer func() {
		if !rmDone {
			var name string
			if fw.opts.Backup {
				name = fw.tmp
			} else if !fw.opts.PreserveOnFail {
				name = fw.name
			}
			if name != "" {
				el.Append(os.Remove(name))
				err = errors.AutoWrapSkip(el.ToError(), 1) // skip = 1 to skip the inner function
			}
		}
		fw.closed = err == nil
	}()
	if fw.bw != nil {
		el.Append(fw.bw.Flush())
	}
	el.Append(fw.c.Close())
	if fw.err == nil && !el.Erroneous() &&
		fw.opts.VerifyFn != nil && !fw.opts.VerifyFn() {
		el.Append(ErrVerificationFail)
	}
	if fw.opts.Backup {
		if fw.err == nil && !el.Erroneous() {
			el.Append(os.Rename(fw.tmp, fw.name))
			if el.Erroneous() {
				el.Append(os.Remove(fw.tmp))
			}
		} else {
			el.Append(os.Remove(fw.tmp))
		}
	} else if (fw.err != nil || el.Erroneous()) && !fw.opts.PreserveOnFail {
		el.Append(os.Remove(fw.name))
	}
	err = errors.AutoWrap(el.ToError()) // only return the errors occurred during Close
	rmDone = true
	if fw.err == nil {
		if err == nil {
			fw.err = errors.AutoWrap(inout.ErrWriterClosed)
		} else {
			fw.err = err
		}
	}
	return
}

// Closed reports whether this writer is closed successfully.
func (fw *writer) Closed() bool {
	return fw.closed
}

// Write writes the contents of p into the buffer.
//
// It returns the number of bytes written and any write error encountered.
// If n < len(p), it also returns an error explaining why the write is short.
//
// It conforms to interface io.Writer.
func (fw *writer) Write(p []byte) (n int, err error) {
	if fw.err != nil {
		return 0, fw.err
	}
	var w io.Writer
	if fw.bw == nil {
		w = fw.ubw
	} else {
		w = fw.bw
	}
	n, err = w.Write(p)
	fw.err = errors.AutoWrap(err)
	return n, fw.err
}

// WriteByte writes a single byte.
//
// It returns any write error encountered.
//
// It conforms to interface io.ByteWriter.
func (fw *writer) WriteByte(c byte) error {
	if fw.err != nil {
		return fw.err
	}
	var err error
	if fw.bw != nil {
		err = fw.bw.WriteByte(c)
	} else if bw, ok := fw.ubw.(io.ByteWriter); ok {
		err = bw.WriteByte(c)
	} else {
		fw.bw = inout.NewBufferedWriterSize(fw.ubw, fw.opts.BufSize)
		err = fw.bw.WriteByte(c)
	}
	fw.err = errors.AutoWrap(err)
	return fw.err
}

// WriteString writes a string.
//
// It returns the number of bytes written and any write error encountered.
// If n is less than len(s),
// it also returns an error explaining why the write is short.
//
// It conforms to interface io.StringWriter.
func (fw *writer) WriteString(s string) (n int, err error) {
	if fw.err != nil {
		return 0, fw.err
	}
	if fw.bw != nil {
		n, err = fw.bw.WriteString(s)
	} else if sw, ok := fw.ubw.(io.StringWriter); ok {
		n, err = sw.WriteString(s)
	} else {
		n, err = fw.ubw.Write([]byte(s))
	}
	fw.err = errors.AutoWrap(err)
	return n, fw.err
}

// ReadFrom reads data from r until EOF or error.
//
// The return value n is the number of bytes read.
// Any error except EOF encountered during the read is also returned.
//
// It conforms to interface io.ReaderFrom.
func (fw *writer) ReadFrom(r io.Reader) (n int64, err error) {
	if fw.err != nil {
		return 0, fw.err
	}
	if fw.bw != nil {
		n, err = fw.bw.ReadFrom(r)
	} else if w, ok := fw.ubw.(io.ReaderFrom); ok {
		n, err = w.ReadFrom(r)
	} else if wt, ok := r.(io.WriterTo); ok {
		n, err = wt.WriteTo(fw.ubw)
	} else {
		fw.bw = inout.NewBufferedWriterSize(fw.ubw, fw.opts.BufSize)
		n, err = fw.bw.ReadFrom(r)
	}
	fw.err = errors.AutoWrap(err)
	return n, fw.err
}

// Flush writes any buffered data to the underlying writer.
//
// It returns any write error encountered.
//
// If there is no buffer, it does nothing and returns nil.
func (fw *writer) Flush() error {
	if fw.err != nil {
		return fw.err
	}
	if fw.bw == nil {
		return nil
	}
	fw.err = errors.AutoWrap(fw.bw.Flush())
	return fw.err
}

// Size returns the size of the underlying buffer in bytes.
//
// If there is no buffer, it returns 0.
func (fw *writer) Size() int {
	if fw.bw == nil {
		return 0
	}
	return fw.bw.Size()
}

// Buffered returns the number of bytes that
// have been written into the current buffer.
//
// If there is no buffer, it returns 0.
func (fw *writer) Buffered() int {
	if fw.bw == nil {
		return 0
	}
	return fw.bw.Buffered()
}

// Available returns the number of bytes unused in the current buffer.
//
// If there is no buffer, it returns 0.
func (fw *writer) Available() int {
	if fw.bw == nil {
		return 0
	}
	return fw.bw.Available()
}

// WriteRune writes a single Unicode code point.
//
// It returns the number of bytes written and any write error encountered.
func (fw *writer) WriteRune(r rune) (size int, err error) {
	if fw.err != nil {
		return 0, fw.err
	}
	if fw.bw == nil {
		fw.bw = inout.NewBufferedWriterSize(fw.ubw, fw.opts.BufSize)
	}
	size, err = fw.bw.WriteRune(r)
	fw.err = errors.AutoWrap(err)
	return size, fw.err
}

// TarEnabled returns true if the file is archived by tar
// (i.e., tape archive) and is not opened in raw mode.
func (fw *writer) TarEnabled() bool {
	return fw.tw != nil
}

// TarWriteHeader writes hdr and prepares to accept the file's contents.
//
// The tar.Header.Size determines how many bytes can be written
// for the next file.
// If the current file is not fully written, it will return an error.
// It implicitly flushes any padding necessary before writing the header.
//
// If the file is not archived by tar, or the file is opened in raw mode,
// it does nothing and returns ErrNotTar.
// (To test whether the error is ErrNotTar, use function errors.Is.)
func (fw *writer) TarWriteHeader(hdr *tar.Header) error {
	if !fw.TarEnabled() {
		return errors.AutoWrap(filesys.ErrNotTar)
	}
	if errors.Is(fw.err, inout.ErrWriterClosed) {
		return fw.err
	}
	if fw.bw != nil {
		fw.err = errors.AutoWrap(fw.bw.Flush())
		if fw.err != nil {
			return fw.err
		}
	}
	fw.err = errors.AutoWrap(fw.tw.WriteHeader(hdr))
	if fw.bw != nil {
		fw.ubw = fw.tw
		fw.bw.Reset(fw.ubw)
	}
	return fw.err
}

// Options returns a copy of options used by this writer.
func (fw *writer) Options() *WriteOptions {
	opts := new(WriteOptions)
	*opts = fw.opts
	return opts
}

// Filename returns the filename as presented to function Write.
func (fw *writer) Filename() string {
	return fw.name
}

// TmpFilename returns the name of the temporary file created by
// function Write if the option Backup enabled.
// Otherwise, it returns an empty string.
func (fw *writer) TmpFilename() string {
	return fw.tmp
}
