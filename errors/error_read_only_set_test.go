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
	"testing"

	"github.com/donyori/gogo/errors"
)

var (
	errorsForErrorReadOnlySet       [][]error
	anotherErrorForErrorReadOnlySet error
)

func init() {
	anotherErrorForErrorReadOnlySet = stderrors.New("another error")

	const N int = 3
	errorsForErrorReadOnlySet = make([][]error, 5)
	errorsForErrorReadOnlySet[0] = make([]error, N)
	for i := range N {
		errorsForErrorReadOnlySet[0][i] = fmt.Errorf("test error %d", i)
	}
	errorsForErrorReadOnlySet[1] = make([]error, N<<1)
	for i := range N {
		errorsForErrorReadOnlySet[1][i<<1] = &errorUnwrap{errorsForErrorReadOnlySet[0][i]}
		errorsForErrorReadOnlySet[1][i<<1+1] = &errorUnwrap{errorsForErrorReadOnlySet[1][i<<1]}
	}
	errorsForErrorReadOnlySet[2] = make([]error, N<<1)
	errorsForErrorReadOnlySet[3] = make([]error, N<<1)
	for i := range N {
		errorsForErrorReadOnlySet[2][i<<1] = &errorIsAlwaysTrue{errorsForErrorReadOnlySet[0][i]}
		errorsForErrorReadOnlySet[2][i<<1+1] = &errorIsAlwaysTrue{errorsForErrorReadOnlySet[1][i<<1]}
		errorsForErrorReadOnlySet[3][i<<1] = &errorIsAlwaysFalse{errorsForErrorReadOnlySet[0][i]}
		errorsForErrorReadOnlySet[3][i<<1+1] = &errorIsAlwaysFalse{errorsForErrorReadOnlySet[1][i<<1]}
	}
	errorsForErrorReadOnlySet[4] = make([]error, N)
	for i := range N {
		errorsForErrorReadOnlySet[4][i] = stderrors.Join(
			errorsForErrorReadOnlySet[1][i<<1],
			errorsForErrorReadOnlySet[3][i<<1],
			anotherErrorForErrorReadOnlySet,
		)
	}
}

type errorReadOnlySetContainsCase struct {
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

func TestErrorReadOnlySetEqual_Contains(t *testing.T) {
	var setErrs []error
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := make([]errorReadOnlySetContainsCase, 0, len(setErrs)*7+1)
	for _, err := range setErrs {
		testCases = append(
			testCases,
			errorReadOnlySetContainsCase{err, true},
			errorReadOnlySetContainsCase{stderrors.New(err.Error()), false},
			errorReadOnlySetContainsCase{&errorUnwrap{err}, false},
			errorReadOnlySetContainsCase{&errorIsAlwaysTrue{err}, false},
			errorReadOnlySetContainsCase{&errorIsAlwaysFalse{err}, false},
			errorReadOnlySetContainsCase{anotherErrorForErrorReadOnlySet, false},
			errorReadOnlySetContainsCase{stderrors.Join(anotherErrorForErrorReadOnlySet, err), false},
		)
	}
	testCases = append(testCases, errorReadOnlySetContainsCase{}) // {nil, false}

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case %d?target=%+q", i, tc.target),
			func(t *testing.T) {
				set := errors.NewErrorReadOnlySetEqual(setErrs...)
				if set.Contains(tc.target) != tc.want {
					t.Errorf("got %t; want %t", !tc.want, tc.want)
				}
			},
		)
	}
}

func TestErrorReadOnlySetEqual_Contains_Nil(t *testing.T) {
	setErrs := []error{nil}
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := []errorReadOnlySetContainsCase{
		{nil, true},
		{stderrors.New("<nil>"), false},
		{new(errorUnwrap), false},
		{new(errorIsAlwaysTrue), false},
		{new(errorIsAlwaysFalse), false},
	}

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case %d?target=%+q", i, tc.target),
			func(t *testing.T) {
				set := errors.NewErrorReadOnlySetEqual(setErrs...)
				if set.Contains(tc.target) != tc.want {
					t.Errorf("got %t; want %t", !tc.want, tc.want)
				}
			},
		)
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

func TestErrorReadOnlySetIs_Contains(t *testing.T) {
	var setErrs []error
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := make([]errorReadOnlySetContainsCase, 0, len(setErrs)*7+5)
	for _, err := range setErrs {
		testCases = append(
			testCases,
			errorReadOnlySetContainsCase{err, true},
			errorReadOnlySetContainsCase{stderrors.New(err.Error()), false},
			errorReadOnlySetContainsCase{&errorUnwrap{err}, true},
			errorReadOnlySetContainsCase{&errorIsAlwaysTrue{err}, true},
			// When method Is returns false,
			// errors.Is will continue testing along the Unwrap error tree
			// rather than return false.
			errorReadOnlySetContainsCase{&errorIsAlwaysFalse{err}, true},
			errorReadOnlySetContainsCase{anotherErrorForErrorReadOnlySet, false},
			errorReadOnlySetContainsCase{stderrors.Join(anotherErrorForErrorReadOnlySet, err), true},
		)
	}
	err := stderrors.New("test error +1")
	testCases = append(
		testCases,
		errorReadOnlySetContainsCase{}, // {nil, false}
		errorReadOnlySetContainsCase{&errorUnwrap{err}, false},
		errorReadOnlySetContainsCase{&errorIsAlwaysTrue{err}, true},
		errorReadOnlySetContainsCase{&errorIsAlwaysFalse{err}, false},
		errorReadOnlySetContainsCase{stderrors.Join(anotherErrorForErrorReadOnlySet, err), false},
	)

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case %d?target=%+q", i, tc.target),
			func(t *testing.T) {
				set := errors.NewErrorReadOnlySetIs(setErrs...)
				if set.Contains(tc.target) != tc.want {
					t.Errorf("got %t; want %t", !tc.want, tc.want)
				}
			},
		)
	}
}

