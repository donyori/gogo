// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2025  Yuan Gao
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
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"path"
	"strings"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/inout"
)

const maxUint16 int = 1<<16 - 1

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
	// BufSize is the size of the buffer for writing the file at least.
	//
	// Nonpositive values for using default value.
	BufSize int

	// Raw indicates whether to not compress the file with gzip
	// and whether to not archive the file with tar (i.e., tape archive)
	// according to the file extension.
	Raw bool

	// DeflateLv is the compression level of DEFLATE.
	//
	// It should be in the range [-2, 9].
	// The zero value (0) stands for no compression
	// other than the default value.
	// To use the default value, set it to compress/flate.DefaultCompression.
	// For more details, see the documentation of compress/flate.NewWriter.
	//
	// This option only takes effect when DEFLATE compression is applied.
	// That is, when Raw is false, and the file extension is ".gz" or ".tgz",
	// or ".zip" and the ZIP archive uses DEFLATE compression.
	DeflateLv int

	// ZipOffset is the offset of the beginning of the ZIP data
	// within the underlying writer.
	//
	// It should be used when the ZIP data is appended to an existing file,
	// such as a binary executable.
	//
	// Nonpositive values are ignored.
	ZipOffset int64

	// ZipComment is the end-of-central-directory comment field
	// of the ZIP archive.
	//
	// It should be 65535 bytes at most.
	// If the comment is too long, an error is reported.
	ZipComment string

	// ZipComp is a method-compressor map for writing the ZIP archive.
	// These compressors are registered to the archive/zip.Writer.
	// (Nil compressors are ignored.)
	//
	// Package archive/zip has two built-in compressors for the common methods
	// archive/zip.Store (0) and archive/zip.Deflate (8).
	//
	// For more details, see the documentation of the method RegisterCompressor
	// of archive/zip.Writer and the function archive/zip.RegisterCompressor.
	ZipComp map[uint16]zip.Compressor
}

// defaultWriteOptions are default options for Write functions.
var defaultWriteOptions = &WriteOptions{DeflateLv: flate.BestCompression}

// Writer is a device to write data to a local file.
//
// Its method Close closes all closable objects opened by this writer
// (may include the file).
// After successfully closing this writer,
// its method Close does nothing and returns nil,
// and its write methods report ErrFileWriterClosed.
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
	// If the current file is not fully written, it returns an error.
	// It implicitly flushes any padding necessary before writing the header.
	//
	// If the file is not archived by tar or is opened in raw mode,
	// it does nothing and reports ErrNotTar.
	// (To test whether the error is ErrNotTar, use function errors.Is.)
	TarWriteHeader(hdr *tar.Header) error

	// TarAddFS adds the files from the specified filesystem
	// to the tape archive.
	// It walks the directory tree starting at the root of the filesystem
	// adding each file to the tape archive
	// while maintaining the directory structure.
	//
	// If the file is not archived by tar or is opened in raw mode,
	// it does nothing and reports ErrNotTar.
	// (To test whether the error is ErrNotTar, use function errors.Is.)
	TarAddFS(fsys fs.FS) error

	// ZipEnabled returns true if the file is archived by ZIP
	// and is not opened in raw mode.
	ZipEnabled() bool

	// ZipCreate adds a file with specified name to the ZIP archive and
	// switches the writer to that file.
	//
	// The file contents are compressed with DEFLATE.
	//
	// The name must be a relative path: it must not start with a drive letter
	// (e.g., "C:") or leading slash. Only forward slashes are allowed;
	// "../" elements are not allowed.
	// To create a directory instead of a file, add a trailing slash
	// to the name (e.g., "dir/").
	//
	// The file's contents must be written to the writer before the next call
	// to ZipCreate, ZipCreateHeader, ZipCreateRaw, ZipCopy, or Close.
	//
	// If the writer's file is not archived by ZIP or is opened in raw mode,
	// it does nothing and reports ErrNotZip.
	// (To test whether the error is ErrNotZip, use function errors.Is.)
	ZipCreate(name string) error

	// ZipCreateHeader adds a file to the ZIP archive using the specified
	// file header for the file metadata and switches the writer to that file.
	//
	// The writer takes ownership of fh and may mutate its fields.
	// The client must not modify fh after calling ZipCreateHeader.
	//
	// The file's contents must be written to the writer before the next call
	// to ZipCreate, ZipCreateHeader, ZipCreateRaw, ZipCopy, or Close.
	//
	// If the writer's file is not archived by ZIP or is opened in raw mode,
	// it does nothing and reports ErrNotZip.
	// (To test whether the error is ErrNotZip, use function errors.Is.)
	ZipCreateHeader(fh *zip.FileHeader) error

	// ZipCreateRaw is like ZipCreateHeader,
	// but the bytes passed to the writer are not compressed.
	ZipCreateRaw(fh *zip.FileHeader) error

	// ZipCopy copies the file f (obtained from an archive/zip.Reader)
	// into the writer.
	//
	// It copies the raw form directly bypassing
	// decompression, compression, and validation.
	//
	// If the writer's file is not archived by ZIP or is opened in raw mode,
	// it does nothing and reports ErrNotZip.
	// (To test whether the error is ErrNotZip, use function errors.Is.)
	ZipCopy(f *zip.File) error

	// ZipAddFS adds the files from the specified filesystem
	// to the ZIP archive.
	// It walks the directory tree starting at the root of the filesystem
	// adding each file to the ZIP using deflate
	// while maintaining the directory structure.
	//
	// If the writer's file is not archived by ZIP or is opened in raw mode,
	// it does nothing and reports ErrNotZip.
	// (To test whether the error is ErrNotZip, use function errors.Is.)
	ZipAddFS(fsys fs.FS) error

	// Options returns a copy of options used by this writer.
	Options() *WriteOptions

	// FileStat returns the io/fs.FileInfo structure describing file.
	FileStat() (info fs.FileInfo, err error)
}

