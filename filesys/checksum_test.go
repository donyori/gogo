// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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

package filesys

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestVerifyChecksumFromFs(t *testing.T) {
	wrongChecksum := Checksum{
		HashGen:   sha256.New,
		HexExpSum: strings.Repeat("0", hex.EncodedLen(sha256.Size)),
	}
	if VerifyChecksumFromFs(testFs, "nonexist") {
		t.Error("True for non-exist file.")
	}
	for _, name := range testFsFilenames {
		if !VerifyChecksumFromFs(testFs, name) {
			t.Errorf("file: %s, no checksum, false for existing file.", name)
		}
		if !VerifyChecksumFromFs(testFs, name, testFsChecksumMap[name]...) {
			t.Errorf("file: %s, false for intact file.", name)
		}
		if VerifyChecksumFromFs(testFs, name, wrongChecksum) {
			t.Errorf("file: %s, only wrong checksum, true for wrong checksum.", name)
		}
		if VerifyChecksumFromFs(testFs, name, testFsChecksumMap[name][0], wrongChecksum) {
			t.Errorf("file: %s, wrong checksum, true for wrong checksum.", name)
		}
		if VerifyChecksumFromFs(testFs, name, wrongChecksum, wrongChecksum) {
			t.Errorf("file: %s, two wrong checksums, true for wrong checksum.", name)
		}
		if VerifyChecksumFromFs(testFs, name, Checksum{}) {
			t.Errorf("file: %s, only empty checksum, true for empty checksum.", name)
		}
		if VerifyChecksumFromFs(testFs, name, Checksum{HashGen: sha256.New}) {
			t.Errorf("file: %s, only empty checksum HexExpSum, true for empty checksum HexExpSum.", name)
		}
		if VerifyChecksumFromFs(testFs, name, Checksum{HexExpSum: testFsChecksumMap[name][0].HexExpSum}) {
			t.Errorf("file: %s, only nil checksum HashGen, true for nil checksum HashGen.", name)
		}
		if VerifyChecksumFromFs(testFs, name, testFsChecksumMap[name][0], Checksum{}) {
			t.Errorf("file: %s, empty checksum, true for empty checksum.", name)
		}
		if VerifyChecksumFromFs(testFs, name, Checksum{}, Checksum{}) {
			t.Errorf("file: %s, two empty checksums, true for empty checksum.", name)
		}
	}
}
