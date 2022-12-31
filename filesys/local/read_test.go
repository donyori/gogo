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

package local_test

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/donyori/gogo/filesys"
	"github.com/donyori/gogo/filesys/local"
)

func TestRead_Raw(t *testing.T) {
	for _, entry := range testFileEntries {
		filename := entry.Name()
		t.Run(fmt.Sprintf("file=%q", filename), func(t *testing.T) {
			name := filepath.Join(testDataDir, filename)
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
			ext == ".tar" || ext == ".tgz" || ext == ".tbz" {
			continue
		}
		t.Run(fmt.Sprintf("file=%q", filename), func(t *testing.T) {
			name := filepath.Join(testDataDir, filename)
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
		if filepath.Ext(filename) != ".gz" || filepath.Ext(filename[:len(filename)-3]) == ".tar" {
			continue
		}
		t.Run(fmt.Sprintf("file=%q", filename), func(t *testing.T) {
			name := filepath.Join(testDataDir, filename)
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
		if filepath.Ext(filename) != ".bz2" || filepath.Ext(filename[:len(filename)-4]) == ".tar" {
			continue
		}
		t.Run(fmt.Sprintf("file=%q", filename), func(t *testing.T) {
			name := filepath.Join(testDataDir, filename)
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
	for _, entry := range testFileEntries {
		filename := entry.Name()
		ext := filepath.Ext(filename)
		if ext == ".gz" || ext == ".bz2" {
			ext = filepath.Ext(filename[:len(filename)-len(ext)])
		}
		if ext != ".tar" && ext != ".tgz" && ext != ".tbz" {
			continue
		}
		t.Run(fmt.Sprintf("file=%q", filename), func(t *testing.T) {
			name := filepath.Join(testDataDir, filename)
			r, err := local.Read(name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func(r filesys.Reader) {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}(r)
			files, err := loadTarFile(name)
			if err != nil {
				t.Fatal("load tar file -", err)
			}
			for i := 0; ; i++ {
				hdr, err := r.TarNext()
				if err != nil {
					if errors.Is(err, io.EOF) {
						if i != len(files) {
							t.Errorf("tar header number: %d != %d, but got EOF", i, len(files))
						}
						return // end of archive
					}
					t.Fatalf("read No.%d tar header - %v", i, err)
				}
				if i >= len(files) {
					t.Fatal("tar headers more than", len(files))
				}
				if hdr.Name != files[i].name {
					t.Errorf("No.%d tar header name unequal - got %s; want %s", i, hdr.Name, files[i].name)
				}
				err = iotest.TestReader(r, files[i].body)
				if err != nil {
					t.Errorf("No.%d tar test read - %v", i, err)
				}
			}
		})
	}
}

func TestRead_Offset(t *testing.T) {
	name := filepath.Join(testDataDir, "file1.txt")
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
			r, err := local.Read(name, &filesys.ReadOptions{Offset: offset, Raw: true})
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

	for _, offset := range []int64{math.MinInt64, -size - 1, size + 1, math.MaxInt64} {
		t.Run(fmt.Sprintf("offset=%d", offset), func(t *testing.T) {
			r, err := local.Read(name, &filesys.ReadOptions{Offset: offset, Raw: true})
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