// writer is an implementation of interface Writer.
//
// Use it with Write functions.
type writer struct {
	err  error // should be one of nil, ErrIsDir, ErrZipWriteBeforeCreate, and ErrFileWriterClosed
	bw   inout.ResettableBufferedWriter
	uw   io.Writer // unbuffered writer
	c    inout.Closer
	opts WriteOptions
	f    WritableFile
	tw   *tar.Writer
	zw   *zip.Writer
}

// Write creates a writer on the specified file with options opts.
//
// If the file is a directory, Write reports ErrIsDir and returns a nil Writer.
// (To test whether err is ErrIsDir, use function errors.Is.)
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - BufSize: 0
//   - Raw: false
//   - DeflateLv: compress/flate.BestCompression
//   - ZipOffset: 0
//   - ZipComment: ""
//   - ZipComp: nil
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
// In this case, the file is closed during the method Close of the writer,
// and it is also closed by this function when encountering an error.
//
// This function panics if file is nil.
func Write(file WritableFile, opts *WriteOptions, closeFile bool) (
	w Writer, err error) {
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
	if opts.DeflateLv < -2 || opts.DeflateLv > 9 {
		return nil, errors.AutoWrap(fmt.Errorf(
			"option DeflateLv (%d) is out of range [-2, 9]",
			opts.DeflateLv,
		))
	}
	if len(opts.ZipComment) > maxUint16 {
		return nil, errors.AutoWrap(fmt.Errorf(
			"option ZipComment (len: %d) exceeds %d bytes",
			len(opts.ZipComment),
			maxUint16,
		))
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
		uw: file,
		opts: WriteOptions{
			BufSize:    opts.BufSize,
			Raw:        opts.Raw,
			DeflateLv:  opts.DeflateLv,
			ZipOffset:  opts.ZipOffset,
			ZipComment: opts.ZipComment,
			ZipComp:    maps.Clone(opts.ZipComp),
		},
		f: file,
	}
	maps.DeleteFunc(
		fw.opts.ZipComp,
		func(method uint16, comp zip.Compressor) bool {
			return comp == nil
		},
	)
	w = fw
	el.Append(fw.init(info, &closers))
	return
}

// init initializes the writer according to the options.
//
// It may update closers.
func (fw *writer) init(info fs.FileInfo, pClosers *[]io.Closer) error {
	err := fw.initRaw(info, pClosers)
	if err != nil {
		return err
	}
	fw.initCloserAndBuffer(*pClosers)
	return nil
}

