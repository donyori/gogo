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
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/donyori/gogo/filesys/local"
)

func TestWriteTrunc(t *testing.T) {
	tmpRoot := t.TempDir()
	name := filepath.Join(tmpRoot, "sub", "test.txt")
	data := []byte("test local.WriteTrunc\n")
	for i := 0; i < 3; i++ {
		func(t *testing.T, i int) {
			w, err := local.WriteTrunc(name, 0600, true, nil)
			if err != nil {
				t.Errorf("i: %d, create - %v", i, err)
				return
			}
			defer func() {
				if err := w.Close(); err != nil {
					t.Errorf("i: %d, close - %v", i, err)
				}
			}()
			_, err = w.Write(data)
			if err != nil {
				t.Errorf("i: %d, write - %v", i, err)
			}
		}(t, i)
		if t.Failed() {
			return
		}
	}
	got, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("got %q; want %q", got, data)
	}
}

func TestWriteAppend(t *testing.T) {
	const N = 3
	tmpRoot := t.TempDir()
	name := filepath.Join(tmpRoot, "sub", "test.txt")
	data := []byte("test local.WriteAppend\n")
	for i := 0; i < N; i++ {
		func(t *testing.T, i int) {
			w, err := local.WriteAppend(name, 0600, true, nil)
			if err != nil {
				t.Errorf("i: %d, create - %v", i, err)
				return
			}
			defer func() {
				if err := w.Close(); err != nil {
					t.Errorf("i: %d, close - %v", i, err)
				}
			}()
			_, err = w.Write(data)
			if err != nil {
				t.Errorf("i: %d, write - %v", i, err)
			}
		}(t, i)
		if t.Failed() {
			return
		}
	}
	got, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	want := bytes.Repeat(data, N)
	if !bytes.Equal(got, want) {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestWriteExcl(t *testing.T) {
	tmpRoot := t.TempDir()
	name := filepath.Join(tmpRoot, "sub", "test.txt")
	data := []byte("test local.WriteExcl\n")
	func(t *testing.T) {
		w, err := local.WriteExcl(name, 0600, true, nil)
		if err != nil {
			t.Error("create -", err)
			return
		}
		defer func() {
			if err := w.Close(); err != nil {
				t.Error("close -", err)
			}
		}()
		_, err = w.Write(data)
		if err != nil {
			t.Error("write -", err)
		}
	}(t)
	if t.Failed() {
		return
	}
	_, err := local.WriteExcl(name, 0600, true, nil)
	if !errors.Is(err, os.ErrExist) {
		t.Fatal("errors.Is(err, os.ErrExist) is false on 2nd call to WriteExcl, err:", err)
	}
	got, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("got %q; want %q", got, data)
	}
}

func TestWriteTrunc_TarTgz(t *testing.T) {
	big := make([]byte, 1_048_576)
	rand.New(rand.NewSource(10)).Read(big)
	tarFiles := []struct {
		name string
		body []byte
	}{
		{"tar file1.txt", []byte("This is tar file 1.")},
		{"tar file2.txt", []byte("Here is tar file 2!")},
		{"roses are red.txt", []byte("Roses are red.\n  Violets are blue.\nSugar is sweet.\n  And so are you.\n")},
		{"1MB.dat", big},
	}
	filenames := []string{"test.tar", "test.tar.gz", "test.tgz"}
	tmpRoot := t.TempDir()
	for _, filename := range filenames {
		t.Run(fmt.Sprintf("file=%q", filename), func(t *testing.T) {
			name := filepath.Join(tmpRoot, "sub", filename)
			writeTarFiles(t, name, tarFiles)
			if t.Failed() {
				return
			}
			testTarTgzFile(t, name, tarFiles)
		})
	}
}

// writeTarFiles uses WriteTrunc to write a tar or tgz file.
//
// name is the local file name.
//
// tarFiles are the files with their names and bodies to be archived in the tar.
//
// Caller should guarantee that name has extension ".tar", ".tar.gz", or ".tgz".
func writeTarFiles(t *testing.T, name string, tarFiles []struct {
	name string
	body []byte
}) {
	w, err := local.WriteTrunc(name, 0600, true, nil)
	if err != nil {
		t.Error("create -", err)
		return
	}
	defer func() {
		if err := w.Close(); err != nil {
			t.Error("close -", err)
		}
	}()
	for i := range tarFiles {
		err = w.TarWriteHeader(&tar.Header{
			Name:    tarFiles[i].name,
			Size:    int64(len(tarFiles[i].body)),
			Mode:    0600,
			ModTime: time.Now(),
		})
		if err != nil {
			t.Errorf("write No.%d tar header - %v", i, err)
			return
		}
		_, err = w.Write(tarFiles[i].body)
		if err != nil {
			t.Errorf("write No.%d tar file body - %v", i, err)
			return
		}
	}
}

// testTarTgzFile checks a tar or tgz file written by function writeTarFiles.
//
// Caller should guarantee that name has extension ".tar", ".tar.gz", or ".tgz".
func testTarTgzFile(t *testing.T, name string, wantTarFiles []struct {
	name string
	body []byte
}) {
	f, err := os.Open(name)
	if err != nil {
		t.Error("open -", err)
		return
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			t.Error("close file -", err)
		}
	}(f)
	var rc io.ReadCloser = f
	ext := filepath.Ext(name)
	if ext == ".gz" || ext == ".tgz" {
		rc, err = gzip.NewReader(rc)
		if err != nil {
			t.Error("create gzip reader -", err)
			return
		}
		defer func(c io.Closer) {
			if err := c.Close(); err != nil {
				t.Error("close gzip reader -", err)
			}
		}(rc)
	}
	tr := tar.NewReader(rc)
	for i := 0; ; i++ {
		hdr, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if i != len(wantTarFiles) {
					t.Errorf("tar header number: %d != %d, but got EOF", i, len(wantTarFiles))
				}
				return // end of archive
			}
			t.Errorf("read No.%d tar header - %v", i, err)
			return
		}
		if i >= len(wantTarFiles) {
			t.Error("tar headers more than", len(wantTarFiles))
			return
		}
		if hdr.Name != wantTarFiles[i].name {
			t.Errorf("No.%d tar header name unequal - got %s; want %s", i, hdr.Name, wantTarFiles[i].name)
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Errorf("read No.%d tar file body - %v", i, err)
			return
		}
		if !bytes.Equal(body, wantTarFiles[i].body) {
			t.Errorf("got No.%d tar file body\n%s\nwant\n%s", i, body, wantTarFiles[i].body)
		}
	}
}
