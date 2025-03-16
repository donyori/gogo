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

package mathalgo_test

import (
	"fmt"
	"math"
	"math/rand/v2"
	"testing"

	"github.com/donyori/gogo/algorithm/mathalgo"
	"github.com/donyori/gogo/constraints"
)

func TestLCM_2Int(t *testing.T) {
	testCases := make([]struct {
		a, b, want int
	}, len(testIntegers)*len(testIntegers)<<2)
	var idx int
	for _, a := range testIntegers {
		for _, b := range testIntegers {
			fsA, fsB := testIntegersFactorMap[a], testIntegersFactorMap[b]
			want := 1
			for i := range NumTestIntegerPrimeFactors {
				if fsA[i] >= fsB[i] {
					want *= fsA[i]
				} else {
					want *= fsB[i]
				}
			}
			testCases[idx] = struct{ a, b, want int }{a: a, b: b, want: want}
			testCases[idx+1] = struct{ a, b, want int }{a: a, b: -b, want: want}
			testCases[idx+2] = struct{ a, b, want int }{a: -a, b: b, want: want}
			testCases[idx+3] = struct{ a, b, want int }{a: -a, b: -b, want: want}
			idx += 4
		}
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("a=%d&b=%d", tc.a, tc.b), func(t *testing.T) {
			got := mathalgo.LCM(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("got %d; want %d", got, tc.want)
			}
		})
	}
}

func TestLCM_0(t *testing.T) {
	xss := [][]int{
		nil,
		{},
		{0},
		{0, 0},
		{0, 6},
		{6, 0},
		{0, 0, 6},
		{0, 6, 0},
		{6, 0, 0},
		{0, 0, 6, 0},
		{0, 6, 0, 0},
		{0, 12, 18},
		{12, 0, 18},
		{12, 18, 0},
		{0, 0, 12, 18},
		{0, 0, 12, 0, 0, 18, 0},
		{0, 12, 18, 21},
		{12, 0, 18, 21},
		{12, 18, 0, 21},
		{12, 18, 21, 0},
		{0, 0, 12, 18, 21},
		{0, 0, 12, 0, 0, 18, 0, 21, 0},
	}

	for _, xs := range xss {
		t.Run("xs="+xsToName(xs), func(t *testing.T) {
			got := mathalgo.LCM(xs...)
			if got != 0 {
				t.Errorf("got %d; want 0", got)
			}
		})
	}
}

func TestLCM_RandomlySelectInt(t *testing.T) {
	const N int = 100
	xsNameSet := make(map[string]struct{}, N)
	xsNameSet[""] = struct{}{}
	random := rand.New(rand.NewChaCha8(ChaCha8Seed))
	for range N {
		xs, xsName := randomlySelectInts(t, random, xsNameSet)
		if t.Failed() {
			return
		}

		var maxF2, maxF3, maxF5 int
		for _, x := range xs {
			if x < 0 {
				x = -x
			}
			fs := testIntegersFactorMap[x]
			if maxF2 < fs[0] {
				maxF2 = fs[0]
			}
			if maxF3 < fs[1] {
				maxF3 = fs[1]
			}
			if maxF5 < fs[2] {
				maxF5 = fs[2]
			}
		}
		want := maxF2 * maxF3 * maxF5

		t.Run("xs="+xsName, func(t *testing.T) {
			got := mathalgo.LCM(xs...)
			if got != want {
				t.Errorf("got %d; want %d", got, want)
			}
		})
	}
}

func TestLCM_AllRandom(t *testing.T) {
	const N int = 100
	xsNameSet := make(map[string]struct{}, N)
	xsNameSet[""] = struct{}{}
	var wantPositive bool
	random := rand.New(rand.NewChaCha8(ChaCha8Seed))
	for range N {
		xs, xsName := randomlyGenerateInts(t, random, xsNameSet)
		if t.Failed() {
			return
		}

		want := lcmBruteForce(xs...)
		if want > 0 {
			wantPositive = true
		}

		t.Run("xs="+xsName, func(t *testing.T) {
			got := mathalgo.LCM(xs...)
			if got != want {
				t.Errorf("got %d; want %d", got, want)
			}
		})
	}

	if !wantPositive {
		t.Error("all cases want 0")
	}
}

