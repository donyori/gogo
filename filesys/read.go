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
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/inout"
)

// ReadOptions are options for Read functions.
type ReadOptions struct {
	// Size of the buffer for reading the file at least.
	// Non-positive values for using default value.
	BufSize int

	// Offset of the file to read, in bytes,
	// relative to the origin of the file for positive values,
	// and relative to the end of the file for negative values.
	Offset int64

	// Limit of the file to read, in bytes.
	// Non-positive values for no limit.
	Limit int64

	// True if not to decompress when the file is compressed by gzip or bzip2,
	// and not to restore when the file is archived by tar (i.e., tape archive).
	Raw bool

	// A method-decompressor map for reading the ZIP archive.
	// These decompressors are registered to the archive/zip.Reader.
	// (Nil decompressors are ignored.)
	//
	// Package archive/zip has two built-in decompressors for the common methods
	// archive/zip.Store (0) and archive/zip.Deflate (8).
	//
	// For more details, see the documentation of the method
	// RegisterDecompressor of archive/zip.Reader and
	// the function archive/zip.RegisterDecompressor.
	ZipDcomp map[uint16]zip.Decompressor

	// A function that wraps the reader to io.ReaderAt
	// to create an archive/zip.Reader.
	//
	// It also returns the size that can be read, and any error encountered.
	//
	// It will only be called when the reader does not implement io.ReaderAt.
	ZipReaderAtFunc func(r io.Reader) (ra io.ReaderAt, size int64, err error)
}

// Reader is a device to read data from a file.
//
// Its method Close closes all closable objects opened by this reader
// (may include the file).
// After successfully closing this reader,
// its method Close does nothing and returns nil,
// and its read methods report ErrFileReaderClosed.
// (To test whether the error is ErrFileReaderClosed, use function errors.Is.)
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
	// If the file is not archived by tar or is opened in raw mode,
	// it does nothing and reports ErrNotTar.
	// (To test whether err is ErrNotTar, use function errors.Is.)
	TarNext() (hdr *tar.Header, err error)

	// ZipEnabled returns true if the file is archived by ZIP
	// and is not opened in raw mode.
	ZipEnabled() bool

	// ZipOpen opens the file with specified name in the ZIP archive.
	//
	// The name must be a relative path: it must not start with a drive letter
	// (e.g., "C:") or leading slash. Only forward slashes are allowed;
	// "../" elements and trailing slashes (e.g., "a/") are not allowed.
	//
	// If the reader's file is not archived by ZIP or is opened in raw mode,
	// it does nothing and reports ErrNotZip.
	// (To test whether the error is ErrNotZip, use function errors.Is.)
	ZipOpen(name string) (file fs.File, err error)

	// ZipFiles returns the files in the ZIP archive, sorted by filename.
	//
	// If the reader's file is not archived by ZIP or is opened in raw mode,
	// it does nothing and reports ErrNotZip.
	// (To test whether the error is ErrNotZip, use function errors.Is.)
	ZipFiles() (files []*zip.File, err error)

	// ZipComment returns the end-of-central-directory comment field
	// of the ZIP archive.
	//
	// If the reader's file is not archived by ZIP or is opened in raw mode,
	// it does nothing and reports ErrNotZip.
	// (To test whether the error is ErrNotZip, use function errors.Is.)
	ZipComment() (comment string, err error)

	// Options returns a copy of options used by this reader.
	Options() *ReadOptions

	// FileStat returns the io/fs.FileInfo structure describing file.
	FileStat() (info fs.FileInfo, err error)
}

// reader is an implementation of interface Reader.
//
// Use it with Read functions.
type reader struct {
	err  error // should be one of nil, ErrIsDir, ErrReadZip, and ErrFileReaderClosed.
	br   inout.ResettableBufferedReader
	ur   io.Reader // unbuffered reader
	c    inout.Closer
	opts ReadOptions
	f    fs.File
	tr   *tar.Reader
	zr   *zip.Reader
}

