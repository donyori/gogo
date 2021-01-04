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
	"math"
	"strings"
)

// EncodeInt64ToString returns hexadecimal representation of integer x.
//
// upper indicates to use uppercase in hexadecimal representation.
// digits specifies the minimum length of the return string
// (excluding the negative sign "-").
// It pads with leading zeros after the sign if the length is not enough.
// If digits is non-positive, no padding will be applied.
//
// The return string is just like the return value of the following function:
//  func foo(x int64, upper bool, digits int) string {
//  	layout := "%"
//  	if digits > 0 {
//  		width := digits
//  		if x < 0 {
//  			width++ // Increase one for the negative sign.
//  		}
//  		layout += "0" + strconv.Itoa(width)
//  	}
//  	if upper {
//  		layout += "X"
//  	} else {
//  		layout += "x"
//  	}
//  	return fmt.Sprintf(layout, x)
//  }
func EncodeInt64ToString(x int64, upper bool, digits int) string {
	// Special cases for 0 and 0x8000000000000000 (minimum value of int64):
	if x == 0 && digits <= 1 {
		return "0"
	}
	if x == math.MinInt64 {
		s := "-8000000000000000"
		if digits < int64BufferLen {
			return s
		}
		var b strings.Builder
		b.Grow(digits + 1)
		b.WriteByte('-')
		for i := 16; i < digits; i++ {
			b.WriteByte('0')
		}
		b.WriteString(s[1:])
		return b.String()
	}

	// Other cases:
	ht := lowercaseHexTable
	if upper {
		ht = uppercaseHexTable
	}
	isNeg := false
	if x < 0 {
		isNeg = true
		x = -x
	}
	bufp := int64BufferPool.Get().(*[]byte)
	defer int64BufferPool.Put(bufp)
	buf := *bufp
	idx := int64BufferLen
	for x != 0 {
		idx--
		buf[idx] = ht[x&0x0f]
		x >>= 4
	}
	if digits < int64BufferLen {
		s := int64BufferLen - digits
		for idx > s {
			idx--
			buf[idx] = '0'
		}
		if isNeg {
			idx--
			buf[idx] = '-'
		}
		return string(buf[idx:])
	} else {
		var b strings.Builder
		if isNeg {
			b.Grow(digits + 1)
			b.WriteByte('-')
		} else {
			b.Grow(digits)
		}
		for i := int64BufferLen - idx; i < digits; i++ {
			b.WriteByte('0')
		}
		b.Write(buf[idx:])
		return b.String()
	}
}
