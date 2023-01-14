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

package mathalgo

import "github.com/donyori/gogo/constraints"

// LCM calculates the least common multiple of the integers xs.
//
// The least common multiple of non-zero integers is the smallest
// positive integer that is divisible by each of the integers.
// In particular, if there is at least one zero,
// the least common multiple is considered zero,
// since zero is the only common multiple of zero and other integers.
//
// According to the above definition,
// LCM always returns a non-negative value,
// and it returns 0 if and only if len(xs) is 0 or
// there is at least one 0 in xs.
func LCM[Int constraints.Integer](xs ...Int) Int {
	if len(xs) == 0 {
		return 0
	}
	for _, x := range xs {
		if x == 0 {
			return 0
		}
	}
	lcm := absIntToUint64(xs[0])
	for i := 1; i < len(xs); i++ {
		x := absIntToUint64(xs[i])
		lcm = lcm / gcd2Uint64Stein(lcm, x) * x
	}
	return Int(lcm)
}
