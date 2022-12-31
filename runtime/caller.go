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

package runtime

import (
	stdruntime "runtime"
	"strings"
)

// FuncPkg returns the package name of the function fn.
//
// fn is the package path-qualified function name that uniquely identifies
// a single function in the program.
func FuncPkg(fn string) string {
	// The first dot ('.') after the last slash ('/') splits the package name
	// and the function name. The dots in the package name after the last slash
	// are escaped to URL encoding ('.' -> "%2e").
	i := strings.LastIndex(fn, "/")
	i = strings.Index(fn[i+1:], ".") + i + 1
	return fn[:i]
}

// CallerFrame returns the stack frame information of caller.
//
// skip is the number of stack frames to ascend,
// with 0 identifying the caller of CallerFrame.
//
// The return value ok is false if the information is unretrievable.
func CallerFrame(skip int) (frame stdruntime.Frame, ok bool) {
	rpc := make([]uintptr, 1)
	n := stdruntime.Callers(skip+2, rpc)
	if n < 1 {
		return
	}
	frame, _ = stdruntime.CallersFrames(rpc).Next()
	return frame, frame.PC != 0
}

// FramePkgFunc returns the full package name and simple function name of
// the function in the specified stack frame.
//
// The return value ok is false if the information is unretrievable.
func FramePkgFunc(frame stdruntime.Frame) (pkg, fn string, ok bool) {
	if frame.PC == 0 || frame.Function == "" {
		return
	}
	pkg = FuncPkg(frame.Function)
	return pkg, frame.Function[len(pkg)+1:], true
}

// CallerPkgFunc returns the full package name and simple function name of
// its caller.
//
// skip is the number of stack frames to ascend,
// with 0 identifying the caller of CallerPkgFunc.
//
// The return value ok is false if the information is unretrievable.
func CallerPkgFunc(skip int) (pkg, fn string, ok bool) {
	frame, ok := CallerFrame(skip + 1)
	if !ok {
		return
	}
	return FramePkgFunc(frame)
}
