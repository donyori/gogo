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

package errors

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

var testErrors []error

func init() {
	testErrors = make([]error, 3)
	for i := range testErrors {
		testErrors[i] = fmt.Errorf("test error %d", i)
	}
}

func TestCombine(t *testing.T) {
	cases := []struct {
		errs   []error
		wanted ErrorList
	}{
		{nil, nil},
		{[]error{}, ErrorList{}},
		{[]error{nil}, ErrorList{}},
		{[]error{testErrors[0]}, []error{testErrors[0]}},
		{append(testErrors[:0:0], testErrors...), testErrors},
		{[]error{testErrors[0], testErrors[0]}, []error{testErrors[0], testErrors[0]}},
		{[]error{testErrors[0], errors.New(testErrors[0].Error())}, nil},
		{[]error{testErrors[0], nil}, testErrors[:1]},
		{[]error{nil, testErrors[0]}, testErrors[:1]},
		{append(testErrors, testErrors...), append(testErrors, testErrors...)},
		{append([]error{nil}, testErrors...), testErrors},
		{append(testErrors, nil), testErrors},
		{append(testErrors[:2:2], append([]error{nil}, testErrors[2:]...)...), testErrors},
	}
	cases[6].wanted = cases[6].errs
	for _, c := range cases {
		el := Combine(c.errs...)
		if errorsNotEqual(el, c.wanted) {
			t.Errorf("Combine(): %v != %v, errs: %v.", el, c.wanted, c.errs)
		}
	}
}

func TestErrorList_Len(t *testing.T) {
	var el ErrorList
	for _, err := range append(testErrors, nil) {
		if n := el.Len(); n != len(el) {
			t.Errorf("el.Len(): %d != %d, el: %v.", n, len(el), el)
		}
		el = append(el, err)
	}
	if n := el.Len(); n != len(el) {
		t.Errorf("el.Len(): %d != %d, el: %v.", n, len(el), el)
	}
}

func TestErrorList_Append(t *testing.T) {
	cases := []struct {
		errs   []error
		wanted ErrorList
	}{
		{nil, nil},
		{[]error{}, ErrorList{}},
		{[]error{nil}, ErrorList{}},
		{[]error{testErrors[0]}, []error{testErrors[0]}},
		{append(testErrors[:0:0], testErrors...), testErrors},
		{[]error{testErrors[0], testErrors[0]}, []error{testErrors[0], testErrors[0]}},
		{[]error{testErrors[0], errors.New(testErrors[0].Error())}, nil},
		{[]error{testErrors[0], nil}, testErrors[:1]},
		{[]error{nil, testErrors[0]}, testErrors[:1]},
		{append(testErrors, testErrors...), append(testErrors, testErrors...)},
		{append([]error{nil}, testErrors...), testErrors},
		{append(testErrors, nil), testErrors},
		{append(testErrors[:2:2], append([]error{nil}, testErrors[2:]...)...), testErrors},
	}
	cases[6].wanted = cases[6].errs
	for _, c := range cases {
		var el ErrorList
		el.Append(c.errs...)
		if errorsNotEqual(el, c.wanted) {
			t.Errorf("Combine(): %v != %v, errs: %v.", el, c.wanted, c.errs)
		}
		el = nil
		for _, err := range c.errs {
			el.Append(err)
		}
		if errorsNotEqual(el, c.wanted) {
			t.Errorf("Combine(): %v != %v, errs: %v.", el, c.wanted, c.errs)
		}
	}
}

func TestErrorList_ToError(t *testing.T) {
	cases := []struct {
		el  ErrorList
		err error
	}{
		{nil, nil},
		{ErrorList{}, nil},
		{ErrorList{nil}, nil},
		{ErrorList{testErrors[0]}, testErrors[0]},
		{append(testErrors[:0:0], testErrors...), nil},
	}
	cases[4].err = &cases[4].el
	for _, c := range cases {
		err := c.el.ToError()
		if !reflect.DeepEqual(err, c.err) {
			t.Errorf("el.ToError(): %v != %v, el: %v.", err, c.err, c.el)
		}
	}
}

