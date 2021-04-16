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
	"testing"

	"github.com/donyori/gogo/errors"
)

type testCloser struct {
	tb     testing.TB
	closed bool
	err    error // Test can set err to simulate a failed call to Close.
}

func (tc *testCloser) Close() error {
	if tc.closed {
		tc.tb.Error("Call testCloser.Close again.")
		return errors.AutoNew("call testCloser.Close again") // Don't use *ClosedError here.
	}
	tc.closed = tc.err == nil
	return tc.err
}

func TestNoOpCloser(t *testing.T) {
	noc := NewNoOpCloser()
	if noc.Closed() {
		t.Error("noc.Closed is true before first call to Close.")
	}
	for i := 1; i <= 10; i++ {
		err := noc.Close()
		if err != nil {
			t.Errorf("Error on %d call to Close: %v.", i, err)
		}
		if !noc.Closed() {
			t.Errorf("noc.Closed is false after %d call to Close.", i)
		}
	}
}

func TestNoErrorCloser(t *testing.T) {
	nec := WrapNoErrorCloser(&testCloser{tb: t})
	if nec.Closed() {
		t.Error("nec.Closed is true before first call to Close.")
	}
	for i := 1; i <= 10; i++ {
		err := nec.Close()
		if err != nil {
			t.Errorf("Error on %d call to Close: %v.", i, err)
		}
		if !nec.Closed() {
			t.Errorf("nec.Closed is false after %d call to Close.", i)
		}
	}
}

func TestErrorCloser(t *testing.T) {
	ec := WrapErrorCloser(&testCloser{tb: t}, "testCloser", nil)
	if ec.Closed() {
		t.Error("ec.Closed is true before first call to Close.")
	}
	err := ec.Close()
	if err != nil {
		t.Errorf("Error on first call to Close: %v.", err)
	}
	if !ec.Closed() {
		t.Error("ec.Closed is false after first call to Close.")
	}
	for i := 2; i <= 10; i++ {
		err = ec.Close()
		if err == nil {
			t.Errorf("No error on %d call to Close, but wanted a *ClosedError.", i)
		} else if !errors.Is(err, ErrClosed) {
			t.Errorf("Error on %d call to Close: %v.", i, err)
		}
		if !ec.Closed() {
			t.Errorf("ec.Closed is false after %d call to Close.", i)
		}
	}
}

func TestMultiCloser(t *testing.T) {
	testMultiCloser(t, false, false)
	testMultiCloser(t, false, true)
	testMultiCloser(t, true, false)
	testMultiCloser(t, true, true)
}

func testMultiCloser(t *testing.T, tryAll, noError bool) {
	failErr := errors.New("an error to simulate a failed call to Close")
	closers := []stdio.Closer{
		&testCloser{tb: t, err: failErr},
		&testCloser{tb: t},
		&testCloser{tb: t, err: failErr},
		&testCloser{tb: t},
	}
	anotherCloser := &testCloser{tb: t}
	mc := NewMultiCloser(tryAll, noError, closers...)

	// Before the first call to Close:
	if mc.Closed() {
		t.Errorf("mc.Closed is true before the first call to mc.Close. tryAll: %t, noError: %t.",
			tryAll, noError)
	}
	for i, closer := range closers {
		closed, ok := mc.CloserClosed(closer)
		if !ok {
			t.Errorf("The %d closer is in mc but mc.CloserClosed returns (%t, %t). tryAll: %t, noError: %t.",
				i, closed, ok, tryAll, noError)
		}
		if closed {
			t.Errorf("mc.CloserClosed for the %d closer is true before the first call to mc.Close. tryAll: %t, noError: %t.",
				i, tryAll, noError)
		}
	}
	closed, ok := mc.CloserClosed(anotherCloser)
	if closed || ok {
		t.Errorf("mc.CloserClosed returns (%t, %t) for anotherCloser. tryAll: %t, noError: %t.",
			closed, ok, tryAll, noError)
	}

	// First call (a failed call) to Close:
	testMultiCloserOneCall(t, tryAll, noError, failErr, closers, anotherCloser, mc, 0)

	closers[2].(*testCloser).err = nil
	// Second call (a failed call) to Close:
	testMultiCloserOneCall(t, tryAll, noError, failErr, closers, anotherCloser, mc, 1)

	closers[0].(*testCloser).err = nil
	// Third call (a successful call) to Close:
	testMultiCloserOneCall(t, tryAll, noError, failErr, closers, anotherCloser, mc, 2)

	// Fourth call (a call to the successfully closed mc) to Close:
	testMultiCloserOneCall(t, tryAll, noError, failErr, closers, anotherCloser, mc, 3)
}

