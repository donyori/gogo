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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/iotest"

	"github.com/donyori/gogo/filesys"
	"github.com/donyori/gogo/filesys/local"
)

const testDataDir = "testdata"

var fileEntries []fs.DirEntry

func init() {
	entries, err := os.ReadDir(testDataDir)
	if err != nil {
		panic(err)
	}
	fileEntries = make([]fs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if entry != nil && !entry.IsDir() {
			fileEntries = append(fileEntries, entry)
		}
	}
}

func TestRead_Raw(t *testing.T) {
	for _, entry := range fileEntries {
		filename := entry.Name()
		t.Run(fmt.Sprintf("file=%q", filename), func(t *testing.T) {
			name := filepath.Join(testDataDir, filename)
			r, err := local.Read(name, &filesys.ReadOptions{Raw: true})
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func() {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}()
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
	for _, entry := range fileEntries {
		filename := entry.Name()
		if ext := filepath.Ext(filename); ext == ".gz" || ext == ".tar" || ext == ".tgz" {
			continue
		}
		t.Run(fmt.Sprintf("file=%q", filename), func(t *testing.T) {
			name := filepath.Join(testDataDir, filename)
			r, err := local.Read(name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func() {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}()
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
	for _, entry := range fileEntries {
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
			defer func() {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}()
			data, err := lazyLoadTestData(name)
			if err != nil {
				t.Fatal("read file -", err)
			}
			gr, err := gzip.NewReader(bytes.NewReader(data))
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

func TestRead_Tar_Tgz(t *testing.T) {
	for _, entry := range fileEntries {
		filename := entry.Name()
		ext := filepath.Ext(filename)
		if ext == ".gz" {
			ext = filepath.Ext(filename[:len(filename)-3])
		}
		if ext != ".tar" && ext != ".tgz" {
			continue
		}
		t.Run(fmt.Sprintf("file=%q", filename), func(t *testing.T) {
			name := filepath.Join(testDataDir, filename)
			r, err := local.Read(name, nil)
			if err != nil {
				t.Fatal("create -", err)
			}
			defer func() {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}()
			list, err := loadTarFile(name)
			if err != nil {
				t.Fatal("load tar file -", err)
			}
			n := len(list)
			for i := 0; ; i++ {
				hdr, err := r.TarNext()
				if err != nil {
					if errors.Is(err, io.EOF) {
						if i != n {
							t.Errorf("tar header number: %d != %d, but got EOF", i, n)
						}
						break // end of archive
					}
					t.Fatalf("read No.%d tar header - %v", i, err)
				}
				if i >= n {
					t.Fatal("tar headers more than", n)
				}
				if hdr.Name != list[i].name {
					t.Errorf("No.%d tar header name unequal - got %s; want %s", i, hdr.Name, list[i].name)
				}
				err = iotest.TestReader(r, list[i].body)
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
			defer func() {
				if err := r.Close(); err != nil {
					t.Error("close -", err)
				}
			}()
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

var (
	testDataMap      map[string][]byte
	loadTestDataLock sync.Mutex
)

// lazyLoadTestData loads a file with specified name.
//
// It stores the file content in the memory the first time reading that file.
// Subsequent reads will get the file content from the memory instead of
// reading the file again.
// Therefore, all modifications to the file after the first read cannot
// take effect on this function.
func lazyLoadTestData(name string) ([]byte, error) {
	loadTestDataLock.Lock()
	defer loadTestDataLock.Unlock()
	var data []byte
	if testDataMap != nil {
		data = testDataMap[name]
		if data != nil {
			return data, nil
		}
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	if testDataMap == nil {
		testDataMap = make(map[string][]byte, len(fileEntries))
	}
	testDataMap[name] = data
	return data, nil
}

// loadTarFile loads a ".tar", ".tgz", or ".tar.gz" file.
//
// It returns a list of (filename, file body) pairs.
// It also returns any error encountered.
//
// Caller should guarantee that the file name has
// a suffix ".tar", ".tgz", or ".tar.gz".
func loadTarFile(name string) ([]struct {
	name string
	body []byte
}, error) {
	data, err := lazyLoadTestData(name)
	if err != nil {
		return nil, err
	}
	var r io.Reader = bytes.NewReader(data)
	ext := filepath.Ext(name)
	if ext == ".gz" || ext == ".tgz" {
		gr, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer func(gr *gzip.Reader) {
			_ = gr.Close() // ignore error
		}(gr)
		r = gr
	}
	tr := tar.NewReader(r)
	var list []struct {
		name string
		body []byte
	}
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break // end of archive
		}
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		list = append(list, struct {
			name string
			body []byte
		}{hdr.Name, data})
	}
	return list, nil
}
