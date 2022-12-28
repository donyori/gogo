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
	"io/fs"

	"github.com/donyori/gogo/encoding/hex"
)

// HashChecksum is a combination of a hash function maker and
// an expected checksum.
type HashChecksum struct {
	// A function that creates a new hash function (e.g., crypto/sha256.New).
	NewHash func() hash.Hash

	// Expected checksum, in hexadecimal representation.
	ExpHex string
}

// VerifyChecksum verifies a file by checksum.
//
// To ensure that this function can work as expected,
// the input file must be ready to be read from the beginning and
// must not be operated by anyone else during the call to this function.
//
// closeFile indicates whether this function should close the file.
// If closeFile is false, the client is responsible for closing file after use.
// If closeFile is true and file is not nil,
// file will be closed by this function.
//
// It returns true if the file can be read and matches all checksums.
//
// Note that it returns false if file is nil,
// or anyone of cs contains a nil NewHash or an empty ExpHex.
// And it returns true if file is not nil and len(cs) is 0.
func VerifyChecksum(file fs.File, closeFile bool, cs ...HashChecksum) bool {
	if file == nil {
		return false
	}
	if closeFile {
		defer func(f fs.File) {
			_ = f.Close() // ignore error
		}(file)
	}
	if len(cs) == 0 {
		return true
	}
	hs := make([]hash.Hash, len(cs))
	ws := make([]io.Writer, len(cs))
	for i := range cs {
		if cs[i].NewHash == nil {
			return false
		}
		if cs[i].ExpHex == "" {
			return false
		}
		hs[i] = cs[i].NewHash()
		ws[i] = hs[i]
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
		if !hex.CanEncodeTo(hs[i].Sum(nil), cs[i].ExpHex) {
			return false
		}
	}
	return true
}

// VerifyChecksumFromFS verifies a file by checksum,
// where the file is opened from fsys by the specified name.
//
// It returns true if and only if the file can be read
// and matches all checksums.
//
// Note that it returns false if fsys is nil,
// or anyone of cs contains a nil NewHash or an empty ExpHex.
// And it returns true if len(cs) is 0 and the file can be opened for reading.
func VerifyChecksumFromFS(fsys fs.FS, name string, cs ...HashChecksum) bool {
	if fsys == nil {
		return false
	}
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	return VerifyChecksum(f, true, cs...)
}
