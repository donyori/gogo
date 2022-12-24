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
	"bytes"
	stdhex "encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/donyori/gogo/encoding/hex"
)

type testEncodeCase struct {
	srcName string
	dstName string
	dst     string
	src     string
	upper   bool
}

var testEncodeCases []*testEncodeCase
var testEncodeCasesDstMaxLen int

func init() {
	srcs := []string{"", "Hello world! 你好，世界！", ""}
	var longStringBuilder strings.Builder
	longStringBuilder.Grow(16384 + len(srcs[1]))
	for longStringBuilder.Len() < 16384 {
		longStringBuilder.WriteString(srcs[1])
	}
	srcs[2] = longStringBuilder.String()
	uppers := []bool{false, true}
	testEncodeCases = make([]*testEncodeCase, len(srcs)*len(uppers))
	var i int
	for _, src := range srcs {
		for _, upper := range uppers {
			s := stdhex.EncodeToString([]byte(src))
			if upper {
				s = strings.ToUpper(s)
			}
			var srcName string
			if len(src) <= 80 {
				srcName = strconv.Quote(src)
			} else {
				srcName = fmt.Sprintf("<long string %d>", len(src))
			}
			var dstName string
			if len(s) <= 80 {
				dstName = strconv.Quote(s)
			} else {
				dstName = fmt.Sprintf("<long string %d>", len(s))
			}
			testEncodeCases[i] = &testEncodeCase{
				srcName: srcName,
				dstName: dstName,
				dst:     s,
				src:     src,
				upper:   upper,
			}
			i++
			if testEncodeCasesDstMaxLen < len(s) {
				testEncodeCasesDstMaxLen = len(s)
			}
		}
	}
}

func TestEncode_CompareWithOfficial(t *testing.T) {
	srcs := [][]byte{nil, {}, []byte("Hello world! 你好，世界！")}
	dst := make([]byte, stdhex.EncodedLen(len(srcs[len(srcs)-1])))
	stdDst := make([]byte, len(dst))
	for _, src := range srcs {
		var srcName string
		if src == nil {
			srcName = "<nil>"
		} else {
			srcName = strconv.Quote(string(src))
		}
		t.Run("src="+srcName, func(t *testing.T) {
			n := hex.Encode(dst, src, false)
			if n2 := stdhex.Encode(stdDst, src); n != n2 {
				t.Fatalf("got n %d; want %d", n, n2)
			}
			if !bytes.Equal(dst[:n], stdDst[:n]) {
				t.Errorf("got %q; want %q", dst[:n], stdDst[:n])
			}
		})
	}
}

func TestEncodedLen(t *testing.T) {
	for _, tc := range testEncodeCases {
		if tc.upper { // only use the lower cases
			continue
		}
		t.Run("src="+tc.srcName, func(t *testing.T) {
			if n := hex.EncodedLen(len(tc.src)); n != len(tc.dst) {
				t.Errorf("got %d; want %d", n, len(tc.dst))
			}
		})
	}
}

func TestEncodedLen64(t *testing.T) {
	for _, tc := range testEncodeCases {
		if tc.upper { // only use the lower cases
			continue
		}
		t.Run("src="+tc.srcName, func(t *testing.T) {
			if n := hex.EncodedLen64(int64(len(tc.src))); n != int64(len(tc.dst)) {
				t.Errorf("got %d; want %d", n, len(tc.dst))
			}
		})
	}
}

func TestEncode(t *testing.T) {
	dst := make([]byte, testEncodeCasesDstMaxLen+1024)
	for _, tc := range testEncodeCases {
		t.Run(fmt.Sprintf("src=%s&upper=%t", tc.srcName, tc.upper), func(t *testing.T) {
			n := hex.Encode(dst, []byte(tc.src), tc.upper)
			if string(dst[:n]) != tc.dst {
				t.Errorf("got %q; want %q", dst[:n], tc.dst)
			}
		})
	}
}

func TestEncodeToString(t *testing.T) {
	for _, tc := range testEncodeCases {
		t.Run(fmt.Sprintf("src=%s&upper=%t", tc.srcName, tc.upper), func(t *testing.T) {
			s := hex.EncodeToString([]byte(tc.src), tc.upper)
			if s != tc.dst {
				t.Errorf("got %q; want %q", s, tc.dst)
			}
		})
	}
}

func TestEncoder_Write(t *testing.T) {
	buf := make([]byte, testEncodeCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	upperEncoder := hex.NewEncoder(w, true)
	lowerEncoder := hex.NewEncoder(w, false)
	for _, tc := range testEncodeCases {
		t.Run(fmt.Sprintf("src=%s&upper=%t", tc.srcName, tc.upper), func(t *testing.T) {
			w.Reset()
			var encoder hex.Encoder
			if tc.upper {
				encoder = upperEncoder
			} else {
				encoder = lowerEncoder
			}
			n, err := encoder.Write([]byte(tc.src))
			if err != nil {
				t.Fatal(err)
			}
			n = hex.EncodedLen(n)
			if string(buf[:n]) != tc.dst {
				t.Errorf("got %q; want %q", buf[:n], tc.dst)
			}
		})
	}
}

func TestEncoder_WriteByte(t *testing.T) {
	buf := make([]byte, testEncodeCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	upperEncoder := hex.NewEncoder(w, true)
	lowerEncoder := hex.NewEncoder(w, false)
	for _, tc := range testEncodeCases {
		t.Run(fmt.Sprintf("src=%s&upper=%t", tc.srcName, tc.upper), func(t *testing.T) {
			w.Reset()
			var encoder hex.Encoder
			if tc.upper {
				encoder = upperEncoder
			} else {
				encoder = lowerEncoder
			}
			var n int
			for _, b := range []byte(tc.src) {
				err := encoder.WriteByte(b)
				if err != nil {
					t.Fatalf("WriteByte(%q) - %v", b, err)
				}
				n++
			}
			n = hex.EncodedLen(n)
			if string(buf[:n]) != tc.dst {
				t.Errorf("got %q; want %q", buf[:n], tc.dst)
			}
		})
	}
}

func TestEncoder_ReadFrom(t *testing.T) {
	buf := make([]byte, testEncodeCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	upperEncoder := hex.NewEncoder(w, true)
	lowerEncoder := hex.NewEncoder(w, false)
	for _, tc := range testEncodeCases {
		t.Run(fmt.Sprintf("src=%s&upper=%t", tc.srcName, tc.upper), func(t *testing.T) {
			w.Reset()
			var encoder hex.Encoder
			if tc.upper {
				encoder = upperEncoder
			} else {
				encoder = lowerEncoder
			}
			n, err := encoder.ReadFrom(strings.NewReader(tc.src))
			if err != nil {
				t.Fatal(err)
			}
			n = hex.EncodedLen64(n)
			if string(buf[:n]) != tc.dst {
				t.Errorf("got %q; want %q", buf[:n], tc.dst)
			}
		})
	}
}
