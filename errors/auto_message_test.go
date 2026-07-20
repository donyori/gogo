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
	"fmt"
	"testing"

	"github.com/donyori/gogo/errors"
)

func TestAutoMsg(t *testing.T) {
	t.Parallel()

	// ".func1" is the anonymous function passed to t.Run
	// that calls function AutoMsg.

	const FullFuncPrefix = "github.com/donyori/gogo/errors_test." +
		"TestAutoMsg.func1: "

	testCases := []struct {
		msg  string
		want string
	}{
		{"", FullFuncPrefix + errors.NoErrorMessage},
		{nonemptyErrorMessage, FullFuncPrefix + nonemptyErrorMessage},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("msg=%+q", tc.msg), func(t *testing.T) {
			t.Parallel()

			s := errors.AutoMsg(tc.msg)
			if s != tc.want {
				t.Errorf("got %q; want %q", s, tc.want)
			}
		})
	}
}

func TestAutoMsgCustom(t *testing.T) {
	t.Parallel()

	// ".func1" is the anonymous function passed to t.Run;
	// ".func1.1" is the anonymous inner function
	// that calls function AutoMsgCustom.

	// Error prefixes.
	const (
		FullFuncSkip0Prefix = "github.com/donyori/gogo/errors_test." +
			"TestAutoMsgCustom.func1.1: "
		FullFuncSkip1Prefix = "github.com/donyori/gogo/errors_test." +
			"TestAutoMsgCustom.func1: "
		FullPkgPrefix = "github.com/donyori/gogo/errors_test: "

		SimpleFuncSkip0Prefix = "TestAutoMsgCustom.func1.1: "
		SimpleFuncSkip1Prefix = "TestAutoMsgCustom.func1: "
		SimplePkgPrefix       = "errors_test: "
	)

	testCases := []struct {
		msg  string
		ms   errors.ErrorMessageStrategy
		skip int
		want string
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
					s := errors.AutoMsgCustom(tc.msg, tc.ms, tc.skip)
					if s != tc.want {
						t.Errorf("got %q; want %q", s, tc.want)
					}
				}()
			},
		)
	}
}

// nonemptyErrorMessage is a nonempty error message for testing.
const nonemptyErrorMessage = "some error"
