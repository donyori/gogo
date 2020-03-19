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
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/donyori/gogo/errors"
	myio "github.com/donyori/gogo/io"
)

func init() {
	errors.AutoWrapExclude(io.EOF) // don't wrap io.EOF to avoid old codes going wrong
}

// Options for function ReadFile.
type ReadOption struct {
	// Offset of the file to read, in bytes,
	// relative to the origin of the file for positive value,
	// and relative to the end of the file for negative value.
	Offset int64

	// Limit of the file to read, in bytes. Non-positive values for no limit.
	Limit int64

	// True if not required to decompress when the file is compressed by gzip or bzip2.
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

	// Return the target file, in order to retrieve information about the file.
	// Caller should NOT read, seek, or do other operations that may change
	// the state of the file to avoid breaking this reader.
	File() *os.File

	// Return a copy of option used by this reader.
	Option() *ReadOption
}

// A reader to read a file.
// Use it with function ReadFile.
type reader struct {
	option  ReadOption
	f       *os.File
	r       io.Reader
	closers []io.Closer
	br      myio.BufferedReader
}

// Open a file with given name for read. If the file is a symlink,
// it will be evaluated by filepath.EvalSymlinks. The file is opened by
// os.Open; the associated file descriptor has mode O_RDONLY.
func ReadFile(name string, option *ReadOption) (r Reader, err error) {
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
		r:       f,
		closers: []io.Closer{f},
		option:  *option,
	}
	if option.Offset > 0 {
		_, err = f.Seek(option.Offset, io.SeekStart)
	} else if option.Offset < 0 {
		_, err = f.Seek(option.Offset, io.SeekEnd)
	}
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	if option.Limit > 0 {
		fr.r = io.LimitReader(fr.r, option.Limit)
	}
	if !option.Raw {
		switch filepath.Ext(name) {
		case ".gz":
			gr, err := gzip.NewReader(fr.r)
			if err != nil {
				return nil, errors.AutoWrap(err)
			}
			fr.closers = append(fr.closers, gr)
			fr.r = gr
		case ".bz2":
			fr.r = bzip2.NewReader(fr.r)
		}
	}
	fr.br, _ = fr.r.(myio.BufferedReader) // enable fr.br if fr.fr is BufferedReader
	if fr.br != nil && fr.br.Size() < option.BufferSize {
		// Make sure that the buffer is large enough.
		fr.br = myio.NewBufferedReaderSize(fr.r, option.BufferSize)
		fr.r = fr.br
	}
	if option.BufferWhenOpen && fr.br == nil {
		fr.createBr()
	}
	return fr, nil
}

// Close all closers used by this reader, including the file.
func (fr *reader) Close() error {
	var errList errors.ErrorList
	for i := len(fr.closers) - 1; i >= 0; i-- {
		err := fr.closers[i].Close()
		errList.Append(err)
	}
	return errors.AutoWrap(errList.ToError())
}

func (fr *reader) Read(p []byte) (n int, err error) {
	n, err = fr.r.Read(p)
	return n, errors.AutoWrap(err)
}

func (fr *reader) ReadByte() (byte, error) {
	if fr.br != nil {
		n, err := fr.br.ReadByte()
		return n, errors.AutoWrap(err)
	}
	if br, ok := fr.r.(io.ByteReader); ok {
		n, err := br.ReadByte()
		return n, errors.AutoWrap(err)
	}
	fr.createBr()
	n, err := fr.br.ReadByte()
	return n, errors.AutoWrap(err)
}

func (fr *reader) UnreadByte() error {
	if fr.br != nil {
		return errors.AutoWrap(fr.br.UnreadByte())
	}
	if br, ok := fr.r.(io.ByteScanner); ok {
		return errors.AutoWrap(br.UnreadByte())
	}
	fr.createBr()
	return errors.AutoWrap(fr.br.UnreadByte())
}

func (fr *reader) ReadRune() (r rune, size int, err error) {
	if fr.br != nil {
		r, size, err = fr.br.ReadRune()
		return r, size, errors.AutoWrap(err)
	}
	if br, ok := fr.r.(io.RuneReader); ok {
		r, size, err = br.ReadRune()
		return r, size, errors.AutoWrap(err)
	}
	fr.createBr()
	r, size, err = fr.br.ReadRune()
	return r, size, errors.AutoWrap(err)
}

func (fr *reader) UnreadRune() error {
	if fr.br != nil {
		return errors.AutoWrap(fr.br.UnreadRune())
	}
	if br, ok := fr.r.(io.RuneScanner); ok {
		return errors.AutoWrap(br.UnreadRune())
	}
	fr.createBr()
	return errors.AutoWrap(fr.br.UnreadRune())
}

func (fr *reader) WriteTo(w io.Writer) (n int64, err error) {
	if fr.br != nil {
		n, err = fr.br.WriteTo(w)
		return n, errors.AutoWrap(err)
	}
	if br, ok := fr.r.(io.WriterTo); ok {
		n, err = br.WriteTo(w)
		return n, errors.AutoWrap(err)
	}
	fr.createBr()
	n, err = fr.br.WriteTo(w)
	return n, errors.AutoWrap(err)
}

func (fr *reader) ReadLine() (line []byte, more bool, err error) {
	if fr.br == nil {
		fr.createBr()
	}
	line, more, err = fr.br.ReadLine()
	return line, more, errors.AutoWrap(err)
}

func (fr *reader) WriteLineTo(w io.Writer) (n int64, err error) {
	if fr.br == nil {
		fr.createBr()
	}
	n, err = fr.br.WriteLineTo(w)
	return n, errors.AutoWrap(err)
}

func (fr *reader) Size() int {
	if fr.br == nil {
		return 0
	}
	return fr.br.Size()
}

func (fr *reader) Buffered() int {
	if fr.br == nil {
		return 0
	}
	return fr.br.Buffered()
}

func (fr *reader) Peek(n int) (data []byte, err error) {
	if fr.br == nil {
		fr.createBr()
	}
	data, err = fr.br.Peek(n)
	return data, errors.AutoWrap(err)
}

func (fr *reader) Discard(n int) (discarded int, err error) {
	if fr.br == nil {
		fr.createBr()
	}
	discarded, err = fr.br.Discard(n)
	return discarded, errors.AutoWrap(err)
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
		fr.br = myio.NewBufferedReader(fr.r)
	} else {
		fr.br = myio.NewBufferedReaderSize(fr.r, fr.option.BufferSize)
	}
	fr.r = fr.br
}