// Read creates a reader on the specified file with options opts.
//
// If the file is a directory, Read reports ErrIsDir and returns a nil Reader.
// (To test whether err is ErrIsDir, use function errors.Is.)
//
// If opts are nil, a zero-value ReadOptions is used.
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
// In this case, the file is closed during the method Close of the reader,
// and it is also closed by this function when encountering an error.
//
// This function panics if file is nil.
func Read(file fs.File, opts *ReadOptions, closeFile bool) (r Reader, err error) {
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
		opts = new(ReadOptions)
	}
	size := info.Size()
	if opts.Offset != 0 &&
		(size < opts.Offset || opts.Offset < 0 && size+opts.Offset < 0) { // check whether opts.Offset < 0 to avoid overflow
		return nil, errors.AutoWrap(fmt.Errorf(
			"option Offset (%d) is out of range; file size: %d",
			opts.Offset,
			size,
		))
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
		ur: file,
		opts: ReadOptions{
			BufSize:         opts.BufSize,
			Offset:          opts.Offset,
			Limit:           opts.Limit,
			Raw:             opts.Raw,
			ZipReaderAtFunc: opts.ZipReaderAtFunc,
		},
		f: file,
	}
	for method, dcomp := range opts.ZipDcomp {
		if dcomp != nil {
			if fr.opts.ZipDcomp == nil {
				fr.opts.ZipDcomp = make(map[uint16]zip.Decompressor, len(opts.ZipDcomp))
			}
			fr.opts.ZipDcomp[method] = dcomp
		}
	}
	r = fr
	el.Append(fr.init(info, size, &closers))
	return
}

// ReadFromFS opens a file from fsys with specified name and
// options opts for reading.
//
// If the file is a directory, ReadFromFS reports ErrIsDir
// and returns a nil Reader.
// (To test whether err is ErrIsDir, use function errors.Is.)
//
// If opts are nil, a zero-value ReadOptions is used.
//
// The file is closed when the returned reader is closed.
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

// init initializes the reader according to the options.
//
// It may update closers.
func (fr *reader) init(info fs.FileInfo, size int64, pClosers *[]io.Closer) error {
	n, err := fr.initOffsetAndLimit(size)
	if err != nil {
		return err
	}
	err = fr.initRaw(info, n, pClosers)
	if err != nil {
		return err
	}
	fr.initCloserAndBuffer(*pClosers)
	return nil
}

// initOffsetAndLimit deals with the options Offset and Limit.
//
// It returns the size that can be read, and any error encountered.
func (fr *reader) initOffsetAndLimit(size int64) (n int64, err error) {
	if fr.opts.Offset == 0 && fr.opts.Limit <= 0 {
		return size, nil
	}
	offset := fr.opts.Offset
	if offset < 0 {
		offset += size
	}
	n = size - offset
	if r, ok := fr.ur.(io.ReaderAt); ok {
		// If fr.ur is an io.ReaderAt, use io.SectionReader
		// so that fr.ur is still an io.ReaderAt.
		if fr.opts.Limit > 0 && fr.opts.Limit < n {
			n = fr.opts.Limit
		}
		fr.ur = io.NewSectionReader(r, offset, n)
		return
	} else if fr.opts.Offset > 0 {
		if seeker, ok := fr.ur.(io.Seeker); ok {
			_, err = seeker.Seek(fr.opts.Offset, io.SeekStart)
		} else {
			// Discard fr.opts.Offset bytes.
			_, err = io.CopyN(io.Discard, fr.ur, fr.opts.Offset)
		}
	} else if fr.opts.Offset < 0 {
		if seeker, ok := fr.ur.(io.Seeker); ok {
			_, err = seeker.Seek(fr.opts.Offset, io.SeekEnd)
		} else {
			// Discard (size + fr.opts.Offset) bytes.
			_, err = io.CopyN(io.Discard, fr.ur, offset)
		}
	}
	if err != nil {
		return 0, err
	} else if fr.opts.Limit > 0 && fr.opts.Limit < n {
		n = fr.opts.Limit
		fr.ur = io.LimitReader(fr.ur, n)
	}
	return
}

