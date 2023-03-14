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
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/donyori/gogo/filesys"
)

func TestRead_NotCloseFile(t *testing.T) {
	const Name = "file1.txt"
	file, err := testFS.Open(Name)
	if err != nil {
		t.Fatal("open file -", err)
	}
	defer func(f fs.File) {
		if err := f.Close(); err != nil {
			t.Error("close file -", err)
		}
	}(file)
	r, err := filesys.Read(file, &filesys.ReadOptions{Raw: true}, false)
	if err != nil {
		t.Fatal("create reader -", err)
	}
	halfSize := int64(len(testFS[Name].Data) / 2)
	hr := io.LimitReader(r, halfSize)
	err = iotest.TestReader(hr, testFS[Name].Data[:halfSize])
	if err != nil {
		t.Error("test read a half -", err)
	}
	var buffered []byte
	if n := r.Buffered(); n > 0 {
		peek, err := r.Peek(n)
		if err != nil {
			t.Fatal("peek buffered data -", err)
		}
		buffered = make([]byte, len(peek), len(peek)+int(halfSize))
		copy(buffered, peek)
	}
	err = r.Close()
	if err != nil {
		t.Fatal("close reader -", err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatal("read all from rest part -", err)
	}
	data = append(buffered, data...)
	want := testFS[Name].Data[halfSize:]
	if !bytes.Equal(data, want) {
		t.Errorf("read all from rest part - got %s; want %s", data, want)
	}
}

func TestRead_NotCloseFile_ErrorOnCreate(t *testing.T) {
	const Name = "file1.txt"
	file, err := testFS.Open(Name)
	if err != nil {
		t.Fatal("open file -", err)
	}
	defer func(f fs.File) {
		if err := f.Close(); err != nil {
			t.Error("close file -", err)
		}
	}(file)
	r, err := filesys.Read(file, &filesys.ReadOptions{Offset: math.MaxInt64, Raw: true}, false)
	if err == nil {
		_ = r.Close()
		t.Fatal("create reader - no error but offset is out of range")
	}
	if !strings.HasSuffix(err.Error(), fmt.Sprintf(
		"option Offset (%d) is out of range; file size: %d",
		math.MaxInt64,
		len(testFS[Name].Data),
	)) {
		t.Fatal("create reader -", err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatal("read all from file -", err)
	}
	want := testFS[Name].Data
	if !bytes.Equal(data, want) {
		t.Errorf("read all from file - got %s; want %s", data, want)
	}
}

func TestReadFromFS_Raw(t *testing.T) {
	for _, name := range testFSFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			testReadFromTestFS(t, name, testFS[name].Data, &filesys.ReadOptions{Raw: true})
		})
	}
}

func TestReadFromFS_Basic(t *testing.T) {
	for _, name := range testFSBasicFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			testReadFromTestFS(t, name, testFS[name].Data, nil)
		})
	}
}

func TestReadFromFS_Gz(t *testing.T) {
	for _, name := range testFSGzFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			dataFilename := name[:len(name)-3]
			mapFile := testFS[dataFilename]
			if mapFile == nil {
				t.Fatalf("file %q does not exist", dataFilename)
			}
			testReadFromTestFS(t, name, mapFile.Data, nil)
		})
	}
}

func TestReadFromFS_TarTgz(t *testing.T) {
	for _, name := range append(testFSTarFilenames, testFSTgzFilenames...) {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			r, err := filesys.ReadFromFS(testFS, name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}(r)

			for i := 0; ; i++ {
				hdr, err := r.TarNext()
				if err != nil {
					if errors.Is(err, io.EOF) {
						if i != len(testFSTarFiles) {
							t.Errorf("tar header number: %d != %d, but got EOF", i, len(testFSTarFiles))
						}
						return // end of archive
					}
					t.Fatalf("read No.%d tar header - %v", i, err)
				}
				if i >= len(testFSTarFiles) {
					t.Fatal("tar headers more than", len(testFSTarFiles))
				}
				if hdr.Name != testFSTarFiles[i].name {
					t.Errorf("No.%d tar header name unequal - got %s; want %s", i, hdr.Name, testFSTarFiles[i].name)
				}
				if filesys.TarHeaderIsDir(hdr) {
					_, err = r.Read([]byte{})
					if !errors.Is(err, filesys.ErrIsDir) {
						t.Errorf("No.%d tar read file body - got %v; want %v", i, err, filesys.ErrIsDir)
					}
				} else {
					err = iotest.TestReader(r, []byte(testFSTarFiles[i].body))
					if err != nil {
						t.Errorf("No.%d tar test read - %v", i, err)
					}
				}
			}
		})
	}
}

func TestReadFromFS_Zip_Open(t *testing.T) {
	for _, name := range testFSZipFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			r, err := filesys.ReadFromFS(testFS, name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}(r)
			testZipOpen(t, r)
		})
	}
}

