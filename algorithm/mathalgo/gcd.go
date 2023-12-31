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

import (
	"math/bits"

	"github.com/donyori/gogo/constraints"
)

// GCD calculates the greatest common divisor of the integers xs.
//
// The greatest common divisor of integers, which are not all zero,
// is the largest positive integer that divides each of the integers
// (zero is considered divisible by everything).
// In particular, if there is no non-zero integer,
// the greatest common divisor is considered zero.
//
// According to the above definition, GCD always returns a non-negative value,
// and it returns 0 if and only if there is no non-zero value in xs.
func GCD[Int constraints.Integer](xs ...Int) Int {
	var i int
	for i < len(xs) && xs[i] == 0 {
		i++
	}
	if i >= len(xs) {
		return 0
	}
	gcd := absIntToUint64(xs[i])
	for i++; i < len(xs); i++ {
		if xs[i] != 0 {
			gcd = gcd2Uint64Stein(absIntToUint64(xs[i]), gcd)
		}
	}
	return Int(gcd)
}

// absIntToUint64 takes the absolute value of the integer x
// and converts it to uint64.
func absIntToUint64[Int constraints.Integer](x Int) uint64 {
	if x < 0 {
		x = -x
	}
	return uint64(x)
}

// gcd2Uint64Stein calculates the greatest common divisor of
// two non-zero 64-bit unsigned integers a and b with Stein's algorithm
// (also known as the binary GCD algorithm or the binary Euclidean algorithm).
//
// gcd2Uint64Stein always returns a non-zero value.
//
// Caller should guarantee that both a and b are not zero.
func gcd2Uint64Stein(a, b uint64) uint64 {
	m, n := bits.TrailingZeros64(a), bits.TrailingZeros64(b)
	a, b = a>>m, b>>n
	if m > n {
		m = n // let m be min(m, n)
	}
	for {
		if a < b {
			a, b = b, a // let a, b be max(a, b), min(a, b) respectively
		}
		a -= b
		if a == 0 {
			return b << m
		}
		a >>= bits.TrailingZeros64(a)
	}
}
