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
	"strings"

	"github.com/donyori/gogo/runtime"
)

// Default error message strategy,
// used by functions AutoMsg, AutoNew, and AutoWrap.
var defaultMessageStrategy = PrefixFullPkgName

// Return the default error message strategy used by functions
// AutoMsg, AutoNew, and AutoWrap.
func DefaultMessageStrategy() ErrorMessageStrategy {
	return defaultMessageStrategy
}

// Set the default error message strategy used by functions
// AutoMsg, AutoNew, and AutoWrap.
func SetDefaultMessageStrategy(ms ErrorMessageStrategy) {
	ms.MustValid()
	defaultMessageStrategy = ms
}

// Generate error message using msg and the default message strategy.
// It automatically adds the package or function name of the caller before msg.
// Caller can specify its action by setting the default message strategy.
func AutoMsg(msg string) string {
	return AutoMsgWithStrategy(msg, defaultMessageStrategy, 1)
}

// Generate error message using msg and ms. skip is the number of
// stack frames to ascend, with 0 identifying the caller of AutoMsgWithStrategy.
// It automatically adds the package or function name of the caller before msg.
// Caller can specify its action by ms.
func AutoMsgWithStrategy(msg string, ms ErrorMessageStrategy, skip int) string {
	if !ms.Valid() {
		ms = defaultMessageStrategy
	}
	if msg == "" {
		msg = "(no error message)"
	}
	if ms == OriginalMsg {
		return msg
	}
	frame, ok := runtime.CallerFrame(skip + 1)
	if !ok {
		return "(retrieving caller failed) error: " + msg
	}
	var prefix string
	switch ms {
	case PrefixFullPkgName:
		prefix, _, ok = runtime.FramePkgFunc(frame)
		if !ok {
			return "(retrieving caller failed) error: " + msg
		}
	case PrefixSimplePkgName:
		prefix, _, ok = runtime.FramePkgFunc(frame)
		if !ok {
			return "(retrieving caller failed) error: " + msg
		}
		prefix = prefix[strings.LastIndex(prefix, "/")+1:]
	case PrefixFullFuncName:
		prefix = frame.Function
	case PrefixSimpleFuncName:
		_, prefix, ok = runtime.FramePkgFunc(frame)
		if !ok {
			return "(retrieving caller failed) error: " + msg
		}
	}
	if prefix == "" {
		prefix = "error"
	}
	return prefix + ": " + msg
}

// Create a new error using msg and the default message strategy.
// It automatically adds the package or function name of the caller before msg.
// Caller can specify its action by setting the default message strategy.
func AutoNew(msg string) error {
	return AutoNewWithStrategy(msg, defaultMessageStrategy, 1)
}

// Create a new error using msg and ms. If ms is OriginalMsg, it performs
// the same as the standard function errors.New(msg). skip is the number of
// stack frames to ascend, with 0 identifying the caller of AutoNewWithStrategy.
// It automatically adds the package or function name of the caller before msg.
// Caller can specify its action by ms.
func AutoNewWithStrategy(msg string, ms ErrorMessageStrategy, skip int) error {
	return New(AutoMsgWithStrategy(msg, ms, skip+1))
}

// Wrap err with the default message strategy. It automatically adds
// the package or function name of the caller before the error message of err.
// Caller can specify its action by setting the default message strategy.
func AutoWrap(err error) WrappingError {
	return AutoWrapWithStrategy(err, defaultMessageStrategy, 1)
}

// Wrap err with ms. skip is the number of stack frames to ascend,
// with 0 identifying the caller of AutoWrapWithStrategy.
// It automatically adds the package or function name of the caller before the
// error message of err. Caller can specify its action by ms.
func AutoWrapWithStrategy(err error, ms ErrorMessageStrategy, skip int) WrappingError {
	ms.MustValid()
	if err == nil {
		return nil
	}
	awe := &autoWrappingError{err: err}
	// Find the first error that isn't an autoWrappingError along the error chain.
	tmpAwe := new(autoWrappingError)
	tmpErr := err
	for As(tmpErr, tmpAwe) {
		tmpErr = Unwrap(tmpErr)
		if tmpErr == nil {
			break
		}
		err = tmpErr
	}
	awe.msg = AutoMsgWithStrategy(err.Error(), ms, skip+1)
	return awe
}

type autoWrappingError struct {
	err error
	msg string
}

func (awe *autoWrappingError) Error() string {
	if awe == nil {
		return "no errors"
	}
	return awe.msg
}

func (awe *autoWrappingError) Unwrap() error {
	if awe == nil {
		return nil
	}
	return awe.err
}
