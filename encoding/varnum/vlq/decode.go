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
	"github.com/donyori/gogo/encoding/varnum/uintconv"
	"github.com/donyori/gogo/errors"
)

// ErrSrcIncomplete is an error indicating that the source bytes src
// is incomplete to be decoded with the variable-length quantity (VLQ)
// style encoding.
//
// The client should use errors.Is to test whether
// an error is ErrSrcIncomplete.
var ErrSrcIncomplete = errors.AutoNewCustom(
	"src is incomplete",
	errors.PrependFullPkgName,
	0,
)

// ErrSrcTooLarge is an error indicating that the source bytes src is
// too large to be decoded into a 64-bit number with the
// variable-length quantity (VLQ) style encoding.
//
// The client should use errors.Is to test whether
// an error is ErrSrcTooLarge.
var ErrSrcTooLarge = errors.AutoNewCustom(
	"src is too large to be decoded into a 64-bit number",
	errors.PrependFullPkgName,
	0,
)

// DecodeUint64 decodes src into a 64-bit unsigned integer
// with variable-length quantity (VLQ) style encoding.
//
// If src is nil or empty, it reports ErrSrcIncomplete.
//
// The decoding stops when obtaining a 64-bit unsigned integer.
// Subsequent content in src (if any) is ignored.
//
// It also returns the number of bytes read from src (n).
// If err is nil, n is exactly Uint64EncodedLen(u).
// If err is not nil, n is 0.
func DecodeUint64(src []byte) (u uint64, n int, err error) {
	end := false
	for _, b := range src {
		n++
		if b&0x80 == 0 {
			u, end = u|uint64(b), true
			break
		}
		u = u | uint64(b&0x7F) + 1
		if u&0xFE00_0000_0000_0000 != 0 {
			return 0, 0, errors.AutoWrap(ErrSrcTooLarge)
		}
		u <<= 7
	}
	if !end {
		return 0, 0, errors.AutoWrap(ErrSrcIncomplete)
	}
	return
}

// DecodeInt64 decodes src into a 64-bit signed integer with
// variable-length quantity (VLQ) style encoding and zigzag encoding.
//
// If src is nil or empty, it reports ErrSrcIncomplete.
//
// The decoding stops when obtaining a 64-bit signed integer.
// Subsequent content in src (if any) is ignored.
//
// It also returns the number of bytes read from src (n).
// If err is nil, n is exactly Int64EncodedLen(i).
// If err is not nil, n is 0.
func DecodeInt64(src []byte) (i int64, n int, err error) {
	u, n, err := DecodeUint64(src)
	if err != nil {
		return 0, 0, errors.AutoWrap(err)
	}
	return uintconv.ToInt64Zigzag(u), n, nil
}

// DecodeFloat64 decodes src into a 64-bit floating-point number
// with variable-length quantity (VLQ) style encoding.
// It first decodes src into a 64-bit unsigned integer.
// The integer is then byte-reversed and converted to
// a 64-bit floating-point number.
//
// If src is nil or empty, it reports ErrSrcIncomplete.
//
// The decoding stops when obtaining a 64-bit floating-point number.
// Subsequent content in src (if any) is ignored.
//
// It also returns the number of bytes read from src (n).
// If err is nil, n is exactly Float64EncodedLen(f).
// If err is not nil, n is 0.
func DecodeFloat64(src []byte) (f float64, n int, err error) {
	u, n, err := DecodeUint64(src)
	if err != nil {
		return 0, 0, errors.AutoWrap(err)
	}
	return uintconv.ToFloat64ByteReversal(u), n, nil
}
