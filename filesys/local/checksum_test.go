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

package local

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/donyori/gogo/filesys"
)

func TestVerifyChecksum(t *testing.T) {
	dir, err := os.MkdirTemp("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer func(dir string) {
		if err := os.RemoveAll(dir); err != nil {
			t.Error(err)
		}
	}(dir)
	filename := filepath.Join(dir, "testfile.dat")

	if VerifyChecksum(filename) {
		t.Error("True for non-exist file.")
	}

	// Make test files:
	fw, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if fw != nil {
			if err := fw.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	filename2 := filepath.Join(dir, "testfile2.dat")
	fw2, err := os.Create(filename2)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if fw2 != nil {
			if err := fw2.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	h := sha256.New()
	w := io.MultiWriter(fw, h)
	w2 := io.MultiWriter(fw, fw2, h)
	for i := 0; i < 5000; i++ {
		_, err = fmt.Fprintln(w2, "gogo test file.")
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 5000; i++ {
		_, err = fmt.Fprintln(w, "gogo test file.")
		if err != nil {
			t.Fatal(err)
		}
	}
	err = fw.Close()
	if err != nil {
		t.Fatal(err)
	}
	fw = nil
	err = fw2.Close()
	if err != nil {
		t.Fatal(err)
	}
	fw2 = nil

	ck := filesys.Checksum{
		HashGen:   sha256.New,
		HexExpSum: hex.EncodeToString(h.Sum(nil)),
	}

	if !VerifyChecksum(filename) {
		t.Error("False for existing file.")
	}
	if !VerifyChecksum(filename, ck) {
		t.Error("False for intact file.")
	}
	if VerifyChecksum(filename2, ck) {
		t.Error("True for damaged file.")
	}
	if VerifyChecksum(filename, filesys.Checksum{}) {
		t.Error("True for empty checksum.")
	}
	if VerifyChecksum(filename, filesys.Checksum{HashGen: sha256.New}) {
		t.Error("True for empty checksum HexExpSum.")
	}
	if VerifyChecksum(filename, filesys.Checksum{HexExpSum: ck.HexExpSum}) {
		t.Error("True for nil checksum HashGen.")
	}
}
