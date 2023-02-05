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

package filesys_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"testing"
	"time"

	"github.com/donyori/gogo/filesys"
	"github.com/donyori/gogo/inout"
)

func TestWrite_Raw(t *testing.T) {
	for _, name := range testFSFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			file := &WritableFileImpl{Name: name}
			data := testFS[name].Data
			writeFile(t, file, data, &filesys.WriteOptions{Raw: true})
			if !t.Failed() && !bytes.Equal(file.Data, data) {
				t.Errorf(
					"file contents - got (len: %d)\n%s\nwant (len: %d)\n%s",
					len(file.Data),
					file.Data,
					len(data),
					data,
				)
			}
		})
	}
}

func TestWrite_Basic(t *testing.T) {
	for _, name := range testFSBasicFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			file := &WritableFileImpl{Name: name}
			data := testFS[name].Data
			writeFile(t, file, data, nil)
			if !t.Failed() && !bytes.Equal(file.Data, data) {
				t.Errorf(
					"file contents - got (len: %d)\n%s\nwant (len: %d)\n%s",
					len(file.Data),
					file.Data,
					len(data),
					data,
				)
			}
		})
	}
}

func TestWrite_Gz(t *testing.T) {
	for _, name := range testFSGzFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			file := &WritableFileImpl{Name: name}
			data := writeGzFile(t, file)
			if t.Failed() {
				return
			}
			gr, err := gzip.NewReader(bytes.NewReader(file.Data))
			if err != nil {
				t.Fatal("create gzip reader -", err)
			}
			defer func(gr *gzip.Reader) {
				if err := gr.Close(); err != nil {
					t.Error("close gzip reader -", err)
				}
			}(gr)
			got, err := io.ReadAll(gr)
			if err != nil {
				t.Fatal("decompress gzip -", err)
			}
			if !bytes.Equal(got, data) {
				t.Errorf("got (len: %d)\n%s\nwant (len: %d)\n%s",
					len(got), got, len(data), data)
			}
		})
	}
}

func TestWrite_TarTgz(t *testing.T) {
	for _, name := range append(testFSTarFilenames, testFSTgzFilenames...) {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			file := &WritableFileImpl{Name: name}
			writeTarFiles(t, file)
			if !t.Failed() {
				testTarTgzFile(t, file)
			}
		})
	}
}

func TestWrite_Zip(t *testing.T) {
	testCases := []struct {
		name string
		f    func(w filesys.Writer, name string) error
	}{
		{"ZipCreate", func(w filesys.Writer, name string) error {
			return w.ZipCreate(name)
		}},
		{"ZipCreateHeader", func(w filesys.Writer, name string) error {
			return w.ZipCreateHeader(&zip.FileHeader{
				Name:   name,
				Method: zip.Store,
			})
		}},
		{"ZipCopy", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file := &WritableFileImpl{Name: "test-write.zip"}
			writeZipFiles(t, file, tc.f)
			if !t.Failed() {
				testZipFile(t, file)
			}
		})
	}
}

func TestWrite_Zip_Raw(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	for name, body := range testFSZipFileNameBodyMap {
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:   name,
			Method: zip.Store,
		})
		if err != nil {
			t.Fatalf("create zip file %q - %v", name, err)
		}
		if len(name) > 0 && name[len(name)-1] == '/' {
			continue
		}
		_, err = w.Write([]byte(body))
		if err != nil {
			t.Fatalf("write zip file %q - %v", name, err)
		}
	}
	err := zw.Close()
	if err != nil {
		t.Fatal("close zip writer -", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal("create zip reader -", err)
	}

	file := &WritableFileImpl{Name: "test-write.zip"}
	writeZipFiles(t, file, func(w filesys.Writer, name string) error {
		for _, f := range zr.File {
			if f.Name == name {
				return w.ZipCreateRaw(&f.FileHeader)
			}
		}
		return fmt.Errorf("unknown file %q", name)
	})
	if !t.Failed() {
		testZipFile(t, file)
	}
}

