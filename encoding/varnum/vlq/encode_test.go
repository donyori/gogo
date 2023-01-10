// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2023  Yuan Gao
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

package vlq_test

import (
	"bytes"
	"fmt"
	"math"
	"testing"

	"github.com/donyori/gogo/encoding/varnum/uintconv"
	"github.com/donyori/gogo/encoding/varnum/vlq"
)

func TestUint64EncodedLen(t *testing.T) {
	for i, u := range uint64s {
		t.Run(fmt.Sprintf("u=%#X", u), func(t *testing.T) {
			if n := vlq.Uint64EncodedLen(u); n != len(encodedUint64s[i]) {
				t.Errorf("got %d; want %d", n, len(encodedUint64s[i]))
			}
		})
	}
}

func BenchmarkUint64EncodedLen(b *testing.B) {
	benchmarks := []struct {
		name string
		f    func(u uint64) int
	}{
		{"package func", vlq.Uint64EncodedLen},
		{"binary func", func(u uint64) int {
			// Binary search.
			// Define: minUint64s[-1] < u,
			//         minUint64s[len(minUint64s)] > u
			// Invariant: minUint64s[low-1] < u,
			//            minUint64s[high] > u
			low, high := 0, len(vlq.MinUint64s)
			for low < high {
				mid := (low + high) / 2
				if vlq.MinUint64s[mid] < u {
					low = mid + 1 // Preserve: minUint64s[low-1] < u
				} else if vlq.MinUint64s[mid] > u {
					high = mid // Preserve: minUint64s[high] > u
				} else {
					return mid + 2
				}
			}
			return high + 1
		}},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				for _, u := range uint64s {
					bm.f(u)
				}
			}
		})
	}
}

func TestEncodeUint64(t *testing.T) {
	for i, u := range uint64s {
		t.Run(fmt.Sprintf("u=%#X", u), func(t *testing.T) {
			dst := make([]byte, 10)
			n := vlq.EncodeUint64(dst, u)
			if !bytes.Equal(dst[:n], encodedUint64s[i]) {
				t.Errorf("got %#X; want %#X", dst[:n], encodedUint64s[i])
			}
		})
	}
}

func TestInt64EncodedLen(t *testing.T) {
	for i, u := range uint64s {
		x := uintconv.ToInt64Zigzag(u)
		t.Run(fmt.Sprintf("i=%#X", x), func(t *testing.T) {
			if n := vlq.Int64EncodedLen(x); n != len(encodedUint64s[i]) {
				t.Errorf("got %d; want %d", n, len(encodedUint64s[i]))
			}
		})
	}
}

func TestEncodeInt64(t *testing.T) {
	for i, u := range uint64s {
		x := uintconv.ToInt64Zigzag(u)
		t.Run(fmt.Sprintf("i=%#X", x), func(t *testing.T) {
			dst := make([]byte, 10)
			n := vlq.EncodeInt64(dst, x)
			if !bytes.Equal(dst[:n], encodedUint64s[i]) {
				t.Errorf("got %#X; want %#X", dst[:n], encodedUint64s[i])
			}
		})
	}
}

func TestFloat64EncodedLen(t *testing.T) {
	for i, u := range uint64s {
		f := uintconv.ToFloat64ByteReversal(u)
		t.Run(fmt.Sprintf("f=%v(bits=%#016X)", f, math.Float64bits(f)), func(t *testing.T) {
			if n := vlq.Float64EncodedLen(f); n != len(encodedUint64s[i]) {
				t.Errorf("got %d; want %d", n, len(encodedUint64s[i]))
			}
		})
	}
}

func TestEncodeFloat64(t *testing.T) {
	for i, u := range uint64s {
		f := uintconv.ToFloat64ByteReversal(u)
		t.Run(fmt.Sprintf("f=%v(bits=%#016X)", f, math.Float64bits(f)), func(t *testing.T) {
			dst := make([]byte, 10)
			n := vlq.EncodeFloat64(dst, f)
			if !bytes.Equal(dst[:n], encodedUint64s[i]) {
				t.Errorf("got %#X; want %#X", dst[:n], encodedUint64s[i])
			}
		})
	}
}
