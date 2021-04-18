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

package local

import (
	"os"

	"github.com/donyori/gogo/fs"
)

// VerifyChecksum verifies a local file by checksum.
//
// It returns true if and only if the file can be read
// and matches all checksums.
//
// Note that it returns false if anyone of cs contains a nil HashGen
// or an empty HexExpSum.
// And it returns true if len(cs) is 0 and the file can be opened for reading.
func VerifyChecksum(filename string, cs ...fs.Checksum) bool {
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer func() {
		_ = f.Close()
	}()
	return fs.VerifyChecksum(f, cs...)
}
