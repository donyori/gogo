// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2025  Yuan Gao
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

package floats

import "math"

// Constants exported from package math.
//
// Floating-point limit values.
// Max is the largest finite value representable by the type.
// SmallestNonzero is the smallest positive,
// nonzero value representable by the type.
const (
	MaxFloat32             = math.MaxFloat32
	SmallestNonzeroFloat32 = math.SmallestNonzeroFloat32

	MaxFloat64             = math.MaxFloat64
	SmallestNonzeroFloat64 = math.SmallestNonzeroFloat64
)

var (
	// NegZero32 is -0.0, of type float32.
	NegZero32 = math.Float32frombits(1 << 31)

	// NegZero64 is -0.0, of type float64.
	NegZero64 = math.Float64frombits(1 << 63)
)

var (
	// Inf32 is positive infinity, of type float32.
	Inf32 = math.Float32frombits(0x7F800000)

	// NegInf32 is negative infinity, of type float32.
	NegInf32 = math.Float32frombits(0xFF800000)

	// Inf64 is positive infinity, of type float64.
	Inf64 = math.Float64frombits(0x7FF0000000000000)

	// NegInf64 is negative infinity, of type float64.
	NegInf64 = math.Float64frombits(0xFFF0000000000000)
)

var (
	// NaN32A is an IEEE 754 NaN, of type float32.
	// Its IEEE 754 binary representation is 0x7FC00001.
	NaN32A = math.Float32frombits(0x7FC00001)

	// NaN32B is an IEEE 754 NaN, of type float32.
	// Its IEEE 754 binary representation is 0x7F800800.
	NaN32B = math.Float32frombits(0x7F800800)

	// NaN32C is an IEEE 754 NaN, of type float32.
	// Its IEEE 754 binary representation is 0x7F800001.
	NaN32C = math.Float32frombits(0x7F800001)

	// NaN32D is an IEEE 754 NaN, of type float32.
	// Its IEEE 754 binary representation is 0x7FC00000.
	NaN32D = math.Float32frombits(0x7FC00000)

	// NaN64A is an IEEE 754 NaN, of type float64.
	// Its IEEE 754 binary representation is 0x7FF8000000000001.
	NaN64A = math.Float64frombits(0x7FF8000000000001)

	// NaN64B is an IEEE 754 NaN, of type float64.
	// Its IEEE 754 binary representation is 0x7FF0000006000000.
	NaN64B = math.Float64frombits(0x7FF0000006000000)

	// NaN64C is an IEEE 754 NaN, of type float64.
	// Its IEEE 754 binary representation is 0x7FF0000000000001.
	NaN64C = math.Float64frombits(0x7FF0000000000001)

	// NaN64D is an IEEE 754 NaN, of type float64.
	// Its IEEE 754 binary representation is 0x7FF8000000000000.
	NaN64D = math.Float64frombits(0x7FF8000000000000)
)
