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
	"io"
	"strings"

	"github.com/donyori/gogo/errors"
)

// Closer is an interface combining the basic Close method and a Closed method.
//
// Method Closed reports whether this closer is closed.
type Closer interface {
	io.Closer

	// Closed reports whether this closer is closed successfully.
	Closed() bool
}

// noOpCloser is an implementation of interface Closer.
//
// Its method Close is a no-op, which does nothing and returns nil.
type noOpCloser struct {
	closed bool
}

// NewNoOpCloser creates a new closer with a no-op method Close.
func NewNoOpCloser() Closer {
	return new(noOpCloser)
}

func (noc *noOpCloser) Close() error {
	noc.closed = true
	return nil
}

func (noc *noOpCloser) Closed() bool {
	return noc.closed
}

// noErrorCloser is an implementation of interface Closer.
//
// After the first successful call to method Close,
// method Close does nothing and returns nil.
type noErrorCloser struct {
	c      io.Closer
	closed bool // closed is true if the closer is successfully closed.
}

// WrapNoErrorCloser wraps the specified closer into a Closer,
// whose method Close does nothing and returns nil
// after the first successful call.
//
// It panics if closer is nil.
func WrapNoErrorCloser(closer io.Closer) Closer {
	if closer == nil {
		panic(errors.AutoMsg("closer is nil"))
	}
	return &noErrorCloser{c: closer}
}

func (nec *noErrorCloser) Close() error {
	if nec.closed {
		return nil
	}
	err := nec.c.Close()
	if err == nil {
		nec.closed = true
	}
	return err // don't wrap the error
}

func (nec *noErrorCloser) Closed() bool {
	return nec.closed
}

// errorCloser is an implementation of interface Closer.
//
// After the first successful call to method Close,
// method Close does nothing and returns a *ClosedError.
type errorCloser struct {
	c      io.Closer
	closed bool   // closed is true if the closer is successfully closed.
	dn     string // Device name.
	pe     error  // Parent error of the ClosedError.
}

// WrapErrorCloser wraps the specified closer into a Closer,
// whose method Close does nothing and returns a *ClosedError
// after the first successful call.
//
// It panics if closer is nil.
//
// deviceName is the name of the specified closer.
// If deviceName is empty, it uses "closer" instead.
//
// parentErr is the parent error of the ClosedError returned by method Close.
// The parent error here is used to classify ClosedError instances,
// not the error that caused the current ClosedError.
// parentErr must either be nil or satisfy that
// errors.Is(parentErr, ErrClosed) is true.
// If not, WrapErrorCloser panics.
// If parentErr is nil, it uses ErrClosed instead.
func WrapErrorCloser(
	closer io.Closer,
	deviceName string,
	parentErr error,
) Closer {
	if closer == nil {
		panic(errors.AutoMsg("closer is nil"))
	}
	deviceName = strings.TrimSpace(deviceName)
	if deviceName == "" {
		deviceName = "closer"
	}
	if parentErr == nil {
		parentErr = ErrClosed
	} else if !errors.Is(parentErr, ErrClosed) {
		panic(errors.AutoMsg(
			"parentErr is neither nil nor an ErrClosed (errors.Is(parentErr, ErrClosed) returns false)"))
	}
	return &errorCloser{
		c:  closer,
		dn: deviceName,
		pe: parentErr,
	}
}

func (ec *errorCloser) Close() error {
	if ec.closed {
		return NewClosedError(ec.dn, ec.pe)
	}
	err := ec.c.Close()
	if err == nil {
		ec.closed = true
	}
	return err // don't wrap the error
}

func (ec *errorCloser) Closed() bool {
	return ec.closed
}

