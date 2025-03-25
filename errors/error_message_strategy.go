// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2025  Yuan Gao
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
	// OriginalMsg uses the error message itself.
	OriginalMsg ErrorMessageStrategy = 1 + iota // OriginalMessage

	// PrependFullFuncName adds the full function name
	// (i.e., the package path-qualified function name;
	// e.g., encoding/json.Marshal)
	// before the error message.
	PrependFullFuncName // PrependFullFunctionName

	// PrependFullPkgName adds the full package name (e.g., encoding/json)
	// before the error message.
	PrependFullPkgName // PrependFullPackageName

	// PrependSimpleFuncName adds the simple function name
	// (e.g., Marshal, rather than encoding/json.Marshal or json.Marshal)
	// before the error message.
	PrependSimpleFuncName // PrependSimpleFunctionName

	// PrependSimplePkgName adds the simple package name
	// (e.g., json, rather than encoding/json)
	// before the error message.
	PrependSimplePkgName // PrependSimplePackageName

	// maxErrorMessageStrategy is the upper bound (exclusive)
	// of the supported error message strategies.
	maxErrorMessageStrategy // ErrorMessageStrategy(6)
)

// Before running the following command, please make sure the numeric value
// in the line comment of maxErrorMessageStrategy is correct.
//
//go:generate stringer -type=ErrorMessageStrategy -output=error_message_strategy_string.go -linecomment

// Valid returns true if the error message strategy is known.
//
// Known error message strategies are shown as follows:
//   - OriginalMsg (1): use the error message itself
//   - PrependFullFuncName (2): add the full function name
//     (i.e., the package path-qualified function name;
//     e.g., encoding/json.Marshal) before the error message
//   - PrependFullPkgName (3): add the full package name
//     (e.g., encoding/json) before the error message
//   - PrependSimpleFuncName (4): add the simple function name
//     (e.g., Marshal, rather than encoding/json.Marshal or json.Marshal)
//     before the error message
//   - PrependSimplePkgName (5): add the simple package name
//     (e.g., json, rather than encoding/json) before the error message
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
