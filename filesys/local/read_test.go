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

package local_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"iter"
	"math"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/filesys"
	"github.com/donyori/gogo/filesys/local"
	"github.com/donyori/gogo/internal/unequal"
)

func TestRead_Raw(t *testing.T) {
	for _, entry := range testFileEntries {
		filename := entry.Name()
		t.Run(fmt.Sprintf("file=%+q", filename), func(t *testing.T) {
			name := filepath.Join(TestDataDir, filename)
			r, err := local.Read(name, &filesys.ReadOptions{Raw: true})
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}(r)
			data, err := lazyLoadTestData(name)
			if err != nil {
				t.Fatal("read file -", err)
			}
			err = iotest.TestReader(r, data)
			if err != nil {
				t.Error("test read -", err)
			}
		})
	}
}

func TestRead_Basic(t *testing.T) {
	for _, entry := range testFileEntries {
		filename := entry.Name()
		if ext := filepath.Ext(filename); ext == ".gz" || ext == ".bz2" ||
			ext == ".tar" || ext == ".tgz" || ext == ".tbz" || ext == ".zip" {
			continue
		}
		t.Run(fmt.Sprintf("file=%+q", filename), func(t *testing.T) {
			name := filepath.Join(TestDataDir, filename)
			r, err := local.Read(name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}(r)
			data, err := lazyLoadTestData(name)
			if err != nil {
				t.Fatal("read file -", err)
			}
			err = iotest.TestReader(r, data)
			if err != nil {
				t.Error("test read -", err)
			}
		})
	}
}

func TestRead_Gz(t *testing.T) {
	for _, entry := range testFileEntries {
		filename := entry.Name()
		if filepath.Ext(filename) != ".gz" ||
			filepath.Ext(filename[:len(filename)-3]) == ".tar" {
			continue
		}
		t.Run(fmt.Sprintf("file=%+q", filename), func(t *testing.T) {
			name := filepath.Join(TestDataDir, filename)
			r, err := local.Read(name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}(r)
			data, err := lazyLoadTestData(name)
			if err != nil {
				t.Fatal("read file -", err)
			}
			gr, err := gzip.NewReader(bytes.NewReader(data))
			if err != nil {
				t.Fatal("create gzip reader -", err)
			}
			defer func(gr *gzip.Reader) {
				if err := gr.Close(); err != nil {
					t.Error("close gzip reader -", err)
				}
			}(gr)
			want, err := io.ReadAll(gr)
			if err != nil {
				t.Fatal("decompress gzip -", err)
			}
			err = iotest.TestReader(r, want)
			if err != nil {
				t.Error("test read -", err)
			}
		})
	}
}

func TestRead_Bz2(t *testing.T) {
	for _, entry := range testFileEntries {
		filename := entry.Name()
		if filepath.Ext(filename) != ".bz2" ||
			filepath.Ext(filename[:len(filename)-4]) == ".tar" {
			continue
		}
		t.Run(fmt.Sprintf("file=%+q", filename), func(t *testing.T) {
			name := filepath.Join(TestDataDir, filename)
			r, err := local.Read(name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}(r)
			data, err := lazyLoadTestData(name)
			if err != nil {
				t.Fatal("read file -", err)
			}
			want, err := io.ReadAll(bzip2.NewReader(bytes.NewReader(data)))
			if err != nil {
				t.Fatal("decompress bzip2 -", err)
			}
			err = iotest.TestReader(r, want)
			if err != nil {
				t.Error("test read -", err)
			}
		})
	}
}

func TestRead_TarTgzTbz(t *testing.T) {
	testReadTarTgzTbzFunc(t, testReadTarTgzTbz)
}

func TestRead_TarTgzTbz_Seq(t *testing.T) {
	testReadTarTgzTbzFunc(t, testReadTarTgzTbzSeq)
}

func TestRead_TarTgzTbz_Seq2(t *testing.T) {
	testReadTarTgzTbzFunc(t, testReadTarTgzTbzSeq2)
}

// testReadTarTgzTbzFunc is the common code for testing reading a tar archive.
//
// f is a handler that reads from the reader r
// and checks with the wanted result files.
//
// It may use t.Run to create subtests for each tar archive.
func testReadTarTgzTbzFunc(
	t *testing.T,
	f func(t *testing.T, r filesys.Reader, files []tarFileNameBody),
) {
	for _, entry := range testFileEntries {
		filename := entry.Name()
		ext := filepath.Ext(filename)
		if ext == ".gz" || ext == ".bz2" {
			ext = filepath.Ext(filename[:len(filename)-len(ext)])
		}
		if ext != ".tar" && ext != ".tgz" && ext != ".tbz" {
			continue
		}
		t.Run(fmt.Sprintf("file=%+q", filename), func(t *testing.T) {
			name := filepath.Join(TestDataDir, filename)
			r, err := local.Read(name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}(r)
			files, err := lazyLoadTarFile(name)
			if err != nil {
				t.Fatal("load tar file -", err)
			}
			f(t, r, files)
		})
	}
}

