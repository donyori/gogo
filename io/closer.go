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

package io

import (
	stdio "io"

	"github.com/donyori/gogo/errors"
)

// Closer is an interface combining the basic Close method and a Closed method.
//
// Method Closed reports whether this closer is closed.
type Closer interface {
	stdio.Closer

	// Closed reports whether this closer is closed successfully.
	Closed() bool
}

// noErrorCloser is an implementation of interface Closer.
//
// After the first successful call to method Close,
// method Close will do nothing and return nil.
type noErrorCloser struct {
	c  stdio.Closer
	ok bool // ok is true if the closer is successfully closed.
}

// WrapNoErrorCloser wraps the specified closer into a Closer,
// whose method Close will do nothing and return nil
// after the first successful call.
//
// It panics if closer is nil.
func WrapNoErrorCloser(closer stdio.Closer) Closer {
	if closer == nil {
		panic(errors.AutoMsg("closer is nil"))
	}
	return &noErrorCloser{c: closer}
}

// Close closes its underlying closer and returns error encountered.
//
// It does nothing and returns nil after the first successful call.
func (nec *noErrorCloser) Close() error {
	if nec.ok {
		return nil
	}
	err := nec.c.Close()
	if err == nil {
		nec.ok = true
	}
	return err // Don't wrap the error.
}

// Closed reports whether this closer is closed successfully.
func (nec *noErrorCloser) Closed() bool {
	return nec.ok
}

// errorCloser is an implementation of interface Closer.
//
// After the first successful call to method Close,
// method Close will do nothing and return a *ClosedError.
type errorCloser struct {
	c  stdio.Closer
	ok bool   // ok is true if the closer is successfully closed.
	dn string // Device name.
	pe error  // Parent error of the ClosedError.
}

// WrapErrorCloser wraps the specified closer into a Closer,
// whose method Close will do nothing and return a *ClosedError
// after the first successful call.
//
// It panics if closer is nil.
//
// deviceName is the name of the specified closer.
// If deviceName is empty, it will use "closer" instead.
//
// parentErr is the parent error of the ClosedError returned by method Close.
// If parentErr is nil, it will use ErrClosed instead.
func WrapErrorCloser(closer stdio.Closer, deviceName string, parentErr error) Closer {
	if closer == nil {
		panic(errors.AutoMsg("closer is nil"))
	}
	if deviceName == "" {
		deviceName = "closer"
	}
	if parentErr == nil {
		parentErr = ErrClosed
	}
	return &errorCloser{
		c:  closer,
		dn: deviceName,
		pe: parentErr,
	}
}

// Close closes its underlying closer and returns error encountered.
//
// It does nothing and returns a *ClosedError after the first successful call.
func (ec *errorCloser) Close() error {
	if ec.ok {
		return NewClosedError(ec.dn, ec.pe)
	}
	err := ec.c.Close()
	if err == nil {
		ec.ok = true
	}
	return err // Don't wrap the error.
}

// Closed reports whether this closer is closed successfully.
func (ec *errorCloser) Closed() bool {
	return ec.ok
}
