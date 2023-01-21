// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2023  Yuan Gao
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
	"io"
	"math"
	"strings"

	"github.com/donyori/gogo/errors"
)

// minInt64Hex is the hexadecimal representation of math.MinInt64, as a string.
const minInt64Hex = "-8000000000000000"

// EncodeInt64DstLen returns a safe length of dst used in EncodeInt64
// to ensure that dst has enough space to keep the encoding result.
//
// digits is the parameter used in EncodeInt64 to specify the minimum length
// of the output content (excluding the negative sign "-").
func EncodeInt64DstLen(digits int) int {
	if digits <= 16 {
		return 17
	}
	return digits + 1
}

// EncodeInt64 encodes x in hexadecimal representation to dst.
//
// upper indicates whether to use uppercase in hexadecimal representation.
// digits specify the minimum length of the output content
// (excluding the negative sign "-").
// It pads with leading zeros after the sign (if any)
// if the length is not enough.
// If digits is non-positive, no padding will be applied.
//
// It returns the number of bytes written into dst.
//
// It panics if dst is too small to keep the result.
// To ensure that dst has enough space to keep the result,
// its length should be at least EncodeInt64DstLen(digits).
func EncodeInt64(dst []byte, x int64, upper bool, digits int) int {
	dstTooSmallMsgFn := func(required int) string {
		return fmt.Sprintf("dst is too small, length: %d, required: %d", len(dst), required)
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
	buf := int64BufferPool.Get().(*[int64BufferLen]byte)
	defer int64BufferPool.Put(buf)
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
// upper indicates whether to use uppercase in hexadecimal representation.
// digits specify the minimum length of the return string
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
	buf := int64BufferPool.Get().(*[int64BufferLen]byte)
	defer int64BufferPool.Put(buf)
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

// EncodeInt64To encodes x in hexadecimal representation to w.
//
// upper indicates whether to use uppercase in hexadecimal representation.
// digits specify the minimum length of the output content
// (excluding the negative sign "-").
// It pads with leading zeros after the sign (if any)
// if the length is not enough.
// If digits is non-positive, no padding will be applied.
//
// It returns the number of bytes written to w, and any write error encountered.
func EncodeInt64To(w io.Writer, x int64, upper bool, digits int) (written int, err error) {
	if w == nil {
		panic(errors.AutoMsg("w is nil"))
	}

	buf := int64BufferPool.Get().(*[int64BufferLen]byte)
	defer int64BufferPool.Put(buf)

	// Special cases for 0 and -0x8000000000000000 (minimum value of int64):
	if x == 0 && digits <= 1 {
		if bw, ok := w.(io.ByteWriter); ok {
			err = errors.AutoWrap(bw.WriteByte('0'))
			if err == nil {
				written = 1
			}
			return
		}
		buf[0] = '0'
		written, err = w.Write(buf[:1])
		return written, errors.AutoWrap(err)
	}
	var n int
	if x == math.MinInt64 {
		if digits < int64BufferLen {
			if sw, ok := w.(io.StringWriter); ok {
				written, err = sw.WriteString(minInt64Hex)
			} else {
				n := copy(buf[:], minInt64Hex)
				written, err = w.Write(buf[:n])
			}
			return written, errors.AutoWrap(err)
		}
		written, err = writeSignAndLeadingZerosTo(w, x, digits, buf)
		if err != nil {
			return written, errors.AutoWrap(err)
		}
		buf[0] = '8'
		n, err = w.Write(buf[:16])
		written += n
		return written, errors.AutoWrap(err)
	}

	// Other cases:
	if digits >= int64BufferLen {
		written, err = writeSignAndLeadingZerosTo(w, x, digits, buf)
		if err != nil {
			return written, errors.AutoWrap(err)
		}
	}
	idx := encodeInt64(buf, x, upper, digits)
	if digits >= int64BufferLen {
		// Items in buf[1:idx] are '0'
		// set by function writeSignAndLeadingZerosTo.
		// buf[idx:] is the hexadecimal representation of x without any sign,
		// set by function encodeInt64.
		// So, buf[1:] is the hexadecimal representation of x
		// with (16 - idx) leading zeros.
		// The negative sign (if x < 0) and other leading zeros have already
		// been written to w by function writeSignAndLeadingZerosTo.
		// Therefore, just write buf[1:] to w and everything is done.
		idx = 1
	}
	n, err = w.Write(buf[idx:])
	written += n
	return written, errors.AutoWrap(err)
}

// encodeInt64 encodes x in hexadecimal representation into buf
// with specified parameters.
//
// buf is the buffer obtained from int64BufferPool.
// upper indicates whether to use uppercase in hexadecimal representation.
// digits specify the minimum length of the output content
// (excluding the negative sign "-").
// It pads with leading zeros after the sign (if any)
// if the length is not enough.
// If digits is non-positive, no padding will be applied.
// In addition, if digits is not less than int64BufferLen,
// no padding will be applied in this function,
// and only the hexadecimal representation of x without any sign will be
// written to buf.
// In this case, the caller should add any needed sign and leading zeros
// after calling this function.
//
// It returns the start index of the valid content in buf.
// This function writes the encoding result to buf[idx:].
// This function will not access (including reading and writing)
// the rest part of the buffer (i.e., buf[:idx]).
//
// Caller should guarantee that (x != 0 || digits > 1) and
// x != math.MinInt64 (= -0x8000000000000000).
// These two special cases should be handled by the caller.
func encodeInt64(buf *[int64BufferLen]byte, x int64, upper bool, digits int) (idx int) {
	ht := lowercaseHexTable
	if upper {
		ht = uppercaseHexTable
	}
	var isNeg bool
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

// writeSignAndLeadingZerosTo is for function EncodeInt64To.
// It writes the negative sign (if x < 0) and
// (digits - 16) leading zeros to w, through using the buffer buf.
//
// w, x, digits are the same as that of its caller.
// buf is obtained from int64BufferPool, with size int64BufferLen (17).
// This function sets all items in buf[1:] to character '0'.
// It sets buf[0] to '-' if x < 0 and digits <= int64BufferLen + 15,
// otherwise, '0'.
//
// It returns the number of bytes written to w,
// and any write error encountered.
//
// Caller should guarantee that w != nil, digits >= int64BufferLen,
// and buf != nil.
func writeSignAndLeadingZerosTo(w io.Writer, x int64, digits int, buf *[int64BufferLen]byte) (written int, err error) {
	if x >= 0 {
		buf[0] = '0'
	} else {
		buf[0] = '-'
	}
	for i := 1; i < int64BufferLen; i++ {
		buf[i] = '0'
	}
	ctr := digits - 16 // counter for the number of leading zeros remaining to be written
	if x < 0 {
		if ctr <= int64BufferLen-1 {
			written, err = w.Write(buf[:ctr+1])
			return written, errors.AutoWrap(err)
		}
		written, err = w.Write(buf[:])
		buf[0] = '0'
		if err != nil {
			return written, errors.AutoWrap(err)
		}
		ctr -= written - 1
	}
	var n int // for the return value of w.Write
	for ctr > 0 {
		if ctr > int64BufferLen {
			n, err = w.Write(buf[:])
		} else {
			n, err = w.Write(buf[:ctr])
		}
		written, ctr = written+n, ctr-n
		if err != nil {
			return written, errors.AutoWrap(err)
		}
	}
	return // err must be nil, so don't need errors.AutoWrap(err)
}
