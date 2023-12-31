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

package inout_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/inout"
)

type testCloser struct {
	tb     testing.TB
	closed bool
	err    error // Test can set err to simulate a failed call to Close.
}

func (tc *testCloser) Close() error {
	if tc.closed {
		msg := "call testCloser.Close again" // don't use *ClosedError here
		tc.tb.Error(msg)
		return errors.AutoNew(msg)
	}
	tc.closed = tc.err == nil
	return tc.err
}

func TestNoOpCloser(t *testing.T) {
	noc := inout.NewNoOpCloser()
	if noc.Closed() {
		t.Error("noc.Closed was true before first call to Close")
	}
	for i := 1; i <= 10; i++ {
		err := noc.Close()
		if err != nil {
			t.Errorf("error on %d call to Close: %v", i, err)
		}
		if !noc.Closed() {
			t.Errorf("noc.Closed was false after %d call to Close", i)
		}
	}
}

func TestNoErrorCloser(t *testing.T) {
	nec := inout.WrapNoErrorCloser(&testCloser{tb: t})
	if nec.Closed() {
		t.Error("nec.Closed was true before first call to Close")
	}
	for i := 1; i <= 10; i++ {
		err := nec.Close()
		if err != nil {
			t.Errorf("error on %d call to Close: %v", i, err)
		}
		if !nec.Closed() {
			t.Errorf("nec.Closed was false after %d call to Close", i)
		}
	}
}

func TestErrorCloser(t *testing.T) {
	ec := inout.WrapErrorCloser(&testCloser{tb: t}, "testCloser", nil)
	if ec.Closed() {
		t.Error("ec.Closed was true before first call to Close")
	}
	err := ec.Close()
	if err != nil {
		t.Error("error on first call to Close:", err)
	}
	if !ec.Closed() {
		t.Error("ec.Closed was false after first call to Close")
	}
	for i := 2; i <= 10; i++ {
		err = ec.Close()
		if err == nil {
			t.Errorf("no error on %d call to Close, but want a *ClosedError", i)
		} else if !errors.Is(err, inout.ErrClosed) {
			t.Errorf("error on %d call to Close: %v", i, err)
		}
		if !ec.Closed() {
			t.Errorf("ec.Closed was false after %d call to Close", i)
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
	t.Run(fmt.Sprintf("tryAll=%t&noError=%t", tryAll, noError), func(t *testing.T) {
		failErr := errors.New("an error to simulate a failed call to Close")
		closers := []io.Closer{
			&testCloser{tb: t, err: failErr},
			&testCloser{tb: t},
			&testCloser{tb: t, err: failErr},
			&testCloser{tb: t},
		}
		anotherCloser := &testCloser{tb: t}
		mc := inout.NewMultiCloser(tryAll, noError, closers...)

		// Before the first call to Close:
		if mc.Closed() {
			t.Error("mc.Closed was true before the first call to mc.Close")
		}
		for i, closer := range closers {
			closed, ok := mc.CloserClosed(closer)
			if !ok {
				t.Errorf("the %d closer is in mc but mc.CloserClosed returned (%t, %t)", i, closed, ok)
			}
			if closed {
				t.Errorf("mc.CloserClosed for the %d closer was true before the first call to mc.Close", i)
			}
		}
		closed, ok := mc.CloserClosed(anotherCloser)
		if closed || ok {
			t.Errorf("mc.CloserClosed returned (%t, %t) for anotherCloser", closed, ok)
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
	})
}

func testMultiCloserOneCall(
	t *testing.T,
	tryAll bool,
	noError bool,
	failErr error,
	closers []io.Closer,
	anotherCloser io.Closer,
	mc inout.MultiCloser,
	callNo int,
) {
	t.Run(fmt.Sprintf("callNo=%d", callNo), func(t *testing.T) {
		testMultiCloserOneCallClose(t, tryAll, noError, failErr, mc, callNo)
		testMultiCloserOneCallCloserClosed(
			t, tryAll, closers, anotherCloser, mc)
	})
}

// testMultiCloserOneCallClose is a subprocess of testMultiCloserOneCall
// to test inout.MultiCloser.Close.
func testMultiCloserOneCallClose(
	t *testing.T,
	tryAll bool,
	noError bool,
	failErr error,
	mc inout.MultiCloser,
	callNo int,
) {
	err := mc.Close()
	var wantErr error
	var wantClosed bool
	callString := "a failed call"
	switch callNo {
	case 0:
		if tryAll {
			el, ok := err.(errors.ErrorList)
			if !ok {
				t.Errorf("mc.Close returned %v (of type %[1]v), not an ErrorList",
					err)
			}
			errs := el.ToList()
			if len(errs) != 2 || errs[0] != failErr || errs[1] != failErr {
				t.Errorf("mc.Close returned %v; want (%v, %[2]v)",
					err, failErr)
			}
		} else {
			wantErr = failErr
		}
	case 1:
		wantErr = failErr
	case 2:
		wantClosed = true
		callString = "a successful call"
	default:
		if !noError {
			wantErr = inout.ErrClosed
		}
		wantClosed = true
		callString = "a call to the successfully closed mc"
	}
	if (!tryAll || callNo > 0) && !errors.Is(err, wantErr) {
		t.Errorf("mc.Close returned %v; want %v", err, wantErr)
	}
	if mc.Closed() != wantClosed {
		t.Errorf("mc.Closed returned %t after the %d call (%s) to Close",
			!wantClosed, callNo, callString)
	}
}

// testMultiCloserOneCallCloserClosed is a subprocess of testMultiCloserOneCall
// to test inout.MultiCloser.CloserClosed after calling Close.
func testMultiCloserOneCallCloserClosed(
	t *testing.T,
	tryAll bool,
	closers []io.Closer,
	anotherCloser io.Closer,
	mc inout.MultiCloser,
) {
	if tryAll {
		for i, closer := range closers {
			wantClosed := closer.(*testCloser).err == nil
			closed, ok := mc.CloserClosed(closer)
			if !ok {
				t.Errorf("the %d closer is in mc but mc.CloserClosed returned (%t, %t)",
					i, closed, ok)
			}
			if closed != wantClosed {
				t.Errorf("mc.CloserClosed for the %d closer was %t; want %t",
					i, closed, wantClosed)
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
			wantClosed := i > idx
			closed, ok := mc.CloserClosed(closer)
			if !ok {
				t.Errorf("the %d closer is in mc but mc.CloserClosed returned (%t, %t)",
					i, closed, ok)
			}
			if closed != wantClosed {
				t.Errorf("mc.CloserClosed for the %d closer was %t; want %t",
					i, closed, wantClosed)
			}
		}
	}

	closed, ok := mc.CloserClosed(anotherCloser)
	if closed || ok {
		t.Errorf("mc.CloserClosed returned (%t, %t) for anotherCloser",
			closed, ok)
	}
}
