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
	"io/fs"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/filesys/internal"
)

// HashChecksum is a combination of a hash function maker and
// an expected checksum.
//
// The checksum is described by a hexadecimal string and a boolean indicator,
// where the indicator is used to report whether the string represents
// a prefix of the checksum or the entire checksum.
//
// HashChecksum has three fields:
//   - NewHash func() hash.Hash // A function that creates a new hash function (e.g., crypto/sha256.New).
//   - WantHex string // Expected checksum, in hexadecimal representation.
//   - IsPrefix bool // True if WantHex is a prefix of the checksum; false if WantHex is the entire checksum.
type HashChecksum = internal.HashChecksum

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
// In particular, it returns true if len(hcs) is 0.
// In this case, the file will not be read.
//
// It panics if file is nil,
// anyone of hcs contains a nil NewHash or an empty WantHex,
// or any NewHash returns nil.
func VerifyChecksum(file fs.File, closeFile bool, hcs ...HashChecksum) bool {
	if file == nil {
		panic(errors.AutoMsg("file is nil"))
	}
	return internal.VerifyChecksum(file, closeFile, hcs, internal.CheckHashChecksums(hcs))
}

// VerifyChecksumFromFS verifies a file by checksum,
// where the file is opened from fsys by the specified name.
//
// It returns true if the file can be read and matches all checksums.
// In particular, it returns true if len(hcs) is 0 and
// the file can be opened for reading.
// In this case, the file will not be read.
//
// It panics if fsys is nil,
// anyone of hcs contains a nil NewHash or an empty WantHex,
// or any NewHash returns nil.
func VerifyChecksumFromFS(fsys fs.FS, name string, hcs ...HashChecksum) bool {
	if fsys == nil {
		panic(errors.AutoMsg("fsys is nil"))
	}
	hs := internal.CheckHashChecksums(hcs)
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	return internal.VerifyChecksum(f, true, hcs, hs)
}
