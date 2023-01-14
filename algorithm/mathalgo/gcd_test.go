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

package mathalgo_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/donyori/gogo/algorithm/mathalgo"
	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/internal/testaux"
)

const NumTestIntegerPrimeFactors int = 3

var (
	testIntegers          []int
	testIntegersFactorMap map[int][NumTestIntegerPrimeFactors]int
)

func init() {
	const MaxD int = 2
	numInts := 1
	for i := 0; i < NumTestIntegerPrimeFactors; i++ {
		numInts *= MaxD + 1
	}
	testIntegers = make([]int, numInts)
	testIntegersFactorMap = make(map[int][NumTestIntegerPrimeFactors]int, numInts)
	d2s := [MaxD + 1]int{1, 2, 4}
	d3s := [MaxD + 1]int{1, 3, 9}
	d5s := [MaxD + 1]int{1, 5, 25}
	var idx int
	for d2 := 0; d2 <= MaxD; d2++ {
		for d3 := 0; d3 <= MaxD; d3++ {
			for d5 := 0; d5 <= MaxD; d5++ {
				if idx >= len(testIntegers) {
					panic("not enough test integers, please update")
				}
				x := d2s[d2] * d3s[d3] * d5s[d5]
				testIntegers[idx] = x
				testIntegersFactorMap[x] = [NumTestIntegerPrimeFactors]int{d2s[d2], d3s[d3], d5s[d5]}
				idx++
			}
		}
	}
	if idx != len(testIntegers) {
		panic("excessive test integers, please update")
	}
}

