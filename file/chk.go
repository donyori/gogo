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
	"hash"
	"io"
	"os"
	"strings"

	"github.com/donyori/gogo/encoding/hex"
)

// A combination of a hash algorithm and an expected checksum.
type Checksum struct {
	// A function to generate a hasher. E.g. crypto/sha256.New.
	HashGen func() hash.Hash

	// Expected checksum, encoding to hexadecimal representation.
	HexExpSum string
}

// Verify a file by checksum.
// It returns true if and only if the file can be read
// and matches all checksums.
// Note that it returns false if any one of chksums contains a nil HashGen
// or an empty HexExpSum.
func VerifyChecksum(filename string, chksums ...Checksum) bool {
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer f.Close() // ignore
	if len(chksums) == 0 {
		return true
	}
	hashes := make([]hash.Hash, len(chksums))
	ws := make([]io.Writer, len(chksums))
	for i := range chksums {
		if chksums[i].HashGen == nil {
			return false
		}
		if chksums[i].HexExpSum == "" {
			return false
		}
		hashes[i] = chksums[i].HashGen()
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
	for i := range chksums {
		sum := hex.EncodeToString(hashes[i].Sum(nil), false)
		if sum != strings.ToLower(chksums[i].HexExpSum) {
			return false
		}
	}
	return true
}
