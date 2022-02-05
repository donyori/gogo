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
	"io"
	"testing"
)

func TestAutoWrap(t *testing.T) {
	err0 := stderrors.New("error 0")
	err1 := &autoWrapError{err0, "manually created: error 1 - " + err0.Error()}
	err2 := &autoWrapError{err1, "manually created: error 2 - " + err1.Error()}
	wantMsg := "github.com/donyori/gogo/errors.TestAutoWrap.func1: " + err0.Error() // ".func1" is the anonymous function passed to t.Run.

	testCases := []struct {
		err     error
		equal   bool
		wantMsg string
	}{
		{nil, true, ""},
		{io.EOF, true, io.EOF.Error()},
		{err0, false, wantMsg},
		{err1, false, wantMsg},
		{err2, false, wantMsg},
		{stderrors.New(""), false, "github.com/donyori/gogo/errors.TestAutoWrap.func1: (no error message)"},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?err=%q", i, tc.err), func(t *testing.T) {
			got := AutoWrap(tc.err)
			if (got == tc.err) != tc.equal {
				if tc.equal {
					t.Errorf("got %q; != tc.err", got)
				} else {
					t.Errorf("got %q; == tc.err", got)
				}
			}
			if tc.err == nil || got == nil || tc.equal {
				return
			}
			if gotMsg := got.Error(); gotMsg != tc.wantMsg {
				t.Errorf("got msg %q; want %q", gotMsg, tc.wantMsg)
			}
			if gotUnwrap := stderrors.Unwrap(got); gotUnwrap != tc.err {
				t.Errorf("got unwrap %q; != tc.err", gotUnwrap)
			}
		})
	}
}

func TestAutoWrapSkip(t *testing.T) {
	err0 := stderrors.New("error 0")
	err1 := &autoWrapError{err0, "manually created: error 1 - " + err0.Error()}
	err2 := &autoWrapError{err1, "manually created: error 2 - " + err1.Error()}
	wantMsg0To2Skip0 := "github.com/donyori/gogo/errors.TestAutoWrapSkip.func1.1: " + err0.Error()
	wantMsg0To2Skip1 := "github.com/donyori/gogo/errors.TestAutoWrapSkip.func1: " + err0.Error()
	err3 := stderrors.New("")
	wantMsg3Skip0 := "github.com/donyori/gogo/errors.TestAutoWrapSkip.func1.1: (no error message)"
	wantMsg3Skip1 := "github.com/donyori/gogo/errors.TestAutoWrapSkip.func1: (no error message)"
	// In the above wantXxx, ".func1" is the anonymous function passed to t.Run;
	// ".func1.1" is the anonymous inner function that calls function AutoWrapSkip.

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
		{err0, 0, false, wantMsg0To2Skip0},
		{err0, 1, false, wantMsg0To2Skip1},
		{err1, 0, false, wantMsg0To2Skip0},
		{err1, 1, false, wantMsg0To2Skip1},
		{err2, 0, false, wantMsg0To2Skip0},
		{err2, 1, false, wantMsg0To2Skip1},
		{err3, 0, false, wantMsg3Skip0},
		{err3, 1, false, wantMsg3Skip1},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?err=%q&skip=%d", i, tc.err, tc.skip), func(t *testing.T) {
			func() { // Use an inner function to test the "skip".
				got := AutoWrapSkip(tc.err, tc.skip)
				if (got == tc.err) != tc.equal {
					if tc.equal {
						t.Errorf("got %q; != tc.err", got)
					} else {
						t.Errorf("got %q; == tc.err", got)
					}
				}
				if tc.err == nil || got == nil || tc.equal {
					return
				}
				if gotMsg := got.Error(); gotMsg != tc.wantMsg {
					t.Errorf("got msg %q; want %q", gotMsg, tc.wantMsg)
				}
				if gotUnwrap := stderrors.Unwrap(got); gotUnwrap != tc.err {
					t.Errorf("got unwrap %q; != tc.err", gotUnwrap)
				}
			}()
		})
	}
}

