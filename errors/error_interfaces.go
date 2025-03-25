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

// Interfaces combining error and other methods (Unwrap, Is, As)
// supported since Go 1.13.

// ErrorUnwrap is an error with method Unwrap,
// to simplify working with errors that contain other errors since Go 1.13.
//
// For more details, see <https://go.dev/blog/go1.13-errors>.
type ErrorUnwrap interface {
	error

	// Unwrap returns the contained error.
	// It returns nil if it contains no error.
	//
	// See errors.Unwrap for detail.
	Unwrap() error
}

// ErrorUnwrapMultiple is an error with method Unwrap
// that returns multiple errors.
//
// Since Go 1.20, the standard library expands support for error wrapping
// to permit an error to wrap multiple other errors.
// For more details, see <https://go.dev/doc/go1.20#errors>.
type ErrorUnwrapMultiple interface {
	error

	// Unwrap returns a list of the contained errors.
	// It permits an error to wrap multiple other errors.
	//
	// The []error returned should not contain any nil-value error.
	Unwrap() []error
}

// ErrorIs is an error with method Is, to custom the behavior of errors.Is.
//
// For more details, see <https://go.dev/blog/go1.13-errors>.
type ErrorIs interface {
	error

	// Is reports whether any error in its Unwrap error tree matches target.
	//
	// See errors.Is for detail.
	Is(target error) bool
}

// ErrorAs is an error with method As, to custom the behavior of errors.As.
//
// For more details, see <https://go.dev/blog/go1.13-errors>.
type ErrorAs interface {
	error

	// As finds the first error in its Unwrap error tree that matches target,
	// and if so, sets target to that error value and returns true.
	// Otherwise, it returns false.
	//
	// See errors.As for detail.
	As(target any) bool
}
