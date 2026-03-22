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
	"io"
	"slices"
	"strconv"
	"testing"

	"github.com/donyori/gogo/errors"
)

// nilErrorString is the string that represents a nil error.
const nilErrorString = "<nil>"

type testError struct {
	err error
}

func (te *testError) Error() string {
	return "testError wrapped on " + strconv.Quote(te.err.Error())
}

func (te *testError) Unwrap() error {
	return te.err
}

func TestAutoWrap(t *testing.T) {
	t.Parallel()

	// ".func1" is the anonymous function passed to t.Run.

	const FullFuncPrefix = "github.com/donyori/gogo/errors_test." +
		"TestAutoWrap.func1: "

	err0 := stderrors.New("error 0")
	err1 := errors.NewAutoWrappedError(err0, "1")
	err2 := errors.NewAutoWrappedError(err1, "2")
	wantMsgErr0To2 := FullFuncPrefix + err0.Error()

	err3 := &testError{err: err2}
	err4 := errors.NewAutoWrappedError(err3, "4")
	err5 := errors.NewAutoWrappedError(err4, "5")
	wantMsgErr3To5 := FullFuncPrefix + err3.Error()

	err6 := stderrors.New("")
	wantMsgErr6 := FullFuncPrefix + "<no error message>"

	testCases := []struct {
		err     error
		equal   bool
		wantMsg string
	}{
		{nil, true, ""},
		{io.EOF, true, io.EOF.Error()},
		{err0, false, wantMsgErr0To2},
		{err1, false, wantMsgErr0To2},
		{err2, false, wantMsgErr0To2},
		{err3, false, wantMsgErr3To5},
		{err4, false, wantMsgErr3To5},
		{err5, false, wantMsgErr3To5},
		{err6, false, wantMsgErr6},
	}

	for i, tc := range testCases {
		errName := nilErrorString
		if tc.err != nil {
			errName = strconv.QuoteToASCII(tc.err.Error())
		}

		t.Run(fmt.Sprintf("case%d?err=%s", i, errName), func(t *testing.T) {
			t.Parallel()

			got := errors.AutoWrap(tc.err)
			checkAutoWrapFamilyResult(t, got, tc.err, tc.equal, tc.wantMsg)
		})
	}
}

func TestAutoWrapSkip(t *testing.T) {
	t.Parallel()

	// ".func1.1" is the anonymous inner function
	// that calls function AutoWrapSkip.

	const FullFuncSkip0Prefix = "github.com/donyori/gogo/errors_test." +
		"TestAutoWrapSkip.func1.1: "

	// ".func1" is the anonymous function passed to t.Run.

	const FullFuncSkip1Prefix = "github.com/donyori/gogo/errors_test." +
		"TestAutoWrapSkip.func1: "

	err0 := stderrors.New("error 0")
	err1 := errors.NewAutoWrappedError(err0, "1")
	err2 := errors.NewAutoWrappedError(err1, "2")
	wantMsgErr0To2Skip0 := FullFuncSkip0Prefix + err0.Error()
	wantMsgErr0To2Skip1 := FullFuncSkip1Prefix + err0.Error()

	err3 := &testError{err: err2}
	err4 := errors.NewAutoWrappedError(err3, "4")
	err5 := errors.NewAutoWrappedError(err4, "5")
	wantMsgErr3To5Skip0 := FullFuncSkip0Prefix + err3.Error()
	wantMsgErr3To5Skip1 := FullFuncSkip1Prefix + err3.Error()

	err6 := stderrors.New("")
	wantMsgErr6Skip0 := FullFuncSkip0Prefix + "<no error message>"
	wantMsgErr6Skip1 := FullFuncSkip1Prefix + "<no error message>"

	testCases := []struct {
		err     error
		skip    int
		equal   bool
		wantMsg string
	}{
		{nil, 0, true, ""},
		{nil, 1, true, ""},
		{io.EOF, 0, true, io.EOF.Error()},
		{io.EOF, 1, true, io.EOF.Error()},
		{err0, 0, false, wantMsgErr0To2Skip0},
		{err0, 1, false, wantMsgErr0To2Skip1},
		{err1, 0, false, wantMsgErr0To2Skip0},
		{err1, 1, false, wantMsgErr0To2Skip1},
		{err2, 0, false, wantMsgErr0To2Skip0},
		{err2, 1, false, wantMsgErr0To2Skip1},
		{err3, 0, false, wantMsgErr3To5Skip0},
		{err3, 1, false, wantMsgErr3To5Skip1},
		{err4, 0, false, wantMsgErr3To5Skip0},
		{err4, 1, false, wantMsgErr3To5Skip1},
		{err5, 0, false, wantMsgErr3To5Skip0},
		{err5, 1, false, wantMsgErr3To5Skip1},
		{err6, 0, false, wantMsgErr6Skip0},
		{err6, 1, false, wantMsgErr6Skip1},
	}

	for i, tc := range testCases {
		errName := nilErrorString
		if tc.err != nil {
			errName = strconv.QuoteToASCII(tc.err.Error())
		}

		t.Run(
			fmt.Sprintf("case%d?err=%s&skip=%d", i, errName, tc.skip),
			func(t *testing.T) {
				t.Parallel()

				func() { // use an inner function to test the "skip"
					got := errors.AutoWrapSkip(tc.err, tc.skip)
					checkAutoWrapFamilyResult(
						t,
						got,
						tc.err,
						tc.equal,
						tc.wantMsg,
					)
				}()
			},
		)
	}
}

