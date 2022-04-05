// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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
	stderrors "errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

var testErrorsForErrorList []error // It will be set in function init.

func init() {
	testErrorsForErrorList = make([]error, 3)
	for i := range testErrorsForErrorList {
		testErrorsForErrorList[i] = fmt.Errorf("test error %d", i)
	}
}

func TestNewErrorList(t *testing.T) {
	// Only test ErrorList with ignoreNil enabled.
	testCases := []struct {
		errs []error
		want []error
	}{
		{nil, []error{}},
		{[]error{}, []error{}},
		{[]error{nil}, []error{}},
		{[]error{testErrorsForErrorList[0]}, []error{testErrorsForErrorList[0]}},
		{append(testErrorsForErrorList[:0:0], testErrorsForErrorList...), testErrorsForErrorList},
		{[]error{testErrorsForErrorList[0], testErrorsForErrorList[0]}, []error{testErrorsForErrorList[0], testErrorsForErrorList[0]}},
		{[]error{testErrorsForErrorList[0], stderrors.New(testErrorsForErrorList[0].Error())}, nil}, // this case, want will be set later
		{[]error{testErrorsForErrorList[0], nil}, testErrorsForErrorList[:1]},
		{[]error{nil, testErrorsForErrorList[0]}, testErrorsForErrorList[:1]},
		{append(testErrorsForErrorList, testErrorsForErrorList...), append(testErrorsForErrorList, testErrorsForErrorList...)},
		{append([]error{nil}, testErrorsForErrorList...), testErrorsForErrorList},
		{append(testErrorsForErrorList, nil), testErrorsForErrorList},
		{append(testErrorsForErrorList[:2:2], append([]error{nil}, testErrorsForErrorList[2:]...)...), testErrorsForErrorList},
	}
	testCases[6].want = append([]error{}, testCases[6].errs...)

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?errs=%#v", i, tc.errs), func(t *testing.T) {
			el := NewErrorList(true, tc.errs...)
			if el == nil {
				t.Error("got a nil error list")
			} else {
				e := el.(*errorList)
				if errorsNotEqual(e.list, tc.want) {
					t.Errorf("got %v; want %v", e.list, tc.want)
				}
			}
		})
	}
}

func TestErrorList_Append(t *testing.T) {
	// Only test ErrorList with ignoreNil enabled.
	testCases := []struct {
		errs []error
		want []error
	}{
		{nil, nil},
		{[]error{}, nil},
		{[]error{nil}, nil},
		{[]error{testErrorsForErrorList[0]}, []error{testErrorsForErrorList[0]}},
		{append(testErrorsForErrorList[:0:0], testErrorsForErrorList...), testErrorsForErrorList},
		{[]error{testErrorsForErrorList[0], testErrorsForErrorList[0]}, []error{testErrorsForErrorList[0], testErrorsForErrorList[0]}},
		{[]error{testErrorsForErrorList[0], stderrors.New(testErrorsForErrorList[0].Error())}, nil}, // this case, want will be set later
		{[]error{testErrorsForErrorList[0], nil}, testErrorsForErrorList[:1]},
		{[]error{nil, testErrorsForErrorList[0]}, testErrorsForErrorList[:1]},
		{append(testErrorsForErrorList, testErrorsForErrorList...), append(testErrorsForErrorList, testErrorsForErrorList...)},
		{append([]error{nil}, testErrorsForErrorList...), testErrorsForErrorList},
		{append(testErrorsForErrorList, nil), testErrorsForErrorList},
		{append(testErrorsForErrorList[:2:2], append([]error{nil}, testErrorsForErrorList[2:]...)...), testErrorsForErrorList},
	}
	testCases[6].want = testCases[6].errs

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?errs=%#v", i, tc.errs), func(t *testing.T) {
			el := NewErrorList(true).(*errorList)
			el.Append(tc.errs...)
			if errorsNotEqual(el.list, tc.want) {
				t.Errorf("got %v; want %v", el.list, tc.want)
			}
		})
	}
}

func TestErrorList_ToError(t *testing.T) {
	testCases := []struct {
		el  *errorList
		err error
	}{
		{NewErrorList(true).(*errorList), nil},
		{NewErrorList(true, []error{}...).(*errorList), nil},
		{NewErrorList(true, nil).(*errorList), nil},
		{NewErrorList(true, testErrorsForErrorList[0]).(*errorList), testErrorsForErrorList[0]},
		{NewErrorList(true, testErrorsForErrorList...).(*errorList), nil}, // this case, err will be set later
	}
	testCases[4].err = testCases[4].el

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?errs=%#v", i, tc.el.list), func(t *testing.T) {
			err := tc.el.ToError()
			if !reflect.DeepEqual(err, tc.err) {
				t.Errorf("got %v; want %v", err, tc.err)
			}
		})
	}
}

