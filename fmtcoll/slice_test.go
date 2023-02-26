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
	"testing"

	"github.com/donyori/gogo/fmtcoll"
)

func TestFormatSlice(t *testing.T) {
	dataList := [][]int{nil, {}, {1}, {1, 2}, {1, 2, 3}}
	testCases := []struct {
		dataIdx                 int
		sep, format             string
		prependType, prependLen bool
		want                    string
	}{
		{0, "", "", false, false, "<nil>"},
		{0, "", "", false, true, "(0)<nil>"},
		{0, "", "", true, false, "([]int)<nil>"},
		{0, "", "", true, true, "([]int,0)<nil>"},

		{1, "", "", false, false, "[]"},
		{1, "", "", false, true, "(0)[]"},
		{1, "", "", true, false, "([]int)[]"},
		{1, "", "", true, true, "([]int,0)[]"},

		{2, "", "", false, false, "[]"},
		{2, "", "", false, true, "(1)[]"},
		{2, "", "", true, false, "([]int)[]"},
		{2, "", "", true, true, "([]int,1)[]"},
		{2, "", "%d", false, false, "[1]"},
		{2, "", "%d", false, true, "(1)[1]"},
		{2, "", "%d", true, false, "([]int)[1]"},
		{2, "", "%d", true, true, "([]int,1)[1]"},

		{3, "", "", false, false, "[]"},
		{3, "", "", false, true, "(2)[]"},
		{3, "", "", true, false, "([]int)[]"},
		{3, "", "", true, true, "([]int,2)[]"},
		{3, "", "%d", false, false, "[12]"},
		{3, "", "%d", false, true, "(2)[12]"},
		{3, "", "%d", true, false, "([]int)[12]"},
		{3, "", "%d", true, true, "([]int,2)[12]"},
		{3, ",", "", false, false, "[,]"},
		{3, ",", "", false, true, "(2)[,]"},
		{3, ",", "", true, false, "([]int)[,]"},
		{3, ",", "", true, true, "([]int,2)[,]"},
		{3, ",", "%d", false, false, "[1,2]"},
		{3, ",", "%d", false, true, "(2)[1,2]"},
		{3, ",", "%d", true, false, "([]int)[1,2]"},
		{3, ",", "%d", true, true, "([]int,2)[1,2]"},

		{4, "", "", false, false, "[]"},
		{4, "", "", false, true, "(3)[]"},
		{4, "", "", true, false, "([]int)[]"},
		{4, "", "", true, true, "([]int,3)[]"},
		{4, "", "%d", false, false, "[123]"},
		{4, "", "%d", false, true, "(3)[123]"},
		{4, "", "%d", true, false, "([]int)[123]"},
		{4, "", "%d", true, true, "([]int,3)[123]"},
		{4, ",", "", false, false, "[,,]"},
		{4, ",", "", false, true, "(3)[,,]"},
		{4, ",", "", true, false, "([]int)[,,]"},
		{4, ",", "", true, true, "([]int,3)[,,]"},
		{4, ",", "%d", false, false, "[1,2,3]"},
		{4, ",", "%d", false, true, "(3)[1,2,3]"},
		{4, ",", "%d", true, false, "([]int)[1,2,3]"},
		{4, ",", "%d", true, true, "([]int,3)[1,2,3]"},
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("dataIdx=%d&sep=%+q&format=%+q&prependType=%t&prependLen=%t",
				tc.dataIdx, tc.sep, tc.format, tc.prependType, tc.prependLen),
			func(t *testing.T) {
				got := fmtcoll.FormatSlice(
					dataList[tc.dataIdx],
					tc.sep,
					tc.format,
					tc.prependType,
					tc.prependLen,
				)
				if got != tc.want {
					t.Errorf("got %q; want %q", got, tc.want)
				}
			},
		)
	}
}
