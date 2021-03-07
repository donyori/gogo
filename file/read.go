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

package file

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/donyori/gogo/errors"
	myio "github.com/donyori/gogo/io"
)

// ErrNotTar is an error indicating that the file is not archived by tar,
// or is opened in raw mode.
//
// The client should use errors.Is to test whether an error is ErrNotTar.
var ErrNotTar = errors.AutoNew("file is not archived by tar, or is opened in raw mode")

// ReadOption is a set of options for function Read.
type ReadOption struct {
	// Offset of the file to read, in bytes,
	// relative to the origin of the file for positive values,
	// and relative to the end of the file for negative values.
	Offset int64

	// Limit of the file to read, in bytes. Non-positive values for no limit.
	Limit int64

	// True if not to decompress when the file is compressed by gzip or bzip2,
	// and not to restore when the file is archived by tar (i.e., tape archive).
	Raw bool

	// Size of the buffer for reading the file at least.
	// Non-positive values for using default value.
	BufferSize int

	// If true, a buffer will be created when open the file. Otherwise,
	// a buffer won't be created until calling methods that need a buffer.
	BufferWhenOpen bool
}

// Reader is a device to read data from a file.
type Reader interface {
	io.Closer
	myio.BufferedReader

	// TarEnabled returns true if the file is archived by tar
	// (i.e., tape archive) and is not opened in raw mode.
	TarEnabled() bool

	// TarNext advances to the next entry in the tar archive.
	//
	// The tar.Header.Size determines how many bytes can be read
	// for the next file.
	// Any remaining data in current file is automatically discarded.
	//
	// io.EOF is returned at the end of the input.
	//
	// If the file is not archived by tar, or the file is opened in raw mode,
	// it does nothing and returns ErrNotTar.
	// (To test whether the error is ErrNotTar, use function errors.Is.)
	TarNext() (header *tar.Header, err error)

	// Option returns a copy of options used by this reader.
	Option() *ReadOption

	// Filename returns the filename as presented to function Read.
	Filename() string

	// FileInfo returns the information of the file.
	FileInfo() (info os.FileInfo, err error)
}

// reader is an implementation of interface Reader.
//
// Use it with function Read.
type reader struct {
	option  ReadOption
	err     error
	f       *os.File
	ubr     io.Reader // unbuffered reader
	br      myio.ResettableBufferedReader
	tr      *tar.Reader
	closers []io.Closer
	closed  bool // true if method Close has been called once and no error occurred during that call
}

// Read opens a file with specified name for reading.
//
// If the file is a symlink, it will be evaluated by filepath.EvalSymlinks.
//
// The file is opened by os.Open;
// the associated file descriptor has mode syscall.O_RDONLY.
func Read(name string, option *ReadOption) (r Reader, err error) {
	name, err = filepath.EvalSymlinks(name)
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	f, err := os.Open(name)
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	if option == nil {
		option = new(ReadOption)
	}
	el := errors.NewErrorList(true)
	defer func() {
		if el.Erroneous() {
			r, err = nil, errors.AutoWrap(el.ToError())
		}
	}()
	fr := &reader{
		option:  *option,
		f:       f,
		ubr:     f,
		closers: []io.Closer{f},
	}
	r = fr
	defer func() {
		if el.Erroneous() {
			for i := len(fr.closers) - 1; i >= 0; i-- {
				el.Append(fr.closers[i].Close())
			}
		}
	}()
	if option.Offset > 0 {
		_, err = f.Seek(option.Offset, io.SeekStart)
	} else if option.Offset < 0 {
		_, err = f.Seek(option.Offset, io.SeekEnd)
	}
	if err != nil {
		el.Append(err)
		return
	}
	if option.Limit > 0 {
		fr.ubr = io.LimitReader(fr.ubr, option.Limit)
	}
	if !option.Raw {
		base := strings.ToLower(filepath.Base(name))
		ext := filepath.Ext(base)
		loop := true
		for loop {
			switch ext {
			case ".gz", ".tgz":
				gr, err1 := gzip.NewReader(fr.ubr)
				if err1 != nil {
					el.Append(err1)
					return
				}
				fr.closers = append(fr.closers, gr)
				fr.ubr = gr
				if ext == ".tgz" {
					ext = ".tar"
					continue
				}
			case ".bz2":
				fr.ubr = bzip2.NewReader(fr.ubr)
			case ".tar":
				fr.tr = tar.NewReader(fr.ubr)
				fr.ubr = fr.tr
				loop = false
			default:
				loop = false
			}
			base = base[:len(base)-len(ext)]
			ext = filepath.Ext(base)
		}
	}
	if option.BufferWhenOpen {
		fr.createBr()
	}
	return
}

