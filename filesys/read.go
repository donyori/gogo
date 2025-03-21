// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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
	"iter"
	"maps"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/inout"
)

// ReadOptions are options for Read functions.
type ReadOptions struct {
	// BufSize is the size of the buffer for reading the file at least.
	//
	// Nonpositive values for using default value.
	BufSize int

	// Offset is the offset of the file to read, in bytes,
	// relative to the origin of the file for positive values,
	// and relative to the end of the file for negative values.
	Offset int64

	// Limit is the limit of the file to read, in bytes.
	//
	// Nonpositive values for no limit.
	Limit int64

	// Raw indicates whether to not decompress
	// when the file is compressed by gzip or bzip2,
	// and whether to not restore
	// when the file is archived by tar (i.e., tape archive).
	Raw bool

	// ZipDcomp is a method-decompressor map for reading the ZIP archive.
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

	// ZipReaderAtFunc is a function that wraps the reader to io.ReaderAt
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
	// If TarNext encounters a non-local name (as defined by filepath.IsLocal),
	// TarNext returns the header with an tar.ErrInsecurePath error.
	// (To test whether err is tar.ErrInsecurePath, use function errors.Is.)
	// The client can ignore the tar.ErrInsecurePath error
	// and use the returned header if it wants to accept non-local names.
	//
	// If the file is not archived by tar or is opened in raw mode,
	// TarNext does nothing and reports ErrNotTar.
	// (To test whether err is ErrNotTar, use function errors.Is.)
	TarNext() (hdr *tar.Header, err error)

	// IterTarFiles returns a single-use iterator
	// over tar headers in the tar archive.
	// The reader is automatically switched to
	// the corresponding tar file content along with the iterator.
	//
	// The iteration early stops when an error occurs.
	// If pErr is not nil and the error is not io.EOF,
	// the error is output to *pErr.
	// Otherwise, the error may be unretrievable.
	// If pErr is not nil and there is no error except for io.EOF,
	// *pErr is set to nil after iteration.
	//
	// acceptNonLocalNames indicates whether to accept non-local names
	// (as defined by filepath.IsLocal).
	// If it is true, tar.ErrInsecurePath is ignored during the iteration.
	//
	// If the file is not archived by tar or is opened in raw mode,
	// it does nothing and reports ErrNotTar.
	// (To test whether err is ErrNotTar, use function errors.Is.)
	//
	// The returned iterator is always non-nil
	// but is a no-op iterator if err is not nil.
	IterTarFiles(pErr *error, acceptNonLocalNames bool) (
		seq iter.Seq[*tar.Header], err error)

	// IterIndexTarFiles returns a single-use iterator
	// over index-header pairs in the tar archive.
	// It is similar to method IterTarFiles but with indices.
	IterIndexTarFiles(pErr *error, acceptNonLocalNames bool) (
		seq2 iter.Seq2[int, *tar.Header], err error)

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

	// IterZipFiles returns an iterator over files in the ZIP archive,
	// traversing it in ascending order by filename.
	//
	// If the reader's file is not archived by ZIP or is opened in raw mode,
	// it does nothing and reports ErrNotZip.
	// (To test whether the error is ErrNotZip, use function errors.Is.)
	//
	// The returned iterator is always non-nil
	// but is a no-op iterator if err is not nil.
	IterZipFiles() (seq iter.Seq[*zip.File], err error)

	// IterIndexZipFiles returns an iterator
	// over index-file pairs in the ZIP archive.
	// It is similar to method IterZipFiles but with indices.
	IterIndexZipFiles() (seq2 iter.Seq2[int, *zip.File], err error)

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
	err  error // should be one of nil, ErrIsDir, ErrReadZip, and ErrFileReaderClosed
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
// If the option Offset is nonzero and the file is not an io.Seeker,
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
func Read(file fs.File, opts *ReadOptions, closeFile bool) (
	r Reader, err error) {
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
			ZipDcomp:        maps.Clone(opts.ZipDcomp),
			ZipReaderAtFunc: opts.ZipReaderAtFunc,
		},
		f: file,
	}
	maps.DeleteFunc(
		fr.opts.ZipDcomp,
		func(method uint16, dcomp zip.Decompressor) bool {
			return dcomp == nil
		},
	)
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
func ReadFromFS(fsys fs.FS, name string, opts *ReadOptions) (
	r Reader, err error) {
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
func (fr *reader) init(
	info fs.FileInfo,
	size int64,
	pClosers *[]io.Closer,
) error {
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
func (fr *reader) initRaw(
	info fs.FileInfo,
	size int64,
	pClosers *[]io.Closer,
) error {
	if fr.opts.Raw {
		return nil
	}
	name := strings.ToLower(info.Name())
	var ext string
	loop := true
	for loop {
		name = name[:len(name)-len(ext)]
		ext = path.Ext(name)
		switch ext {
		case ".tgz":
			name = name[:len(name)-len(ext)] + ".tar.gz"
			ext = ""
		case ".tbz":
			name = name[:len(name)-len(ext)] + ".tar.bz2"
			ext = ""
		case ".gz":
			gr, err := gzip.NewReader(fr.ur)
			if err != nil {
				return err
			}
			*pClosers = append(*pClosers, gr)
			fr.ur = gr
		case ".bz2":
			fr.ur = bzip2.NewReader(fr.ur)
		case ".tar":
			fr.tr = tar.NewReader(fr.ur)
			fr.ur = fr.tr
			loop = false
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
			loop = false
		default:
			loop = false
		}
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

func (fr *reader) ConsumeByte(target byte, n int64) (
	consumed int64, err error) {
	if fr.err != nil {
		return 0, errors.AutoWrap(fr.err)
	}
	consumed, err = fr.br.ConsumeByte(target, n)
	return consumed, errors.AutoWrap(err)
}

func (fr *reader) ConsumeByteFunc(f func(c byte) bool, n int64) (
	consumed int64, err error) {
	if fr.err != nil {
		return 0, errors.AutoWrap(fr.err)
	}
	consumed, err = fr.br.ConsumeByteFunc(f, n)
	return consumed, errors.AutoWrap(err)
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

func (fr *reader) ConsumeRune(target rune, n int64) (
	consumed int64, err error) {
	if fr.err != nil {
		return 0, errors.AutoWrap(fr.err)
	}
	consumed, err = fr.br.ConsumeRune(target, n)
	return consumed, errors.AutoWrap(err)
}

func (fr *reader) ConsumeRuneFunc(f func(r rune, size int) bool, n int64) (
	consumed int64, err error) {
	if fr.err != nil {
		return 0, errors.AutoWrap(fr.err)
	}
	consumed, err = fr.br.ConsumeRuneFunc(f, n)
	return consumed, errors.AutoWrap(err)
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

func (fr *reader) IterLines(pErr *error) iter.Seq[[]byte] {
	if fr.err != nil {
		if pErr == nil {
			return noOpSeq
		}
		// Make a dedicated variable to ensure that
		// the iterator reports the same error each time.
		err := fr.err
		return func(func([]byte) bool) {
			*pErr = errors.AutoWrap(err)
		}
	}
	return fr.br.IterLines(pErr)
}

func (fr *reader) IterCountLines(pErr *error) iter.Seq2[int64, []byte] {
	if fr.err != nil {
		if pErr == nil {
			return noOpSeq2
		}
		// Make a dedicated variable to ensure that
		// the iterator reports the same error each time.
		err := fr.err
		return func(func(int64, []byte) bool) {
			*pErr = errors.AutoWrap(err)
		}
	}
	return fr.br.IterCountLines(pErr)
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
	if err == nil && hdr != nil && !filepath.IsLocal(hdr.Name) {
		err = tar.ErrInsecurePath
	}
	switch {
	case err != nil &&
		!errors.Is(err, io.EOF) &&
		!errors.Is(err, tar.ErrInsecurePath):
		return nil, errors.AutoWrap(err)
	case tarHeaderIsDir(hdr):
		fr.ur, fr.err = isDirErrorReader, ErrIsDir
	default:
		fr.ur, fr.err = fr.tr, nil
	}
	fr.br.Reset(fr.ur)
	return hdr, errors.AutoWrap(err)
}

func (fr *reader) IterTarFiles(pErr *error, acceptNonLocalNames bool) (
	seq iter.Seq[*tar.Header], err error) {
	if fr.tr == nil {
		return noOpSeq, errors.AutoWrap(ErrNotTar)
	} else if fr.c.Closed() {
		return noOpSeq, errors.AutoWrap(ErrFileReaderClosed)
	}
	var tarNextErr error
	return func(yield func(*tar.Header) bool) {
		if pErr != nil {
			defer func(pErr *error) {
				err := tarNextErr
				if errors.Is(err, io.EOF) ||
					acceptNonLocalNames && errors.Is(err, tar.ErrInsecurePath) {
					err = nil
				}
				*pErr = errors.AutoWrapSkip(err, 1) // skip = 1 to skip the inner function
			}(pErr)
		}
		if tarNextErr != nil {
			return
		}
		for {
			var hdr *tar.Header
			hdr, tarNextErr = fr.TarNext()
			if acceptNonLocalNames &&
				errors.Is(tarNextErr, tar.ErrInsecurePath) {
				tarNextErr = nil
			}
			if tarNextErr != nil || yield != nil && !yield(hdr) {
				return
			}
		}
	}, nil
}

func (fr *reader) IterIndexTarFiles(pErr *error, acceptNonLocalNames bool) (
	seq2 iter.Seq2[int, *tar.Header], err error) {
	if fr.tr == nil {
		return noOpSeq2, errors.AutoWrap(ErrNotTar)
	} else if fr.c.Closed() {
		return noOpSeq2, errors.AutoWrap(ErrFileReaderClosed)
	}
	var tarNextErr error
	return func(yield func(int, *tar.Header) bool) {
		if pErr != nil {
			defer func(pErr *error) {
				err := tarNextErr
				if errors.Is(err, io.EOF) ||
					acceptNonLocalNames && errors.Is(err, tar.ErrInsecurePath) {
					err = nil
				}
				*pErr = errors.AutoWrapSkip(err, 1) // skip = 1 to skip the inner function
			}(pErr)
		}
		if tarNextErr != nil {
			return
		}
		for i := 0; ; i++ {
			var hdr *tar.Header
			hdr, tarNextErr = fr.TarNext()
			if acceptNonLocalNames &&
				errors.Is(tarNextErr, tar.ErrInsecurePath) {
				tarNextErr = nil
			}
			if tarNextErr != nil || yield != nil && !yield(i, hdr) {
				return
			}
		}
	}, nil
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
	slices.SortFunc(files, func(a, b *zip.File) int {
		if a.Name < b.Name {
			return -1
		} else if a.Name > b.Name {
			return 1
		}
		return 0
	})
	return
}

func (fr *reader) IterZipFiles() (seq iter.Seq[*zip.File], err error) {
	files, err := fr.ZipFiles()
	if err != nil || len(files) == 0 {
		return noOpSeq, errors.AutoWrap(err)
	}
	return func(yield func(*zip.File) bool) {
		for _, file := range files {
			if !yield(file) {
				return
			}
		}
	}, nil
}

func (fr *reader) IterIndexZipFiles() (
	seq2 iter.Seq2[int, *zip.File], err error) {
	files, err := fr.ZipFiles()
	if err != nil || len(files) == 0 {
		return noOpSeq2, errors.AutoWrap(err)
	}
	return func(yield func(int, *zip.File) bool) {
		for i, file := range files {
			if !yield(i, file) {
				return
			}
		}
	}, nil
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
		ZipDcomp:        maps.Clone(fr.opts.ZipDcomp),
		ZipReaderAtFunc: fr.opts.ZipReaderAtFunc,
	}
	return opts
}

func (fr *reader) FileStat() (info fs.FileInfo, err error) {
	return fr.f.Stat()
}

// tarHeaderIsDir reports whether the tar header represents a directory.
func tarHeaderIsDir(hdr *tar.Header) bool {
	return hdr != nil &&
		(hdr.Typeflag == '\x00' &&
			len(hdr.Name) > 0 &&
			hdr.Name[len(hdr.Name)-1] == '/' ||
			hdr.FileInfo().IsDir())
}

// errorReader implements io.Reader and io.WriterTo.
// Its methods always return 0 and report the specified error.
type errorReader struct {
	// err is the error reported by methods Read and WriteTo.
	//
	// In particular, if err is io.EOF,
	// its method WriteTo returns (0, nil) instead of (0, io.EOF).
	err error
}

func (er *errorReader) Read([]byte) (n int, err error) {
	return 0, errors.AutoWrap(er.err)
}

func (er *errorReader) WriteTo(io.Writer) (n int64, err error) {
	if errors.Is(er.err, io.EOF) {
		return 0, nil
	}
	return 0, errors.AutoWrap(er.err)
}

var (
	isDirErrorReader   = &errorReader{err: ErrIsDir}
	readZipErrorReader = &errorReader{err: ErrReadZip}
	closedErrorReader  = &errorReader{err: ErrFileReaderClosed}
)

// noOpSeq is a no-op iterator of type iter.Seq[V any].
func noOpSeq[V any](func(V) bool) {}

// noOpSeq2 is a no-op iterator of type iter.Seq2[K, V any].
func noOpSeq2[K, V any](func(K, V) bool) {}
