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

package unequal_test

import (
	"fmt"
	"testing"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/fmtcoll"
	"github.com/donyori/gogo/internal/unequal"
)

func TestSlice(t *testing.T) {
	testCases := []struct {
		s1, s2 []int
		want   bool
	}{
		{nil, nil, false},
		{nil, []int{}, true},
		{nil, []int{1}, true},
		{nil, []int{2}, true},
		{nil, []int{1, 2}, true},

		{[]int{}, nil, true},
		{[]int{}, []int{}, false},
		{[]int{}, []int{1}, true},
		{[]int{}, []int{2}, true},
		{[]int{}, []int{1, 2}, true},

		{[]int{1}, nil, true},
		{[]int{1}, []int{}, true},
		{[]int{1}, []int{1}, false},
		{[]int{1}, []int{2}, true},
		{[]int{1}, []int{1, 2}, true},

		{[]int{2}, nil, true},
		{[]int{2}, []int{}, true},
		{[]int{2}, []int{1}, true},
		{[]int{2}, []int{2}, false},
		{[]int{2}, []int{1, 2}, true},

		{[]int{1, 2}, nil, true},
		{[]int{1, 2}, []int{}, true},
		{[]int{1, 2}, []int{1}, true},
		{[]int{1, 2}, []int{2}, true},
		{[]int{1, 2}, []int{1, 2}, false},
	}

	for _, tc := range testCases {
		format := &fmtcoll.SequenceFormat[int]{
			CommonFormat: fmtcoll.CommonFormat{
				Separator: ",",
			},
			FormatItemFn: fmtcoll.FprintfToFormatFunc[int]("%d"),
		}
		s1Name := fmtcoll.MustFormatSliceToString(tc.s1, format)
		s2Name := fmtcoll.MustFormatSliceToString(tc.s2, format)
		t.Run(fmt.Sprintf("s1=%s&s2=%s", s1Name, s2Name), func(t *testing.T) {
			got := unequal.Slice(tc.s1, tc.s2)
			if got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
		})
	}
}

func TestMap(t *testing.T) {
	testCases := []struct {
		m1, m2 map[string]int
		want   bool
	}{
		{nil, nil, false},
		{nil, map[string]int{}, true},
		{nil, map[string]int{"A": 1}, true},
		{nil, map[string]int{"A": 2}, true},
		{nil, map[string]int{"B": 1}, true},
		{nil, map[string]int{"A": 1, "B": 2}, true},

		{map[string]int{}, nil, true},
		{map[string]int{}, map[string]int{}, false},
		{map[string]int{}, map[string]int{"A": 1}, true},
		{map[string]int{}, map[string]int{"A": 2}, true},
		{map[string]int{}, map[string]int{"B": 1}, true},
		{map[string]int{}, map[string]int{"A": 1, "B": 2}, true},

		{map[string]int{"A": 1}, nil, true},
		{map[string]int{"A": 1}, map[string]int{}, true},
		{map[string]int{"A": 1}, map[string]int{"A": 1}, false},
		{map[string]int{"A": 1}, map[string]int{"A": 2}, true},
		{map[string]int{"A": 1}, map[string]int{"B": 1}, true},
		{map[string]int{"A": 1}, map[string]int{"A": 1, "B": 2}, true},

		{map[string]int{"A": 2}, nil, true},
		{map[string]int{"A": 2}, map[string]int{}, true},
		{map[string]int{"A": 2}, map[string]int{"A": 1}, true},
		{map[string]int{"A": 2}, map[string]int{"A": 2}, false},
		{map[string]int{"A": 2}, map[string]int{"B": 1}, true},
		{map[string]int{"A": 2}, map[string]int{"A": 1, "B": 2}, true},

		{map[string]int{"B": 1}, nil, true},
		{map[string]int{"B": 1}, map[string]int{}, true},
		{map[string]int{"B": 1}, map[string]int{"A": 1}, true},
		{map[string]int{"B": 1}, map[string]int{"A": 2}, true},
		{map[string]int{"B": 1}, map[string]int{"B": 1}, false},
		{map[string]int{"B": 1}, map[string]int{"A": 1, "B": 2}, true},

		{map[string]int{"A": 1, "B": 2}, nil, true},
		{map[string]int{"A": 1, "B": 2}, map[string]int{}, true},
		{map[string]int{"A": 1, "B": 2}, map[string]int{"A": 1}, true},
		{map[string]int{"A": 1, "B": 2}, map[string]int{"A": 2}, true},
		{map[string]int{"A": 1, "B": 2}, map[string]int{"B": 1}, true},
		{map[string]int{"A": 1, "B": 2}, map[string]int{"A": 1, "B": 2}, false},
	}

	for _, tc := range testCases {
		format := &fmtcoll.MapFormat[string, int]{
			CommonFormat: fmtcoll.CommonFormat{
				Separator: ",",
			},
			FormatKeyFn:   fmtcoll.FprintfToFormatFunc[string]("%+q"),
			FormatValueFn: fmtcoll.FprintfToFormatFunc[int]("%d"),
			CompareKeyValueFn: func(
				key1 string,
				value1 int,
				key2 string,
				value2 int,
			) int {
				switch {
				case key1 < key2:
					return -1
				case key1 > key2:
					return 1
				case value1 < value2:
					return -1
				case value1 > value2:
					return 1
				}
				return 0
			},
		}
		m1Name := fmtcoll.MustFormatMapToString(tc.m1, format)
		m2Name := fmtcoll.MustFormatMapToString(tc.m2, format)
		t.Run(fmt.Sprintf("m1=%s&m2=%s", m1Name, m2Name), func(t *testing.T) {
			got := unequal.Map(tc.m1, tc.m2)
			if got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
		})
	}
}

