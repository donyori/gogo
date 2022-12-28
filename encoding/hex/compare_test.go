// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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
		xSet := make(map[string]bool, len(testEncodeCases))
		for _, xCase := range testEncodeCases {
			if xSet[xCase.dstStr] {
				continue
			}
			xSet[xCase.dstStr] = true
			want := srcCase.srcStr == xCase.srcStr
			t.Run(
				fmt.Sprintf("src=%s&x=%s&upper=%t", srcCase.srcName, xCase.dstName, xCase.upper),
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

	// Test about hex.LetterCaseDiff.
	skipAll := true
	xSet := make(map[string]bool, len(testEncodeCases))
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
		if skip || xSet[xStr] {
			continue
		}
		xSet[xStr] = true
		t.Run(
			fmt.Sprintf("src=%s&x=%s&upper=%t&numeric-xor-%#x", tc.srcName, stringName(xStr), tc.upper, hex.LetterCaseDiff),
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
		t.Errorf("No test about numeric character xor %#x as dst!", hex.LetterCaseDiff)
	}
}

func TestCanEncodeToPrefix(t *testing.T) {
	const maxI int = 8
	for _, srcCase := range testEncodeCases {
		if srcCase.upper { // only use the lower cases to avoid redundant sources
			continue
		}
		prefixSet := make(map[string]bool, len(testEncodeCases)*maxI)
		for i := 0; i < maxI; i++ {
			for _, prefixCase := range testEncodeCases {
				var prefix string
				var prefixBytes []byte
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
					// This should never happen, but will act as a safeguard for later,
					// as a default value doesn't make sense here.
					t.Fatalf("i (%d) is out of range", i)
				}
				if prefixSet[prefix] {
					continue
				}
				prefixSet[prefix] = true
				want := strings.HasPrefix(srcCase.dstStr, strings.ToLower(prefix))
				t.Run(
					fmt.Sprintf("src=%s&prefix=%s&upper=%t", srcCase.srcName, stringName(prefix), prefixCase.upper),
					func(t *testing.T) {
						t.Run("srcType=[]byte&prefixType=[]byte", func(t *testing.T) {
							got := hex.CanEncodeToPrefix(srcCase.srcBytes, prefixBytes)
							if got != want {
								t.Errorf("got %t; want %t", got, want)
							}
						})
						t.Run("srcType=[]byte&prefixType=string", func(t *testing.T) {
							got := hex.CanEncodeToPrefix(srcCase.srcBytes, prefix)
							if got != want {
								t.Errorf("got %t; want %t", got, want)
							}
						})
						t.Run("srcType=string&prefixType=[]byte", func(t *testing.T) {
							got := hex.CanEncodeToPrefix(srcCase.srcStr, prefixBytes)
							if got != want {
								t.Errorf("got %t; want %t", got, want)
							}
						})
						t.Run("srcType=string&prefixType=string", func(t *testing.T) {
							got := hex.CanEncodeToPrefix(srcCase.srcStr, prefix)
							if got != want {
								t.Errorf("got %t; want %t", got, want)
							}
						})
					},
				)
			}
		}
	}
}

func BenchmarkCanEncodeTo(b *testing.B) {
	src := testEncodeLongSrcCases[0].srcBytes
	dst := testEncodeLongSrcCases[0].dstStr
	var sameLen string
	if strings.ContainsRune(dst, '0') {
		sameLen = strings.Replace(dst, "0", "1", 2)
	} else {
		for r := '1'; r <= '9'; r++ {
			if strings.ContainsRune(dst, r) {
				sameLen = strings.Replace(dst, string(r), "0", 2)
				break
			}
		}
		if sameLen == "" {
			for r := 'a'; r <= 'f'; r++ {
				if strings.ContainsRune(dst, r) {
					sameLen = strings.Replace(dst, string(r), "0", 2)
					break
				}
			}
		}
	}
	if sameLen == "" {
		b.Fatal("cannot build sameLen")
	}
	data := []struct {
		name string
		x    string
		want bool
	}{
		{"Success", dst, true},
		{"FailSameLen", sameLen, false},
		{"FailDiffLen", dst[:len(dst)/2], false},
	}
	fns := []struct {
		name string
		fn   func(src []byte, x string) bool
	}{
		{"MyFunc", hex.CanEncodeTo[[]byte, string]},
		{"Another1", canEncodeToBytesString1},
		{"Another2", canEncodeToBytesString2},
	}
	benchmarks := make([]struct {
		name string
		fn   func(src []byte, x string) bool
		x    string
		want bool
	}, len(fns)*len(data))
	var idx int
	for i := range data {
		for k := range fns {
			benchmarks[idx].name = fns[k].name + "_" + data[i].name
			benchmarks[idx].fn = fns[k].fn
			benchmarks[idx].x = data[i].x
			benchmarks[idx].want = data[i].want
			idx++
		}
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if got := bm.fn(src, bm.x); got != bm.want {
					b.Errorf("got %t; want %t", got, bm.want)
				}
			}
		})
	}
}

func BenchmarkCanEncodeToPrefix(b *testing.B) {
	src := testEncodeLongSrcCases[0].srcBytes
	prefix := testEncodeLongSrcCases[0].dstStr[:12]
	var sameLen string
	if strings.ContainsRune(prefix, '0') {
		sameLen = strings.Replace(prefix, "0", "1", 2)
	} else {
		for r := '1'; r <= '9'; r++ {
			if strings.ContainsRune(prefix, r) {
				sameLen = strings.Replace(prefix, string(r), "0", 2)
				break
			}
		}
		if sameLen == "" {
			for r := 'a'; r <= 'f'; r++ {
				if strings.ContainsRune(prefix, r) {
					sameLen = strings.Replace(prefix, string(r), "0", 2)
					break
				}
			}
		}
	}
	if sameLen == "" {
		b.Fatal("cannot build sameLen")
	}
	data := []struct {
		name   string
		prefix string
		want   bool
	}{
		{"Success", prefix, true},
		{"Fail", sameLen, false},
		{"FailTooLong", testEncodeLongSrcCases[0].dstStr + "00", false},
	}
	fns := []struct {
		name string
		fn   func(src []byte, prefix string) bool
	}{
		{"MyFunc", hex.CanEncodeToPrefix[[]byte, string]},
		{"Another", canEncodeToPrefixBytesString},
	}
	benchmarks := make([]struct {
		name   string
		fn     func(src []byte, prefix string) bool
		prefix string
		want   bool
	}, len(fns)*len(data))
	var idx int
	for i := range data {
		for k := range fns {
			benchmarks[idx].name = fns[k].name + "_" + data[i].name
			benchmarks[idx].fn = fns[k].fn
			benchmarks[idx].prefix = data[i].prefix
			benchmarks[idx].want = data[i].want
			idx++
		}
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if got := bm.fn(src, bm.prefix); got != bm.want {
					b.Errorf("got %t; want %t", got, bm.want)
				}
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
	return strings.HasPrefix(stdhex.EncodeToString(src), strings.ToLower(prefix))
}
