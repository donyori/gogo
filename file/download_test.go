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
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/donyori/gogo/errors"
)

var (
	testHttpDlUrl      = `https://www.gnu.org/licenses/agpl-3.0.txt`
	testHttpDlChecksum = Checksum{
		HashGen: sha256.New,
		// This SHA256 checksum was generated on March 9, 2021.
		HexExpSum: "0d96a4ff68ad6d4b6f1f30f713b18d5184912ba8dd389f86aa7710db079abcb0",
	}
	testHttpDlWrongChecksum = Checksum{
		HashGen:   sha256.New,
		HexExpSum: "0d96a4ff68ad6d4b6f1f30f713b18d5184912ba8dd389f86aa7710db07912345",
	}
)

func TestHttpDownload(t *testing.T) {
	testHttpDownloadFn(t, func(filename string) error {
		return HttpDownload(testHttpDlUrl, filename, 0600, testHttpDlChecksum)
	})
}

func TestHttpDownload_ChecksumFailed(t *testing.T) {
	testHttpDownloadFnChecksumFailed(t, func(filename string, cs ...Checksum) error {
		return HttpDownload(testHttpDlUrl, filename, 0600, cs...)
	})
}

func TestHttpCustomDownload(t *testing.T) {
	testHttpDownloadFn(t, func(filename string) error {
		req, err := http.NewRequest("", testHttpDlUrl, nil)
		if err != nil {
			return err
		}
		return HttpCustomDownload(req, filename, 0600, testHttpDlChecksum)
	})
}

func TestHttpCustomDownload_ChecksumFailed(t *testing.T) {
	testHttpDownloadFnChecksumFailed(t, func(filename string, cs ...Checksum) error {
		req, err := http.NewRequest("", testHttpDlUrl, nil)
		if err != nil {
			return err
		}
		return HttpCustomDownload(req, filename, 0600, cs...)
	})
}

func TestHttpUpdate(t *testing.T) {
	testHttpUpdateFn(t, func(filename string, cs ...Checksum) (updated bool, err error) {
		return HttpUpdate(testHttpDlUrl, filename, 0600, cs...)
	})
}

func TestHttpUpdate_ChecksumFailed(t *testing.T) {
	testHttpDownloadFnChecksumFailed(t, func(filename string, cs ...Checksum) error {
		updated, err := HttpUpdate(testHttpDlUrl, filename, 0600, cs...)
		if updated {
			t.Error("Checksum Failed Case, updated is true.")
		}
		return err
	})
}

func TestHttpCustomUpdate(t *testing.T) {
	testHttpUpdateFn(t, func(filename string, cs ...Checksum) (updated bool, err error) {
		req, err := http.NewRequest("", testHttpDlUrl, nil)
		if err != nil {
			return false, err
		}
		return HttpCustomUpdate(req, filename, 0600, cs...)
	})
}

func TestHttpCustomUpdate_ChecksumFailed(t *testing.T) {
	testHttpDownloadFnChecksumFailed(t, func(filename string, cs ...Checksum) error {
		req, err := http.NewRequest("", testHttpDlUrl, nil)
		if err != nil {
			return err
		}
		updated, err := HttpCustomUpdate(req, filename, 0600, cs...)
		if updated {
			t.Error("Checksum Failed Case, updated is true.")
		}
		return err
	})
}

func testHttpDownloadFn(t *testing.T, fn func(filename string) error) {
	dir, err := os.MkdirTemp("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	filename := filepath.Join(dir, "testfile.dat")
	err = fn(filename)
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

func testHttpUpdateFn(t *testing.T, fn func(filename string, cs ...Checksum) (updated bool, err error)) {
	dir, err := os.MkdirTemp("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	filename := filepath.Join(dir, "testfile.dat")
	updated, err := fn(filename, testHttpDlChecksum)
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
	updated, err = fn(filename, testHttpDlChecksum)
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
	updated, err = fn(filename)
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
	defer func() {
		if f != nil {
			f.Close() // ignore error
		}
	}()
	_, err = f.WriteString("abc")
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	f = nil
	now := time.Now()
	updated, err = fn(filename, testHttpDlChecksum)
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

func testHttpDownloadFnChecksumFailed(t *testing.T, fn func(filename string, cs ...Checksum) error) {
	dir, err := os.MkdirTemp("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	filename := filepath.Join(dir, "testfile.dat")
	var client http.Client
	resp, err := client.Get(testHttpDlUrl)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close() // ignore error
	f, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if f != nil {
			f.Close() // ignore error
		}
	}()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	f = nil
	info, err := os.Lstat(filename)
	if err != nil {
		t.Fatal(err)
	}
	modTime := info.ModTime()

	err = fn(filename, testHttpDlWrongChecksum)
	if !errors.Is(err, ErrVerificationFail) {
		t.Errorf("Checksum Failed Case, err: %v != %v.", err, ErrVerificationFail)
	}

	info, err = os.Lstat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !modTime.Equal(info.ModTime()) {
		t.Error("Checksum Failed Case, file has been modified.")
	}
}
