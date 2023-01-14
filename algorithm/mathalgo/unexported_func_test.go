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

import (
	"fmt"
	"testing"
)

func BenchmarkGcd2Uint64Stein(b *testing.B) {
	testCases := []struct {
		a, b, want uint64
	}{
		{1, 3, 1},
		{21, 7, 7},
		{12, 18, 6},
		{4096, 6144, 2048},
		{5_274_763, 169, 13},            // 5_274_763 = 13*47*89*97, 169 = 13*13
		{5_274_763, 4_463_261, 405_751}, // 5_274_763 = 13*405_751, 4_463_261 = 11*405_751, 405_751=47*89*97
	}

	fns := []struct {
		name string
		f    func(a, b uint64) uint64
	}{
		{"Stein", gcd2Uint64Stein},
		{"SteinAnother", gcd2Uint64SteinAnother},
		{"BruteForce", gcd2Uint64BruteForce},
		{"Euclidean", gcd2Uint64Euclidean},
	}

	benchmarks := make([]struct {
		name string
		f    func(a, b uint64) uint64
		a, b uint64
		want uint64
	}, len(fns)*len(testCases))
	var idx int
	for _, tc := range testCases {
		for _, fn := range fns {
			benchmarks[idx].name = fmt.Sprintf("%s_a=%d&b=%d", fn.name, tc.a, tc.b)
			benchmarks[idx].f = fn.f
			benchmarks[idx].a = tc.a
			benchmarks[idx].b = tc.b
			benchmarks[idx].want = tc.want
			idx++
		}
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if got := bm.f(bm.a, bm.b); got != bm.want {
					b.Errorf("got %d; want %d", got, bm.want)
				}
			}
		})
	}
}

// gcd2Uint64SteinAnother is another implementation of gcd2Uint64Stein.
func gcd2Uint64SteinAnother(a, b uint64) uint64 {
	var m, n int
	for a&1 == 0 {
		a, m = a>>1, m+1
	}
	for b&1 == 0 {
		b, n = b>>1, n+1
	}
	if m > n {
		m = n
	}
	for {
		if a < b {
			a, b = b, a
		}
		a -= b
		if a == 0 {
			return b << m
		}
		for a&1 == 0 {
			a >>= 1
		}
	}
}

// gcd2Uint64BruteForce finds the greatest common divisor of
// two non-zero 64-bit unsigned integers a and b by testing
// the value from min(a, b) to zero, one by one.
//
// Caller should guarantee that both a and b are not zero.
func gcd2Uint64BruteForce(a, b uint64) uint64 {
	d := a
	if d > b {
		d = b
	}
	for d > 0 && (a%d != 0 || b%d != 0) {
		d--
	}
	return d
}

// gcd2Uint64Euclidean calculates the greatest common divisor of
// two non-zero 64-bit unsigned integers a and b with the Euclidean algorithm
// (also known as Euclid's algorithm).
//
// Caller should guarantee that both a and b are not zero.
func gcd2Uint64Euclidean(a, b uint64) uint64 {
	if a < b {
		a, b = b, a
	}
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
