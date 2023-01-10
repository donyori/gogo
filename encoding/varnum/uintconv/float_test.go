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

package uintconv_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/donyori/gogo/encoding/varnum/uintconv"
)

var float64Uint64ByteReversalPair = []struct {
	f float64
	u uint64
}{
	{0., 0},
	{1., 0xF03F},
	{2., 0x40},
	{17., 0x3140},
	{0.1, 0x9A9999999999B93F},
	{0.2, 0x9A9999999999C93F},
	{1.1, 0x9A9999999999F13F},
	{-1., 0xF0BF},
	{-2., 0xC0},
	{-17., 0x31C0},
	{-0.1, 0x9A9999999999B9BF},
	{-0.2, 0x9A9999999999C9BF},
	{-1.1, 0x9A9999999999F1BF},
	{math.MaxFloat64, 0xFFFFFFFFFFFFEF7F},
	{math.SmallestNonzeroFloat64, 0x0100000000000000},
	{-math.MaxFloat64, 0xFFFFFFFFFFFFEFFF},
	{-math.SmallestNonzeroFloat64, 0x0100000000000080},
	{math.NaN(), 0x010000000000F87F},
	{math.Inf(1), 0xF07F},
	{math.Inf(-1), 0xF0FF},
}

func TestFromFloat64ByteReversal(t *testing.T) {
	for _, pair := range float64Uint64ByteReversalPair {
		t.Run(fmt.Sprintf("f=%v(bits=%#016X)", pair.f, math.Float64bits(pair.f)), func(t *testing.T) {
			if got := uintconv.FromFloat64ByteReversal(pair.f); got != pair.u {
				t.Errorf("got %#X; want %#X", got, pair.u)
			}
		})
	}
}

func TestToFloat64ByteReversal(t *testing.T) {
	for _, pair := range float64Uint64ByteReversalPair {
		t.Run(fmt.Sprintf("u=%#X", pair.u), func(t *testing.T) {
			got := uintconv.ToFloat64ByteReversal(pair.u)
			if math.IsNaN(pair.f) {
				if !math.IsNaN(got) {
					t.Errorf("got %v (bits: %#016X); want NaN", got, math.Float64bits(got))
				}
			} else if got != pair.f {
				t.Errorf(
					"got %v (bits: %#016X); want %v (bits: %#016X)",
					got,
					math.Float64bits(got),
					pair.f,
					math.Float64bits(pair.f),
				)
			}
		})
	}
}
