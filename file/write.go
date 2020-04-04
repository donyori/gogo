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

package file

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/donyori/gogo/errors"
	myio "github.com/donyori/gogo/io"
)

// An error indicating that the file verification failed.
var ErrVerificationFail = errors.AutoNew("file verification failed")

// Options for function Write.
type WriteOption struct {
	// True if not to compress the file with gzip and not to archive the file
	// with tar (i.e., tape archive) according to the filename.
	Raw bool

	// Size of the buffer for writing the file at least.
	// Non-positive values for using default value.
	BufferSize int

	// If true, a buffer will be created when open the file. Otherwise,
	// a buffer won't be created until calling methods that need a buffer.
	BufferWhenOpen bool

	// Let the writer write data to a temporary file. After calling method
	// Close, the writer move the temporary file to the target file. If any
	// error occurs during writing, the temporary file will be discarded, and
	// the original target file won't be changed.
	Backup bool

	// Make parent directories before creating the file.
	MakeDirs bool

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

var defaultWriterOption = &WriteOption{
	Backup:   true,
	MakeDirs: true,
	GzipLv:   gzip.DefaultCompression,
}

// A writer to write data to a file.
type Writer interface {
	io.Closer
	myio.BufferedWriter

	// Return true if the file is archived by tar (i.e., tape archive) and is
	// not opened in raw mode.
	TarEnabled() bool

	// Write hdr and prepare to accept the file's contents. The Header.Size
	// determines how many bytes can be written for the next file. If the
	// current file is not fully written, it will return an error.
	// It implicitly flushes any padding necessary before writing the header.
	//
	// If the file is not archived by tar (i.e., tape archive), or the file is
	// opened in raw mode, it does nothing but returns ErrNotTar.
	TarWriteHeader(hdr *tar.Header) error

	// Return a copy of option used by this writer.
	Option() *WriteOption
}

// A writer to write a file.
// Use it with function New.
type writer struct {
	option   WriteOption
	filename string
	err      error
	f        *os.File
	ubw      io.Writer // unbuffered writer
	bw       myio.ResettableBufferedWriter
	tw       *tar.Writer
	closers  []io.Closer
}

// Create a file for writing, with given permission perm.
//
// Data ultimately written to the file will also be written to copies.
// The client can use copies to monitor the data, such as calculating the
// checksum to verify the file.
// Note that: 1. The writer won't manage copies. It's the client's
// responsibility to manage copies, including closing them when no longer
// needed. 2. If an error occurred when writing to copies, other writing will
// also stop and the writer will fall into the error state.
func New(name string, perm os.FileMode, option *WriteOption, copies ...io.Writer) (w Writer, err error) {
	if name == "" {
		return nil, errors.AutoNew("name is empty")
	}
	if option == nil {
		option = defaultWriterOption
	}
	fw := &writer{
		option:   *option,
		filename: name,
	}
	dir, base := filepath.Split(name)
	if option.MakeDirs {
		err = os.MkdirAll(dir, perm)
		if err != nil {
			return nil, errors.AutoWrap(err)
		}
	}
	if option.Backup {
		fw.f, err = Tmp(dir, base+".*.tmp", perm)
	} else {
		fw.f, err = os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	}
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	fw.ubw = fw.f
	fw.closers = []io.Closer{fw.f}
	defer func() {
		if err != nil {
			w = nil
			fw.Close() // ignore error
		}
	}()
	if len(copies) > 0 {
		fw.ubw = io.MultiWriter(append([]io.Writer{fw.ubw}, copies...)...)
	}
	if !option.Raw {
		base = strings.ToLower(base)
		ext := filepath.Ext(base)
		loop := true
		for loop {
			switch ext {
			case ".gz", ".tgz":
				var gw *gzip.Writer
				gw, err = gzip.NewWriterLevel(fw.ubw, option.GzipLv)
				if err != nil {
					return nil, errors.AutoWrap(err)
				}
				fw.closers = append(fw.closers, gw)
				fw.ubw = gw
				if ext == ".tgz" {
					ext = ".tar"
					continue
				}
			case ".tar":
				fw.tw = tar.NewWriter(fw.ubw)
				fw.closers = append(fw.closers, fw.tw)
				fw.ubw = fw.tw
				loop = false
			default:
				loop = false
			}
			base = base[:len(base)-len(ext)]
			ext = filepath.Ext(base)
		}
	}
	if option.BufferWhenOpen {
		fw.createBw()
	}
	return fw, nil
}

// Close all closers used by this writer, verify the written file, and
// process the temporary file if Backup enabled.
// The written file will be removed if any error occurred during writing.
func (fw *writer) Close() error {
	if fw == nil || errors.Is(fw.err, myio.ErrWriterClosed) {
		return nil
	}
	rmDone := false
	defer func() {
		if !rmDone {
			os.Remove(fw.f.Name()) // ignore error
		}
	}()
	var err error // err records the first error occurred during Close.
	if fw.bw != nil {
		err = errors.AutoWrap(fw.bw.Flush())
		if fw.err == nil {
			fw.err = err
		}
	}
	var errList errors.ErrorList
	for i := len(fw.closers) - 1; i >= 0; i-- {
		errList.Append(fw.closers[i].Close())
	}
	err1 := errors.AutoWrap(errList.ToError())
	if err == nil {
		err = err1
	}
	if fw.err == nil {
		fw.err = err
	}
	if fw.err == nil && fw.option.VerifyFn != nil && !fw.option.VerifyFn() {
		fw.err = errors.AutoWrap(ErrVerificationFail)
		if err == nil {
			err = fw.err
		}
	}
	if fw.option.Backup {
		if fw.err == nil {
			fw.err = errors.AutoWrap(os.Rename(fw.f.Name(), fw.filename))
			if err == nil {
				err = fw.err
			}
			if fw.err != nil {
				os.Remove(fw.f.Name()) // ignore error
			}
		} else {
			err1 := errors.AutoWrap(os.Remove(fw.f.Name()))
			if err == nil {
				err = err1
			}
		}
	} else if fw.err != nil {
		err1 := errors.AutoWrap(os.Remove(fw.f.Name()))
		if err == nil {
			err = err1
		}
	}
	rmDone = true
	if fw.err == nil {
		fw.err = errors.AutoWrap(myio.ErrWriterClosed)
	}
	return err // only return the error occurred during Close
}

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
		fw.createBw()
		err = fw.bw.WriteByte(c)
	}
	fw.err = errors.AutoWrap(err)
	return fw.err
}

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
		fw.createBw()
		n, err = fw.bw.ReadFrom(r)
	}
	fw.err = errors.AutoWrap(err)
	return n, fw.err
}

