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

package inout

import (
	"fmt"
	"io"

	"github.com/donyori/gogo/errors"
)

// MustFprintf is like fmt.Fprintf but panics when encountering an error.
//
// If it panics, the error value passed to the call to panic
// must be exactly of type *WritePanicError.
func MustFprintf(w io.Writer, format string, arg ...any) (n int) {
	n, err := fmt.Fprintf(w, format, arg...)
	if err != nil {
		panic(NewWritePanicError(errors.AutoWrap(err)))
	}

	return
}

// MustFprint is like fmt.Fprint but panics when encountering an error.
//
// If it panics, the error value passed to the call to panic
// must be exactly of type *WritePanicError.
func MustFprint(w io.Writer, arg ...any) (n int) {
	n, err := fmt.Fprint(w, arg...)
	if err != nil {
		panic(NewWritePanicError(errors.AutoWrap(err)))
	}

	return
}

// MustFprintln is like fmt.Fprintln but panics when encountering an error.
//
// If it panics, the error value passed to the call to panic
// must be exactly of type *WritePanicError.
func MustFprintln(w io.Writer, arg ...any) (n int) {
	n, err := fmt.Fprintln(w, arg...)
	if err != nil {
		panic(NewWritePanicError(errors.AutoWrap(err)))
	}

	return
}

// Printer contains methods Printf, Print, Println, and their "Must" versions.
type Printer interface {
	// Printf formats arguments and writes to its underlying data stream.
	// Arguments are handled in the manner of fmt.Printf.
	Printf(format string, arg ...any) (n int, err error)

	// MustPrintf is like Printf but panics when encountering an error.
	//
	// If it panics, the error value passed to the call to panic
	// must be exactly of type *WritePanicError.
	MustPrintf(format string, arg ...any) (n int)

	// Print formats arguments and writes to its underlying data stream.
	// Arguments are handled in the manner of fmt.Print.
	Print(arg ...any) (n int, err error)

	// MustPrint is like Print but panics when encountering an error.
	//
	// If it panics, the error value passed to the call to panic
	// must be exactly of type *WritePanicError.
	MustPrint(arg ...any) (n int)

	// Println formats arguments and writes to its underlying data stream.
	// Arguments are handled in the manner of fmt.Println.
	Println(arg ...any) (n int, err error)

	// MustPrintln is like Println but panics when encountering an error.
	//
	// If it panics, the error value passed to the call to panic
	// must be exactly of type *WritePanicError.
	MustPrintln(arg ...any) (n int)
}