// initRaw deals with the option Raw.
//
// It may update closers.
func (fw *writer) initRaw(info fs.FileInfo, pClosers *[]io.Closer) error {
	if fw.opts.Raw {
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
		case ".gz":
			gw, err := gzip.NewWriterLevel(fw.uw, fw.opts.DeflateLv)
			if err != nil {
				return err
			}
			*pClosers = append(*pClosers, gw)
			fw.uw = gw
		case ".tar":
			fw.tw = tar.NewWriter(fw.uw)
			*pClosers = append(*pClosers, fw.tw)
			fw.uw = fw.tw
			loop = false
		case ".zip":
			fw.zw = zip.NewWriter(fw.uw)
			*pClosers = append(*pClosers, fw.zw)
			if fw.opts.ZipOffset > 0 {
				fw.zw.SetOffset(fw.opts.ZipOffset)
			}
			if fw.opts.ZipComment != "" {
				err := fw.zw.SetComment(fw.opts.ZipComment)
				if err != nil {
					return err
				}
			}
			noDeflate := true
			for method, comp := range fw.opts.ZipComp {
				if method == zip.Deflate {
					noDeflate = false
				}
				fw.zw.RegisterCompressor(method, comp)
			}
			if noDeflate {
				fw.zw.RegisterCompressor(
					zip.Deflate,
					func(w io.Writer) (io.WriteCloser, error) {
						return flate.NewWriter(w, fw.opts.DeflateLv)
					},
				)
			}
			fw.uw = zipWriteBeforeCreateErrorWriter
			fw.err = ErrZipWriteBeforeCreate
			loop = false
		default:
			loop = false
		}
	}
	return nil
}

// initCloserAndBuffer sets fw.c and creates a buffer.
func (fw *writer) initCloserAndBuffer(closers []io.Closer) {
	switch len(closers) {
	case 1:
		fw.c = inout.WrapNoErrorCloser(closers[0])
	case 0:
		fw.c = inout.NewNoOpCloser()
	default:
		fw.c = inout.NewMultiCloser(true, true, closers...)
	}
	fw.bw = inout.NewBufferedWriterSize(fw.uw, fw.opts.BufSize)
}

func (fw *writer) Close() error {
	if fw.c.Closed() {
		return nil
	}
	flushErr := fw.bw.Flush()
	closeErr := fw.c.Close()
	if fw.c.Closed() {
		fw.uw, fw.err = closedErrorWriter, ErrFileWriterClosed
		fw.bw.Reset(fw.uw)
	}
	return errors.AutoWrap(errors.Combine(flushErr, closeErr))
}

func (fw *writer) Closed() bool {
	return fw.c.Closed()
}

func (fw *writer) Write(p []byte) (n int, err error) {
	if fw.err != nil {
		return 0, errors.AutoWrap(fw.err)
	}
	n, err = fw.bw.Write(p)
	return n, errors.AutoWrap(err)
}

