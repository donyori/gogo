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

package uintconv

import (
	"math"
	"math/bits"
)

// FromFloat64ByteReversal maps a 64-bit floating-point number
// to a 64-bit unsigned integer.
//
// It converts the specified float64 to uint64 using math.Float64bits
// and then returns the byte-reversal of that uint64.
//
// It satisfies:
//
//	FromFloat64ByteReversal(ToFloat64ByteReversal(x)) == x.
func FromFloat64ByteReversal(f float64) uint64 {
	return bits.ReverseBytes64(math.Float64bits(f))
}

// ToFloat64ByteReversal maps a 64-bit unsigned integer back to
// a 64-bit floating-point number.
//
// It converts the byte-reversal of the specified uint64 to float64
// using math.Float64frombits.
//
// It satisfies:
//
//	ToFloat64ByteReversal(FromFloat64ByteReversal(x)) == x.
func ToFloat64ByteReversal(u uint64) float64 {
	return math.Float64frombits(bits.ReverseBytes64(u))
}
