// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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

package hex

import (
	stdhex "encoding/hex"
	"strings"
	"testing"
)

func TestCanEncode(t *testing.T) {
	for _, srcCase := range testEncodeCases {
		src := srcCase.src
		for _, dstCase := range testEncodeCases {
			dst := dstCase.dst
			if r := CanEncode([]byte(src), []byte(dst)); r != (src == dstCase.src) {
				t.Errorf("CanEncode: %t\n  src: %q\n  dst:         %q\n  encode(src): %q", r, src, dst, srcCase.dst)
			}
		}
	}
	skipAll := true
	for _, c := range testEncodeCases {
		dst := []byte(c.dst)
		skip := true
		for i := range dst {
			if dst[i] <= '9' {
				dst[i] ^= letterCaseDiff
				skip = false
			}
		}
		if skip {
			continue
		}
		if CanEncode([]byte(c.src), dst) {
			t.Errorf("CanEncode: true\n  src: %q\n  dst:         %q\n  encode(src): %q", c.src, dst, c.dst)
		}
		skipAll = false
	}
	if skipAll {
		t.Errorf("No test about numeric character xor %x as dst!", letterCaseDiff)
	}
}

func TestCanEncodeToString(t *testing.T) {
	for _, srcCase := range testEncodeCases {
		src := srcCase.src
		for _, dstCase := range testEncodeCases {
			dst := dstCase.dst
			if r := CanEncodeToString([]byte(src), dst); r != (src == dstCase.src) {
				t.Errorf("CanEncodeToString: %t\n  src: %q\n  dst:         %q\n  encode(src): %q", r, src, dst, srcCase.dst)
			}
		}
	}
	skipAll := true
	for _, c := range testEncodeCases {
		dst := []byte(c.dst)
		skip := true
		for i := range dst {
			if dst[i] <= '9' {
				dst[i] ^= letterCaseDiff
				skip = false
			}
		}
		if skip {
			continue
		}
		if CanEncodeToString([]byte(c.src), string(dst)) {
			t.Errorf("CanEncodeToString: true\n  src: %q\n  dst:         %q\n  encode(src): %q", c.src, dst, c.dst)
		}
		skipAll = false
	}
	if skipAll {
		t.Errorf("No test about numeric character xor %x as dst!", letterCaseDiff)
	}
}

func BenchmarkCanEncodeToString(b *testing.B) {
	fns := []struct {
		name string
		fn   func(src []byte, x string) bool
	}{
		{"MyFunc", CanEncodeToString},
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
	bms := make([]struct {
		name string
		fn   func(src []byte, x string) bool
		x    string
		r    bool
	}, len(fns)*len(data))
	var idx int
	for i := range data {
		for k := range fns {
			bms[idx].name = fns[k].name + "_" + data[i].name
			bms[idx].fn = fns[k].fn
			bms[idx].x = data[i].x
			bms[idx].r = data[i].r
			idx++
		}
	}

	for _, bm := range bms {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if r := bm.fn(src, bm.x); r != bm.r {
					b.Errorf("Case %q, got: %t\n  x:           %q\n  encode(src): %q", bm.name, r, bm.x, dst)
				}
			}
		})
	}
}

// testCanEncodeToString1 is another implementation of
// function CanEncodeToString, based on EncodeToString and strings.ToLower.
func testCanEncodeToString1(src []byte, x string) bool {
	return EncodeToString(src, false) == strings.ToLower(x)
}

// testCanEncodeToString2 is another implementation of
// function CanEncodeToString,
// based on standard library function hex.EncodeToString and strings.EqualFold.
func testCanEncodeToString2(src []byte, x string) bool {
	return strings.EqualFold(stdhex.EncodeToString(src), x)
}
