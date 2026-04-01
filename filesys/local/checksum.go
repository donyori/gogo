// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2026  Yuan Gao
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
	"hash"
	"os"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/filesys"
)

// Checksum calculates hash checksums of a local file,
// and returns the result in hexadecimal representation and
// any error encountered during opening and reading the file.
//
// If the file is a directory, Checksum reports filesys.ErrIsDir
// and returns nil checksums.
// (To test whether err is filesys.ErrIsDir, use function errors.Is.)
//
// upper indicates whether to use uppercase in hexadecimal representation.
//
// newHash is a list of functions that create new hash functions
// (e.g., crypto/sha256.New, crypto.SHA256.New).
//
// The length of the returned checksums is the same as that of newHash.
// The hash result of newHash[i] is checksums[i], encoded in hexadecimal.
// In particular, if newHash[i] is nil or returns nil,
// checksums[i] is an empty string.
// If len(newHash) is 0, checksums is nil.
func Checksum(
	filename string,
	upper bool,
	newHash ...func() hash.Hash,
) (checksums []string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.AutoWrap(err)
	}
	return filesys.Checksum(f, true, upper, newHash...)
}

// VerifyChecksum verifies a local file by hash checksum.
//
// It returns true if the file can be read and matches all
// filesys.HashVerifier in hv
// (nil and duplicate filesys.HashVerifier are ignored).
// In particular, it returns true if there is no non-nil filesys.HashVerifier
// in hv and the file can be opened for reading.
// In this case, the file is not read.
//
// Note that VerifyChecksum does not reset the hash state of anyone in hv.
// The client should use new filesys.HashVerifier
// returned by filesys.NewHashVerifier or
// call the Reset method of filesys.HashVerifier
// before calling this function if needed.
func VerifyChecksum(filename string, hv ...filesys.HashVerifier) bool {
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	return filesys.VerifyChecksum(f, true, hv...)
}
