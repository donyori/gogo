// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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
	stderrors "errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/donyori/gogo/errors"
)

var errorsForErrorList []error // it is set in function init

func init() {
	errorsForErrorList = make([]error, 3)
	for i := range errorsForErrorList {
		errorsForErrorList[i] = fmt.Errorf("test error %d", i)
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
		{[]error{errorsForErrorList[0]}, []error{errorsForErrorList[0]}},
		{append(errorsForErrorList[:0:0], errorsForErrorList...), errorsForErrorList},
		{[]error{errorsForErrorList[0], errorsForErrorList[0]}, []error{errorsForErrorList[0], errorsForErrorList[0]}},
		{[]error{errorsForErrorList[0], stderrors.New(errorsForErrorList[0].Error())}, nil}, // this case, want will be set later
		{[]error{errorsForErrorList[0], nil}, errorsForErrorList[:1]},
		{[]error{nil, errorsForErrorList[0]}, errorsForErrorList[:1]},
		{append(errorsForErrorList, errorsForErrorList...), append(errorsForErrorList, errorsForErrorList...)},
		{append([]error{nil}, errorsForErrorList...), errorsForErrorList},
		{append(errorsForErrorList, nil), errorsForErrorList},
		{append(errorsForErrorList[:2:2], append([]error{nil}, errorsForErrorList[2:]...)...), errorsForErrorList},
	}
	testCases[6].want = append([]error{}, testCases[6].errs...)

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?errs=%#v", i, tc.errs), func(t *testing.T) {
			el := errors.NewErrorList(true, tc.errs...)
			if el == nil {
				t.Error("got a nil error list")
			} else {
				e := el.(*errors.ErrorListImpl)
				if list := e.GetList(); errorsUnequal(list, tc.want) {
					t.Errorf("got %v; want %v", list, tc.want)
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
		{[]error{errorsForErrorList[0]}, []error{errorsForErrorList[0]}},
		{append(errorsForErrorList[:0:0], errorsForErrorList...), errorsForErrorList},
		{[]error{errorsForErrorList[0], errorsForErrorList[0]}, []error{errorsForErrorList[0], errorsForErrorList[0]}},
		{[]error{errorsForErrorList[0], stderrors.New(errorsForErrorList[0].Error())}, nil}, // this case, want will be set later
		{[]error{errorsForErrorList[0], nil}, errorsForErrorList[:1]},
		{[]error{nil, errorsForErrorList[0]}, errorsForErrorList[:1]},
		{append(errorsForErrorList, errorsForErrorList...), append(errorsForErrorList, errorsForErrorList...)},
		{append([]error{nil}, errorsForErrorList...), errorsForErrorList},
		{append(errorsForErrorList, nil), errorsForErrorList},
		{append(errorsForErrorList[:2:2], append([]error{nil}, errorsForErrorList[2:]...)...), errorsForErrorList},
	}
	testCases[6].want = testCases[6].errs

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?errs=%#v", i, tc.errs), func(t *testing.T) {
			el := errors.NewErrorList(true).(*errors.ErrorListImpl)
			el.Append(tc.errs...)
			if list := el.GetList(); errorsUnequal(list, tc.want) {
				t.Errorf("got %v; want %v", list, tc.want)
			}
		})
	}
}

