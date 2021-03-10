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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/donyori/gogo/errors"
)

func TestWrite_TarGz(t *testing.T) {
	dir, err := ioutil.TempDir("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	filename := filepath.Join(dir, "simple.tar.gz")
	var perm os.FileMode = 0740
	w, err := Write(filename, perm, &WriteOptions{
		BufOpen: true,
		Backup:  true,
		MkDirs:  true,
		GzipLv:  gzip.BestCompression,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if w != nil {
			w.Close() // ignore error
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
		err = w.TarWriteHeader(hdr)
		if err != nil {
			t.Fatal(err)
		}
		_, err = w.WriteString(files[i].body)
		if err != nil {
			t.Fatal(err)
		}
		err = w.Flush()
		if err != nil {
			t.Fatal(err)
		}
	}
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	w = nil

	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close() // ignore error
	if runtime.GOOS != "windows" {
		dirInfo, err := os.Lstat(dir)
		if err != nil {
			t.Fatal(err)
		}
		info, err := f.Stat()
		if err != nil {
			t.Fatal(err)
		}
		if p := info.Mode().Perm(); p != dirInfo.Mode().Perm()&perm {
			t.Errorf("Permission: %o != %o.", p, perm)
		}
	}
	gzr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gzr.Close() // ignore error
	tr := tar.NewReader(gzr)
	for i := 0; true; i++ {
		hdr, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatal(err)
		}
		if i >= len(files) {
			t.Fatal("i:", i, ">= len(files):", len(files))
		}
		read, err := ioutil.ReadAll(tr)
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

func TestWrite_Append(t *testing.T) {
	testWriteAppend(t, true)
	testWriteAppend(t, false)
}

func testWriteAppend(tb testing.TB, backup bool) {
	dir, err := ioutil.TempDir("", "gogo_test_")
	if err != nil {
		tb.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	const content = "gogo test file."
	filename := filepath.Join(dir, "testfile.dat")
	options := &WriteOptions{
		Append: true,
		Raw:    true,
		Backup: backup,
		MkDirs: true,
	}
	var perm os.FileMode = 0740
	w1, err := Write(filename, perm, options)
	if err != nil {
		tb.Fatal(err)
	}
	defer func() {
		if w1 != nil {
			w1.Close() // ignore error
		}
	}()
	_, err = fmt.Fprint(w1, content)
	if err != nil {
		tb.Fatal(err)
	}
	err = w1.Close()
	if err != nil {
		tb.Fatal(err)
	}
	w1 = nil

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		tb.Fatal(err)
	}
	if string(data) != content {
		tb.Errorf("After first write: data: %s\nwanted: %s", data, content)
	}

	w2, err := Write(filename, perm, options)
	if err != nil {
		tb.Fatal(err)
	}
	defer func() {
		if w2 != nil {
			w2.Close() // ignore error
		}
	}()
	_, err = fmt.Fprint(w2, content)
	if err != nil {
		tb.Fatal(err)
	}
	err = w2.Close()
	if err != nil {
		tb.Fatal(err)
	}
	w2 = nil

	data, err = ioutil.ReadFile(filename)
	if err != nil {
		tb.Fatal(err)
	}
	if string(data) != content+content {
		tb.Errorf("After second write: data: %s\nwanted: %s", data, content+content)
	}
}
