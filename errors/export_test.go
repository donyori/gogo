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

import "github.com/donyori/gogo/runtime"

// Export for testing only.

// NewAutoWrappedError returns an error of the type that is the same as
// the error type returned by functions AutoWrap, AutoWrapSkip,
// and AutoWrapCustom.
//
// It is used to simulate the error generated by AutoWrap, AutoWrapSkip,
// and AutoWrapCustom for testing.
func NewAutoWrappedError(err error) error {
	frame, ok := runtime.CallerFrame(1)
	if !ok || frame.Function == "" ||
		len(runtime.FuncPkg(frame.Function)) >= len(frame.Function) {
		panic(AutoMsg("cannot retrieve caller function name"))
	}
	return &autoWrappedError{
		err:      err,
		ms:       PrependSimpleFuncName,
		fullFunc: frame.Function,
	}
}

type ErrorListImpl = errorList

func (el *ErrorListImpl) GetList() []error {
	if el == nil {
		return nil
	}
	return el.list
}
