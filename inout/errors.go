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

package inout

import (
	"strings"

	"github.com/donyori/gogo/errors"
)

// ClosedError is an error indicating that a closable device is already closed.
//
// It records the device name and possible parent error.
//
// The client should use function NewClosedError to create a ClosedError
// rather than directly using a variable declaration statement.
//
// ErrClosed is a (direct or indirect) parent error of
// all ClosedError except itself.
// Therefore, the client can use errors.Is(err, ErrClosed)
// to test whether err is a ClosedError.
type ClosedError struct {
	deviceName string
	parentErr  error // Parent error, which is wrapped by this ClosedError.
}

// NewClosedError creates a new ClosedError
// with specified device name and parent error.
//
// if deviceName is empty, it will use "closer" instead.
// if parentErr is nil, it will use ErrClosed instead.
func NewClosedError(deviceName string, parentErr error) *ClosedError {
	if deviceName == "" {
		deviceName = "closer"
	}
	if parentErr == nil {
		parentErr = ErrClosed
	}
	return &ClosedError{strings.TrimSpace(deviceName), parentErr}
}

// Error reports the error message, which performs the same as method String.
func (ce *ClosedError) Error() string {
	return ce.String()
}

// String returns the error message, which performs the same as method Error.
func (ce *ClosedError) String() string {
	return ce.deviceName + " is already closed"
}

// Unwrap returns its parent error (if any).
//
// If this error has no parent error, it returns nil.
func (ce *ClosedError) Unwrap() error {
	return ce.parentErr
}

// ErrClosed is an error indicating that the closer is already closed.
var ErrClosed = errors.AutoWrap(&ClosedError{deviceName: "closer"})

// ErrReaderClosed is an error indicating that the reader is already closed.
var ErrReaderClosed = errors.AutoWrap(&ClosedError{"reader", ErrClosed})

// ErrWriterClosed is an error indicating that the writer is already closed.
var ErrWriterClosed = errors.AutoWrap(&ClosedError{"writer", ErrClosed})
