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
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/donyori/gogo/errors"
)

func TestWrite_Tgz(t *testing.T) {
	dir, err := os.MkdirTemp("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	filename := filepath.Join(dir, "simple.tgz")
	var perm fs.FileMode = 0740
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
			t.Errorf("Permission: %3o != %3o.", p, perm)
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
		read, err := io.ReadAll(tr)
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

func testWriteAppend(t *testing.T, backup bool) {
	dir, err := os.MkdirTemp("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
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
	var perm fs.FileMode = 0740
	w1, err := Write(filename, perm, options)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if w1 != nil {
			w1.Close() // ignore error
		}
	}()
	_, err = w1.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}
	err = w1.Close()
	if err != nil {
		t.Fatal(err)
	}
	w1 = nil

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("Backup: %t. After first write, data: %s\nwanted: %s",
			backup, data, content)
	}

	w2, err := Write(filename, perm, options)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if w2 != nil {
			w2.Close() // ignore error
		}
	}()
	_, err = w2.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}
	err = w2.Close()
	if err != nil {
		t.Fatal(err)
	}
	w2 = nil

	data, err = os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content+content {
		t.Errorf("Backup: %t. After second write, data: %s\nwanted: %s",
			backup, data, content+content)
	}
}

// testErrorWriter always returns an error err
// to simulate the failed writing scenario.
type testErrorWriter struct {
	err error
}

func (tew *testErrorWriter) Write([]byte) (n int, err error) {
	return 0, tew.err
}

func TestWrite_Error(t *testing.T) {
	testWriteError(t, false, false)
	testWriteError(t, false, true)
	testWriteError(t, true, false)
	testWriteError(t, true, true)
}

func testWriteError(t *testing.T, backup, preserveOnFail bool) {
	dir, err := os.MkdirTemp("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	const content = "gogo test file."
	filename := filepath.Join(dir, "testfile.dat")
	fw, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if fw != nil {
			fw.Close() // ignore error
		}
	}()
	_, err = fw.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}
	err = fw.Close()
	if err != nil {
		t.Fatal(err)
	}
	fw = nil

	options := &WriteOptions{
		Raw:            true,
		Backup:         backup,
		PreserveOnFail: preserveOnFail,
		MkDirs:         true,
	}
	var perm fs.FileMode = 0740
	wErr := errors.New("testErrorWriter error")
	w, err := Write(filename, perm, options, &testErrorWriter{wErr})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if w != nil {
			w.Close() // ignore error
		}
	}()
	_, err = w.WriteString(content + content)
	if !errors.Is(err, wErr) {
		t.Errorf("Write err: %v != %v. Backup: %t, PreserveOnFail: %t.",
			err, wErr, backup, preserveOnFail)
	}
	err = w.Close()
	if err != nil {
		t.Errorf("Close err: %v != nil. Backup: %t, PreserveOnFail: %t.",
			err, backup, preserveOnFail)
	} else if !w.Closed() {
		t.Errorf("Closed: false. Backup: %t, PreserveOnFail: %t.",
			backup, preserveOnFail)
	}
	w = nil

	data, err := os.ReadFile(filename)
	var wantedContent string
	if backup || preserveOnFail {
		if err != nil {
			t.Errorf("Open file after writing, err: %v != nil. Backup: %t, PreserveOnFail: %t.",
				err, backup, preserveOnFail)
			return
		}
		if backup {
			wantedContent = content
		} else {
			// The data should be written to the file before
			// the testErrorWriter returns an error.
			wantedContent = content + content
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Open file after writing, err: %v is not %v. Backup: %t, PreserveOnFail: %t.",
			err, os.ErrNotExist, backup, preserveOnFail)
		return
	}
	if string(data) != wantedContent {
		t.Errorf("After writing, Backup: %t, PreserveOnFail: %t, file: %q\nwanted: %q",
			backup, preserveOnFail, data, wantedContent)
	}
}
