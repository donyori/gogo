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
	"testing"

	"github.com/donyori/gogo/errors"
)

type testCloser struct {
	tb     testing.TB
	closed bool
}

func (tc *testCloser) Close() error {
	if tc.closed {
		tc.tb.Error("Call testCloser.Close again.")
		return errors.AutoNew("call testCloser.Close again") // Don't use *ClosedError here.
	}
	tc.closed = true
	return nil
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
