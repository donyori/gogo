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

package randbytes_test

import (
	"bytes"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/donyori/gogo/function/compare"
	"github.com/donyori/gogo/randbytes"
)

// ChaCha8Seed is the seed for ChaCha8 used for testing.
var ChaCha8Seed = [32]byte([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"))

// Random223BytesByChaCha8 is a byte array of length 223
// generated by ChaCha8 seeded with ChaCha8Seed.
var Random223BytesByChaCha8 [223]byte

func init() {
	src := rand.NewChaCha8(ChaCha8Seed)
	var x uint64
	for i := range Random223BytesByChaCha8 {
		if i%8 == 0 {
			x = src.Uint64()
		}
		Random223BytesByChaCha8[i] = byte(x)
		x >>= 8
	}
}

func TestFill(t *testing.T) {
	want := Random223BytesByChaCha8[:]
	p := make([]byte, len(want))
	randbytes.Fill(rand.NewChaCha8(ChaCha8Seed), p)
	if !bytes.Equal(p, want) {
		t.Errorf("got (len %d)\n%x\nwant (len %d)\n%x",
			len(p), p, len(want), want)
	}
}

func TestFill_NilAndEmpty(t *testing.T) {
	testCases := []struct {
		name string
		p    []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			randbytes.Fill(rand.NewChaCha8(ChaCha8Seed), tc.p)
		})
	}
}

func TestMake(t *testing.T) {
	want := Random223BytesByChaCha8[:]
	got := randbytes.Make(rand.NewChaCha8(ChaCha8Seed), len(want))
	if cap(got) != len(want) || !bytes.Equal(got, want) {
		t.Errorf("got (len %d, cap %d)\n%x\nwant (len %d, cap %[4]d)\n%x",
			len(got), cap(got), got, len(want), want)
	}
}

func TestMake_Empty(t *testing.T) {
	got := randbytes.Make(rand.NewChaCha8(ChaCha8Seed), 0)
	if got == nil {
		t.Error("got <nil>; want []byte{}")
	} else if len(got) != 0 || cap(got) != 0 {
		t.Errorf("got len %d, cap %d; want 0, 0", len(got), cap(got))
	}
}

func TestMakeCapacity(t *testing.T) {
	want := Random223BytesByChaCha8[:]
	wantCap := len(want) + 20
	got := randbytes.MakeCapacity(
		rand.NewChaCha8(ChaCha8Seed), len(want), wantCap)
	if cap(got) != wantCap || !bytes.Equal(got, want) {
		t.Errorf("got (len %d, cap %d)\n%x\nwant (len %d, cap %d)\n%x",
			len(got), cap(got), got, len(want), wantCap, want)
	}
}

func TestMakeCapacity_Empty(t *testing.T) {
	got := randbytes.MakeCapacity(rand.NewChaCha8(ChaCha8Seed), 0, 0)
	if got == nil {
		t.Error("got <nil>; want []byte{}")
	} else if len(got) != 0 || cap(got) != 0 {
		t.Errorf("got len %d, cap %d; want 0, 0", len(got), cap(got))
	}
}

func TestMakeCapacity_EmptyButCapacityOne(t *testing.T) {
	got := randbytes.MakeCapacity(rand.NewChaCha8(ChaCha8Seed), 0, 1)
	if got == nil {
		t.Error("got <nil>; want []byte{}")
	} else if len(got) != 0 || cap(got) != 1 {
		t.Errorf("got len %d, cap %d; want 0, 1", len(got), cap(got))
	}
}

func TestAppend(t *testing.T) {
	p := []byte("Append")
	want := make([]byte, len(p)+len(Random223BytesByChaCha8))
	copy(want[copy(want, p):], Random223BytesByChaCha8[:])
	got := randbytes.Append(
		rand.NewChaCha8(ChaCha8Seed), p, len(Random223BytesByChaCha8))
	if !bytes.Equal(got, want) {
		t.Errorf("got (len %d)\n%x\nwant (len %d)\n%x",
			len(got), got, len(want), want)
	}
}

func TestAppend_NZero(t *testing.T) {
	testCases := []struct {
		name string
		p    []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
		{"nonempty", []byte("Append")},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			want := slices.Clone(tc.p)
			got := randbytes.Append(rand.NewChaCha8(ChaCha8Seed), tc.p, 0)
			if !compare.SliceEqual(got, want) {
				switch {
				case got == nil:
					// want is non-nil.
					t.Errorf("got <nil>; want %x", want)
				case want == nil:
					// got is non-nil.
					t.Errorf("got %x; want <nil>", got)
				default:
					// Both got and want are non-nil.
					t.Errorf("got %x; want %x", got, want)
				}
			}
		})
	}
}

func TestAppend_ToNilAndEmpty(t *testing.T) {
	want := Random223BytesByChaCha8[:]
	testCases := []struct {
		name string
		p    []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := randbytes.Append(
				rand.NewChaCha8(ChaCha8Seed), tc.p, len(want))
			if !bytes.Equal(got, want) {
				t.Errorf("got (len %d)\n%x\nwant (len %d)\n%x",
					len(got), got, len(want), want)
			}
		})
	}
}

func TestWriteN(t *testing.T) {
	want := Random223BytesByChaCha8[:]
	var buf bytes.Buffer
	buf.Grow(len(want))
	written, err := randbytes.WriteN(
		rand.NewChaCha8(ChaCha8Seed), &buf, len(want))
	if err != nil {
		t.Fatal(err)
	} else if written != len(want) {
		t.Errorf("got written %d; want %d", written, len(want))
	}
	got := buf.Bytes()
	if !bytes.Equal(got, want) {
		t.Errorf("got (len %d)\n%x\nwant (len %d)\n%x",
			len(got), got, len(want), want)
	}
}

func TestWriteN_NZero(t *testing.T) {
	var buf bytes.Buffer
	written, err := randbytes.WriteN(rand.NewChaCha8(ChaCha8Seed), &buf, 0)
	if err != nil {
		t.Fatal(err)
	} else if written != 0 {
		t.Errorf("got written %d; want 0", written)
	}
	if buf.Len() != 0 || buf.Cap() != 0 {
		t.Errorf("got buf.Len() %d, buf.Cap() %d; want 0, 0",
			buf.Len(), buf.Cap())
		if buf.Len() > 0 {
			t.Errorf("got data %x", buf.Bytes())
		}
	}
}
