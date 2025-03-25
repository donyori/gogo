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

package randbytes

import (
	"fmt"
	"io"
	"math/rand/v2"
	"slices"

	"github.com/donyori/gogo/errors"
)

// Fill generates random bytes using the specified random value source
// and writes them into p.
//
// The random value source should not be used by others concurrently.
//
// Fill panics if the random value source is nil.
func Fill(src rand.Source, p []byte) {
	if src == nil {
		panic(errors.AutoMsg("random value source is nil"))
	}
	var x uint64
	for i := range p {
		if i&7 == 0 {
			x = src.Uint64()
		}
		p[i] = byte(x)
		x >>= 8
	}
}

// Make allocates a []byte with the specified length and
// fills it with random bytes generated from the specified random value source.
//
// The random value source should not be used by others concurrently.
//
// Make panics if the random value source is nil or the length is negative.
func Make(src rand.Source, length int) []byte {
	switch {
	case src == nil:
		panic(errors.AutoMsg("random value source is nil"))
	case length < 0:
		panic(errors.AutoMsg(fmt.Sprintf("length (%d) is negative", length)))
	case length == 0:
		return []byte{}
	}
	p := make([]byte, length)
	Fill(src, p)
	return p
}

// MakeCapacity is like Make but specifies a different capacity.
//
// The random value source should not be used by others concurrently.
//
// MakeCapacity panics if the random value source is nil,
// the length is negative, or the capacity is less than the length.
func MakeCapacity(src rand.Source, length int, capacity int) []byte {
	switch {
	case src == nil:
		panic(errors.AutoMsg("random value source is nil"))
	case length < 0:
		panic(errors.AutoMsg(fmt.Sprintf("length (%d) is negative", length)))
	case length > capacity:
		panic(errors.AutoMsg(fmt.Sprintf(
			"capacity (%d) is less than length (%d)", capacity, length)))
	case length == 0:
		return make([]byte, 0, capacity)
	}
	p := make([]byte, length, capacity)
	Fill(src, p)
	return p
}

// Append generates n random bytes using the specified random value source,
// appends them to p, and returns the extended byte slice.
//
// The random value source should not be used by others concurrently.
//
// Append panics if the random value source is nil or n is negative.
func Append(src rand.Source, p []byte, n int) []byte {
	switch {
	case src == nil:
		panic(errors.AutoMsg("random value source is nil"))
	case n < 0:
		panic(errors.AutoMsg(fmt.Sprintf("n (%d) is negative", n)))
	case n == 0:
		return p
	}
	p = slices.Grow(p, n)
	Fill(src, p[len(p):][:n])
	return p[:len(p)+n]
}

// WriteN generates n random bytes using the specified random value source
// and writes them into the specified byte writer.
//
// The random value source should not be used by others concurrently.
//
// WriteN returns the number of bytes written and any write error encountered.
//
// It panics if the random value source or the byte writer is nil
// or n is negative.
func WriteN(src rand.Source, w io.ByteWriter, n int) (written int, err error) {
	switch {
	case src == nil:
		panic(errors.AutoMsg("random value source is nil"))
	case w == nil:
		panic(errors.AutoMsg("byte writer is nil"))
	case n < 0:
		panic(errors.AutoMsg(fmt.Sprintf("n (%d) is negative", n)))
	}
	var x uint64
	for written < n {
		if written&7 == 0 {
			x = src.Uint64()
		}
		err = w.WriteByte(byte(x))
		if err != nil {
			return written, errors.AutoWrap(err)
		}
		written, x = written+1, x>>8
	}
	return
}
