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

import "strconv"

// Strategy for auto generating error message.
type ErrorMessageStrategy int8

// Enumeration of supported error message strategies.
const (
	OriginalMsg          ErrorMessageStrategy = iota // OriginalMessage
	PrefixFullPkgName                                // PrefixFullPackageName
	PrefixSimplePkgName                              // PrefixSimplePackageName
	PrefixFullFuncName                               // PrefixFullFunctionName
	PrefixSimpleFuncName                             // PrefixSimpleFunctionName
)

//go:generate stringer -type=ErrorMessageStrategy -output=error_message_strategy_string.go -linecomment

// Return true if the error message strategy is known.
func (i ErrorMessageStrategy) Valid() bool {
	return i >= OriginalMsg && i <= PrefixSimpleFuncName
}

// Panic if i is invalid. Do nothing otherwise.
func (i ErrorMessageStrategy) MustValid() {
	if !i.Valid() {
		panic(AutoMsgWithStrategy("unknown message strategy: "+strconv.FormatInt(int64(i), 10), defaultMessageStrategy, 1))
	}
}