func TestReadFromFS_Zip_Files(t *testing.T) {
	for _, name := range testFSZipFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			r, err := filesys.ReadFromFS(testFS, name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}(r)
			testZipFiles(t, r)
		})
	}
}

func TestReadFromFS_Zip_Comment(t *testing.T) {
	for _, name := range testFSZipFilenames {
		t.Run(fmt.Sprintf("file=%+q", name), func(t *testing.T) {
			r, err := filesys.ReadFromFS(testFS, name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}(r)

			comment, err := r.ZipComment()
			if err != nil {
				t.Fatal("zip comment -", err)
			}
			if comment != testFSZipComment {
				t.Errorf("got %q; want %q", comment, testFSZipComment)
			}
		})
	}
}

func TestReadFromFS_Zip_Offset_Open(t *testing.T) {
	r, err := filesys.ReadFromFS(
		testFS,
		testFSZipOffsetName,
		&filesys.ReadOptions{Offset: testFSZipOffset},
	)
	if err != nil {
		t.Fatal("create -", err)
	}
	defer func(r filesys.Reader) {
		if err := r.Close(); err != nil {
			t.Error("close -", err)
		}
	}(r)
	testZipOpen(t, r)
}

func TestReadFromFS_Zip_Offset_Files(t *testing.T) {
	r, err := filesys.ReadFromFS(
		testFS,
		testFSZipOffsetName,
		&filesys.ReadOptions{Offset: testFSZipOffset},
	)
	if err != nil {
		t.Fatal("create -", err)
	}
	defer func(r filesys.Reader) {
		if err := r.Close(); err != nil {
			t.Error("close -", err)
		}
	}(r)
	testZipFiles(t, r)
}

func TestReadFromFS_Zip_Offset_Comment(t *testing.T) {
	r, err := filesys.ReadFromFS(
		testFS,
		testFSZipOffsetName,
		&filesys.ReadOptions{Offset: testFSZipOffset},
	)
	if err != nil {
		t.Fatal("create -", err)
	}
	defer func(r filesys.Reader) {
		if err := r.Close(); err != nil {
			t.Error("close -", err)
		}
	}(r)

	comment, err := r.ZipComment()
	if err != nil {
		t.Fatal("zip comment -", err)
	}
	if comment != testFSZipComment {
		t.Errorf("got %q; want %q", comment, testFSZipComment)
	}
}

func TestReadFromFS_Offset(t *testing.T) {
	const Name = "file1.txt"
	fileData := testFS[Name].Data
	size := int64(len(fileData))

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
			testReadFromTestFS(t, Name, fileData[pos:], &filesys.ReadOptions{Offset: offset, Raw: true})
		})
	}

	for _, offset := range []int64{math.MinInt64, -size - 1, size + 1, math.MaxInt64} {
		t.Run(fmt.Sprintf("offset=%d", offset), func(t *testing.T) {
			r, err := filesys.ReadFromFS(testFS, Name, &filesys.ReadOptions{Offset: offset, Raw: true})
			if err == nil {
				_ = r.Close() // ignore error
				t.Fatal("create - no error but offset is out of range")
			}
			if !strings.HasSuffix(err.Error(), fmt.Sprintf(
				"option Offset (%d) is out of range; file size: %d",
				offset,
				size,
			)) {
				t.Error("create -", err)
			}
		})
	}
}

func TestReadFromFS_AfterClose(t *testing.T) {
	const RegFile = "file1.txt"
	const TarFile = "tar file.tar"
	const ZipFile = "zip basic.zip"
	testCases := []struct {
		methodName string
		filename   string
		f          func(t *testing.T, r filesys.Reader) error
		wantErr    error
	}{
		{
			"Close",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				return r.Close()
			},
			nil,
		},
		{
			"Closed",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				if !r.Closed() {
					t.Error("r.Closed - got false; want true")
				}
				return nil
			},
			nil,
		},
		{
			"Read",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.Read(nil)
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"ReadByte",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.ReadByte()
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"UnreadByte",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				return r.UnreadByte()
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"ReadRune",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				_, _, err := r.ReadRune()
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"UnreadRune",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				return r.UnreadRune()
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"WriteTo",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.WriteTo(nil)
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"ReadLine",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				_, _, err := r.ReadLine()
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"ReadEntireLine",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.ReadEntireLine()
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"WriteLineTo",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.WriteLineTo(nil)
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"Peek",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.Peek(0)
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"Discard",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.Discard(0)
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"TarNext-notTar",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.TarNext()
				return err
			},
			filesys.ErrNotTar,
		},
		{
			"TarNext-isTar",
			TarFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.TarNext()
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"ZipOpen-notZip",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.ZipOpen("")
				return err
			},
			filesys.ErrNotZip,
		},
		{
			"ZipOpen-isZip",
			ZipFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.ZipOpen("")
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"ZipFiles-notZip",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.ZipFiles()
				return err
			},
			filesys.ErrNotZip,
		},
		{
			"ZipFiles-isZip",
			ZipFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.ZipFiles()
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"ZipComment-notZip",
			RegFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.ZipComment()
				return err
			},
			filesys.ErrNotZip,
		},
		{
			"ZipComment-isZip",
			ZipFile,
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.ZipComment()
				return err
			},
			filesys.ErrFileReaderClosed,
		},
	}

	for _, tc := range testCases {
		t.Run("method="+tc.methodName, func(t *testing.T) {
			r, err := filesys.ReadFromFS(testFS, tc.filename, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			err = r.Close()
			if err != nil {
				t.Fatal("close -", err)
			}
			err = tc.f(t, r)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got error %v; want %v", err, tc.wantErr)
			}
		})
	}
}

