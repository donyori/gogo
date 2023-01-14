// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2023  Yuan Gao
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
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/donyori/gogo/filesys"
	"github.com/donyori/gogo/filesys/local"
)

func TestWriteTrunc(t *testing.T) {
	tmpRoot := t.TempDir()
	name := filepath.Join(tmpRoot, "test.txt")
	data := []byte("test local.WriteTrunc\n")
	for i := 0; i < 3; i++ {
		func(t *testing.T, i int) {
			w, err := local.WriteTrunc(name, 0600, true, nil)
			if err != nil {
				t.Errorf("i: %d, create - %v", i, err)
				return
			}
			defer func(w filesys.Writer) {
				if err := w.Close(); err != nil {
					t.Errorf("i: %d, close - %v", i, err)
				}
			}(w)
			n, err := w.Write(data)
			if n != len(data) || err != nil {
				t.Errorf("i: %d, write - got (%d, %v); want (%d, nil)", i, n, err, len(data))
			}
		}(t, i)
		if t.Failed() {
			return
		}
	}
	got, err := os.ReadFile(name)
	if err != nil {
		t.Fatal("read output -", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("got %q; want %q", got, data)
	}
}

func TestWriteTrunc_MkDirs(t *testing.T) {
	testWriteFuncMkDirs(t, local.WriteTrunc)
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
		t.Run(fmt.Sprintf("file=%+q", filename), func(t *testing.T) {
			name := filepath.Join(tmpRoot, filename)
			writeTarFiles(t, name, tarFiles)
			if !t.Failed() {
				testTarTgzFile(t, name, tarFiles)
			}
		})
	}
}

func TestWriteAppend(t *testing.T) {
	const N int = 3
	tmpRoot := t.TempDir()
	name := filepath.Join(tmpRoot, "test.txt")
	data := []byte("test local.WriteAppend\n")
	for i := 0; i < N; i++ {
		func(t *testing.T, i int) {
			w, err := local.WriteAppend(name, 0600, true, nil)
			if err != nil {
				t.Errorf("i: %d, create - %v", i, err)
				return
			}
			defer func(w filesys.Writer) {
				if err := w.Close(); err != nil {
					t.Errorf("i: %d, close - %v", i, err)
				}
			}(w)
			n, err := w.Write(data)
			if n != len(data) || err != nil {
				t.Errorf("i: %d, write - got (%d, %v); want (%d, nil)", i, n, err, len(data))
			}
		}(t, i)
		if t.Failed() {
			return
		}
	}
	got, err := os.ReadFile(name)
	if err != nil {
		t.Fatal("read output -", err)
	}
	want := bytes.Repeat(data, N)
	if !bytes.Equal(got, want) {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestWriteAppend_MkDirs(t *testing.T) {
	testWriteFuncMkDirs(t, local.WriteAppend)
}

func TestWriteExcl(t *testing.T) {
	tmpRoot := t.TempDir()
	name := filepath.Join(tmpRoot, "test.txt")
	data := []byte("test local.WriteExcl\n")
	func(t *testing.T) {
		w, err := local.WriteExcl(name, 0600, true, nil)
		if err != nil {
			t.Error("create -", err)
			return
		}
		defer func(w filesys.Writer) {
			if err := w.Close(); err != nil {
				t.Error("close -", err)
			}
		}(w)
		n, err := w.Write(data)
		if n != len(data) || err != nil {
			t.Errorf("write - got (%d, %v); want (%d, nil)", n, err, len(data))
		}
	}(t)
	if t.Failed() {
		return
	}
	_, err := local.WriteExcl(name, 0600, true, nil)
	if !errors.Is(err, fs.ErrExist) {
		t.Fatal("errors.Is(err, fs.ErrExist) is false on 2nd call to WriteExcl, err:", err)
	}
	got, err := os.ReadFile(name)
	if err != nil {
		t.Fatal("read output -", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("got %q; want %q", got, data)
	}
}

func TestWriteExcl_MkDirs(t *testing.T) {
	testWriteFuncMkDirs(t, local.WriteExcl)
}

// testWriteFuncMkDirs tests parameter mkDirs of
// functions WriteTrunc, WriteAppend, and WriteExcl.
//
// writeFn should be one of WriteTrunc, WriteAppend, and WriteExcl.
func testWriteFuncMkDirs(t *testing.T,
	writeFn func(name string, perm fs.FileMode, mkDirs bool, opts *filesys.WriteOptions) (w filesys.Writer, err error)) {
	tmpRoot := t.TempDir()
	data := []byte("test local.WriteTrunc - mkDirs\n")

	t.Run("mkDirs=true", func(t *testing.T) {
		name := filepath.Join(tmpRoot, "true", "test.txt")
		func(t *testing.T) {
			w, err := writeFn(name, 0600, true, nil)
			if err != nil {
				t.Error("create -", err)
				return
			}
			defer func(w filesys.Writer) {
				if err := w.Close(); err != nil {
					t.Error("close -", err)
				}
			}(w)
			n, err := w.Write(data)
			if n != len(data) || err != nil {
				t.Errorf("write - got (%d, %v); want (%d, nil)", n, err, len(data))
			}
		}(t)
		if t.Failed() {
			return
		}
		got, err := os.ReadFile(name)
		if err != nil {
			t.Fatal("read output -", err)
		}
		if !bytes.Equal(got, data) {
			t.Errorf("got %q; want %q", got, data)
		}
	})

	t.Run("mkDirs=false", func(t *testing.T) {
		name := filepath.Join(tmpRoot, "false", "test.txt")
		w, err := writeFn(name, 0600, false, nil)
		if err == nil {
			if err := w.Close(); err != nil {
				t.Error("close -", err)
			}
		}
		if !errors.Is(err, fs.ErrNotExist) {
			t.Error("errors.Is(err, fs.ErrNotExist) is false, err:", err)
		}
	})
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
	defer func(w filesys.Writer) {
		if err := w.Close(); err != nil {
			t.Error("close -", err)
		}
	}(w)
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
		var n int
		n, err = w.Write(tarFiles[i].body)
		if n != len(tarFiles[i].body) || err != nil {
			t.Errorf("write No.%d tar file body - got (%d, %v); want (%d, nil)", i, n, err, len(tarFiles[i].body))
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
	var r io.Reader = f
	ext := filepath.Ext(name)
	if ext == ".gz" || ext == ".tgz" {
		gr, err := gzip.NewReader(r)
		if err != nil {
			t.Error("create gzip reader -", err)
			return
		}
		defer func(gr *gzip.Reader) {
			if err := gr.Close(); err != nil {
				t.Error("close gzip reader -", err)
			}
		}(gr)
		r = gr
	}
	tr := tar.NewReader(r)
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
			t.Errorf(
				"got No.%d tar file body (len: %d)\n%s\nwant (len: %d)\n%s",
				i,
				len(body),
				body,
				len(wantTarFiles[i].body),
				wantTarFiles[i].body,
			)
		}
	}
}
