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

import stderrors "errors"

// AutoNew creates a new error with specified error message msg,
// and then wraps it by prepending the full function name
// (i.e., the package path-qualified function name) of its caller
// to that error message.
//
// If msg is empty, it will use "<no error message>" instead.
func AutoNew(msg string) error {
	return AutoWrapCustom(stderrors.New(msg), PrependFullFuncName, 1, nil)
}

// AutoNewCustom creates a new error with specified error message msg,
// and then wraps it by prepending the function or package name of the caller
// to that error message.
// Caller can specify its behavior by ms.
//
// skip is the number of stack frames to ascend,
// with 0 identifying the caller of AutoNewCustom.
// For example, if skip is 1, instead of the caller of AutoNewCustom,
// the information of that caller's caller will be used.
//
// If msg is empty, it will use "<no error message>" instead.
// If ms is invalid, it will use PrependFullFuncName instead.
func AutoNewCustom(msg string, ms ErrorMessageStrategy, skip int) error {
	return AutoWrapCustom(stderrors.New(msg), ms, skip+1, nil)
}
