// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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

package compare_test

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
	NaN32A = math.Float32frombits(0x7F800001)

	// NaN32B is another IEEE 754 NaN, of type float32.
	NaN32B = math.Float32frombits(0x7FC00000)

	// NaN64A is an IEEE 754 NaN, of type float64.
	NaN64A = math.Float64frombits(0x7FF0000000000001)

	// NaN64B is another IEEE 754 NaN, of type float64.
	NaN64B = math.Float64frombits(0x7FF8000000000000)
)

func init() {
	// Check the above values.
	initCheckNegZeros()
	initCheckInfs()
	initCheckNaNs()
}

// initCheckNegZeros checks NegZero32 and NegZero64.
// It panics if something is wrong.
func initCheckNegZeros() {
	switch {
	case NegZero32 != 0.0:
		panic("NegZero32 is not zero")
	case math.Float32bits(NegZero32)>>31 == 0:
		panic("NegZero32 is not negative")
	case NegZero64 != 0.0:
		panic("NegZero64 is not zero")
	case math.Float64bits(NegZero64)>>63 == 0:
		panic("NegZero64 is not negative")
	}
}

// initCheckInfs checks Inf32, NegInf32, Inf64, and NegInf64.
// It panics if something is wrong.
func initCheckInfs() {
	switch {
	// !(x > y) is not equivalent to (x <= y) when NaN values are involved.
	case !(Inf32 > MaxFloat32), !math.IsInf(float64(Inf32), 1):
		panic("Inf32 is not positive infinity")
	// !(x < y) is not equivalent to (x >= y) when NaN values are involved.
	case !(NegInf32 < -MaxFloat32), !math.IsInf(float64(NegInf32), -1):
		panic("NegInf32 is not negative infinity")
	case !math.IsInf(Inf64, 1):
		panic("Inf64 is not positive infinity")
	case !math.IsInf(NegInf64, -1):
		panic("NegInf64 is not negative infinity")
	}
}

// initCheckNaNs checks NaN32A, NaN32B, NaN64A, and NaN64B.
// It panics if something is wrong.
func initCheckNaNs() {
	switch {
	case NaN32A == NaN32A, !math.IsNaN(float64(NaN32A)):
		panic("NaN32A is not a NaN")
	case NaN32B == NaN32B, !math.IsNaN(float64(NaN32B)):
		panic("NaN32B is not a NaN")
	case !math.IsNaN(NaN64A):
		panic("NaN64A is not a NaN")
	case !math.IsNaN(NaN64B):
		panic("NaN64B is not a NaN")
	}
}
