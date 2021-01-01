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

package helper

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/donyori/gogo/encoding/hex"
	"github.com/donyori/gogo/errors"
)

// PrefixBytesNo creates a function that generates a prefix that representing
// the offset of the beginning of current line, in bytes.
//
// If cfg != nil, cfg.BlockLen > 0, and cfg.BlocksPerLine > 0,
// it returns a prefix function used in DumpConfig that generates a prefix
// representing the byte offset of the beginning of current line,
// in hexadecimal representation, at least "digits" digits
// (padding with 0 if not enough).
// If digits is non-positive, it will use 8 instead.
// upper indicates to use uppercase in the hexadecimal representation.
// initCount specifies the byte offset of the first line.
//
// Otherwise (cfg == nil, cfg.BlockLen <= 0, or cfg.BlocksPerLine <= 0),
// it returns nil.
func PrefixBytesNo(cfg *hex.DumpConfig, upper bool, digits int, initCount int64) func() []byte {
	if cfg == nil || cfg.BlockLen <= 0 || cfg.BlocksPerLine <= 0 {
		return nil
	}
	if digits <= 0 {
		digits = 8
	}
	x := "x"
	if upper {
		x = "X"
	}
	layout := fmt.Sprintf("%%0%d%s: ", digits, x)
	count := initCount
	length := int64(cfg.BlockLen * cfg.BlocksPerLine)
	buf := bytes.NewBuffer(make([]byte, 0, digits+2))
	return func() []byte {
		buf.Reset()
		_, err := fmt.Fprintf(buf, layout, count)
		if err != nil {
			panic(errors.AutoWrap(err))
		}
		count += length
		return buf.Bytes()
	}
}

// SuffixQuotedPretty creates a function that generates a suffix consisting of
// " | " and the argument line represented in double-quoted Go string literal.
//
// If cfg != nil, cfg.BlockLen > 0, and cfg.BlocksPerLine > 0,
// it returns a suffix function used in DumpConfig that generates a suffix
// consisting of " | " and the argument line represented in double-quoted
// Go string literal.
// If the length of line is less than cfg.BlockLen * cfg.BlocksPerLine,
// it pads spaces before " | ".
//
// Otherwise (cfg == nil, cfg.BlockLen <= 0, or cfg.BlocksPerLine <= 0),
// it returns nil.
func SuffixQuotedPretty(cfg *hex.DumpConfig) func(line []byte) []byte {
	if cfg == nil || cfg.BlockLen <= 0 || cfg.BlocksPerLine <= 0 {
		return nil
	}
	cfgCopy := new(hex.DumpConfig)
	*cfgCopy = *cfg
	length := cfg.BlockLen * cfg.BlocksPerLine
	buf := bytes.NewBuffer(make([]byte, 0, hex.FormattedLen(length, &cfg.FormatConfig)+length+10))
	return func(line []byte) []byte {
		buf.Reset()
		if len(line) < length {
			n := hex.FormattedLen(length, &cfgCopy.FormatConfig) -
				hex.FormattedLen(len(line), &cfgCopy.FormatConfig)
			for i := 0; i < n; i++ {
				buf.WriteRune(' ')
			}
		}
		buf.WriteString(" | ")
		buf.WriteString(strconv.Quote(string(line)))
		return buf.Bytes()
	}
}

// ExampleDumpConfig returns an example hex.DumpConfig,
// with a prefix function generated by PrefixBytesNo and
// a suffix function generated by SuffixQuotedPretty.
//
// upper indicates to use uppercase in the hexadecimal representation.
// bytesPerLine is the number of bytes to show in one line.
// If bytesPerLine is non-positive, it will use 16 instead.
func ExampleDumpConfig(upper bool, bytesPerLine int) *hex.DumpConfig {
	if bytesPerLine <= 0 {
		bytesPerLine = 16
	}
	cfg := &hex.DumpConfig{
		FormatConfig: hex.FormatConfig{
			Upper:    upper,
			Sep:      " ",
			BlockLen: 1,
		},
		LineSep:       "\n",
		BlocksPerLine: bytesPerLine,
	}
	if bytesPerLine%2 == 0 {
		cfg.BlockLen = 2
		cfg.BlocksPerLine /= 2
	}
	cfg.PrefixFn = PrefixBytesNo(cfg, upper, 0, 0)
	cfg.SuffixFn = SuffixQuotedPretty(cfg)
	return cfg
}
