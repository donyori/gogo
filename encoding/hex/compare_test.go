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
	"strconv"
	"strings"
	"testing"

	"github.com/donyori/gogo/encoding/hex"
)

// letterCaseDiff is copied from github.com/donyori/gogo/encoding/hex.
const letterCaseDiff byte = 'A' ^ 'a'

func TestCanEncode(t *testing.T) {
	for _, srcCase := range testEncodeCases {
		if srcCase.upper { // only use the lower cases
			continue
		}
		src := srcCase.src
		dstSet := make(map[string]bool, len(testEncodeCases))
		for _, dstCase := range testEncodeCases {
			dst := dstCase.dst
			if dstSet[dst] {
				continue
			}
			dstSet[dst] = true
			t.Run("src="+srcCase.srcName+"&x="+dstCase.dstName, func(t *testing.T) {
				if r := hex.CanEncode([]byte(src), []byte(dst)); r != (src == dstCase.src) {
					t.Errorf("got %t; want %t", r, src == dstCase.src)
				}
			})
		}
	}
	skipAll := true
	dstSet := make(map[string]bool, len(testEncodeCases))
	for _, tc := range testEncodeCases {
		dst := []byte(tc.dst)
		skip := true
		for i := range dst {
			if dst[i] <= '9' {
				dst[i] ^= letterCaseDiff
				skip = false
			}
		}
		if skip || dstSet[string(dst)] {
			continue
		}
		dstSet[string(dst)] = true
		var dstName string
		if len(dst) <= 80 {
			dstName = strconv.Quote(string(dst))
		} else {
			dstName = fmt.Sprintf("<long string %d>", len(dst))
		}
		t.Run(fmt.Sprintf("src=%s&dst=%s&numeric-xor-%#x", tc.srcName, dstName, letterCaseDiff), func(t *testing.T) {
			if hex.CanEncode([]byte(tc.src), dst) {
				t.Error("got true; want false")
			}
		})
		skipAll = false
	}
	if skipAll {
		t.Errorf("No test about numeric character xor %#x as dst!", letterCaseDiff)
	}
}

func TestCanEncodeToString(t *testing.T) {
	for _, srcCase := range testEncodeCases {
		if srcCase.upper { // only use the lower cases
			continue
		}
		src := srcCase.src
		dstSet := make(map[string]bool, len(testEncodeCases))
		for _, dstCase := range testEncodeCases {
			dst := dstCase.dst
			if dstSet[dst] {
				continue
			}
			dstSet[dst] = true
			t.Run("src="+srcCase.srcName+"&x="+dstCase.dstName, func(t *testing.T) {
				if r := hex.CanEncodeToString([]byte(src), dst); r != (src == dstCase.src) {
					t.Errorf("got %t; want %t", r, src == dstCase.src)
				}
			})
		}
	}
	skipAll := true
	dstSet := make(map[string]bool, len(testEncodeCases))
	for _, tc := range testEncodeCases {
		dst := []byte(tc.dst)
		skip := true
		for i := range dst {
			if dst[i] <= '9' {
				dst[i] ^= letterCaseDiff
				skip = false
			}
		}
		if skip || dstSet[string(dst)] {
			continue
		}
		dstSet[string(dst)] = true
		var dstName string
		if len(dst) <= 80 {
			dstName = strconv.Quote(string(dst))
		} else {
			dstName = fmt.Sprintf("<long string %d>", len(dst))
		}
		t.Run(fmt.Sprintf("src=%s&dst=%s&numeric-xor-%#x", tc.srcName, dstName, letterCaseDiff), func(t *testing.T) {
			if hex.CanEncodeToString([]byte(tc.src), string(dst)) {
				t.Error("got true; want false")
			}
		})
		skipAll = false
	}
	if skipAll {
		t.Errorf("No test about numeric character xor %#x as dst!", letterCaseDiff)
	}
}

func BenchmarkCanEncodeToString(b *testing.B) {
	fns := []struct {
		name string
		fn   func(src []byte, x string) bool
	}{
		{"MyFunc", hex.CanEncodeToString},
		{"Another1", testCanEncodeToString1},
		{"Another2", testCanEncodeToString2},
	}
	src := make([]byte, 9999)
	for i := range src {
		src[i] = byte(i % (1 << 8))
	}
	dst := stdhex.EncodeToString(src)
	data := []struct {
		name string
		x    string
		r    bool
	}{
		{"Success", dst, true},
		{"FailSameLen", strings.Replace(dst, "a", "B", 4), false},
		{"FailDiffLen", dst[:len(dst)/2], false},
	}
	benchmarks := make([]struct {
		name string
		fn   func(src []byte, x string) bool
		x    string
		r    bool
	}, len(fns)*len(data))
	var idx int
	for i := range data {
		for k := range fns {
			benchmarks[idx].name = fns[k].name + "_" + data[i].name
			benchmarks[idx].fn = fns[k].fn
			benchmarks[idx].x = data[i].x
			benchmarks[idx].r = data[i].r
			idx++
		}
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if r := bm.fn(src, bm.x); r != bm.r {
					b.Errorf("got %t; want %t", r, bm.r)
				}
			}
		})
	}
}

// testCanEncodeToString1 is another implementation of
// function CanEncodeToString, based on EncodeToString and strings.ToLower.
func testCanEncodeToString1(src []byte, x string) bool {
	return hex.EncodeToString(src, false) == strings.ToLower(x)
}

// testCanEncodeToString2 is another implementation of
// function CanEncodeToString,
// based on standard library function hex.EncodeToString and strings.EqualFold.
func testCanEncodeToString2(src []byte, x string) bool {
	return strings.EqualFold(stdhex.EncodeToString(src), x)
}