func TestWrite_AfterClose(t *testing.T) {
	const RegFile = "test-write-after-close.txt"
	const TarFile = "test-write-after-close.tar"
	const ZipFile = "test-write-after-close.zip"
	testCases := []struct {
		methodName string
		filename   string
		f          func(t *testing.T, w filesys.Writer) error
		wantErr    error
		writePanic bool
	}{
		{
			"Close",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				return w.Close()
			},
			nil,
			false,
		},
		{
			"Closed",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				if !w.Closed() {
					t.Error("w.Closed - got false; want true")
				}
				return nil
			},
			nil,
			false,
		},
		{
			"Write",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				_, err := w.Write(nil)
				return err
			},
			filesys.ErrFileWriterClosed,
			false,
		},
		{
			"MustWrite",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				w.MustWrite(nil)
				return nil
			},
			filesys.ErrFileWriterClosed,
			true,
		},
		{
			"WriteByte",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				return w.WriteByte('0')
			},
			filesys.ErrFileWriterClosed,
			false,
		},
		{
			"MustWriteByte",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				w.MustWriteByte('0')
				return nil
			},
			filesys.ErrFileWriterClosed,
			true,
		},
		{
			"WriteRune",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				_, err := w.WriteRune('汉')
				return err
			},
			filesys.ErrFileWriterClosed,
			false,
		},
		{
			"MustWriteRune",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				w.MustWriteRune('汉')
				return nil
			},
			filesys.ErrFileWriterClosed,
			true,
		},
		{
			"WriteString",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				_, err := w.WriteString("")
				return err
			},
			filesys.ErrFileWriterClosed,
			false,
		},
		{
			"MustWriteString",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				w.MustWriteString("")
				return nil
			},
			filesys.ErrFileWriterClosed,
			true,
		},
		{
			"ReadFrom",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				_, err := w.ReadFrom(nil)
				return err
			},
			filesys.ErrFileWriterClosed,
			false,
		},
		{
			"Printf",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				_, err := w.Printf("")
				return err
			},
			filesys.ErrFileWriterClosed,
			false,
		},
		{
			"MustPrintf",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				w.MustPrintf("")
				return nil
			},
			filesys.ErrFileWriterClosed,
			true,
		},
		{
			"Print",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				_, err := w.Print()
				return err
			},
			filesys.ErrFileWriterClosed,
			false,
		},
		{
			"MustPrint",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				w.MustPrint()
				return nil
			},
			filesys.ErrFileWriterClosed,
			true,
		},
		{
			"Println",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				_, err := w.Println()
				return err
			},
			filesys.ErrFileWriterClosed,
			false,
		},
		{
			"MustPrintln",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				w.MustPrintln()
				return nil
			},
			filesys.ErrFileWriterClosed,
			true,
		},
		{
			"Flush-noBufferedData",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				return w.Flush()
			},
			nil,
			false,
		},
		{
			"TarWriteHeader-notTar",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				return w.TarWriteHeader(nil)
			},
			filesys.ErrNotTar,
			false,
		},
		{
			"TarWriteHeader-isTar",
			TarFile,
			func(t *testing.T, w filesys.Writer) error {
				return w.TarWriteHeader(nil)
			},
			filesys.ErrFileWriterClosed,
			false,
		},
		{
			"ZipCreate-notZip",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				return w.ZipCreate("")
			},
			filesys.ErrNotZip,
			false,
		},
		{
			"ZipCreate-isZip",
			ZipFile,
			func(t *testing.T, w filesys.Writer) error {
				return w.ZipCreate("")
			},
			filesys.ErrFileWriterClosed,
			false,
		},
		{
			"ZipCreateHeader-notZip",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				return w.ZipCreateHeader(nil)
			},
			filesys.ErrNotZip,
			false,
		},
		{
			"ZipCreateHeader-isZip",
			ZipFile,
			func(t *testing.T, w filesys.Writer) error {
				return w.ZipCreateHeader(nil)
			},
			filesys.ErrFileWriterClosed,
			false,
		},
		{
			"ZipCreateRaw-notZip",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				return w.ZipCreateRaw(nil)
			},
			filesys.ErrNotZip,
			false,
		},
		{
			"ZipCreateRaw-isZip",
			ZipFile,
			func(t *testing.T, w filesys.Writer) error {
				return w.ZipCreateRaw(nil)
			},
			filesys.ErrFileWriterClosed,
			false,
		},
		{
			"ZipCopy-notZip",
			RegFile,
			func(t *testing.T, w filesys.Writer) error {
				return w.ZipCopy(nil)
			},
			filesys.ErrNotZip,
			false,
		},
		{
			"ZipCopy-isZip",
			ZipFile,
			func(t *testing.T, w filesys.Writer) error {
				return w.ZipCopy(nil)
			},
			filesys.ErrFileWriterClosed,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run("method="+tc.methodName, func(t *testing.T) {
			file := &WritableFileImpl{Name: tc.filename}
			w, err := filesys.Write(file, nil, true)
			if err != nil {
				t.Fatal("create -", err)
			}
			err = w.Close()
			if err != nil {
				t.Fatal("close -", err)
			}
			defer func() {
				e := recover()
				if tc.writePanic {
					if wp, ok := e.(*inout.WritePanic); ok {
						if !errors.Is(wp, tc.wantErr) {
							t.Errorf("got panic %v; want %v", e, tc.wantErr)
						}
					} else {
						t.Errorf("got panic %v (type: %[1]T); want *inout.WritePanic", e)
					}
				} else if e != nil {
					t.Error("panic -", e)
				}
			}()
			err = tc.f(t, w)
			if tc.writePanic {
				t.Error("function returned but want panic")
			} else if !errors.Is(err, tc.wantErr) {
				t.Errorf("got error %v; want %v", err, tc.wantErr)
			}
		})
	}

	t.Run("method=Flush-hasBufferedData", func(t *testing.T) {
		file := &WritableFileImpl{Name: RegFile}
		const Input = "Flush should return nil"
		w, err := filesys.Write(file, &filesys.WriteOptions{BufSize: len(Input) + 10}, true)
		if err != nil {
			t.Fatal("create -", err)
		}
		_, err = w.WriteString(Input)
		if err != nil {
			t.Fatal("write string -", err)
		}
		if n := w.Buffered(); n != len(Input) {
			t.Fatalf("got w.Buffered %d; want %d", n, len(Input))
		}
		err = w.Close()
		if err != nil {
			t.Fatal("close -", err)
		}
		err = w.Flush()
		// The buffered data should be flushed in Close, so err should be nil.
		if err != nil {
			t.Errorf("got error %v; want nil", err)
		}
	})
}