// initRaw deals with the option Raw.
//
// size is obtained from initOffsetAndLimit, not the file size.
//
// It may update closers.
func (fr *reader) initRaw(info fs.FileInfo, size int64, pClosers *[]io.Closer) error {
	if fr.opts.Raw {
		return nil
	}
	name := strings.ToLower(info.Name())
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
		case ".zip":
			r, ok := fr.ur.(io.ReaderAt)
			var err error
			if !ok {
				if fr.opts.ZipReaderAtFunc != nil {
					r, size, err = fr.opts.ZipReaderAtFunc(fr.ur)
					if err != nil {
						return err
					}
				} else {
					return errors.New("cannot wrap to io.ReaderAt")
				}
			}
			fr.zr, err = zip.NewReader(r, size)
			if err != nil {
				return err
			}
			for method, dcomp := range fr.opts.ZipDcomp {
				fr.zr.RegisterDecompressor(method, dcomp)
			}
			fr.ur, fr.err = readZipErrorReader, ErrReadZip
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
	return nil
}

// initCloserAndBuffer sets fr.c and creates a buffer.
func (fr *reader) initCloserAndBuffer(closers []io.Closer) {
	switch len(closers) {
	case 1:
		fr.c = inout.WrapNoErrorCloser(closers[0])
	case 0:
		fr.c = inout.NewNoOpCloser()
	default:
		fr.c = inout.NewMultiCloser(true, true, closers...)
	}
	if fr.opts.BufSize <= 0 {
		fr.br = inout.NewBufferedReader(fr.ur)
	} else {
		fr.br = inout.NewBufferedReaderSize(fr.ur, fr.opts.BufSize)
	}
}

func (fr *reader) Close() error {
	if fr.c.Closed() {
		return nil
	}
	err := fr.c.Close()
	if fr.c.Closed() {
		fr.ur, fr.err = closedErrorReader, ErrFileReaderClosed
		fr.br.Reset(fr.ur)
	}
	return errors.AutoWrap(err)
}

func (fr *reader) Closed() bool {
	return fr.c.Closed()
}

func (fr *reader) Read(p []byte) (n int, err error) {
	if fr.err != nil {
		return 0, errors.AutoWrap(fr.err)
	}
	n, err = fr.br.Read(p)
	return n, errors.AutoWrap(err)
}

func (fr *reader) ReadByte() (byte, error) {
	if fr.err != nil {
		return 0, errors.AutoWrap(fr.err)
	}
	c, err := fr.br.ReadByte()
	return c, errors.AutoWrap(err)
}

func (fr *reader) UnreadByte() error {
	if fr.err != nil {
		return errors.AutoWrap(fr.err)
	}
	return errors.AutoWrap(fr.br.UnreadByte())
}

func (fr *reader) ReadRune() (r rune, size int, err error) {
	if fr.err != nil {
		return 0, 0, errors.AutoWrap(fr.err)
	}
	r, size, err = fr.br.ReadRune()
	return r, size, errors.AutoWrap(err)
}

func (fr *reader) UnreadRune() error {
	if fr.err != nil {
		return errors.AutoWrap(fr.err)
	}
	return errors.AutoWrap(fr.br.UnreadRune())
}

func (fr *reader) WriteTo(w io.Writer) (n int64, err error) {
	if fr.err != nil {
		return 0, errors.AutoWrap(fr.err)
	}
	n, err = fr.br.WriteTo(w)
	return n, errors.AutoWrap(err)
}

func (fr *reader) ReadLine() (line []byte, more bool, err error) {
	if fr.err != nil {
		return nil, false, errors.AutoWrap(fr.err)
	}
	line, more, err = fr.br.ReadLine()
	return line, more, errors.AutoWrap(err)
}

func (fr *reader) ReadEntireLine() (line []byte, err error) {
	if fr.err != nil {
		return nil, errors.AutoWrap(fr.err)
	}
	line, err = fr.br.ReadEntireLine()
	return line, errors.AutoWrap(err)
}

func (fr *reader) WriteLineTo(w io.Writer) (n int64, err error) {
	if fr.err != nil {
		return 0, errors.AutoWrap(fr.err)
	}
	n, err = fr.br.WriteLineTo(w)
	return n, errors.AutoWrap(err)
}

func (fr *reader) Size() int {
	return fr.br.Size()
}

func (fr *reader) Buffered() int {
	return fr.br.Buffered()
}

