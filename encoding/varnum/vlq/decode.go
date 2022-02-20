// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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

var (
	// ErrSrcIncomplete is an error indicating that the source bytes src
	// is incomplete to be decoded with the variable-length quantity (VLQ)
	// style encoding.
	//
	// The client should use errors.Is to test whether
	// an error is ErrSrcIncomplete.
	ErrSrcIncomplete = errors.AutoNewCustom("src is incomplete",
		errors.PrependFullPkgName, 0)

	// ErrSrcTooLarge is an error indicating that the source bytes src is
	// too large to be decoded into a 64-bit number with the
	// variable-length quantity (VLQ) style encoding.
	//
	// The client should use errors.Is to test whether
	// an error is ErrSrcTooLarge.
	ErrSrcTooLarge = errors.AutoNewCustom("src is too large to be decoded into a 64-bit number",
		errors.PrependFullPkgName, 0)
)

// DecodeUint64 decodes src into a 64-bit unsigned integer
// with variable-length quantity (VLQ) style encoding.
//
// If src is nil or empty, it reports ErrSrcIncomplete.
//
// The decoding will stop when obtaining a 64-bit unsigned integer.
// Subsequent content in src (if any) will be ignored.
func DecodeUint64(src []byte) (u uint64, err error) {
	end := false
	for _, b := range src {
		if b&0x80 == 0 {
			u, end = u|uint64(b), true
			break
		}
		u = u | uint64(b&0x7F) + 1
		if u&0xFE00_0000_0000_0000 != 0 {
			return 0, errors.AutoWrap(ErrSrcTooLarge)
		}
		u <<= 7
	}
	if !end {
		return 0, errors.AutoWrap(ErrSrcIncomplete)
	}
	return
}

// DecodeInt64 decodes src into a 64-bit signed integer with
// variable-length quantity (VLQ) style encoding and zigzag encoding.
//
// If src is nil or empty, it reports ErrSrcIncomplete.
//
// The decoding will stop when obtaining a 64-bit signed integer.
// Subsequent content in src (if any) will be ignored.
func DecodeInt64(src []byte) (i int64, err error) {
	u, err := DecodeUint64(src)
	if err != nil {
		return 0, errors.AutoWrap(err)
	}
	return uintconv.ToInt64Zigzag(u), nil
}

// DecodeFloat64 decodes src into a 64-bit floating-point number
// with variable-length quantity (VLQ) style encoding.
// It first decodes src into a 64-bit unsigned integer.
// The integer is then byte-reversed and converted to
// a 64-bit floating-point number.
//
// If src is nil or empty, it reports ErrSrcIncomplete.
//
// The decoding will stop when obtaining a 64-bit floating-point number.
// Subsequent content in src (if any) will be ignored.
func DecodeFloat64(src []byte) (f float64, err error) {
	u, err := DecodeUint64(src)
	if err != nil {
		return 0, errors.AutoWrap(err)
	}
	return uintconv.ToFloat64ByteReversal(u), nil
}
