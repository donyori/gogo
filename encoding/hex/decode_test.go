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

package hex_test

import (
	"testing"

	"github.com/donyori/gogo/encoding/hex"
)

func TestDecodedLen(t *testing.T) {
	for _, tc := range testEncodeCases {
		if tc.upper { // only use the lower cases
			continue
		}
		t.Run("dst="+tc.dstName, func(t *testing.T) {
			if n := hex.DecodedLen(len(tc.dst)); n != len(tc.src) {
				t.Errorf("got %d; want %d", n, len(tc.src))
			}
		})
	}
}

func TestDecodedLen64(t *testing.T) {
	for _, tc := range testEncodeCases {
		if tc.upper { // only use the lower cases
			continue
		}
		t.Run("dst="+tc.dstName, func(t *testing.T) {
			if n := hex.DecodedLen64(int64(len(tc.dst))); n != int64(len(tc.src)) {
				t.Errorf("got %d; want %d", n, len(tc.src))
			}
		})
	}
}
