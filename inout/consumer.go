// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2025  Yuan Gao
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

package inout

// ByteConsumer contains methods to consume
// (read and discard) bytes from an io.Reader.
type ByteConsumer interface {
	// ConsumeByte repeats detecting the next byte in the underlying reader
	// and consuming it until the byte is not the specified byte
	// or ConsumeByte has consumed n bytes (if n is positive).
	//
	// If n is zero, it does nothing and returns (0, nil).
	// If n is negative, it consumes the bytes with no limit.
	//
	// It returns the number of bytes consumed and
	// any error encountered (including io.EOF).
	ConsumeByte(target byte, n int64) (consumed int64, err error)

	// ConsumeByteFunc repeats detecting the next byte in the underlying reader
	// and consuming it until the byte does not satisfy the specified function f
	// or ConsumeByteFunc has consumed n bytes (if n is positive).
	//
	// If n is zero, it does nothing and returns (0, nil).
	// If n is negative, it consumes the bytes with no limit.
	//
	// It returns the number of bytes consumed and
	// any error encountered (including io.EOF).
	ConsumeByteFunc(f func(c byte) bool, n int64) (consumed int64, err error)
}

// RuneConsumer contains methods to consume (read and discard)
// runes (Unicode code points) from an io.Reader.
type RuneConsumer interface {
	// ConsumeRune repeats detecting the next rune in the underlying reader
	// and consuming it until the rune is not the specified rune
	// or ConsumeRune has consumed n runes (if n is positive).
	// In particular, if the rune is not a valid Unicode code point in UTF-8,
	// the rune is considered unicode.ReplacementChar (U+FFFD) with a size of 1.
	//
	// If n is zero, ConsumeRune does nothing and returns (0, nil).
	// If n is negative, it consumes the runes with no limit.
	//
	// It returns the number of runes consumed and
	// any error encountered (including io.EOF).
	ConsumeRune(target rune, n int64) (consumed int64, err error)

	// ConsumeRuneFunc repeats detecting the next rune in the underlying reader
	// and consuming it until the rune does not satisfy the specified function f
	// or ConsumeRuneFunc has consumed n runes (if n is positive).
	//
	// If n is zero, it does nothing and returns (0, nil).
	// If n is negative, it consumes the runes with no limit.
	//
	// The parameters of f are the rune and the size of the rune in bytes.
	// In particular, if the rune is not a valid Unicode code point in UTF-8,
	// f gets unicode.ReplacementChar (U+FFFD) with a size of 1.
	//
	// ConsumeRuneFunc returns the number of runes consumed and
	// any error encountered (including io.EOF).
	ConsumeRuneFunc(f func(r rune, size int) bool, n int64) (
		consumed int64, err error)
}
