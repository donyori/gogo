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

	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/fmtcoll"
)

func TestFormatSliceToString(t *testing.T) {
	testFormatSequenceToString(
		t,
		"[]*int",
		func(
			tc sequenceTestCase,
			dataList [][]*int,
			commonFormatList []fmtcoll.CommonFormat,
		) (result string, err error) {
			return fmtcoll.FormatSliceToString(
				dataList[tc.dataIdx],
				&fmtcoll.SequenceFormat[*int]{
					CommonFormat: commonFormatList[tc.commonFormatIdx],
					FormatItemFn: tc.formatItemFn,
				},
			)
		},
	)
}

func TestMustFormatSliceToString_Panic(t *testing.T) {
	testMustFormatToStringPanic(
		t,
		func(errorFormatItemFn fmtcoll.FormatFunc[int]) {
			fmtcoll.MustFormatSliceToString(
				[]int{0},
				&fmtcoll.SequenceFormat[int]{FormatItemFn: errorFormatItemFn},
			)
		},
	)
}

func TestFormatSequenceToString(t *testing.T) {
	testFormatSequenceToString(
		t,
		"sequence.Sequence[*int]",
		func(
			tc sequenceTestCase,
			dataList [][]*int,
			commonFormatList []fmtcoll.CommonFormat,
		) (result string, err error) {
			var s sequence.Sequence[*int]
			if dataList[tc.dataIdx] != nil {
				s = (*array.SliceDynamicArray[*int])(&dataList[tc.dataIdx])
			}
			return fmtcoll.FormatSequenceToString(
				s, &fmtcoll.SequenceFormat[*int]{
					CommonFormat: commonFormatList[tc.commonFormatIdx],
					FormatItemFn: tc.formatItemFn,
				},
			)
		},
	)
}

func TestMustFormatSequenceToString_Panic(t *testing.T) {
	testMustFormatToStringPanic(
		t,
		func(errorFormatItemFn fmtcoll.FormatFunc[int]) {
			fmtcoll.MustFormatSequenceToString[int](
				&array.SliceDynamicArray[int]{0},
				&fmtcoll.SequenceFormat[int]{FormatItemFn: errorFormatItemFn},
			)
		},
	)
}

type sequenceTestCase struct {
	dataIdx         int
	commonFormatIdx int
	formatItemFn    fmtcoll.FormatFunc[*int]
	want            string
}

func testFormatSequenceToString(
	t *testing.T,
	typeStr string,
	f func(
		tc sequenceTestCase,
		dataList [][]*int,
		commonFormatList []fmtcoll.CommonFormat,
	) (result string, err error),
) {
	const Separator, Prefix, Suffix = ",", "<PREFIX>", "<SUFFIX>"

	two, three := 2, 3
	dataList := [][]*int{nil, {}, {nil}, {nil, &two}, {nil, &two, &three}}
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

	for _, tc := range getTestCasesForFormatSequenceToString(
		typeStr, dataList, commonFormatList) {
		name := fmt.Sprintf("dataIdx=%d&commonFormatIdx=%d",
			tc.dataIdx, tc.commonFormatIdx)
		if tc.formatItemFn == nil {
			name += "&formatItemFn=<nil>"
		}
		t.Run(name, func(t *testing.T) {
			got, err := f(tc, dataList, commonFormatList)
			if err != nil {
				t.Error("err -", err)
			} else if got != tc.want {
				t.Errorf("got %q; want %q", got, tc.want)
			}
		})
	}
}

// getTestCasesForFormatSequenceToString returns test cases
// for testFormatSequenceToString.
func getTestCasesForFormatSequenceToString(
	typeStr string,
	dataList [][]*int,
	commonFormatList []fmtcoll.CommonFormat,
) []sequenceTestCase {
	const NilItemStr = "<nilptr>"
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

	testCases := make([]sequenceTestCase, len(dataList)*len(commonFormatList)*2)
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

			for _, fmtItemFn := range []fmtcoll.FormatFunc[*int]{
				nil,
				formatItemFn,
			} {
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
							b.WriteString(
								commonFormatList[commonFormatIdx].Separator)
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
	return testCases
}

func testMustFormatToStringPanic(
	t *testing.T,
	f func(errorFormatItemFn fmtcoll.FormatFunc[int]),
) {
	wantErr := errors.New("want error")
	errorFormatItemFn := func(io.Writer, int) error {
		return errors.AutoWrap(wantErr)
	}
	defer func() {
		e := recover()
		if e == nil {
			t.Error("want panic but not")
			return
		}
		err, ok := e.(error)
		if !ok || !errors.Is(err, wantErr) {
			t.Error("panic -", e)
		}
	}()
	f(errorFormatItemFn)
}