// testReadTarTgzTbz is a subprocess of TestRead_TarTgzTbz
// that tests reading a tar archive.
//
// It may use t.Fatal and t.Fatalf to stop the test.
func testReadTarTgzTbz(
	t *testing.T,
	r filesys.Reader,
	files []tarFileNameBody,
) {
	for i := 0; ; i++ {
		hdr, err := r.TarNext()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if i != len(files) {
					t.Errorf("tar header number: %d != %d, but got EOF",
						i, len(files))
				}
				return // end of archive
			}
			t.Fatalf("read No.%d tar header - %v", i, err)
		}
		testTarFileReadFromR(t, r, files, i, hdr)
	}
}

// testReadTarTgzTbzSeq is a subprocess of TestRead_TarTgzTbz_Seq
// that tests reading a tar archive with the iterator method IterTarFiles.
//
// It may use t.Fatal to stop the test.
func testReadTarTgzTbzSeq(
	t *testing.T,
	r filesys.Reader,
	files []tarFileNameBody,
) {
	outErr := errors.New("init error") // initialize as a non-nil error
	seq, err := r.IterTarFiles(&outErr, false)
	if err != nil {
		t.Fatal("create iterator -", err)
	} else if seq == nil {
		t.Fatal("got nil iterator")
	}
	var i int
	for hdr := range seq {
		testTarFileReadFromR(t, r, files, i, hdr)
		i++
	}
	if outErr != nil {
		t.Error("iteration ended with", outErr)
	} else if i != len(files) {
		t.Errorf("tar header number: %d != %d, but iteration has ended",
			i, len(files))
	}
	// Test whether the iterator is single-use.
	prevErr := outErr
	for hdr := range seq {
		if hdr != nil {
			t.Errorf("not single-use iterator; got %q", hdr.Name)
		} else {
			t.Error("not single-use iterator; got <nil>")
		}
		break
	}
	if unequal.ErrorUnwrapAuto(outErr, prevErr) {
		t.Errorf("output error changed from %v to %v", prevErr, outErr)
	}
}

// testReadTarTgzTbzSeq2 is a subprocess of TestRead_TarTgzTbz_Seq2
// that tests reading a tar archive with the iterator method IterIndexTarFiles.
//
// It may use t.Fatal to stop the test.
func testReadTarTgzTbzSeq2(
	t *testing.T,
	r filesys.Reader,
	files []tarFileNameBody,
) {
	outErr := errors.New("init error") // initialize as a non-nil error
	seq2, err := r.IterIndexTarFiles(&outErr, false)
	if err != nil {
		t.Fatal("create iterator -", err)
	} else if seq2 == nil {
		t.Fatal("got nil iterator")
	}
	var ctr int
	for i, hdr := range seq2 {
		if i != ctr {
			t.Errorf("got index %d; want %d", i, ctr)
		}
		testTarFileReadFromR(t, r, files, ctr, hdr)
		ctr++
	}
	if outErr != nil {
		t.Error("iteration ended with", outErr)
	} else if ctr != len(files) {
		t.Errorf("tar header number: %d != %d, but iteration has ended",
			ctr, len(files))
	}
	// Test whether the iterator is single-use.
	prevErr := outErr
	for i, hdr := range seq2 {
		if hdr != nil {
			t.Errorf("not single-use iterator; got %d, %q", i, hdr.Name)
		} else {
			t.Errorf("not single-use iterator; got %d, <nil>", i)
		}
		break
	}
	if unequal.ErrorUnwrapAuto(outErr, prevErr) {
		t.Errorf("output error changed from %v to %v", prevErr, outErr)
	}
}

// testTarFileReadFromR tests the i-th tar file read from the reader r.
//
// It may use t.Fatal to stop the test.
func testTarFileReadFromR(
	t *testing.T,
	r filesys.Reader,
	files []tarFileNameBody,
	i int,
	hdr *tar.Header,
) {
	switch {
	case i >= len(files):
		t.Fatal("tar headers more than", len(files))
	case hdr == nil:
		t.Errorf("No.%d got nil tar header", i)
		return
	case hdr.Name != files[i].name:
		t.Errorf("No.%d tar header name unequal - got %q; want %q",
			i, hdr.Name, files[i].name)
	}
	if TarHeaderIsDir(hdr) {
		_, err := r.Read([]byte{})
		if !errors.Is(err, filesys.ErrIsDir) {
			t.Errorf("No.%d tar read file body - got %v; want %v",
				i, err, filesys.ErrIsDir)
		}
	} else {
		err := iotest.TestReader(r, files[i].body)
		if err != nil {
			t.Errorf("No.%d tar test read - %v", i, err)
		}
	}
}

