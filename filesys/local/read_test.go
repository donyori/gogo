// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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

	"github.com/donyori/gogo/filesys"
)

func TestRead_Basic(t *testing.T) {
	dir, err := os.MkdirTemp("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer func(dir string) {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	}(dir)
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
	defer func(r filesys.Reader) {
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	}(r)
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
	defer func(dir string) {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	}(dir)
	data := []byte("Some data in the temporary gzip file.\n")
	filename := filepath.Join(dir, "simple.txt.gz")
	f, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if f != nil {
			if err := f.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	gzw := gzip.NewWriter(f)
	defer func() {
		if gzw != nil {
			if err := gzw.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	_, err = gzw.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	err = gzw.Close()
	if err != nil {
		t.Fatal(err)
	}
	gzw = nil
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	f = nil

	r, err := Read(filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func(r filesys.Reader) {
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	}(r)
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
	defer func(dir string) {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	}(dir)
	filename := filepath.Join(dir, "simple.tar")
	f, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if f != nil {
			if err := f.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	tw := tar.NewWriter(f)
	defer func() {
		if tw != nil {
			if err := tw.Close(); err != nil {
				t.Error(err)
			}
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
	err = tw.Close()
	if err != nil {
		t.Fatal(err)
	}
	tw = nil
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	f = nil

	r, err := Read(filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func(r filesys.Reader) {
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	}(r)
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
	defer func(dir string) {
		if err := os.RemoveAll(dir); err != nil {
			t.Error("useTgz:", useTgz, err)
		}
	}(dir)
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
	defer func() {
		if f != nil {
			if err := f.Close(); err != nil {
				t.Error("useTgz:", useTgz, err)
			}
		}
	}()
	gzw := gzip.NewWriter(f)
	defer func() {
		if gzw != nil {
			if err := gzw.Close(); err != nil {
				t.Error("useTgz:", useTgz, err)
			}
		}
	}()
	tw := tar.NewWriter(gzw)
	defer func() {
		if tw != nil {
			if err := tw.Close(); err != nil {
				t.Error("useTgz:", useTgz, err)
			}
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
	err = tw.Close()
	if err != nil {
		t.Error("useTgz:", useTgz, err)
		return
	}
	tw = nil
	err = gzw.Close()
	if err != nil {
		t.Error("useTgz:", useTgz, err)
		return
	}
	gzw = nil
	err = f.Close()
	if err != nil {
		t.Error("useTgz:", useTgz, err)
		return
	}
	f = nil

	r, err := Read(filename, &filesys.ReadOptions{BufOpen: true})
	if err != nil {
		t.Error("useTgz:", useTgz, err)
		return
	}
	defer func(r filesys.Reader) {
		if err := r.Close(); err != nil {
			t.Error("useTgz:", useTgz, err)
		}
	}(r)
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
