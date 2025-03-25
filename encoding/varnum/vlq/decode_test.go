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

package vlq_test

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/donyori/gogo/encoding/varnum/uintconv"
	"github.com/donyori/gogo/encoding/varnum/vlq"
)

func TestDecodeUint64(t *testing.T) {
	for i, src := range encodedUint64s {
		t.Run(fmt.Sprintf("src=%#X", src), func(t *testing.T) {
			u, n, err := vlq.DecodeUint64(src)
			if err != nil {
				t.Fatal(err)
			}
			if u != uint64s[i] {
				t.Fatalf("got %#X; want %#X", u, uint64s[i])
			}
			if wantN := vlq.Uint64EncodedLen(u); n != wantN {
				t.Errorf("got n %d; want %d", n, wantN)
			}
		})
	}
	for _, src := range incompleteSrcs {
		var name string
		if len(src) == 0 {
			if src == nil {
				name = "src=<nil>(Incomplete)"
			} else {
				name = "src=[](Incomplete)"
			}
		} else {
			name = fmt.Sprintf("src=%#X(Incomplete)", src)
		}
		t.Run(name, func(t *testing.T) {
			u, n, err := vlq.DecodeUint64(src)
			if !errors.Is(err, vlq.ErrSrcIncomplete) {
				t.Fatalf("got %#X, %v", u, err)
			}
			if n != 0 {
				t.Errorf("got n %d; want 0", n)
			}
		})
	}
	for _, src := range tooLargeSrcs {
		t.Run(fmt.Sprintf("src=%#X(TooLarge)", src), func(t *testing.T) {
			u, n, err := vlq.DecodeUint64(src)
			if !errors.Is(err, vlq.ErrSrcTooLarge) {
				t.Fatalf("got %#X, %v", u, err)
			}
			if n != 0 {
				t.Errorf("got n %d; want 0", n)
			}
		})
	}
}

func TestDecodeInt64(t *testing.T) {
	for i, src := range encodedUint64s {
		t.Run(fmt.Sprintf("src=%#X", src), func(t *testing.T) {
			x, n, err := vlq.DecodeInt64(src)
			if err != nil {
				t.Fatal(err)
			}
			if want := uintconv.ToInt64Zigzag(uint64s[i]); x != want {
				t.Fatalf("got %#X; want %#X", x, want)
			}
			if wantN := vlq.Int64EncodedLen(x); n != wantN {
				t.Errorf("got n %d; want %d", n, wantN)
			}
		})
	}
}

func TestDecodeFloat64(t *testing.T) {
	for i, src := range encodedUint64s {
		t.Run(fmt.Sprintf("src=%#X", src), func(t *testing.T) {
			f, n, err := vlq.DecodeFloat64(src)
			if err != nil {
				t.Fatal(err)
			}
			want := uintconv.ToFloat64ByteReversal(uint64s[i])
			if math.IsNaN(want) {
				if !math.IsNaN(f) {
					t.Fatalf("got %v (bits: %#016X); want NaN",
						f, math.Float64bits(f))
				}
			} else if f != want {
				t.Fatalf(
					"got %v (bits: %#016X); want %v (bits: %#016X)",
					f,
					math.Float64bits(f),
					want,
					math.Float64bits(want),
				)
			}
			if wantN := vlq.Float64EncodedLen(f); n != wantN {
				t.Errorf("got n %d; want %d", n, wantN)
			}
		})
	}
}
