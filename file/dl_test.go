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
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHttpDownload(t *testing.T) {
	dir, err := ioutil.TempDir("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	filename := filepath.Join(dir, "go1.4.tar.gz")
	url := `https://dl.google.com/go/go1.4-bootstrap-20170531.tar.gz`
	chksum := Checksum{
		HashGen:   sha256.New,
		HexExpSum: "49f806f66762077861b7de7081f586995940772d29d4c45068c134441a743fa2",
	}
	err = HttpDownload(url, filename, 0600, chksum)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close() // ignore error
	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		t.Fatal(err)
	}
	if sum := hex.EncodeToString(h.Sum(nil)); sum != chksum.HexExpSum {
		t.Errorf("Checksum: %s != %s.", sum, chksum.HexExpSum)
	}
}

func TestHttpUpdate(t *testing.T) {
	dir, err := ioutil.TempDir("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	filename := filepath.Join(dir, "go1.4.tar.gz")
	url := `https://dl.google.com/go/go1.4-bootstrap-20170531.tar.gz`
	chksum := Checksum{
		HashGen:   sha256.New,
		HexExpSum: "49f806f66762077861b7de7081f586995940772d29d4c45068c134441a743fa2",
	}
	err = HttpUpdate(url, filename, 0600, chksum)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(filename)
	if err != nil {
		t.Fatal(err)
	}
	modTime := info.ModTime()
	err = HttpUpdate(url, filename, 0600, chksum)
	if err != nil {
		t.Fatal(err)
	}
	info, err = os.Lstat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !modTime.Equal(info.ModTime()) {
		t.Error("File has been modified.")
	}
	err = HttpUpdate(url, filename, 0600)
	if err != nil {
		t.Fatal(err)
	}
	info, err = os.Lstat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !modTime.Equal(info.ModTime()) {
		t.Error("File has been modified.")
	}
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString("abc")
	f.Close() // ignore error
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	err = HttpUpdate(url, filename, 0600, chksum)
	if err != nil {
		t.Fatal(err)
	}
	info, err = os.Lstat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !now.Before(info.ModTime()) {
		t.Error("File has not been updated after damaging.")
	}
}