func TestErrorReadOnlySetIs_Contains_Nil(t *testing.T) {
	setErrs := []error{nil}
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := []errorReadOnlySetContainsCase{
		{nil, true},
		{stderrors.New("<nil>"), false},
		{new(errorUnwrap), false},
		{new(errorIsAlwaysTrue), true},
		{new(errorIsAlwaysFalse), false},
	}

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case %d?target=%+q", i, tc.target),
			func(t *testing.T) {
				set := errors.NewErrorReadOnlySetIs(setErrs...)
				if set.Contains(tc.target) != tc.want {
					t.Errorf("got %t; want %t", !tc.want, tc.want)
				}
			},
		)
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

func TestErrorReadOnlySetSameMessage_Contains(t *testing.T) {
	var setErrs []error
	msgSet := make(map[string]struct{})
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
		for _, err := range errs {
			msgSet[err.Error()] = struct{}{}
		}
	}
	testCases := make([]errorReadOnlySetContainsCase, 0, len(setErrs)*7+1)
	for _, err := range setErrs {
		eu := &errorUnwrap{err}
		_, wantEU := msgSet[eu.Error()]
		et := &errorIsAlwaysTrue{err}
		_, wantET := msgSet[et.Error()]
		ef := &errorIsAlwaysFalse{err}
		_, wantEF := msgSet[ef.Error()]
		testCases = append(
			testCases,
			errorReadOnlySetContainsCase{err, true},
			errorReadOnlySetContainsCase{stderrors.New(err.Error()), true},
			errorReadOnlySetContainsCase{eu, wantEU},
			errorReadOnlySetContainsCase{et, wantET},
			errorReadOnlySetContainsCase{ef, wantEF},
			errorReadOnlySetContainsCase{anotherErrorForErrorReadOnlySet, false},
			errorReadOnlySetContainsCase{stderrors.Join(anotherErrorForErrorReadOnlySet, err), false},
		)
	}
	testCases = append(testCases, errorReadOnlySetContainsCase{}) // {nil, false}

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case %d?target=%+q", i, tc.target),
			func(t *testing.T) {
				set := errors.NewErrorReadOnlySetSameMessage(setErrs...)
				if set.Contains(tc.target) != tc.want {
					t.Errorf("got %t; want %t", !tc.want, tc.want)
				}
			},
		)
	}
}

func TestErrorReadOnlySetSameMessage_Contains_Nil(t *testing.T) {
	setErrs := []error{nil}
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}
	testCases := []errorReadOnlySetContainsCase{
		{nil, true},
		{stderrors.New("<nil>"), true},
		{new(errorUnwrap), false},
		{new(errorIsAlwaysTrue), false},
		{new(errorIsAlwaysFalse), false},
	}

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case %d?target=%+q", i, tc.target),
			func(t *testing.T) {
				set := errors.NewErrorReadOnlySetSameMessage(setErrs...)
				if set.Contains(tc.target) != tc.want {
					t.Errorf("got %t; want %t", !tc.want, tc.want)
				}
			},
		)
	}
}

type errorUnwrap struct {
	err error
}

func (eu *errorUnwrap) Error() string {
	if eu.err == nil {
		return "test error - <nil>"
	}
	return "test error unwrap - " + eu.err.Error()
}

func (eu *errorUnwrap) Unwrap() error {
	return eu.err
}

type errorIsAlwaysTrue struct {
	err error
}

func (eiat *errorIsAlwaysTrue) Error() string {
	if eiat.err == nil {
		return "test error Is always true - <nil>"
	}
	return "test error Is always true - " + eiat.err.Error()
}

func (eiat *errorIsAlwaysTrue) Unwrap() error {
	return eiat.err
}

func (eiat *errorIsAlwaysTrue) Is(error) bool {
	return true
}

type errorIsAlwaysFalse struct {
	err error
}

func (eiaf *errorIsAlwaysFalse) Error() string {
	if eiaf.err == nil {
		return "test error Is always false - <nil>"
	}
	return "test error Is always false - " + eiaf.err.Error()
}

func (eiaf *errorIsAlwaysFalse) Unwrap() error {
	return eiaf.err
}

func (eiaf *errorIsAlwaysFalse) Is(error) bool {
	return false
}
