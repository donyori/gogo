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

package fmtcoll_test

import (
	"fmt"
	"io"

	"github.com/donyori/gogo/fmtcoll"
)

func ExampleFormatSliceToString() {
	data := [][]*[2]int{
		{{0, 0}, {0, 1}, {0, 2}},
		{{1, 0}, nil, {1, 2}},
		{{2, 0}, {2, 1}, {2, 2}},
	}
	itemFormat := &fmtcoll.SequenceFormat[*[2]int]{
		CommonFormat: fmtcoll.CommonFormat{
			Separator: ", ",
		},
		FormatItemFn: func(w io.Writer, x *[2]int) error {
			var err error
			if x != nil {
				_, err = fmt.Fprintf(w, "(%d, %d)", x[0], x[1])
			} else if sw, ok := w.(io.StringWriter); ok {
				_, err = sw.WriteString(" <nil>")
			} else {
				_, err = w.Write([]byte(" <nil>"))
			}
			return err
		},
	}
	rowFormat := &fmtcoll.SequenceFormat[[]*[2]int]{
		CommonFormat: fmtcoll.CommonFormat{
			Separator:   ",\n    ",
			Prefix:      "\n    ",
			Suffix:      ",\n",
			PrependType: true,
			PrependSize: true,
		},
		FormatItemFn: func(w io.Writer, x []*[2]int) error {
			s, err := fmtcoll.FormatSliceToString(x, itemFormat)
			if err != nil {
				return err
			}
			if sw, ok := w.(io.StringWriter); ok {
				_, err = sw.WriteString(s)
			} else {
				_, err = w.Write([]byte(s))
			}
			return err
		},
	}

	s, err := fmtcoll.FormatSliceToString(data, rowFormat)
	if err != nil {
		panic(err) // handle error
	}
	fmt.Println(s)

	// Output:
	// ([][]*[2]int,3)[
	//     [(0, 0), (0, 1), (0, 2)],
	//     [(1, 0),  <nil>, (1, 2)],
	//     [(2, 0), (2, 1), (2, 2)],
	// ]
}
