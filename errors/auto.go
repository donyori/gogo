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
	defaultAutoWrapper.SetMessageStrategy(ms)
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

// An error wrapper that automatically adds the package or function name of
// the caller before the error message of the wrapped error. Caller can specify
// its action by setting the error message strategy.
type AutoWrapper struct {
	msgStrategy ErrorMessageStrategy
	exclusions  map[string][]error
}

// Create a AutoWrapper with ms and exclusions. It panics if ms is invalid.
func NewAutoWrapper(ms ErrorMessageStrategy, exclusions ...error) *AutoWrapper {
	aw := new(AutoWrapper)
	aw.SetMessageStrategy(ms)
	aw.Exclude(exclusions...)
	return aw
}

// Return its error message strategy.
func (aw *AutoWrapper) MessageStrategy() ErrorMessageStrategy {
	if aw == nil {
		return -1
	}
	return aw.msgStrategy
}

// Set its error message strategy to ms. It panics if ms is invalid.
func (aw *AutoWrapper) SetMessageStrategy(ms ErrorMessageStrategy) {
	ms.MustValid()
	aw.msgStrategy = ms
}

// Return its exclusion list.
func (aw *AutoWrapper) Exclusions() []error {
	if aw == nil {
		return nil
	}
	list := make([]error, 0, len(aw.exclusions))
	for _, item := range aw.exclusions {
		list = append(list, item...)
	}
	return list
}

// Exclude errors in exclusions, i.e., for every error in exclusions,
// the wrapper won't wrap it and return itself directly.
func (aw *AutoWrapper) Exclude(exclusions ...error) {
	for _, err := range exclusions {
		if err == nil {
			continue
		}
		if aw.exclusions == nil {
			aw.exclusions = make(map[string][]error)
		}
		errMsg := err.Error()
		list := aw.exclusions[errMsg]
		skip := false
		for _, item := range list {
			if Is(err, item) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		aw.exclusions[errMsg] = append(list, err)
	}
}

// Wrap err as a new error. Or return err itself directly if err is in the
// exclusion list. skip is the number of stack frames to ascend,
// with 0 identifying the caller of WrapSkip.
func (aw *AutoWrapper) WrapSkip(err error, skip int) error {
	if err == nil {
		return nil
	}
	list := aw.exclusions[err.Error()]
	for _, item := range list {
		if Is(err, item) {
			return err
		}
	}
	awe := &autoWrappingError{err: err}
	// Find the first error that isn't an autoWrappingError along the error chain.
	tmpAwe := new(autoWrappingError)
	tmpErr := err
	for As(tmpErr, &tmpAwe) {
		tmpErr = Unwrap(tmpErr)
		if tmpErr == nil {
			break
		}
		err = tmpErr
	}
	awe.msg = AutoMsgWithStrategy(err.Error(), aw.msgStrategy, skip+1)
	return awe
}

func (aw *AutoWrapper) Wrap(err error) error {
	return aw.WrapSkip(err, 1)
}

// Auto wrapper used by function AutoWrap.
var defaultAutoWrapper = NewAutoWrapper(defaultMessageStrategy, io.EOF)

// Return the exclusion list used by function AutoWrap.
func AutoWrapExclusions() []error {
	return defaultAutoWrapper.Exclusions()
}

// Exclude errors in exclusions to let function AutoWrap
// return these errors directly.
func AutoWrapExclude(exclusions ...error) {
	defaultAutoWrapper.Exclude(exclusions...)
}

// Wrap err with the default message strategy. It automatically adds
// the package or function name of the caller before the error message of err.
// Caller can specify its action by setting the default message strategy.
// It returns nil if err is nil.
func AutoWrap(err error) error {
	return defaultAutoWrapper.WrapSkip(err, 1)
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
