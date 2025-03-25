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

package uintconv_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/donyori/gogo/encoding/varnum/uintconv"
)

var int64Uint64ZigzagPairs = []struct {
	i int64
	u uint64
}{
	{0, 0},
	{-1, 1},
	{1, 2},
	{-2, 3},
	{2147483647, 4294967294},
	{-2147483648, 4294967295},
	{math.MaxInt64, math.MaxUint64 - 1},
	{math.MinInt64, math.MaxUint64},
}

func TestFromInt64Zigzag(t *testing.T) {
	for _, pair := range int64Uint64ZigzagPairs {
		t.Run(fmt.Sprintf("i=%#X", pair.i), func(t *testing.T) {
			if got := uintconv.FromInt64Zigzag(pair.i); got != pair.u {
				t.Errorf("got %#X; want %#X", got, pair.u)
			}
		})
	}
}

func TestToInt64Zigzag(t *testing.T) {
	for _, pair := range int64Uint64ZigzagPairs {
		t.Run(fmt.Sprintf("u=%#X", pair.u), func(t *testing.T) {
			if got := uintconv.ToInt64Zigzag(pair.u); got != pair.i {
				t.Errorf("got %#X; want %#X", got, pair.i)
			}
		})
	}
}
