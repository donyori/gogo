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

package local

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"testing/iotest"
	"time"

	"github.com/donyori/gogo/fs"
)

func TestRead_Basic(t *testing.T) {
	dir, err := os.MkdirTemp("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	data := []byte("Some data in the temporary file.\n")
	filename := filepath.Join(dir, "basic.txt")
	err = os.WriteFile(filename, data, 0600)
	if err != nil {
		t.Fatal(err)
	}
	r, err := Read(filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	}()
	err = iotest.TestReader(r, data)
	if err != nil {
		t.Error(err)
	}
}

func TestRead_Gz(t *testing.T) {
	dir, err := os.MkdirTemp("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	data := []byte("Some data in the temporary gzip file.\n")
	filename := filepath.Join(dir, "simple.txt.gz")
	f, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	defer func() {
		if !closed {
			f.Close() // ignore error
		}
	}()
	gzw := gzip.NewWriter(f)
	defer func() {
		if !closed {
			gzw.Close() // ignore error
		}
	}()
	_, err = gzw.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	gzw.Close() // ignore error
	f.Close()   // ignore error
	closed = true

	r, err := Read(filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	}()
	err = iotest.TestReader(r, data)
	if err != nil {
		t.Error(err)
	}
}

func TestRead_Tar(t *testing.T) {
	dir, err := os.MkdirTemp("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	filename := filepath.Join(dir, "simple.tar")
	f, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	defer func() {
		if !closed {
			f.Close() // ignore error
		}
	}()
	tw := tar.NewWriter(f)
	defer func() {
		if !closed {
			tw.Close() // ignore error
		}
	}()
	files := []struct {
		name string
		body []byte
	}{
		{"file1.txt", []byte("This is file1.")},
		{"file2.txt", []byte("This is file2.")},
		{"file3.txt", []byte("This is file3.")},
	}
	for i := range files {
		err = tw.WriteHeader(&tar.Header{
			Name:    files[i].name,
			Size:    int64(len(files[i].body)),
			Mode:    0600,
			ModTime: time.Now(),
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = tw.Write(files[i].body)
		if err != nil {
			t.Fatal(err)
		}
	}
	tw.Close() // ignore error
	f.Close()  // ignore error
	closed = true

	r, err := Read(filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	}()
	for i := 0; true; i++ {
		hdr, err := r.TarNext()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if i != len(files) {
					t.Error("i:", i, "!= len(files):", len(files), "but got EOF.")
				}
				break
			}
			t.Fatal(err)
		}
		if i >= len(files) {
			t.Fatal("i:", i, ">= len(files):", len(files))
		}
		if hdr.Name != files[i].name {
			t.Errorf("hdr.Name: %q != %q.", hdr.Name, files[i].name)
		}
		err = iotest.TestReader(r, files[i].body)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestRead_TarGz_Tgz(t *testing.T) {
	testReadTarGz(t, false)
	testReadTarGz(t, true)
}

func testReadTarGz(t *testing.T, useTgz bool) {
	dir, err := os.MkdirTemp("", "gogo_test_")
	if err != nil {
		t.Error("useTgz:", useTgz, err)
		return
	}
	defer os.RemoveAll(dir) // ignore error
	basename := "simple.tar.gz"
	if useTgz {
		basename = "simple.tgz"
	}
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		t.Error("useTgz:", useTgz, err)
		return
	}
	closed := false
	defer func() {
		if !closed {
			f.Close() // ignore error
		}
	}()
	gzw := gzip.NewWriter(f)
	defer func() {
		if !closed {
			gzw.Close() // ignore error
		}
	}()
	tw := tar.NewWriter(gzw)
	defer func() {
		if !closed {
			tw.Close() // ignore error
		}
	}()
	files := []struct {
		name string
		body []byte
	}{
		{"file1.txt", []byte("This is file1.")},
		{"file2.txt", []byte("This is file2.")},
		{"file3.txt", []byte("This is file3.")},
	}
	for i := range files {
		err = tw.WriteHeader(&tar.Header{
			Name:    files[i].name,
			Size:    int64(len(files[i].body)),
			Mode:    0600,
			ModTime: time.Now(),
		})
		if err != nil {
			t.Error("useTgz:", useTgz, err)
			return
		}
		_, err = tw.Write(files[i].body)
		if err != nil {
			t.Error("useTgz:", useTgz, err)
			return
		}
	}
	tw.Close()  // ignore error
	gzw.Close() // ignore error
	f.Close()   // ignore error
	closed = true

	r, err := Read(filename, &fs.ReadOptions{BufOpen: true})
	if err != nil {
		t.Error("useTgz:", useTgz, err)
		return
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Error("useTgz:", useTgz, err)
		}
	}()
	for i := 0; true; i++ {
		hdr, err := r.TarNext()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if i != len(files) {
					t.Error("useTgz:", useTgz, "i:", i, "!= len(files):", len(files), "but got EOF.")
				}
				break
			}
			t.Error("useTgz:", useTgz, err)
			return
		}
		if i >= len(files) {
			t.Error("useTgz:", useTgz, "i:", i, ">= len(files):", len(files))
			return
		}
		if hdr.Name != files[i].name {
			t.Errorf("useTgz: %t hdr.Name: %q != %q.", useTgz, hdr.Name, files[i].name)
		}
		err = iotest.TestReader(r, files[i].body)
		if err != nil {
			t.Error("useTgz:", useTgz, err)
		}
	}
}
