// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2023  Yuan Gao
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

package internal

import (
	"fmt"
	"hash"
	"io"
	"io/fs"

	"github.com/donyori/gogo/encoding/hex"
	"github.com/donyori/gogo/errors"
)

// HashChecksum is a combination of a hash function maker and
// an expected checksum.
//
// The checksum is described by a hexadecimal string and a boolean indicator,
// where the indicator is used to report whether the string represents
// a prefix of the checksum or the entire checksum.
type HashChecksum struct {
	// A function that creates a new hash function (e.g., crypto/sha256.New).
	NewHash func() hash.Hash

	// Expected checksum, in hexadecimal representation.
	WantHex string

	// True if WantHex is a prefix of the checksum.
	// False if WantHex is the entire checksum.
	IsPrefix bool
}

// These functions are shared between packages filesys and local.

// CheckHashChecksums checks hcs and panics if
// anyone of hcs contains a nil NewHash or an empty WantHex,
// or any NewHash returns nil.
//
// The panic message will be prepended with the full function name of
// the caller of CheckHashChecksums.
//
// It returns a hash.Hash list hs obtained by
// calling NewHash for each item in hcs.
// hs has the same length as hcs.
func CheckHashChecksums(hcs []HashChecksum) (hs []hash.Hash) {
	hs = make([]hash.Hash, len(hcs))
	for i := range hcs {
		// Set skip to 1 to report the caller of CheckHashChecksums.
		if hcs[i].NewHash == nil {
			panic(errors.AutoMsgCustom(fmt.Sprintf("HashChecksum with index %d has a nil NewHash", i), -1, 1))
		}
		if hcs[i].WantHex == "" {
			panic(errors.AutoMsgCustom(fmt.Sprintf("HashChecksum with index %d has an empty WantHex", i), -1, 1))
		}
		hs[i] = hcs[i].NewHash()
		if hs[i] == nil {
			panic(errors.AutoMsgCustom(fmt.Sprintf("NewHash of HashChecksum with index %d returns nil", i), -1, 1))
		}
	}
	return
}

// VerifyChecksum is an implementation of
// github.com/donyori/gogo/filesys.VerifyChecksum
// without checking the arguments.
//
// It requires one more argument hs,
// which should be obtained from CheckHashChecksums.
//
// Caller should guarantee that file is non-nil,
// and hs is returned by CheckHashChecksums(hcs).
func VerifyChecksum(file fs.File, closeFile bool, hcs []HashChecksum, hs []hash.Hash) bool {
	if closeFile {
		defer func(f fs.File) {
			_ = f.Close() // ignore error
		}(file)
	}
	if len(hcs) == 0 {
		return true
	}
	ws := make([]io.Writer, len(hs))
	for i := range hs {
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
	for i := range hcs {
		checksum := hs[i].Sum(nil)
		if hcs[i].IsPrefix {
			if !hex.CanEncodeToPrefix(checksum, hcs[i].WantHex) {
				return false
			}
		} else {
			if !hex.CanEncodeTo(checksum, hcs[i].WantHex) {
				return false
			}
		}
	}
	return true
}
