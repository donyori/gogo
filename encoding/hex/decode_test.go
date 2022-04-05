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

package hex

import "testing"

func TestDecodedLen(t *testing.T) {
	for _, c := range testEncodeCases {
		if n := DecodedLen(len(c.dst)); n != len(c.src) {
			t.Errorf("DecodedLen: %d != %d, src: %q, dst: %q, upper: %t.", n, len(c.src), c.src, c.dst, c.upper)
		}
	}
}

func TestDecodedLen64(t *testing.T) {
	for _, c := range testEncodeCases {
		if n := DecodedLen64(int64(len(c.dst))); n != int64(len(c.src)) {
			t.Errorf("DecodedLen: %d != %d, src: %q, dst: %q, upper: %t.", n, len(c.src), c.src, c.dst, c.upper)
		}
	}
}
