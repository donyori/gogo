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
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	testHttpDlUrl      = `https://www.gnu.org/licenses/agpl-3.0.txt`
	testHttpDlChecksum = Checksum{
		HashGen: sha256.New,
		// This SHA256 checksum was generated on March 9, 2021.
		HexExpSum: "0d96a4ff68ad6d4b6f1f30f713b18d5184912ba8dd389f86aa7710db079abcb0",
	}
)

func TestHttpDownload(t *testing.T) {
	dir, err := ioutil.TempDir("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	filename := filepath.Join(dir, "testfile.dat")
	err = HttpDownload(testHttpDlUrl, filename, 0600, testHttpDlChecksum)
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
	if sum := hex.EncodeToString(h.Sum(nil)); sum != testHttpDlChecksum.HexExpSum {
		t.Errorf("Checksum: %s != %s.", sum, testHttpDlChecksum.HexExpSum)
	}
}

func TestHttpUpdate(t *testing.T) {
	dir, err := ioutil.TempDir("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	filename := filepath.Join(dir, "testfile.dat")
	updated, err := HttpUpdate(testHttpDlUrl, filename, 0600, testHttpDlChecksum)
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Error("updated is false for the first call.")
	}
	info, err := os.Lstat(filename)
	if err != nil {
		t.Fatal(err)
	}
	modTime := info.ModTime()
	updated, err = HttpUpdate(testHttpDlUrl, filename, 0600, testHttpDlChecksum)
	if err != nil {
		t.Fatal(err)
	}
	if updated {
		t.Error("updated is true for the second call.")
	}
	info, err = os.Lstat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !modTime.Equal(info.ModTime()) {
		t.Error("File has been modified.")
	}
	updated, err = HttpUpdate(testHttpDlUrl, filename, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if updated {
		t.Error("updated is true for the third call.")
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
	updated, err = HttpUpdate(testHttpDlUrl, filename, 0600, testHttpDlChecksum)
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Error("updated is false for the fourth call (after damaging).")
	}
	info, err = os.Lstat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !now.Before(info.ModTime()) {
		t.Error("File has not been updated after damaging.")
	}
}