type autoWrapCustomTestCase struct {
	err     error
	ms      errors.ErrorMessageStrategy
	skip    int
	excl    errors.ErrorReadOnlySet
	equal   bool
	wantMsg string
}

func TestAutoWrapCustom(t *testing.T) {
	t.Parallel()

	for i, tc := range getTestCasesForAutoWrapCustom(t) {
		errName := nilErrorString
		if tc.err != nil {
			errName = strconv.QuoteToASCII(tc.err.Error())
		}

		t.Run(
			fmt.Sprintf("case%d?err=%s&ms=%s(%[3]d)&skip=%d&hasExcl=%t",
				i, errName, tc.ms, tc.skip, tc.excl != nil),
			func(t *testing.T) {
				t.Parallel()

				func() { // use an inner function to test the "skip"
					got := errors.AutoWrapCustom(
						tc.err,
						tc.ms,
						tc.skip,
						tc.excl,
					)
					checkAutoWrapFamilyResult(
						t,
						got,
						tc.err,
						tc.equal,
						tc.wantMsg,
					)
				}()
			},
		)
	}
}

// getTestCasesForAutoWrapCustom returns test cases for TestAutoWrapCustom.
//
// It uses t.Fatal to stop the test if there is something wrong.
func getTestCasesForAutoWrapCustom(t *testing.T) []autoWrapCustomTestCase {
	t.Helper()

	err0 := stderrors.New("error 0")
	err1 := errors.NewAutoWrappedError(err0, "1")
	err2 := errors.NewAutoWrappedError(err1, "2")
	err3 := &testError{err: err2}
	err4 := errors.NewAutoWrappedError(err3, "4")
	err5 := errors.NewAutoWrappedError(err4, "5")
	err6 := stderrors.New("")

	excl := errors.NewErrorReadOnlySetIs(io.EOF, err0)

	errorList := []error{nil, io.EOF, err0, err1, err2, err3, err4, err5, err6}
	exclList := []errors.ErrorReadOnlySet{nil, excl}
	testCases := make(
		[]autoWrapCustomTestCase,
		len(errorList)*len(exclList)*(int(errors.PrependSimplePkgName)+3)<<1,
	)

	var idx int

	for _, err := range errorList {
		ms := errors.ErrorMessageStrategy(-1)
		for ms <= errors.PrependSimplePkgName+1 {
			for skip := range 2 {
				for _, excl := range exclList {
					if idx >= len(testCases) {
						t.Fatal("not enough test cases; please update")
					}

					testCases[idx].err = err
					testCases[idx].ms = ms
					testCases[idx].skip = skip
					testCases[idx].excl = excl
					testCases[idx].equal = err == nil ||
						excl != nil && excl.Contains(err)
					testCases[idx].wantMsg = getWantMessageForAutoWrapCustom(
						err0,
						err1,
						err2,
						err3,
						err4,
						err5,
						err,
						ms,
						skip,
						excl,
					)
					idx++
				}
			}

			ms++
		}
	}

	if idx != len(testCases) {
		t.Fatal("excessive test cases; please update")
	}

	return testCases
}

// getWantMessageForAutoWrapCustom returns the expected error message
// for TestAutoWrapCustom.
func getWantMessageForAutoWrapCustom(
	err0 error,
	err1 error,
	err2 error,
	err3 error,
	err4 error,
	err5 error,
	err error,
	ms errors.ErrorMessageStrategy,
	skip int,
	excl errors.ErrorReadOnlySet,
) string {
	if err == nil {
		return ""
	} else if excl != nil && excl.Contains(err) {
		return err.Error()
	}

	wantPrefix := getWantMessagePrefixForAutoWrapCustom(ms, skip)

	var errMsg string

	// Compare the interface directly here, don't use errors.Is.
	//
	//nolint:errorlint // as stated above
	switch err {
	case err1, err2:
		errMsg = err0.Error()
	case err4, err5:
		errMsg = err3.Error()
	default:
		errMsg = err.Error()
	}

	if errMsg == "" {
		errMsg = "<no error message>"
	}

	return wantPrefix + errMsg
}