// writeFile writes data to file using Write.
//
// It closes file after writing.
func writeFile(t *testing.T, file *WritableFileImpl, data []byte, opts *filesys.WriteOptions) {
	w, err := filesys.Write(file, opts, true)
	if err != nil {
		t.Error("create -", err)
		return
	}
	defer func(w filesys.Writer) {
		if err := w.Close(); err != nil {
			t.Error("close -", err)
		}
	}(w)
	n, err := w.Write(data)
	if n != len(data) || err != nil {
		t.Errorf("write - got (%d, %v); want (%d, nil)", n, err, len(data))
	}
}

// writeGzFile loads data from testFS according to file.Name,
// and then writes the data to file using Write.
//
// It returns the data written to file before gzip compression.
//
// Caller should set file.Name before calling this function and
// guarantee that file.Name has extension ".gz".
func writeGzFile(t *testing.T, file *WritableFileImpl) (data []byte) {
	dataFilename := file.Name[:len(file.Name)-3]
	mapFile := testFS[dataFilename]
	if mapFile == nil {
		t.Errorf("file %q does not exist", dataFilename)
		return
	}
	data = mapFile.Data
	writeFile(t, file, data, nil)
	return
}

// writeTarFiles writes testFSTarFiles to file using Write
// and then close the file.
//
// Caller should set file.Name before calling this function and
// guarantee that file.Name has extension ".tar", ".tar.gz", or ".tgz".
func writeTarFiles(t *testing.T, file *WritableFileImpl) {
	w, err := filesys.Write(file, nil, true)
	if err != nil {
		t.Error("create -", err)
		return
	}
	defer func(w filesys.Writer) {
		if err := w.Close(); err != nil {
			t.Error("close -", err)
		}
	}(w)
	for i := range testFSTarFiles {
		hdr := &tar.Header{
			Name:    testFSTarFiles[i].name,
			Size:    int64(len(testFSTarFiles[i].body)),
			Mode:    0600,
			ModTime: time.Now(),
		}
		err = w.TarWriteHeader(hdr)
		if err != nil {
			t.Errorf("write No.%d tar header - %v", i, err)
			return
		}
		var n int
		n, err = w.WriteString(testFSTarFiles[i].body)
		if filesys.TarHeaderIsDir(hdr) {
			if n != 0 || !errors.Is(err, filesys.ErrIsDir) {
				t.Errorf("write No.%d tar file body - got (%d, %v); want (0, %v)",
					i, n, err, filesys.ErrIsDir)
				return
			}
		} else if n != len(testFSTarFiles[i].body) || err != nil {
			t.Errorf("write No.%d tar file body - got (%d, %v); want (%d, nil)",
				i, n, err, len(testFSTarFiles[i].body))
			return
		}
	}
}

