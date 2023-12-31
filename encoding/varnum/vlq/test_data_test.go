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

package vlq_test

import (
	"math"

	"github.com/donyori/gogo/encoding/varnum/vlq"
)

var (
	uint64s        []uint64
	encodedUint64s [][]byte

	incompleteSrcs = [][]byte{nil, {}, {0x80, 0x81}}
	tooLargeSrcs   = [][]byte{
		{0x80, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFF, 0},
		{0xFF, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0x7F},
		{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0},
		{0x80, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0x80, 0},
	}
)

func init() {
	uint64s = make([]uint64, 53)
	uint64s[1], uint64s[2], uint64s[3], uint64s[4] = 1, 7, 8, 9
	for i, x := range vlq.MinUint64s {
		uint64s[5+i*5] = x - 2
		uint64s[6+i*5] = x - 1
		uint64s[7+i*5] = x
		uint64s[8+i*5] = x + 1
		uint64s[9+i*5] = x + 2
	}
	uint64s[50] = math.MaxUint64 - 2
	uint64s[51] = math.MaxUint64 - 1
	uint64s[52] = math.MaxUint64

	encodedUint64s = make([][]byte, 53)
	for i := 0; i < 7; i++ {
		encodedUint64s[i] = []byte{byte(uint64s[i])}
	}
	for i := 1; i <= len(vlq.MinUint64s); i++ {
		for k, lastByte := range []byte{0, 1, 2, 0x7E, 0x7F} {
			idx := 2 + i*5 + k
			if idx >= 50 {
				break
			}
			encodedUint64s[idx] = make([]byte, i+1)
			for j := 0; j < i; j++ {
				if k < 3 {
					encodedUint64s[idx][j] = 0x80
				} else {
					encodedUint64s[idx][j] = 0xFF
				}
			}
			encodedUint64s[idx][i] = lastByte
		}
	}
	encodedUint64s[50] = []byte{
		0x80, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0x7D,
	}
	encodedUint64s[51] = []byte{
		0x80, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0x7E,
	}
	encodedUint64s[52] = []byte{
		0x80, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0x7F,
	}
}
