// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2026  Yuan Gao
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
	"maps"
	"slices"
	"testing"

	"github.com/donyori/gogo/errors"
)

var (
	errorsForErrorReadOnlySet     [][]error
	errAnotherForErrorReadOnlySet error
)

func init() {
	errAnotherForErrorReadOnlySet = stderrors.New("another error")

	const N int = 3

	errorsForErrorReadOnlySet = make([][]error, 5)

	errorsForErrorReadOnlySet[0] = make([]error, N)
	for i := range N {
		errorsForErrorReadOnlySet[0][i] = fmt.Errorf("test error %d", i)
	}

	errorsForErrorReadOnlySet[1] = make([]error, N<<1)
	for i := range N {
		errorsForErrorReadOnlySet[1][i<<1] = &unwrapError{
			errorsForErrorReadOnlySet[0][i],
		}
		errorsForErrorReadOnlySet[1][i<<1+1] = &unwrapError{
			errorsForErrorReadOnlySet[1][i<<1],
		}
	}

	errorsForErrorReadOnlySet[2] = make([]error, N<<1)
	errorsForErrorReadOnlySet[3] = make([]error, N<<1)

	for i := range N {
		errorsForErrorReadOnlySet[2][i<<1] = &isAlwaysTrueError{
			errorsForErrorReadOnlySet[0][i],
		}
		errorsForErrorReadOnlySet[2][i<<1+1] = &isAlwaysTrueError{
			errorsForErrorReadOnlySet[1][i<<1],
		}
		errorsForErrorReadOnlySet[3][i<<1] = &isAlwaysFalseError{
			errorsForErrorReadOnlySet[0][i],
		}
		errorsForErrorReadOnlySet[3][i<<1+1] = &isAlwaysFalseError{
			errorsForErrorReadOnlySet[1][i<<1],
		}
	}

	errorsForErrorReadOnlySet[4] = make([]error, N)
	for i := range N {
		errorsForErrorReadOnlySet[4][i] = stderrors.Join(
			errorsForErrorReadOnlySet[1][i<<1],
			errorsForErrorReadOnlySet[3][i<<1],
			errAnotherForErrorReadOnlySet,
		)
	}
}

type errorReadOnlySetContainsCase struct {
	target error
	want   bool
}

func TestErrorReadOnlySetEqual_Len(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	var setErrs []error
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}

	testCases := make([]errorReadOnlySetContainsCase, 0, len(setErrs)*7+1)

	for _, err := range setErrs {
		testCases = append(
			testCases,
			errorReadOnlySetContainsCase{
				err,
				true,
			},
			errorReadOnlySetContainsCase{
				stderrors.New(err.Error()),
				false,
			},
			errorReadOnlySetContainsCase{
				&unwrapError{err},
				false,
			},
			errorReadOnlySetContainsCase{
				&isAlwaysTrueError{err},
				false,
			},
			errorReadOnlySetContainsCase{
				&isAlwaysFalseError{err},
				false,
			},
			errorReadOnlySetContainsCase{
				errAnotherForErrorReadOnlySet,
				false,
			},
			errorReadOnlySetContainsCase{
				stderrors.Join(errAnotherForErrorReadOnlySet, err),
				false,
			},
		)
	}

	testCases = append(testCases, errorReadOnlySetContainsCase{
		nil,
		false,
	})

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case%d?target=%+q", i, tc.target),
			func(t *testing.T) {
				t.Parallel()

				set := errors.NewErrorReadOnlySetEqual(setErrs...)
				if set.Contains(tc.target) != tc.want {
					t.Errorf("got %t; want %t", !tc.want, tc.want)
				}
			},
		)
	}
}

func TestErrorReadOnlySetEqual_Contains_Nil(t *testing.T) {
	t.Parallel()

	setErrs := []error{nil}
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}

	testCases := []errorReadOnlySetContainsCase{
		{nil, true},
		{stderrors.New("<nil>"), false},
		{new(unwrapError), false},
		{new(isAlwaysTrueError), false},
		{new(isAlwaysFalseError), false},
	}

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case%d?target=%+q", i, tc.target),
			func(t *testing.T) {
				t.Parallel()

				set := errors.NewErrorReadOnlySetEqual(setErrs...)
				if set.Contains(tc.target) != tc.want {
					t.Errorf("got %t; want %t", !tc.want, tc.want)
				}
			},
		)
	}
}

func TestErrorReadOnlySetEqual_Range(t *testing.T) {
	t.Parallel()

	errs := slices.Clone(errorsForErrorReadOnlySet[0])

	counterMap := make(map[error]int, len(errs))
	for _, err := range errs {
		counterMap[err] = 1
	}

	set := errors.NewErrorReadOnlySetEqual(errs...)
	set.Range(func(err error) (cont bool) {
		counterMap[err]--
		return true
	})

	for err, ctr := range counterMap {
		if ctr > 0 {
			t.Error("insufficient accesses to", err)
		} else if ctr < 0 {
			t.Error("too many accesses to", err)
		}
	}
}

