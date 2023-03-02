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

package fmtcoll_test

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/donyori/gogo/fmtcoll"
)

func TestFormatSliceToString(t *testing.T) {
	const NilItemStr = "<nilptr>"
	const Separator, Prefix, Suffix = ",", "<PREFIX>", "<SUFFIX>"
	two, three := 2, 3
	dataList := [][]*int{nil, {}, {nil}, {nil, &two}, {nil, &two, &three}}
	formatItemFn := func(w io.Writer, x *int) error {
		var err error
		if x != nil {
			_, err = fmt.Fprintf(w, "%d", *x)
		} else if sw, ok := w.(io.StringWriter); ok {
			_, err = sw.WriteString(NilItemStr)
		} else {
			_, err = w.Write([]byte(NilItemStr))
		}
		return err
	}
	commonFormatList := []fmtcoll.CommonFormat{
		{},
		{PrependType: true},
		{PrependSize: true},
		{PrependType: true, PrependSize: true},
		{Separator: Separator},
		{Separator: Separator, PrependType: true},
		{Separator: Separator, PrependSize: true},
		{Separator: Separator, PrependType: true, PrependSize: true},
		{Separator: Separator, Prefix: Prefix, Suffix: Suffix},
		{Separator: Separator, Prefix: Prefix, Suffix: Suffix, PrependType: true},
		{Separator: Separator, Prefix: Prefix, Suffix: Suffix, PrependSize: true},
		{Separator: Separator, Prefix: Prefix, Suffix: Suffix, PrependType: true, PrependSize: true},
	}

	testCases := make([]struct {
		dataIdx         int
		commonFormatIdx int
		formatItemFn    fmtcoll.FormatFunc[*int]
		want            string
	}, len(dataList)*len(commonFormatList)*2)
	var idx int
	for dataIdx, data := range dataList {
		for commonFormatIdx := range commonFormatList {
			var prefix string
			if commonFormatList[commonFormatIdx].PrependType {
				if commonFormatList[commonFormatIdx].PrependSize {
					prefix = fmt.Sprintf("([]*int,%d)", len(data))
				} else {
					prefix = "([]*int)"
				}
			} else if commonFormatList[commonFormatIdx].PrependSize {
				prefix = fmt.Sprintf("(%d)", len(data))
			}

			for _, fmtItemFn := range []fmtcoll.FormatFunc[*int]{nil, formatItemFn} {
				var s string
				switch {
				case data == nil:
					s = "<nil>"
				case len(data) == 0:
					s = "[]"
				case fmtItemFn == nil:
					s = "[...]"
				default:
					var b strings.Builder
					b.WriteByte('[')
					b.WriteString(commonFormatList[commonFormatIdx].Prefix)
					for i, x := range data {
						if i > 0 {
							b.WriteString(commonFormatList[commonFormatIdx].Separator)
						}
						if x != nil {
							b.WriteString(strconv.Itoa(*x))
						} else {
							b.WriteString(NilItemStr)
						}
					}
					b.WriteString(commonFormatList[commonFormatIdx].Suffix)
					b.WriteByte(']')
					s = b.String()
				}

				testCases[idx].dataIdx = dataIdx
				testCases[idx].commonFormatIdx = commonFormatIdx
				testCases[idx].formatItemFn = fmtItemFn
				testCases[idx].want = prefix + s
				idx++
			}
		}
	}

	for _, tc := range testCases {
		name := fmt.Sprintf("dataIdx=%d&commonFormatIdx=%d", tc.dataIdx, tc.commonFormatIdx)
		if tc.formatItemFn == nil {
			name += "&formatItemFn=<nil>"
		}
		t.Run(name, func(t *testing.T) {
			got, err := fmtcoll.FormatSliceToString(
				dataList[tc.dataIdx],
				&fmtcoll.SequenceFormat[*int]{
					CommonFormat: commonFormatList[tc.commonFormatIdx],
					FormatItemFn: tc.formatItemFn,
				},
			)
			if err != nil {
				t.Fatal("err -", err)
			}
			if got != tc.want {
				t.Errorf("got %q; want %q", got, tc.want)
			}
		})
	}
}
