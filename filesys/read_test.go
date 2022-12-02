// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/donyori/gogo/filesys"
)

func TestRead_NotCloseFile(t *testing.T) {
	const name = "file1.txt"
	file, err := testFS.Open(name)
	if err != nil {
		t.Fatal("open file -", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Error("close file -", err)
		}
	}()
	r, err := filesys.Read(file, &filesys.ReadOptions{Raw: true}, false)
	if err != nil {
		t.Fatal("create reader -", err)
	}
	halfSize := int64(len(testFS[name].Data) / 2)
	hr := io.LimitReader(r, halfSize)
	err = iotest.TestReader(hr, testFS[name].Data[:halfSize])
	if err != nil {
		t.Error("test read a half -", err)
	}
	err = r.Close()
	if err != nil {
		t.Fatal("close reader -", err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatal("read all from rest part -", err)
	}
	want := testFS[name].Data[halfSize:]
	if !bytes.Equal(data, want) {
		t.Errorf("read all from rest part - got %s; want %s", data, want)
	}
}

func TestRead_NotCloseFile_ErrorOnCreate(t *testing.T) {
	const name = "file1.txt"
	file, err := testFS.Open(name)
	if err != nil {
		t.Fatal("open file -", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Error("close file -", err)
		}
	}()
	r, err := filesys.Read(file, &filesys.ReadOptions{Offset: math.MaxInt64, Raw: true}, false)
	if err == nil {
		_ = r.Close()
		t.Fatal("create reader - no error but offset is out of range")
	}
	if !strings.HasSuffix(err.Error(), fmt.Sprintf("option Offset is out of range, file size: %d, Offset: %d", len(testFS[name].Data), math.MaxInt64)) {
		t.Fatal("create reader -", err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatal("read all from file -", err)
	}
	want := testFS[name].Data
	if !bytes.Equal(data, want) {
		t.Errorf("read all from file - got %s; want %s", data, want)
	}
}

func TestReadFromFS_Raw(t *testing.T) {
	for _, name := range testFSFilenames {
		t.Run(fmt.Sprintf("file=%q", name), func(t *testing.T) {
			r, err := filesys.ReadFromFS(testFS, name, &filesys.ReadOptions{Raw: true})
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func() {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}()
			err = iotest.TestReader(r, testFS[name].Data)
			if err != nil {
				t.Error("test read -", err)
			}
		})
	}
}

func TestReadFromFS_Basic(t *testing.T) {
	for _, name := range testFSBasicFilenames {
		t.Run(fmt.Sprintf("file=%q", name), func(t *testing.T) {
			r, err := filesys.ReadFromFS(testFS, name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func() {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}()
			err = iotest.TestReader(r, testFS[name].Data)
			if err != nil {
				t.Error("test read -", err)
			}
		})
	}
}

func TestReadFromFS_Gz(t *testing.T) {
	for _, name := range testFSGzFilenames {
		t.Run(fmt.Sprintf("file=%q", name), func(t *testing.T) {
			r, err := filesys.ReadFromFS(testFS, name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func() {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}()
			gr, err := gzip.NewReader(bytes.NewReader(testFS[name].Data))
			if err != nil {
				t.Fatal("create gzip reader -", err)
			}
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

func TestReadFromFS_Tar_Tgz(t *testing.T) {
	for _, name := range append(testFSTarFilenames, testFSTgzFilenames...) {
		t.Run(fmt.Sprintf("file=%q", name), func(t *testing.T) {
			r, err := filesys.ReadFromFS(testFS, name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func() {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}()
			for i := 0; ; i++ {
				hdr, err := r.TarNext()
				if err != nil {
					if errors.Is(err, io.EOF) {
						if i != len(testFSTarFiles) {
							t.Errorf("tar header number: %d != %d, but got EOF", i, len(testFSTarFiles))
						}
						break // end of archive
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
	const name = "file1.txt"
	fileData := testFS[name].Data
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
			r, err := filesys.ReadFromFS(testFS, name, &filesys.ReadOptions{Offset: offset, Raw: true})
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func() {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}()
			err = iotest.TestReader(r, fileData[pos:])
			if err != nil {
				t.Error("test read -", err)
			}
		})
	}

	for _, offset := range []int64{math.MinInt64, -size - 1, size + 1, math.MaxInt64} {
		t.Run(fmt.Sprintf("offset=%d", offset), func(t *testing.T) {
			r, err := filesys.ReadFromFS(testFS, name, &filesys.ReadOptions{Offset: offset, Raw: true})
			if err == nil {
				_ = r.Close() // ignore error
				t.Fatal("create - no error but offset is out of range")
			}
			if !strings.HasSuffix(err.Error(), fmt.Sprintf("option Offset is out of range, file size: %d, Offset: %d", size, offset)) {
				t.Error("create -", err)
			}
		})
	}
}
