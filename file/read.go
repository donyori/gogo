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
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/donyori/gogo/errors"
	myio "github.com/donyori/gogo/io"
)

// An error indicating that the file is not archived by tar, or is opened in
// raw mode. Clients should use errors.Is to test whether an error is ErrNotTar.
var ErrNotTar = errors.AutoNew("file is not archived by tar, or is opened in raw mode")

// Options for function Read.
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

	// Size of the buffer for reading the file at least. 0 and negative values
	// for using default value (the default value depends on package bufio).
	BufferSize int

	// If true, a buffer will be created when open the file. Otherwise,
	// a buffer won't be created until calling methods that need a buffer.
	BufferWhenOpen bool
}

// A reader to read data from a file.
type Reader interface {
	io.Closer
	myio.BufferedReader

	// Return true if the file is archived by tar (i.e., tape archive) and is
	// not opened in raw mode.
	TarEnabled() bool

	// Advance to the next entry in the tar archive. The Header.Size determines
	// how many bytes can be read for the next file. Any remaining data in
	// current file is automatically discarded. io.EOF is returned at the end of
	// the input.
	//
	// If the file is not archived by tar (i.e., tape archive), or the file is
	// opened in raw mode, it does nothing but returns ErrNotTar.
	TarNext() (header *tar.Header, err error)

	// Return the target file, in order to retrieve information about the file.
	// Caller should NOT read, seek, or do other operations that may change
	// the state of the file to avoid breaking this reader.
	File() *os.File

	// Return a copy of option used by this reader.
	Option() *ReadOption
}

// A reader to read a file.
// Use it with function Read.
type reader struct {
	option  ReadOption
	err     error
	f       *os.File
	ubr     io.Reader // unbuffered reader
	br      myio.ResettableBufferedReader
	tr      *tar.Reader
	closers []io.Closer
}

