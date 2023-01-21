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
		"option Offset (%d) is out of range, file size: %d",
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
				err = iotest.TestReader(r, []byte(testFSTarFiles[i].body))
				if err != nil {
					t.Errorf("No.%d tar test read - %v", i, err)
				}
			}
		})
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
				"option Offset (%d) is out of range, file size: %d",
				offset,
				size,
			)) {
				t.Error("create -", err)
			}
		})
	}
}

func TestReadFromFS_AfterClose(t *testing.T) {
	testCases := []struct {
		methodName string
		f          func(t *testing.T, r filesys.Reader) error
		wantErr    error
	}{
		{
			"Close",
			func(t *testing.T, r filesys.Reader) error {
				return r.Close()
			},
			nil,
		},
		{
			"Closed",
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
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.Read([]byte{})
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"ReadByte",
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.ReadByte()
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"UnreadByte",
			func(t *testing.T, r filesys.Reader) error {
				return r.UnreadByte()
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"ReadRune",
			func(t *testing.T, r filesys.Reader) error {
				_, _, err := r.ReadRune()
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"UnreadRune",
			func(t *testing.T, r filesys.Reader) error {
				return r.UnreadRune()
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"WriteTo",
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.WriteTo(io.Discard)
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"ReadLine",
			func(t *testing.T, r filesys.Reader) error {
				_, _, err := r.ReadLine()
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"ReadEntireLine",
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.ReadEntireLine()
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"WriteLineTo",
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.WriteLineTo(io.Discard)
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"Peek",
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.Peek(0)
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"Discard",
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.Discard(0)
				return err
			},
			filesys.ErrFileReaderClosed,
		},
		{
			"TarNext-notTar",
			func(t *testing.T, r filesys.Reader) error {
				_, err := r.TarNext()
				return err
			},
			filesys.ErrNotTar,
		},
	}

	for _, tc := range testCases {
		t.Run("method="+tc.methodName, func(t *testing.T) {
			const Name = "file1.txt"
			r, err := filesys.ReadFromFS(testFS, Name, nil)
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

	t.Run("method=TarNext-isTar", func(t *testing.T) {
		const Name = "tar file.tar"
		r, err := filesys.ReadFromFS(testFS, Name, nil)
		if err != nil {
			t.Fatal("create -", err)
		}
		err = r.Close()
		if err != nil {
			t.Fatal("close -", err)
		}
		_, err = r.TarNext()
		if !errors.Is(err, filesys.ErrFileReaderClosed) {
			t.Errorf("got error %v; want %v", err, filesys.ErrFileReaderClosed)
		}
	})
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
