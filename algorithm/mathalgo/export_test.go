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

package mathalgo

import "github.com/donyori/gogo/constraints"

// Export for testing only.

// AbsIntToUint64 takes the absolute value of the integer x
// and converts it to uint64.
func AbsIntToUint64[Int constraints.Integer](x Int) uint64 {
	return absIntToUint64(x)
}

var GCD2Uint64Stein = gcd2Uint64Stein
