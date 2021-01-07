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
	"fmt"
	"math"
	"strings"

	"github.com/donyori/gogo/errors"
)

// minInt64Hex is the hexadecimal representation of math.MinInt64, as a string.
const minInt64Hex = "-8000000000000000"

// EncodeInt64 encodes x in hexadecimal representation into dst.
//
// upper indicates to use uppercase in hexadecimal representation.
// digits specifies the minimum length of the output content
// (excluding the negative sign "-").
// It pads with leading zeros after the sign (if any)
// if the length is not enough.
// If digits is non-positive, no padding will be applied.
//
// It returns the number of bytes written into dst.
//
// It panics if dst is too small to store the result.
// To ensure that dst has enough space to store the result,
// its length should be at least (digits + 1) and not less than 17.
func EncodeInt64(dst []byte, x int64, upper bool, digits int) int {
	dstTooSmallMsgFn := func(need int) string {
		return fmt.Sprintf("dst is too small, len(dst): %d, need: %d", len(dst), need)
	}

	// Special cases for 0 and -0x8000000000000000 (minimum value of int64):
	if x == 0 && digits <= 1 {
		if len(dst) < 1 {
			panic(errors.AutoMsg(dstTooSmallMsgFn(1)))
		}
		dst[0] = '0'
		return 1
	}
	if x == math.MinInt64 {
		if digits < int64BufferLen {
			if len(dst) < len(minInt64Hex) {
				panic(errors.AutoMsg(dstTooSmallMsgFn(len(minInt64Hex))))
			}
			return copy(dst, minInt64Hex)
		}
		if len(dst) < digits+1 {
			panic(errors.AutoMsg(dstTooSmallMsgFn(digits + 1)))
		}
		dst[0] = '-'
		idx := 1
		for i := 16; i < digits; i++ {
			dst[idx] = '0'
			idx++
		}
		return idx + copy(dst[idx:], minInt64Hex[1:])
	}

	// Other cases:
	bufp := int64BufferPool.Get().(*[]byte)
	defer int64BufferPool.Put(bufp)
	buf := *bufp
	bufIdx := encodeInt64(buf, x, upper, digits)
	if digits < int64BufferLen {
		if len(dst) < int64BufferLen-bufIdx {
			panic(errors.AutoMsg(dstTooSmallMsgFn(int64BufferLen - bufIdx)))
		}
		return copy(dst, buf[bufIdx:])
	}
	var idx int
	if x < 0 {
		if len(dst) < digits+1 {
			panic(errors.AutoMsg(dstTooSmallMsgFn(digits + 1)))
		}
		dst[idx] = '-'
		idx++
	} else if len(dst) < digits {
		panic(errors.AutoMsg(dstTooSmallMsgFn(digits)))
	}
	for i := int64BufferLen - bufIdx; i < digits; i++ {
		dst[idx] = '0'
		idx++
	}
	return idx + copy(dst[idx:], buf[bufIdx:])
}

// EncodeInt64ToString returns hexadecimal representation of integer x.
//
// upper indicates to use uppercase in hexadecimal representation.
// digits specifies the minimum length of the return string
// (excluding the negative sign "-").
// It pads with leading zeros after the sign (if any)
// if the length is not enough.
// If digits is non-positive, no padding will be applied.
func EncodeInt64ToString(x int64, upper bool, digits int) string {
	// Special cases for 0 and -0x8000000000000000 (minimum value of int64):
	if x == 0 && digits <= 1 {
		return "0"
	}
	if x == math.MinInt64 {
		if digits < int64BufferLen {
			return minInt64Hex
		}
		var b strings.Builder
		b.Grow(digits + 1)
		b.WriteByte('-')
		for i := 16; i < digits; i++ {
			b.WriteByte('0')
		}
		b.WriteString(minInt64Hex[1:])
		return b.String()
	}

	// Other cases:
	bufp := int64BufferPool.Get().(*[]byte)
	defer int64BufferPool.Put(bufp)
	buf := *bufp
	idx := encodeInt64(buf, x, upper, digits)
	if digits < int64BufferLen {
		return string(buf[idx:])
	}
	var b strings.Builder
	if x < 0 {
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

// encodeInt64 encodes x in hexadecimal representation into buf
// with given parameters.
//
// buf is the buffer obtained from int64BufferPool.
// upper indicates to use uppercase in hexadecimal representation.
// digits specifies the minimum length of the output content
// (excluding the negative sign "-").
// It pads with leading zeros after the sign (if any)
// if the length is not enough.
// If digits is non-positive, no padding will be applied.
// Note that, if digits is not less than int64BufferLen,
// no padding will be applied in this function,
// and only the hexadecimal representation of x without any sign will be
// written to buf.
// In this case, the caller should add any needed sign and leading zeros
// by themself after calling this function.
//
// It returns the start index of the valid content in buf.
// This function writes the encoding result to buf[idx:].
//
// The caller should guarantee that (x != 0 || digits > 1) and
// x != math.MinInt64 (= -0x8000000000000000).
// These two special cases should be handled by the caller.
func encodeInt64(buf []byte, x int64, upper bool, digits int) (idx int) {
	ht := lowercaseHexTable
	if upper {
		ht = uppercaseHexTable
	}
	isNeg := false
	if x < 0 {
		isNeg = true
		x = -x
	}
	idx = int64BufferLen
	for x != 0 {
		idx--
		buf[idx] = ht[x&0x0f]
		x >>= 4
	}
	if digits >= int64BufferLen {
		return
	}
	for idx > int64BufferLen-digits {
		idx--
		buf[idx] = '0'
	}
	if isNeg {
		idx--
		buf[idx] = '-'
	}
	return
}