func (fr *reader) Peek(n int) (data []byte, err error) {
	if fr.err != nil {
		return nil, errors.AutoWrap(fr.err)
	}
	data, err = fr.br.Peek(n)
	return data, errors.AutoWrap(err)
}

func (fr *reader) Discard(n int) (discarded int, err error) {
	if fr.err != nil {
		return 0, errors.AutoWrap(fr.err)
	}
	discarded, err = fr.br.Discard(n)
	return discarded, errors.AutoWrap(err)
}

func (fr *reader) TarEnabled() bool {
	return fr.tr != nil
}

func (fr *reader) TarNext() (hdr *tar.Header, err error) {
	if fr.tr == nil {
		return nil, errors.AutoWrap(ErrNotTar)
	} else if fr.c.Closed() {
		return nil, errors.AutoWrap(ErrFileReaderClosed)
	}
	hdr, err = fr.tr.Next()
	switch {
	case err != nil && !errors.Is(err, io.EOF):
		return nil, errors.AutoWrap(err)
	case tarHeaderIsDir(hdr):
		fr.ur, fr.err = isDirErrorReader, ErrIsDir
	default:
		fr.ur, fr.err = fr.tr, nil
	}
	fr.br.Reset(fr.ur)
	return hdr, errors.AutoWrap(err)
}

func (fr *reader) ZipEnabled() bool {
	return fr.zr != nil
}

func (fr *reader) ZipOpen(name string) (file fs.File, err error) {
	if fr.zr == nil {
		return nil, errors.AutoWrap(ErrNotZip)
	} else if fr.c.Closed() {
		return nil, errors.AutoWrap(ErrFileReaderClosed)
	}
	file, err = fr.zr.Open(name)
	return file, errors.AutoWrap(err)
}

func (fr *reader) ZipFiles() (files []*zip.File, err error) {
	switch {
	case fr.zr == nil:
		return nil, errors.AutoWrap(ErrNotZip)
	case fr.c.Closed():
		return nil, errors.AutoWrap(ErrFileReaderClosed)
	case len(fr.zr.File) == 0:
		return
	}
	files = make([]*zip.File, len(fr.zr.File))
	copy(files, fr.zr.File)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})
	return
}

func (fr *reader) ZipComment() (comment string, err error) {
	if fr.zr == nil {
		return "", errors.AutoWrap(ErrNotZip)
	} else if fr.c.Closed() {
		return "", errors.AutoWrap(ErrFileReaderClosed)
	}
	return fr.zr.Comment, nil
}

func (fr *reader) Options() *ReadOptions {
	opts := &ReadOptions{
		BufSize:         fr.opts.BufSize,
		Offset:          fr.opts.Offset,
		Limit:           fr.opts.Limit,
		Raw:             fr.opts.Raw,
		ZipReaderAtFunc: fr.opts.ZipReaderAtFunc,
	}
	if len(fr.opts.ZipDcomp) > 0 {
		opts.ZipDcomp = make(map[uint16]zip.Decompressor, len(fr.opts.ZipDcomp))
		for method, dcomp := range fr.opts.ZipDcomp {
			opts.ZipDcomp[method] = dcomp
		}
	}
	return opts
}

func (fr *reader) FileStat() (info fs.FileInfo, err error) {
	return fr.f.Stat()
}

// tarHeaderIsDir reports whether the tar header represents a directory.
func tarHeaderIsDir(hdr *tar.Header) bool {
	return hdr != nil &&
		(hdr.Typeflag == '\x00' && len(hdr.Name) > 0 && hdr.Name[len(hdr.Name)-1] == '/' ||
			hdr.FileInfo().IsDir())
}

// errorReader implements io.Reader and io.WriterTo.
// Its methods always return 0 and report the specified error.
type errorReader struct {
	// Error reported by methods Write and ReadFrom.
	err error
}

func (er *errorReader) Read([]byte) (n int, err error) {
	return 0, errors.AutoWrap(er.err)
}

func (er *errorReader) WriteTo(io.Writer) (n int64, err error) {
	return 0, errors.AutoWrap(er.err)
}

var (
	isDirErrorReader   = &errorReader{err: ErrIsDir}
	readZipErrorReader = &errorReader{err: ErrReadZip}
	closedErrorReader  = &errorReader{err: ErrFileReaderClosed}
)
