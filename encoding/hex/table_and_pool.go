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

package hex

import "sync"

// uppercaseHexTable is an uppercase character table for hexadecimal encoding.
const uppercaseHexTable = "0123456789ABCDEF"

// lowercaseHexTable is a lowercase character table for hexadecimal encoding.
const lowercaseHexTable = "0123456789abcdef"

// sourceBufferLen is the length of a chunk of source data.
const sourceBufferLen int = 512

// letterCaseDiff is the result of 'A' xor 'a'.
// It indicates the different bit of the uppercase and lowercase letters.
//
// A letter byte ('A'-'Z','a'-'z', binary representation:
// 0b0100_0001-0b0101_1010,0b0110_0001-0b0111_1010) bitwise or letterCaseDiff
// gets the corresponding lowercase letter ('a'-'z').
// A number byte ('0'-'9', binary representation: 0b0011_0000-0b0011_1001)
// bitwise or letterCaseDiff gets the number itself.
// But the bytes 0b0001_0000-0b0001_1001 bitwise or letterCaseDiff
// gets the number byte ('0'-'9').
// Therefore, a necessary and sufficient condition to test whether
// a specified byte c equals to a byte x of the lowercase character table
// (lowercaseHexTable) case-insensitively, is:
//
//	c >= '0' && c|letterCaseDiff == x
const letterCaseDiff byte = 'A' ^ 'a'

// sourceBufferPool is a set of temporary buffers to load source data
// from readers.
//
// The type of the buffers is *[sourceBufferLen]byte.
var sourceBufferPool = sync.Pool{
	New: func() any {
		return new([sourceBufferLen]byte)
	},
}

// encodeBufferPool is a set of temporary buffers to hold encoding results
// that will be written to destination writers.
//
// The type of the buffers is *[sourceBufferLen*2]byte.
var encodeBufferPool = sync.Pool{
	New: func() any {
		return new([sourceBufferLen * 2]byte)
	},
}

// int64BufferLen is the length of a buffer that holds
// a hexadecimal representation of a 64-bit integer.
const int64BufferLen int = 17

// int64BufferPool is a set of temporary buffers to hold
// the hexadecimal representation of a 64-bit integer.
//
// The type of the buffers is *[int64BufferLen]byte.
var int64BufferPool = sync.Pool{
	New: func() any {
		return new([int64BufferLen]byte)
	},
}