func TestErrorUnwrapAuto(t *testing.T) {
	err1 := errors.New("error-1")
	err1AutoWrap := errors.AutoWrap(err1)
	err2 := errors.New("error-2")
	err2AutoWrap := errors.AutoWrap(err2)
	var err1AutoWrap2, err2AutoWrap2 error
	func() {
		// Wrap them in an inner function.
		err1AutoWrap2 = errors.AutoWrap(err1AutoWrap)
		err2AutoWrap2 = errors.AutoWrap(err2AutoWrap)
	}()
	testCases := []struct {
		err1, err2 error
		want       bool
	}{
		{nil, nil, false},
		{nil, err1, true},
		{nil, err1AutoWrap, true},
		{nil, err1AutoWrap2, true},
		{nil, err2, true},
		{nil, err2AutoWrap, true},
		{nil, err2AutoWrap2, true},

		{err1, nil, true},
		{err1, err1, false},
		{err1, err1AutoWrap, false},
		{err1, err1AutoWrap2, false},
		{err1, err2, true},
		{err1, err2AutoWrap, true},
		{err1, err2AutoWrap2, true},

		{err1AutoWrap, nil, true},
		{err1AutoWrap, err1, false},
		{err1AutoWrap, err1AutoWrap, false},
		{err1AutoWrap, err1AutoWrap2, false},
		{err1AutoWrap, err2, true},
		{err1AutoWrap, err2AutoWrap, true},
		{err1AutoWrap, err2AutoWrap2, true},

		{err1AutoWrap2, nil, true},
		{err1AutoWrap2, err1, false},
		{err1AutoWrap2, err1AutoWrap, false},
		{err1AutoWrap2, err1AutoWrap2, false},
		{err1AutoWrap2, err2, true},
		{err1AutoWrap2, err2AutoWrap, true},
		{err1AutoWrap2, err2AutoWrap2, true},

		{err2, nil, true},
		{err2, err1, true},
		{err2, err1AutoWrap, true},
		{err2, err1AutoWrap2, true},
		{err2, err2, false},
		{err2, err2AutoWrap, false},
		{err2, err2AutoWrap2, false},

		{err2AutoWrap, nil, true},
		{err2AutoWrap, err1, true},
		{err2AutoWrap, err1AutoWrap, true},
		{err2AutoWrap, err1AutoWrap2, true},
		{err2AutoWrap, err2, false},
		{err2AutoWrap, err2AutoWrap, false},
		{err2AutoWrap, err2AutoWrap2, false},

		{err2AutoWrap2, nil, true},
		{err2AutoWrap2, err1, true},
		{err2AutoWrap2, err1AutoWrap, true},
		{err2AutoWrap2, err1AutoWrap2, true},
		{err2AutoWrap2, err2, false},
		{err2AutoWrap2, err2AutoWrap, false},
		{err2AutoWrap2, err2AutoWrap2, false},
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("err1=%v&err2=%v", tc.err1, tc.err2),
			func(t *testing.T) {
				got := unequal.ErrorUnwrapAuto(tc.err1, tc.err2)
				if got != tc.want {
					t.Errorf("got %t; want %t", got, tc.want)
				}
			},
		)
	}
}
