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

package errors_test

import (
	stderrors "errors"
	"fmt"
	"testing"

	"github.com/donyori/gogo/errors"
)

func TestAutoNew(t *testing.T) {
	testCases := []struct {
		msg     string
		wantMsg string
	}{
		{"", "github.com/donyori/gogo/errors_test.TestAutoNew.func1: <no error message>"},
		{"some error", "github.com/donyori/gogo/errors_test.TestAutoNew.func1: some error"},
	}
	// In the above testCases.wantMsg, ".func1" is the anonymous function passed to t.Run.

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("msg=%q", tc.msg), func(t *testing.T) {
			got := errors.AutoNew(tc.msg)
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
		ms      errors.ErrorMessageStrategy
		skip    int
		wantMsg string
	}{
		{"", -1, 0, "github.com/donyori/gogo/errors_test.TestAutoNewCustom.func1.1: <no error message>"},
		{"", -1, 1, "github.com/donyori/gogo/errors_test.TestAutoNewCustom.func1: <no error message>"},
		{"", errors.OriginalMsg, 0, "<no error message>"},
		{"", errors.OriginalMsg, 1, "<no error message>"},
		{"", errors.PrependFullFuncName, 0, "github.com/donyori/gogo/errors_test.TestAutoNewCustom.func1.1: <no error message>"},
		{"", errors.PrependFullFuncName, 1, "github.com/donyori/gogo/errors_test.TestAutoNewCustom.func1: <no error message>"},
		{"", errors.PrependFullPkgName, 0, "github.com/donyori/gogo/errors_test: <no error message>"},
		{"", errors.PrependFullPkgName, 1, "github.com/donyori/gogo/errors_test: <no error message>"},
		{"", errors.PrependSimpleFuncName, 0, "TestAutoNewCustom.func1.1: <no error message>"},
		{"", errors.PrependSimpleFuncName, 1, "TestAutoNewCustom.func1: <no error message>"},
		{"", errors.PrependSimplePkgName, 0, "errors_test: <no error message>"},
		{"", errors.PrependSimplePkgName, 1, "errors_test: <no error message>"},
		{"", errors.PrependSimplePkgName + 1, 0, "github.com/donyori/gogo/errors_test.TestAutoNewCustom.func1.1: <no error message>"},
		{"", errors.PrependSimplePkgName + 1, 1, "github.com/donyori/gogo/errors_test.TestAutoNewCustom.func1: <no error message>"},
		{"some error", -1, 0, "github.com/donyori/gogo/errors_test.TestAutoNewCustom.func1.1: some error"},
		{"some error", -1, 1, "github.com/donyori/gogo/errors_test.TestAutoNewCustom.func1: some error"},
		{"some error", errors.OriginalMsg, 0, "some error"},
		{"some error", errors.OriginalMsg, 1, "some error"},
		{"some error", errors.PrependFullFuncName, 0, "github.com/donyori/gogo/errors_test.TestAutoNewCustom.func1.1: some error"},
		{"some error", errors.PrependFullFuncName, 1, "github.com/donyori/gogo/errors_test.TestAutoNewCustom.func1: some error"},
		{"some error", errors.PrependFullPkgName, 0, "github.com/donyori/gogo/errors_test: some error"},
		{"some error", errors.PrependFullPkgName, 1, "github.com/donyori/gogo/errors_test: some error"},
		{"some error", errors.PrependSimpleFuncName, 0, "TestAutoNewCustom.func1.1: some error"},
		{"some error", errors.PrependSimpleFuncName, 1, "TestAutoNewCustom.func1: some error"},
		{"some error", errors.PrependSimplePkgName, 0, "errors_test: some error"},
		{"some error", errors.PrependSimplePkgName, 1, "errors_test: some error"},
		{"some error", errors.PrependSimplePkgName + 1, 0, "github.com/donyori/gogo/errors_test.TestAutoNewCustom.func1.1: some error"},
		{"some error", errors.PrependSimplePkgName + 1, 1, "github.com/donyori/gogo/errors_test.TestAutoNewCustom.func1: some error"},
	}
	// In the above testCases.wantMsg, ".func1" is the anonymous function passed to t.Run;
	// ".func1.1" is the anonymous inner function that calls function AutoNewCustom.

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?msg=%q&ms=%s(%[3]d)&skip=%d", i, tc.msg, tc.ms, tc.skip), func(t *testing.T) {
			func() { // Use an inner function to test the "skip".
				got := errors.AutoNewCustom(tc.msg, tc.ms, tc.skip)
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
