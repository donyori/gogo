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

package local

import (
	"os"

	"github.com/donyori/gogo/filesys"
)

// VerifyChecksum verifies a local file by hash checksum.
//
// It returns true if the file can be read and matches all
// filesys.HashVerifier in hvs
// (nil and duplicate filesys.HashVerifier will be ignored).
// In particular, it returns true if there is no non-nil filesys.HashVerifier
// in hvs and the file can be opened for reading.
// In this case, the file will not be read.
//
// Note that VerifyChecksum will not reset the hash state of anyone in hvs.
// The client should use new filesys.HashVerifier
// returned by filesys.NewHashVerifier or
// call the Reset method of filesys.HashVerifier
// before calling this function if needed.
func VerifyChecksum(filename string, hvs ...filesys.HashVerifier) bool {
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	return filesys.VerifyChecksum(f, true, hvs...)
}
