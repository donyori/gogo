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

package hex_test

import (
	stdhex "encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/donyori/gogo/encoding/hex"
)

func TestCanEncodeTo(t *testing.T) {
	for _, srcCase := range testEncodeCases {
		if srcCase.upper { // only use the lower cases to avoid redundant sources
			continue
		}
		xSet := make(map[string]struct{}, len(testEncodeCases))
		for _, xCase := range testEncodeCases {
			if _, ok := xSet[xCase.dstStr]; ok {
				continue
			}
			xSet[xCase.dstStr] = struct{}{}
			want := srcCase.srcStr == xCase.srcStr
			t.Run(
				fmt.Sprintf(
					"src=%s&x=%s&upper=%t",
					srcCase.srcName,
					xCase.dstName,
					xCase.upper,
				),
				func(t *testing.T) {
					t.Run("srcType=[]byte&xType=[]byte", func(t *testing.T) {
						got := hex.CanEncodeTo(srcCase.srcBytes, xCase.dstBytes)
						if got != want {
							t.Errorf("got %t; want %t", got, want)
						}
					})
					t.Run("srcType=[]byte&xType=string", func(t *testing.T) {
						got := hex.CanEncodeTo(srcCase.srcBytes, xCase.dstStr)
						if got != want {
							t.Errorf("got %t; want %t", got, want)
						}
					})
					t.Run("srcType=string&xType=[]byte", func(t *testing.T) {
						got := hex.CanEncodeTo(srcCase.srcStr, xCase.dstBytes)
						if got != want {
							t.Errorf("got %t; want %t", got, want)
						}
					})
					t.Run("srcType=string&xType=string", func(t *testing.T) {
						got := hex.CanEncodeTo(srcCase.srcStr, xCase.dstStr)
						if got != want {
							t.Errorf("got %t; want %t", got, want)
						}
					})
				},
			)
		}
	}
}

func TestCanEncodeTo_LetterCaseDiff(t *testing.T) {
	skipAll := true
	xSet := make(map[string]struct{}, len(testEncodeCases))
	for _, tc := range testEncodeCases {
		if len(tc.dstStr) == 0 {
			continue
		}
		x := make([]byte, len(tc.dstBytes))
		copy(x, tc.dstBytes)
		skip := true
		for i := range x {
			if x[i] <= '9' {
				x[i] ^= hex.LetterCaseDiff
				skip = false
			}
		}
		xStr := string(x)
		if skip {
			continue
		} else if _, ok := xSet[xStr]; ok {
			continue
		}
		xSet[xStr] = struct{}{}
		t.Run(
			fmt.Sprintf(
				"src=%s&x=%s&upper=%t&numeric-xor-%#x",
				tc.srcName,
				stringName(xStr),
				tc.upper,
				hex.LetterCaseDiff,
			),
			func(t *testing.T) {
				t.Run("srcType=[]byte&xType=[]byte", func(t *testing.T) {
					if hex.CanEncodeTo(tc.srcBytes, x) {
						t.Error("got true; want false")
					}
				})
				t.Run("srcType=[]byte&xType=string", func(t *testing.T) {
					if hex.CanEncodeTo(tc.srcBytes, xStr) {
						t.Error("got true; want false")
					}
				})
				t.Run("srcType=string&xType=[]byte", func(t *testing.T) {
					if hex.CanEncodeTo(tc.srcStr, x) {
						t.Error("got true; want false")
					}
				})
				t.Run("srcType=string&xType=string", func(t *testing.T) {
					if hex.CanEncodeTo(tc.srcStr, xStr) {
						t.Error("got true; want false")
					}
				})
			},
		)
		skipAll = false
	}
	if skipAll {
		t.Errorf("No test about numeric character xor %#x as dst!",
			hex.LetterCaseDiff)
	}
}

