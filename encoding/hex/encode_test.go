// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

type testEncodeCase struct {
	dst   string
	src   string
	upper bool
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
			s := hex.EncodeToString([]byte(src))
			if upper {
				s = strings.ToUpper(s)
			}
			testEncodeCases[i] = &testEncodeCase{
				dst:   s,
				src:   src,
				upper: upper,
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
	dst := make([]byte, hex.EncodedLen(len(srcs[len(srcs)-1])))
	stdDst := make([]byte, len(dst))
	for _, src := range srcs {
		n := Encode(dst, src, false)
		if n2 := hex.Encode(stdDst, src); n != n2 {
			t.Errorf(`Encode(dst, src, "", false): %d != hex.Encode(dst, src): %d, src: %q.`, n, n2, src)
		}
		if string(dst[:n]) != string(stdDst[:n]) {
			t.Errorf("dst: %q != stdDst: %q.", dst[:n], stdDst[:n])
		}
	}
}

func TestEncodedLen(t *testing.T) {
	for _, c := range testEncodeCases {
		if n := EncodedLen(len(c.src)); n != len(c.dst) {
			t.Errorf("EncodedLen: %d != %d, src: %q, upper: %t.", n, len(c.dst), c.src, c.upper)
		}
	}
}

func TestEncodedLen64(t *testing.T) {
	for _, c := range testEncodeCases {
		if n := EncodedLen64(int64(len(c.src))); n != int64(len(c.dst)) {
			t.Errorf("EncodedLen: %d != %d, src: %q, upper: %t.", n, len(c.dst), c.src, c.upper)
		}
	}
}

func TestEncode(t *testing.T) {
	dst := make([]byte, testEncodeCasesDstMaxLen+1024)
	for _, c := range testEncodeCases {
		n := Encode(dst, []byte(c.src), c.upper)
		if string(dst[:n]) != c.dst {
			t.Errorf("dst: %q != %q, src: %q, upper: %t.", dst[:n], c.dst, c.src, c.upper)
		}
	}
}

func TestEncodeToString(t *testing.T) {
	for _, c := range testEncodeCases {
		s := EncodeToString([]byte(c.src), c.upper)
		if s != c.dst {
			t.Errorf("EncodeToString: %q != %q, src: %q, upper: %t.", s, c.dst, c.src, c.upper)
		}
	}
}

func TestEncoder_Write(t *testing.T) {
	buf := make([]byte, testEncodeCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	upperEncoder := NewEncoder(w, true)
	lowerEncoder := NewEncoder(w, false)
	for _, c := range testEncodeCases {
		var encoder Encoder
		if c.upper {
			encoder = upperEncoder
		} else {
			encoder = lowerEncoder
		}
		n, err := encoder.Write([]byte(c.src))
		if err != nil {
			t.Errorf("Error: %v, src: %q, upper: %t.", err, c.src, c.upper)
		}
		n = EncodedLen(n)
		if string(buf[:n]) != c.dst {
			t.Errorf("Output: %q != %q, src: %q, upper: %t.", buf[:n], c.dst, c.src, c.upper)
		}
		w.Reset()
	}
}

func TestEncoder_WriteByte(t *testing.T) {
	buf := make([]byte, testEncodeCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	upperEncoder := NewEncoder(w, true)
	lowerEncoder := NewEncoder(w, false)
	for _, c := range testEncodeCases {
		var encoder Encoder
		if c.upper {
			encoder = upperEncoder
		} else {
			encoder = lowerEncoder
		}
		var n int
		for _, b := range []byte(c.src) {
			err := encoder.WriteByte(b)
			if err != nil {
				t.Errorf("Error: %v, src: %q, upper: %t.", err, c.src, c.upper)
				break
			}
			n++
		}
		n = EncodedLen(n)
		if string(buf[:n]) != c.dst {
			t.Errorf("Output: %q != %q, src: %q, upper: %t.", buf[:n], c.dst, c.src, c.upper)
		}
		w.Reset()
	}
}

func TestEncoder_ReadFrom(t *testing.T) {
	buf := make([]byte, testEncodeCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	upperEncoder := NewEncoder(w, true)
	lowerEncoder := NewEncoder(w, false)
	for _, c := range testEncodeCases {
		var encoder Encoder
		if c.upper {
			encoder = upperEncoder
		} else {
			encoder = lowerEncoder
		}
		n, err := encoder.ReadFrom(strings.NewReader(c.src))
		if err != nil {
			t.Errorf("Error: %v, src: %q, upper: %t.", err, c.src, c.upper)
		}
		n = EncodedLen64(n)
		if string(buf[:n]) != c.dst {
			t.Errorf("Output: %q != %q, src: %q, upper: %t.", buf[:n], c.dst, c.src, c.upper)
		}
		w.Reset()
	}
}
