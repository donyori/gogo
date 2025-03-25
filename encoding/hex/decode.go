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

package hex

import (
	"fmt"

	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/errors"
)

// DecodedLen returns the length of decoding of x source bytes, exactly x / 2.
//
// It panics if x is negative or odd.
func DecodedLen[Int constraints.Integer](x Int) Int {
	if x < 0 {
		panic(errors.AutoMsg(fmt.Sprintf("x (%d) is negative", x)))
	} else if x&1 != 0 {
		panic(errors.AutoMsg(fmt.Sprintf("x (%d) is odd", x)))
	}
	return x >> 1
}