func TestCanEncodeToPrefix(t *testing.T) {
	const MaxI int = 7
	for _, srcCase := range testEncodeCases {
		if srcCase.upper { // only use the lower cases to avoid redundant sources
			continue
		}
		prefixSet := make(map[string]struct{}, len(testEncodeCases)*(MaxI+1))
		for i := 0; i <= MaxI; i++ {
			for _, prefixCase := range testEncodeCases {
				prefix, prefixBytes := getPrefixAndPrefixBytes(t, i, prefixCase)
				if _, ok := prefixSet[prefix]; ok {
					continue
				}
				prefixSet[prefix] = struct{}{}
				// srcCase.dstStr is in lowercase as the uppercase is skipped.
				want := strings.HasPrefix(
					srcCase.dstStr, strings.ToLower(prefix))
				t.Run(
					fmt.Sprintf(
						"src=%s&prefix=%s&upper=%t",
						srcCase.srcName,
						stringName(prefix),
						prefixCase.upper,
					),
					func(t *testing.T) {
						t.Run(
							"srcType=[]byte&prefixType=[]byte",
							func(t *testing.T) {
								got := hex.CanEncodeToPrefix(
									srcCase.srcBytes, prefixBytes)
								if got != want {
									t.Errorf("got %t; want %t", got, want)
								}
							},
						)
						t.Run(
							"srcType=[]byte&prefixType=string",
							func(t *testing.T) {
								got := hex.CanEncodeToPrefix(
									srcCase.srcBytes, prefix)
								if got != want {
									t.Errorf("got %t; want %t", got, want)
								}
							},
						)
						t.Run(
							"srcType=string&prefixType=[]byte",
							func(t *testing.T) {
								got := hex.CanEncodeToPrefix(
									srcCase.srcStr, prefixBytes)
								if got != want {
									t.Errorf("got %t; want %t", got, want)
								}
							},
						)
						t.Run(
							"srcType=string&prefixType=string",
							func(t *testing.T) {
								got := hex.CanEncodeToPrefix(
									srcCase.srcStr, prefix)
								if got != want {
									t.Errorf("got %t; want %t", got, want)
								}
							},
						)
					},
				)
			}
		}
	}
}

// getPrefixAndPrefixBytes returns the prefix and prefix bytes
// for TestCanEncodeToPrefix according to the specified i and test encode case.
//
// It uses t.Fatalf to stop the test if i is out of range [0, 7].
func getPrefixAndPrefixBytes(t *testing.T, i int, prefixCase *testEncodeCase) (
	prefix string, prefixBytes []byte) {
	switch i {
	case 0:
		if len(prefixCase.dstStr) > 1 {
			prefix = prefixCase.dstStr[:1]
			prefixBytes = prefixCase.dstBytes[:1]
		}
	case 1:
		if len(prefixCase.dstStr) > 2 {
			prefix = prefixCase.dstStr[:2]
			prefixBytes = prefixCase.dstBytes[:2]
		}
	case 2:
		if end := len(prefixCase.dstStr)/2 - 1; end > 0 {
			prefix = prefixCase.dstStr[:end]
			prefixBytes = prefixCase.dstBytes[:end]
		}
	case 3:
		prefix = prefixCase.dstStr[:len(prefixCase.dstStr)/2]
		prefixBytes = prefixCase.dstBytes[:len(prefixCase.dstBytes)/2]
	case 4:
		if end := len(prefixCase.dstStr) - 1; end > 0 {
			prefix = prefixCase.dstStr[:end]
			prefixBytes = prefixCase.dstBytes[:end]
		}
	case 5:
		prefix = prefixCase.dstStr
		prefixBytes = prefixCase.dstBytes
	case 6:
		prefix = prefixCase.dstStr + "0"
		prefixBytes = []byte(prefix)
	case 7:
		prefix = prefixCase.dstStr + "00"
		prefixBytes = []byte(prefix)
	default:
		t.Fatalf("i (%d) is out of range [0, 7]", i)
	}
	return
}

func TestCanEncodeToBytesStringFunctions(t *testing.T) {
	src, dst, _, sameLen :=
		makeCanEncodeToPrefixBytesStringFunctionsTestData(t, -1)
	if t.Failed() {
		return
	}

	fns := []struct {
		name string
		f    func(src []byte, x string) bool
	}{
		{"MyFunc", hex.CanEncodeTo[[]byte, string]},
		{"Another1", canEncodeToBytesString1},
		{"Another2", canEncodeToBytesString2},
	}

	testCases := []struct {
		name string
		x    string
		want bool
	}{
		{"Match", dst, true},
		{"FailSameLen", sameLen, false},
		{"FailDiffLen", dst[:len(dst)/2], false},
	}

	for _, fn := range fns {
		t.Run(fn.name, func(t *testing.T) {
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					got := fn.f(src, tc.x)
					if got != tc.want {
						t.Errorf("got %t; want %t", got, tc.want)
					}
				})
			}
		})
	}
}

func BenchmarkCanEncodeToBytesStringFunctions(b *testing.B) {
	src, dst, _, sameLen :=
		makeCanEncodeToPrefixBytesStringFunctionsTestData(b, -1)
	if b.Failed() {
		return
	}

	dataList := []struct {
		name, x string
	}{
		{"Match", dst},
		{"FailSameLen", sameLen},
		{"FailDiffLen", dst[:len(dst)/2]},
	}

	fns := []struct {
		name string
		f    func(src []byte, x string) bool
	}{
		{"MyFunc", hex.CanEncodeTo[[]byte, string]},
		{"Another1", canEncodeToBytesString1},
		{"Another2", canEncodeToBytesString2},
	}

	for _, data := range dataList {
		b.Run(data.name, func(b *testing.B) {
			for _, fn := range fns {
				b.Run(fn.name, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						fn.f(src, data.x)
					}
				})
			}
		})
	}
}

