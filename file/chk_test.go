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
	filename := filepath.Join(dir, "go1.4.tar.gz")
	v := VerifyChecksum(filename)
	if v {
		t.Error("True for non-exist file.")
	}

	chksum := Checksum{
		HashGen:   sha256.New,
		HexExpSum: "49f806f66762077861b7de7081f586995940772d29d4c45068c134441a743fa2",
	}
	url := `https://dl.google.com/go/go1.4-bootstrap-20170531.tar.gz`
	err = HttpDownload(url, filename, 0600, chksum)
	if err != nil {
		t.Fatal(err)
	}

	v = VerifyChecksum(filename)
	if !v {
		t.Error("False for existing file.")
	}
	v = VerifyChecksum(filename, chksum)
	if !v {
		t.Error("False for intact file.")
	}
	chksum.HexExpSum = "49f806f66762077861b7de7081f586995940772d29d4c45068c134441a743fa3"
	v = VerifyChecksum(filename, chksum)
	if v {
		t.Error("True for damaged file.")
	}
	v = VerifyChecksum(filename, Checksum{})
	if v {
		t.Error("True for empty checksum.")
	}
}
