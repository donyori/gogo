// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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

// An error wrapper to wrap an error as another error with method Unwrap.
// It is allowed to return the error itself directly if the error that will be
// wrapped is in the exclusion list.
type Wrapper interface {
	// Wrap err as a new error. Or return err itself directly
	// if err is in the exclusion list.
	Wrap(err error) error
}

// Interfaces combining error and other methods supported since Go 1.13.

// An error with method Unwrap, to simplify working with errors that
// contain other errors since Go 1.13. For more details,
// see <https://blog.golang.org/go1.13-errors>.
type WrappingError interface {
	error

	// Return the contained error.
	// Return nil if it contains no error.
	Unwrap() error
}

// An error with method Is, to custom the behavior of errors.Is.
// For more details, see <https://blog.golang.org/go1.13-errors>.
type ErrorIs interface {
	error

	// Report whether any error in its error chain matches target.
	// See errors.Is for detail.
	Is(target error) bool
}

// An error with method As, to custom the behavior of errors.As.
// For more details, see <https://blog.golang.org/go1.13-errors>.
type ErrorAs interface {
	error

	// Find the first error in its error chain that matches target,
	// and if so, set target to that error value and returns true.
	// See errors.As for detail.
	As(target interface{}) bool
}
