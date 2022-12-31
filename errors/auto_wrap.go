// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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

import "io"

// defaultExclusionSet is an exclusion set used by
// functions AutoWrap and AutoWrapSkip.
// It contains one error: io.EOF.
var defaultExclusionSet = NewErrorReadOnlySetIs(io.EOF)

// AutoWrap wraps err by prepending the full function name
// (i.e., the package path-qualified function name)
// of its caller to the error message of err.
//
// In particular, if err is generated by AutoWrap, AutoWrapSkip,
// or AutoWrapCustom, AutoWrap will find the first error that is not
// generated by these functions along the Unwrap error chain and
// use its error message instead.
//
// It returns err itself if err is nil or io.EOF
// (i.e., err == nil || errors.Is(err, io.EOF)).
//
// If the target error message is empty,
// it will use "<no error message>" instead.
func AutoWrap(err error) error {
	if err == nil { // nil error is the most common case, so add a short path here
		return nil
	}
	return AutoWrapCustom(err, -1, 1, defaultExclusionSet)
}

// AutoWrapSkip wraps err by prepending the full function name
// (i.e., the package path-qualified function name)
// of the caller to the error message of err.
//
// In particular, if err is generated by AutoWrap, AutoWrapSkip,
// or AutoWrapCustom, AutoWrapSkip will find the first error that is not
// generated by these functions along the Unwrap error chain and
// use its error message instead.
//
// skip is the number of stack frames to ascend,
// with 0 identifying the caller of AutoWrapSkip.
// For example, if skip is 1, instead of the caller of AutoWrapSkip,
// the information of that caller's caller will be used.
//
// It returns err itself if err is nil or io.EOF
// (i.e., err == nil || errors.Is(err, io.EOF)).
//
// If the target error message is empty,
// it will use "<no error message>" instead.
func AutoWrapSkip(err error, skip int) error {
	if err == nil { // nil error is the most common case, so add a short path here
		return nil
	}
	return AutoWrapCustom(err, -1, skip+1, defaultExclusionSet)
}

// AutoWrapCustom wraps err by prepending the function or package name
// of the caller to the error message of err.
// Caller can specify its behavior by ms.
//
// In particular, if err is generated by AutoWrap, AutoWrapSkip,
// or AutoWrapCustom, AutoWrapCustom will find the first error that is not
// generated by these functions along the Unwrap error chain and
// use its error message instead.
//
// skip is the number of stack frames to ascend,
// with 0 identifying the caller of AutoWrapCustom.
// For example, if skip is 1, instead of the caller of AutoWrapCustom,
// the information of that caller's caller will be used.
//
// exclusions are a set of errors.
// If err is nil or in exclusions, AutoWrapCustom returns err itself.
//
// If the target error message is empty,
// it will use "<no error message>" instead.
// If ms is invalid, it will use PrependFullFuncName instead.
func AutoWrapCustom(err error, ms ErrorMessageStrategy, skip int, exclusions ErrorReadOnlySet) error {
	if err == nil || exclusions != nil && exclusions.Contains(err) {
		return err
	}
	unwrapped, _ := UnwrapAllAutoWrappedErrors(err)
	return &autoWrappedError{
		err: err,
		msg: AutoMsgCustom(unwrapped.Error(), ms, skip+1),
	}
}

// IsAutoWrappedError reports whether err is generated by
// AutoWrap, AutoWrapSkip, or AutoWrapCustom.
func IsAutoWrappedError(err error) bool {
	// The type testing should not go along the Unwrap error chain,
	// so type assertion is used here instead of errors.As.
	_, ok := err.(*autoWrappedError)
	return ok
}

// UnwrapAutoWrappedError unwraps err and returns the result and true
// if err is generated by AutoWrap, AutoWrapSkip, or AutoWrapCustom.
// Otherwise, UnwrapAutoWrappedError returns err itself and false.
func UnwrapAutoWrappedError(err error) (error, bool) {
	// The type testing should not go along the Unwrap error chain,
	// so type assertion is used here instead of errors.As.
	awe, ok := err.(*autoWrappedError)
	if ok {
		err = awe.err
	}
	return err, ok
}

// UnwrapAllAutoWrappedErrors repeatedly unwraps err until the result is not
// an error generated by AutoWrap, AutoWrapSkip, or AutoWrapCustom,
// and then returns the result and true if err is generated by these functions.
// If err is not generated by these functions,
// UnwrapAllAutoWrappedErrors returns err itself and false.
func UnwrapAllAutoWrappedErrors(err error) (error, bool) {
	var isUnwrapped bool
	// Each type testing should not go along the Unwrap error chain,
	// so type assertion is used here instead of errors.As.
	for awe, ok := err.(*autoWrappedError); ok; awe, ok = err.(*autoWrappedError) {
		err, isUnwrapped = awe.err, true
	}
	return err, isUnwrapped
}

// autoWrappedError is the error generated by
// AutoWrap, AutoWrapSkip, and AutoWrapCustom.
//
// It consists of the wrapped error and an error message.
type autoWrappedError struct {
	err error // the wrapped error, must be non-nil
	msg string
}

// Error reports the error message.
func (awe *autoWrappedError) Error() string {
	return awe.msg
}

// Unwrap returns the wrapped error.
func (awe *autoWrappedError) Unwrap() error {
	return awe.err
}
