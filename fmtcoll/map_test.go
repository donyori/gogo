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

package fmtcoll_test

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/donyori/gogo/container/mapping"
	"github.com/donyori/gogo/fmtcoll"
)

func TestFormatMapToString(t *testing.T) {
	testFormatMapToString(
		t,
		"map[string]*int",
		func(
			tc mapTestCase,
			dataList []map[string]*int,
			commonFormatList []fmtcoll.CommonFormat,
			keyValueLess func(key1 string, key2 string, _ *int, _ *int) bool,
		) (result string, err error) {
			return fmtcoll.FormatMapToString(
				dataList[tc.dataIdx],
				&fmtcoll.MapFormat[string, *int]{
					CommonFormat:  commonFormatList[tc.commonFormatIdx],
					FormatKeyFn:   tc.formatKeyFn,
					FormatValueFn: tc.formatValueFn,
					KeyValueLess:  keyValueLess,
				},
			)
		},
	)
}

func TestMustFormatMapToString_Panic(t *testing.T) {
	testMustFormatToStringPanic(
		t,
		func(errorFormatItemFn fmtcoll.FormatFunc[int]) {
			fmtcoll.MustFormatMapToString(
				map[int]int{0: 0},
				&fmtcoll.MapFormat[int, int]{FormatKeyFn: errorFormatItemFn},
			)
		},
	)
}

func TestFormatGogoMapToString(t *testing.T) {
	testFormatMapToString(
		t,
		"mapping.Map[string,*int]",
		func(
			tc mapTestCase,
			dataList []map[string]*int,
			commonFormatList []fmtcoll.CommonFormat,
			keyValueLess func(key1 string, key2 string, _ *int, _ *int) bool,
		) (result string, err error) {
			var m mapping.Map[string, *int]
			if dataList[tc.dataIdx] != nil {
				m = (*mapping.GoMap[string, *int])(&dataList[tc.dataIdx])
			}
			return fmtcoll.FormatGogoMapToString(
				m, &fmtcoll.MapFormat[string, *int]{
					CommonFormat:  commonFormatList[tc.commonFormatIdx],
					FormatKeyFn:   tc.formatKeyFn,
					FormatValueFn: tc.formatValueFn,
					KeyValueLess:  keyValueLess,
				},
			)
		},
	)
}

func TestMustFormatGogoMapToString_Panic(t *testing.T) {
	testMustFormatToStringPanic(
		t,
		func(errorFormatItemFn fmtcoll.FormatFunc[int]) {
			fmtcoll.MustFormatGogoMapToString[int, int](
				&mapping.GoMap[int, int]{0: 0},
				&fmtcoll.MapFormat[int, int]{FormatKeyFn: errorFormatItemFn},
			)
		},
	)
}

type mapTestCase struct {
	dataIdx         int
	commonFormatIdx int
	formatKeyFn     fmtcoll.FormatFunc[string]
	formatValueFn   fmtcoll.FormatFunc[*int]
	want            string
}