// testTarTgzFile checks file written by function writeTarFiles.
//
// Caller should guarantee that file.Name has extension
// ".tar", ".tar.gz", or ".tgz".
func testTarTgzFile(t *testing.T, file *WritableFileImpl) {
	var r io.Reader = bytes.NewReader(file.Data)
	ext := path.Ext(file.Name)
	if ext == ".gz" || ext == ".tgz" {
		gr, err := gzip.NewReader(r)
		if err != nil {
			t.Error("create gzip reader -", err)
			return
		}
		defer func(gr *gzip.Reader) {
			if err := gr.Close(); err != nil {
				t.Error("close gzip reader -", err)
			}
		}(gr)
		r = gr
	}
	tr := tar.NewReader(r)
	for i := 0; ; i++ {
		hdr, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if i != len(testFSTarFiles) {
					t.Errorf("tar header number: %d != %d, but got EOF",
						i, len(testFSTarFiles))
				}
				return // end of archive
			}
			t.Errorf("read No.%d tar header - %v", i, err)
			return
		}
		if i >= len(testFSTarFiles) {
			t.Error("tar headers more than", len(testFSTarFiles))
			return
		}
		if hdr.Name != testFSTarFiles[i].name {
			t.Errorf("No.%d tar header name unequal - got %s; want %s",
				i, hdr.Name, testFSTarFiles[i].name)
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Errorf("read No.%d tar file body - %v", i, err)
			return
		}
		if string(body) != testFSTarFiles[i].body {
			t.Errorf(
				"got No.%d tar file body (len: %d)\n%s\nwant (len: %d)\n%s",
				i,
				len(body),
				body,
				len(testFSTarFiles[i].body),
				testFSTarFiles[i].body,
			)
		}
	}
}

