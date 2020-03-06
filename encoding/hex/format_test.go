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

type testFormatCase struct {
	dst string
	src string
	cfg *FormatConfig
}

var testFormatCases []*testFormatCase
var testFormatCasesDstMaxLen int

func init() {
	srcs := []string{"", "Hello world! 你好，世界！", ""}
	var longStringBuilder strings.Builder
	longStringBuilder.Grow(16384 + len(srcs[1]))
	for longStringBuilder.Len() < 16384 {
		longStringBuilder.WriteString(srcs[1])
	}
	srcs[2] = longStringBuilder.String()
	uppers := []bool{false, true}
	seps := []string{"", " ", "分隔符"}
	blockLens := []int{-1, 0, 1, 2, 3, 5, 7, 11, 2048}
	testFormatCases = make([]*testFormatCase, len(srcs)*len(uppers)*len(seps)*len(blockLens))
	var i int
	for _, src := range srcs {
		for _, upper := range uppers {
			for _, sep := range seps {
				for _, blockLen := range blockLens {
					s := hex.EncodeToString([]byte(src))
					if upper {
						s = strings.ToUpper(s)
					}
					capacity := 0
					if blockLen > 0 {
						capacity = (len(src) + blockLen - 1) / blockLen
					}
					blocks := make([]string, 0, capacity)
					blockSize := hex.EncodedLen(blockLen)
					if blockSize <= 0 {
						blockSize = len(s)
					}
					p := s
					for len(p) > 0 {
						end := blockSize
						if len(p) < end {
							end = len(p)
						}
						blocks = append(blocks, p[:end])
						p = p[end:]
					}
					dst := strings.Join(blocks, sep)
					testFormatCases[i] = &testFormatCase{
						dst: dst,
						src: src,
						cfg: &FormatConfig{
							Upper:    upper,
							Sep:      sep,
							BlockLen: blockLen,
						},
					}
					i++
					if testFormatCasesDstMaxLen < len(dst) {
						testFormatCasesDstMaxLen = len(dst)
					}
				}
			}
		}
	}
}

func TestFormattedLen(t *testing.T) {
	for _, c := range testFormatCases {
		if n := FormattedLen(len(c.src), c.cfg); n != len(c.dst) {
			t.Errorf("FormattedLen: %d != %d, src: %q, cfg: %+v.", n, len(c.dst), c.src, c.cfg)
		}
	}
}

func TestFormattedLen64(t *testing.T) {
	for _, c := range testFormatCases {
		if n := FormattedLen64(int64(len(c.src)), c.cfg); n != int64(len(c.dst)) {
			t.Errorf("FormattedLen: %d != %d, src: %q, cfg: %+v.", n, len(c.dst), c.src, c.cfg)
		}
	}
}

func TestFormat(t *testing.T) {
	dst := make([]byte, testFormatCasesDstMaxLen+1024)
	for _, c := range testFormatCases {
		n := Format(dst, []byte(c.src), c.cfg)
		if string(dst[:n]) != c.dst {
			t.Errorf("dst: %q != %q, src: %q, cfg: %+v.", dst[:n], c.dst, c.src, c.cfg)
		}
	}
}

func TestFormatToString(t *testing.T) {
	for _, c := range testFormatCases {
		s := FormatToString([]byte(c.src), c.cfg)
		if s != c.dst {
			t.Errorf("FormatToString: %q != %q, src: %q, cfg: %+v.", s, c.dst, c.src, c.cfg)
		}
	}
}

func TestFormatTo(t *testing.T) {
	buf := make([]byte, testFormatCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	for _, c := range testFormatCases {
		n, err := FormatTo(w, []byte(c.src), c.cfg)
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		n = FormattedLen(n, c.cfg)
		if string(buf[:n]) != c.dst {
			t.Errorf("Output: %q != %q, src: %q, cfg: %+v.", buf[:n], c.dst, c.src, c.cfg)
		}
		w.Reset()
	}
}

func TestFormatter_Write(t *testing.T) {
	buf := make([]byte, testFormatCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	for _, c := range testFormatCases {
		formatter := NewFormatter(w, c.cfg)
		n, err := formatter.Write([]byte(c.src))
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		err = formatter.Close()
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		err = formatter.Close() // Close() again, to detect whether two Close() can make output wrong.
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		n = FormattedLen(n, c.cfg)
		if string(buf[:n]) != c.dst {
			t.Errorf("Output: %q != %q, src: %q, cfg: %+v.", buf[:n], c.dst, c.src, c.cfg)
		}
		w.Reset()
	}
}

func TestFormatter_WriteByte(t *testing.T) {
	buf := make([]byte, testFormatCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	for _, c := range testFormatCases {
		formatter := NewFormatter(w, c.cfg)
		var n int
		for _, b := range []byte(c.src) {
			err := formatter.WriteByte(b)
			if err != nil {
				t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
				break
			}
			n++
		}
		err := formatter.Close()
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		err = formatter.Close() // Close() again, to detect whether two Close() can make output wrong.
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		n = FormattedLen(n, c.cfg)
		if string(buf[:n]) != c.dst {
			t.Errorf("Output: %q != %q, src: %q, cfg: %+v.", buf[:n], c.dst, c.src, c.cfg)
		}
		w.Reset()
	}
}

func TestFormatter_ReadFrom(t *testing.T) {
	buf := make([]byte, testFormatCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	for _, c := range testFormatCases {
		formatter := NewFormatter(w, c.cfg)
		n, err := formatter.ReadFrom(strings.NewReader(c.src))
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		err = formatter.Close()
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		err = formatter.Close() // Close() again, to detect whether two Close() can make output wrong.
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		n = FormattedLen64(n, c.cfg)
		if string(buf[:n]) != c.dst {
			t.Errorf("Output: %q != %q, src: %q, cfg: %+v.", buf[:n], c.dst, c.src, c.cfg)
		}
		w.Reset()
	}
}
