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
	stderrors "errors"
	"fmt"
	"testing"
)

var testErrorsForErrorReadOnlySet [][]error // It will be set in function init.

func init() {
	const N int = 3
	testErrorsForErrorReadOnlySet = make([][]error, 4)
	testErrorsForErrorReadOnlySet[0] = make([]error, N)
	for i := 0; i < N; i++ {
		testErrorsForErrorReadOnlySet[0][i] = fmt.Errorf("test error %d", i)
	}
	testErrorsForErrorReadOnlySet[1] = make([]error, N*2)
	for i := 0; i < N; i++ {
		testErrorsForErrorReadOnlySet[1][i*2] = &testErrorUnwrap{testErrorsForErrorReadOnlySet[0][i]}
		testErrorsForErrorReadOnlySet[1][i*2+1] = &testErrorUnwrap{testErrorsForErrorReadOnlySet[1][i*2]}
	}
	testErrorsForErrorReadOnlySet[2], testErrorsForErrorReadOnlySet[3] = make([]error, N*2), make([]error, N*2)
	for i := 0; i < N; i++ {
		testErrorsForErrorReadOnlySet[2][i*2] = &testErrorIsAlwaysTrue{testErrorsForErrorReadOnlySet[0][i]}
		testErrorsForErrorReadOnlySet[2][i*2+1] = &testErrorIsAlwaysTrue{testErrorsForErrorReadOnlySet[1][i*2]}
		testErrorsForErrorReadOnlySet[3][i*2] = &testErrorIsAlwaysFalse{testErrorsForErrorReadOnlySet[0][i]}
		testErrorsForErrorReadOnlySet[3][i*2+1] = &testErrorIsAlwaysFalse{testErrorsForErrorReadOnlySet[1][i*2]}
	}
}

type testErrorReadOnlySetContainCase struct {
	target error
	want   bool
}

func TestErrorReadOnlySetEqual_Len(t *testing.T) {
	var setErrs []error
	for _, errs := range testErrorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	set := NewErrorReadOnlySetEqual(setErrs...)
	if n := set.Len(); n != len(setErrs) {
		t.Errorf("got %d; want %d", n, len(setErrs))
	}
}

func TestErrorReadOnlySetEqual_Len_Nil(t *testing.T) {
	setErrs := []error{nil, stderrors.New("<nil>")}
	for _, errs := range testErrorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	set := NewErrorReadOnlySetEqual(setErrs...)
	if n := set.Len(); n != len(setErrs) {
		t.Errorf("got %d; want %d", n, len(setErrs))
	}
}

func TestErrorReadOnlySetEqual_Contain(t *testing.T) {
	var setErrs []error
	for _, errs := range testErrorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := make([]testErrorReadOnlySetContainCase, 0, len(setErrs)*5+1)
	for _, err := range setErrs {
		testCases = append(
			testCases,
			testErrorReadOnlySetContainCase{err, true},
			testErrorReadOnlySetContainCase{stderrors.New(err.Error()), false},
			testErrorReadOnlySetContainCase{&testErrorUnwrap{err}, false},
			testErrorReadOnlySetContainCase{&testErrorIsAlwaysTrue{err}, false},
			testErrorReadOnlySetContainCase{&testErrorIsAlwaysFalse{err}, false},
		)
	}
	testCases = append(testCases, testErrorReadOnlySetContainCase{}) // {nil, false}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?target=%q", i, tc.target), func(t *testing.T) {
			set := NewErrorReadOnlySetEqual(setErrs...)
			if set.Contain(tc.target) != tc.want {
				t.Errorf("got %t; want %t", !tc.want, tc.want)
			}
		})
	}
}

func TestErrorReadOnlySetEqual_Contain_Nil(t *testing.T) {
	setErrs := []error{nil}
	for _, errs := range testErrorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := []testErrorReadOnlySetContainCase{
		{nil, true},
		{stderrors.New("<nil>"), false},
		{new(testErrorUnwrap), false},
		{new(testErrorIsAlwaysTrue), false},
		{new(testErrorIsAlwaysFalse), false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?target=%q", i, tc.target), func(t *testing.T) {
			set := NewErrorReadOnlySetEqual(setErrs...)
			if set.Contain(tc.target) != tc.want {
				t.Errorf("got %t; want %t", !tc.want, tc.want)
			}
		})
	}
}

func TestErrorReadOnlySetIs_Len(t *testing.T) {
	var setErrs []error
	for _, errs := range testErrorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	set := NewErrorReadOnlySetIs(setErrs...)
	if n := set.Len(); n != len(setErrs) {
		t.Errorf("got %d; want %d", n, len(setErrs))
	}
}

func TestErrorReadOnlySetIs_Len_Nil(t *testing.T) {
	setErrs := []error{nil, stderrors.New("<nil>")}
	for _, errs := range testErrorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	set := NewErrorReadOnlySetIs(setErrs...)
	if n := set.Len(); n != len(setErrs) {
		t.Errorf("got %d; want %d", n, len(setErrs))
	}
}

