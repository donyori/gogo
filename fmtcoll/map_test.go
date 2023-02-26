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

func TestFormatMap(t *testing.T) {
	dataList := []map[string]int{
		nil,
		{},
		{"A": 1},
		{"A": 1, "B": 2},
		{"A": 1, "B": 2, "C": 3},
	}
	testCases := []struct {
		dataIdx                     int
		sep, keyFormat, valueFormat string
		prependType, prependLen     bool
		want                        string
	}{
		{0, "", "", "", false, false, "<nil>"},
		{0, "", "", "", false, true, "(0)<nil>"},
		{0, "", "", "", true, false, "(map[string]int)<nil>"},
		{0, "", "", "", true, true, "(map[string]int,0)<nil>"},

		{1, "", "", "", false, false, "{}"},
		{1, "", "", "", false, true, "(0){}"},
		{1, "", "", "", true, false, "(map[string]int){}"},
		{1, "", "", "", true, true, "(map[string]int,0){}"},

		{2, "", "", "", false, false, "{}"},
		{2, "", "", "", false, true, "(1){}"},
		{2, "", "", "", true, false, "(map[string]int){}"},
		{2, "", "", "", true, true, "(map[string]int,1){}"},
		{2, "", "", "%d", false, false, "{1}"},
		{2, "", "", "%d", false, true, "(1){1}"},
		{2, "", "", "%d", true, false, "(map[string]int){1}"},
		{2, "", "", "%d", true, true, "(map[string]int,1){1}"},
		{2, "", "%q", "", false, false, `{"A"}`},
		{2, "", "%q", "", false, true, `(1){"A"}`},
		{2, "", "%q", "", true, false, `(map[string]int){"A"}`},
		{2, "", "%q", "", true, true, `(map[string]int,1){"A"}`},
		{2, "", "%q", "%d", false, false, `{"A":1}`},
		{2, "", "%q", "%d", false, true, `(1){"A":1}`},
		{2, "", "%q", "%d", true, false, `(map[string]int){"A":1}`},
		{2, "", "%q", "%d", true, true, `(map[string]int,1){"A":1}`},

		{3, "", "", "", false, false, "{}"},
		{3, "", "", "", false, true, "(2){}"},
		{3, "", "", "", true, false, "(map[string]int){}"},
		{3, "", "", "", true, true, "(map[string]int,2){}"},
		{3, "", "", "%d", false, false, "{12}"},
		{3, "", "", "%d", false, true, "(2){12}"},
		{3, "", "", "%d", true, false, "(map[string]int){12}"},
		{3, "", "", "%d", true, true, "(map[string]int,2){12}"},
		{3, "", "%q", "", false, false, `{"A""B"}`},
		{3, "", "%q", "", false, true, `(2){"A""B"}`},
		{3, "", "%q", "", true, false, `(map[string]int){"A""B"}`},
		{3, "", "%q", "", true, true, `(map[string]int,2){"A""B"}`},
		{3, "", "%q", "%d", false, false, `{"A":1"B":2}`},
		{3, "", "%q", "%d", false, true, `(2){"A":1"B":2}`},
		{3, "", "%q", "%d", true, false, `(map[string]int){"A":1"B":2}`},
		{3, "", "%q", "%d", true, true, `(map[string]int,2){"A":1"B":2}`},
		{3, ",", "", "", false, false, "{,}"},
		{3, ",", "", "", false, true, "(2){,}"},
		{3, ",", "", "", true, false, "(map[string]int){,}"},
		{3, ",", "", "", true, true, "(map[string]int,2){,}"},
		{3, ",", "", "%d", false, false, "{1,2}"},
		{3, ",", "", "%d", false, true, "(2){1,2}"},
		{3, ",", "", "%d", true, false, "(map[string]int){1,2}"},
		{3, ",", "", "%d", true, true, "(map[string]int,2){1,2}"},
		{3, ",", "%q", "", false, false, `{"A","B"}`},
		{3, ",", "%q", "", false, true, `(2){"A","B"}`},
		{3, ",", "%q", "", true, false, `(map[string]int){"A","B"}`},
		{3, ",", "%q", "", true, true, `(map[string]int,2){"A","B"}`},
		{3, ",", "%q", "%d", false, false, `{"A":1,"B":2}`},
		{3, ",", "%q", "%d", false, true, `(2){"A":1,"B":2}`},
		{3, ",", "%q", "%d", true, false, `(map[string]int){"A":1,"B":2}`},
		{3, ",", "%q", "%d", true, true, `(map[string]int,2){"A":1,"B":2}`},

		{4, "", "", "", false, false, "{}"},
		{4, "", "", "", false, true, "(3){}"},
		{4, "", "", "", true, false, "(map[string]int){}"},
		{4, "", "", "", true, true, "(map[string]int,3){}"},
		{4, "", "", "%d", false, false, "{123}"},
		{4, "", "", "%d", false, true, "(3){123}"},
		{4, "", "", "%d", true, false, "(map[string]int){123}"},
		{4, "", "", "%d", true, true, "(map[string]int,3){123}"},
		{4, "", "%q", "", false, false, `{"A""B""C"}`},
		{4, "", "%q", "", false, true, `(3){"A""B""C"}`},
		{4, "", "%q", "", true, false, `(map[string]int){"A""B""C"}`},
		{4, "", "%q", "", true, true, `(map[string]int,3){"A""B""C"}`},
		{4, "", "%q", "%d", false, false, `{"A":1"B":2"C":3}`},
		{4, "", "%q", "%d", false, true, `(3){"A":1"B":2"C":3}`},
		{4, "", "%q", "%d", true, false, `(map[string]int){"A":1"B":2"C":3}`},
		{4, "", "%q", "%d", true, true, `(map[string]int,3){"A":1"B":2"C":3}`},
		{4, ",", "", "", false, false, "{,,}"},
		{4, ",", "", "", false, true, "(3){,,}"},
		{4, ",", "", "", true, false, "(map[string]int){,,}"},
		{4, ",", "", "", true, true, "(map[string]int,3){,,}"},
		{4, ",", "", "%d", false, false, "{1,2,3}"},
		{4, ",", "", "%d", false, true, "(3){1,2,3}"},
		{4, ",", "", "%d", true, false, "(map[string]int){1,2,3}"},
		{4, ",", "", "%d", true, true, "(map[string]int,3){1,2,3}"},
		{4, ",", "%q", "", false, false, `{"A","B","C"}`},
		{4, ",", "%q", "", false, true, `(3){"A","B","C"}`},
		{4, ",", "%q", "", true, false, `(map[string]int){"A","B","C"}`},
		{4, ",", "%q", "", true, true, `(map[string]int,3){"A","B","C"}`},
		{4, ",", "%q", "%d", false, false, `{"A":1,"B":2,"C":3}`},
		{4, ",", "%q", "%d", false, true, `(3){"A":1,"B":2,"C":3}`},
		{4, ",", "%q", "%d", true, false, `(map[string]int){"A":1,"B":2,"C":3}`},
		{4, ",", "%q", "%d", true, true, `(map[string]int,3){"A":1,"B":2,"C":3}`},
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("dataIdx=%d&sep=%+q&keyFormat=%+q&valueFormat=%+q&prependType=%t&prependLen=%t",
				tc.dataIdx, tc.sep, tc.keyFormat, tc.valueFormat, tc.prependType, tc.prependLen),
			func(t *testing.T) {
				got := fmtcoll.FormatMap(
					dataList[tc.dataIdx],
					tc.sep,
					tc.keyFormat,
					tc.valueFormat,
					tc.prependType,
					tc.prependLen,
					func(key1 string, value1 int, key2 string, value2 int) bool {
						if key1 != key2 {
							return key1 < key2
						}
						return value1 < value2
					},
				)
				if got != tc.want {
					t.Errorf("got %#q; want %#q", got, tc.want)
				}
			},
		)
	}
}