// Open a file with given name for read. If the file is a symlink,
// it will be evaluated by filepath.EvalSymlinks. The file is opened by
// os.Open; the associated file descriptor has mode O_RDONLY.
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
	fr := &reader{
		f:       f,
		ubr:     f,
		closers: []io.Closer{f},
		option:  *option,
	}
	defer func() {
		if err != nil {
			r = nil
			fr.Close()
		}
	}()
	if option.Offset > 0 {
		_, err = f.Seek(option.Offset, io.SeekStart)
	} else if option.Offset < 0 {
		_, err = f.Seek(option.Offset, io.SeekEnd)
	}
	if err != nil {
		return nil, errors.AutoWrap(err)
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
				var gr *gzip.Reader
				gr, err = gzip.NewReader(fr.ubr)
				if err != nil {
					return nil, errors.AutoWrap(err)
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
	if option.BufferWhenOpen && fr.br == nil {
		fr.createBr()
	}
	return fr, nil
}

// Close all closers used by this reader, including the file.
func (fr *reader) Close() error {
	if fr == nil || errors.Is(fr.err, myio.ErrReaderClosed) {
		return nil
	}
	var errList errors.ErrorList
	for i := len(fr.closers) - 1; i >= 0; i-- {
		err := fr.closers[i].Close()
		errList.Append(err)
	}
	err := errors.AutoWrap(errList.ToError())
	if fr.err == nil {
		if err == nil {
			fr.err = myio.ErrReaderClosed
		} else {
			fr.err = err
		}
	}
	return err
}

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

func (fr *reader) ReadByte() (byte, error) {
	if fr.err != nil {
		return 0, fr.err
	}
	if fr.br != nil {
		n, err := fr.br.ReadByte()
		fr.err = errors.AutoWrap(err)
		return n, fr.err
	}
	if br, ok := fr.ubr.(io.ByteReader); ok {
		n, err := br.ReadByte()
		fr.err = errors.AutoWrap(err)
		return n, fr.err
	}
	fr.createBr()
	n, err := fr.br.ReadByte()
	fr.err = errors.AutoWrap(err)
	return n, fr.err
}

func (fr *reader) UnreadByte() error {
	if fr.err != nil {
		return fr.err
	}
	if fr.br != nil {
		fr.err = errors.AutoWrap(fr.br.UnreadByte())
		return fr.err
	}
	if br, ok := fr.ubr.(io.ByteScanner); ok {
		fr.err = errors.AutoWrap(br.UnreadByte())
		return fr.err
	}
	fr.createBr()
	fr.err = errors.AutoWrap(fr.br.UnreadByte())
	return fr.err
}

func (fr *reader) ReadRune() (r rune, size int, err error) {
	if fr.err != nil {
		return 0, 0, fr.err
	}
	if fr.br != nil {
		r, size, err = fr.br.ReadRune()
		fr.err = errors.AutoWrap(err)
		return r, size, fr.err
	}
	if br, ok := fr.ubr.(io.RuneReader); ok {
		r, size, err = br.ReadRune()
		fr.err = errors.AutoWrap(err)
		return r, size, fr.err
	}
	fr.createBr()
	r, size, err = fr.br.ReadRune()
	fr.err = errors.AutoWrap(err)
	return r, size, fr.err
}

func (fr *reader) UnreadRune() error {
	if fr.err != nil {
		return fr.err
	}
	if fr.br != nil {
		fr.err = errors.AutoWrap(fr.br.UnreadRune())
		return fr.err
	}
	if br, ok := fr.ubr.(io.RuneScanner); ok {
		fr.err = errors.AutoWrap(br.UnreadRune())
		return fr.err
	}
	fr.createBr()
	fr.err = errors.AutoWrap(fr.br.UnreadRune())
	return fr.err
}

func (fr *reader) WriteTo(w io.Writer) (n int64, err error) {
	if fr.err != nil {
		return 0, fr.err
	}
	if fr.br != nil {
		n, err = fr.br.WriteTo(w)
		fr.err = errors.AutoWrap(err)
		return n, fr.err
	}
	if br, ok := fr.ubr.(io.WriterTo); ok {
		n, err = br.WriteTo(w)
		fr.err = errors.AutoWrap(err)
		return n, fr.err
	}
	fr.createBr()
	n, err = fr.br.WriteTo(w)
	fr.err = errors.AutoWrap(err)
	return n, fr.err
}

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

func (fr *reader) Size() int {
	if fr.br == nil {
		if br, ok := fr.ubr.(*bufio.Reader); ok {
			return br.Size()
		}
		return 0
	}
	return fr.br.Size()
}

func (fr *reader) Buffered() int {
	if fr.br == nil {
		if br, ok := fr.ubr.(*bufio.Reader); ok {
			return br.Buffered()
		}
		return 0
	}
	return fr.br.Buffered()
}

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

func (fr *reader) TarEnabled() bool {
	return fr != nil && fr.tr != nil
}

func (fr *reader) TarNext() (header *tar.Header, err error) {
	if !fr.TarEnabled() {
		return nil, errors.AutoWrap(ErrNotTar)
	}
	if errors.Is(fr.err, myio.ErrReaderClosed) {
		return nil, fr.err
	}
	header, err = fr.tr.Next()
	fr.err = nil
	if fr.br != nil {
		fr.ubr = fr.tr
		fr.br.Reset(fr.ubr)
	}
	return header, errors.AutoWrap(err) // don't record Next error to fr.err
}

func (fr *reader) File() *os.File {
	if fr == nil {
		return nil
	}
	return fr.f
}

func (fr *reader) Option() *ReadOption {
	if fr == nil {
		return nil
	}
	option := new(ReadOption)
	*option = fr.option
	return option
}

// Wrap a BufferedReader on current reader.
// Caller should guarantee that reader.br == nil.
func (fr *reader) createBr() {
	if fr.option.BufferSize <= 0 {
		fr.br = myio.NewBufferedReader(fr.ubr)
	} else {
		fr.br = myio.NewBufferedReaderSize(fr.ubr, fr.option.BufferSize)
	}
}