func TestErrorList_Deduplicate(t *testing.T) {
	testCases := []struct {
		errs []error
		want []error
	}{
		{nil, nil},
		{[]error{}, nil},
		{[]error{nil}, nil},
		{[]error{testErrorsForErrorList[0]}, testErrorsForErrorList[:1]},
		{append(testErrorsForErrorList[:0:0], testErrorsForErrorList...), testErrorsForErrorList},
		{[]error{testErrorsForErrorList[0], testErrorsForErrorList[0]}, testErrorsForErrorList[:1]},
		{[]error{testErrorsForErrorList[0], stderrors.New(testErrorsForErrorList[0].Error())}, testErrorsForErrorList[:1]},
		{[]error{testErrorsForErrorList[0], nil}, testErrorsForErrorList[:1]},
		{[]error{nil, testErrorsForErrorList[0]}, testErrorsForErrorList[:1]},
		{append(testErrorsForErrorList, testErrorsForErrorList...), testErrorsForErrorList},
		{append([]error{nil}, testErrorsForErrorList...), testErrorsForErrorList},
		{append(testErrorsForErrorList, nil), testErrorsForErrorList},
		{append(testErrorsForErrorList[:2:2], append([]error{nil}, testErrorsForErrorList[2:]...)...), testErrorsForErrorList},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?errs=%#v", i, tc.errs), func(t *testing.T) {
			el := NewErrorList(false, tc.errs...)
			el.Deduplicate()
			if errorsNotEqual(el.(*errorList).list, tc.want) {
				t.Errorf("el after deduplicate: %v; want %v", el, tc.want)
			}
		})
	}
}

func TestErrorList_Error(t *testing.T) {
	testCases := []struct {
		errs []error
		want string
	}{
		{nil, "no error"},
		{[]error{}, "no error"},
		{[]error{nil}, fmt.Sprintf("%v", error(nil))},
		{[]error{testErrorsForErrorList[0]}, fmt.Sprintf("%v", testErrorsForErrorList[0])},
		{append(testErrorsForErrorList[:0:0], testErrorsForErrorList...), ""},
		{append(testErrorsForErrorList, nil), ""},
		{append([]error{nil}, testErrorsForErrorList...), ""},
		{append(testErrorsForErrorList[:2:2], append([]error{nil}, testErrorsForErrorList[2:]...)...), ""},
	}
	for i := 4; i < len(testCases); i++ {
		strs := make([]string, len(testCases[i].errs))
		for j, err := range testCases[i].errs {
			if err != nil {
				strs[j] = fmt.Sprintf("%q", testCases[i].errs[j].Error())
			} else {
				strs[j] = `"<nil>"`
			}
		}
		testCases[i].want = fmt.Sprintf("%d errors: [%s]", len(testCases[i].errs), strings.Join(strs, ", "))
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?errs=%#v", i, tc.errs), func(t *testing.T) {
			el := NewErrorList(false, tc.errs...)
			s := el.Error()
			if s != tc.want {
				t.Errorf("got %q; want %q", s, tc.want)
			}
		})
	}
}

func TestCombine(t *testing.T) {
	testCases := []struct {
		errs []error
		want []error
	}{
		{nil, nil},
		{[]error{}, nil},
		{[]error{nil}, nil},
		{[]error{testErrorsForErrorList[0]}, []error{testErrorsForErrorList[0]}},
		{append(testErrorsForErrorList[:0:0], testErrorsForErrorList...), testErrorsForErrorList},
		{[]error{testErrorsForErrorList[0], testErrorsForErrorList[0]}, []error{testErrorsForErrorList[0], testErrorsForErrorList[0]}},
		{[]error{testErrorsForErrorList[0], stderrors.New(testErrorsForErrorList[0].Error())}, nil}, // this case, want will be set later
		{[]error{testErrorsForErrorList[0], nil}, testErrorsForErrorList[:1]},
		{[]error{nil, testErrorsForErrorList[0]}, testErrorsForErrorList[:1]},
		{append(testErrorsForErrorList, testErrorsForErrorList...), append(testErrorsForErrorList, testErrorsForErrorList...)},
		{append([]error{nil}, testErrorsForErrorList...), testErrorsForErrorList},
		{append(testErrorsForErrorList, nil), testErrorsForErrorList},
		{append(testErrorsForErrorList[:2:2], append([]error{nil}, testErrorsForErrorList[2:]...)...), testErrorsForErrorList},
	}
	testCases[6].want = testCases[6].errs

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?errs=%#v", i, tc.errs), func(t *testing.T) {
			err := Combine(tc.errs...)
			if el, ok := err.(*errorList); ok {
				if errorsNotEqual(el.list, tc.want) {
					t.Errorf("got %v; want %v", el.list, tc.want)
				}
			} else if len(tc.want) > 1 {
				t.Errorf("got %v; want multiple errors in a list: %v", err, tc.want)
			} else if len(tc.want) == 1 {
				if err != tc.want[0] {
					t.Errorf("got %v; want %v", err, tc.want[0])
				}
			} else {
				if err != nil {
					t.Errorf("got %v; want nil", err)
				}
			}
		})
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
