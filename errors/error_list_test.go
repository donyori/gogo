// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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

var testErrors []error // It will be set in function init.

func init() {
	testErrors = make([]error, 3)
	for i := range testErrors {
		testErrors[i] = fmt.Errorf("test error %d", i)
	}
}

func TestNewErrorList(t *testing.T) {
	// Only test ErrorList with ignoreNil enabled.
	cases := []struct {
		errs   []error
		wanted []error
	}{
		{nil, []error{}},
		{[]error{}, []error{}},
		{[]error{nil}, []error{}},
		{[]error{testErrors[0]}, []error{testErrors[0]}},
		{append(testErrors[:0:0], testErrors...), testErrors},
		{[]error{testErrors[0], testErrors[0]}, []error{testErrors[0], testErrors[0]}},
		{[]error{testErrors[0], errors.New(testErrors[0].Error())}, nil}, // this case, wanted will be set later
		{[]error{testErrors[0], nil}, testErrors[:1]},
		{[]error{nil, testErrors[0]}, testErrors[:1]},
		{append(testErrors, testErrors...), append(testErrors, testErrors...)},
		{append([]error{nil}, testErrors...), testErrors},
		{append(testErrors, nil), testErrors},
		{append(testErrors[:2:2], append([]error{nil}, testErrors[2:]...)...), testErrors},
	}
	cases[6].wanted = append([]error{}, cases[6].errs...)
	for _, c := range cases {
		el := NewErrorList(true, c.errs...)
		if el == nil {
			t.Errorf("NewErrorList returns a nil error list, errs: %v.", c.errs)
		} else {
			e := el.(*errorList)
			if errorsNotEqual(e.list, c.wanted) {
				t.Errorf("NewErrorList: %v != %v, errs: %v.", e.list, c.wanted, c.errs)
			}
		}
	}
}

func TestErrorList_Append(t *testing.T) {
	// Only test ErrorList with ignoreNil enabled.
	cases := []struct {
		errs   []error
		wanted []error
	}{
		{nil, nil},
		{[]error{}, nil},
		{[]error{nil}, nil},
		{[]error{testErrors[0]}, []error{testErrors[0]}},
		{append(testErrors[:0:0], testErrors...), testErrors},
		{[]error{testErrors[0], testErrors[0]}, []error{testErrors[0], testErrors[0]}},
		{[]error{testErrors[0], errors.New(testErrors[0].Error())}, nil}, // this case, wanted will be set later
		{[]error{testErrors[0], nil}, testErrors[:1]},
		{[]error{nil, testErrors[0]}, testErrors[:1]},
		{append(testErrors, testErrors...), append(testErrors, testErrors...)},
		{append([]error{nil}, testErrors...), testErrors},
		{append(testErrors, nil), testErrors},
		{append(testErrors[:2:2], append([]error{nil}, testErrors[2:]...)...), testErrors},
	}
	cases[6].wanted = cases[6].errs
	for _, c := range cases {
		el := NewErrorList(true).(*errorList)
		el.Append(c.errs...)
		if errorsNotEqual(el.list, c.wanted) {
			t.Errorf("Append: %v != %v, errs: %v.", el.list, c.wanted, c.errs)
		}
	}
}

func TestErrorList_ToError(t *testing.T) {
	cases := []struct {
		el  ErrorList
		err error
	}{
		{NewErrorList(true), nil},
		{NewErrorList(true, []error{}...), nil},
		{NewErrorList(true, nil), nil},
		{NewErrorList(true, testErrors[0]), testErrors[0]},
		{NewErrorList(true, testErrors...), nil}, // this case, err will be set later
	}
	cases[4].err = cases[4].el
	for _, c := range cases {
		err := c.el.ToError()
		if !reflect.DeepEqual(err, c.err) {
			t.Errorf("el.ToError(): %v != %v, el: %v.", err, c.err, c.el)
		}
	}
}

func TestErrorList_Deduplicate(t *testing.T) {
	cases := []struct {
		errs   []error
		wanted []error
	}{
		{nil, nil},
		{[]error{}, nil},
		{[]error{nil}, nil},
		{[]error{testErrors[0]}, testErrors[:1]},
		{append(testErrors[:0:0], testErrors...), testErrors},
		{[]error{testErrors[0], testErrors[0]}, testErrors[:1]},
		{[]error{testErrors[0], errors.New(testErrors[0].Error())}, testErrors[:1]},
		{[]error{testErrors[0], nil}, testErrors[:1]},
		{[]error{nil, testErrors[0]}, testErrors[:1]},
		{append(testErrors, testErrors...), testErrors},
		{append([]error{nil}, testErrors...), testErrors},
		{append(testErrors, nil), testErrors},
		{append(testErrors[:2:2], append([]error{nil}, testErrors[2:]...)...), testErrors},
	}
	for _, c := range cases {
		el := NewErrorList(false, c.errs...)
		el.Deduplicate()
		if errorsNotEqual(el.(*errorList).list, c.wanted) {
			t.Errorf("el after deduplicate: %v != %v, errs: %v.", el, c.wanted, c.errs)
		}
	}
}