func TestRead_Zip(t *testing.T) {
	testReadZipFunc(t, testReadZip)
}

func TestRead_Zip_Seq(t *testing.T) {
	testReadZipFunc(t, testReadZipSeqMain)
}

func TestRead_Zip_Seq2(t *testing.T) {
	testReadZipFunc(t, testReadZipSeq2Main)
}

// testReadZipFunc is the common code for testing reading a ZIP archive.
//
// f is a handler that reads from the reader r
// and checks with the wanted result fileMap.
//
// It may use t.Run to create subtests for each ZIP archive.
func testReadZipFunc(
	t *testing.T,
	f func(t *testing.T, r filesys.Reader, fileMap map[string]*zipHeaderBody),
) {
	for _, entry := range testFileEntries {
		filename := entry.Name()
		if filepath.Ext(filename) != ".zip" {
			continue
		}
		t.Run(fmt.Sprintf("file=%+q", filename), func(t *testing.T) {
			name := filepath.Join(TestDataDir, filename)
			r, err := local.Read(name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}(r)
			fileMap, err := lazyLoadZipFile(name)
			if err != nil {
				t.Fatal("load zip file -", err)
			}
			f(t, r, fileMap)
		})
	}
}

// testReadZip is a subprocess of TestRead_Zip
// that tests reading a ZIP archive.
//
// It may use t.Fatal to stop the test.
//
// It may use t.Run to create subtests for each file in the ZIP archive.
func testReadZip(
	t *testing.T,
	r filesys.Reader,
	fileMap map[string]*zipHeaderBody,
) {
	zipFiles, err := r.ZipFiles()
	if err != nil {
		t.Fatal("ZipFiles -", err)
	} else if len(zipFiles) != len(fileMap) {
		t.Errorf("got %d zip files; want %d", len(zipFiles), len(fileMap))
	}
	for i, file := range zipFiles {
		testZipFileReadFromR(t, "", "", fileMap, i, file)
	}
}

// testReadZipSeqMain is a subprocess of TestRead_Zip_Seq
// that tests reading a ZIP archive with the iterator method IterZipFiles.
//
// It may use t.Fatal to stop the test.
//
// It may use t.Run to create subtests for each file in the ZIP archive.
func testReadZipSeqMain(
	t *testing.T,
	r filesys.Reader,
	fileMap map[string]*zipHeaderBody,
) {
	seq, err := r.IterZipFiles()
	if err != nil {
		t.Fatal("create iterator -", err)
	} else if seq == nil {
		t.Fatal("got nil iterator")
	}
	testReadZipSeqSub(t, fileMap, seq, false)
	// Rewind the iterator and test it again.
	testReadZipSeqSub(t, fileMap, seq, true)
}

// testReadZipSeqSub is a subprocess of testReadZipSeqMain
// that tests the iterator with a single call.
//
// It may use t.Run to create subtests for each file in the ZIP archive.
func testReadZipSeqSub(
	t *testing.T,
	fileMap map[string]*zipHeaderBody,
	seq iter.Seq[*zip.File],
	isRewound bool,
) {
	var logPrefix, subtestNameSuffix string
	if isRewound {
		logPrefix, subtestNameSuffix = "rewind - ", "&rewind"
	}
	var i int
	for file := range seq {
		testZipFileReadFromR(t, logPrefix, subtestNameSuffix, fileMap, i, file)
		i++
	}
	if i != len(fileMap) {
		t.Errorf("%szip file number: %d != %d, but iteration has ended",
			logPrefix, i, len(fileMap))
	}
}

// testReadZipSeq2Main is a subprocess of TestRead_Zip_Seq2
// that tests reading a ZIP archive with the iterator method IterIndexZipFiles.
//
// It may use t.Fatal to stop the test.
//
// It may use t.Run to create subtests for each file in the ZIP archive.
func testReadZipSeq2Main(
	t *testing.T,
	r filesys.Reader,
	fileMap map[string]*zipHeaderBody,
) {
	seq2, err := r.IterIndexZipFiles()
	if err != nil {
		t.Fatal("create iterator -", err)
	} else if seq2 == nil {
		t.Fatal("got nil iterator")
	}
	testReadZipSeq2Sub(t, fileMap, seq2, false)
	// Rewind the iterator and test it again.
	testReadZipSeq2Sub(t, fileMap, seq2, true)
}

