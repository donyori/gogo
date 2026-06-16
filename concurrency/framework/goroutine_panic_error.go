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

package framework

import "fmt"

// GoroutinePanicError is an error recording the name of a panicked goroutine
// and the panic value (i.e., the argument passed to the function panic).
type GoroutinePanicError struct {
	// goroutine is the name of the panicked goroutine.
	goroutine string

	// value is the argument passed to the function panic.
	value any
}

// NewGoroutinePanicError creates a new GoroutinePanicError
// with the specified goroutine name and panic value.
//
// If the goroutine name is empty, it uses "<unknown>" instead.
func NewGoroutinePanicError(
	goroutineName string,
	panicValue any,
) *GoroutinePanicError {
	if goroutineName == "" {
		goroutineName = "<unknown>"
	}

	return &GoroutinePanicError{
		goroutine: goroutineName,
		value:     panicValue,
	}
}

// GoroutineName returns the name of the goroutine recorded in this error.
//
// If gpe is nil, it returns "<nil>".
func (gpe *GoroutinePanicError) GoroutineName() string {
	if gpe == nil {
		return "<nil>"
	}

	return gpe.goroutine
}

// PanicValue returns the panic value recorded in this error.
//
// If gpe is nil, it returns nil.
func (gpe *GoroutinePanicError) PanicValue() any {
	if gpe == nil {
		return nil
	}

	return gpe.value
}

// Error formats the panic record into a string
// and reports it as an error message.
//
// If gpe is nil or its value is nil, it returns "no panic".
func (gpe *GoroutinePanicError) Error() string {
	if gpe == nil || gpe.value == nil {
		return "no panic"
	}

	return fmt.Sprintf("panic on goroutine %s: %v", gpe.goroutine, gpe.value)
}

// Unwrap returns the panic value
// if the panic value implements the interface error.
// Otherwise, it returns nil.
//
// If gpe is nil, it returns nil as well.
func (gpe *GoroutinePanicError) Unwrap() error {
	if gpe == nil {
		return nil
	}

	err, _ := gpe.value.(error)

	return err
}
