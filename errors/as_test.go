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

package errors_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/donyori/gogo/errors"
)

// MyError is a trivial implementation of interface error
// for testing errors.As.
type MyError string

func (err MyError) Error() string {
	return string(err)
}

func TestAs(t *testing.T) {
	newErr := errors.New("newError")
	myErr := MyError("myError")
	testCases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{newErr, false},
		{myErr, true},
		{errors.Join(newErr, myErr), true},
		{fmt.Errorf("fmtErrorf%w", newErr), false},
		{fmt.Errorf("fmtErrorf%w", myErr), true},
		{fmt.Errorf("fmtErrorf%w%w", newErr, myErr), true},
		{errors.AutoWrap(newErr), false},
		{errors.AutoWrap(myErr), true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("err=%v", tc.err), func(t *testing.T) {
			var target MyError
			got := errors.As(tc.err, &target)
			if got != tc.want {
				t.Errorf("got %t; want %t", got, tc.want)
			}
			if tc.want {
				if target != myErr {
					t.Errorf("got target %v; want %v", target, myErr)
				}
			} else if target != "" {
				t.Error("target was modified unexpectedly; target:", target)
			}
		})
	}
}

func TestAs_PanicForErrorPointer(t *testing.T) {
	target := new(error)
	err := errors.New("test error")
	defer func() {
		e := recover()
		if e == nil {
			t.Error("want panic but not")
			return
		}
		s, ok := e.(string)
		if !ok || !strings.HasSuffix(s,
			"target is of type *error; As always returns true for that") {
			t.Error("panic -", e)
		}
	}()
	errors.As(err, target)
}