func TestErrorList_Error(t *testing.T) {
	cases := []struct {
		errs []error
		s    string
	}{
		{nil, "no error"},
		{[]error{}, "no error"},
		{[]error{nil}, fmt.Sprintf("%v", error(nil))},
		{[]error{testErrors[0]}, fmt.Sprintf("%v", testErrors[0])},
		{append(testErrors[:0:0], testErrors...), ""},
		{append(testErrors, nil), ""},
		{append([]error{nil}, testErrors...), ""},
		{append(testErrors[:2:2], append([]error{nil}, testErrors[2:]...)...), ""},
	}
	for i := 4; i < len(cases); i++ {
		strs := make([]string, len(cases[i].errs))
		for j, err := range cases[i].errs {
			if err != nil {
				strs[j] = fmt.Sprintf("%q", cases[i].errs[j].Error())
			} else {
				strs[j] = `"<nil>"`
			}
		}
		cases[i].s = fmt.Sprintf("%d errors: [%s]", len(cases[i].errs), strings.Join(strs, ", "))
	}
	for _, c := range cases {
		el := NewErrorList(false, c.errs...)
		s := el.Error()
		if s != c.s {
			t.Errorf("el.Error(): %q != %q, errs: %v.", s, c.s, c.errs)
		}
	}
}

func TestErrorList_String(t *testing.T) {
	cases := []struct {
		errs []error
		s    string
	}{
		{nil, "no error"},
		{[]error{}, "no error"},
		{[]error{nil}, fmt.Sprintf("%v", error(nil))},
		{[]error{testErrors[0]}, fmt.Sprintf("%v", testErrors[0])},
		{append(testErrors[:0:0], testErrors...), ""},
		{append(testErrors, nil), ""},
		{append([]error{nil}, testErrors...), ""},
		{append(testErrors[:2:2], append([]error{nil}, testErrors[2:]...)...), ""},
	}
	for i := 4; i < len(cases); i++ {
		strs := make([]string, len(cases[i].errs))
		for j, err := range cases[i].errs {
			if err != nil {
				strs[j] = fmt.Sprintf("%q", cases[i].errs[j].Error())
			} else {
				strs[j] = `"<nil>"`
			}
		}
		cases[i].s = fmt.Sprintf("%d errors: [%s]", len(cases[i].errs), strings.Join(strs, ", "))
	}
	t.Log("String() example:\n" + cases[len(cases)-1].s)
	for _, c := range cases {
		el := NewErrorList(false, c.errs...)
		s := el.String()
		if s != c.s {
			t.Errorf("el.String(): %q != %q, errs: %v.", s, c.s, c.errs)
		}
	}
}

func TestCombine(t *testing.T) {
	cases := []struct {
		errs   []error
		wanted []error
	}{
		{nil, nil},
		{[]error{}, nil},
		{[]error{nil}, nil},
		{[]error{testErrors[0]}, []error{testErrors[0]}},
		{append(testErrors[:0:0], testErrors...), testErrors},
		{[]error{testErrors[0], testErrors[0]}, []error{testErrors[0], testErrors[0]}},
		{[]error{testErrors[0], errors.New(testErrors[0].Error())}, nil}, // this case, wanted will be set later
		{[]error{testErrors[0], nil}, testErrors[:1]},
		{[]error{nil, testErrors[0]}, testErrors[:1]},
		{append(testErrors, testErrors...), append(testErrors, testErrors...)},
		{append([]error{nil}, testErrors...), testErrors},
		{append(testErrors, nil), testErrors},
		{append(testErrors[:2:2], append([]error{nil}, testErrors[2:]...)...), testErrors},
	}
	cases[6].wanted = cases[6].errs
	for _, c := range cases {
		err := Combine(c.errs...)
		if el, ok := err.(*errorList); ok {
			if errorsNotEqual(el.list, c.wanted) {
				t.Errorf("Combine: %v != %v, errs: %v.", el.list, c.wanted, c.errs)
			}
		} else if len(c.wanted) > 1 {
			t.Errorf("Combine: %v, want multiple errors in a list: %v, errs: %v.", err, c.wanted, c.errs)
		} else if len(c.wanted) == 1 {
			if err != c.wanted[0] {
				t.Errorf("Combine: %v != %v. errs: %v.", err, c.wanted[0], c.errs)
			}
		} else {
			if err != nil {
				t.Errorf("Combine: %v != nil. errs: %v.", err, c.errs)
			}
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