// testReadZipSeq2Sub is a subprocess of testReadZipSeq2Main
// that tests the iterator with a single call.
//
// It may use t.Run to create subtests for each file in the ZIP archive.
func testReadZipSeq2Sub(
	t *testing.T,
	fileMap map[string]*zipHeaderBody,
	seq2 iter.Seq2[int, *zip.File],
	isRewound bool,
) {
	var logPrefix, subtestNameSuffix string
	if isRewound {
		logPrefix, subtestNameSuffix = "rewind - ", "&rewind"
	}
	var ctr int
	for i, file := range seq2 {
		if i != ctr {
			t.Errorf("%sgot index %d; want %d", logPrefix, i, ctr)
		}
		testZipFileReadFromR(
			t, logPrefix, subtestNameSuffix, fileMap, ctr, file)
		ctr++
	}
	if ctr != len(fileMap) {
		t.Errorf("%szip file number: %d != %d, but iteration has ended",
			logPrefix, ctr, len(fileMap))
	}
}

// testZipFileReadFromR tests the i-th ZIP file read from the filesys.Reader.
//
// It may use t.Run to create subtests for each file in the ZIP archive.
func testZipFileReadFromR(
	t *testing.T,
	logPrefix string,
	subtestNameSuffix string,
	fileMap map[string]*zipHeaderBody,
	i int,
	file *zip.File,
) {
	if file == nil {
		t.Errorf("%sNo.%d zip file is nil", logPrefix, i)
		return
	}
	t.Run(
		fmt.Sprintf("zipFile=%+q%s", file.Name, subtestNameSuffix),
		func(t *testing.T) {
			hb, ok := fileMap[file.Name]
			if !ok {
				t.Fatalf("unknown zip file %q", file.Name)
			}
			isDir := hb.header.FileInfo().IsDir()
			if d := file.FileInfo().IsDir(); d != isDir {
				t.Errorf("got IsDir %t; want %t", d, isDir)
			}
			if isDir {
				return
			}
			rc, err := file.Open()
			if err != nil {
				t.Fatal("open -", err)
			}
			defer func(rc io.ReadCloser) {
				if err := rc.Close(); err != nil {
					t.Error("close -", err)
				}
			}(rc)
			data, err := io.ReadAll(rc)
			if err != nil {
				t.Error("read -", err)
			} else if !bytes.Equal(data, hb.body) {
				t.Errorf(
					"file contents - got (len: %d)\n%s\nwant (len: %d)\n%s",
					len(data),
					data,
					len(hb.body),
					hb.body,
				)
			}
		},
	)
}

func TestRead_Offset(t *testing.T) {
	name := filepath.Join(TestDataDir, "file1.txt")
	data, err := lazyLoadTestData(name)
	if err != nil {
		t.Fatal("read file -", err)
	}
	size := int64(len(data))

	for offset := -size; offset <= size; offset++ {
		var pos int64
		if offset > 0 {
			pos = offset
		} else if offset < 0 {
			pos = size + offset
		} else {
			continue
		}
		t.Run(fmt.Sprintf("offset=%d", offset), func(t *testing.T) {
			r, err := local.Read(name, &filesys.ReadOptions{
				Offset: offset,
				Raw:    true,
			})
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}(r)
			err = iotest.TestReader(r, data[pos:])
			if err != nil {
				t.Error("test read -", err)
			}
		})
	}

	for _, offset := range []int64{
		math.MinInt64,
		-size - 1,
		size + 1,
		math.MaxInt64,
	} {
		t.Run(fmt.Sprintf("offset=%d", offset), func(t *testing.T) {
			r, err := local.Read(name, &filesys.ReadOptions{
				Offset: offset,
				Raw:    true,
			})
			if err == nil {
				_ = r.Close() // ignore error
				t.Error("create - no error but offset is out of range")
			} else if !strings.HasSuffix(err.Error(), fmt.Sprintf(
				"option Offset (%d) is out of range; file size: %d",
				offset,
				size,
			)) {
				t.Error("create -", err)
			}
		})
	}
}

// TarHeaderIsDir reports whether the tar header represents a directory.
func TarHeaderIsDir(hdr *tar.Header) bool {
	return hdr != nil &&
		(hdr.Typeflag == '\x00' &&
			len(hdr.Name) > 0 &&
			hdr.Name[len(hdr.Name)-1] == '/' ||
			hdr.FileInfo().IsDir())
}