func TestLCM_Type(t *testing.T) {
	xss := []any{
		[]int{6, -8, 10},
		[]int8{6, -8, 10},
		[]int16{6, -8, 10},
		[]int32{6, -8, 10},
		[]int64{6, -8, 10},
		[]uint{6, 8, 10},
		[]uint8{6, 8, 10},
		[]uint16{6, 8, 10},
		[]uint32{6, 8, 10},
		[]uint64{6, 8, 10},
		[]uintptr{6, 8, 10},
	}
	const Want uint64 = 120

	for _, xsAny := range xss {
		t.Run(fmt.Sprintf("xs-type=%T", xsAny), func(t *testing.T) {
			var got uint64
			switch xs := xsAny.(type) {
			case []int:
				got = uint64(mathalgo.LCM(xs...))
			case []int8:
				got = uint64(mathalgo.LCM(xs...))
			case []int16:
				got = uint64(mathalgo.LCM(xs...))
			case []int32:
				got = uint64(mathalgo.LCM(xs...))
			case []int64:
				got = uint64(mathalgo.LCM(xs...))
			case []uint:
				got = uint64(mathalgo.LCM(xs...))
			case []uint8:
				got = uint64(mathalgo.LCM(xs...))
			case []uint16:
				got = uint64(mathalgo.LCM(xs...))
			case []uint32:
				got = uint64(mathalgo.LCM(xs...))
			case []uint64:
				got = mathalgo.LCM(xs...)
			case []uintptr:
				got = uint64(mathalgo.LCM(xs...))
			default:
				// This should never happen, but will act as a safeguard for later.
				t.Fatal("type of xs is unacceptable")
			}

			if got != Want {
				t.Errorf("got %d; want %d", got, Want)
			}
		})
	}
}

// lcmBruteForce finds the least common multiple of the integers xs
// by testing the value from the maximum value in xs
// to the maximum value of Int, one by one.
//
// The definition of the least common multiple here
// is the same as function LCM.
//
// It returns 0 if len(xs) is 0.
func lcmBruteForce[Int constraints.Integer](xs ...Int) Int {
	if len(xs) == 0 || xs[0] == 0 {
		return 0
	}
	m := mathalgo.AbsIntToUint64(xs[0])
	for i := 1; i < len(xs); i++ {
		if xs[i] == 0 {
			return 0
		}
		x := mathalgo.AbsIntToUint64(xs[i])
		if m < x {
			m = x
		}
	}

	limit := getLimitAccordingToType(xs...)
	for m < limit {
		var i int
		for i < len(xs) && m%mathalgo.AbsIntToUint64(xs[i]) == 0 {
			i++
		}
		if i >= len(xs) {
			return Int(m)
		}
		m++
	}
	for _, x := range xs {
		if m%mathalgo.AbsIntToUint64(x) != 0 {
			panic("lcm overflows")
		}
	}
	return Int(m)
}

// getLimitAccordingToType returns the maximum value
// of the corresponding integer type.
// The result is of type uint64.
func getLimitAccordingToType[Int constraints.Integer](xs ...Int) uint64 {
	var limit uint64
	switch any(xs).(type) {
	case []int:
		limit = math.MaxInt
	case []int8:
		limit = math.MaxInt8
	case []int16:
		limit = math.MaxInt16
	case []int32:
		limit = math.MaxInt32
	case []int64:
		limit = math.MaxInt64
	case []uint:
		limit = math.MaxUint
	case []uint8:
		limit = math.MaxUint8
	case []uint16:
		limit = math.MaxUint16
	case []uint32:
		limit = math.MaxUint32
	case []uint64, []uintptr:
		limit = math.MaxUint64
	default:
		// This should never happen, but will act as a safeguard for later.
		panic("type of xs is unacceptable")
	}
	return limit
}