func TestErrorReadOnlySetIs_Contain(t *testing.T) {
	var setErrs []error
	for _, errs := range testErrorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := make([]testErrorReadOnlySetContainCase, 0, len(setErrs)*5+4)
	for _, err := range setErrs {
		testCases = append(
			testCases,
			testErrorReadOnlySetContainCase{err, true},
			testErrorReadOnlySetContainCase{stderrors.New(err.Error()), false},
			testErrorReadOnlySetContainCase{&testErrorUnwrap{err}, true},
			testErrorReadOnlySetContainCase{&testErrorIsAlwaysTrue{err}, true},
			// When method Is returns false, errors.Is will continue testing along the Unwrap error chain rather than return false.
			testErrorReadOnlySetContainCase{&testErrorIsAlwaysFalse{err}, true},
		)
	}
	err := stderrors.New("test error +1")
	testCases = append(
		testCases,
		testErrorReadOnlySetContainCase{}, // {nil, false}
		testErrorReadOnlySetContainCase{&testErrorUnwrap{err}, false},
		testErrorReadOnlySetContainCase{&testErrorIsAlwaysTrue{err}, true},
		testErrorReadOnlySetContainCase{&testErrorIsAlwaysFalse{err}, false},
	)

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?target=%q", i, tc.target), func(t *testing.T) {
			set := NewErrorReadOnlySetIs(setErrs...)
			if set.Contain(tc.target) != tc.want {
				t.Errorf("got %t; want %t", !tc.want, tc.want)
			}
		})
	}
}

func TestErrorReadOnlySetIs_Contain_Nil(t *testing.T) {
	setErrs := []error{nil}
	for _, errs := range testErrorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := []testErrorReadOnlySetContainCase{
		{nil, true},
		{stderrors.New("<nil>"), false},
		{new(testErrorUnwrap), false},
		{new(testErrorIsAlwaysTrue), true},
		{new(testErrorIsAlwaysFalse), false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?target=%q", i, tc.target), func(t *testing.T) {
			set := NewErrorReadOnlySetIs(setErrs...)
			if set.Contain(tc.target) != tc.want {
				t.Errorf("got %t; want %t", !tc.want, tc.want)
			}
		})
	}
}

func TestErrorReadOnlySetSameMessage_Len(t *testing.T) {
	var setErrs []error
	for _, errs := range testErrorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	set := NewErrorReadOnlySetSameMessage(setErrs...)
	if n := set.Len(); n != len(setErrs) {
		t.Errorf("got %d; want %d", n, len(setErrs))
	}
}

func TestErrorReadOnlySetSameMessage_Len_Nil(t *testing.T) {
	setErrs := []error{nil, stderrors.New("<nil>")}
	for _, errs := range testErrorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	set := NewErrorReadOnlySetSameMessage(setErrs...)
	if n := set.Len(); n != len(setErrs) {
		t.Errorf("got %d; want %d", n, len(setErrs))
	}
}

func TestErrorReadOnlySetSameMessage_Contain(t *testing.T) {
	var setErrs []error
	msgSet := make(map[string]bool)
	for _, errs := range testErrorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
		for _, err := range errs {
			msgSet[err.Error()] = true
		}
	}
	testCases := make([]testErrorReadOnlySetContainCase, 0, len(setErrs)*5+1)
	for _, err := range setErrs {
		eu, et, ef := &testErrorUnwrap{err}, &testErrorIsAlwaysTrue{err}, &testErrorIsAlwaysFalse{err}
		testCases = append(
			testCases,
			testErrorReadOnlySetContainCase{err, true},
			testErrorReadOnlySetContainCase{stderrors.New(err.Error()), true},
			testErrorReadOnlySetContainCase{eu, msgSet[eu.Error()]},
			testErrorReadOnlySetContainCase{et, msgSet[et.Error()]},
			testErrorReadOnlySetContainCase{ef, msgSet[ef.Error()]},
		)
	}
	testCases = append(testCases, testErrorReadOnlySetContainCase{}) // {nil, false}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?target=%q", i, tc.target), func(t *testing.T) {
			set := NewErrorReadOnlySetSameMessage(setErrs...)
			if set.Contain(tc.target) != tc.want {
				t.Errorf("got %t; want %t", !tc.want, tc.want)
			}
		})
	}
}

func TestErrorReadOnlySetSameMessage_Contain_Nil(t *testing.T) {
	setErrs := []error{nil}
	for _, errs := range testErrorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := []testErrorReadOnlySetContainCase{
		{nil, true},
		{stderrors.New("<nil>"), true},
		{new(testErrorUnwrap), false},
		{new(testErrorIsAlwaysTrue), false},
		{new(testErrorIsAlwaysFalse), false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?target=%q", i, tc.target), func(t *testing.T) {
			set := NewErrorReadOnlySetSameMessage(setErrs...)
			if set.Contain(tc.target) != tc.want {
				t.Errorf("got %t; want %t", !tc.want, tc.want)
			}
		})
	}
}

type testErrorUnwrap struct {
	wrapped error
}

func (teu *testErrorUnwrap) Error() string {
	if teu.wrapped == nil {
		return "test error - <nil>"
	}
	return "test error unwrap - " + teu.wrapped.Error()
}

func (teu *testErrorUnwrap) Unwrap() error {
	return teu.wrapped
}

type testErrorIsAlwaysTrue struct {
	wrapped error
}

func (teiat *testErrorIsAlwaysTrue) Error() string {
	if teiat.wrapped == nil {
		return "test error Is always true - <nil>"
	}
	return "test error Is always true - " + teiat.wrapped.Error()
}

func (teiat *testErrorIsAlwaysTrue) Unwrap() error {
	return teiat.wrapped
}

func (teiat *testErrorIsAlwaysTrue) Is(error) bool {
	return true
}

type testErrorIsAlwaysFalse struct {
	wrapped error
}

func (teiaf *testErrorIsAlwaysFalse) Error() string {
	if teiaf.wrapped == nil {
		return "test error Is always false - <nil>"
	}
	return "test error Is always false - " + teiaf.wrapped.Error()
}

func (teiaf *testErrorIsAlwaysFalse) Unwrap() error {
	return teiaf.wrapped
}

func (teiaf *testErrorIsAlwaysFalse) Is(error) bool {
	return false
}
