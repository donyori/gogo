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

package hex

import "sync"

const hexTable = "0123456789ABCDEF0123456789abcdef"

func getHexTable(upper bool) string {
	if upper {
		return hexTable[:16]
	}
	return hexTable[16:]
}

const chunkLen = 512

var chunkPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, chunkLen)
		return &b
	},
}

var encodeBufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, EncodedLen(chunkLen))
		return &b
	},
}

const formatBufferLen = 1024

var formatBufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, formatBufferLen)
		return &b
	},
}
