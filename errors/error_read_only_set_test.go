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

package errors_test

import (
	stderrors "errors"
	"fmt"
	"testing"

	"github.com/donyori/gogo/errors"
)

var errorsForErrorReadOnlySet [][]error // It will be set in function init.

func init() {
	const N int = 3
	errorsForErrorReadOnlySet = make([][]error, 4)
	errorsForErrorReadOnlySet[0] = make([]error, N)
	for i := 0; i < N; i++ {
		errorsForErrorReadOnlySet[0][i] = fmt.Errorf("test error %d", i)
	}
	errorsForErrorReadOnlySet[1] = make([]error, N*2)
	for i := 0; i < N; i++ {
		errorsForErrorReadOnlySet[1][i*2] = &errorUnwrap{errorsForErrorReadOnlySet[0][i]}
		errorsForErrorReadOnlySet[1][i*2+1] = &errorUnwrap{errorsForErrorReadOnlySet[1][i*2]}
	}
	errorsForErrorReadOnlySet[2], errorsForErrorReadOnlySet[3] = make([]error, N*2), make([]error, N*2)
	for i := 0; i < N; i++ {
		errorsForErrorReadOnlySet[2][i*2] = &errorIsAlwaysTrue{errorsForErrorReadOnlySet[0][i]}
		errorsForErrorReadOnlySet[2][i*2+1] = &errorIsAlwaysTrue{errorsForErrorReadOnlySet[1][i*2]}
		errorsForErrorReadOnlySet[3][i*2] = &errorIsAlwaysFalse{errorsForErrorReadOnlySet[0][i]}
		errorsForErrorReadOnlySet[3][i*2+1] = &errorIsAlwaysFalse{errorsForErrorReadOnlySet[1][i*2]}
	}
}

type errorReadOnlySetContainCase struct {
	target error
	want   bool
}

func TestErrorReadOnlySetEqual_Len(t *testing.T) {
	var setErrs []error
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	set := errors.NewErrorReadOnlySetEqual(setErrs...)
	if n := set.Len(); n != len(setErrs) {
		t.Errorf("got %d; want %d", n, len(setErrs))
	}
}

func TestErrorReadOnlySetEqual_Len_Nil(t *testing.T) {
	setErrs := []error{nil, stderrors.New("<nil>")}
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	set := errors.NewErrorReadOnlySetEqual(setErrs...)
	if n := set.Len(); n != len(setErrs) {
		t.Errorf("got %d; want %d", n, len(setErrs))
	}
}

func TestErrorReadOnlySetEqual_Contain(t *testing.T) {
	var setErrs []error
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := make([]errorReadOnlySetContainCase, 0, len(setErrs)*5+1)
	for _, err := range setErrs {
		testCases = append(
			testCases,
			errorReadOnlySetContainCase{err, true},
			errorReadOnlySetContainCase{stderrors.New(err.Error()), false},
			errorReadOnlySetContainCase{&errorUnwrap{err}, false},
			errorReadOnlySetContainCase{&errorIsAlwaysTrue{err}, false},
			errorReadOnlySetContainCase{&errorIsAlwaysFalse{err}, false},
		)
	}
	testCases = append(testCases, errorReadOnlySetContainCase{}) // {nil, false}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?target=%q", i, tc.target), func(t *testing.T) {
			set := errors.NewErrorReadOnlySetEqual(setErrs...)
			if set.Contain(tc.target) != tc.want {
				t.Errorf("got %t; want %t", !tc.want, tc.want)
			}
		})
	}
}

func TestErrorReadOnlySetEqual_Contain_Nil(t *testing.T) {
	setErrs := []error{nil}
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := []errorReadOnlySetContainCase{
		{nil, true},
		{stderrors.New("<nil>"), false},
		{new(errorUnwrap), false},
		{new(errorIsAlwaysTrue), false},
		{new(errorIsAlwaysFalse), false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?target=%q", i, tc.target), func(t *testing.T) {
			set := errors.NewErrorReadOnlySetEqual(setErrs...)
			if set.Contain(tc.target) != tc.want {
				t.Errorf("got %t; want %t", !tc.want, tc.want)
			}
		})
	}
}

func TestErrorReadOnlySetIs_Len(t *testing.T) {
	var setErrs []error
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	set := errors.NewErrorReadOnlySetIs(setErrs...)
	if n := set.Len(); n != len(setErrs) {
		t.Errorf("got %d; want %d", n, len(setErrs))
	}
}

func TestErrorReadOnlySetIs_Len_Nil(t *testing.T) {
	setErrs := []error{nil, stderrors.New("<nil>")}
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	set := errors.NewErrorReadOnlySetIs(setErrs...)
	if n := set.Len(); n != len(setErrs) {
		t.Errorf("got %d; want %d", n, len(setErrs))
	}
}