func TestErrorList_ToError(t *testing.T) {
	testCases := []struct {
		el  *errors.ErrorListImpl
		err error
	}{
		{errors.NewErrorList(true).(*errors.ErrorListImpl), nil},
		{errors.NewErrorList(true, []error{}...).(*errors.ErrorListImpl), nil},
		{errors.NewErrorList(true, nil).(*errors.ErrorListImpl), nil},
		{errors.NewErrorList(true, errorsForErrorList[0]).(*errors.ErrorListImpl), errorsForErrorList[0]},
		{errors.NewErrorList(true, errorsForErrorList...).(*errors.ErrorListImpl), nil}, // this case, err will be set later
	}
	testCases[4].err = testCases[4].el

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case %d?errs=%#v", i, tc.el.GetList()),
			func(t *testing.T) {
				err := tc.el.ToError()
				if err != tc.err {
					t.Errorf("got %v; want %v", err, tc.err)
				}
			},
		)
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
		{[]error{errorsForErrorList[0]}, errorsForErrorList[:1]},
		{append(errorsForErrorList[:0:0], errorsForErrorList...), errorsForErrorList},
		{[]error{errorsForErrorList[0], errorsForErrorList[0]}, errorsForErrorList[:1]},
		{[]error{errorsForErrorList[0], stderrors.New(errorsForErrorList[0].Error())}, errorsForErrorList[:1]},
		{[]error{errorsForErrorList[0], nil}, errorsForErrorList[:1]},
		{[]error{nil, errorsForErrorList[0]}, errorsForErrorList[:1]},
		{append(errorsForErrorList, errorsForErrorList...), errorsForErrorList},
		{append([]error{nil}, errorsForErrorList...), errorsForErrorList},
		{append(errorsForErrorList, nil), errorsForErrorList},
		{append(errorsForErrorList[:2:2], append([]error{nil}, errorsForErrorList[2:]...)...), errorsForErrorList},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?errs=%#v", i, tc.errs), func(t *testing.T) {
			el := errors.NewErrorList(false, tc.errs...)
			el.Deduplicate()
			if errorsUnequal(el.(*errors.ErrorListImpl).GetList(), tc.want) {
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
		{[]error{errorsForErrorList[0]}, fmt.Sprintf("%v", errorsForErrorList[0])},
		{append(errorsForErrorList[:0:0], errorsForErrorList...), ""},
		{append(errorsForErrorList, nil), ""},
		{append([]error{nil}, errorsForErrorList...), ""},
		{append(errorsForErrorList[:2:2], append([]error{nil}, errorsForErrorList[2:]...)...), ""},
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
		testCases[i].want = fmt.Sprintf("%d errors: [%s]",
			len(testCases[i].errs), strings.Join(strs, ", "))
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?errs=%#v", i, tc.errs), func(t *testing.T) {
			el := errors.NewErrorList(false, tc.errs...)
			s := el.Error()
			if s != tc.want {
				t.Errorf("got %q; want %q", s, tc.want)
			}
		})
	}
}

// indexError combines an index and an error,
// for testing github.com/donyori/gogo/errors.ErrorList.
type indexError struct {
	i   int
	err error
}

func TestErrorList_Range(t *testing.T) {
	errs := slices.Clone(errorsForErrorList)
	want := []indexError{{0, errs[0]}, {1, errs[1]}}
	el := errors.NewErrorList(false, errs...)
	gotData := make([]indexError, 0, len(errs))
	el.Range(func(i int, err error) (cont bool) {
		gotData = append(gotData, indexError{i: i, err: err})
		return i == 0
	})
	if !slices.Equal(gotData, want) {
		t.Errorf("got %v; want %v", gotData, want)
	}
}

func TestErrorList_Range_Empty(t *testing.T) {
	el := errors.NewErrorList(false)
	el.Range(func(i int, err error) (cont bool) {
		t.Errorf("handler was called, i: %d, err: %v", i, err)
		return true
	})
}

func TestErrorList_Range_NilHandler(t *testing.T) {
	errs := slices.Clone(errorsForErrorList)
	el := errors.NewErrorList(false, errs...)
	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()
	el.Range(nil)
}

func TestErrorList_RangeBackward(t *testing.T) {
	errs := slices.Clone(errorsForErrorList)
	want := []indexError{{2, errs[2]}, {1, errs[1]}}
	el := errors.NewErrorList(false, errs...)
	gotData := make([]indexError, 0, len(errs))
	el.RangeBackward(func(i int, err error) (cont bool) {
		gotData = append(gotData, indexError{i: i, err: err})
		return i == len(errs)-1
	})
	if !slices.Equal(gotData, want) {
		t.Errorf("got %v; want %v", gotData, want)
	}
}

func TestErrorList_RangeBackward_Empty(t *testing.T) {
	el := errors.NewErrorList(false)
	el.RangeBackward(func(i int, err error) (cont bool) {
		t.Errorf("handler was called, i: %d, err: %v", i, err)
		return true
	})
}

func TestErrorList_RangeBackward_NilHandler(t *testing.T) {
	errs := slices.Clone(errorsForErrorList)
	el := errors.NewErrorList(false, errs...)
	defer func() {
		if e := recover(); e != nil {
			t.Error("panic -", e)
		}
	}()
	el.RangeBackward(nil)
}

func TestErrorList_IterErrors(t *testing.T) {
	errs := slices.Clone(errorsForErrorList)
	want := []error{errs[0], errs[1]}
	el := errors.NewErrorList(false, errs...)
	seq := el.IterErrors()
	if seq == nil {
		t.Fatal("got nil iterator")
	}
	gotData := make([]error, 0, len(errs))
	for err := range seq {
		gotData = append(gotData, err)
		if len(gotData) >= 2 {
			break
		}
	}
	if errorsUnequal(gotData, want) {
		t.Errorf("got %v; want %v", gotData, want)
	}
}

func TestErrorList_IterErrors_Empty(t *testing.T) {
	el := errors.NewErrorList(false)
	seq := el.IterErrors()
	if seq == nil {
		t.Fatal("got nil iterator")
	}
	for err := range seq {
		t.Error("yielded", err)
	}
}

