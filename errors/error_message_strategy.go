// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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

import "strconv"

// ErrorMessageStrategy is a strategy for auto-generating error messages.
type ErrorMessageStrategy int8

// Enumeration of supported error message strategies.
const (
	OriginalMsg             ErrorMessageStrategy = 1 + iota // OriginalMessage
	PrependFullFuncName                                     // PrependFullFunctionName
	PrependFullPkgName                                      // PrependFullPackageName
	PrependSimpleFuncName                                   // PrependSimpleFunctionName
	PrependSimplePkgName                                    // PrependSimplePackageName
	maxErrorMessageStrategy                                 // ErrorMessageStrategy(6)
)

// Before running the following command, please make sure the numeric value
// in the line comment of maxErrorMessageStrategy is correct.
//
//go:generate stringer -type=ErrorMessageStrategy -output=error_message_strategy_string.go -linecomment

// Valid returns true if the error message strategy is known.
//
// Known error message strategies are shown as follows:
//   - OriginalMsg: use the error message itself
//   - PrependFullFuncName: add the full function name (i.e., the package path-qualified function name) before the error message
//   - PrependFullPkgName: add the full package name before the error message
//   - PrependSimpleFuncName: add the simple function name before the error message
//   - PrependSimplePkgName: add the simple package name before the error message
func (i ErrorMessageStrategy) Valid() bool {
	return i > 0 && i < maxErrorMessageStrategy
}

// MustValid panics if i is invalid.
// Otherwise, it does nothing.
func (i ErrorMessageStrategy) MustValid() {
	if !i.Valid() {
		panic(AutoMsgCustom(
			"unknown message strategy: "+strconv.FormatInt(int64(i), 10),
			-1,
			1,
		))
	}
}
