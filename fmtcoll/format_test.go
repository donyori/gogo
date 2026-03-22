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
	"math"
	"strings"
	"testing"

	"github.com/donyori/gogo/fmtcoll"
	"github.com/donyori/gogo/internal/floats"
)

func TestNewDefaultSequenceFormat(t *testing.T) {
	t.Parallel()

	xs := []any{
		0,
		-1,
		int64(math.MaxInt64),
		0.,
		-.1,
		floats.Inf64,
		floats.NaN32A,
		floats.SmallestNonzeroFloat64,
		rune(0),
		'A',
		'\'',
		'"',
		'`',
		'\u6C49',
		"",
		"A",
		"'single-quoted'",
		`"Double-quoted"`,
		"`Back-quoted`",
		"\u6C49\u5B57",
	}
	wantCF := fmtcoll.CommonFormat{Separator: ","}
	got := fmtcoll.NewDefaultSequenceFormat[any]()

	if got.CommonFormat != wantCF {
		t.Errorf("got common format %#v; want %#v", got.CommonFormat, wantCF)
	}

	checkFormatItemFn(t, "item", "%v", xs, got.FormatItemFn)
}

func TestNewSequenceFormatPrepend(t *testing.T) {
	t.Parallel()

	xs := []any{
		0,
		-1,
		int64(math.MaxInt64),
		0.,
		-.1,
		floats.Inf64,
		floats.NaN32A,
		floats.SmallestNonzeroFloat64,
		rune(0),
		'A',
		'\'',
		'"',
		'`',
		'\u6C49',
		"",
		"A",
		"'single-quoted'",
		`"Double-quoted"`,
		"`Back-quoted`",
		"\u6C49\u5B57",
	}
	wantCF := fmtcoll.CommonFormat{
		Separator:   ",",
		PrependType: true,
		PrependSize: true,
	}
	got := fmtcoll.NewSequenceFormatPrepend[any]()

	if got.CommonFormat != wantCF {
		t.Errorf("got common format %#v; want %#v", got.CommonFormat, wantCF)
	}

	checkFormatItemFn(t, "item", "%v", xs, got.FormatItemFn)
}

func TestNewSequenceFormatQuoted(t *testing.T) {
	t.Parallel()

	xs := []any{
		rune(0),
		'A',
		'\'',
		'"',
		'`',
		'\u6C49',
		"",
		"A",
		"'single-quoted'",
		`"Double-quoted"`,
		"`Back-quoted`",
		"\u6C49\u5B57",
	}
	wantCF := fmtcoll.CommonFormat{Separator: ","}
	testCases := []struct {
		asciiOnly bool
		format    string
	}{
		{false, "%q"},
		{true, "%+q"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("asciiOnly=%t", tc.asciiOnly), func(t *testing.T) {
			t.Parallel()

			got := fmtcoll.NewSequenceFormatQuoted[any](tc.asciiOnly)
			if got.CommonFormat != wantCF {
				t.Errorf("got common format %#v; want %#v",
					got.CommonFormat, wantCF)
			}

			checkFormatItemFn(t, "item", tc.format, xs, got.FormatItemFn)
		})
	}
}

func TestNewDefaultMapFormat(t *testing.T) {
	t.Parallel()

	xs := []any{
		0,
		-1,
		int64(math.MaxInt64),
		0.,
		-.1,
		floats.Inf64,
		floats.NaN32A,
		floats.SmallestNonzeroFloat64,
		rune(0),
		'A',
		'\'',
		'"',
		'`',
		'\u6C49',
		"",
		"A",
		"'single-quoted'",
		`"Double-quoted"`,
		"`Back-quoted`",
		"\u6C49\u5B57",
	}
	wantCF := fmtcoll.CommonFormat{Separator: ","}
	got := fmtcoll.NewDefaultMapFormat[any, any]()

	if got.CommonFormat != wantCF {
		t.Errorf("got common format %#v; want %#v", got.CommonFormat, wantCF)
	}

	if got.CompareKeyValueFn != nil {
		t.Error("got non-nil CompareKeyValueFn")
	}

	checkFormatItemFn(t, "key", "%v", xs, got.FormatKeyFn)
	checkFormatItemFn(t, "value", "%v", xs, got.FormatValueFn)
}

func TestNewMapFormatPrepend(t *testing.T) {
	t.Parallel()

	xs := []any{
		0,
		-1,
		int64(math.MaxInt64),
		0.,
		-.1,
		floats.Inf64,
		floats.NaN32A,
		floats.SmallestNonzeroFloat64,
		rune(0),
		'A',
		'\'',
		'"',
		'`',
		'\u6C49',
		"",
		"A",
		"'single-quoted'",
		`"Double-quoted"`,
		"`Back-quoted`",
		"\u6C49\u5B57",
	}
	wantCF := fmtcoll.CommonFormat{
		Separator:   ",",
		PrependType: true,
		PrependSize: true,
	}
	got := fmtcoll.NewMapFormatPrepend[any, any]()

	if got.CommonFormat != wantCF {
		t.Errorf("got common format %#v; want %#v", got.CommonFormat, wantCF)
	}

	if got.CompareKeyValueFn != nil {
		t.Error("got non-nil CompareKeyValueFn")
	}

	checkFormatItemFn(t, "key", "%v", xs, got.FormatKeyFn)
	checkFormatItemFn(t, "value", "%v", xs, got.FormatValueFn)
}

