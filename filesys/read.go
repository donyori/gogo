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
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/inout"
)

// ErrNotTar is an error indicating that the file is not archived by tar,
// or is opened in raw mode.
//
// The client should use errors.Is to test whether an error is ErrNotTar.
var ErrNotTar = errors.AutoNewCustom("file is not archived by tar, or is opened in raw mode",
	errors.PrependFullPkgName, 0)

// ReadOptions are options for Read functions.
type ReadOptions struct {
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
	BufSize int
}

// Reader is a device to read data from a file.
//
// Its method Close closes all closable objects opened by this reader
// (may include the file).
// After successfully closing this reader,
// its method Close will do nothing and return nil.
type Reader interface {
	inout.Closer
	inout.BufferedReader

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
	// it does nothing and reports ErrNotTar.
	// (To test whether err is ErrNotTar, use function errors.Is.)
	TarNext() (header *tar.Header, err error)

	// Options returns a copy of options used by this reader.
	Options() *ReadOptions

	// FileInfo returns the information of the file.
	FileInfo() (info fs.FileInfo, err error)
}

// reader is an implementation of interface Reader.
//
// Use it with Read functions.
type reader struct {
	err  error
	br   inout.ResettableBufferedReader
	ur   io.Reader // unbuffered reader
	opts ReadOptions
	c    inout.Closer
	f    fs.File
	tr   *tar.Reader
}

// Read creates a reader on the specified file with options opts.
//
// If opts are nil, a zero-value ReadOptions will be used.
//
// To ensure that this function and the returned reader can work as expected,
// the input file must not be operated by anyone else
// before closing the returned reader.
// If the option Offset is non-zero and the file is not an io.Seeker,
// the file must be ready to be read from the beginning.
//
// closeFile indicates whether the reader should close the file
// when calling its method Close.
// If closeFile is false, the client is responsible for closing file
// after closing the reader.
// If closeFile is true, the client should not close the file,
// even if this function reports an error.
// In this case, the file will be closed during the method Close of the reader,
// and it will also be closed by this function when encountering an error.
//
// This function panics if file is nil.
func Read(file fs.File, opts *ReadOptions, closeFile bool) (r Reader, err error) {
	if file == nil {
		panic(errors.AutoMsg("file is nil"))
	}
	if opts == nil {
		opts = new(ReadOptions)
	}
	info, err := file.Stat()
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	if opts.Offset != 0 {
		if size := info.Size(); size < opts.Offset || size+opts.Offset < 0 {
			return nil, errors.AutoNew(fmt.Sprintf("option Offset is out of range, file size: %d, Offset: %d", size, opts.Offset))
		}
	}
	el := errors.NewErrorList(true)
	defer func() {
		if el.Erroneous() {
			r, err = nil, errors.AutoWrapSkip(el.ToError(), 1) // skip = 1 to skip the inner function
		}
	}()
	closers := make([]io.Closer, 0, 2)
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
	fr := &reader{
		ur:   file,
		opts: *opts,
		f:    file,
	}
	r = fr
	err = readSubOffsetAndLimit(fr, info)
	if err != nil {
		el.Append(err)
		return
	}
	err = readSubRawClosersAndCreateBuffer(fr, info, &closers)
	if err != nil {
		el.Append(err)
	}
	return
}

// readSubOffsetAndLimit is a sub-process of function Read
// to deal with the options Offset and Limit.
func readSubOffsetAndLimit(fr *reader, info fs.FileInfo) error {
	var err error
	if fr.opts.Offset > 0 {
		if seeker, ok := fr.f.(io.Seeker); ok {
			_, err = seeker.Seek(fr.opts.Offset, io.SeekStart)
		} else {
			// Discard fr.opts.Offset bytes.
			_, err = io.CopyN(io.Discard, fr.f, fr.opts.Offset)
		}
	} else if fr.opts.Offset < 0 {
		if seeker, ok := fr.f.(io.Seeker); ok {
			_, err = seeker.Seek(fr.opts.Offset, io.SeekEnd)
		} else {
			// Discard (size + fr.opts.Offset) bytes.
			_, err = io.CopyN(io.Discard, fr.f, info.Size()+fr.opts.Offset)
		}
	}
	if err != nil {
		return err
	}
	if fr.opts.Limit > 0 {
		fr.ur = io.LimitReader(fr.ur, fr.opts.Limit)
	}
	return nil
}

