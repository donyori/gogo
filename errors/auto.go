// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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

// defaultMessageStrategy is the default error message strategy,
// used by functions AutoMsg, AutoNew, and AutoWrap.
var defaultMessageStrategy = PrefixFullPkgName

// DefaultMessageStrategy returns the default error message strategy
// used by functions AutoMsg, AutoNew, and AutoWrap.
func DefaultMessageStrategy() ErrorMessageStrategy {
	return defaultMessageStrategy
}

// SetDefaultMessageStrategy sets the default error message strategy
// used by functions AutoMsg, AutoNew, and AutoWrap.
//
// It panics if ms is invalid.
func SetDefaultMessageStrategy(ms ErrorMessageStrategy) {
	ms.MustValid()
	defaultMessageStrategy = ms
	defaultAutoWrapper.SetMessageStrategy(ms)
}

// AutoMsg generates an error message using msg and
// the default error message strategy.
// It automatically adds the package or function name of the caller before msg.
// Caller can specify its behavior by setting the default error message
// strategy.
//
// If msg is empty, it will use "(no error message)" instead.
func AutoMsg(msg string) string {
	return AutoMsgWithStrategy(msg, defaultMessageStrategy, 1)
}

// AutoMsgWithStrategy generates an error message using msg and ms.
// It automatically adds the package or function name of the caller before msg.
// Caller can specify its behavior by ms.
//
// skip is the number of stack frames to ascend,
// with 0 identifying the caller of AutoMsgWithStrategy.
//
// If msg is empty, it will use "(no error message)" instead.
// If ms is invalid, it will use the default error message strategy instead.
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
		return msg
	}
	return prefix + ": " + msg
}

// Wrapper is a device to wrap an error as another error with method Unwrap.
// It is permitted to return the error itself directly by adding the error to
// its exclusion list.
type Wrapper interface {
	// Exclusions returns its exclusion list.
	//
	// Errors in the exclusion list will not be wrapped by this wrapper.
	Exclusions() []error

	// Exclude adds errors in exclusions to the exclusion list of this wrapper.
	//
	// Errors in the exclusion list will not be wrapped by this wrapper.
	Exclude(exclusions ...error)

	// Wrap wraps err as a new error,
	// or returns err itself directly if err is nil or in the exclusion list.
	Wrap(err error) error
}

// AutoWrapper is a wrapper to automatically adds the package or function name
// of the caller before the error message of the contained error.
// Caller can specify its behavior by setting the error message strategy.
//
// The implementations of AutoWrapper should guarantee that errors not in
// the exclusion list will be wrapped as new AutoError.
type AutoWrapper interface {
	Wrapper

	// MessageStrategy returns its error message strategy.
	MessageStrategy() ErrorMessageStrategy

	// SetMessageStrategy sets its error message strategy to ms.
	//
	// It panics if ms is invalid.
	SetMessageStrategy(ms ErrorMessageStrategy)

	// WrapSkip wraps err as a new error,
	// or returns err itself directly if err is nil or in the exclusion list.
	//
	// skip is the number of stack frames to ascend,
	// with 0 identifying the caller of WrapSkip.
	WrapSkip(err error, skip int) error
}

// autoWrapper is an implementation of interface AutoWrapper.
type autoWrapper struct {
	// Error message strategy used by this wrapper.
	msgStrategy ErrorMessageStrategy

	// Exclusion list of this wrapper.
	// The errors are sorted by their error messages.
	exclusions map[string][]error
}

// NewAutoWrapper creates an AutoWrapper with ms and exclusions.
//
// It panics if ms is invalid.
func NewAutoWrapper(ms ErrorMessageStrategy, exclusions ...error) AutoWrapper {
	aw := new(autoWrapper)
	aw.SetMessageStrategy(ms)
	aw.Exclude(exclusions...)
	return aw
}

// Exclusions returns its exclusion list.
//
// Errors in the exclusion list will not be wrapped by this wrapper.
func (aw *autoWrapper) Exclusions() []error {
	if aw == nil || len(aw.exclusions) == 0 {
		return nil
	}
	list := make([]error, 0, len(aw.exclusions))
	for _, item := range aw.exclusions {
		list = append(list, item...)
	}
	return list
}

