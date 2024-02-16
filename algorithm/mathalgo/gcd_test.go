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
	"github.com/donyori/gogo/fmtcoll"
)

const NumTestIntegerPrimeFactors int = 3

// ChaCha8Seed is the seed for ChaCha8 used for testing.
var ChaCha8Seed = [32]byte([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"))

var (
	testIntegers          []int
	testIntegersFactorMap map[int][NumTestIntegerPrimeFactors]int
)

func init() {
	const MaxD int = 2
	numInts := 1
	for range NumTestIntegerPrimeFactors {
		numInts *= MaxD + 1
	}
	testIntegers = make([]int, numInts)
	testIntegersFactorMap = make(
		map[int][NumTestIntegerPrimeFactors]int, numInts)
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
				testIntegersFactorMap[x] = [NumTestIntegerPrimeFactors]int{
					d2s[d2],
					d3s[d3],
					d5s[d5],
				}
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
			for i := range NumTestIntegerPrimeFactors {
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

func TestGCD_RandomlySelectInt(t *testing.T) {
	const N int = 100
	xsNameSet := make(map[string]struct{}, N)
	xsNameSet[""] = struct{}{}
	random := rand.New(rand.NewChaCha8(ChaCha8Seed))
	for range N {
		xs, xsName := randomlySelectInts(t, random, xsNameSet)
		if t.Failed() {
			return
		}

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
	xsNameSet := make(map[string]struct{}, N)
	xsNameSet[""] = struct{}{}
	var isWantMoreThan1 bool
	random := rand.New(rand.NewChaCha8(ChaCha8Seed))
	for range N {
		xs, xsName := randomlyGenerateInts(t, random, xsNameSet)
		if t.Failed() {
			return
		}

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

func TestGCD2Uint64Functions(t *testing.T) {
	fns := []struct {
		name string
		f    func(a, b uint64) uint64
	}{
		{"Stein", mathalgo.GCD2Uint64Stein},
		{"SteinAnother", gcd2Uint64SteinAnother},
		{"BruteForce", gcd2Uint64BruteForce},
		{"Euclidean", gcd2Uint64Euclidean},
	}

	testCases := make([]struct {
		a, b, want uint64
	}, 6+len(testIntegers)*len(testIntegers))
	// The following 6 cases are used in the benchmark.
	testCases[0] = struct{ a, b, want uint64 }{a: 1, b: 7, want: 1}
	testCases[1] = struct{ a, b, want uint64 }{a: 3 * 7, b: 7, want: 7}
	testCases[2] = struct{ a, b, want uint64 }{a: 2 * 6, b: 3 * 6, want: 6}
	testCases[3] = struct{ a, b, want uint64 }{a: 2 * 2048, b: 3 * 2048, want: 2048}
	testCases[4] = struct{ a, b, want uint64 }{a: 13 * 405_751, b: 13 * 13, want: 13}
	testCases[5] = struct{ a, b, want uint64 }{a: 13 * 405_751, b: 11 * 405_751, want: 405_751}
	idx := 6
	for _, a := range testIntegers {
		for _, b := range testIntegers {
			fsA, fsB := testIntegersFactorMap[a], testIntegersFactorMap[b]
			var want uint64 = 1
			for i := range NumTestIntegerPrimeFactors {
				if fsA[i] <= fsB[i] {
					want *= uint64(fsA[i])
				} else {
					want *= uint64(fsB[i])
				}
			}
			testCases[idx].a = uint64(a)
			testCases[idx].b = uint64(b)
			testCases[idx].want = want
			idx++
		}
	}

	for _, fn := range fns {
		t.Run(fn.name, func(t *testing.T) {
			for _, tc := range testCases {
				t.Run(fmt.Sprintf("a=%d&b=%d", tc.a, tc.b), func(t *testing.T) {
					got := fn.f(tc.a, tc.b)
					if got != tc.want {
						t.Errorf("got %d; want %d", got, tc.want)
					}
				})
			}
		})
	}
}

func BenchmarkGCD2Uint64Functions(b *testing.B) {
	dataList := []struct {
		a, b uint64
	}{
		{1, 7},
		{3 * 7, 7},
		{2 * 6, 3 * 6},
		{2 * 2048, 3 * 2048},
		{13 * 405_751, 13 * 13},
		{13 * 405_751, 11 * 405_751},
	}

	fns := []struct {
		name string
		f    func(a, b uint64) uint64
	}{
		{"Stein", mathalgo.GCD2Uint64Stein},
		{"SteinAnother", gcd2Uint64SteinAnother},
		{"BruteForce", gcd2Uint64BruteForce},
		{"Euclidean", gcd2Uint64Euclidean},
	}

	for _, data := range dataList {
		b.Run(fmt.Sprintf("a=%d&b=%d", data.a, data.b), func(b *testing.B) {
			for _, fn := range fns {
				b.Run(fn.name, func(b *testing.B) {
					for range b.N {
						fn.f(data.a, data.b)
					}
				})
			}
		})
	}
}

func xsToName[Int constraints.Integer](xs []Int) string {
	return fmtcoll.MustFormatSliceToString(xs, &fmtcoll.SequenceFormat[Int]{
		CommonFormat: fmtcoll.CommonFormat{
			Separator: ",",
		},
		FormatItemFn: fmtcoll.FprintfToFormatFunc[Int]("%d"),
	})
}

// randomlySelectInts selects 1 to 64 integers from testIntegers randomly.
//
// random is the source of random numbers.
//
// xsNameSet is the set to record generated integer lists.
//
// If there is something wrong,
// randomlySelectInts reports the error using t.Errorf and returns zero values.
func randomlySelectInts(
	t *testing.T,
	random *rand.Rand,
	xsNameSet map[string]struct{},
) (xs []int, xsName string) {
	const MaxTry int = 100
	const MaxLen int = 64
	xs = make([]int, 1+random.IntN(MaxLen))
	_, duplicated := xsNameSet[xsName]
	for try := 0; duplicated && try < MaxTry; try++ {
		for i := range xs {
			xs[i] = testIntegers[random.IntN(len(testIntegers))]
			if random.Uint64()&1 == 0 {
				xs[i] = -xs[i]
			}
		}
		xsName = xsToName(xs)
		_, duplicated = xsNameSet[xsName]
	}
	if _, duplicated = xsNameSet[xsName]; duplicated {
		t.Errorf("try %d times but always get tested xs", MaxTry)
		return nil, ""
	}
	xsNameSet[xsName] = struct{}{}
	return
}

// randomlyGenerateInts generates 1 to 4 integers randomly.
// Each integer is in the interval [0, 127].
//
// random is the source of random numbers.
//
// xsNameSet is the set to record generated integer lists.
//
// If there is something wrong,
// randomlyGenerateInts reports the error using t.Errorf
// and returns zero values.
func randomlyGenerateInts(
	t *testing.T,
	random *rand.Rand,
	xsNameSet map[string]struct{},
) (xs []int, xsName string) {
	const MaxTry int = 100
	const MaxLen int = 4
	const MaxX int = 127
	xs = make([]int, 1+random.IntN(MaxLen))
	_, duplicated := xsNameSet[xsName]
	for try := 0; duplicated && try < MaxTry; try++ {
		for i := range xs {
			xs[i] = random.IntN(MaxX + 1)
		}
		xsName = xsToName(xs)
		_, duplicated = xsNameSet[xsName]
	}
	if _, duplicated = xsNameSet[xsName]; duplicated {
		t.Errorf("try %d times but always get tested xs", MaxTry)
		return nil, ""
	}
	xsNameSet[xsName] = struct{}{}
	return
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

// gcd2Uint64SteinAnother is another implementation of
// function gcd2Uint64Stein.
// It uses for loop instead of math/bits.TrailingZeros64
// to remove trailing zeros.
//
// Caller should guarantee that both a and b are not zero.
// If a or b is zero, gcd2Uint64SteinAnother falls into an infinite loop.
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
