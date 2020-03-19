// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestReadFile_Basic(t *testing.T) {
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
	reader, err := ReadFile(filename, nil)
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

func TestReadFile_Gzip(t *testing.T) {
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
	var closed bool
	defer func() {
		if !closed {
			f.Close() // ignore error
		}
	}()
	gzw := gzip.NewWriter(f)
	_, err = gzw.Write(data)
	defer func() {
		if !closed {
			gzw.Close()
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
	gzw.Close() // ignore error
	f.Close()   // ignore error
	closed = true

	reader, err := ReadFile(filename, nil)
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
