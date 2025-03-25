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

package inout_test

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/donyori/gogo/inout"
)

func TestResettableBufferedReader_Reset(t *testing.T) {
	const BufferSize int = 256
	data := strings.Repeat("z", BufferSize+10)
	testCases := []struct {
		name         string
		toItself     bool
		r            io.Reader
		wantBuffered int
	}{
		{"toItself", true, nil, BufferSize - 1},
		{"<nil>", false, nil, 0},
		{"(*bufio.Reader)(nil)", false, (*bufio.Reader)(nil), 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			br := inout.NewBufferedReaderSize(
				strings.NewReader(data), BufferSize)
			_, err := br.ReadByte() // read one byte to fill the buffer
			if err != nil {
				t.Fatal(err)
			}
			buffered := br.Buffered()
			if buffered != BufferSize-1 {
				t.Fatalf("got Buffered %d after reading one byte; want %d",
					buffered, BufferSize-1)
			} else if tc.toItself {
				br.Reset(br)
			} else {
				br.Reset(tc.r)
			}
			if newBuffered := br.Buffered(); newBuffered != tc.wantBuffered {
				t.Errorf("got Buffered %d after resetting; want %d",
					newBuffered, tc.wantBuffered)
			}
		})
	}
}

func TestResettableBufferedWriter_Reset(t *testing.T) {
	const BufferSize int = 256
	testCases := []struct {
		name          string
		toItself      bool
		w             io.Writer
		wantBuffered  int
		wantAvailable int
	}{
		{"toItself", true, nil, 1, BufferSize - 1},
		{"<nil>", false, nil, 0, BufferSize},
		{"(*bufio.Writer)(nil)", false, (*bufio.Writer)(nil), 0, BufferSize},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bw := inout.NewBufferedWriterSize(io.Discard, BufferSize)
			err := bw.WriteByte('z') // write one byte to fill the buffer
			if err != nil {
				t.Fatal(err)
			}
			buffered := bw.Buffered()
			if buffered != 1 {
				t.Fatalf("got Buffered %d after writing one byte; want 1",
					buffered)
			}
			available := bw.Available()
			if available != BufferSize-1 {
				t.Fatalf("got Available %d after writing one byte; want %d",
					available, BufferSize-1)
			} else if tc.toItself {
				bw.Reset(bw)
			} else {
				bw.Reset(tc.w)
			}
			if newBuffered := bw.Buffered(); newBuffered != tc.wantBuffered {
				t.Errorf("got Buffered %d after resetting; want %d",
					newBuffered, tc.wantBuffered)
			}
			if newAvailable := bw.Available(); newAvailable != tc.wantAvailable {
				t.Errorf("got Available %d after resetting; want %d",
					newAvailable, tc.wantAvailable)
			}
		})
	}
}