func TestNewMapFormatKeyQuoted(t *testing.T) {
	t.Parallel()

	keys := []any{
		rune(0),
		'A',
		'\'',
		'"',
		'`',
		'\u6C49',
		"",
		"A",
		"'single-quoted'",
		`"Double-quoted"`,
		"`Back-quoted`",
		"\u6C49\u5B57",
	}
	values := []any{
		0,
		-1,
		int64(math.MaxInt64),
		0.,
		-.1,
		floats.Inf64,
		floats.NaN32A,
		floats.SmallestNonzeroFloat64,
		rune(0),
		'A',
		'\'',
		'"',
		'`',
		'\u6C49',
		"",
		"A",
		"'single-quoted'",
		`"Double-quoted"`,
		"`Back-quoted`",
		"\u6C49\u5B57",
	}
	wantCF := fmtcoll.CommonFormat{Separator: ","}
	testCases := []struct {
		asciiOnly bool
		format    string
	}{
		{false, "%q"},
		{true, "%+q"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("asciiOnly=%t", tc.asciiOnly), func(t *testing.T) {
			t.Parallel()

			got := fmtcoll.NewMapFormatKeyQuoted[any, any](tc.asciiOnly)

			if got.CommonFormat != wantCF {
				t.Errorf("got common format %#v; want %#v",
					got.CommonFormat, wantCF)
			}

			if got.CompareKeyValueFn != nil {
				t.Error("got non-nil CompareKeyValueFn")
			}

			checkFormatItemFn(t, "key", tc.format, keys, got.FormatKeyFn)
			checkFormatItemFn(t, "value", "%v", values, got.FormatValueFn)
		})
	}
}

func TestNewMapFormatValueQuoted(t *testing.T) {
	t.Parallel()

	keys := []any{
		0,
		-1,
		int64(math.MaxInt64),
		0.,
		-.1,
		floats.Inf64,
		floats.NaN32A,
		floats.SmallestNonzeroFloat64,
		rune(0),
		'A',
		'\'',
		'"',
		'`',
		'\u6C49',
		"",
		"A",
		"'single-quoted'",
		`"Double-quoted"`,
		"`Back-quoted`",
		"\u6C49\u5B57",
	}
	values := []any{
		rune(0),
		'A',
		'\'',
		'"',
		'`',
		'\u6C49',
		"",
		"A",
		"'single-quoted'",
		`"Double-quoted"`,
		"`Back-quoted`",
		"\u6C49\u5B57",
	}
	wantCF := fmtcoll.CommonFormat{Separator: ","}
	testCases := []struct {
		asciiOnly bool
		format    string
	}{
		{false, "%q"},
		{true, "%+q"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("asciiOnly=%t", tc.asciiOnly), func(t *testing.T) {
			t.Parallel()

			got := fmtcoll.NewMapFormatValueQuoted[any, any](tc.asciiOnly)

			if got.CommonFormat != wantCF {
				t.Errorf("got common format %#v; want %#v",
					got.CommonFormat, wantCF)
			}

			if got.CompareKeyValueFn != nil {
				t.Error("got non-nil CompareKeyValueFn")
			}

			checkFormatItemFn(t, "key", "%v", keys, got.FormatKeyFn)
			checkFormatItemFn(t, "value", tc.format, values, got.FormatValueFn)
		})
	}
}

func TestNewMapFormatKeyValueQuoted(t *testing.T) {
	t.Parallel()

	xs := []any{
		rune(0),
		'A',
		'\'',
		'"',
		'`',
		'\u6C49',
		"",
		"A",
		"'single-quoted'",
		`"Double-quoted"`,
		"`Back-quoted`",
		"\u6C49\u5B57",
	}
	wantCF := fmtcoll.CommonFormat{Separator: ","}
	testCases := []struct {
		asciiOnly bool
		format    string
	}{
		{false, "%q"},
		{true, "%+q"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("asciiOnly=%t", tc.asciiOnly), func(t *testing.T) {
			t.Parallel()

			got := fmtcoll.NewMapFormatKeyValueQuoted[any, any](tc.asciiOnly)

			if got.CommonFormat != wantCF {
				t.Errorf("got common format %#v; want %#v",
					got.CommonFormat, wantCF)
			}

			if got.CompareKeyValueFn != nil {
				t.Error("got non-nil CompareKeyValueFn")
			}

			checkFormatItemFn(t, "key", tc.format, xs, got.FormatKeyFn)
			checkFormatItemFn(t, "value", tc.format, xs, got.FormatValueFn)
		})
	}
}

// checkFormatItemFn checks the specified fmtcoll.FormatFunc formatItemFn
// by comparing its return value with
// the string returned by fmt.Sprintf(wantFormat, x),
// where x is each item in the test data xs.
//
// itemName is used to form error messages.
// If it is empty, "item" is used instead.
func checkFormatItemFn(
	t *testing.T,
	itemName string,
	wantFormat string,
	xs []any,
	formatItemFn fmtcoll.FormatFunc[any],
) {
	t.Helper()

	if itemName == "" {
		itemName = "item"
	}

	var b strings.Builder

	for _, x := range xs {
		want := fmt.Sprintf(wantFormat, x)

		b.Reset()

		err := formatItemFn(&b, x)
		if err != nil {
			t.Error("format", itemName, "failed:", err)
		} else if b.String() != want {
			t.Errorf("x=%v: got %s string %q; want %q",
				x, itemName, b.String(), want)
		}
	}
}
