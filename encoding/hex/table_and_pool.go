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

// hexTable combines a uppercase character table and a lowercase character table
// used in hexadecimal encoding.
const hexTable = "0123456789ABCDEF0123456789abcdef"

// getHexTable returns a character table used in hexadecimal encoding.
// upper indicates to use uppercase letters in the encoding.
func getHexTable(upper bool) string {
	if upper {
		return hexTable[:16]
	}
	return hexTable[16:]
}

// chunkLen is the length of a chunk of source data.
const chunkLen = 512

// chunkPool is a set of temporary buffers to load source data from readers.
var chunkPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, chunkLen)
		return &b
	},
}

// encodeBufferPool is a set of temporary buffers to hold encoding results
// that will be written to destination writers.
var encodeBufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, EncodedLen(chunkLen))
		return &b
	},
}

// formatBufferLen is the length of a buffer that holds formatting results.
const formatBufferLen = 1024

// formatBufferPool is a set of temporary buffers to hold formatting results
// that will be written to destination writers.
var formatBufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, formatBufferLen)
		return &b
	},
}