// getWantMessagePrefixForAutoWrapCustom returns
// the expected error message prefix for TestAutoWrapCustom.
func getWantMessagePrefixForAutoWrapCustom(
	ms errors.ErrorMessageStrategy,
	skip int,
) string {
	// ".func1" is the anonymous function passed to t.Run.
	// ".func1.1" is the anonymous inner function
	// that calls function AutoWrapCustom.

	// Error prefixes.
	const (
		FullFuncSkip0Prefix = "github.com/donyori/gogo/errors_test." +
			"TestAutoWrapCustom.func1.1: "
		FullFuncSkip1Prefix = "github.com/donyori/gogo/errors_test." +
			"TestAutoWrapCustom.func1: "
		FullPkgPrefix = "github.com/donyori/gogo/errors_test: "

		SimpleFuncSkip0Prefix = "TestAutoWrapCustom.func1.1: "
		SimpleFuncSkip1Prefix = "TestAutoWrapCustom.func1: "
		SimplePkgPrefix       = "errors_test: "
	)

	switch ms {
	case errors.OriginalMsg:
		return ""
	case errors.PrependFullPkgName:
		return FullPkgPrefix
	case errors.PrependSimpleFuncName:
		if skip == 0 {
			return SimpleFuncSkip0Prefix
		}

		return SimpleFuncSkip1Prefix
	case errors.PrependSimplePkgName:
		return SimplePkgPrefix
	}

	if skip == 0 {
		return FullFuncSkip0Prefix
	}

	return FullFuncSkip1Prefix
}

// checkAutoWrapFamilyResult checks the result error returned by
// AutoWrap, AutoWrapSkip, and AutoWrapCustom.
func checkAutoWrapFamilyResult(
	t *testing.T,
	got error,
	inputErr error,
	wantEqual bool,
	wantMsg string,
) {
	t.Helper()

	// Compare the interface directly here, don't use errors.Is.
	if (got == inputErr) != wantEqual { //nolint:err113,errorlint // as stated above
		if wantEqual {
			t.Errorf("got %q; != tc.err", got)
		} else {
			t.Errorf("got %q; == tc.err", got)
		}
	}

	if inputErr == nil || got == nil || wantEqual {
		return
	}

	gotMsg := got.Error()
	if gotMsg != wantMsg {
		t.Errorf("got msg %q; want %q", gotMsg, wantMsg)
	}

	gotUnwrap := stderrors.Unwrap(got)

	// Compare the interface directly here, don't use errors.Is.
	if gotUnwrap != inputErr { //nolint:err113,errorlint // as stated above
		t.Errorf("got unwrap %q; != tc.err", gotUnwrap)
	}
}

func TestIsAutoWrappedError(t *testing.T) {
	t.Parallel()

	err0 := stderrors.New("error 0")
	err1 := errors.NewAutoWrappedError(err0, "1")
	err2 := errors.NewAutoWrappedError(err1, "2")
	err3 := errors.NewAutoWrappedError(err2, "3")
	err4 := &testError{err: err2}
	err5 := errors.NewAutoWrappedError(err4, "5")
	err6 := errors.NewAutoWrappedError(err5, "6")

	testCases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{err0, false},
		{err1, true},
		{err2, true},
		{err3, true},
		{err4, false},
		{err5, true},
		{err6, true},
	}

	for i, tc := range testCases {
		errName := nilErrorString
		if tc.err != nil {
			errName = strconv.QuoteToASCII(tc.err.Error())
		}

		t.Run(fmt.Sprintf("case%d?err=%s", i, errName), func(t *testing.T) {
			t.Parallel()

			got := errors.IsAutoWrappedError(tc.err)
			if got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
		})
	}
}

func TestUnwrapAutoWrappedError(t *testing.T) {
	t.Parallel()

	err0 := stderrors.New("error 0")
	err1 := errors.NewAutoWrappedError(err0, "1")
	err2 := errors.NewAutoWrappedError(err1, "2")
	err3 := errors.NewAutoWrappedError(err2, "3")
	err4 := &testError{err: err2}
	err5 := errors.NewAutoWrappedError(err4, "5")
	err6 := errors.NewAutoWrappedError(err5, "6")

	testCases := []struct {
		err     error
		wantErr error
	}{
		{nil, nil},
		{err0, err0},
		{err1, err0},
		{err2, err1},
		{err3, err2},
		{err4, err4},
		{err5, err4},
		{err6, err5},
	}

	for i, tc := range testCases {
		errName := nilErrorString
		if tc.err != nil {
			errName = strconv.QuoteToASCII(tc.err.Error())
		}

		// Compare the interface directly here, don't use errors.Is.
		wantBool := tc.err != tc.wantErr //nolint:err113,errorlint // as stated above

		t.Run(fmt.Sprintf("case%d?err=%s", i, errName), func(t *testing.T) {
			t.Parallel()

			gotErr, gotBool := errors.UnwrapAutoWrappedError(tc.err)

			// Compare the interface directly here, don't use errors.Is.
			if gotErr != tc.wantErr || gotBool != wantBool { //nolint:err113,errorlint // as stated above
				t.Errorf("got (%q, %t); want (%q, %t)",
					gotErr, gotBool, tc.wantErr, wantBool)
			}
		})
	}
}