func TestGCD_2Int(t *testing.T) {
	testCases := make([]struct {
		a, b, want int
	}, len(testIntegers)*len(testIntegers)*4)
	var idx int
	for _, a := range testIntegers {
		for _, b := range testIntegers {
			fsA, fsB := testIntegersFactorMap[a], testIntegersFactorMap[b]
			want := 1
			for i := 0; i < NumTestIntegerPrimeFactors; i++ {
				if fsA[i] <= fsB[i] {
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
			got := mathalgo.GCD(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("got %d; want %d", got, tc.want)
			}
		})
	}
}

func TestGCD_0(t *testing.T) {
	testCases := []struct {
		xs   []int
		want int
	}{
		{nil, 0},
		{[]int{}, 0},
		{[]int{0}, 0},
		{[]int{0, 0}, 0},
		{[]int{0, 6}, 6},
		{[]int{6, 0}, 6},
		{[]int{0, 0, 6}, 6},
		{[]int{0, 6, 0}, 6},
		{[]int{6, 0, 0}, 6},
		{[]int{0, 0, 6, 0}, 6},
		{[]int{0, 6, 0, 0}, 6},
		{[]int{0, 12, 18}, 6},
		{[]int{12, 0, 18}, 6},
		{[]int{12, 18, 0}, 6},
		{[]int{0, 0, 12, 18}, 6},
		{[]int{0, 0, 12, 0, 0, 18, 0}, 6},
		{[]int{0, 12, 18, 21}, 3},
		{[]int{12, 0, 18, 21}, 3},
		{[]int{12, 18, 0, 21}, 3},
		{[]int{12, 18, 21, 0}, 3},
		{[]int{0, 0, 12, 18, 21}, 3},
		{[]int{0, 0, 12, 0, 0, 18, 0, 21, 0}, 3},
	}

	for _, tc := range testCases {
		t.Run("xs="+xsToName(tc.xs), func(t *testing.T) {
			got := mathalgo.GCD(tc.xs...)
			if got != tc.want {
				t.Errorf("got %d; want %d", got, tc.want)
			}
		})
	}
}

func TestGCD_RandomSelectInt(t *testing.T) {
	const N int = 100
	const MaxTry int = 100
	const MaxLen int = 64

	nameSet := make(map[string]bool, N)
	nameSet[""] = true
	random := rand.New(rand.NewSource(20))

	for i := 0; i < N; i++ {
		var n int
		for try := 0; n == 0 && try < MaxTry; try++ {
			n = random.Intn(MaxLen + 1)
		}
		if n == 0 {
			t.Fatalf("try %d times but always get n as 0", MaxTry)
		}

		xs := make([]int, n)
		var xsName string
		for try := 0; nameSet[xsName] && try < MaxTry; try++ {
			for i := range xs {
				xs[i] = testIntegers[random.Intn(len(testIntegers))]
				if random.Int31n(2) == 0 {
					xs[i] = -xs[i]
				}
			}
			xsName = xsToName(xs)
		}
		if nameSet[xsName] {
			t.Fatalf("try %d times but always get tested xs", MaxTry)
		}
		nameSet[xsName] = true

		minF2, minF3, minF5 := math.MaxInt, math.MaxInt, math.MaxInt
		for _, x := range xs {
			if x < 0 {
				x = -x
			}
			fs := testIntegersFactorMap[x]
			if minF2 > fs[0] {
				minF2 = fs[0]
			}
			if minF3 > fs[1] {
				minF3 = fs[1]
			}
			if minF5 > fs[2] {
				minF5 = fs[2]
			}
		}
		want := minF2 * minF3 * minF5

		t.Run("xs="+xsName, func(t *testing.T) {
			got := mathalgo.GCD(xs...)
			if got != want {
				t.Errorf("got %d; want %d", got, want)
			}
		})
	}
}

func TestGCD_AllRandom(t *testing.T) {
	const N int = 100
	const MaxTry int = 100
	const MaxLen int = 4
	const MaxX int = 32

	nameSet := make(map[string]bool, N)
	nameSet[""] = true
	var isWantMoreThan1 bool
	random := rand.New(rand.NewSource(30))

	for i := 0; i < N; i++ {
		var n int
		for try := 0; n == 0 && try < MaxTry; try++ {
			n = random.Intn(MaxLen + 1)
		}
		if n == 0 {
			t.Fatalf("try %d times but always get n as 0", MaxTry)
		}

		xs := make([]int, n)
		var xsName string
		for try := 0; nameSet[xsName] && try < MaxTry; try++ {
			for i := range xs {
				xs[i] = random.Intn(MaxX + 1)
			}
			xsName = xsToName(xs)
		}
		if nameSet[xsName] {
			t.Fatalf("try %d times but always get tested xs", MaxTry)
		}
		nameSet[xsName] = true

		want := gcdBruteForce(xs...)
		if want > 1 {
			isWantMoreThan1 = true
		}

		t.Run("xs="+xsName, func(t *testing.T) {
			got := mathalgo.GCD(xs...)
			if got != want {
				t.Errorf("got %d; want %d", got, want)
			}
		})
	}

	if !isWantMoreThan1 {
		t.Error("all cases want 0 or 1")
	}
}

func TestGCD_Type(t *testing.T) {
	xss := []any{
		[]int{12, -16, 0, 20},
		[]int8{12, -16, 0, 20},
		[]int16{12, -16, 0, 20},
		[]int32{12, -16, 0, 20},
		[]int64{12, -16, 0, 20},
		[]uint{12, 16, 0, 20},
		[]uint8{12, 16, 0, 20},
		[]uint16{12, 16, 0, 20},
		[]uint32{12, 16, 0, 20},
		[]uint64{12, 16, 0, 20},
		[]uintptr{12, 16, 0, 20},
	}
	const Want uint64 = 4

	for _, xsAny := range xss {
		t.Run(fmt.Sprintf("xs-type=%T", xsAny), func(t *testing.T) {
			var got uint64
			switch xs := xsAny.(type) {
			case []int:
				got = uint64(mathalgo.GCD(xs...))
			case []int8:
				got = uint64(mathalgo.GCD(xs...))
			case []int16:
				got = uint64(mathalgo.GCD(xs...))
			case []int32:
				got = uint64(mathalgo.GCD(xs...))
			case []int64:
				got = uint64(mathalgo.GCD(xs...))
			case []uint:
				got = uint64(mathalgo.GCD(xs...))
			case []uint8:
				got = uint64(mathalgo.GCD(xs...))
			case []uint16:
				got = uint64(mathalgo.GCD(xs...))
			case []uint32:
				got = uint64(mathalgo.GCD(xs...))
			case []uint64:
				got = mathalgo.GCD(xs...)
			case []uintptr:
				got = uint64(mathalgo.GCD(xs...))
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

func xsToName[Int constraints.Integer](xs []Int) string {
	return testaux.SliceToName(xs, ",", "%d", false)
}

// gcdBruteForce finds the greatest common divisor of the integers xs
// by testing the value from the minimum value in xs to zero, one by one.
//
// The definition of the greatest common divisor here
// is the same as function GCD.
//
// It returns 0 if len(xs) is 0.
func gcdBruteForce[Int constraints.Integer](xs ...Int) Int {
	if len(xs) == 0 {
		return 0
	}
	d := mathalgo.AbsIntToUint64(xs[0])
	for i := 1; i < len(xs); i++ {
		x := mathalgo.AbsIntToUint64(xs[i])
		if d == 0 || x != 0 && d > x {
			d = x
		}
	}
	for d > 0 {
		var i int
		for i < len(xs) && mathalgo.AbsIntToUint64(xs[i])%d == 0 {
			i++
		}
		if i >= len(xs) {
			break
		}
		d--
	}
	return Int(d)
}