// Close closes all closers used by this reader, including the file.
func (fr *reader) Close() error {
	if fr.closed {
		return nil
	}
	errList := errors.NewErrorList(true) // errList records the errors occurred during Close.
	for i := len(fr.closers) - 1; i >= 0; i-- {
		errList.Append(fr.closers[i].Close())
	}
	err := errors.AutoWrap(errList.ToError())
	fr.closed = err == nil
	if fr.err == nil {
		if fr.closed {
			fr.err = errors.AutoWrap(myio.ErrReaderClosed)
		} else {
			fr.err = err
		}
	}
	return err
}

// Read reads data into p.
//
// It conforms to interface io.Reader.
func (fr *reader) Read(p []byte) (n int, err error) {
	if fr.err != nil {
		return 0, fr.err
	}
	var r io.Reader
	if fr.br != nil {
		r = fr.br
	} else {
		r = fr.ubr
	}
	n, err = r.Read(p)
	fr.err = errors.AutoWrap(err)
	return n, fr.err
}

// ReadByte reads and returns a single byte.
//
// It conforms to interface io.ByteReader.
func (fr *reader) ReadByte() (byte, error) {
	if fr.err != nil {
		return 0, fr.err
	}
	var c byte
	var err error
	if fr.br != nil {
		c, err = fr.br.ReadByte()
	} else if br, ok := fr.ubr.(io.ByteReader); ok {
		c, err = br.ReadByte()
	} else {
		fr.createBr()
		c, err = fr.br.ReadByte()
	}
	fr.err = errors.AutoWrap(err)
	return c, fr.err
}

// UnreadByte unreads the last byte.
// Only the most recently read byte can be unread.
//
// It conforms to interface io.ByteScanner.
func (fr *reader) UnreadByte() error {
	if fr.err != nil {
		return fr.err
	}
	if fr.br != nil {
		fr.err = errors.AutoWrap(fr.br.UnreadByte())
	} else if bs, ok := fr.ubr.(io.ByteScanner); ok {
		fr.err = errors.AutoWrap(bs.UnreadByte())
	} else {
		fr.createBr()
		fr.err = errors.AutoWrap(fr.br.UnreadByte())
	}
	return fr.err
}

// ReadRune reads a single UTF-8 encoded Unicode character and
// returns the rune and its size in bytes.
//
// It conforms to interface io.RuneReader.
func (fr *reader) ReadRune() (r rune, size int, err error) {
	if fr.err != nil {
		return 0, 0, fr.err
	}
	if fr.br != nil {
		r, size, err = fr.br.ReadRune()
	} else if rr, ok := fr.ubr.(io.RuneReader); ok {
		r, size, err = rr.ReadRune()
	} else {
		fr.createBr()
		r, size, err = fr.br.ReadRune()
	}
	fr.err = errors.AutoWrap(err)
	return r, size, fr.err
}

// UnreadRune unreads the last rune.
//
// It conforms to interface io.RuneScanner.
func (fr *reader) UnreadRune() error {
	if fr.err != nil {
		return fr.err
	}
	if fr.br != nil {
		fr.err = errors.AutoWrap(fr.br.UnreadRune())
	} else if rs, ok := fr.ubr.(io.RuneScanner); ok {
		fr.err = errors.AutoWrap(rs.UnreadRune())
	} else {
		fr.createBr()
		fr.err = errors.AutoWrap(fr.br.UnreadRune())
	}
	return fr.err
}

