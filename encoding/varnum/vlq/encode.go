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

package vlq

import (
	"fmt"
	"sync"

	"github.com/donyori/gogo/encoding/varnum/uintconv"
	"github.com/donyori/gogo/errors"
)

// minUint64s are the minimum 64-bit unsigned integers in the VLQ encoding
// of different lengths (from 2 to 10) implemented in this package.
//
// They are used by function Uint64EncodedLen to determine
// the encoding length of a specified input.
//
// They satisfy:
//
//	If u < minUint64s[0], then Uint64EncodedLen(u) = 1;
//	If minUint64s[i] <= u < minUint64s[i+1] (i=0,1,...,7), then Uint64EncodedLen(u) = i+2;
//	If u >= minUint64s[8], then Uint64EncodedLen(u) = 10.
var minUint64s = [...]uint64{
	0x80, 0x4080, 0x204080, 0x10204080, 0x810204080,
	0x40810204080, 0x2040810204080, 0x102040810204080, 0x8102040810204080,
}

// Uint64EncodedLen returns the length of variable-length quantity (VLQ)
// style encoding of the 64-bit unsigned integer u.
//
// It is at most 10.
func Uint64EncodedLen(u uint64) int {
	// It is more efficient to simply test items one by one
	// since the size of minUint64s is very small.
	for i, mx := range minUint64s {
		if u < mx {
			return i + 1
		}
	}
	return 10
}

// EncodeUint64 encodes the 64-bit unsigned integer u in variable-length
// quantity (VLQ) style encoding to dst.
//
// It panics if dst doesn't have enough space to hold the encoding result.
// The client should guarantee that len(dst) >= Uint64EncodedLen(u).
//
// It returns the number of bytes written into dst,
// exactly Uint64EncodedLen(u).
func EncodeUint64(dst []byte, u uint64) int {
	if reqLen := Uint64EncodedLen(u); reqLen > len(dst) {
		panic(errors.AutoMsg(fmt.Sprintf("dst is too small, length: %d, required: %d", len(dst), reqLen)))
	}
	return encodeUint64(dst, u)
}

// Int64EncodedLen returns the length of variable-length quantity (VLQ)
// style encoding with zigzag encoding of the 64-bit signed integer i.
//
// It is at most 10.
func Int64EncodedLen(i int64) int {
	return Uint64EncodedLen(uintconv.FromInt64Zigzag(i))
}

// EncodeInt64 encodes the 64-bit signed integer i in variable-length
// quantity (VLQ) style encoding with zigzag encoding to dst.
//
// It panics if dst doesn't have enough space to hold the encoding result.
// The client should guarantee that len(dst) >= Int64EncodedLen(i).
//
// It returns the number of bytes written into dst,
// exactly Int64EncodedLen(i).
func EncodeInt64(dst []byte, i int64) int {
	u := uintconv.FromInt64Zigzag(i)
	if reqLen := Uint64EncodedLen(u); reqLen > len(dst) {
		panic(errors.AutoMsg(fmt.Sprintf("dst is too small, length: %d, required: %d", len(dst), reqLen)))
	}
	return encodeUint64(dst, u)
}

// Float64EncodedLen returns the length of variable-length quantity (VLQ)
// style encoding of the corresponding byte-reversed 64-bit
// unsigned integer of the 64-bit floating-point number f.
//
// It is at most 10.
func Float64EncodedLen(f float64) int {
	return Uint64EncodedLen(uintconv.FromFloat64ByteReversal(f))
}

// EncodeFloat64 encodes the 64-bit floating-point number f in variable-length
// quantity (VLQ) style encoding.
// It converts f to a 64-bit unsigned integer.
// The integer is then byte-reversed and encoded as a regular unsigned integer.
//
// It panics if dst doesn't have enough space to hold the encoding result.
// The client should guarantee that len(dst) >= Float64EncodedLen(i).
//
// It returns the number of bytes written into dst,
// exactly Float64EncodedLen(i).
func EncodeFloat64(dst []byte, f float64) int {
	u := uintconv.FromFloat64ByteReversal(f)
	if reqLen := Uint64EncodedLen(u); reqLen > len(dst) {
		panic(errors.AutoMsg(fmt.Sprintf("dst is too small, length: %d, required: %d", len(dst), reqLen)))
	}
	return encodeUint64(dst, u)
}

// bufferLen is the length of the buffer in bufferPool.
const bufferLen = 10

// bufferPool is a set of temporary buffers used during encoding.
//
// The type of the buffers is *[bufferLen]byte.
//
// Each buffer is large enough to hold the encoding result of
// a 64-bit unsigned integer.
var bufferPool = sync.Pool{
	New: func() any {
		return new([bufferLen]byte)
	},
}

// encodeUint64 is an implementation of function EncodeUint64,
// without checking the length of dst.
//
// Caller should guarantee that len(dst) >= Uint64EncodedLen(u).
func encodeUint64(dst []byte, u uint64) int {
	if u < 0x80 {
		// If u does not exceed 7 bits, simply write itself to dst and return 1.
		dst[0] = byte(u)
		return 1
	}
	buf := bufferPool.Get().(*[bufferLen]byte)
	defer bufferPool.Put(buf)
	buf[0] = byte(u & 0x7F)
	t, n := u>>7, 1
	for t != 0 {
		t-- // Remove the prepending redundancy in typical VLQ.
		buf[n], n, t = byte(t&0x7F|0x80), n+1, t>>7
	}
	for i := 0; i < n; i++ {
		dst[i] = buf[n-1-i]
	}
	return n
}
