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

package uintconv

import "math"

// FromFloat64ByteReversal maps a 64-bit floating-point number
// to a 64-bit unsigned integer.
//
// It converts the specified float64 to uint64 using math.Float64bits
// and then returns the byte-reversal of that uint64.
//
// FromFloat64ByteReversal(ToFloat64ByteReversal(x)) == x.
func FromFloat64ByteReversal(f float64) uint64 {
	return uint64ByteReverse(math.Float64bits(f))
}

// ToFloat64ByteReversal maps a 64-bit unsigned integer back to
// a 64-bit floating-point number.
//
// It converts the byte-reversal of the specified uint64 to float64
// using math.Float64frombits.
//
// ToFloat64ByteReversal(FromFloat64ByteReversal(x)) == x.
func ToFloat64ByteReversal(u uint64) float64 {
	return math.Float64frombits(uint64ByteReverse(u))
}

// uint64ByteReverse returns the byte-reversal of u.
func uint64ByteReverse(u uint64) uint64 {
	var x uint64
	for i := 0; i < 64; i += 8 {
		b := 0xFF << i & u
		offset := (28 - i) << 1
		if offset > 0 {
			x |= b << offset
		} else {
			x |= b >> -offset
		}
	}
	return x
}
