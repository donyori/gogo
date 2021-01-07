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

package file

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestRead_Basic(t *testing.T) {
	dir, err := ioutil.TempDir("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	data := []byte("Some data in the temporary file.\n")
	filename := filepath.Join(dir, "basic.txt")
	err = ioutil.WriteFile(filename, data, 0600)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := Read(filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := reader.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	read, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(read) != string(data) {
		t.Errorf("Got: %q != %q.", read, data)
	}
}

func TestRead_Gzip(t *testing.T) {
	dir, err := ioutil.TempDir("", "gogo_test_")
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

	reader, err := Read(filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := reader.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	read, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(read) != string(data) {
		t.Errorf("Got: %q != %q.", read, data)
	}
}

func TestRead_Tar(t *testing.T) {
	dir, err := ioutil.TempDir("", "gogo_test_")
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
		name, body string
	}{
		{"file1.txt", "This is file1."},
		{"file2.txt", "This is file2."},
		{"file3.txt", "This is file3."},
	}
	for i := range files {
		hdr := &tar.Header{
			Name: files[i].name,
			Mode: 0600,
			Size: int64(len(files[i].body)),
		}
		err = tw.WriteHeader(hdr)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tw.Write([]byte(files[i].body))
		if err != nil {
			t.Fatal(err)
		}
	}
	tw.Close() // ignore error
	f.Close()  // ignore error
	closed = true

	reader, err := Read(filename, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := reader.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	for i := 0; true; i++ {
		hdr, err := reader.TarNext()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatal(err)
		}
		if i >= len(files) {
			t.Fatal("i:", i, ">= len(files):", len(files))
		}
		read, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		if hdr.Name != files[i].name {
			t.Errorf("hdr.Name: %q != %q.", hdr.Name, files[i].name)
		}
		if string(read) != files[i].body {
			t.Errorf("Got: %q != %q.", read, files[i].body)
		}
	}
}

func TestRead_Tgz(t *testing.T) {
	dir, err := ioutil.TempDir("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	filename := filepath.Join(dir, "simple.tgz")
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
	tw := tar.NewWriter(gzw)
	defer func() {
		if !closed {
			tw.Close() // ignore error
		}
	}()
	files := []struct {
		name, body string
	}{
		{"file1.txt", "This is file1."},
		{"file2.txt", "This is file2."},
		{"file3.txt", "This is file3."},
	}
	for i := range files {
		hdr := &tar.Header{
			Name: files[i].name,
			Mode: 0600,
			Size: int64(len(files[i].body)),
		}
		err = tw.WriteHeader(hdr)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tw.Write([]byte(files[i].body))
		if err != nil {
			t.Fatal(err)
		}
	}
	tw.Close()  // ignore error
	gzw.Close() // ignore error
	f.Close()   // ignore error
	closed = true

	reader, err := Read(filename, &ReadOption{BufferWhenOpen: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := reader.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	for i := 0; true; i++ {
		hdr, err := reader.TarNext()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatal(err)
		}
		if i >= len(files) {
			t.Fatal("i:", i, ">= len(files):", len(files))
		}
		read, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		if hdr.Name != files[i].name {
			t.Errorf("hdr.Name: %q != %q.", hdr.Name, files[i].name)
		}
		if string(read) != files[i].body {
			t.Errorf("Got: %q != %q.", read, files[i].body)
		}
	}
}

func TestRead_TarGz(t *testing.T) {
	dir, err := ioutil.TempDir("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	filename := filepath.Join(dir, "simple.tar.gz")
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
	tw := tar.NewWriter(gzw)
	defer func() {
		if !closed {
			tw.Close() // ignore error
		}
	}()
	files := []struct {
		name, body string
	}{
		{"file1.txt", "This is file1."},
		{"file2.txt", "This is file2."},
		{"file3.txt", "This is file3."},
	}
	for i := range files {
		hdr := &tar.Header{
			Name: files[i].name,
			Mode: 0600,
			Size: int64(len(files[i].body)),
		}
		err = tw.WriteHeader(hdr)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tw.Write([]byte(files[i].body))
		if err != nil {
			t.Fatal(err)
		}
	}
	tw.Close()  // ignore error
	gzw.Close() // ignore error
	f.Close()   // ignore error
	closed = true

	reader, err := Read(filename, &ReadOption{BufferWhenOpen: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := reader.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	for i := 0; true; i++ {
		hdr, err := reader.TarNext()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatal(err)
		}
		if i >= len(files) {
			t.Fatal("i:", i, ">= len(files):", len(files))
		}
		read, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		if hdr.Name != files[i].name {
			t.Errorf("hdr.Name: %q != %q.", hdr.Name, files[i].name)
		}
		if string(read) != files[i].body {
			t.Errorf("Got: %q != %q.", read, files[i].body)
		}
	}
}
