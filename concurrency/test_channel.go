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

package concurrency

// TestChannel returns true if the specified channel c is closed or
// if a message has been received or buffered in c.
//
// The code snippet
//
//	if concurrency.TestChannel(c) {
//		// branch A
//	} else {
//		// branch B
//	}
//
// is equivalent to
//
//	select {
//	case <-c:
//		// branch A
//	default:
//		// branch B
//	}
func TestChannel[T any](c <-chan T) bool {
	select {
	case <-c:
		return true
	default:
		return false
	}
}

// TestF returns true if the specified function f is not nil
// and the channel acquired by f() is closed or
// a message has been received or buffered in the channel.
// It is shorthand for
//
//	f != nil && concurrency.TestChannel(f())
//
// The code snippet
//
//	if concurrency.TestF(f) {
//		// branch A
//	} else {
//		// branch B
//	}
//
// is equivalent to
//
//	var c <-chan T
//	if f != nil {
//		c = f()
//	}
//
//	select {
//	case <-c:
//		// branch A
//	default:
//		// branch B
//	}
func TestF[T any](f func() <-chan T) bool {
	return f != nil && TestChannel(f())
}

// TestC returns true if the specified interface itf is not nil
// and the channel acquired by itf.C() is closed or
// a message has been received or buffered in the channel.
// It is shorthand for
//
//	itf != nil && concurrency.TestChannel(itf.C())
//
// The code snippet
//
//	if concurrency.TestC(itf) {
//		// branch A
//	} else {
//		// branch B
//	}
//
// is equivalent to
//
//	var c <-chan T
//	if itf != nil {
//		c = itf.C()
//	}
//
//	select {
//	case <-c:
//		// branch A
//	default:
//		// branch B
//	}
func TestC[T any](itf interface{ C() <-chan T }) bool {
	return itf != nil && TestChannel(itf.C())
}

// TestDone returns true if the specified interface itf is not nil
// and the channel acquired by itf.Done() is closed or
// a message has been received or buffered in the channel.
// It is shorthand for
//
//	itf != nil && concurrency.TestChannel(itf.Done())
//
// The code snippet
//
//	if concurrency.TestDone(itf) {
//		// branch A
//	} else {
//		// branch B
//	}
//
// is equivalent to
//
//	var c <-chan T
//	if itf != nil {
//		c = itf.Done()
//	}
//
//	select {
//	case <-c:
//		// branch A
//	default:
//		// branch B
//	}
//
// A typical usage of TestDone is to test whether a context.Context
// is done (canceled, deadline expired, or timeout elapsed), for example:
//
//	// ctx is of type context.Context.
//	if concurrency.TestDone(ctx) {
//		// do something when the context is done
//	} else {
//		// do something else before the context is done
//	}
func TestDone[T any](itf interface{ Done() <-chan T }) bool {
	return itf != nil && TestChannel(itf.Done())
}