func testFormatMapToString(
	t *testing.T,
	typeStr string,
	f func(
		tc mapTestCase,
		dataList []map[string]*int,
		commonFormatList []fmtcoll.CommonFormat,
		keyValueLess func(key1, key2 string, _, _ *int) bool,
	) (result string, err error),
) {
	const Separator, Prefix, Suffix = ",", "<PREFIX>", "<SUFFIX>"

	keyValueLess := func(key1, key2 string, _, _ *int) bool {
		return key1 < key2
	}
	two, three := 2, 3
	dataList := []map[string]*int{
		nil,
		{},
		{"A": nil},
		{"A": nil, "B": &two},
		{"A": nil, "B": &two, "C": &three},
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

	for _, tc := range getTestCasesForFormatMapToString(
		typeStr, dataList, commonFormatList) {
		name := fmt.Sprintf("dataIdx=%d&commonFormatIdx=%d",
			tc.dataIdx, tc.commonFormatIdx)
		if tc.formatKeyFn == nil {
			name += "&formatKeyFn=<nil>"
		}
		if tc.formatValueFn == nil {
			name += "&formatValueFn=<nil>"
		}
		t.Run(name, func(t *testing.T) {
			got, err := f(tc, dataList, commonFormatList, keyValueLess)
			if err != nil {
				t.Error("err -", err)
			} else if got != tc.want {
				t.Errorf("got %#q; want %#q", got, tc.want)
			}
		})
	}
}

// getTestCasesForFormatMapToString returns test cases
// for testFormatMapToString.
func getTestCasesForFormatMapToString(
	typeStr string,
	dataList []map[string]*int,
	commonFormatList []fmtcoll.CommonFormat,
) []mapTestCase {
	const NilItemStr = "<nilptr>"
	formatKeyFn := fmtcoll.FprintfToFormatFunc[string]("%q")
	formatValueFn := func(w io.Writer, x *int) error {
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

	testCases := make([]mapTestCase, len(dataList)*len(commonFormatList)*2*2)
	var idx int
	for dataIdx, data := range dataList {
		for commonFormatIdx := range commonFormatList {
			var prefix string
			if commonFormatList[commonFormatIdx].PrependType {
				if commonFormatList[commonFormatIdx].PrependSize {
					prefix = fmt.Sprintf("(%s,%d)", typeStr, len(data))
				} else {
					prefix = fmt.Sprintf("(%s)", typeStr)
				}
			} else if commonFormatList[commonFormatIdx].PrependSize {
				prefix = fmt.Sprintf("(%d)", len(data))
			}

			for _, fmtKeyFn := range []fmtcoll.FormatFunc[string]{
				nil,
				formatKeyFn,
			} {
				for _, fmtValueFn := range []fmtcoll.FormatFunc[*int]{
					nil,
					formatValueFn,
				} {
					testCases[idx].dataIdx = dataIdx
					testCases[idx].commonFormatIdx = commonFormatIdx
					testCases[idx].formatKeyFn = fmtKeyFn
					testCases[idx].formatValueFn = fmtValueFn
					testCases[idx].want = getWantStringForFormatMapToString(
						NilItemStr,
						commonFormatList,
						commonFormatIdx,
						data,
						prefix,
						fmtKeyFn,
						fmtValueFn,
					)
					idx++
				}
			}
		}
	}
	return testCases
}

// getWantStringForFormatMapToString returns the expected result string
// for testFormatMapToString.
func getWantStringForFormatMapToString(
	nilItemStr string,
	commonFormatList []fmtcoll.CommonFormat,
	commonFormatIdx int,
	data map[string]*int,
	prefix string,
	fmtKeyFn fmtcoll.FormatFunc[string],
	fmtValueFn fmtcoll.FormatFunc[*int],
) string {
	var s string
	switch {
	case data == nil:
		s = "<nil>"
	case len(data) == 0:
		s = "{}"
	case fmtKeyFn == nil && fmtValueFn == nil:
		s = "{...}"
	default:
		keys := make([]string, 0, len(data))
		for key := range data {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		var b strings.Builder
		b.WriteByte('{')
		b.WriteString(commonFormatList[commonFormatIdx].Prefix)
		for i, key := range keys {
			if i > 0 {
				b.WriteString(commonFormatList[commonFormatIdx].Separator)
			}
			if fmtKeyFn != nil {
				b.WriteString(strconv.Quote(key))
				if fmtValueFn != nil {
					b.WriteByte(':')
				}
			}
			if fmtValueFn != nil {
				if x := data[key]; x != nil {
					b.WriteString(strconv.Itoa(*x))
				} else {
					b.WriteString(nilItemStr)
				}
			}
		}
		b.WriteString(commonFormatList[commonFormatIdx].Suffix)
		b.WriteByte('}')
		s = b.String()
	}
	return prefix + s
}