// readSubRawClosersAndCreateBuffer is a sub-process of function Read
// to deal with the options Raw, set fr.c, update closers, and create a buffer.
func readSubRawClosersAndCreateBuffer(fr *reader, info fs.FileInfo, pClosers *[]io.Closer) error {
	if !fr.opts.Raw {
		name := strings.ToLower(path.Clean(filepath.ToSlash(info.Name())))
		ext := path.Ext(name)
		for {
			var endLoop bool
			switch ext {
			case ".gz", ".tgz":
				gr, err := gzip.NewReader(fr.ur)
				if err != nil {
					return err
				}
				*pClosers = append(*pClosers, gr)
				fr.ur = gr
				if ext == ".tgz" {
					ext = ".tar"
					continue
				}
			case ".bz2", ".tbz":
				fr.ur = bzip2.NewReader(fr.ur)
				if ext == ".tbz" {
					ext = ".tar"
					continue
				}
			case ".tar":
				fr.tr = tar.NewReader(fr.ur)
				fr.ur = fr.tr
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
		fr.c = inout.WrapNoErrorCloser((*pClosers)[0])
	case 0:
		fr.c = inout.NewNoOpCloser()
	default:
		fr.c = inout.NewMultiCloser(true, true, *pClosers...)
	}
	if fr.opts.BufSize <= 0 {
		fr.br = inout.NewBufferedReader(fr.ur)
	} else {
		fr.br = inout.NewBufferedReaderSize(fr.ur, fr.opts.BufSize)
	}
	return nil
}

// ReadFromFS opens a file from fsys with specified name and
// options opts for reading.
//
// If opts are nil, a zero-value ReadOptions will be used.
//
// The file will be closed when the returned reader is closed.
//
// This function panics if fsys is nil.
func ReadFromFS(fsys fs.FS, name string, opts *ReadOptions) (r Reader, err error) {
	if fsys == nil {
		panic(errors.AutoMsg("fsys is nil"))
	}
	f, err := fsys.Open(name)
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	r, err = Read(f, opts, true)
	return r, errors.AutoWrap(err)
}

// Close closes all closers used by this reader.
func (fr *reader) Close() error {
	if fr.c.Closed() {
		return nil
	}
	err := errors.AutoWrap(fr.c.Close())
	if fr.err == nil {
		if fr.c.Closed() {
			fr.err = errors.AutoWrap(inout.ErrReaderClosed)
		} else {
			fr.err = err
		}
	}
	return err
}

// Closed reports whether this reader is closed successfully.
func (fr *reader) Closed() bool {
	return fr.c.Closed()
}

// Read reads data into p.
//
// It conforms to interface io.Reader.
func (fr *reader) Read(p []byte) (n int, err error) {
	if fr.err != nil {
		return 0, fr.err
	}
	n, err = fr.br.Read(p)
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
	c, err := fr.br.ReadByte()
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
	fr.err = errors.AutoWrap(fr.br.UnreadByte())
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
	r, size, err = fr.br.ReadRune()
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
	fr.err = errors.AutoWrap(fr.br.UnreadRune())
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
	n, err = fr.br.WriteTo(w)
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
// If an error (including io.EOF) occurs after reading some content,
// it returns the content as a line and a nil error.
// The error encountered will be reported on future read calls.
//
// No indication or error is given if the input ends
// without a final line end.
// Even if the input ends without end-of-line bytes,
// the content before EOF is treated as a line.
//
// Caller should not keep the return value line,
// and line is only valid until the next call to the reader,
// including the method ReadLine and any other possible methods.
func (fr *reader) ReadLine() (line []byte, more bool, err error) {
	if fr.err != nil {
		return nil, false, fr.err
	}
	line, more, err = fr.br.ReadLine()
	fr.err = errors.AutoWrap(err)
	return line, more, fr.err
}

// ReadEntireLine reads an entire line excluding the end-of-line bytes.
//
// It either returns a non-nil line or it returns an error, never both.
// If an error (including io.EOF) occurs after reading some content,
// it returns the content as a line and a nil error.
// The error encountered will be reported on future read calls.
//
// No indication or error is given if the input ends
// without a final line end.
// Even if the input ends without end-of-line bytes,
// the content before EOF is treated as a line.
//
// Unlike the method ReadLine of interface LineReader,
// the returned line is always valid.
// Caller can keep the returned line safely.
//
// If the line is too long to be stored in a []byte
// (hardly happens in text files), it may panic or report an error.
func (fr *reader) ReadEntireLine() (line []byte, err error) {
	if fr.err != nil {
		return nil, fr.err
	}
	line, err = fr.br.ReadEntireLine()
	fr.err = errors.AutoWrap(err)
	return line, fr.err
}

// WriteLineTo reads a line excluding the end-of-line bytes
// from its underlying reader and writes it to w.
//
// It stops writing data if an error occurs.
//
// It returns the number of bytes written to w and any error encountered.
//
// If an error (including io.EOF) occurs while reading from
// the underlying reader, but some content has already been read,
// it writes the content as a line and returns a nil error.
// The error encountered will be reported on future read calls.
//
// No indication or error is given if the input ends
// without a final line end.
// Even if the input ends without end-of-line bytes,
// the content before EOF is treated as a line.
func (fr *reader) WriteLineTo(w io.Writer) (n int64, err error) {
	if fr.err != nil {
		return 0, fr.err
	}
	n, err = fr.br.WriteLineTo(w)
	fr.err = errors.AutoWrap(err)
	return n, fr.err
}

// Size returns the size of the underlying buffer in bytes.
func (fr *reader) Size() int {
	return fr.br.Size()
}

// Buffered returns the number of bytes
// that can be read from the current buffer.
func (fr *reader) Buffered() int {
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
// Calling Peek prevents an UnreadByte or UnreadRune call from succeeding
// until the next read operation.
func (fr *reader) Peek(n int) (data []byte, err error) {
	if fr.err != nil {
		return nil, fr.err
	}
	data, err = fr.br.Peek(n)
	err = errors.AutoWrap(err)
	if !errors.Is(err, bufio.ErrBufferFull) { // don't record bufio.ErrBufferFull
		fr.err = err
	}
	return
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
// it does nothing and reports ErrNotTar.
// (To test whether err is ErrNotTar, use function errors.Is.)
func (fr *reader) TarNext() (header *tar.Header, err error) {
	if !fr.TarEnabled() {
		return nil, errors.AutoWrap(ErrNotTar)
	}
	if errors.Is(fr.err, inout.ErrReaderClosed) {
		return nil, fr.err
	}
	header, err = fr.tr.Next()
	err = errors.AutoWrap(err)
	fr.err = err
	if errors.Is(err, io.EOF) {
		fr.err = nil // don't record io.EOF to enable following reading
	}
	fr.br.Reset(fr.ur) // discard current buffered data
	return header, err
}

// Options returns a copy of options used by this reader.
func (fr *reader) Options() *ReadOptions {
	opts := new(ReadOptions)
	*opts = fr.opts
	return opts
}

// FileInfo returns the information of the file.
func (fr *reader) FileInfo() (info fs.FileInfo, err error) {
	return fr.f.Stat()
}
