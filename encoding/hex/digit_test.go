// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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

import (
	"fmt"
	"math"
	"strconv"
	"testing"
)

func TestEncodeInt64ToString(t *testing.T) {
	f := func(x int64, upper bool, digits int) string {
		layout := "%"
		if digits > 0 {
			width := digits
			if x < 0 {
				width++ // Increase one for the negative sign.
			}
			layout += "0" + strconv.Itoa(width)
		}
		if upper {
			layout += "X"
		} else {
			layout += "x"
		}
		return fmt.Sprintf(layout, x)
	}
	xs := []int64{
		0, 1, 2, 7, 8, 15, 16, 31, 32, 63, 64, 1234567890, math.MaxInt64,
		-1, -2, -7, -8, -15, -16, -31, -32, -63, -64, -1234567890, math.MinInt64 + 1, math.MinInt64,
	}
	uppers := []bool{false, true}
	digitsValues := []int{-1, 0, 2, 4, 8, 9, 15, 16, 17, 18, 500}
	for _, x := range xs {
		for _, upper := range uppers {
			for _, digits := range digitsValues {
				r := EncodeInt64ToString(x, upper, digits)
				wanted := f(x, upper, digits)
				if r != wanted {
					t.Errorf("x: %d, upper: %t, digits: %d, r != wanted.\n\tr: %s\n\twanted: %s", x, upper, digits, r, wanted)
				}
			}
		}
	}
}
