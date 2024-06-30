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

import (
	"fmt"
	"strings"

	"github.com/donyori/gogo/runtime"
)

// AutoMsg generates an error message by prepending the full function name
// (i.e., the package path-qualified function name; e.g., encoding/json.Marshal)
// of its caller to msg.
//
// If msg is empty, it uses "<no error message>" instead.
func AutoMsg(msg string) string {
	return AutoMsgCustom(msg, -1, 1)
}

// AutoMsgCustom generates an error message using msg.
// It prepends the function or package name of the caller to msg.
// Caller can specify its behavior by ms.
//
// skip is the number of stack frames to ascend,
// with 0 identifying the caller of AutoMsgCustom.
// For example, if skip is 1, instead of the caller of AutoMsgCustom,
// the information of that caller's caller is used.
//
// If msg is empty, it uses "<no error message>" instead.
// If ms is invalid, it uses PrependFullFuncName instead.
func AutoMsgCustom(msg string, ms ErrorMessageStrategy, skip int) string {
	if msg == "" {
		msg = "<no error message>"
	}
	if !ms.Valid() {
		ms = PrependFullFuncName
	}
	if ms == OriginalMsg {
		return msg
	}
	const AutoMsgCustomPrefix = "github.com/donyori/gogo/errors.AutoMsgCustom: "
	const CannotRetrieveCallerPanicMsg = AutoMsgCustomPrefix +
		"cannot retrieve caller function name"
	frame, ok := runtime.CallerFrame(skip + 1)
	if !ok || frame.Function == "" {
		panic(CannotRetrieveCallerPanicMsg)
	}
	prefix := frame.Function
	pkg := runtime.FuncPkg(frame.Function)
	if len(pkg) >= len(frame.Function) {
		panic(CannotRetrieveCallerPanicMsg)
	}
	if ms != PrependFullFuncName {
		switch ms {
		case PrependFullPkgName:
			prefix = pkg
		case PrependSimpleFuncName:
			prefix = frame.Function[len(pkg)+1:]
		case PrependSimplePkgName:
			prefix = pkg[strings.LastIndexByte(pkg, '/')+1:]
		default:
			// This should never happen, but will act as a safeguard for later,
			// as a default value doesn't make sense here.
			panic(fmt.Sprintf(
				AutoMsgCustomPrefix+
					"error message strategy is invalid (%v), which should never happen",
				ms,
			))
		}
	}
	if prefix == "" {
		return msg
	}
	return prefix + ": " + msg
}
