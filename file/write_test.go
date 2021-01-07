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
	"runtime"
	"testing"
)

func TestNew_TarGz(t *testing.T) {
	dir, err := ioutil.TempDir("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	filename := filepath.Join(dir, "simple.tar.gz")
	var perm os.FileMode = 0740
	writer, err := New(filename, perm, &WriteOption{
		BufferWhenOpen: true,
		Backup:         true,
		MakeDirs:       true,
		GzipLv:         gzip.BestCompression,
	})
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	defer func() {
		if !closed {
			writer.Close() // ignore error
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
		err = writer.TarWriteHeader(hdr)
		if err != nil {
			t.Fatal(err)
		}
		_, err = writer.WriteString(files[i].body)
		if err != nil {
			t.Fatal(err)
		}
		err = writer.Flush()
		if err != nil {
			t.Fatal(err)
		}
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}
	closed = true

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
