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

package local

import (
	"os"

	"github.com/donyori/gogo/filesys"
	"github.com/donyori/gogo/filesys/internal"
)

// VerifyChecksum verifies a local file by checksum.
//
// It returns true if the file can be read and matches all checksums.
// In particular, it returns true if len(hcs) is 0 and
// the file can be opened for reading.
// In this case, the file will not be read.
//
// It panics if anyone of hcs contains a nil NewHash or an empty WantHex,
// or any NewHash returns nil.
func VerifyChecksum(filename string, hcs ...filesys.HashChecksum) bool {
	hs := internal.CheckHashChecksums(hcs)
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	return internal.VerifyChecksum(f, true, hcs, hs)
}
