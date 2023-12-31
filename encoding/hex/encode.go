// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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

	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/errors"
)

// EncodedLen returns the length of encoding of n source bytes, exactly n * 2.
func EncodedLen[Int constraints.Integer](n Int) Int {
	return n * 2
}

// Encode encodes src in hexadecimal representation to dst.
//
// It panics if dst doesn't have enough space to hold the encoding result.
// The client should guarantee that len(dst) >= EncodedLen(len(src)).
//
// upper indicates whether to use uppercase in hexadecimal representation.
//
// It returns the number of bytes written into dst,
// exactly EncodedLen(len(src)).
//
// Encode[[]byte](dst, src, false) is equivalent to Encode(dst, src)
// in official package encoding/hex.
func Encode[Bytes constraints.ByteString](
	dst []byte,
	src Bytes,
	upper bool,
) int {
	if reqLen := EncodedLen(len(src)); reqLen > len(dst) {
		panic(errors.AutoMsg(fmt.Sprintf(
			"dst is too small, length: %d, required: %d", len(dst), reqLen)))
	}
	return encode(dst, src, upper)
}

// EncodeToString returns hexadecimal encoding of src.
//
// upper indicates whether to use uppercase in hexadecimal representation.
//
// EncodeToString[[]byte](src, false) is equivalent to EncodeToString(src)
// in official package encoding/hex.
func EncodeToString[Bytes constraints.ByteString](
	src Bytes,
	upper bool,
) string {
	dst := make([]byte, EncodedLen(len(src)))
	encode(dst, src, upper)
	return string(dst)
}

// Encoder is a device to write hexadecimal encoding of input data
// to the destination writer.
//
// It combines io.Writer, io.ByteWriter, io.StringWriter, and io.ReaderFrom.
// All the methods write hexadecimal encoding of input data to
// the destination writer.
type Encoder interface {
	io.Writer
	io.ByteWriter
	io.StringWriter
	io.ReaderFrom

	// EncodeDst returns the destination writer of this encoder.
	EncodeDst() io.Writer
}

// encoder is an implementation of interface Encoder.
type encoder struct {
	w     io.Writer
	upper bool
}

// NewEncoder creates an encoder to write hexadecimal characters to w.
//
// upper indicates whether to use uppercase in hexadecimal representation.
//
// It panics if w is nil.
func NewEncoder(w io.Writer, upper bool) Encoder {
	if w == nil {
		panic(errors.AutoMsg("w is nil"))
	}
	return &encoder{
		w:     w,
		upper: upper,
	}
}

func (enc *encoder) Write(p []byte) (n int, err error) {
	buf := encodeBufferPool.Get().(*[sourceBufferLen * 2]byte)
	defer encodeBufferPool.Put(buf)
	size := sourceBufferLen
	for len(p) > 0 && err == nil {
		if len(p) < size {
			size = len(p)
		}
		encoded := encode(buf[:], p[:size], enc.upper)
		var written int
		written, err = enc.w.Write(buf[:encoded])
		n += DecodedLen(written)
		p = p[size:]
	}
	return n, errors.AutoWrap(err)
}

func (enc *encoder) WriteByte(c byte) error {
	buf := encodeBufferPool.Get().(*[sourceBufferLen * 2]byte)
	defer encodeBufferPool.Put(buf)
	buf[0] = c
	encoded := encode(buf[1:], buf[:1], enc.upper)
	_, err := enc.w.Write(buf[1 : 1+encoded])
	return errors.AutoWrap(err)
}

func (enc *encoder) WriteString(s string) (n int, err error) {
	buf := encodeBufferPool.Get().(*[sourceBufferLen * 2]byte)
	defer encodeBufferPool.Put(buf)
	size := sourceBufferLen
	for len(s) > 0 && err == nil {
		if len(s) < size {
			size = len(s)
		}
		encoded := encode(buf[:], s[:size], enc.upper)
		var written int
		written, err = enc.w.Write(buf[:encoded])
		n += DecodedLen(written)
		s = s[size:]
	}
	return n, errors.AutoWrap(err)
}

func (enc *encoder) ReadFrom(r io.Reader) (n int64, err error) {
	buf := sourceBufferPool.Get().(*[sourceBufferLen]byte)
	defer sourceBufferPool.Put(buf)
	for {
		readLen, readErr := r.Read(buf[:])
		var writeErr error
		if readLen > 0 {
			n += int64(readLen)
			_, writeErr = enc.Write(buf[:readLen])
		}
		err = readErr
		if errors.Is(err, io.EOF) {
			err = nil
		}
		if readErr != nil {
			if err != nil {
				return n, errors.AutoWrap(err)
			}
			return n, errors.AutoWrap(writeErr)
		} else if writeErr != nil {
			return n, errors.AutoWrap(writeErr)
		}
	}
}

func (enc *encoder) EncodeDst() io.Writer {
	return enc.w
}

// encode is an implementation of function Encode,
// without checking the length of dst.
//
// Caller should guarantee that len(dst) >= EncodedLen(len(src)).
func encode[Bytes constraints.ByteString](
	dst []byte,
	src Bytes,
	upper bool,
) int {
	ht := lowercaseHexTable
	if upper {
		ht = uppercaseHexTable
	}
	end := EncodedLen(len(src))
	var n int
	for n < end {
		dst[n] = ht[src[n>>1]>>4]
		dst[n+1] = ht[src[n>>1]&0x0f]
		n += 2
	}
	return n
}