func TestErrorReadOnlySetEqual_Range_Empty(t *testing.T) {
	t.Parallel()

	set := errors.NewErrorReadOnlySetEqual()
	set.Range(func(err error) (cont bool) {
		t.Error("handler was called, err:", err)
		return true
	})
}

func TestErrorReadOnlySetEqual_Range_NilHandler(t *testing.T) {
	t.Parallel()

	errs := slices.Clone(errorsForErrorReadOnlySet[0])
	set := errors.NewErrorReadOnlySetEqual(errs...)

	defer func() {
		if e := recover(); e != nil {
			t.Error("panic:", e)
		}
	}()

	set.Range(nil)
}

func TestErrorReadOnlySetEqual_IterErrors(t *testing.T) {
	t.Parallel()

	errs := slices.Clone(errorsForErrorReadOnlySet[0])

	counterMap := make(map[error]int, len(errs))
	for _, err := range errs {
		counterMap[err] = 1
	}

	set := errors.NewErrorReadOnlySetEqual(errs...)

	seq := set.IterErrors()
	if seq == nil {
		t.Fatal("got nil iterator")
	}

	counterMapCopy := maps.Clone(counterMap)
	for err := range seq {
		counterMap[err]--
	}

	for err, ctr := range counterMap {
		if ctr > 0 {
			t.Error("insufficient accesses to", err)
		} else if ctr < 0 {
			t.Error("too many accesses to", err)
		}
	}

	// Rewind the iterator and test it again.
	for err := range seq {
		counterMapCopy[err]--
	}

	for err, ctr := range counterMapCopy {
		if ctr > 0 {
			t.Error("rewind, insufficient accesses to", err)
		} else if ctr < 0 {
			t.Error("rewind, too many accesses to", err)
		}
	}
}

func TestErrorReadOnlySetEqual_IterErrors_Empty(t *testing.T) {
	t.Parallel()

	set := errors.NewErrorReadOnlySetEqual()

	seq := set.IterErrors()
	if seq == nil {
		t.Fatal("got nil iterator")
	}

	for err := range seq {
		t.Error("yielded", err)
	}
}

func TestErrorReadOnlySetIs_Len(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	var setErrs []error
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}

	testCases := make([]errorReadOnlySetContainsCase, 0, len(setErrs)*7+5)

	for _, err := range setErrs {
		testCases = append(
			testCases,
			errorReadOnlySetContainsCase{
				err,
				true,
			},
			errorReadOnlySetContainsCase{
				stderrors.New(err.Error()),
				false,
			},
			errorReadOnlySetContainsCase{
				&unwrapError{err},
				true,
			},
			errorReadOnlySetContainsCase{
				&isAlwaysTrueError{err},
				true,
			},

			// When method Is returns false,
			// errors.Is will continue testing along the Unwrap error tree
			// rather than return false.
			errorReadOnlySetContainsCase{
				&isAlwaysFalseError{err},
				true,
			},

			errorReadOnlySetContainsCase{
				errAnotherForErrorReadOnlySet,
				false,
			},
			errorReadOnlySetContainsCase{
				stderrors.Join(errAnotherForErrorReadOnlySet, err),
				true,
			},
		)
	}

	err := stderrors.New("test error +1")
	testCases = append(
		testCases,
		errorReadOnlySetContainsCase{
			nil,
			false,
		},
		errorReadOnlySetContainsCase{
			&unwrapError{err},
			false,
		},
		errorReadOnlySetContainsCase{
			&isAlwaysTrueError{err},
			true,
		},
		errorReadOnlySetContainsCase{
			&isAlwaysFalseError{err},
			false,
		},
		errorReadOnlySetContainsCase{
			stderrors.Join(errAnotherForErrorReadOnlySet, err),
			false,
		},
	)

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case%d?target=%+q", i, tc.target),
			func(t *testing.T) {
				t.Parallel()

				set := errors.NewErrorReadOnlySetIs(setErrs...)
				if set.Contains(tc.target) != tc.want {
					t.Errorf("got %t; want %t", !tc.want, tc.want)
				}
			},
		)
	}
}

func TestErrorReadOnlySetIs_Contains_Nil(t *testing.T) {
	t.Parallel()

	setErrs := []error{nil}
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}

	testCases := []errorReadOnlySetContainsCase{
		{nil, true},
		{stderrors.New("<nil>"), false},
		{new(unwrapError), false},
		{new(isAlwaysTrueError), true},
		{new(isAlwaysFalseError), false},
	}

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case%d?target=%+q", i, tc.target),
			func(t *testing.T) {
				t.Parallel()

				set := errors.NewErrorReadOnlySetIs(setErrs...)
				if set.Contains(tc.target) != tc.want {
					t.Errorf("got %t; want %t", !tc.want, tc.want)
				}
			},
		)
	}
}

