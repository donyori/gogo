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

package errors

import (
	stderrors "errors"
	"reflect"
)

// Export variables and functions from standard package errors for convenience.

// ErrUnsupported is exactly standard errors.ErrUnsupported.
var ErrUnsupported error = stderrors.ErrUnsupported

// New directly calls standard errors.New.
func New(msg string) error {
	return stderrors.New(msg)
}

// Unwrap directly calls standard errors.Unwrap.
func Unwrap(err error) error {
	return stderrors.Unwrap(err)
}

// Is directly calls standard errors.Is.
func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

// As calls standard errors.As,
// but panics if target is of type *error,
// because As always returns true for that,
// so that the function call is senseless.
func As(err error, target any) bool {
	if reflect.TypeOf(target) == errorPointerType {
		panic(AutoMsg("target is of type *error; " +
			"As always returns true for that"))
	}
	return stderrors.As(err, target)
}

// AsType calls standard errors.AsType,
// but panics if the type E is exactly error,
// because AsType always returns an error of the same type and true for that,
// so that the function call is senseless.
func AsType[E error](err error) (E, bool) {
	if reflect.TypeFor[E]() == errorType {
		panic(AutoMsg("type E is exactly error; AsType[error] is senseless"))
	}
	return stderrors.AsType[E](err)
}

// Join directly calls standard errors.Join.
func Join(err ...error) error {
	return stderrors.Join(err...)
}

var (
	// errorPointerType is the reflect.Type of *error.
	errorPointerType = reflect.TypeFor[*error]()

	// errorType is the reflect.Type of error.
	errorType = reflect.TypeFor[error]()
)