// MultiCloser is a device to close multiple closers sequentially.
//
// It closes its closers sequentially from the last one to the first one.
//
// Its method Closed returns true if and only if
// all its closers are already successfully closed.
//
// If its option tryAll is enabled,
// its method Close tries to close all its closers,
// regardless of whether any error occurs, and returns all errors encountered.
// (It returns an ErrorList if there are multiple errors.)
//
// If its option tryAll is disabled, when an error occurs,
// its method Close stops closing other closers and returns this error.
//
// If its option noError is enabled, its method Close does nothing and
// returns nil after the first successful call.
//
// If its option noError is disabled, its method Close does nothing and
// returns a *ClosedError after the first successful call.
type MultiCloser interface {
	Closer

	// CloserClosed reports whether the specified closer is closed successfully.
	//
	// It returns two boolean indicators:
	// closed reports whether the specified closer
	// has been successfully closed by this MultiCloser.
	// ok reports whether the specified closer is in this MultiCloser.
	CloserClosed(closer io.Closer) (closed, ok bool)
}

// Masks for the field flag of struct multiCloser.
const (
	multiCloserTryAllMask  byte = 0x01
	multiCloserNoErrorMask byte = 0x02
)

// multiCloser is an implementation of interface MultiCloser.
type multiCloser struct {
	cm   map[io.Closer]bool // Closed map, to record whether the closer is closed successfully.
	cs   []io.Closer
	idx  int  // Index of the last successfully closed closer. (It equals len(cs) initially.)
	flag byte // The first bit (0x01) is the option tryAll, and the second bit (0x02) is the option noError.
}

// NewMultiCloser creates a new MultiCloser.
//
// If the option tryAll is enabled,
// its method Close tries to close all its closers,
// regardless of whether any error occurs, and returns all errors encountered.
// (It returns an ErrorList if there are multiple errors.)
//
// If the option tryAll is disabled, when an error occurs,
// its method Close stops closing other closers and returns this error.
//
// If the option noError is enabled, its method Close does nothing and
// returns nil after the first successful call.
//
// If the option noError is disabled, its method Close does nothing and
// returns a *ClosedError after the first successful call.
//
// closer is the closers provided to the MultiCloser.
// All nil closers are ignored.
// If there is no non-nil closer,
// the MultiCloser performs as an already closed closer.
func NewMultiCloser(tryAll, noError bool, closer ...io.Closer) MultiCloser {
	mc := &multiCloser{
		cm: make(map[io.Closer]bool),
		cs: make([]io.Closer, 0, len(closer)),
	}
	for _, c := range closer {
		if c != nil {
			mc.cm[c] = false
			mc.cs = append(mc.cs, c)
		}
	}
	mc.idx = len(mc.cs)
	if tryAll {
		mc.flag |= multiCloserTryAllMask
	}
	if noError {
		mc.flag |= multiCloserNoErrorMask
	}
	return mc
}

func (mc *multiCloser) Close() error {
	if mc.idx == 0 {
		var err error
		if mc.flag&multiCloserNoErrorMask == 0 {
			err = NewClosedError("MultiCloser", nil)
		}
		return err
	}

	if mc.flag&multiCloserTryAllMask == 0 {
		for mc.idx > 0 {
			err := mc.cs[mc.idx-1].Close()
			if err != nil {
				return err // don't wrap the error
			}
			mc.cm[mc.cs[mc.idx-1]] = true
			mc.idx--
		}
		return nil
	}
	el := errors.NewErrorList(true)
	for i := mc.idx - 1; i >= 0; i-- {
		if mc.cm[mc.cs[i]] {
			continue
		}
		err := mc.cs[i].Close()
		el.Append(err)
		if err == nil {
			mc.cm[mc.cs[i]] = true
			if !el.Erroneous() {
				mc.idx = i
			}
		}
	}
	return el.ToError() // don't wrap the error
}

func (mc *multiCloser) Closed() bool {
	return mc.idx == 0
}

func (mc *multiCloser) CloserClosed(closer io.Closer) (closed, ok bool) {
	closed, ok = mc.cm[closer]
	return
}
