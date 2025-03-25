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

package hex_test

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/donyori/gogo/encoding/hex"
	"github.com/donyori/gogo/errors"
)

// encodeInt64Baseline implements all requirements of
// function EncodeInt64 with package fmt.
func encodeInt64Baseline(dst []byte, x int64, upper bool, digits int) int {
	layout := "%"
	if digits > 0 {
		width := digits
		if x < 0 {
			width++ // increase one for the negative sign
		}
		layout += "0" + strconv.Itoa(width)
	}
	if upper {
		layout += "X"
	} else {
		layout += "x"
	}
	w := bytes.NewBuffer(dst[:0])
	n, err := fmt.Fprintf(w, layout, x)
	if err != nil {
		panic(errors.AutoWrap(err))
	} else if n > len(dst) {
		panic(errors.AutoMsg(fmt.Sprintf(
			"dst is too small, len(dst): %d, need: %d", len(dst), n)))
	}
	return n
}

// encodeInt64ToStringBaseline implements all requirements of
// function EncodeInt64ToString with package fmt.
func encodeInt64ToStringBaseline(x int64, upper bool, digits int) string {
	layout := "%"
	if digits > 0 {
		width := digits
		if x < 0 {
			width++ // increase one for the negative sign
		}
		layout += "0" + strconv.Itoa(width)
	}
	if upper {
		layout += "X"
	} else {
		layout += "x"
	}
	return fmt.Sprintf(layout, x)
}

// encodeInt64ToBaseline implements all requirements of
// function EncodeInt64To with package fmt.
func encodeInt64ToBaseline(w io.Writer, x int64, upper bool, digits int) (
	written int, err error) {
	layout := "%"
	if digits > 0 {
		width := digits
		if x < 0 {
			width++ // increase one for the negative sign
		}
		layout += "0" + strconv.Itoa(width)
	}
	if upper {
		layout += "X"
	} else {
		layout += "x"
	}
	return fmt.Fprintf(w, layout, x)
}

var testEncodeInt64Xs = [...]int64{
	0, 1, 2, 3, 7, 8, 9, 15, 16, 31, 32, 33, 63, 64, 65,
	1234567890, math.MaxInt64 - 1, math.MaxInt64,
	-1, -2, -3, -7, -8, -9, -15, -16, -31, -32, -33, -63, -64, -65,
	-1234567890, math.MinInt64 + 1, math.MinInt64,
}

var testEncodeIntegerDigits = [...]int{
	-10, -1, 0, 1, 2, 3, 4, 8, 9, 14, 15, 16, 17, 18, 500,
}

func TestEncodeInt64(t *testing.T) {
	for _, x := range testEncodeInt64Xs {
		for _, upper := range []bool{false, true} {
			for _, digits := range testEncodeIntegerDigits {
				t.Run(
					fmt.Sprintf("x=%d&upper=%t&digits=%d", x, upper, digits),
					func(t *testing.T) {
						length := hex.EncodeInt64DstLen(digits)
						dst1, dst2 := make([]byte, length), make([]byte, length)
						n1 := hex.EncodeInt64(dst1, x, upper, digits)
						n2 := encodeInt64Baseline(dst2, x, upper, digits)
						if n1 != n2 || !bytes.Equal(dst1[:n1], dst2[:n2]) {
							t.Errorf("got (%d) %s; want (%d) %s",
								n1, dst1[:n1], n2, dst2[:n2])
						}
					},
				)
			}
		}
	}
}

func TestEncodeInt64ToString(t *testing.T) {
	for _, x := range testEncodeInt64Xs {
		for _, upper := range []bool{false, true} {
			for _, digits := range testEncodeIntegerDigits {
				t.Run(
					fmt.Sprintf("x=%d&upper=%t&digits=%d", x, upper, digits),
					func(t *testing.T) {
						r := hex.EncodeInt64ToString(x, upper, digits)
						want := encodeInt64ToStringBaseline(x, upper, digits)
						if r != want {
							t.Errorf("got %s; want %s", r, want)
						}
					},
				)
			}
		}
	}
}

func TestEncodeInt64To(t *testing.T) {
	var b1, b2 strings.Builder
	for _, x := range testEncodeInt64Xs {
		for _, upper := range []bool{false, true} {
			for _, digits := range testEncodeIntegerDigits {
				t.Run(
					fmt.Sprintf("x=%d&upper=%t&digits=%d", x, upper, digits),
					func(t *testing.T) {
						b1.Reset()
						b2.Reset()
						n1, err1 := hex.EncodeInt64To(&b1, x, upper, digits)
						if err1 != nil {
							t.Fatalf("written %d - %v", n1, err1)
						}
						n2, err2 := encodeInt64ToBaseline(&b2, x, upper, digits)
						if err2 != nil {
							t.Fatalf("baseline, written %d - %v", n2, err2)
						}
						if n1 != n2 || b1.String() != b2.String() {
							t.Errorf("got (%d) %s; want (%d) %s",
								n1, b1.String(), n2, b2.String())
						}
					},
				)
			}
		}
	}
}
