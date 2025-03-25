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

package floats_test

import (
	"math"
	"testing"

	"github.com/donyori/gogo/internal/floats"
)

func TestNegZeros(t *testing.T) {
	if floats.NegZero32 != 0.0 {
		t.Error("NegZero32 is not zero")
	} else if math.Float32bits(floats.NegZero32)>>31 == 0 {
		t.Error("NegZero32 is not negative")
	}
	if floats.NegZero64 != 0.0 {
		t.Error("NegZero64 is not zero")
	} else if math.Float64bits(floats.NegZero64)>>63 == 0 {
		t.Error("NegZero64 is not negative")
	}
}

func TestInfs(t *testing.T) {
	// !(x > y) is not equivalent to (x <= y) when NaN values are involved.
	if !(floats.Inf32 > math.MaxFloat32) ||
		!math.IsInf(float64(floats.Inf32), 1) {
		t.Error("Inf32 is not positive infinity")
	}
	// !(x < y) is not equivalent to (x >= y) when NaN values are involved.
	if !(floats.NegInf32 < -math.MaxFloat32) ||
		!math.IsInf(float64(floats.NegInf32), -1) {
		t.Error("NegInf32 is not negative infinity")
	}
	if !math.IsInf(floats.Inf64, 1) {
		t.Error("Inf64 is not positive infinity")
	}
	if !math.IsInf(floats.NegInf64, -1) {
		t.Error("NegInf64 is not negative infinity")
	}
}

func TestNaNs(t *testing.T) {
	names := [][]string{
		{"NaN32A", "NaN32B", "NaN32C", "NaN32D"},
		{"NaN64A", "NaN64B", "NaN64C", "NaN64D"},
	}
	nan32s := []float32{
		floats.NaN32A, floats.NaN32B, floats.NaN32C, floats.NaN32D,
	}
	nan64s := []float64{
		floats.NaN64A, floats.NaN64B, floats.NaN64C, floats.NaN64D,
	}
	for i := range nan32s {
		if nan32s[i] == nan32s[i] || !math.IsNaN(float64(nan32s[i])) {
			t.Error(names[0][i], "is not a NaN")
		}
	}
	for i := range nan64s {
		if !math.IsNaN(nan64s[i]) {
			t.Error(names[1][i], "is not a NaN")
		}
	}
}