func TestAutoWrapCustom(t *testing.T) {
	err0 := stderrors.New("error 0")
	err1 := &autoWrapError{err0, "manually created: error 1 - " + err0.Error()}
	err2 := &autoWrapError{err1, "manually created: error 2 - " + err1.Error()}
	excl := NewErrorReadOnlySetIs(io.EOF, err0)
	wantMsgPrefixPrependFullFuncNameSkip0 := "github.com/donyori/gogo/errors.TestAutoWrapCustom.func1.1: "
	wantMsgPrefixPrependFullFuncNameSkip1 := "github.com/donyori/gogo/errors.TestAutoWrapCustom.func1: "
	wantMsgPrefixPrependFullPkgName := "github.com/donyori/gogo/errors: "
	wantMsgPrefixPrependSimpleFuncNameSkip0 := "TestAutoWrapCustom.func1.1: "
	wantMsgPrefixPrependSimpleFuncNameSkip1 := "TestAutoWrapCustom.func1: "
	wantMsgPrefixPrependSimplePkgName := "errors: "
	// In the above wantXxx, ".func1" is the anonymous function passed to t.Run;
	// ".func1.1" is the anonymous inner function that calls function AutoWrapCustom.

	var testCases []struct {
		err     error
		ms      ErrorMessageStrategy
		skip    int
		excl    ErrorReadOnlySet
		equal   bool
		wantMsg string
	}
	for _, err := range []error{nil, io.EOF, err0, err1, err2, stderrors.New("")} {
		for ms := ErrorMessageStrategy(-1); ms <= PrependSimplePkgName+1; ms++ {
			for skip := 0; skip <= 1; skip++ {
				for _, excl := range []ErrorReadOnlySet{nil, excl} {
					var wantMsg string
					if err != nil {
						if excl == nil || !excl.Contain(err) {
							switch ms {
							case OriginalMsg:
								// The prefix is "". Do nothing here.
							case PrependFullPkgName:
								wantMsg = wantMsgPrefixPrependFullPkgName
							case PrependSimpleFuncName:
								if skip == 0 {
									wantMsg = wantMsgPrefixPrependSimpleFuncNameSkip0
								} else {
									wantMsg = wantMsgPrefixPrependSimpleFuncNameSkip1
								}
							case PrependSimplePkgName:
								wantMsg = wantMsgPrefixPrependSimplePkgName
							default:
								if skip == 0 {
									wantMsg = wantMsgPrefixPrependFullFuncNameSkip0
								} else {
									wantMsg = wantMsgPrefixPrependFullFuncNameSkip1
								}
							}
							var errMsg string
							if err == err1 || err == err2 {
								errMsg = err0.Error()
							} else {
								errMsg = err.Error()
							}
							if errMsg != "" {
								wantMsg += errMsg
							} else {
								wantMsg += "(no error message)"
							}
						} else {
							wantMsg = err.Error()
						}
					}

					testCases = append(testCases, struct {
						err     error
						ms      ErrorMessageStrategy
						skip    int
						excl    ErrorReadOnlySet
						equal   bool
						wantMsg string
					}{
						err,
						ms,
						skip,
						excl,
						err == nil || excl != nil && excl.Contain(err),
						wantMsg,
					})
				}
			}
		}
	}

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case %d?err=%q&ms=%s(%[3]d)&skip=%d&hasExcl=%t", i, tc.err, tc.ms, tc.skip, tc.excl != nil),
			func(t *testing.T) {
				func() { // Use an inner function to test the "skip".
					got := AutoWrapCustom(tc.err, tc.ms, tc.skip, tc.excl)
					if (got == tc.err) != tc.equal {
						if tc.equal {
							t.Errorf("got %q; != tc.err", got)
						} else {
							t.Errorf("got %q; == tc.err", got)
						}
					}
					if tc.err == nil || got == nil || tc.equal {
						return
					}
					if gotMsg := got.Error(); gotMsg != tc.wantMsg {
						t.Errorf("got msg %q; want %q", gotMsg, tc.wantMsg)
					}
					if gotUnwrap := stderrors.Unwrap(got); gotUnwrap != tc.err {
						t.Errorf("got unwrap %q; != tc.err", gotUnwrap)
					}
				}()
			},
		)
	}
}