func TestErrorList_Deduplicate(t *testing.T) {
	cases := []struct {
		el     ErrorList
		wanted []error
	}{
		{nil, nil},
		{ErrorList{}, ErrorList{}},
		{ErrorList{nil}, ErrorList{}},
		{ErrorList{testErrors[0]}, testErrors[:1]},
		{append(testErrors[:0:0], testErrors...), testErrors},
		{ErrorList{testErrors[0], testErrors[0]}, testErrors[:1]},
		{ErrorList{testErrors[0], errors.New(testErrors[0].Error())}, testErrors[:1]},
		{ErrorList{testErrors[0], nil}, testErrors[:1]},
		{ErrorList{nil, testErrors[0]}, testErrors[:1]},
		{append(testErrors, testErrors...), testErrors},
		{append(ErrorList{nil}, testErrors...), testErrors},
		{append(testErrors, nil), testErrors},
		{append(testErrors[:2:2], append([]error{nil}, testErrors[2:]...)...), testErrors},
	}
	for _, c := range cases {
		el := append(c.el[:0:0], c.el...)
		el.Deduplicate()
		if errorsNotEqual(el, c.wanted) {
			t.Errorf("el after deduplicate: %v != %v, el: %v.", el, c.wanted, c.el)
		}
	}
}

func TestErrorList_Error(t *testing.T) {
	cases := []struct {
		el ErrorList
		s  string
	}{
		{nil, "no errors"},
		{ErrorList{}, "no errors"},
		{ErrorList{nil}, fmt.Sprintf("%v", error(nil))},
		{ErrorList{testErrors[0]}, fmt.Sprintf("%v", testErrors[0])},
		{append(testErrors[:0:0], testErrors...), ""},
		{append(testErrors, nil), ""},
		{append(ErrorList{nil}, testErrors...), ""},
		{append(testErrors[:2:2], append([]error{nil}, testErrors[2:]...)...), ""},
	}
	for i := 4; i < len(cases); i++ {
		el := cases[i].el
		strs := make([]string, len(el))
		for j, err := range el {
			if err != nil {
				strs[j] = fmt.Sprintf("%q", el[j].Error())
			} else {
				strs[j] = `"<nil>"`
			}
		}
		cases[i].s = fmt.Sprintf("%d errors: [%s]", len(el), strings.Join(strs, ", "))
	}
	for _, c := range cases {
		s := c.el.Error()
		if s != c.s {
			t.Errorf("el.Error(): %q != %q, el: %v.", s, c.s, c.el)
		}
	}
}

func TestErrorList_Unwrap(t *testing.T) {
	cases := []struct {
		el  ErrorList
		err error
	}{
		{nil, nil},
		{ErrorList{}, nil},
		{ErrorList{nil}, nil},
		{ErrorList{testErrors[0]}, testErrors[0]},
		{append(testErrors[:0:0], testErrors...), nil},
	}
	for _, c := range cases {
		err := c.el.Unwrap()
		if err != c.err {
			t.Errorf("el.Unwrap(): %v != %v, el: %v", err, c.err, c.el)
		}
	}
}

func TestErrorList_String(t *testing.T) {
	cases := []struct {
		el ErrorList
		s  string
	}{
		{nil, "no errors"},
		{ErrorList{}, "no errors"},
		{ErrorList{nil}, fmt.Sprintf("%v", error(nil))},
		{ErrorList{testErrors[0]}, fmt.Sprintf("%v", testErrors[0])},
		{append(testErrors[:0:0], testErrors...), ""},
		{append(testErrors, nil), ""},
		{append(ErrorList{nil}, testErrors...), ""},
		{append(testErrors[:2:2], append([]error{nil}, testErrors[2:]...)...), ""},
	}
	for i := 4; i < len(cases); i++ {
		el := cases[i].el
		strs := make([]string, len(el))
		for j, err := range el {
			if err != nil {
				strs[j] = fmt.Sprintf("%q", el[j].Error())
			} else {
				strs[j] = `"<nil>"`
			}
		}
		cases[i].s = fmt.Sprintf("%d errors: [%s]", len(el), strings.Join(strs, ", "))
	}
	t.Log("String() example:\n" + cases[len(cases)-1].s)
	for _, c := range cases {
		s := c.el.String()
		if s != c.s {
			t.Errorf("el.String(): %q != %q, el: %v.", s, c.s, c.el)
		}
	}
}

func errorsNotEqual(errs1, errs2 []error) bool {
	if len(errs1) != len(errs2) {
		return true
	}
	for i := range errs1 {
		if errs1[i] != errs2[i] { // compare the interface directly, don't use errors.Is
			return true
		}
	}
	return false
}
