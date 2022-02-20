// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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

package vlq

import "math"

var (
	testUint64s        []uint64
	testEncodedUint64s [][]byte

	testIncompleteSrcs = [][]byte{nil, {}, {0x80, 0x81}}
	testTooLargeSrcs   = [][]byte{
		{0x80, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFF, 0},
		{0xFF, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0x7F},
		{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0},
		{0x80, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0x80, 0},
	}
)

func init() {
	testUint64s = make([]uint64, 53)
	testUint64s[1], testUint64s[2], testUint64s[3], testUint64s[4] = 1, 7, 8, 9
	for i, x := range minUint64s {
		testUint64s[5+i*5] = x - 2
		testUint64s[6+i*5] = x - 1
		testUint64s[7+i*5] = x
		testUint64s[8+i*5] = x + 1
		testUint64s[9+i*5] = x + 2
	}
	testUint64s[50], testUint64s[51], testUint64s[52] = math.MaxUint64-2, math.MaxUint64-1, math.MaxUint64

	testEncodedUint64s = make([][]byte, 53)
	for i := 0; i < 7; i++ {
		testEncodedUint64s[i] = []byte{byte(testUint64s[i])}
	}
	for i := 1; i <= len(minUint64s); i++ {
		for k, lastByte := range []byte{0, 1, 2, 0x7E, 0x7F} {
			idx := 2 + i*5 + k
			if idx >= 50 {
				break
			}
			testEncodedUint64s[idx] = make([]byte, i+1)
			for j := 0; j < i; j++ {
				if k < 3 {
					testEncodedUint64s[idx][j] = 0x80
				} else {
					testEncodedUint64s[idx][j] = 0xFF
				}
			}
			testEncodedUint64s[idx][i] = lastByte
		}
	}
	testEncodedUint64s[50] = []byte{0x80, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0x7D}
	testEncodedUint64s[51] = []byte{0x80, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0x7E}
	testEncodedUint64s[52] = []byte{0x80, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0xFE, 0x7F}
}