func (fw *writer) Flush() error {
	if fw == nil {
		return nil
	}
	if fw.err != nil {
		return fw.err
	}
	if fw.bw == nil {
		return nil
	}
	fw.err = errors.AutoWrap(fw.bw.Flush())
	return fw.err
}

func (fw *writer) Size() int {
	if fw == nil || fw.bw == nil {
		return 0
	}
	return fw.bw.Size()
}

func (fw *writer) Buffered() int {
	if fw == nil || fw.bw == nil {
		return 0
	}
	return fw.bw.Buffered()
}

func (fw *writer) Available() int {
	if fw == nil || fw.bw == nil {
		return 0
	}
	return fw.bw.Available()
}

func (fw *writer) WriteRune(r rune) (size int, err error) {
	if fw.err != nil {
		return 0, fw.err
	}
	if fw.bw == nil {
		fw.createBw()
	}
	size, err = fw.bw.WriteRune(r)
	fw.err = errors.AutoWrap(err)
	return size, fw.err
}

func (fw *writer) TarEnabled() bool {
	return fw != nil && fw.tw != nil
}

func (fw *writer) TarWriteHeader(hdr *tar.Header) error {
	if !fw.TarEnabled() {
		return errors.AutoWrap(ErrNotTar)
	}
	if errors.Is(fw.err, myio.ErrWriterClosed) {
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

func (fw *writer) Option() *WriteOption {
	if fw == nil {
		return nil
	}
	option := new(WriteOption)
	*option = fw.option
	return option
}

// Wrap a bufio.Writer on current writer.
// Caller should guarantee that fw.bw == nil.
func (fw *writer) createBw() {
	if fw.option.BufferSize <= 0 {
		fw.bw = bufio.NewWriter(fw.ubw)
	} else {
		fw.bw = bufio.NewWriterSize(fw.ubw, fw.option.BufferSize)
	}
}