func TestErrorReadOnlySetIs_Range(t *testing.T) {
	t.Parallel()

	errs := slices.Clone(errorsForErrorReadOnlySet[0])

	counterMap := make(map[error]int, len(errs))
	for _, err := range errs {
		counterMap[err] = 1
	}

	set := errors.NewErrorReadOnlySetIs(errs...)
	set.Range(func(err error) (cont bool) {
		counterMap[err]--
		return true
	})

	for err, ctr := range counterMap {
		if ctr > 0 {
			t.Error("insufficient accesses to", err)
		} else if ctr < 0 {
			t.Error("too many accesses to", err)
		}
	}
}

func TestErrorReadOnlySetIs_Range_Empty(t *testing.T) {
	t.Parallel()

	set := errors.NewErrorReadOnlySetIs()
	set.Range(func(err error) (cont bool) {
		t.Error("handler was called, err:", err)
		return true
	})
}

func TestErrorReadOnlySetIs_Range_NilHandler(t *testing.T) {
	t.Parallel()

	errs := slices.Clone(errorsForErrorReadOnlySet[0])
	set := errors.NewErrorReadOnlySetIs(errs...)

	defer func() {
		if e := recover(); e != nil {
			t.Error("panic:", e)
		}
	}()

	set.Range(nil)
}

func TestErrorReadOnlySetIs_IterErrors(t *testing.T) {
	t.Parallel()

	errs := slices.Clone(errorsForErrorReadOnlySet[0])

	counterMap := make(map[error]int, len(errs))
	for _, err := range errs {
		counterMap[err] = 1
	}

	set := errors.NewErrorReadOnlySetIs(errs...)

	seq := set.IterErrors()
	if seq == nil {
		t.Fatal("got nil iterator")
	}

	counterMapCopy := maps.Clone(counterMap)
	for err := range seq {
		counterMap[err]--
	}

	for err, ctr := range counterMap {
		if ctr > 0 {
			t.Error("insufficient accesses to", err)
		} else if ctr < 0 {
			t.Error("too many accesses to", err)
		}
	}

	// Rewind the iterator and test it again.
	for err := range seq {
		counterMapCopy[err]--
	}

	for err, ctr := range counterMapCopy {
		if ctr > 0 {
			t.Error("rewind, insufficient accesses to", err)
		} else if ctr < 0 {
			t.Error("rewind, too many accesses to", err)
		}
	}
}

func TestErrorReadOnlySetIs_IterErrors_Empty(t *testing.T) {
	t.Parallel()

	set := errors.NewErrorReadOnlySetIs()

	seq := set.IterErrors()
	if seq == nil {
		t.Fatal("got nil iterator")
	}

	for err := range seq {
		t.Error("yielded", err)
	}
}

func TestErrorReadOnlySetSameMessage_Len(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	msgSet := make(map[string]struct{})

	var setErrs []error
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
		for _, err := range errs {
			msgSet[err.Error()] = struct{}{}
		}
	}

	testCases := make([]errorReadOnlySetContainsCase, 0, len(setErrs)*7+1)

	for _, err := range setErrs {
		ue := &unwrapError{err}
		_, wantEU := msgSet[ue.Error()]
		te := &isAlwaysTrueError{err}
		_, wantET := msgSet[te.Error()]
		fe := &isAlwaysFalseError{err}
		_, wantEF := msgSet[fe.Error()]
		testCases = append(
			testCases,
			errorReadOnlySetContainsCase{
				err,
				true,
			},
			errorReadOnlySetContainsCase{
				stderrors.New(err.Error()),
				true,
			},
			errorReadOnlySetContainsCase{
				ue,
				wantEU,
			},
			errorReadOnlySetContainsCase{
				te,
				wantET,
			},
			errorReadOnlySetContainsCase{
				fe,
				wantEF,
			},
			errorReadOnlySetContainsCase{
				errAnotherForErrorReadOnlySet,
				false,
			},
			errorReadOnlySetContainsCase{
				stderrors.Join(errAnotherForErrorReadOnlySet, err),
				false,
			},
		)
	}

	testCases = append(testCases, errorReadOnlySetContainsCase{
		nil,
		false,
	})

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case%d?target=%+q", i, tc.target),
			func(t *testing.T) {
				t.Parallel()

				set := errors.NewErrorReadOnlySetSameMessage(setErrs...)
				if set.Contains(tc.target) != tc.want {
					t.Errorf("got %t; want %t", !tc.want, tc.want)
				}
			},
		)
	}
}

