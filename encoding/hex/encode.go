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

package hex

import (
	"io"
	"sync"
)

const hexTable = "0123456789ABCDEF0123456789abcdef"

// Return the length of encoding of n source bytes, exactly n * 2.
func EncodedLen(n int) int {
	return n * 2
}

// Encode src into dst.
// upper indicates to use upper case in hexadecimal representation.
// It returns the number of bytes written into dst, exactly EncodedLen(len(src)).
// Encode(dst, src, false) is equivalent to Encode(dst, src) in official package encoding/hex.
func Encode(dst, src []byte, upper bool) int {
	var ht string
	if upper {
		ht = hexTable[:16]
	} else {
		ht = hexTable[16:]
	}
	var n int
	for _, b := range src {
		dst[n] = ht[b>>4]
		dst[n+1] = ht[b&0x0f]
		n += 2
	}
	return n
}

// Return hexadecimal encoding of src.
// upper indicates to use upper case in hexadecimal representation.
// EncodeToString(src, false) is equivalent to EncodeToString(src) in official package encoding/hex.
func EncodeToString(src []byte, upper bool) string {
	dst := make([]byte, EncodedLen(len(src)))
	Encode(dst, src, upper)
	return string(dst)
}

const chunkSize = 512

var chunkPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, chunkSize)
		return &b
	},
}

var encodeBufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, EncodedLen(chunkSize))
		return &b
	},
}

// An encoder to write hexadecimal characters.
type Encoder struct {
	w     io.Writer
	upper bool
}

// Create an encoder to write hexadecimal characters to w.
// upper indicates to use upper case in hexadecimal representation.
func NewEncoder(w io.Writer, upper bool) *Encoder {
	return &Encoder{
		w:     w,
		upper: upper,
	}
}

func (e *Encoder) Write(p []byte) (n int, err error) {
	size := chunkSize
	bufp := encodeBufferPool.Get().(*[]byte)
	buf := *bufp
	for len(p) > 0 && err == nil {
		if len(p) < size {
			size = len(p)
		}
		encoded := Encode(buf[:], p[:size], e.upper)
		var written int
		written, err = e.w.Write(buf[:encoded])
		n += written / 2
		p = p[size:]
	}
	encodeBufferPool.Put(bufp)
	return n, err
}

func (e *Encoder) WriteByte(c byte) error {
	bufp := encodeBufferPool.Get().(*[]byte)
	buf := *bufp
	buf[0] = c
	encoded := Encode(buf[1:], buf[:1], e.upper)
	_, err := e.w.Write(buf[1 : 1+encoded])
	encodeBufferPool.Put(bufp)
	return err
}

func (e *Encoder) ReadFrom(r io.Reader) (n int64, err error) {
	bufp := chunkPool.Get().(*[]byte)
	defer func() {
		chunkPool.Put(bufp)
	}()
	buf := *bufp
	var readSize int
	var readErr, writeErr error
	for {
		readSize, readErr = r.Read(buf)
		if readSize > 0 {
			n += int64(readSize)
			_, writeErr = e.Write(buf[:readSize])
		}
		err = readErr
		if err == io.EOF {
			err = nil
		}
		if readErr != nil {
			if err != nil {
				return n, err
			}
			return n, writeErr
		} else if writeErr != nil {
			return n, writeErr
		}
	}
}