func TestCanEncodeToPrefixBytesStringFunctions(t *testing.T) {
	src, dst, prefixEven, sameLenEven :=
		makeCanEncodeToPrefixBytesStringFunctionsTestData(t, 12)
	if t.Failed() {
		return
	}
	prefixOdd := prefixEven[:len(prefixEven)-1]
	sameLenOdd := sameLenEven[:len(sameLenEven)-1]

	fns := []struct {
		name string
		f    func(src []byte, prefix string) bool
	}{
		{"MyFunc", hex.CanEncodeToPrefix[[]byte, string]},
		{"Another", canEncodeToPrefixBytesString},
	}

	testCases := []struct {
		name   string
		prefix string
		want   bool
	}{
		{"MatchEven", prefixEven, true},
		{"MatchOdd", prefixOdd, true},
		{"FailEven", sameLenEven, false},
		{"FailOdd", sameLenOdd, false},
		{"FailTooLongEven", dst + "00", false},
		{"FailTooLongOdd", dst + "0", false},
	}

	for _, fn := range fns {
		t.Run(fn.name, func(t *testing.T) {
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					got := fn.f(src, tc.prefix)
					if got != tc.want {
						t.Errorf("got %t; want %t", got, tc.want)
					}
				})
			}
		})
	}
}

func BenchmarkCanEncodeToPrefixBytesStringFunctions(b *testing.B) {
	src, dst, prefixEven, sameLenEven :=
		makeCanEncodeToPrefixBytesStringFunctionsTestData(b, 12)
	if b.Failed() {
		return
	}
	prefixOdd := prefixEven[:len(prefixEven)-1]
	sameLenOdd := sameLenEven[:len(sameLenEven)-1]

	dataList := []struct {
		name, prefix string
	}{
		{"MatchEven", prefixEven},
		{"MatchOdd", prefixOdd},
		{"FailEven", sameLenEven},
		{"FailOdd", sameLenOdd},
		{"FailTooLongEven", dst + "00"},
		{"FailTooLongOdd", dst + "0"},
	}

	fns := []struct {
		name string
		f    func(src []byte, prefix string) bool
	}{
		{"MyFunc", hex.CanEncodeToPrefix[[]byte, string]},
		{"Another", canEncodeToPrefixBytesString},
	}

	for _, data := range dataList {
		b.Run(data.name, func(b *testing.B) {
			for _, fn := range fns {
				b.Run(fn.name, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						fn.f(src, data.prefix)
					}
				})
			}
		})
	}
}

// canEncodeToBytesString1 is another implementation of
// function CanEncodeTo[[]byte, string],
// based on EncodeToString and strings.ToLower.
func canEncodeToBytesString1(src []byte, x string) bool {
	return hex.EncodeToString(src, false) == strings.ToLower(x)
}

// canEncodeToBytesString2 is another implementation of
// function CanEncodeTo[[]byte, string],
// based on standard library function hex.EncodeToString and strings.EqualFold.
func canEncodeToBytesString2(src []byte, x string) bool {
	return strings.EqualFold(stdhex.EncodeToString(src), x)
}

// canEncodeToPrefixBytesString is another implementation of
// function CanEncodeToPrefix[[]byte, string],
// based on standard library function hex.EncodeToString,
// strings.HasPrefix, and strings.ToLower.
func canEncodeToPrefixBytesString(src []byte, prefix string) bool {
	return strings.HasPrefix(
		stdhex.EncodeToString(src), strings.ToLower(prefix))
}

// makeCanEncodeToPrefixBytesStringFunctionsTestData generates data for testing
// CanEncodeTo[[]byte, string] and CanEncodeToPrefix[[]byte, string],
// and their alternative implementations.
//
// If prefixLen is not positive, prefix is the same as dst.
// If prefixLen is greater than len(dst),
// makeCanEncodeToPrefixBytesStringFunctionsTestData logs an error and
// returns zero-value src, dst, prefix, and sameLen.
func makeCanEncodeToPrefixBytesStringFunctionsTestData(
	tb testing.TB,
	prefixLen int,
) (
	src []byte, dst, prefix, sameLen string) {
	dst = testEncodeLongSrcCases[0].dstStr
	if prefixLen <= 0 {
		prefixLen = len(dst)
	} else if prefixLen > len(dst) {
		tb.Errorf(
			"prefixLen (%d) is out of range (should be up to %d)",
			prefixLen,
			len(dst),
		)
		dst = ""
		return
	}
	prefix = dst[:prefixLen]

	var b strings.Builder
	b.Grow(prefixLen)
	half := prefixLen / 2
	if half > 0 {
		b.WriteString(prefix[:half])
	}
	if prefix[half] != '0' {
		b.WriteByte('0')
	} else {
		b.WriteByte('1')
	}
	if half+1 < prefixLen {
		b.WriteString(prefix[half+1:])
	}
	sameLen = b.String()

	src = testEncodeLongSrcCases[0].srcBytes
	return
}
