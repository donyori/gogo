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

import (
	"strings"

	"github.com/donyori/gogo/errors"
)

// ClosedError is an error indicating that a closable device is already closed.
//
// It records the device name and possible parent error.
// The parent error here is used to classify ClosedError instances,
// not the error that caused the current ClosedError.
//
// The client should use function NewClosedError to create a ClosedError.
//
// ErrClosed is a (direct or indirect) parent error of
// all ClosedError instances except itself.
// Therefore, the client can use errors.Is(err, ErrClosed)
// to test whether err is a ClosedError.
type ClosedError struct {
	// device is the name of the closable device.
	device string

	// parent is the parent error, which is wrapped by this ClosedError.
	parent error
}

// NewClosedError creates a new ClosedError
// with specified device name and parent error.
//
// parentErr must either be nil or satisfy that
// errors.Is(parentErr, ErrClosed) is true.
// If not, NewClosedError panics.
//
// If deviceName is empty, it uses "closer" instead.
// If parentErr is nil, it uses ErrClosed instead.
func NewClosedError(deviceName string, parentErr error) *ClosedError {
	deviceName = strings.TrimSpace(deviceName)
	if deviceName == "" {
		deviceName = "closer"
	}
	if parentErr == nil {
		parentErr = ErrClosed
	} else if !errors.Is(parentErr, ErrClosed) {
		panic(errors.AutoMsg("parentErr is neither nil nor an ErrClosed " +
			"(errors.Is(parentErr, ErrClosed) returns false)"))
	}
	return &ClosedError{
		device: deviceName,
		parent: parentErr,
	}
}

// DeviceName returns the device name passed to NewClosedError.
//
// If ce is nil, it returns "<nil>".
func (ce *ClosedError) DeviceName() string {
	if ce == nil {
		return "<nil>"
	}
	return ce.device
}

// Error reports the error message.
//
// If ce is nil, it returns "<nil>".
func (ce *ClosedError) Error() string {
	if ce == nil {
		return "<nil>"
	} else if ce.device == "" {
		if ce.parent != nil {
			// This should never happen, but will act as a safeguard for later.
			panic(errors.AutoMsg(
				"ClosedError has an empty device but a non-nil parent"))
		}
		return "closer is already closed (zero-value ClosedError)"
	}
	return ce.device + " is already closed"
}

// Unwrap returns its parent error (if any).
//
// If ce has no parent error, it returns nil.
func (ce *ClosedError) Unwrap() error {
	if ce == nil {
		return nil
	}
	return ce.parent
}

// ErrClosed is an error indicating that the closer is already closed.
//
// The client should use errors.Is to test whether an error is ErrClosed.
var ErrClosed = errors.AutoWrapCustom(
	&ClosedError{device: "closer"},
	errors.PrependFullPkgName,
	0,
	nil,
)

// ErrReaderClosed is an error indicating that the reader is already closed.
//
// The client should use errors.Is to test whether an error is ErrReaderClosed.
var ErrReaderClosed = errors.AutoWrapCustom(
	NewClosedError("reader", nil),
	errors.PrependFullPkgName,
	0,
	nil,
)

// ErrWriterClosed is an error indicating that the writer is already closed.
//
// The client should use errors.Is to test whether an error is ErrWriterClosed.
var ErrWriterClosed = errors.AutoWrapCustom(
	NewClosedError("writer", nil),
	errors.PrependFullPkgName,
	0,
	nil,
)

// WritePanic is the error passed to the call of panic
// in MustWrite methods and MustPrint methods.
//
// It records the error that caused the panic.
type WritePanic struct {
	// err is the error that caused the panic.
	err error
}

// NewWritePanic creates a new WritePanic with
// specified error that caused the panic.
func NewWritePanic(causeErr error) *WritePanic {
	return &WritePanic{err: causeErr}
}

// Error reports the error message.
//
// If wp is nil, it returns "<nil>".
func (wp *WritePanic) Error() string {
	if wp == nil {
		return "<nil>"
	} else if wp.err != nil {
		if msg := wp.err.Error(); msg != "" {
			return msg
		}
	}
	return "<no error message>"
}

// Unwrap returns the error that caused this panic.
//
// If wp is nil, it returns nil.
func (wp *WritePanic) Unwrap() error {
	if wp == nil {
		return nil
	}
	return wp.err
}