func TestUnwrapAllAutoWrappedErrors(t *testing.T) {
	t.Parallel()

	err0 := stderrors.New("error 0")
	err1 := errors.NewAutoWrappedError(err0, "1")
	err2 := errors.NewAutoWrappedError(err1, "2")
	err3 := errors.NewAutoWrappedError(err2, "3")
	err4 := &testError{err: err2}
	err5 := errors.NewAutoWrappedError(err4, "5")
	err6 := errors.NewAutoWrappedError(err5, "6")

	testCases := []struct {
		err     error
		wantErr error
	}{
		{nil, nil},
		{err0, err0},
		{err1, err0},
		{err2, err0},
		{err3, err0},
		{err4, err4},
		{err5, err4},
		{err6, err4},
	}

	for i, tc := range testCases {
		errName := nilErrorString
		if tc.err != nil {
			errName = strconv.QuoteToASCII(tc.err.Error())
		}

		// Compare the interface directly here, don't use errors.Is.
		wantBool := tc.err != tc.wantErr //nolint:err113,errorlint // as stated above

		t.Run(fmt.Sprintf("case%d?err=%s", i, errName), func(t *testing.T) {
			t.Parallel()

			gotErr, gotBool := errors.UnwrapAllAutoWrappedErrors(tc.err)

			// Compare the interface directly here, don't use errors.Is.
			if gotErr != tc.wantErr || gotBool != wantBool { //nolint:err113,errorlint // as stated above
				t.Errorf("got (%q, %t); want (%q, %t)",
					gotErr, gotBool, tc.wantErr, wantBool)
			}
		})
	}
}

func TestListFunctionNamesInAutoWrappedErrors(t *testing.T) {
	t.Parallel()

	const FullFunc = "github.com/donyori/gogo/errors_test." +
		"TestListFunctionNamesInAutoWrappedErrors"

	err0 := stderrors.New("error 0")
	err1 := errors.NewAutoWrappedError(err0, "1")
	err2 := errors.NewAutoWrappedError(err1, "2")
	err3 := errors.NewAutoWrappedError(err2, "3")
	err4 := &testError{err: err2}
	err5 := errors.NewAutoWrappedError(err4, "5")
	err6 := errors.NewAutoWrappedError(err5, "6")

	testCases := []struct {
		err       error
		wantNames []string
		wantRoot  error
	}{
		{
			nil,
			nil,
			nil,
		},
		{
			err0,
			nil,
			err0,
		},
		{
			err1,
			[]string{FullFunc + "_1"},
			err0,
		},
		{
			err2,
			[]string{FullFunc + "_2", FullFunc + "_1"},
			err0,
		},
		{
			err3,
			[]string{FullFunc + "_3", FullFunc + "_2", FullFunc + "_1"},
			err0,
		},
		{
			err4,
			nil,
			err4,
		},
		{
			err5,
			[]string{FullFunc + "_5"},
			err4,
		},
		{
			err6,
			[]string{FullFunc + "_6", FullFunc + "_5"},
			err4,
		},
	}

	for i, tc := range testCases {
		errName := nilErrorString
		if tc.err != nil {
			errName = strconv.QuoteToASCII(tc.err.Error())
		}

		t.Run(fmt.Sprintf("case%d?err=%s", i, errName), func(t *testing.T) {
			t.Parallel()

			gotNames, gotRoot := errors.ListFunctionNamesInAutoWrappedErrors(
				tc.err)

			switch {
			case tc.wantNames == nil:
				if gotNames != nil {
					t.Errorf("got names %v; want <nil>", gotNames)
				}
			case gotNames == nil:
				t.Errorf("got names <nil>; want %v", tc.wantNames)
			case !slices.Equal(gotNames, tc.wantNames):
				t.Errorf("got names %v; want %v", gotNames, tc.wantNames)
			}

			// Compare the interface directly here, don't use errors.Is.
			if gotRoot != tc.wantRoot { //nolint:err113,errorlint // as stated above
				t.Errorf("got root %v; want %v", gotRoot, tc.wantRoot)
			}
		})
	}
}
