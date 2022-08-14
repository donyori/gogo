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

package inout

// This file requires the unexported variables: minReadBufferSize and defaultBufferSize,
// and the unexported type: resettableBufferedReader.

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestNewBufferedReaderSize(t *testing.T) {
	r := strings.NewReader("123456")
	bufr := bufio.NewReader(r)
	bufr64 := bufio.NewReaderSize(r, 64)
	bufr128 := bufio.NewReaderSize(r, 128)
	bufr256 := bufio.NewReaderSize(r, 256)
	br := NewBufferedReader(r)
	br64 := NewBufferedReaderSize(r, 64)
	br128 := NewBufferedReaderSize(r, 128)
	br256 := NewBufferedReaderSize(r, 256)
	const size = 128

	testCases := []struct {
		name             string
		r                io.Reader
		argSize          int
		wantSize         int
		wantAsUnderlying bool
		wantAsSelf       bool
	}{
		{name: "on r", r: r, argSize: size, wantSize: size},
		{name: "on bufr", r: bufr, argSize: size, wantAsUnderlying: true},
		{name: "on bufr64", r: bufr64, argSize: size, wantSize: size},
		{name: "on bufr128", r: bufr128, argSize: size, wantAsUnderlying: true},
		{name: "on bufr256", r: bufr256, argSize: size, wantAsUnderlying: true},
		{name: "on br", r: br, argSize: size, wantAsSelf: true},
		{name: "on br64", r: br64, argSize: size, wantSize: size},
		{name: "on br128", r: br128, argSize: size, wantAsSelf: true},
		{name: "on br256", r: br256, argSize: size, wantAsSelf: true},
		{name: "on r?size=0", r: r, argSize: 0, wantSize: minReadBufferSize},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := NewBufferedReaderSize(tc.r, tc.argSize)
			switch true {
			case tc.wantSize != 0:
				if n := b.Size(); n != tc.wantSize {
					t.Errorf("got size %d; want %d", n, tc.wantSize)
				}
			case tc.wantAsUnderlying:
				if b.(*resettableBufferedReader).br != tc.r {
					t.Error("underlying buffered reader not input")
				}
			case tc.wantAsSelf:
				if b != tc.r {
					t.Error("buffered reader not input")
				}
			default:
				t.Errorf("invalid test case %+v", tc)
			}
		})
	}
}

func TestNewBufferedWriterSize(t *testing.T) {
	w := bytes.NewBuffer(make([]byte, 0, 10))
	bufw := bufio.NewWriter(w)
	bufw64 := bufio.NewWriterSize(w, 64)
	bufw128 := bufio.NewWriterSize(w, 128)
	bufw256 := bufio.NewWriterSize(w, 256)
	bw := NewBufferedWriter(w)
	bw64 := NewBufferedWriterSize(w, 64)
	bw128 := NewBufferedWriterSize(w, 128)
	bw256 := NewBufferedWriterSize(w, 256)
	const size = 128

	testCases := []struct {
		name       string
		w          io.Writer
		argSize    int
		wantSize   int
		wantAsSelf bool
	}{
		{name: "on w", w: w, argSize: size, wantSize: size},
		{name: "on bufw", w: bufw, argSize: size, wantAsSelf: true},
		{name: "on bufw64", w: bufw64, argSize: size, wantSize: size},
		{name: "on bufw128", w: bufw128, argSize: size, wantAsSelf: true},
		{name: "on bufw256", w: bufw256, argSize: size, wantAsSelf: true},
		{name: "on bw", w: bw, argSize: size, wantAsSelf: true},
		{name: "on bw64", w: bw64, argSize: size, wantSize: size},
		{name: "on bw128", w: bw128, argSize: size, wantAsSelf: true},
		{name: "on bw256", w: bw256, argSize: size, wantAsSelf: true},
		{name: "on w?size=0", w: w, argSize: 0, wantSize: defaultBufferSize},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := NewBufferedWriterSize(tc.w, tc.argSize)
			switch true {
			case tc.wantSize != 0:
				if n := b.Size(); n != tc.wantSize {
					t.Errorf("got size %d; want %d", n, tc.wantSize)
				}
			case tc.wantAsSelf:
				if b != tc.w {
					t.Error("buffered writer not input")
				}
			default:
				t.Errorf("invalid test case %+v", tc)
			}
		})
	}
}
