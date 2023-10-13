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

package concurrency

import (
	"context"

	"github.com/donyori/gogo/errors"
)

// Canceler is a device to broadcast and detect cancellation signals.
type Canceler interface {
	// Cancel broadcasts the cancellation signal.
	//
	// It takes effect only once for one instance.
	// After the first call, subsequent calls to Cancel do nothing.
	Cancel()

	// C returns a channel for the cancellation signal.
	// The channel will be closed after the first call to the method Cancel.
	C() <-chan struct{}

	// Canceled reports whether the Canceler has broadcast
	// the cancellation signal (i.e., the method Cancel has been called).
	Canceled() bool
}

// ErrCanceled is an error indicating that
// the Canceler has broadcast a cancellation signal.
var ErrCanceled = errors.AutoNewCustom(
	"canceler broadcast a cancellation signal",
	errors.PrependFullPkgName,
	0,
)

// NewCanceler creates a new Canceler.
func NewCanceler() Canceler {
	return &onceCanceler{o: NewOnce(nil)}
}

// onceCanceler is an implementation of interface Canceler based on Once.
type onceCanceler struct {
	o Once
}

func (c *onceCanceler) Cancel() {
	c.o.Do()
}

func (c *onceCanceler) C() <-chan struct{} {
	return c.o.C()
}

func (c *onceCanceler) Canceled() bool {
	return c.o.Done()
}

// NewCancelerFromContext wraps the specified context and
// cancel function as a Canceler.
//
// The method Cancel of the returned Canceler
// calls the specified cancel function.
//
// The method C of the returned Canceler returns ctx.Done().
//
// The method Canceled of the returned Canceler
// reports whether ctx.Done() is closed.
//
// The client is responsible for guaranteeing that
// the specified cancel function is for the specified context.
//
// NewCancelerFromContext panics if ctx or cancel is nil.
func NewCancelerFromContext(
	ctx context.Context,
	cancel context.CancelFunc,
) Canceler {
	if ctx == nil {
		panic(errors.AutoMsg("the provided context is nil"))
	}
	if cancel == nil {
		panic(errors.AutoMsg("the provided cancel function is nil"))
	}
	return &contextCanceler{
		ctx:    ctx,
		cancel: cancel,
	}
}

// contextCanceler is an implementation of interface Canceler
// based on context.Context and context.CancelFunc.
type contextCanceler struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *contextCanceler) Cancel() {
	c.cancel()
}

func (c *contextCanceler) C() <-chan struct{} {
	return c.ctx.Done()
}

func (c *contextCanceler) Canceled() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}

// NewCancelerFromContextCause wraps the specified context and
// cancel function as a Canceler.
//
// The method Cancel of the returned Canceler
// calls the specified cancel function with ErrCanceled.
//
// The method C of the returned Canceler returns ctx.Done().
//
// The method Canceled of the returned Canceler
// reports whether ctx.Done() is closed.
//
// The client is responsible for guaranteeing that
// the specified cancel function is for the specified context.
//
// NewCancelerFromContextCause panics if ctx or cancel is nil.
func NewCancelerFromContextCause(
	ctx context.Context,
	cancel context.CancelCauseFunc,
) Canceler {
	if ctx == nil {
		panic(errors.AutoMsg("the provided context is nil"))
	}
	if cancel == nil {
		panic(errors.AutoMsg("the provided cancel function is nil"))
	}
	return &contextCauseCanceler{
		ctx:    ctx,
		cancel: cancel,
	}
}

// contextCauseCanceler is an implementation of interface Canceler
// based on context.Context and context.CancelCauseFunc.
type contextCauseCanceler struct {
	ctx    context.Context
	cancel context.CancelCauseFunc
}

func (c *contextCauseCanceler) Cancel() {
	c.cancel(ErrCanceled)
}

func (c *contextCauseCanceler) C() <-chan struct{} {
	return c.ctx.Done()
}

func (c *contextCauseCanceler) Canceled() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}
