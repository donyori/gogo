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
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/donyori/gogo/filesys/local"
)

func TestWriteTrunc(t *testing.T) {
	tmpRoot := t.TempDir()
	name := filepath.Join(tmpRoot, "sub", "test.txt")
	data := []byte("test local.WriteTrunc\n")
	for i := 0; i < 3; i++ {
		func(t *testing.T, i int) {
			w, err := local.WriteTrunc(name, 0600, nil)
			if err != nil {
				t.Errorf("i: %d, WriteTrunc - %v", i, err)
				return
			}
			defer func() {
				if err := w.Close(); err != nil {
					t.Errorf("i: %d, w.Close - %v", i, err)
				}
			}()
			_, err = w.Write(data)
			if err != nil {
				t.Errorf("i: %d, w.Write - %v", i, err)
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
			w, err := local.WriteAppend(name, 0600, nil)
			if err != nil {
				t.Errorf("i: %d, WriteAppend - %v", i, err)
				return
			}
			defer func() {
				if err := w.Close(); err != nil {
					t.Errorf("i: %d, w.Close - %v", i, err)
				}
			}()
			_, err = w.Write(data)
			if err != nil {
				t.Errorf("i: %d, w.Write - %v", i, err)
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
		w, err := local.WriteExcl(name, 0600, nil)
		if err != nil {
			t.Error("WriteExcl -", err)
			return
		}
		defer func() {
			if err := w.Close(); err != nil {
				t.Error("w.Close -", err)
			}
		}()
		_, err = w.Write(data)
		if err != nil {
			t.Error("w.Write -", err)
		}
	}(t)
	_, err := local.WriteExcl(name, 0600, nil)
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
