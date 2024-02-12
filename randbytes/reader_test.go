// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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

package randbytes_test

import (
	"io"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/donyori/gogo/randbytes"
)

func TestNewReader_NilSrc(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			if !t.Failed() {
				t.Error("panic with nil")
			}
			return
		}
		s, ok := r.(string)
		if !ok || !strings.HasSuffix(s, "random value source is nil") {
			t.Error("unexpected panic:", r)
		}
	}()
	randbytes.NewReader(nil)
	t.Error("want panic but not")
}

func TestReader_Read(t *testing.T) {
	reader := randbytes.NewReader(rand.NewChaCha8(ChaCha8Seed))
	want := Random223BytesByChaCha8[:]
	p := make([]byte, len(want))
	n, err := reader.Read(p)
	if err != nil {
		t.Fatal(err)
	} else if n != len(want) {
		t.Errorf("got n %d; want %d", n, len(want))
	}
	if !slices.Equal(p, want) {
		t.Errorf("got (len %d)\n%x\nwant (len %d)\n%x",
			len(p), p, len(want), want)
	}
}

func TestReader_Read_ReadByOneByte(t *testing.T) {
	reader := randbytes.NewReader(rand.NewChaCha8(ChaCha8Seed))
	want := Random223BytesByChaCha8[:]
	p := make([]byte, len(want))
	_, err := io.ReadFull(iotest.OneByteReader(reader), p)
	if err != nil {
		t.Error(err)
	} else if !slices.Equal(p, want) {
		t.Errorf("got (len %d)\n%x\nwant (len %d)\n%x",
			len(p), p, len(want), want)
	}
}

func TestReader_Read_NilAndEmpty(t *testing.T) {
	testCases := []struct {
		name string
		p    []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := randbytes.NewReader(rand.NewChaCha8(ChaCha8Seed))
			n, err := reader.Read(tc.p)
			if err != nil {
				t.Error(err)
			} else if n != 0 {
				t.Errorf("got n %d; want 0", n)
			}
		})
	}
}

func TestReader_ReadByte(t *testing.T) {
	reader := randbytes.NewReader(rand.NewChaCha8(ChaCha8Seed))
	for i := range Random223BytesByChaCha8 {
		c, err := reader.ReadByte()
		if err != nil {
			t.Fatalf("Call %d, %v", i+1, err)
		} else if c != Random223BytesByChaCha8[i] {
			t.Errorf("Call %d, got %q; want %q",
				i+1, c, Random223BytesByChaCha8[i])
		}
	}
}
