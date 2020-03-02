// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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

var testEncodeCases []struct {
	src   string
	upper bool
	dst   string
}

func init() {
	testEncodeCases = []struct {
		src   string
		upper bool
		dst   string
	}{
		{"", false, ""},
		{"", true, ""},
		{"Hello world! 你好，世界！", false, ""},
		{"Hello world! 你好，世界！", true, ""},
	}
	// Generate dst in cases
	for i, c := range testEncodeCases {
		s := hex.EncodeToString([]byte(c.src))
		if c.upper {
			s = strings.ToUpper(s)
		}
		testEncodeCases[i].dst = s // don't use c.dst, because c is a copy of cases[i]
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

func TestEncode(t *testing.T) {
	dst := make([]byte, 1024)
	for _, c := range testEncodeCases {
		n := Encode(dst, []byte(c.src), c.upper)
		if n != len(c.dst) || string(dst[:n]) != c.dst {
			t.Errorf("dst: %q != %q, src: %q, upper: %t.", dst[:n], c.dst, c.src, c.upper)
		}
	}
}

func TestEncodedLen(t *testing.T) {
	dst := make([]byte, 1024)
	for _, c := range testEncodeCases {
		n := EncodedLen(len(c.src))
		if n2 := Encode(dst, []byte(c.src), c.upper); n != n2 {
			t.Errorf("EncodedLen: %d != %d, src: %q, upper: %t.", n, n2, c.src, c.upper)
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
	buf := make([]byte, 1024)
	w := bytes.NewBuffer(buf)
	upperEncoder := NewEncoder(w, true)
	lowerEncoder := NewEncoder(w, false)
	for _, c := range testEncodeCases {
		var encoder *Encoder
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
	buf := make([]byte, 1024)
	w := bytes.NewBuffer(buf)
	upperEncoder := NewEncoder(w, true)
	lowerEncoder := NewEncoder(w, false)
	for _, c := range testEncodeCases {
		var encoder *Encoder
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
				continue
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
	buf := make([]byte, 1024)
	w := bytes.NewBuffer(buf)
	upperEncoder := NewEncoder(w, true)
	lowerEncoder := NewEncoder(w, false)
	for _, c := range testEncodeCases {
		var encoder *Encoder
		if c.upper {
			encoder = upperEncoder
		} else {
			encoder = lowerEncoder
		}
		r := strings.NewReader(c.src)
		n, err := encoder.ReadFrom(r)
		if err != nil {
			t.Errorf("Error: %v, src: %q, upper: %t.", err, c.src, c.upper)
		}
		n = int64(EncodedLen(int(n)))
		if string(buf[:n]) != c.dst {
			t.Errorf("Output: %q != %q, src: %q, upper: %t.", buf[:n], c.dst, c.src, c.upper)
		}
		w.Reset()
	}
}
