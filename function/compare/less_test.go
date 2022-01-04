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

package compare

import (
	"math"
	"testing"
)

func TestIntLess(t *testing.T) {
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		if r := IntLess(pair[0], pair[1]); r != (pair[0] < pair[1]) {
			t.Errorf("IntLess(%d, %d): %t.", pair[0], pair[1], r)
		}
	}
}

func TestFloat64Less(t *testing.T) {
	floatPairs := [][2]float64{
		{0.0, 0.5}, {1.0, 0.5}, {0.5, 0.5},
		{math.NaN(), 0.0}, {0.0, math.NaN()}, {math.NaN(), math.NaN()},
		{math.Inf(1), 0.0}, {math.Inf(-1), 0.0},
		{0.0, math.Inf(1)}, {0.0, math.Inf(-1)},
		{math.Inf(1), math.Inf(1)}, {math.Inf(1), math.Inf(-1)},
		{math.Inf(-1), math.Inf(1)}, {math.Inf(-1), math.Inf(-1)},
		{math.NaN(), math.Inf(1)}, {math.Inf(1), math.NaN()},
		{math.NaN(), math.Inf(-1)}, {math.Inf(-1), math.NaN()},
	}
	for _, pair := range floatPairs {
		if r := Float64Less(pair[0], pair[1]); r != (pair[0] < pair[1] || (math.IsNaN(pair[0]) && !math.IsNaN(pair[1]))) {
			t.Errorf("Float64Less(%f, %f): %t.", pair[0], pair[1], r)
		}
	}
}

func TestStringLess(t *testing.T) {
	stringPairs := [][2]string{
		{"hello", "hell"}, {"hell", "hello"}, {"hello", "hello"},
	}
	for _, pair := range stringPairs {
		if r := StringLess(pair[0], pair[1]); r != (pair[0] < pair[1]) {
			t.Errorf("StringLess(%s, %s): %t.", pair[0], pair[1], r)
		}
	}
}

func TestBuiltinRealNumberLess(t *testing.T) {
	groups := [][]interface{}{
		{math.NaN()},
		{math.Inf(-1)},
		{-math.MaxFloat64},
		{float32(-math.MaxFloat32), -math.MaxFloat32},
		{int64(math.MinInt64)},
		{
			math.MinInt32, int32(math.MinInt32),
			int64(math.MinInt32), rune(math.MinInt32),
			float64(math.MinInt32),
		},
		{
			math.MinInt16, int16(math.MinInt16), int32(math.MinInt16),
			int64(math.MinInt16), rune(math.MinInt16),
			float32(math.MinInt16), float64(math.MinInt16),
		},
		{
			math.MinInt8, int8(math.MinInt8), int16(math.MinInt8),
			int32(math.MinInt8), int64(math.MinInt8), rune(math.MinInt8),
			float32(math.MinInt8), float64(math.MinInt8),
		},
		{
			-1, int8(-1), int16(-1), int32(-1), int64(-1), rune(-1),
			float32(-1.), -1.,
		},
		{float32(-math.SmallestNonzeroFloat32)}, {-math.SmallestNonzeroFloat64},
		{
			nil,
			0, int8(0), int16(0), int32(0), int64(0), rune(0),
			uint(0), uintptr(0), uint8(0), uint16(0), uint32(0), uint64(0), byte(0),
			float32(0.), 0.,
		},
		{math.SmallestNonzeroFloat64}, {float32(math.SmallestNonzeroFloat32)},
		{
			1, int8(1), int16(1), int32(1), int64(1), rune(1),
			uint(1), uintptr(1), uint8(1), uint16(1), uint32(1), uint64(1), byte(1),
			float32(1.), 1.,
		},
		{
			math.MaxInt8, int8(math.MaxInt8), int16(math.MaxInt8),
			int32(math.MaxInt8), int64(math.MaxInt8), rune(math.MaxInt8),
			uint(math.MaxInt8), uintptr(math.MaxInt8), uint8(math.MaxInt8),
			uint16(math.MaxInt8), uint32(math.MaxInt8), uint64(math.MaxInt8),
			byte(math.MaxInt8),
			float32(math.MaxInt8), float64(math.MaxInt8),
		},
		{
			math.MaxUint8, int16(math.MaxUint8), int32(math.MaxUint8),
			int64(math.MaxUint8), rune(math.MaxUint8),
			uint(math.MaxUint8), uintptr(math.MaxUint8), uint8(math.MaxUint8),
			uint16(math.MaxUint8), uint32(math.MaxUint8), uint64(math.MaxUint8),
			byte(math.MaxUint8),
			float32(math.MaxUint8), float64(math.MaxUint8),
		},
		{
			math.MaxInt16, int16(math.MaxInt16), int32(math.MaxInt16),
			int64(math.MaxInt16), rune(math.MaxInt16),
			uint(math.MaxInt16), uintptr(math.MaxInt16), uint16(math.MaxInt16),
			uint32(math.MaxInt16), uint64(math.MaxInt16),
			float32(math.MaxInt16), float64(math.MaxInt16),
		},
		{
			math.MaxUint16, int32(math.MaxUint16),
			int64(math.MaxUint16), rune(math.MaxUint16),
			uint(math.MaxUint16), uintptr(math.MaxUint16), uint16(math.MaxUint16),
			uint32(math.MaxUint16), uint64(math.MaxUint16),
			float32(math.MaxUint16), float64(math.MaxUint16),
		},
		{
			math.MaxInt32, int32(math.MaxInt32),
			int64(math.MaxInt32), rune(math.MaxInt32),
			uint(math.MaxInt32), uintptr(math.MaxInt32),
			uint32(math.MaxInt32), uint64(math.MaxInt32),
			float64(math.MaxInt32),
		},
		{
			int64(math.MaxUint32),
			uint(math.MaxUint32), uintptr(math.MaxUint32),
			uint32(math.MaxUint32), uint64(math.MaxUint32),
			float64(math.MaxUint32),
		},
		{int64(math.MaxInt64), uint64(math.MaxInt64)},
		{uint64(math.MaxUint64)},
		{float32(math.MaxFloat32), math.MaxFloat32},
		{math.MaxFloat64},
		{math.Inf(1)},
	}
	for i := range groups {
		for k := range groups {
			for _, a := range groups[i] {
				for _, b := range groups[k] {
					if r := BuiltinRealNumberLess(a, b); r != (i < k) {
						t.Errorf("BuiltinRealNumberLess(%v<%[1]T>, %v<%[2]T>): %t.", a, b, r)
					}
				}
			}
		}
	}
}

func TestLessFunc_Not(t *testing.T) {
	nLess := IntLess.Not()
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		if !IntLess(pair[0], pair[1]) != nLess(pair[0], pair[1]) {
			t.Errorf("nLess(%d, %d) != !less(%[1]d, %d).", pair[0], pair[1])
		}
	}
}

func TestLessFunc_Reverse(t *testing.T) {
	rLess := IntLess.Reverse()
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		if IntLess(pair[1], pair[0]) != rLess(pair[0], pair[1]) {
			t.Errorf("rLess(%d, %d) != less(%[2]d, %[1]d).", pair[0], pair[1])
		}
	}
}

func TestLessFunc_ToEqual(t *testing.T) {
	eq := IntLess.ToEqual()
	intPairs := [][2]int{{1, 2}, {2, 1}, {1, 1}}
	for _, pair := range intPairs {
		r := pair[0] == pair[1]
		if eq(pair[0], pair[1]) != r {
			t.Errorf("eq(%d, %d) != %t.", pair[0], pair[1], r)
		}
	}
}
