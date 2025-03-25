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

package inout

import "io"

// Writer extends io.Writer by adding a MustWrite method.
type Writer interface {
	io.Writer

	// MustWrite is like Write of io.Writer
	// but panics when encountering an error.
	//
	// If it panics, the error value passed to the call of panic
	// must be exactly of type *WritePanic.
	MustWrite(p []byte) (n int)
}

// ByteWriter extends io.ByteWriter by adding a MustWriteByte method.
type ByteWriter interface {
	io.ByteWriter

	// MustWriteByte is like WriteByte of io.ByteWriter
	// but panics when encountering an error.
	//
	// If it panics, the error value passed to the call of panic
	// must be exactly of type *WritePanic.
	MustWriteByte(c byte)
}

// RuneWriter contains a WriteRune method and a MustWriteRune method.
type RuneWriter interface {
	// WriteRune writes a single Unicode code point.
	//
	// It returns the number of bytes written and any write error encountered.
	WriteRune(r rune) (size int, err error)

	// MustWriteRune is like WriteRune but panics when encountering an error.
	//
	// If it panics, the error value passed to the call of panic
	// must be exactly of type *WritePanic.
	MustWriteRune(r rune) (size int)
}

// StringWriter extends io.StringWriter by adding a MustWriteString method.
type StringWriter interface {
	io.StringWriter

	// MustWriteString is like WriteString of io.StringWriter
	// but panics when encountering an error.
	//
	// If it panics, the error value passed to the call of panic
	// must be exactly of type *WritePanic.
	MustWriteString(s string) (n int)
}
