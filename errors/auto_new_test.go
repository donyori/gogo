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

package errors_test

import (
	stderrors "errors"
	"fmt"
	"testing"

	"github.com/donyori/gogo/errors"
)

func TestAutoNew(t *testing.T) {
	t.Parallel()

	// ".func1" is the anonymous function passed to t.Run
	// that calls function AutoNew.

	const FullFuncPrefix = "github.com/donyori/gogo/errors_test." +
		"TestAutoNew.func1: "

	testCases := []struct {
		msg     string
		wantMsg string
	}{
		{"", FullFuncPrefix + errors.NoErrorMessage},
		{nonemptyErrorMessage, FullFuncPrefix + nonemptyErrorMessage},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("msg=%+q", tc.msg), func(t *testing.T) {
			t.Parallel()

			got := errors.AutoNew(tc.msg)
			if gotMsg := got.Error(); gotMsg != tc.wantMsg {
				t.Errorf("got msg %q; want %q", gotMsg, tc.wantMsg)
			}

			unwrap := stderrors.Unwrap(got)
			if unwrap == nil {
				t.Fatal("Unwrap returns nil")
			}

			unwrapMsg := unwrap.Error()
			if unwrapMsg != tc.msg {
				t.Errorf("unwrap msg %q; want %q", unwrapMsg, tc.msg)
			}
		})
	}
}

func TestAutoNewCustom(t *testing.T) {
	t.Parallel()

	// ".func1" is the anonymous function passed to t.Run;
	// ".func1.1" is the anonymous inner function
	// that calls function AutoNewCustom.

	// Error prefixes.
	const (
		FullFuncSkip0Prefix = "github.com/donyori/gogo/errors_test." +
			"TestAutoNewCustom.func1.1: "
		FullFuncSkip1Prefix = "github.com/donyori/gogo/errors_test." +
			"TestAutoNewCustom.func1: "
		FullPkgPrefix = "github.com/donyori/gogo/errors_test: "

		SimpleFuncSkip0Prefix = "TestAutoNewCustom.func1.1: "
		SimpleFuncSkip1Prefix = "TestAutoNewCustom.func1: "
		SimplePkgPrefix       = "errors_test: "
	)

	testCases := []struct {
		msg     string
		ms      errors.ErrorMessageStrategy
		skip    int
		wantMsg string
	}{
		{
			"",
			-1,
			0,
			FullFuncSkip0Prefix + errors.NoErrorMessage,
		},
		{
			"",
			-1,
			1,
			FullFuncSkip1Prefix + errors.NoErrorMessage,
		},
		{
			"",
			errors.OriginalMsg,
			0,
			errors.NoErrorMessage,
		},
		{
			"",
			errors.OriginalMsg,
			1,
			errors.NoErrorMessage,
		},
		{
			"",
			errors.PrependFullFuncName,
			0,
			FullFuncSkip0Prefix + errors.NoErrorMessage,
		},
		{
			"",
			errors.PrependFullFuncName,
			1,
			FullFuncSkip1Prefix + errors.NoErrorMessage,
		},
		{
			"",
			errors.PrependFullPkgName,
			0,
			FullPkgPrefix + errors.NoErrorMessage,
		},
		{
			"",
			errors.PrependFullPkgName,
			1,
			FullPkgPrefix + errors.NoErrorMessage,
		},
		{
			"",
			errors.PrependSimpleFuncName,
			0,
			SimpleFuncSkip0Prefix + errors.NoErrorMessage,
		},
		{
			"",
			errors.PrependSimpleFuncName,
			1,
			SimpleFuncSkip1Prefix + errors.NoErrorMessage,
		},
		{
			"",
			errors.PrependSimplePkgName,
			0,
			SimplePkgPrefix + errors.NoErrorMessage,
		},
		{
			"",
			errors.PrependSimplePkgName,
			1,
			SimplePkgPrefix + errors.NoErrorMessage,
		},
		{
			"",
			errors.PrependSimplePkgName + 1,
			0,
			FullFuncSkip0Prefix + errors.NoErrorMessage,
		},
		{
			"",
			errors.PrependSimplePkgName + 1,
			1,
			FullFuncSkip1Prefix + errors.NoErrorMessage,
		},
		{
			nonemptyErrorMessage,
			-1,
			0,
			FullFuncSkip0Prefix + nonemptyErrorMessage,
		},
		{
			nonemptyErrorMessage,
			-1,
			1,
			FullFuncSkip1Prefix + nonemptyErrorMessage,
		},
		{
			nonemptyErrorMessage,
			errors.OriginalMsg,
			0,
			nonemptyErrorMessage,
		},
		{
			nonemptyErrorMessage,
			errors.OriginalMsg,
			1,
			nonemptyErrorMessage,
		},
		{
			nonemptyErrorMessage,
			errors.PrependFullFuncName,
			0,
			FullFuncSkip0Prefix + nonemptyErrorMessage,
		},
		{
			nonemptyErrorMessage,
			errors.PrependFullFuncName,
			1,
			FullFuncSkip1Prefix + nonemptyErrorMessage,
		},
		{
			nonemptyErrorMessage,
			errors.PrependFullPkgName,
			0,
			FullPkgPrefix + nonemptyErrorMessage,
		},
		{
			nonemptyErrorMessage,
			errors.PrependFullPkgName,
			1,
			FullPkgPrefix + nonemptyErrorMessage,
		},
		{
			nonemptyErrorMessage,
			errors.PrependSimpleFuncName,
			0,
			SimpleFuncSkip0Prefix + nonemptyErrorMessage,
		},
		{
			nonemptyErrorMessage,
			errors.PrependSimpleFuncName,
			1,
			SimpleFuncSkip1Prefix + nonemptyErrorMessage,
		},
		{
			nonemptyErrorMessage,
			errors.PrependSimplePkgName,
			0,
			SimplePkgPrefix + nonemptyErrorMessage,
		},
		{
			nonemptyErrorMessage,
			errors.PrependSimplePkgName,
			1,
			SimplePkgPrefix + nonemptyErrorMessage,
		},
		{
			nonemptyErrorMessage,
			errors.PrependSimplePkgName + 1,
			0,
			FullFuncSkip0Prefix + nonemptyErrorMessage,
		},
		{
			nonemptyErrorMessage,
			errors.PrependSimplePkgName + 1,
			1,
			FullFuncSkip1Prefix + nonemptyErrorMessage,
		},
	}

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case%d?msg=%+q&ms=%s(%[3]d)&skip=%d",
				i, tc.msg, tc.ms, tc.skip),
			func(t *testing.T) {
				t.Parallel()

				func() { // use an inner function to test the "skip"
					got := errors.AutoNewCustom(tc.msg, tc.ms, tc.skip)
					if gotMsg := got.Error(); gotMsg != tc.wantMsg {
						t.Errorf("got msg %q; want %q", gotMsg, tc.wantMsg)
					}

					unwrap := stderrors.Unwrap(got)
					if unwrap == nil {
						t.Fatal("Unwrap returns nil")
					}

					unwrapMsg := unwrap.Error()
					if unwrapMsg != tc.msg {
						t.Errorf("unwrap msg %q; want %q", unwrapMsg, tc.msg)
					}
				}()
			},
		)
	}
}
