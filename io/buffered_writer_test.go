// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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

package io

import (
	"bufio"
	"bytes"
	"testing"
)

func TestNewBufferedWriterSize(t *testing.T) {
	w := new(bytes.Buffer)
	bufw := bufio.NewWriter(w)
	bufw128 := bufio.NewWriterSize(w, 128)
	bufw64 := bufio.NewWriterSize(w, 64)
	bw := NewBufferedWriter(w)
	bw128 := NewBufferedWriterSize(w, 128)
	bw64 := NewBufferedWriterSize(w, 64)
	const size = 128
	b := NewBufferedWriterSize(w, size)
	if n := b.Size(); n != size {
		t.Errorf("(on w) size: %d != %d.", n, size)
	}
	b = NewBufferedWriterSize(bufw, size)
	if b != bufw {
		t.Error("(on bufw) b != bufw.")
	}
	b = NewBufferedWriterSize(bufw128, size)
	if b != bufw128 {
		t.Error("(on bufw128) b != bufw128.")
	}
	b = NewBufferedWriterSize(bufw64, size)
	if n := b.Size(); n != size {
		t.Errorf("(on bufw64) size: %d != %d.", n, size)
	}
	b = NewBufferedWriterSize(bw, size)
	if b != bw {
		t.Error("(on bw) b != bw.")
	}
	b = NewBufferedWriterSize(bw128, size)
	if b != bw128 {
		t.Error("(on bw128) b != bw128.")
	}
	b = NewBufferedWriterSize(bw64, size)
	if n := b.Size(); n != size {
		t.Errorf("(on bw64) size: %d != %d.", n, size)
	}

	b = NewBufferedWriterSize(w, 0)
	if n := b.Size(); n != defaultBufferSize {
		t.Errorf("(on w, size 0) size: %d != defaultBufferSize (%d).", n, defaultBufferSize)
	}
}