// writeZipFiles writes testFSZipFileNameBodyMap to file using Write
// and then close the file.
//
// createFn is a function that calls
// w.ZipCreate, w.ZipCreateHeader, or w.ZipCreateRaw.
// If createFn is nil, writeZipFiles will write the file through ZipCopy.
//
// Caller should set file.Name before calling this function and
// guarantee that file.Name has extension ".zip".
func writeZipFiles(
	t *testing.T,
	file *WritableFileImpl,
	createFn func(w filesys.Writer, name string) error,
) {
	w, err := filesys.Write(
		file,
		&filesys.WriteOptions{
			DeflateLv:  flate.BestCompression,
			ZipComment: testFSZipComment,
		},
		true,
	)
	if err != nil {
		t.Error("create -", err)
		return
	}
	defer func(w filesys.Writer) {
		if err := w.Close(); err != nil {
			t.Error("close -", err)
		}
	}(w)

	if createFn != nil {
		for name, body := range testFSZipFileNameBodyMap {
			err = createFn(w, name)
			if err != nil {
				t.Errorf("create %q - %v", name, err)
				return
			}
			var n int
			n, err = w.WriteString(body)
			if len(name) > 0 && name[len(name)-1] == '/' {
				if n != 0 || !errors.Is(err, filesys.ErrIsDir) {
					t.Errorf("write %q file body - got (%d, %v); want (0, %v)",
						name, n, err, filesys.ErrIsDir)
					return
				}
			} else if n != len(body) || err != nil {
				t.Errorf("write %q file body - got (%d, %v); want (%d, nil)",
					name, n, err, len(body))
				return
			}
		}
		return
	}

	zipFile, err := testFS.Open(testFSZipFilenames[0])
	if err != nil {
		t.Errorf("open zip file %q - %v", testFSZipFilenames[0], err)
		return
	}
	defer func(f fs.File) {
		_ = f.Close() // ignore error
	}(zipFile)
	zipInfo, err := zipFile.Stat()
	if err != nil {
		t.Errorf("zip file %q stat - %v", testFSZipFilenames[0], err)
		return
	}
	zr, err := zip.NewReader(zipFile.(io.ReaderAt), zipInfo.Size())
	if err != nil {
		t.Error("create zip reader -", err)
	}
	for _, f := range zr.File {
		err = w.ZipCopy(f)
		if err != nil {
			t.Errorf("copy %q - %v", f.Name, err)
			return
		}
		_, err = w.Write(nil)
		if !errors.Is(err, filesys.ErrZipWriteBeforeCreate) {
			t.Errorf("call Write after ZipCopy - got %v; want %v",
				err, filesys.ErrZipWriteBeforeCreate)
		}
	}
}

// testTarTgzFile checks file written by function writeZipFiles.
//
// Caller should guarantee that file.Name has extension ".zip".
func testZipFile(t *testing.T, file *WritableFileImpl) {
	r, err := zip.NewReader(bytes.NewReader(file.Data), int64(len(file.Data)))
	if err != nil {
		t.Error("create zip reader -", err)
		return
	}
	if r.Comment != testFSZipComment {
		t.Errorf("got comment %q; want %q", r.Comment, testFSZipComment)
	}
	if len(r.File) != len(testFSZipFileNameBodyMap) {
		t.Errorf("got %d zip files; want %d",
			len(r.File), len(testFSZipFileNameBodyMap))
	}

	for _, file := range r.File {
		body, ok := testFSZipFileNameBodyMap[file.Name]
		if !ok {
			t.Errorf("unknown zip file %q", file.Name)
			continue
		}
		isDir := len(file.Name) > 0 && file.Name[len(file.Name)-1] == '/'
		if d := file.FileInfo().IsDir(); d != isDir {
			t.Errorf("got IsDir %t; want %t", d, isDir)
		}
		if isDir {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			t.Errorf("open %q - %v", file.Name, err)
			return
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close() // ignore error
		if err != nil {
			t.Errorf("read %q - %v", file.Name, err)
			return
		}
		if string(data) != body {
			t.Errorf(
				"%q file contents - got (len: %d)\n%s\nwant (len: %d)\n%s",
				file.Name,
				len(data),
				data,
				len(body),
				body,
			)
		}
	}
}
