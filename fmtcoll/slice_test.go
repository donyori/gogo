// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2026  Yuan Gao
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
	"strings"
	"testing"

	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/fmtcoll"
)

func TestFormatSliceToString(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	testMustFormatToStringPanic(
		t,
		func(errorFormatItemFn fmtcoll.FormatFunc[int]) {
			fmtcoll.MustFormatSequenceToString(
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
	t.Helper()

	const Separator, Prefix, Suffix = ",", "<PREFIX>", "<SUFFIX>"

	two, three := 2, 3
	dataList := [][]*int{nil, {}, {nil}, {nil, &two}, {nil, &two, &three}}
	commonFormatList := []fmtcoll.CommonFormat{
		{},
		{PrependType: true},
		{PrependSize: true},
		{
			PrependType: true,
			PrependSize: true,
		},
		{Separator: Separator},
		{
			Separator:   Separator,
			PrependType: true,
		},
		{
			Separator:   Separator,
			PrependSize: true,
		},
		{
			Separator:   Separator,
			PrependType: true,
			PrependSize: true,
		},
		{
			Separator: Separator,
			Prefix:    Prefix,
			Suffix:    Suffix,
		},
		{
			Separator:   Separator,
			Prefix:      Prefix,
			Suffix:      Suffix,
			PrependType: true,
		},
		{
			Separator:   Separator,
			Prefix:      Prefix,
			Suffix:      Suffix,
			PrependSize: true,
		},
		{
			Separator:   Separator,
			Prefix:      Prefix,
			Suffix:      Suffix,
			PrependType: true,
			PrependSize: true,
		},
	}

	for _, tc := range getTestCasesForFormatSequenceToString(
		typeStr, dataList, commonFormatList) {
		name := fmt.Sprintf("dataIdx=%d&commonFormatIdx=%d",
			tc.dataIdx, tc.commonFormatIdx)
		if tc.formatItemFn == nil {
			name += "&formatItemFn=<nil>"
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

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
	const NilPointerStr = "<nilptr>"

	formatItemFn := func(w io.Writer, x *int) error {
		var err error
		if x != nil {
			_, err = fmt.Fprintf(w, "%d", *x)
		} else if sw, ok := w.(io.StringWriter); ok {
			_, err = sw.WriteString(NilPointerStr)
		} else {
			_, err = w.Write([]byte(NilPointerStr))
		}

		return err
	}

	testCases := make(
		[]sequenceTestCase,
		len(dataList)*len(commonFormatList)<<1,
	)

	var idx int

	for dataIdx, data := range dataList {
		for commonFormatIdx := range commonFormatList {
			prefix := getPrefixByCommonFormat(
				&commonFormatList[commonFormatIdx],
				typeStr,
				len(data),
			)

			for _, fmtItemFn := range []fmtcoll.FormatFunc[*int]{
				nil,
				formatItemFn,
			} {
				testCases[idx].dataIdx = dataIdx
				testCases[idx].commonFormatIdx = commonFormatIdx
				testCases[idx].formatItemFn = fmtItemFn
				testCases[idx].want = prefix + formatSequenceContentToString(
					data,
					&commonFormatList[commonFormatIdx],
					fmtItemFn,
				)
				idx++
			}
		}
	}

	return testCases
}

// formatSequenceContentToString formats the specified sequence content
// to a string with the specified format options.
func formatSequenceContentToString(
	data []*int,
	cf *fmtcoll.CommonFormat,
	formatItemFn fmtcoll.FormatFunc[*int],
) string {
	switch {
	case data == nil:
		return "<nil>"
	case len(data) == 0:
		return "[]"
	case formatItemFn == nil:
		return "[...]"
	}

	var b strings.Builder
	b.WriteByte('[')
	b.WriteString(cf.Prefix)

	for i, x := range data {
		if i > 0 {
			b.WriteString(cf.Separator)
		}

		writeItemToStringBuilder(&b, x, formatItemFn)
	}

	b.WriteString(cf.Suffix)
	b.WriteByte(']')

	return b.String()
}

func testMustFormatToStringPanic(
	t *testing.T,
	f func(errorFormatItemFn fmtcoll.FormatFunc[int]),
) {
	t.Helper()

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

// getPrefixByCommonFormat returns the prefix that conforms to
// the specified format options, type string, and collection size.
func getPrefixByCommonFormat(
	cf *fmtcoll.CommonFormat,
	typeStr string,
	size int,
) string {
	var prefix string

	if cf.PrependType {
		if cf.PrependSize {
			prefix = fmt.Sprintf("(%s,%d)", typeStr, size)
		} else {
			prefix = fmt.Sprintf("(%s)", typeStr)
		}
	} else if cf.PrependSize {
		prefix = fmt.Sprintf("(%d)", size)
	}

	return prefix
}

// writeItemToStringBuilder writes the item string returned by
// the specified format function to the specified string builder.
//
// If the format function returns an error,
// writeItemToStringBuilder writes an error string to
// the specified string builder.
func writeItemToStringBuilder[T any](
	b *strings.Builder,
	x T,
	formatItemFn fmtcoll.FormatFunc[T],
) {
	var xb strings.Builder

	err := formatItemFn(&xb, x)
	if err == nil {
		b.WriteString(xb.String())
	} else {
		b.WriteString("!ERROR:" + err.Error() + "!")
	}
}
