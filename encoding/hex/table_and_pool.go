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

package hex

import "sync"

// uppercaseHexTable is a uppercase character table for hexadecimal encoding.
const uppercaseHexTable = "0123456789ABCDEF"

// lowercaseHexTable is a lowercase character table for hexadecimal encoding.
const lowercaseHexTable = "0123456789abcdef"

// sourceBufferLen is the length of a chunk of source data.
const sourceBufferLen = 512

// letterCaseDiff is the result of 'A' xor 'a'.
// It indicates the different bit of the uppercase and lowercase letters.
const letterCaseDiff byte = 'A' ^ 'a'

// sourceBufferPool is a set of temporary buffers to load source data
// from readers.
//
// The user should guarantee that the size of the buffer
// put into this pool is exactly sourceBufferLen (512).
var sourceBufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, sourceBufferLen)
		return &b
	},
}

// encodeBufferPool is a set of temporary buffers to hold encoding results
// that will be written to destination writers.
//
// The user should guarantee that the size of the buffer
// put into this pool is exactly EncodedLen(sourceBufferLen) (1024).
var encodeBufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, EncodedLen(sourceBufferLen))
		return &b
	},
}

// formatBufferLen is the length of a buffer that holds formatting results.
const formatBufferLen = 1024

// formatBufferPool is a set of temporary buffers to hold formatting results
// that will be written to destination writers.
//
// The user should guarantee that the size of the buffer
// put into this pool is exactly formatBufferLen (1024).
var formatBufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, formatBufferLen)
		return &b
	},
}

// int64BufferLen is the length of a buffer that holds
// a hexadecimal representation of a 64-bit integer.
const int64BufferLen = 17

// int64BufferPool is a set of temporary buffers to hold
// the hexadecimal representation of a 64-bit integer.
//
// The user should guarantee that the size of the buffer
// put into this pool is exactly int64BufferLen (17).
var int64BufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, int64BufferLen)
		return &b
	},
}
