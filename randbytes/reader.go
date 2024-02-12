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

package randbytes

import (
	"io"
	"math/rand/v2"

	"github.com/donyori/gogo/errors"
)

// Reader is a pseudorandom byte generator.
//
// It combines io.Reader and io.ByteReader.
// Its method Read always returns the buffer length (i.e., len(p))
// and a nil error.
// Its method ReadByte always returns a nil error as well.
//
// It is based on the standard library math/rand/v2.
// During the call to its read methods,
// its random value source should not be used by others concurrently.
type Reader interface {
	io.Reader
	io.ByteReader

	// Source returns the random value source used by this reader.
	Source() rand.Source
}

// reader is an implementation of interface Reader.
type reader struct {
	x   uint64 // x is the last generated random number.
	ctr int    // ctr is a counter for updating x.

	src rand.Source // src is the random value source.
}

// NewReader creates a new pseudorandom byte generator
// with the specified random value source.
//
// During the call to the read methods of the returned Reader,
// the random value source should not be used by others concurrently.
//
// NewReader panics if the random value source is nil.
func NewReader(src rand.Source) Reader {
	if src == nil {
		panic(errors.AutoMsg("random value source is nil"))
	}
	return &reader{src: src}
}

func (r *reader) Read(p []byte) (n int, err error) {
	// Don't use "n = range p" or "n = range len(p)",
	// because these statements make n end at len(p)-1, not len(p).
	for n = 0; n < len(p); n++ {
		p[n], _ = r.ReadByte() // the error is always nil
	}
	return
}

func (r *reader) ReadByte() (c byte, err error) {
	if r.ctr == 0 {
		r.x, r.ctr = r.src.Uint64(), 8
	}
	c = byte(r.x)
	r.x, r.ctr = r.x>>8, r.ctr-1
	return
}

func (r *reader) Source() rand.Source {
	return r.src
}
