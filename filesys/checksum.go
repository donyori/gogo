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

package filesys

import (
	"hash"
	"io"

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
// To ensure that this function can work as expected,
// the input file must be ready to be read from the beginning and
// must not be operated by anyone else during the call to this function.
//
// This function will not close file.
// The client is responsible for closing file after use.
//
// It returns true if the file can be read and matches all checksums.
//
// Note that it returns false if file is nil,
// or anyone of cs contains a nil HashGen or an empty HexExpSum.
// And it returns true if len(cs) is 0.
func VerifyChecksum(file File, cs ...Checksum) bool {
	if file == nil {
		return false
	}
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
	_, err := io.Copy(w, file)
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

// VerifyChecksumFromFs verifies a file by checksum,
// where the file is opened from fsys by the specified name.
//
// It returns true if and only if the file can be read
// and matches all checksums.
//
// Note that it returns false if fsys is nil,
// or anyone of cs contains a nil HashGen or an empty HexExpSum.
// And it returns true if len(cs) is 0 and the file can be opened for reading.
func VerifyChecksumFromFs(fsys FS, name string, cs ...Checksum) bool {
	if fsys == nil {
		return false
	}
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	defer func() {
		_ = f.Close()
	}()
	return VerifyChecksum(f, cs...)
}