func testMultiCloserOneCall(t *testing.T, tryAll, noError bool, failErr error, closers []stdio.Closer, anotherCloser stdio.Closer, mc MultiCloser, callNo int) {
	err := mc.Close()
	var wantedErr error
	var wantedClosed bool
	callString := "a failed call"
	switch callNo {
	case 0:
		if tryAll {
			el, ok := err.(errors.ErrorList)
			if !ok {
				t.Errorf("The %d call to mc.Close returns %v, not a ErrorList.",
					callNo, err)
			}
			errs := el.ToList()
			if len(errs) != 2 || errs[0] != failErr || errs[1] != failErr {
				t.Errorf("The %d call to mc.Close returns %v != [%v, %[3]v].",
					callNo, err, failErr)
			}
		} else {
			wantedErr = failErr
		}
	case 1:
		wantedErr = failErr
	case 2:
		wantedClosed = true
		callString = "a successful call"
	default:
		if !noError {
			wantedErr = ErrClosed
		}
		wantedClosed = true
		callString = "a call to the successfully closed mc"
	}
	if (!tryAll || callNo > 0) && !errors.Is(err, wantedErr) {
		t.Errorf("The %d call to mc.Close returns %v != %v. tryAll: %t, noError: %t.",
			callNo, err, wantedErr, tryAll, noError)
	}
	if mc.Closed() != wantedClosed {
		t.Errorf("mc.Closed returns %t after the %d call (%s) to Close. tryAll: %t, noError: %t.",
			!wantedClosed, callNo, callString, tryAll, noError)
	}

	if tryAll {
		for i, closer := range closers {
			wantedClosed := closer.(*testCloser).err == nil
			closed, ok := mc.CloserClosed(closer)
			if !ok {
				t.Errorf("The %d closer is in mc but mc.CloserClosed returns (%t, %t). tryAll: %t, noError: %t.",
					i, closed, ok, tryAll, noError)
			}
			if closed != wantedClosed {
				t.Errorf("mc.CloserClosed for the %d closer is %t != %t. tryAll: %t, noError: %t.",
					i, closed, wantedClosed, tryAll, noError)
			}
		}
	} else {
		var idx int
		for idx = len(closers) - 1; idx >= 0; idx-- {
			if closers[idx].(*testCloser).err != nil {
				break
			}
		}
		for i, closer := range closers {
			wantedClosed := i > idx
			closed, ok := mc.CloserClosed(closer)
			if !ok {
				t.Errorf("The %d closer is in mc but mc.CloserClosed returns (%t, %t). tryAll: %t, noError: %t.",
					i, closed, ok, tryAll, noError)
			}
			if closed != wantedClosed {
				t.Errorf("mc.CloserClosed for the %d closer is %t != %t. tryAll: %t, noError: %t.",
					i, closed, wantedClosed, tryAll, noError)
			}
		}
	}

	closed, ok := mc.CloserClosed(anotherCloser)
	if closed || ok {
		t.Errorf("mc.CloserClosed returns (%t, %t) for anotherCloser. tryAll: %t, noError: %t.",
			closed, ok, tryAll, noError)
	}
}
