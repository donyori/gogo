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
	"hash"
	"io"
	"os"

	"github.com/donyori/gogo/encoding/hex"
)

// Checksum is a combination of a hash function generator and
// an expected checksum.
type Checksum struct {
	// A function to generate a hash function. E.g. crypto/sha256.New.
	HashGen func() hash.Hash

	// Expected checksum, in hexadecimal representation.
	HexExpSum string
}

// VerifyChecksum verifies a file by checksum.
//
// It returns true if and only if the file can be read
// and matches all checksums.
// Note that it returns false if anyone of cs contains a nil HashGen
// or an empty HexExpSum,
// and it returns true if len(cs) is 0 and the file can be opened for reading.
func VerifyChecksum(filename string, cs ...Checksum) bool {
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer f.Close() // ignore error
	if len(cs) == 0 {
		return true
	}
	hashes := make([]hash.Hash, len(cs))
	ws := make([]io.Writer, len(cs))
	for i := range cs {
		if cs[i].HashGen == nil {
			return false
		}
		if cs[i].HexExpSum == "" {
			return false
		}
		hashes[i] = cs[i].HashGen()
		ws[i] = hashes[i]
	}
	w := ws[0]
	if len(ws) > 1 {
		w = io.MultiWriter(ws...)
	}
	_, err = io.Copy(w, f)
	if err != nil {
		return false
	}
	for i := range cs {
		if !hex.CanEncodeToString(hashes[i].Sum(nil), cs[i].HexExpSum) {
			return false
		}
	}
	return true
}
