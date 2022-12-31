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

import (
	"fmt"
	"strings"

	"github.com/donyori/gogo/runtime"
)

// AutoMsg generates an error message by prepending the full function name
// (i.e., the package path-qualified function name) of its caller to msg.
//
// If msg is empty, it will use "<no error message>" instead.
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
// the information of that caller's caller will be used.
//
// If msg is empty, it will use "<no error message>" instead.
// If ms is invalid, it will use PrependFullFuncName instead.
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
	frame, ok := runtime.CallerFrame(skip + 1)
	if !ok {
		return "(retrieving caller failed) error: " + msg
	}
	var prefix string
	switch ms {
	case PrependFullFuncName:
		prefix = frame.Function
	case PrependFullPkgName:
		prefix, _, ok = runtime.FramePkgFunc(frame)
		if !ok {
			return "(retrieving caller failed) error: " + msg
		}
	case PrependSimpleFuncName:
		_, prefix, ok = runtime.FramePkgFunc(frame)
		if !ok {
			return "(retrieving caller failed) error: " + msg
		}
	case PrependSimplePkgName:
		prefix, _, ok = runtime.FramePkgFunc(frame)
		if !ok {
			return "(retrieving caller failed) error: " + msg
		}
		prefix = prefix[strings.LastIndex(prefix, "/")+1:]
	default:
		// This should never happen, but will act as a safeguard for later,
		// as a default value doesn't make sense here.
		panic(fmt.Sprintf(
			"github.com/donyori/gogo/errors.AutoMsgCustom: "+
				"error message strategy is invalid (%v), which should never happen",
			ms,
		))
	}
	if prefix == "" {
		return msg
	}
	return prefix + ": " + msg
}