// Exclude adds errors in exclusions to the exclusion list of this wrapper.
//
// Errors in the exclusion list will not be wrapped by this wrapper.
func (aw *autoWrapper) Exclude(exclusions ...error) {
	for _, err := range exclusions {
		if err == nil {
			continue
		}
		errMsg := err.Error()
		if aw.exclusions == nil {
			aw.exclusions = make(map[string][]error)
			aw.exclusions[errMsg] = []error{err}
			continue
		}
		list, skip := aw.exclusions[errMsg], false
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

// Wrap wraps err as a new error,
// or returns err itself directly if err is nil or in the exclusion list.
func (aw *autoWrapper) Wrap(err error) error {
	return aw.WrapSkip(err, 1)
}

// MessageStrategy returns its error message strategy.
func (aw *autoWrapper) MessageStrategy() ErrorMessageStrategy {
	if aw == nil {
		return -1
	}
	return aw.msgStrategy
}

// SetMessageStrategy sets its error message strategy to ms.
//
// It panics if ms is invalid.
func (aw *autoWrapper) SetMessageStrategy(ms ErrorMessageStrategy) {
	ms.MustValid()
	aw.msgStrategy = ms
}

// WrapSkip wraps err as a new error,
// or returns err itself directly if err is nil or in the exclusion list.
//
// skip is the number of stack frames to ascend,
// with 0 identifying the caller of WrapSkip.
func (aw *autoWrapper) WrapSkip(err error, skip int) error {
	if err == nil {
		return nil
	}
	if aw.exclusions != nil {
		list := aw.exclusions[err.Error()]
		for _, item := range list {
			if Is(err, item) {
				return err
			}
		}
	}
	ae := &autoError{err: err}
	// Find the first error that isn't an AutoError along the error chain.
	var tmpAe AutoError
	tmpErr := err
	for As(tmpErr, &tmpAe) {
		tmpErr = Unwrap(tmpErr)
		if tmpErr == nil {
			break
		}
		err = tmpErr
	}
	ae.msg = AutoMsgWithStrategy(err.Error(), aw.msgStrategy, skip+1)
	return ae
}

// defaultAutoWrapper is the auto wrapper used by function AutoWrap.
var defaultAutoWrapper = NewAutoWrapper(defaultMessageStrategy, io.EOF)

// AutoWrapExclusions returns the exclusion list used by function AutoWrap.
func AutoWrapExclusions() []error {
	return defaultAutoWrapper.Exclusions()
}

// AutoWrapExclude excludes errors in exclusions to
// let function AutoWrap return these errors directly.
func AutoWrapExclude(exclusions ...error) {
	defaultAutoWrapper.Exclude(exclusions...)
}

// AutoNew creates a new error using msg and the default error message strategy.
// It automatically adds the package or function name of the caller before msg.
// Caller can specify its behavior by setting the default error message
// strategy via SetDefaultMessageStrategy.
func AutoNew(msg string) AutoError {
	return defaultAutoWrapper.WrapSkip(New(msg), 1).(AutoError)
}

// AutoNewWithStrategy creates a new error using msg and ms.
// It automatically adds the package or function name of the caller before msg.
// Caller can specify its behavior by ms.
//
// skip is the number of stack frames to ascend,
// with 0 identifying the caller of AutoNewWithStrategy.
//
// If ms is invalid, it will use the default error message strategy instead.
func AutoNewWithStrategy(msg string, ms ErrorMessageStrategy, skip int) AutoError {
	if !ms.Valid() {
		ms = defaultMessageStrategy
	}
	wrapper := NewAutoWrapper(ms)
	return wrapper.WrapSkip(New(msg), skip+1).(AutoError)
}

// AutoWrap wraps err with the default AutoWrapper.
// It automatically adds the package or function name of the caller
// before the error message of err.
// Caller can specify its behavior by setting the default error message
// strategy via SetDefaultMessageStrategy.
//
// It returns nil if err is nil.
func AutoWrap(err error) error {
	return defaultAutoWrapper.WrapSkip(err, 1)
}

// AutoWrapSkip wraps err with the default AutoWrapper.
// It automatically adds the package or function name of the caller
// before the error message of err.
// Caller can specify its behavior by setting the default error message
// strategy via SetDefaultMessageStrategy.
//
// skip is the number of stack frames to ascend,
// with 0 identifying the caller of AutoWrapSkip.
//
// It returns nil if err is nil.
func AutoWrapSkip(err error, skip int) error {
	return defaultAutoWrapper.WrapSkip(err, skip+1)
}

// AutoError is an error created by AutoWrapper.
//
// For an AutoError, AutoWrapper will use the message of its contained error
// to generate new error message, other than using the message of itself.
type AutoError interface {
	WrappingError

	// GogoAutoMade is a dummy method with an empty body to indicate that
	// this error is created by AutoWrapper.
	GogoAutoMade()
}

// autoError is an implementation of interface AutoError.
type autoError struct {
	err error  // Contained error.
	msg string // Error message.
}

// Error reports the error message.
func (ae *autoError) Error() string {
	return ae.msg
}

// Unwrap returns the contained error.
func (ae *autoError) Unwrap() error {
	return ae.err
}

// GogoAutoMade is a dummy method with an empty body to indicate that
// this error is created by AutoWrapper.
func (ae *autoError) GogoAutoMade() {}

// String returns the error message,
// which performs the same as the method Error.
func (ae *autoError) String() string {
	return ae.msg
}
