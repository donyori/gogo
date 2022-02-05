// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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

package errors

import (
	stderrors "errors"
	"fmt"
	"testing"
)

func TestAutoNew(t *testing.T) {
	testCases := []struct {
		msg     string
		wantMsg string
	}{
		{"", "github.com/donyori/gogo/errors.TestAutoNew.func1: (no error message)"},
		{"some error", "github.com/donyori/gogo/errors.TestAutoNew.func1: some error"},
	}
	// In the above testCases.wantMsg, ".func1" is the anonymous function passed to t.Run.

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("msg=%q", tc.msg), func(t *testing.T) {
			got := AutoNew(tc.msg)
			if gotMsg := got.Error(); gotMsg != tc.wantMsg {
				t.Errorf("got msg %q; want %q", gotMsg, tc.wantMsg)
			}
			unwrap := stderrors.Unwrap(got)
			if unwrap == nil {
				t.Fatal("Unwrap returns nil")
			}
			if unwrapMsg := unwrap.Error(); unwrapMsg != tc.msg {
				t.Errorf("unwrap msg %q; want %q", unwrapMsg, tc.msg)
			}
		})
	}
}

func TestAutoNewCustom(t *testing.T) {
	testCases := []struct {
		msg     string
		ms      ErrorMessageStrategy
		skip    int
		wantMsg string
	}{
		{"", -1, 0, "github.com/donyori/gogo/errors.TestAutoNewCustom.func1.1: (no error message)"},
		{"", -1, 1, "github.com/donyori/gogo/errors.TestAutoNewCustom.func1: (no error message)"},
		{"", OriginalMsg, 0, "(no error message)"},
		{"", OriginalMsg, 1, "(no error message)"},
		{"", PrependFullFuncName, 0, "github.com/donyori/gogo/errors.TestAutoNewCustom.func1.1: (no error message)"},
		{"", PrependFullFuncName, 1, "github.com/donyori/gogo/errors.TestAutoNewCustom.func1: (no error message)"},
		{"", PrependFullPkgName, 0, "github.com/donyori/gogo/errors: (no error message)"},
		{"", PrependFullPkgName, 1, "github.com/donyori/gogo/errors: (no error message)"},
		{"", PrependSimpleFuncName, 0, "TestAutoNewCustom.func1.1: (no error message)"},
		{"", PrependSimpleFuncName, 1, "TestAutoNewCustom.func1: (no error message)"},
		{"", PrependSimplePkgName, 0, "errors: (no error message)"},
		{"", PrependSimplePkgName, 1, "errors: (no error message)"},
		{"", PrependSimplePkgName + 1, 0, "github.com/donyori/gogo/errors.TestAutoNewCustom.func1.1: (no error message)"},
		{"", PrependSimplePkgName + 1, 1, "github.com/donyori/gogo/errors.TestAutoNewCustom.func1: (no error message)"},
		{"some error", -1, 0, "github.com/donyori/gogo/errors.TestAutoNewCustom.func1.1: some error"},
		{"some error", -1, 1, "github.com/donyori/gogo/errors.TestAutoNewCustom.func1: some error"},
		{"some error", OriginalMsg, 0, "some error"},
		{"some error", OriginalMsg, 1, "some error"},
		{"some error", PrependFullFuncName, 0, "github.com/donyori/gogo/errors.TestAutoNewCustom.func1.1: some error"},
		{"some error", PrependFullFuncName, 1, "github.com/donyori/gogo/errors.TestAutoNewCustom.func1: some error"},
		{"some error", PrependFullPkgName, 0, "github.com/donyori/gogo/errors: some error"},
		{"some error", PrependFullPkgName, 1, "github.com/donyori/gogo/errors: some error"},
		{"some error", PrependSimpleFuncName, 0, "TestAutoNewCustom.func1.1: some error"},
		{"some error", PrependSimpleFuncName, 1, "TestAutoNewCustom.func1: some error"},
		{"some error", PrependSimplePkgName, 0, "errors: some error"},
		{"some error", PrependSimplePkgName, 1, "errors: some error"},
		{"some error", PrependSimplePkgName + 1, 0, "github.com/donyori/gogo/errors.TestAutoNewCustom.func1.1: some error"},
		{"some error", PrependSimplePkgName + 1, 1, "github.com/donyori/gogo/errors.TestAutoNewCustom.func1: some error"},
	}
	// In the above testCases.wantMsg, ".func1" is the anonymous function passed to t.Run;
	// ".func1.1" is the anonymous inner function that calls function AutoNewCustom.

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?msg=%q&ms=%s(%[3]d)&skip=%d", i, tc.msg, tc.ms, tc.skip), func(t *testing.T) {
			func() { // Use an inner function to test the "skip".
				got := AutoNewCustom(tc.msg, tc.ms, tc.skip)
				if gotMsg := got.Error(); gotMsg != tc.wantMsg {
					t.Errorf("got msg %q; want %q", gotMsg, tc.wantMsg)
				}
				unwrap := stderrors.Unwrap(got)
				if unwrap == nil {
					t.Fatal("Unwrap returns nil")
				}
				if unwrapMsg := unwrap.Error(); unwrapMsg != tc.msg {
					t.Errorf("unwrap msg %q; want %q", unwrapMsg, tc.msg)
				}
			}()
		})
	}
}
