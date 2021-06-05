// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"
	"testing/iotest"
)

func TestRead_NotCloseFile(t *testing.T) {
	const name = "testFile1.txt"
	file, err := testFs.Open(name)
	if err != nil {
		t.Fatalf("open file, %v.", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("close file, %v.", err)
		}
	}()
	r, err := Read(file, &ReadOptions{Raw: true}, false)
	if err != nil {
		t.Fatalf("create reader, %v.", err)
	}
	halfSize := int64(len(testFs[name].Data) / 2)
	hr := io.LimitReader(r, halfSize)
	err = iotest.TestReader(hr, testFs[name].Data[:halfSize])
	if err != nil {
		t.Errorf("test read a half, %v.", err)
	}
	err = r.Close()
	if err != nil {
		t.Fatalf("close reader, %v.", err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("read all from rest part, %v.", err)
	}
	wanted := testFs[name].Data[halfSize:]
	if !bytes.Equal(data, wanted) {
		t.Errorf("read all from rest part, got: %s, wanted: %s.", data, wanted)
	}
}

func TestRead_NotCloseFile_ErrorOnCreate(t *testing.T) {
	const name = "testFile1.txt"
	file, err := testFs.Open(name)
	if err != nil {
		t.Fatalf("open file, %v.", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("close file, %v.", err)
		}
	}()
	r, err := Read(file, &ReadOptions{Offset: math.MaxInt64, Raw: true}, false)
	if err == nil {
		_ = r.Close()
		t.Fatal("create reader, no error but offset is out of range.")
	}
	if !strings.HasSuffix(err.Error(), fmt.Sprintf("option Offset is out of range, file size: %d, Offset: %d", len(testFs[name].Data), math.MaxInt64)) {
		t.Fatalf("create reader, %v.", err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("read all from file, %v.", err)
	}
	wanted := testFs[name].Data
	if !bytes.Equal(data, wanted) {
		t.Errorf("read all from file, got: %s, wanted: %s.", data, wanted)
	}
}

func TestReadFromFs_Raw(t *testing.T) {
	for _, name := range testFsFilenames {
		func(name string) {
			r, err := ReadFromFs(testFs, name, &ReadOptions{Raw: true})
			if err != nil {
				t.Errorf("file: %s, create, %v.", name, err)
				return
			}
			defer func() {
				if err := r.Close(); err != nil {
					t.Errorf("file: %s, close, %v.", name, err)
				}
			}()
			err = iotest.TestReader(r, testFs[name].Data)
			if err != nil {
				t.Errorf("file: %s, test read, %v.", name, err)
			}
		}(name)
	}
}

func TestReadFromFs_Basic(t *testing.T) {
	for _, name := range testFsBasicFilenames {
		func(name string) {
			r, err := ReadFromFs(testFs, name, nil)
			if err != nil {
				t.Errorf("file: %s, create, %v.", name, err)
				return
			}
			defer func() {
				if err := r.Close(); err != nil {
					t.Errorf("file: %s, close, %v.", name, err)
				}
			}()
			err = iotest.TestReader(r, testFs[name].Data)
			if err != nil {
				t.Errorf("file: %s, test read, %v.", name, err)
			}
		}(name)
	}
}

func TestReadFromFs_Gz(t *testing.T) {
	for _, name := range testFsGzFilenames {
		func(name string) {
			r, err := ReadFromFs(testFs, name, nil)
			if err != nil {
				t.Errorf("file: %s, create, %v.", name, err)
				return
			}
			defer func() {
				if err := r.Close(); err != nil {
					t.Errorf("file: %s, close, %v.", name, err)
				}
			}()
			gr, err := gzip.NewReader(bytes.NewReader(testFs[name].Data))
			if err != nil {
				t.Errorf("file: %s, create gzip reader, %v.", name, err)
				return
			}
			wanted, err := io.ReadAll(gr)
			if err != nil {
				t.Errorf("file: %s, decompress gzip, %v.", name, err)
				return
			}
			err = iotest.TestReader(r, wanted)
			if err != nil {
				t.Errorf("file: %s, test read, %v.", name, err)
			}
		}(name)
	}
}

func TestReadFromFs_Tar_Tgz(t *testing.T) {
	for _, name := range append(testFsTarFilenames, testFsTgzFilenames...) {
		func(name string) {
			r, err := ReadFromFs(testFs, name, nil)
			if err != nil {
				t.Errorf("file: %s, create, %v.", name, err)
				return
			}
			defer func() {
				if err := r.Close(); err != nil {
					t.Errorf("file: %s, close, %v.", name, err)
				}
			}()
			for i := 0; true; i++ {
				hdr, err := r.TarNext()
				if err != nil {
					if errors.Is(err, io.EOF) {
						if i != len(testFsTarFiles) {
							t.Errorf("file: %s, tar header number: %d != %d, but got EOF.", name, i, len(testFsTarFiles))
						}
						break
					}
					t.Errorf("file: %s, read No.%d tar header, %v.", name, i, err)
					return
				}
				if i >= len(testFsTarFiles) {
					t.Errorf("file: %s, tar headers more than %d.", name, len(testFsTarFiles))
					return
				}
				if hdr.Name != testFsTarFiles[i].name {
					t.Errorf("file: %s, No.%d tar header name unequal, got: %s, wanted: %s.", name, i, hdr.Name, testFsTarFiles[i].name)
				}
				err = iotest.TestReader(r, []byte(testFsTarFiles[i].body))
				if err != nil {
					t.Errorf("file: %s, No.%d tar test read, %v.", name, i, err)
				}
			}
		}(name)
	}
}

func TestReadFromFs_Offset(t *testing.T) {
	const name = "testFile1.txt"
	fileData := testFs[name].Data
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
		func(offset, pos int64) {
			r, err := ReadFromFs(testFs, name, &ReadOptions{Offset: offset, Raw: true})
			if err != nil {
				t.Errorf("offset: %d, create, %v.", offset, err)
				return
			}
			defer func() {
				if err := r.Close(); err != nil {
					t.Errorf("offset: %d, close, %v.", offset, err)
				}
			}()
			err = iotest.TestReader(r, fileData[pos:])
			if err != nil {
				t.Errorf("offset: %d, test read, %v.", offset, err)
			}
		}(offset, pos)
	}

	for _, offset := range []int64{math.MinInt64, -size - 1, size + 1, math.MaxInt64} {
		func(offset int64) {
			r, err := ReadFromFs(testFs, name, &ReadOptions{Offset: offset, Raw: true})
			if err == nil {
				_ = r.Close()
				t.Errorf("offset: %d, create, no error but offset is out of range.", offset)
				return
			}
			if !strings.HasSuffix(err.Error(), fmt.Sprintf("option Offset is out of range, file size: %d, Offset: %d", size, offset)) {
				t.Errorf("offset: %d, create, %v.", offset, err)
			}
		}(offset)
	}
}