// WriteTo writes data to w until there's no more data to write or
// when an error occurs.
//
// The return value n is the number of bytes written.
// Any error encountered during the write is also returned.
//
// It conforms to interface io.WriterTo.
func (fr *reader) WriteTo(w io.Writer) (n int64, err error) {
	if fr.err != nil {
		return 0, fr.err
	}
	if fr.br != nil {
		n, err = fr.br.WriteTo(w)
	} else if r, ok := fr.ubr.(io.WriterTo); ok {
		n, err = r.WriteTo(w)
	} else if rf, ok := w.(io.ReaderFrom); ok {
		n, err = rf.ReadFrom(fr.ubr)
	} else {
		fr.createBr()
		n, err = fr.br.WriteTo(w)
	}
	fr.err = errors.AutoWrap(err)
	return n, fr.err
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
func (fr *reader) ReadLine() (line []byte, more bool, err error) {
	if fr.err != nil {
		return nil, false, fr.err
	}
	if fr.br == nil {
		fr.createBr()
	}
	line, more, err = fr.br.ReadLine()
	fr.err = errors.AutoWrap(err)
	return line, more, fr.err
}

// WriteLineTo reads a line from its underlying reader and writes it to w.
//
// It stops writing data if an error occurs.
//
// It returns the number of bytes written to w and any error encountered.
func (fr *reader) WriteLineTo(w io.Writer) (n int64, err error) {
	if fr.err != nil {
		return 0, fr.err
	}
	if fr.br == nil {
		fr.createBr()
	}
	n, err = fr.br.WriteLineTo(w)
	fr.err = errors.AutoWrap(err)
	return n, fr.err
}

// Size returns the size of the underlying buffer in bytes.
//
// If there is no buffer, it returns 0.
func (fr *reader) Size() int {
	if fr.br == nil {
		return 0
	}
	return fr.br.Size()
}

// Buffered returns the number of bytes
// that can be read from the current buffer.
//
// If there is no buffer, it returns 0.
func (fr *reader) Buffered() int {
	if fr.br == nil {
		return 0
	}
	return fr.br.Buffered()
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
func (fr *reader) Peek(n int) (data []byte, err error) {
	if fr.err != nil {
		return nil, fr.err
	}
	if fr.br == nil {
		fr.createBr()
	}
	data, err = fr.br.Peek(n)
	fr.err = errors.AutoWrap(err)
	return data, fr.err
}

// Discard skips the next n bytes and returns the number of bytes discarded.
//
// If it skips fewer than n bytes, it also returns an error explaining why.
//
// If 0 <= n <= Buffered(),
// it is guaranteed to succeed without reading from the underlying reader.
func (fr *reader) Discard(n int) (discarded int, err error) {
	if fr.err != nil {
		return 0, fr.err
	}
	if fr.br == nil {
		fr.createBr()
	}
	discarded, err = fr.br.Discard(n)
	fr.err = errors.AutoWrap(err)
	return discarded, fr.err
}

// TarEnabled returns true if the file is archived by tar
// (i.e., tape archive) and is not opened in raw mode.
func (fr *reader) TarEnabled() bool {
	return fr.tr != nil
}

// TarNext advances to the next entry in the tar archive.
//
// The tar.Header.Size determines how many bytes can be read
// for the next file.
// Any remaining data in current file is automatically discarded.
//
// io.EOF is returned at the end of the input.
//
// If the file is not archived by tar, or the file is opened in raw mode,
// it does nothing and returns ErrNotTar.
// (To test whether the error is ErrNotTar, use function errors.Is.)
func (fr *reader) TarNext() (header *tar.Header, err error) {
	if !fr.TarEnabled() {
		return nil, errors.AutoWrap(ErrNotTar)
	}
	if errors.Is(fr.err, myio.ErrReaderClosed) {
		return nil, fr.err
	}
	header, err = fr.tr.Next()
	err = errors.AutoWrap(err)
	fr.err = err
	if errors.Is(err, io.EOF) {
		fr.err = nil // don't record io.EOF to enable following reading
	}
	if fr.br != nil {
		fr.ubr = fr.tr
		fr.br.Reset(fr.ubr)
	}
	return header, err
}

// Option returns a copy of options used by this reader.
func (fr *reader) Option() *ReadOption {
	option := new(ReadOption)
	*option = fr.option
	return option
}

// Filename returns the filename as presented to function Read.
func (fr *reader) Filename() string {
	return fr.f.Name()
}

// FileInfo returns the information of the file.
func (fr *reader) FileInfo() (info os.FileInfo, err error) {
	return fr.f.Stat()
}

// createBr wraps a BufferedReader on current reader.
//
// Caller should guarantee that fr.br == nil.
func (fr *reader) createBr() {
	if fr.option.BufferSize <= 0 {
		fr.br = myio.NewBufferedReader(fr.ubr)
	} else {
		fr.br = myio.NewBufferedReaderSize(fr.ubr, fr.option.BufferSize)
	}
}