func (fw *writer) MustWrite(p []byte) (n int) {
	n, err := fw.Write(p)
	if err != nil {
		panic(inout.NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

func (fw *writer) WriteByte(c byte) error {
	if fw.err != nil {
		return errors.AutoWrap(fw.err)
	}
	return errors.AutoWrap(fw.bw.WriteByte(c))
}

func (fw *writer) MustWriteByte(c byte) {
	err := fw.WriteByte(c)
	if err != nil {
		panic(inout.NewWritePanic(errors.AutoWrap(err)))
	}
}

func (fw *writer) WriteRune(r rune) (size int, err error) {
	if fw.err != nil {
		return 0, errors.AutoWrap(fw.err)
	}
	size, err = fw.bw.WriteRune(r)
	return size, errors.AutoWrap(err)
}

func (fw *writer) MustWriteRune(r rune) (size int) {
	size, err := fw.WriteRune(r)
	if err != nil {
		panic(inout.NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

func (fw *writer) WriteString(s string) (n int, err error) {
	if fw.err != nil {
		return 0, errors.AutoWrap(fw.err)
	}
	n, err = fw.bw.WriteString(s)
	return n, errors.AutoWrap(err)
}

func (fw *writer) MustWriteString(s string) (n int) {
	n, err := fw.WriteString(s)
	if err != nil {
		panic(inout.NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

func (fw *writer) ReadFrom(r io.Reader) (n int64, err error) {
	if fw.err != nil {
		return 0, errors.AutoWrap(fw.err)
	}
	n, err = fw.bw.ReadFrom(r)
	return n, errors.AutoWrap(err)
}

func (fw *writer) Printf(format string, args ...any) (n int, err error) {
	if fw.err != nil {
		return 0, errors.AutoWrap(fw.err)
	}
	n, err = fw.bw.Printf(format, args...)
	return n, errors.AutoWrap(err)
}

func (fw *writer) MustPrintf(format string, args ...any) (n int) {
	n, err := fw.Printf(format, args...)
	if err != nil {
		panic(inout.NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

func (fw *writer) Print(args ...any) (n int, err error) {
	if fw.err != nil {
		return 0, errors.AutoWrap(fw.err)
	}
	n, err = fw.bw.Print(args...)
	return n, errors.AutoWrap(err)
}

func (fw *writer) MustPrint(args ...any) (n int) {
	n, err := fw.Print(args...)
	if err != nil {
		panic(inout.NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

func (fw *writer) Println(args ...any) (n int, err error) {
	if fw.err != nil {
		return 0, errors.AutoWrap(fw.err)
	}
	n, err = fw.bw.Println(args...)
	return n, errors.AutoWrap(err)
}

func (fw *writer) MustPrintln(args ...any) (n int) {
	n, err := fw.Println(args...)
	if err != nil {
		panic(inout.NewWritePanic(errors.AutoWrap(err)))
	}
	return
}

func (fw *writer) Flush() error {
	if fw.bw.Buffered() == 0 {
		// If there is no buffered data, do nothing and return nil,
		// even if fw.err is not nil.
		return nil
	} else if fw.err != nil {
		return errors.AutoWrap(fw.err)
	}
	return errors.AutoWrap(fw.bw.Flush())
}

func (fw *writer) Size() int {
	return fw.bw.Size()
}

func (fw *writer) Buffered() int {
	return fw.bw.Buffered()
}

func (fw *writer) Available() int {
	return fw.bw.Available()
}

func (fw *writer) TarEnabled() bool {
	return fw.tw != nil
}

func (fw *writer) TarWriteHeader(hdr *tar.Header) error {
	err := fw.tarCheckAndFlush()
	if err != nil {
		return errors.AutoWrap(err)
	}
	err = fw.tw.WriteHeader(hdr)
	switch {
	case err != nil:
		return errors.AutoWrap(err)
	case tarHeaderIsDir(hdr):
		fw.uw, fw.err = isDirErrorWriter, ErrIsDir
	default:
		fw.uw, fw.err = fw.tw, nil
	}
	fw.bw.Reset(fw.uw)
	return nil
}

func (fw *writer) TarAddFS(fsys fs.FS) error {
	err := fw.tarCheckAndFlush()
	if err != nil {
		return errors.AutoWrap(err)
	} else if fsys == nil {
		return nil
	}
	return errors.AutoWrap(fw.tw.AddFS(fsys))
}

func (fw *writer) ZipEnabled() bool {
	return fw.zw != nil
}

func (fw *writer) ZipCreate(name string) error {
	return errors.AutoWrap(fw.zipCreateFunc(
		nil,
		name,
		func() (io.Writer, error) {
			return fw.zw.Create(name)
		},
	))
}

func (fw *writer) ZipCreateHeader(fh *zip.FileHeader) error {
	return errors.AutoWrap(fw.zipCreateFunc(
		fh,
		"",
		func() (io.Writer, error) {
			return fw.zw.CreateHeader(fh)
		},
	))
}

func (fw *writer) ZipCreateRaw(fh *zip.FileHeader) error {
	return errors.AutoWrap(fw.zipCreateFunc(
		fh,
		"",
		func() (io.Writer, error) {
			return fw.zw.CreateRaw(fh)
		},
	))
}

func (fw *writer) ZipCopy(f *zip.File) error {
	err := fw.zipCheckAndFlush()
	if err != nil {
		return errors.AutoWrap(err)
	}
	fw.uw, fw.err = zipWriteBeforeCreateErrorWriter, ErrZipWriteBeforeCreate
	fw.bw.Reset(fw.uw)
	return errors.AutoWrap(fw.zw.Copy(f))
}

func (fw *writer) ZipAddFS(fsys fs.FS) error {
	err := fw.zipCheckAndFlush()
	if err != nil {
		return errors.AutoWrap(err)
	} else if fsys == nil {
		return nil
	}
	err = fw.zw.AddFS(fsys)
	if err != nil {
		return errors.AutoWrap(err)
	}
	fw.uw, fw.err = zipWriteBeforeCreateErrorWriter, ErrZipWriteBeforeCreate
	fw.bw.Reset(fw.uw)
	return nil
}

func (fw *writer) Options() *WriteOptions {
	opts := &WriteOptions{
		BufSize:    fw.opts.BufSize,
		Raw:        fw.opts.Raw,
		DeflateLv:  fw.opts.DeflateLv,
		ZipOffset:  fw.opts.ZipOffset,
		ZipComment: fw.opts.ZipComment,
		ZipComp:    maps.Clone(fw.opts.ZipComp),
	}
	return opts
}

func (fw *writer) FileStat() (info fs.FileInfo, err error) {
	return fw.f.Stat()
}

// tarCheckAndFlush checks whether the writer is in tar mode and not closed.
// If so, it flushes the buffer and returns any error encountered.
// If not, it reports the corresponding error.
func (fw *writer) tarCheckAndFlush() error {
	if fw.tw == nil {
		return errors.AutoWrap(ErrNotTar)
	} else if fw.c.Closed() {
		return errors.AutoWrap(ErrFileWriterClosed)
	}
	return errors.AutoWrap(fw.bw.Flush())
}

// zipCheckAndFlush checks whether the writer is in ZIP mode and not closed.
// If so, it flushes the buffer and returns any error encountered.
// If not, it reports the corresponding error.
func (fw *writer) zipCheckAndFlush() error {
	if fw.zw == nil {
		return errors.AutoWrap(ErrNotZip)
	} else if fw.c.Closed() {
		return errors.AutoWrap(ErrFileWriterClosed)
	}
	return errors.AutoWrap(fw.bw.Flush())
}

// zipCreateFunc is a framework for ZipCreate, ZipCreateHeader,
// and ZipCreateRaw.
//
// It may mutate fw.err, fw.uw, and fw.bw.
//
// fh is the argument of ZipCreateHeader and ZipCreateRaw or nil for ZipCreate.
//
// name is the argument of ZipCreate or
// empty for ZipCreateHeader and ZipCreateRaw.
//
// f is a function that calls Create, CreateHeader, or CreateRaw of fw.zw.
func (fw *writer) zipCreateFunc(
	fh *zip.FileHeader,
	name string,
	f func() (io.Writer, error),
) error {
	err := fw.zipCheckAndFlush()
	if err != nil {
		return errors.AutoWrap(err)
	}
	w, err := f()
	if fh != nil {
		name = fh.Name
	}
	if err == nil {
		if len(name) == 0 || name[len(name)-1] != '/' {
			fw.uw, fw.err = w, nil
		} else {
			fw.uw, fw.err = isDirErrorWriter, ErrIsDir
		}
	} else {
		fw.uw = zipWriteBeforeCreateErrorWriter
		fw.err = ErrZipWriteBeforeCreate
	}
	fw.bw.Reset(fw.uw)
	return errors.AutoWrap(err)
}

// errorWriter implements io.Writer and io.ReaderFrom.
// Its methods always return 0 and report the specified error.
type errorWriter struct {
	// err is the error reported by methods Write and ReadFrom.
	//
	// In particular, if err is io.EOF,
	// its method ReadFrom returns (0, nil) instead of (0, io.EOF).
	err error
}

func (ew *errorWriter) Write([]byte) (n int, err error) {
	return 0, errors.AutoWrap(ew.err)
}

func (ew *errorWriter) ReadFrom(io.Reader) (n int64, err error) {
	if errors.Is(ew.err, io.EOF) {
		return 0, nil
	}
	return 0, errors.AutoWrap(ew.err)
}

var (
	isDirErrorWriter                = &errorWriter{err: ErrIsDir}
	zipWriteBeforeCreateErrorWriter = &errorWriter{err: ErrZipWriteBeforeCreate}
	closedErrorWriter               = &errorWriter{err: ErrFileWriterClosed}
)