func TestErrorList_IterErrorsBackward(t *testing.T) {
	errs := slices.Clone(errorsForErrorList)
	want := []error{errs[2], errs[1]}
	el := errors.NewErrorList(false, errs...)
	seq := el.IterErrorsBackward()
	if seq == nil {
		t.Fatal("got nil iterator")
	}
	gotData := make([]error, 0, len(errs))
	for err := range seq {
		gotData = append(gotData, err)
		if len(gotData) >= 2 {
			break
		}
	}
	if errorsUnequal(gotData, want) {
		t.Errorf("got %v; want %v", gotData, want)
	}
}

func TestErrorList_IterErrorsBackward_Empty(t *testing.T) {
	el := errors.NewErrorList(false)
	seq := el.IterErrorsBackward()
	if seq == nil {
		t.Fatal("got nil iterator")
	}
	for err := range seq {
		t.Error("yielded", err)
	}
}

func TestErrorList_IterIndexErrors(t *testing.T) {
	errs := slices.Clone(errorsForErrorList)
	want := []indexError{{0, errs[0]}, {1, errs[1]}}
	el := errors.NewErrorList(false, errs...)
	seq2 := el.IterIndexErrors()
	if seq2 == nil {
		t.Fatal("got nil iterator")
	}
	gotData := make([]indexError, 0, len(errs))
	for i, err := range seq2 {
		gotData = append(gotData, indexError{i: i, err: err})
		if len(gotData) >= 2 {
			break
		}
	}
	if !slices.Equal(gotData, want) {
		t.Errorf("got %v; want %v", gotData, want)
	}
}

func TestErrorList_IterIndexErrors_Empty(t *testing.T) {
	el := errors.NewErrorList(false)
	seq2 := el.IterIndexErrors()
	if seq2 == nil {
		t.Fatal("got nil iterator")
	}
	for i, err := range seq2 {
		t.Errorf("yielded %d: %v", i, err)
	}
}

func TestErrorList_IterIndexErrorsBackward(t *testing.T) {
	errs := slices.Clone(errorsForErrorList)
	want := []indexError{{2, errs[2]}, {1, errs[1]}}
	el := errors.NewErrorList(false, errs...)
	seq2 := el.IterIndexErrorsBackward()
	if seq2 == nil {
		t.Fatal("got nil iterator")
	}
	gotData := make([]indexError, 0, len(errs))
	for i, err := range seq2 {
		gotData = append(gotData, indexError{i: i, err: err})
		if len(gotData) >= 2 {
			break
		}
	}
	if !slices.Equal(gotData, want) {
		t.Errorf("got %v; want %v", gotData, want)
	}
}

func TestErrorList_IterIndexErrorsBackward_Empty(t *testing.T) {
	el := errors.NewErrorList(false)
	seq2 := el.IterIndexErrorsBackward()
	if seq2 == nil {
		t.Fatal("got nil iterator")
	}
	for i, err := range seq2 {
		t.Errorf("yielded %d: %v", i, err)
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
		{[]error{errorsForErrorList[0]}, []error{errorsForErrorList[0]}},
		{append(errorsForErrorList[:0:0], errorsForErrorList...), errorsForErrorList},
		{[]error{errorsForErrorList[0], errorsForErrorList[0]}, []error{errorsForErrorList[0], errorsForErrorList[0]}},
		{[]error{errorsForErrorList[0], stderrors.New(errorsForErrorList[0].Error())}, nil}, // this case, want will be set later
		{[]error{errorsForErrorList[0], nil}, errorsForErrorList[:1]},
		{[]error{nil, errorsForErrorList[0]}, errorsForErrorList[:1]},
		{append(errorsForErrorList, errorsForErrorList...), append(errorsForErrorList, errorsForErrorList...)},
		{append([]error{nil}, errorsForErrorList...), errorsForErrorList},
		{append(errorsForErrorList, nil), errorsForErrorList},
		{append(errorsForErrorList[:2:2], append([]error{nil}, errorsForErrorList[2:]...)...), errorsForErrorList},
	}
	testCases[6].want = testCases[6].errs

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?errs=%#v", i, tc.errs), func(t *testing.T) {
			err := errors.Combine(tc.errs...)
			if el, ok := err.(*errors.ErrorListImpl); ok {
				if list := el.GetList(); errorsUnequal(list, tc.want) {
					t.Errorf("got %v; want %v", list, tc.want)
				}
			} else if len(tc.want) > 1 {
				t.Errorf("got %v; want multiple errors in a list: %v",
					err, tc.want)
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

func errorsUnequal(errs1, errs2 []error) bool {
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
