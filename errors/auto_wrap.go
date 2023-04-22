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

package errors

import (
	"fmt"
	"io"
	"strings"

	"github.com/donyori/gogo/runtime"
)

// AutoWrappedError is the error generated by
// AutoWrap, AutoWrapSkip, and AutoWrapCustom.
//
// It records the information of the function that reports the error.
//
// Its method Unwrap returns the error wrapped by
// AutoWrap, AutoWrapSkip, or AutoWrapCustom.
// It is always a non-nil error.
//
// The client cannot implement this interface.
type AutoWrappedError interface {
	ErrorUnwrap

	// Root finds and returns the first error along the Unwrap error chain
	// that is not generated by AutoWrap, AutoWrapSkip, or AutoWrapCustom.
	// The result is always a non-nil error.
	Root() error

	// FullFunction returns the full function name
	// (i.e., the package path-qualified function name)
	// recorded in this error.
	FullFunction() string

	// FullPackage returns the full package name recorded in this error.
	FullPackage() string

	// SimpleFunction returns the simple function name recorded in this error.
	SimpleFunction() string

	// SimplePackage returns the simple package name recorded in this error.
	SimplePackage() string

	// MessageStrategy returns the strategy for auto-generating error messages
	// used by this error.
	MessageStrategy() ErrorMessageStrategy

	// private prevents others from implementing this interface.
	private()
}

// autoWrappedError is the only implementation of interface AutoWrappedError.
type autoWrappedError struct {
	err      error                // the wrapped error, must be non-nil
	ms       ErrorMessageStrategy // the error message strategy, must be valid
	fullFunc string               // the full function name, must be non-empty and legal
}

var _ AutoWrappedError = (*autoWrappedError)(nil)

func (awe *autoWrappedError) Error() string {
	msg := awe.Root().Error()
	if msg == "" {
		msg = "<no error message>"
	}
	if awe.ms == OriginalMsg {
		return msg
	}
	var prefix string
	switch awe.ms {
	case PrependFullFuncName:
		prefix = awe.FullFunction()
	case PrependFullPkgName:
		prefix = awe.FullPackage()
	case PrependSimpleFuncName:
		prefix = awe.SimpleFunction()
	case PrependSimplePkgName:
		prefix = awe.SimplePackage()
	default:
		// This should never happen, but will act as a safeguard for later,
		// as a default value doesn't make sense here.
		panic(AutoMsg(fmt.Sprintf(
			"error message strategy is invalid (%v), which should never happen",
			awe.ms,
		)))
	}
	if prefix == "" {
		return msg
	}
	return prefix + ": " + msg
}

func (awe *autoWrappedError) Unwrap() error {
	return awe.err
}

func (awe *autoWrappedError) Root() error {
	err := awe.err
	// Each type testing should not go along the Unwrap error tree,
	// so type assertion is used here instead of errors.As.
	for {
		wrapped, ok := err.(*autoWrappedError)
		if !ok {
			return err
		}
		err = wrapped.err
	}
}

func (awe *autoWrappedError) FullFunction() string {
	return awe.fullFunc
}

func (awe *autoWrappedError) FullPackage() string {
	return runtime.FuncPkg(awe.fullFunc)
}

func (awe *autoWrappedError) SimpleFunction() string {
	return awe.fullFunc[len(runtime.FuncPkg(awe.fullFunc))+1:]
}

func (awe *autoWrappedError) SimplePackage() string {
	pkg := runtime.FuncPkg(awe.fullFunc)
	return pkg[strings.LastIndexByte(pkg, '/')+1:]
}

func (awe *autoWrappedError) MessageStrategy() ErrorMessageStrategy {
	return awe.ms
}

func (awe *autoWrappedError) private() {}

// defaultExclusionSet is an exclusion set used by
// functions AutoWrap and AutoWrapSkip.
// It contains one error: io.EOF.
var defaultExclusionSet = NewErrorReadOnlySetIs(io.EOF)

// AutoWrap wraps err by prepending the full function name
// (i.e., the package path-qualified function name)
// of its caller to the error message of err.
//
// In particular, if err is generated by AutoWrap, AutoWrapSkip,
// or AutoWrapCustom, AutoWrap finds the first error that is not
// generated by these functions along the Unwrap error tree and
// uses its error message instead.
//
// It returns err itself if err is nil or io.EOF
// (i.e., err == nil || errors.Is(err, io.EOF)).
//
// If the target error message is empty,
// it uses "<no error message>" instead.
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
// or AutoWrapCustom, AutoWrapSkip finds the first error that is not
// generated by these functions along the Unwrap error tree and
// uses its error message instead.
//
// skip is the number of stack frames to ascend,
// with 0 identifying the caller of AutoWrapSkip.
// For example, if skip is 1, instead of the caller of AutoWrapSkip,
// the information of that caller's caller is used.
//
// It returns err itself if err is nil or io.EOF
// (i.e., err == nil || errors.Is(err, io.EOF)).
//
// If the target error message is empty,
// it uses "<no error message>" instead.
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
// or AutoWrapCustom, AutoWrapCustom finds the first error that is not
// generated by these functions along the Unwrap error tree and
// uses its error message instead.
//
// skip is the number of stack frames to ascend,
// with 0 identifying the caller of AutoWrapCustom.
// For example, if skip is 1, instead of the caller of AutoWrapCustom,
// the information of that caller's caller is used.
//
// exclusions are a set of errors.
// If err is nil or in exclusions, AutoWrapCustom returns err itself.
//
// If the target error message is empty,
// it uses "<no error message>" instead.
// If ms is invalid, it uses PrependFullFuncName instead.
func AutoWrapCustom(err error, ms ErrorMessageStrategy, skip int, exclusions ErrorReadOnlySet) error {
	if err == nil || exclusions != nil && exclusions.Contains(err) {
		return err
	}
	frame, ok := runtime.CallerFrame(skip + 1)
	if !ok || frame.Function == "" ||
		len(runtime.FuncPkg(frame.Function)) >= len(frame.Function) {
		panic(AutoMsg("cannot retrieve caller function name"))
	}
	if !ms.Valid() {
		ms = PrependFullFuncName
	}
	return &autoWrappedError{
		err:      err,
		ms:       ms,
		fullFunc: frame.Function,
	}
}

// IsAutoWrappedError reports whether err is generated by
// AutoWrap, AutoWrapSkip, or AutoWrapCustom.
func IsAutoWrappedError(err error) bool {
	// The type testing should not go along the Unwrap error tree,
	// so type assertion is used here instead of errors.As.
	_, ok := err.(*autoWrappedError)
	return ok
}

// UnwrapAutoWrappedError unwraps err and returns the result and true
// if err is generated by AutoWrap, AutoWrapSkip, or AutoWrapCustom.
// Otherwise, UnwrapAutoWrappedError returns err itself and false.
func UnwrapAutoWrappedError(err error) (error, bool) {
	// The type testing should not go along the Unwrap error tree,
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
	awe, ok := err.(*autoWrappedError)
	if ok {
		err = awe.Root()
	}
	return err, ok
}
