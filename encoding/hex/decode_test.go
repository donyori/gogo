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

package hex_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/encoding/hex"
)

func TestDecodedLen(t *testing.T) {
	for _, tc := range testEncodeCases {
		if tc.upper { // only use the lower cases to avoid redundant sources
			continue
		}
		t.Run("dst="+tc.dstName, func(t *testing.T) {
			t.Run("type=int", func(t *testing.T) {
				n := hex.DecodedLen(len(tc.dstStr))
				if n != len(tc.srcStr) {
					t.Errorf("got %d; want %d", n, len(tc.srcStr))
				}
			})
			t.Run("type=int64", func(t *testing.T) {
				n := hex.DecodedLen(int64(len(tc.dstStr)))
				if n != int64(len(tc.srcStr)) {
					t.Errorf("got %d; want %d", n, len(tc.srcStr))
				}
			})
		})
	}
}

func TestDecodedLen_Negative(t *testing.T) {
	t.Run("type=int", func(t *testing.T) {
		testDecodedLenNegative[int](t)
	})
	t.Run("type=int64", func(t *testing.T) {
		testDecodedLenNegative[int64](t)
	})
}

// testDecodedLenNegative is the common process of
// the subtests of TestDecodedLen_Negative.
func testDecodedLenNegative[Int constraints.SignedInteger](t *testing.T) {
	var x Int = -1
	defer func() {
		if e := recover(); e != nil {
			msg, ok := e.(string)
			if !ok || !strings.HasSuffix(
				msg, fmt.Sprintf("x (%d) is negative", x)) {
				t.Error(e)
			}
		}
	}()
	got := hex.DecodedLen(x) // want panic here
	t.Errorf("want panic but got %d (%#[1]x)", got)
}

func TestDecodedLen_Odd(t *testing.T) {
	t.Run("type=int", func(t *testing.T) {
		testDecodedLenOdd[int](t)
	})
	t.Run("type=uint", func(t *testing.T) {
		testDecodedLenOdd[uint](t)
	})
	t.Run("type=int64", func(t *testing.T) {
		testDecodedLenOdd[int64](t)
	})
	t.Run("type=uint64", func(t *testing.T) {
		testDecodedLenOdd[uint64](t)
	})
}

// testDecodedLenOdd is the common process of
// the subtests of TestDecodedLen_Odd.
func testDecodedLenOdd[Int constraints.Integer](t *testing.T) {
	var x Int = 3
	defer func() {
		if e := recover(); e != nil {
			msg, ok := e.(string)
			if !ok || !strings.HasSuffix(
				msg, fmt.Sprintf("x (%d) is odd", x)) {
				t.Error(e)
			}
		}
	}()
	got := hex.DecodedLen(x) // want panic here
	t.Errorf("want panic but got %d (%#[1]x)", got)
}
