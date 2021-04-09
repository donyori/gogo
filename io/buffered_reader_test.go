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
	"errors"
	"io"
	"strings"
	"testing"
)

func TestNewBufferedReaderSize(t *testing.T) {
	r := bytes.NewReader([]byte("123456"))
	bufr := bufio.NewReader(r)
	bufr128 := bufio.NewReaderSize(r, 128)
	bufr64 := bufio.NewReaderSize(r, 64)
	br := NewBufferedReader(r)
	br128 := NewBufferedReaderSize(r, 128)
	br64 := NewBufferedReaderSize(r, 64)
	const size = 128
	b := NewBufferedReaderSize(r, size)
	if n := b.Size(); n != size {
		t.Errorf("(on r) size: %d != %d.", n, size)
	}
	b = NewBufferedReaderSize(bufr, size)
	if b.(*resettableBufferedReader).br != bufr {
		t.Error("(on bufr) b.br != bufr.")
	}
	b = NewBufferedReaderSize(bufr128, size)
	if b.(*resettableBufferedReader).br != bufr128 {
		t.Error("(on bufr128) b.br != bufr128.")
	}
	b = NewBufferedReaderSize(bufr64, size)
	if n := b.Size(); n != size {
		t.Errorf("(on bufr64) size: %d != %d.", n, size)
	}
	b = NewBufferedReaderSize(br, size)
	if b != br {
		t.Error("(on br) b != br.")
	}
	b = NewBufferedReaderSize(br128, size)
	if b != br128 {
		t.Error("(on br128) b != br128.")
	}
	b = NewBufferedReaderSize(br64, size)
	if n := b.Size(); n != size {
		t.Errorf("(on br64) size: %d != %d.", n, size)
	}

	b = NewBufferedReaderSize(r, 0)
	if n := b.Size(); n != minReadBufferSize {
		t.Errorf("(on r, size 0) size: %d != minReadBufferSize (%d).", n, minReadBufferSize)
	}
}

func TestBufferedReader_WriteLineTo(t *testing.T) {
	var longLineBuilder strings.Builder
	longLineBuilder.Grow(16390)
	for longLineBuilder.Len() < 16384 {
		longLineBuilder.WriteString("12345678")
	}
	longLineBuilder.WriteByte('\n')
	data := make([]byte, 0, 65560)
	for i := 0; i < 4; i++ {
		data = append(data, longLineBuilder.String()...)
	}
	data = data[:len(data)-1]
	br := NewBufferedReader(bytes.NewReader(data))
	var output strings.Builder
	output.Grow(longLineBuilder.Len() + 100)
	var err error
	for err == nil {
		output.Reset()
		_, err = br.WriteLineTo(&output)
		if err == nil {
			output.WriteByte('\n')
			if output.String() != longLineBuilder.String() {
				t.Errorf("Output line wrong. Line length: %d. Line: %q\nWanted: %q.", output.Len(), output.String(), longLineBuilder.String())
			}
		} else if !errors.Is(err, io.EOF) {
			t.Errorf("Error: %v.", err)
		}
	}
}