// testReadFromTestFS reads the file with specified name from testFS
// using ReadFromFS and tests the reader using iotest.TestReader.
//
// want is the expected data read from the file.
func testReadFromTestFS(t *testing.T, name string, want []byte, opts *filesys.ReadOptions) {
	r, err := filesys.ReadFromFS(testFS, name, opts)
	if err != nil {
		t.Error("create -", err)
		return
	}
	defer func(r filesys.Reader) {
		if err := r.Close(); err != nil {
			t.Error("close -", err)
		}
	}(r)
	err = iotest.TestReader(r, want)
	if err != nil {
		t.Error("test read -", err)
	}
}

// testZipOpen tests filesys.Reader.ZipOpen.
func testZipOpen(t *testing.T, r filesys.Reader) {
	dirSet := make(map[string]bool, len(testFSZipFileNameBodyMap))
	for zipFilename, body := range testFSZipFileNameBodyMap {
		for i := 1; i < len(zipFilename)-1; i++ {
			if zipFilename[i-1] != '/' && zipFilename[i] == '/' {
				dirSet[zipFilename[:i]] = true
			}
		}
		if len(zipFilename) > 1 &&
			zipFilename[len(zipFilename)-2] != '/' &&
			zipFilename[len(zipFilename)-1] == '/' {
			dirSet[zipFilename[:len(zipFilename)-1]] = true
			continue
		}
		t.Run(fmt.Sprintf("zipFile=%+q", zipFilename), func(t *testing.T) {
			file, err := r.ZipOpen(zipFilename)
			if err != nil {
				t.Fatal("open -", err)
			}
			defer func(f fs.File) {
				if err := f.Close(); err != nil {
					t.Error("close -", err)
				}
			}(file)
			info, err := file.Stat()
			if err != nil {
				t.Fatal("stat -", err)
			}
			if d := info.IsDir(); d {
				t.Errorf("got IsDir %t; want false", d)
			}
			data, err := io.ReadAll(file)
			if err != nil {
				t.Fatal("read -", err)
			}
			if string(data) != body {
				t.Errorf(
					"file contents - got (len: %d)\n%s\nwant (len: %d)\n%s",
					len(data),
					data,
					len(body),
					body,
				)
			}
		})
	}

	for dir := range dirSet {
		t.Run(fmt.Sprintf("zipFile=%+q", dir), func(t *testing.T) {
			file, err := r.ZipOpen(dir)
			if err != nil {
				t.Fatal("open -", err)
			}
			defer func(f fs.File) {
				if err := f.Close(); err != nil {
					t.Error("close -", err)
				}
			}(file)
			info, err := file.Stat()
			if err != nil {
				t.Fatal("stat -", err)
			}
			if d := info.IsDir(); !d {
				t.Errorf("got IsDir %t; want true", d)
			}
		})
	}
}

// testZipFiles tests filesys.Reader.ZipFiles.
func testZipFiles(t *testing.T, r filesys.Reader) {
	files, err := r.ZipFiles()
	if err != nil {
		t.Error("ZipFiles -", err)
		return
	} else if len(files) != len(testFSZipFileNameBodyMap) {
		t.Errorf("got %d zip files; want %d",
			len(files), len(testFSZipFileNameBodyMap))
	}

	for i, file := range files {
		if file == nil {
			t.Errorf("No.%d zip file is nil", i)
			continue
		}
		t.Run(fmt.Sprintf("zipFile=%+q", file.Name), func(t *testing.T) {
			body, ok := testFSZipFileNameBodyMap[file.Name]
			if !ok {
				t.Fatalf("unknown zip file %q", file.Name)
			}
			isDir := len(file.Name) > 0 && file.Name[len(file.Name)-1] == '/'
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
				t.Fatal("read -", err)
			}
			if string(data) != body {
				t.Errorf(
					"file contents - got (len: %d)\n%s\nwant (len: %d)\n%s",
					len(data),
					data,
					len(body),
					body,
				)
			}
		})
	}
}