func TestErrorReadOnlySetIs_Contain(t *testing.T) {
	var setErrs []error
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := make([]errorReadOnlySetContainCase, 0, len(setErrs)*5+4)
	for _, err := range setErrs {
		testCases = append(
			testCases,
			errorReadOnlySetContainCase{err, true},
			errorReadOnlySetContainCase{stderrors.New(err.Error()), false},
			errorReadOnlySetContainCase{&errorUnwrap{err}, true},
			errorReadOnlySetContainCase{&errorIsAlwaysTrue{err}, true},
			// When method Is returns false, errors.Is will continue testing along the Unwrap error chain rather than return false.
			errorReadOnlySetContainCase{&errorIsAlwaysFalse{err}, true},
		)
	}
	err := stderrors.New("test error +1")
	testCases = append(
		testCases,
		errorReadOnlySetContainCase{}, // {nil, false}
		errorReadOnlySetContainCase{&errorUnwrap{err}, false},
		errorReadOnlySetContainCase{&errorIsAlwaysTrue{err}, true},
		errorReadOnlySetContainCase{&errorIsAlwaysFalse{err}, false},
	)

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?target=%q", i, tc.target), func(t *testing.T) {
			set := errors.NewErrorReadOnlySetIs(setErrs...)
			if set.Contain(tc.target) != tc.want {
				t.Errorf("got %t; want %t", !tc.want, tc.want)
			}
		})
	}
}

func TestErrorReadOnlySetIs_Contain_Nil(t *testing.T) {
	setErrs := []error{nil}
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := []errorReadOnlySetContainCase{
		{nil, true},
		{stderrors.New("<nil>"), false},
		{new(errorUnwrap), false},
		{new(errorIsAlwaysTrue), true},
		{new(errorIsAlwaysFalse), false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?target=%q", i, tc.target), func(t *testing.T) {
			set := errors.NewErrorReadOnlySetIs(setErrs...)
			if set.Contain(tc.target) != tc.want {
				t.Errorf("got %t; want %t", !tc.want, tc.want)
			}
		})
	}
}

func TestErrorReadOnlySetSameMessage_Len(t *testing.T) {
	var setErrs []error
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	set := errors.NewErrorReadOnlySetSameMessage(setErrs...)
	if n := set.Len(); n != len(setErrs) {
		t.Errorf("got %d; want %d", n, len(setErrs))
	}
}

func TestErrorReadOnlySetSameMessage_Len_Nil(t *testing.T) {
	setErrs := []error{nil, stderrors.New("<nil>")}
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	set := errors.NewErrorReadOnlySetSameMessage(setErrs...)
	if n := set.Len(); n != len(setErrs) {
		t.Errorf("got %d; want %d", n, len(setErrs))
	}
}

func TestErrorReadOnlySetSameMessage_Contain(t *testing.T) {
	var setErrs []error
	msgSet := make(map[string]bool)
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
		for _, err := range errs {
			msgSet[err.Error()] = true
		}
	}
	testCases := make([]errorReadOnlySetContainCase, 0, len(setErrs)*5+1)
	for _, err := range setErrs {
		eu, et, ef := &errorUnwrap{err}, &errorIsAlwaysTrue{err}, &errorIsAlwaysFalse{err}
		testCases = append(
			testCases,
			errorReadOnlySetContainCase{err, true},
			errorReadOnlySetContainCase{stderrors.New(err.Error()), true},
			errorReadOnlySetContainCase{eu, msgSet[eu.Error()]},
			errorReadOnlySetContainCase{et, msgSet[et.Error()]},
			errorReadOnlySetContainCase{ef, msgSet[ef.Error()]},
		)
	}
	testCases = append(testCases, errorReadOnlySetContainCase{}) // {nil, false}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?target=%q", i, tc.target), func(t *testing.T) {
			set := errors.NewErrorReadOnlySetSameMessage(setErrs...)
			if set.Contain(tc.target) != tc.want {
				t.Errorf("got %t; want %t", !tc.want, tc.want)
			}
		})
	}
}

func TestErrorReadOnlySetSameMessage_Contain_Nil(t *testing.T) {
	setErrs := []error{nil}
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := []errorReadOnlySetContainCase{
		{nil, true},
		{stderrors.New("<nil>"), true},
		{new(errorUnwrap), false},
		{new(errorIsAlwaysTrue), false},
		{new(errorIsAlwaysFalse), false},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d?target=%q", i, tc.target), func(t *testing.T) {
			set := errors.NewErrorReadOnlySetSameMessage(setErrs...)
			if set.Contain(tc.target) != tc.want {
				t.Errorf("got %t; want %t", !tc.want, tc.want)
			}
		})
	}
}

type errorUnwrap struct {
	wrapped error
}

func (eu *errorUnwrap) Error() string {
	if eu.wrapped == nil {
		return "test error - <nil>"
	}
	return "test error unwrap - " + eu.wrapped.Error()
}

func (eu *errorUnwrap) Unwrap() error {
	return eu.wrapped
}

type errorIsAlwaysTrue struct {
	wrapped error
}

func (eiat *errorIsAlwaysTrue) Error() string {
	if eiat.wrapped == nil {
		return "test error Is always true - <nil>"
	}
	return "test error Is always true - " + eiat.wrapped.Error()
}

func (eiat *errorIsAlwaysTrue) Unwrap() error {
	return eiat.wrapped
}

func (eiat *errorIsAlwaysTrue) Is(error) bool {
	return true
}

type errorIsAlwaysFalse struct {
	wrapped error
}

func (eiaf *errorIsAlwaysFalse) Error() string {
	if eiaf.wrapped == nil {
		return "test error Is always false - <nil>"
	}
	return "test error Is always false - " + eiaf.wrapped.Error()
}

func (eiaf *errorIsAlwaysFalse) Unwrap() error {
	return eiaf.wrapped
}

func (eiaf *errorIsAlwaysFalse) Is(error) bool {
	return false
}
