// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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

// This file indirectly requires the unexported variable: minUint64s.

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/donyori/gogo/encoding/varnum/uintconv"
)

func TestDecodeUint64(t *testing.T) {
	for i, src := range encodedUint64s {
		t.Run(fmt.Sprintf("src=%X", src), func(t *testing.T) {
			u, err := DecodeUint64(src)
			if err != nil {
				t.Fatal(err)
			}
			if u != uint64s[i] {
				t.Errorf("got %#X; want %#X", u, uint64s[i])
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
			name = fmt.Sprintf("src=%X(Incomplete)", src)
		}
		t.Run(name, func(t *testing.T) {
			u, err := DecodeUint64(src)
			if !errors.Is(err, ErrSrcIncomplete) {
				t.Errorf("got %#X, %v", u, err)
			}
		})
	}
	for _, src := range tooLargeSrcs {
		t.Run(fmt.Sprintf("src=%X(TooLarge)", src), func(t *testing.T) {
			u, err := DecodeUint64(src)
			if !errors.Is(err, ErrSrcTooLarge) {
				t.Errorf("got %#X, %v", u, err)
			}
		})
	}
}

func TestDecodeInt64(t *testing.T) {
	for i, src := range encodedUint64s {
		t.Run(fmt.Sprintf("src=%X", src), func(t *testing.T) {
			x, err := DecodeInt64(src)
			if err != nil {
				t.Fatal(err)
			}
			if want := uintconv.ToInt64Zigzag(uint64s[i]); x != want {
				t.Errorf("got %#X; want %#X", x, want)
			}
		})
	}
}

func TestDecodeFloat64(t *testing.T) {
	for i, src := range encodedUint64s {
		t.Run(fmt.Sprintf("src=%X", src), func(t *testing.T) {
			f, err := DecodeFloat64(src)
			if err != nil {
				t.Fatal(err)
			}
			want := uintconv.ToFloat64ByteReversal(uint64s[i])
			if math.IsNaN(want) {
				if !math.IsNaN(f) {
					t.Errorf("got %f; want NaN", f)
				}
			} else if f != want {
				t.Errorf("got %f; want %f", f, want)
			}
		})
	}
}