func TestErrorReadOnlySetSameMessage_Contains_Nil(t *testing.T) {
	t.Parallel()

	setErrs := []error{nil}
	for _, errs := range errorsForErrorReadOnlySet {
		setErrs = append(setErrs, errs...)
	}

	testCases := []errorReadOnlySetContainsCase{
		{nil, true},
		{stderrors.New("<nil>"), true},
		{new(unwrapError), false},
		{new(isAlwaysTrueError), false},
		{new(isAlwaysFalseError), false},
	}

	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("case%d?target=%+q", i, tc.target),
			func(t *testing.T) {
				t.Parallel()

				set := errors.NewErrorReadOnlySetSameMessage(setErrs...)
				if set.Contains(tc.target) != tc.want {
					t.Errorf("got %t; want %t", !tc.want, tc.want)
				}
			},
		)
	}
}

func TestErrorReadOnlySetSameMessage_Range(t *testing.T) {
	t.Parallel()

	errs := slices.Clone(errorsForErrorReadOnlySet[0])

	counterMap := make(map[error]int, len(errs))
	for _, err := range errs {
		counterMap[err] = 1
	}

	set := errors.NewErrorReadOnlySetSameMessage(errs...)
	set.Range(func(err error) (cont bool) {
		counterMap[err]--
		return true
	})

	for err, ctr := range counterMap {
		if ctr > 0 {
			t.Error("insufficient accesses to", err)
		} else if ctr < 0 {
			t.Error("too many accesses to", err)
		}
	}
}

func TestErrorReadOnlySetSameMessage_Range_Empty(t *testing.T) {
	t.Parallel()

	set := errors.NewErrorReadOnlySetSameMessage()
	set.Range(func(err error) (cont bool) {
		t.Error("handler was called, err:", err)
		return true
	})
}

func TestErrorReadOnlySetSameMessage_Range_NilHandler(t *testing.T) {
	t.Parallel()

	errs := slices.Clone(errorsForErrorReadOnlySet[0])
	set := errors.NewErrorReadOnlySetSameMessage(errs...)

	defer func() {
		if e := recover(); e != nil {
			t.Error("panic:", e)
		}
	}()

	set.Range(nil)
}

func TestErrorReadOnlySetSameMessage_IterErrors(t *testing.T) {
	t.Parallel()

	errs := slices.Clone(errorsForErrorReadOnlySet[0])

	counterMap := make(map[error]int, len(errs))
	for _, err := range errs {
		counterMap[err] = 1
	}

	set := errors.NewErrorReadOnlySetSameMessage(errs...)

	seq := set.IterErrors()
	if seq == nil {
		t.Fatal("got nil iterator")
	}

	counterMapCopy := maps.Clone(counterMap)
	for err := range seq {
		counterMap[err]--
	}

	for err, ctr := range counterMap {
		if ctr > 0 {
			t.Error("insufficient accesses to", err)
		} else if ctr < 0 {
			t.Error("too many accesses to", err)
		}
	}

	// Rewind the iterator and test it again.
	for err := range seq {
		counterMapCopy[err]--
	}

	for err, ctr := range counterMapCopy {
		if ctr > 0 {
			t.Error("rewind, insufficient accesses to", err)
		} else if ctr < 0 {
			t.Error("rewind, too many accesses to", err)
		}
	}
}

func TestErrorReadOnlySetSameMessage_IterErrors_Empty(t *testing.T) {
	t.Parallel()

	set := errors.NewErrorReadOnlySetSameMessage()

	seq := set.IterErrors()
	if seq == nil {
		t.Fatal("got nil iterator")
	}

	for err := range seq {
		t.Error("yielded", err)
	}
}

type unwrapError struct {
	err error
}

func (ue *unwrapError) Error() string {
	if ue.err == nil {
		return "test error: <nil>"
	}

	return "test error unwrap: " + ue.err.Error()
}

func (ue *unwrapError) Unwrap() error {
	return ue.err
}

type isAlwaysTrueError struct {
	err error
}

func (te *isAlwaysTrueError) Error() string {
	if te.err == nil {
		return "test error Is always true: <nil>"
	}

	return "test error Is always true: " + te.err.Error()
}

func (te *isAlwaysTrueError) Unwrap() error {
	return te.err
}

func (te *isAlwaysTrueError) Is(error) bool {
	return true
}

type isAlwaysFalseError struct {
	err error
}

func (fe *isAlwaysFalseError) Error() string {
	if fe.err == nil {
		return "test error Is always false: <nil>"
	}

	return "test error Is always false: " + fe.err.Error()
}

func (fe *isAlwaysFalseError) Unwrap() error {
	return fe.err
}

func (fe *isAlwaysFalseError) Is(error) bool {
	return false
}
