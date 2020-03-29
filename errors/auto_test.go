// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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
	"io"
	"testing"
)

func TestAutoMsgWithStrategy_1(t *testing.T) {
	dms := DefaultMessageStrategy()
	defer SetDefaultMessageStrategy(dms)
	SetDefaultMessageStrategy(PrefixFullFuncName)
	cases := []struct {
		msg    string
		ms     ErrorMessageStrategy
		wanted string
	}{
		{"", -1, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_1: (no error message)"},
		{"", OriginalMsg, "(no error message)"},
		{"", PrefixFullPkgName, "github.com/donyori/gogo/errors: (no error message)"},
		{"", PrefixSimplePkgName, "errors: (no error message)"},
		{"", PrefixFullFuncName, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_1: (no error message)"},
		{"", PrefixSimpleFuncName, "TestAutoMsgWithStrategy_1: (no error message)"},
		{"", PrefixSimpleFuncName + 1, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_1: (no error message)"},
		{"some error", -1, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_1: some error"},
		{"some error", OriginalMsg, "some error"},
		{"some error", PrefixFullPkgName, "github.com/donyori/gogo/errors: some error"},
		{"some error", PrefixSimplePkgName, "errors: some error"},
		{"some error", PrefixFullFuncName, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_1: some error"},
		{"some error", PrefixSimpleFuncName, "TestAutoMsgWithStrategy_1: some error"},
		{"some error", PrefixSimpleFuncName + 1, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_1: some error"},
	}
	for _, c := range cases {
		s := AutoMsgWithStrategy(c.msg, c.ms, 0)
		if s != c.wanted {
			t.Errorf("AutoMsgWithStrategy: %q != %q, msg: %q, ms: %v, skip: 0.", s, c.wanted, c.msg, c.ms)
		}
	}
}

func TestAutoMsgWithStrategy_2(t *testing.T) {
	dms := DefaultMessageStrategy()
	defer SetDefaultMessageStrategy(dms)
	SetDefaultMessageStrategy(PrefixFullFuncName)
	cases := []struct {
		msg    string
		ms     ErrorMessageStrategy
		skip   int
		wanted string
	}{
		{"", -1, 0, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_2.func1: (no error message)"},
		{"", -1, 1, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_2: (no error message)"},
		{"", OriginalMsg, 0, "(no error message)"},
		{"", OriginalMsg, 1, "(no error message)"},
		{"", PrefixFullPkgName, 0, "github.com/donyori/gogo/errors: (no error message)"},
		{"", PrefixFullPkgName, 1, "github.com/donyori/gogo/errors: (no error message)"},
		{"", PrefixSimplePkgName, 0, "errors: (no error message)"},
		{"", PrefixSimplePkgName, 1, "errors: (no error message)"},
		{"", PrefixFullFuncName, 0, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_2.func1: (no error message)"},
		{"", PrefixFullFuncName, 1, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_2: (no error message)"},
		{"", PrefixSimpleFuncName, 0, "TestAutoMsgWithStrategy_2.func1: (no error message)"},
		{"", PrefixSimpleFuncName, 1, "TestAutoMsgWithStrategy_2: (no error message)"},
		{"", PrefixSimpleFuncName + 1, 0, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_2.func1: (no error message)"},
		{"", PrefixSimpleFuncName + 1, 1, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_2: (no error message)"},
		{"some error", -1, 0, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_2.func1: some error"},
		{"some error", -1, 1, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_2: some error"},
		{"some error", OriginalMsg, 0, "some error"},
		{"some error", OriginalMsg, 1, "some error"},
		{"some error", PrefixFullPkgName, 0, "github.com/donyori/gogo/errors: some error"},
		{"some error", PrefixFullPkgName, 1, "github.com/donyori/gogo/errors: some error"},
		{"some error", PrefixSimplePkgName, 0, "errors: some error"},
		{"some error", PrefixSimplePkgName, 1, "errors: some error"},
		{"some error", PrefixFullFuncName, 0, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_2.func1: some error"},
		{"some error", PrefixFullFuncName, 1, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_2: some error"},
		{"some error", PrefixSimpleFuncName, 0, "TestAutoMsgWithStrategy_2.func1: some error"},
		{"some error", PrefixSimpleFuncName, 1, "TestAutoMsgWithStrategy_2: some error"},
		{"some error", PrefixSimpleFuncName + 1, 0, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_2.func1: some error"},
		{"some error", PrefixSimpleFuncName + 1, 1, "github.com/donyori/gogo/errors.TestAutoMsgWithStrategy_2: some error"},
	}
	func() {
		for _, c := range cases {
			s := AutoMsgWithStrategy(c.msg, c.ms, c.skip)
			if s != c.wanted {
				t.Errorf("AutoMsgWithStrategy: %q != %q, msg: %q, ms: %v, skip: %d.", s, c.wanted, c.msg, c.ms, c.skip)
			}
		}
	}()
}

func TestAutoMsg(t *testing.T) {
	dms := DefaultMessageStrategy()
	defer SetDefaultMessageStrategy(dms)
	cases := []struct {
		msg    string
		ms     ErrorMessageStrategy
		wanted string
	}{
		{"", OriginalMsg, "(no error message)"},
		{"", PrefixFullPkgName, "github.com/donyori/gogo/errors: (no error message)"},
		{"", PrefixSimplePkgName, "errors: (no error message)"},
		{"", PrefixFullFuncName, "github.com/donyori/gogo/errors.TestAutoMsg: (no error message)"},
		{"", PrefixSimpleFuncName, "TestAutoMsg: (no error message)"},
		{"some error", OriginalMsg, "some error"},
		{"some error", PrefixFullPkgName, "github.com/donyori/gogo/errors: some error"},
		{"some error", PrefixSimplePkgName, "errors: some error"},
		{"some error", PrefixFullFuncName, "github.com/donyori/gogo/errors.TestAutoMsg: some error"},
		{"some error", PrefixSimpleFuncName, "TestAutoMsg: some error"},
	}
	for _, c := range cases {
		SetDefaultMessageStrategy(c.ms)
		s := AutoMsg(c.msg)
		if s != c.wanted {
			t.Errorf("AutoMsg: %q != %q, msg: %q, ms: %v.", s, c.wanted, c.msg, c.ms)
		}
	}
}

func TestAutoNew(t *testing.T) {
	dms := DefaultMessageStrategy()
	defer SetDefaultMessageStrategy(dms)
	cases := []struct {
		msg    string
		ms     ErrorMessageStrategy
		wanted string
	}{
		{"", OriginalMsg, "(no error message)"},
		{"", PrefixFullPkgName, "github.com/donyori/gogo/errors: (no error message)"},
		{"", PrefixSimplePkgName, "errors: (no error message)"},
		{"", PrefixFullFuncName, "github.com/donyori/gogo/errors.TestAutoNew: (no error message)"},
		{"", PrefixSimpleFuncName, "TestAutoNew: (no error message)"},
		{"some error", OriginalMsg, "some error"},
		{"some error", PrefixFullPkgName, "github.com/donyori/gogo/errors: some error"},
		{"some error", PrefixSimplePkgName, "errors: some error"},
		{"some error", PrefixFullFuncName, "github.com/donyori/gogo/errors.TestAutoNew: some error"},
		{"some error", PrefixSimpleFuncName, "TestAutoNew: some error"},
	}
	for _, c := range cases {
		SetDefaultMessageStrategy(c.ms)
		err := AutoNew(c.msg)
		if s := err.Error(); s != c.wanted {
			t.Errorf("AutoNew: %q != %q, msg: %q, ms: %v.", s, c.wanted, c.msg, c.ms)
		}
		tmpAme := new(autoMadeError)
		if !As(err, &tmpAme) {
			t.Errorf("The error returned by AutoNew is not a *autoMadeError, msg: %q, ms: %v.", c.msg, c.ms)
		}
	}
}

func TestAutoNewWithStrategy(t *testing.T) {
	dms := DefaultMessageStrategy()
	defer SetDefaultMessageStrategy(dms)
	SetDefaultMessageStrategy(PrefixFullFuncName)
	cases := []struct {
		msg    string
		ms     ErrorMessageStrategy
		skip   int
		wanted string
	}{
		{"", -1, 0, "github.com/donyori/gogo/errors.TestAutoNewWithStrategy.func1: (no error message)"},
		{"", -1, 1, "github.com/donyori/gogo/errors.TestAutoNewWithStrategy: (no error message)"},
		{"", OriginalMsg, 0, "(no error message)"},
		{"", OriginalMsg, 1, "(no error message)"},
		{"", PrefixFullPkgName, 0, "github.com/donyori/gogo/errors: (no error message)"},
		{"", PrefixFullPkgName, 1, "github.com/donyori/gogo/errors: (no error message)"},
		{"", PrefixSimplePkgName, 0, "errors: (no error message)"},
		{"", PrefixSimplePkgName, 1, "errors: (no error message)"},
		{"", PrefixFullFuncName, 0, "github.com/donyori/gogo/errors.TestAutoNewWithStrategy.func1: (no error message)"},
		{"", PrefixFullFuncName, 1, "github.com/donyori/gogo/errors.TestAutoNewWithStrategy: (no error message)"},
		{"", PrefixSimpleFuncName, 0, "TestAutoNewWithStrategy.func1: (no error message)"},
		{"", PrefixSimpleFuncName, 1, "TestAutoNewWithStrategy: (no error message)"},
		{"", PrefixSimpleFuncName + 1, 0, "github.com/donyori/gogo/errors.TestAutoNewWithStrategy.func1: (no error message)"},
		{"", PrefixSimpleFuncName + 1, 1, "github.com/donyori/gogo/errors.TestAutoNewWithStrategy: (no error message)"},
		{"some error", -1, 0, "github.com/donyori/gogo/errors.TestAutoNewWithStrategy.func1: some error"},
		{"some error", -1, 1, "github.com/donyori/gogo/errors.TestAutoNewWithStrategy: some error"},
		{"some error", OriginalMsg, 0, "some error"},
		{"some error", OriginalMsg, 1, "some error"},
		{"some error", PrefixFullPkgName, 0, "github.com/donyori/gogo/errors: some error"},
		{"some error", PrefixFullPkgName, 1, "github.com/donyori/gogo/errors: some error"},
		{"some error", PrefixSimplePkgName, 0, "errors: some error"},
		{"some error", PrefixSimplePkgName, 1, "errors: some error"},
		{"some error", PrefixFullFuncName, 0, "github.com/donyori/gogo/errors.TestAutoNewWithStrategy.func1: some error"},
		{"some error", PrefixFullFuncName, 1, "github.com/donyori/gogo/errors.TestAutoNewWithStrategy: some error"},
		{"some error", PrefixSimpleFuncName, 0, "TestAutoNewWithStrategy.func1: some error"},
		{"some error", PrefixSimpleFuncName, 1, "TestAutoNewWithStrategy: some error"},
		{"some error", PrefixSimpleFuncName + 1, 0, "github.com/donyori/gogo/errors.TestAutoNewWithStrategy.func1: some error"},
		{"some error", PrefixSimpleFuncName + 1, 1, "github.com/donyori/gogo/errors.TestAutoNewWithStrategy: some error"},
	}
	func() {
		for _, c := range cases {
			err := AutoNewWithStrategy(c.msg, c.ms, c.skip)
			if s := err.Error(); s != c.wanted {
				t.Errorf("AutoNewWithStrategy: %q != %q, msg: %q, ms: %v, skip: %d.", s, c.wanted, c.msg, c.ms, c.skip)
			}
			tmpAme := new(autoMadeError)
			if !As(err, &tmpAme) {
				t.Errorf("The error returned by AutoNewWithStrategy is not a *autoMadeError, msg: %q, ms: %v.", c.msg, c.ms)
			}
		}
	}()
}

func TestAutoWrap(t *testing.T) {
	dms := DefaultMessageStrategy()
	defer SetDefaultMessageStrategy(dms)
	cases := []struct {
		msg    string
		ms     ErrorMessageStrategy
		wanted string
	}{
		{"", OriginalMsg, "(no error message)"},
		{"", PrefixFullPkgName, "github.com/donyori/gogo/errors: (no error message)"},
		{"", PrefixSimplePkgName, "errors: (no error message)"},
		{"", PrefixFullFuncName, "github.com/donyori/gogo/errors.TestAutoWrap: (no error message)"},
		{"", PrefixSimpleFuncName, "TestAutoWrap: (no error message)"},
		{"some error", OriginalMsg, "some error"},
		{"some error", PrefixFullPkgName, "github.com/donyori/gogo/errors: some error"},
		{"some error", PrefixSimplePkgName, "errors: some error"},
		{"some error", PrefixFullFuncName, "github.com/donyori/gogo/errors.TestAutoWrap: some error"},
		{"some error", PrefixSimpleFuncName, "TestAutoWrap: some error"},
	}
	for _, c := range cases {
		SetDefaultMessageStrategy(c.ms)
		var err error = AutoNew(c.msg)
		err = AutoWrap(err)
		if s := err.Error(); s != c.wanted {
			t.Errorf("AutoWrap: %q != %q, msg: %q, ms: %v.", s, c.wanted, c.msg, c.ms)
		}
		tmpAme := new(autoMadeError)
		if !As(err, &tmpAme) {
			t.Errorf("The error returned by AutoWrap is not a *autoMadeError, msg: %q, ms: %v.", c.msg, c.ms)
		}
	}
	// Test whether io.EOF is excluded.
	for ms := OriginalMsg; ms <= PrefixSimpleFuncName; ms++ {
		SetDefaultMessageStrategy(ms)
		err := AutoWrap(io.EOF)
		if err != io.EOF { // Don't use Is(err, io.EOF)! err should be io.EOF itself.
			t.Error("AutoWrap wraps io.EOF.")
		}
	}
}

func TestAutoWrapSkip(t *testing.T) {
	dms := DefaultMessageStrategy()
	defer SetDefaultMessageStrategy(dms)
	SetDefaultMessageStrategy(PrefixFullFuncName)
	cases := []struct {
		msg    string
		ms     ErrorMessageStrategy
		skip   int
		wanted string
	}{
		{"", OriginalMsg, 0, "(no error message)"},
		{"", OriginalMsg, 1, "(no error message)"},
		{"", PrefixFullPkgName, 0, "github.com/donyori/gogo/errors: (no error message)"},
		{"", PrefixFullPkgName, 1, "github.com/donyori/gogo/errors: (no error message)"},
		{"", PrefixSimplePkgName, 0, "errors: (no error message)"},
		{"", PrefixSimplePkgName, 1, "errors: (no error message)"},
		{"", PrefixFullFuncName, 0, "github.com/donyori/gogo/errors.TestAutoWrapSkip.func1: (no error message)"},
		{"", PrefixFullFuncName, 1, "github.com/donyori/gogo/errors.TestAutoWrapSkip: (no error message)"},
		{"", PrefixSimpleFuncName, 0, "TestAutoWrapSkip.func1: (no error message)"},
		{"", PrefixSimpleFuncName, 1, "TestAutoWrapSkip: (no error message)"},
		{"some error", OriginalMsg, 0, "some error"},
		{"some error", OriginalMsg, 1, "some error"},
		{"some error", PrefixFullPkgName, 0, "github.com/donyori/gogo/errors: some error"},
		{"some error", PrefixFullPkgName, 1, "github.com/donyori/gogo/errors: some error"},
		{"some error", PrefixSimplePkgName, 0, "errors: some error"},
		{"some error", PrefixSimplePkgName, 1, "errors: some error"},
		{"some error", PrefixFullFuncName, 0, "github.com/donyori/gogo/errors.TestAutoWrapSkip.func1: some error"},
		{"some error", PrefixFullFuncName, 1, "github.com/donyori/gogo/errors.TestAutoWrapSkip: some error"},
		{"some error", PrefixSimpleFuncName, 0, "TestAutoWrapSkip.func1: some error"},
		{"some error", PrefixSimpleFuncName, 1, "TestAutoWrapSkip: some error"},
	}
	func() {
		for _, c := range cases {
			SetDefaultMessageStrategy(c.ms)
			var err error = AutoNew(c.msg)
			err = AutoWrapSkip(err, c.skip)
			if s := err.Error(); s != c.wanted {
				t.Errorf("AutoWrapSkip: %q != %q, msg: %q, ms: %v, skip: %d.", s, c.wanted, c.msg, c.ms, c.skip)
			}
			tmpAme := new(autoMadeError)
			if !As(err, &tmpAme) {
				t.Errorf("The error returned by AutoWrapSkip is not a *autoMadeError, msg: %q, ms: %v, skip: %d.", c.msg, c.ms, c.skip)
			}
		}
		// Test whether io.EOF is excluded.
		for ms := OriginalMsg; ms <= PrefixSimpleFuncName; ms++ {
			SetDefaultMessageStrategy(ms)
			for skip := 0; skip <= 1; skip++ {
				err := AutoWrapSkip(io.EOF, skip)
				if err != io.EOF { // Don't use Is(err, io.EOF)! err should be io.EOF itself.
					t.Errorf("AutoWrapSkip wraps io.EOF, skip: %d.", skip)
				}
			}
		}
	}()
}
