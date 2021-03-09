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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyChecksum(t *testing.T) {
	dir, err := ioutil.TempDir("", "gogo_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // ignore error
	filename := filepath.Join(dir, "testfile.dat")

	v := VerifyChecksum(filename)
	if v {
		t.Error("True for non-exist file.")
	}

	// Make test file:
	filename2 := filepath.Join(dir, "testfile2.dat")
	fw, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if fw != nil {
			fw.Close() // ignore error
		}
	}()
	fw2, err := os.Create(filename2)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if fw2 != nil {
			fw2.Close() // ignore error
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
	ck := Checksum{
		HashGen:   sha256.New,
		HexExpSum: hex.EncodeToString(h.Sum(nil)),
	}

	v = VerifyChecksum(filename)
	if !v {
		t.Error("False for existing file.")
	}
	v = VerifyChecksum(filename, ck)
	if !v {
		t.Error("False for intact file.")
	}
	v = VerifyChecksum(filename2, ck)
	if v {
		t.Error("True for damaged file.")
	}
	v = VerifyChecksum(filename, Checksum{})
	if v {
		t.Error("True for empty checksum.")
	}
}
