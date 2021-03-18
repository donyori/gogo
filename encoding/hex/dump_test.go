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
	stdhex "encoding/hex"
	"strings"
	"testing"
)

type testDumpCase struct {
	dst string
	src string
	cfg *DumpConfig
}

var testDumpCases []*testDumpCase
var testDumpCasesDstMaxLen int

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
	lineSeps := []string{"", "\n", "\n\n"}
	blocksPerLines := []int{-1, 0, 1, 2, 3, 5, 7, 11, 20}
	prefixFns := []func() []byte{
		nil,
		func() []byte {
			return []byte("Prefix")
		},
	}
	suffixFns := []func(line []byte) []byte{
		nil,
		func(line []byte) []byte {
			return line
		},
		func(line []byte) []byte {
			suffix := make([]byte, 4+len(line))
			suffix[0] = ' '
			suffix[1] = ' '
			suffix[2] = '|'
			suffix[3] = ' '
			copy(suffix[4:], line)
			return suffix
		},
	}
	testDumpCases = make([]*testDumpCase, len(srcs)*len(uppers)*len(seps)*
		len(blockLens)*len(lineSeps)*len(blocksPerLines)*len(prefixFns)*len(suffixFns))
	var builder strings.Builder
	var i int
	for _, src := range srcs {
		for _, upper := range uppers {
			for _, sep := range seps {
				for _, blockLen := range blockLens {
					for _, lineSep := range lineSeps {
						for _, blocksPerLine := range blocksPerLines {
							for _, prefixFn := range prefixFns {
								for _, suffixFn := range suffixFns {
									cfg := &DumpConfig{
										FormatConfig: FormatConfig{
											Upper:    upper,
											Sep:      sep,
											BlockLen: blockLen,
										},
										LineSep:       lineSep,
										BlocksPerLine: blocksPerLine,
										PrefixFn:      prefixFn,
										SuffixFn:      suffixFn,
									}
									s := stdhex.EncodeToString([]byte(src))
									if upper {
										s = strings.ToUpper(s)
									}
									blockSize := stdhex.EncodedLen(blockLen)
									if blockSize <= 0 {
										blockSize = len(s)
									}
									if !cfg.dumpCfgLineNotValid() && prefixFn != nil {
										builder.Write(prefixFn())
									}
									var line []byte
									p := s
									t := src
									var j int
									for len(p) > 0 {
										if !cfg.dumpCfgLineNotValid() && j > 0 && j%blocksPerLine == 0 && prefixFn != nil {
											builder.Write(prefixFn())
										}
										end := blockSize
										if len(p) < end {
											end = len(p)
										}
										builder.WriteString(p[:end])
										line = append(line, t[:DecodedLen(end)]...)
										p = p[end:]
										t = t[DecodedLen(end):]
										j++
										if !cfg.dumpCfgLineNotValid() && j > 0 && j%blocksPerLine == 0 {
											if suffixFn != nil {
												builder.Write(suffixFn(line))
											}
											line = line[:0]
											builder.WriteString(lineSep)
										} else if !cfg.formatCfgNotValid() && len(p) > 0 {
											builder.WriteString(sep)
										}
									}
									if !cfg.dumpCfgLineNotValid() && (j == 0 || j%blocksPerLine != 0) {
										if suffixFn != nil {
											builder.Write(suffixFn(line))
										}
										builder.WriteString(lineSep)
									}
									testDumpCases[i] = &testDumpCase{
										dst: builder.String(),
										src: src,
										cfg: cfg,
									}
									i++
									if testDumpCasesDstMaxLen < builder.Len() {
										testDumpCasesDstMaxLen = builder.Len()
									}
									builder.Reset()
								}
							}
						}
					}
				}
			}
		}
	}
}

func TestDumpToString(t *testing.T) {
	for _, c := range testDumpCases {
		dst := DumpToString([]byte(c.src), c.cfg)
		if dst != c.dst {
			t.Errorf("dst: %q != %q, src: %q, cfg: %+v.", dst, c.dst, c.src, c.cfg)
		}
	}
}

func TestDumpTo(t *testing.T) {
	buf := make([]byte, testDumpCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	for _, c := range testDumpCases {
		_, err := DumpTo(w, []byte(c.src), c.cfg)
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		var n int
		for buf[n] != 0 {
			n++
		}
		if string(buf[:n]) != c.dst {
			t.Errorf("Output: %q != %q, src: %q, cfg: %+v.", buf[:n], c.dst, c.src, c.cfg)
		}
		w.Reset()
		for i := range buf {
			if buf[i] == 0 {
				break
			}
			buf[i] = 0
		}
	}
}

func TestDumper_Write(t *testing.T) {
	buf := make([]byte, testDumpCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	for _, c := range testDumpCases {
		d := NewDumper(w, c.cfg)
		_, err := d.Write([]byte(c.src))
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		err = d.Close()
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		err = d.Close() // Close() again, to detect whether two Close() can make output wrong.
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		var n int
		for buf[n] != 0 {
			n++
		}
		if string(buf[:n]) != c.dst {
			t.Errorf("Output: %q != %q, src: %q, cfg: %+v.", buf[:n], c.dst, c.src, c.cfg)
		}
		w.Reset()
		for i := range buf {
			if buf[i] == 0 {
				break
			}
			buf[i] = 0
		}
	}
}

func TestDumper_WriteByte(t *testing.T) {
	buf := make([]byte, testDumpCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	for _, c := range testDumpCases {
		d := NewDumper(w, c.cfg)
		for _, b := range []byte(c.src) {
			err := d.WriteByte(b)
			if err != nil {
				t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
				break
			}
		}
		err := d.Close()
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		err = d.Close() // Close() again, to detect whether two Close() can make output wrong.
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		var n int
		for buf[n] != 0 {
			n++
		}
		if string(buf[:n]) != c.dst {
			t.Errorf("Output: %q != %q, src: %q, cfg: %+v.", buf[:n], c.dst, c.src, c.cfg)
		}
		w.Reset()
		for i := range buf {
			if buf[i] == 0 {
				break
			}
			buf[i] = 0
		}
	}
}

func TestDumper_ReadFrom(t *testing.T) {
	buf := make([]byte, testDumpCasesDstMaxLen+1024)
	w := bytes.NewBuffer(buf)
	for _, c := range testDumpCases {
		d := NewDumper(w, c.cfg)
		_, err := d.ReadFrom(strings.NewReader(c.src))
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		err = d.Close()
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		err = d.Close() // Close() again, to detect whether two Close() can make output wrong.
		if err != nil {
			t.Errorf("Error: %v, src: %q, cfg: %+v.", err, c.src, c.cfg)
		}
		var n int
		for buf[n] != 0 {
			n++
		}
		if string(buf[:n]) != c.dst {
			t.Errorf("Output: %q != %q, src: %q, cfg: %+v.", buf[:n], c.dst, c.src, c.cfg)
		}
		w.Reset()
		for i := range buf {
			if buf[i] == 0 {
				break
			}
			buf[i] = 0
		}
	}
}
