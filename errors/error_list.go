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

package errors

import (
	"iter"
	"strconv"
	"strings"
)

// ErrorList is a list of errors, to collect multiple errors in sequence.
//
// If it is empty, it reports "no error".
// If there is only one item, it performs the same as this item.
// If there are two or more items, it reports the number of errors,
// followed by an error array, in which each item is quoted.
//
// It is designed to log errors that occur during a process and
// report them after exiting the process.
// Therefore, it only supports append operation, not delete operation.
type ErrorList interface {
	ErrorUnwrapMultiple

	// Len returns the number of errors in the error list.
	Len() int

	// Erroneous reports whether the error list contains non-nil errors.
	Erroneous() bool

	// ToList returns a copy of the list of errors, as type []error.
	//
	// If there is no errors in the list, it returns nil.
	ToList() []error

	// ToError returns a necessary error.
	//
	// If there is no error, it returns nil.
	// If there is only one item, it returns this item.
	// Otherwise, it returns the error list itself.
	ToError() error

	// Range calls handler on all items in the list one by one.
	//
	// handler has two parameters: i (the index of the error) and
	// err (the error value), and returns an indicator cont to report
	// whether to continue the iteration.
	Range(handler func(i int, err error) (cont bool))

	// RangeBackward is like Range,
	// but the order of access is from last to first.
	RangeBackward(handler func(i int, err error) (cont bool))

	// IterErrors returns an iterator over all errors in the list,
	// traversing it from first to last.
	//
	// The returned iterator is always non-nil.
	IterErrors() iter.Seq[error]

	// IterErrorsBackward returns an iterator over all errors in the list,
	// traversing it from last to first.
	//
	// The returned iterator is always non-nil.
	IterErrorsBackward() iter.Seq[error]

	// IterIndexErrors returns an iterator over index-error pairs in the list,
	// traversing it from first to last.
	//
	// The returned iterator is always non-nil.
	IterIndexErrors() iter.Seq2[int, error]

	// IterIndexErrorsBackward returns an iterator
	// over index-error pairs in the list,
	// traversing it from last to first with descending indices.
	//
	// The returned iterator is always non-nil.
	IterIndexErrorsBackward() iter.Seq2[int, error]

	// Append appends err to the error list.
	Append(err ...error)

	// Deduplicate removes duplicate and nil errors.
	//
	// An error is regarded as duplicate if its method Error returns
	// the same string as that of a previous error.
	Deduplicate()
}

// errorList is an implementation of interface ErrorList.
type errorList struct {
	list      []error
	ignoreNil bool
}

// NewErrorList creates a new ErrorList.
//
// ignoreNil indicates whether the ErrorList ignores nil errors.
// If ignoreNil is true, the ErrorList discards all nil errors.
//
// err is errors added to the ErrorList initially.
func NewErrorList(ignoreNil bool, err ...error) ErrorList {
	el := &errorList{ignoreNil: ignoreNil}
	if len(err) > 0 {
		el.list = make([]error, 0, len(err))
		el.Append(err...)
	}
	return el
}

// Error returns the error message of this error list, as follows:
//
// If the error list is empty, it returns "no error".
//
// If there is only one item (denoted by t), then
// it returns t.Error() if t != nil, or returns "<nil>" otherwise.
//
// If there are two or more items, it returns the number of errors,
// followed by an error array, in which each item is double-quoted
// in Go string literal.
// In particular, nil error item is "<nil>".
func (el *errorList) Error() string {
	if len(el.list) == 0 {
		return "no error"
	} else if len(el.list) == 1 {
		err := el.list[0]
		if err != nil {
			return err.Error()
		}
		return "<nil>"
	}
	var builder strings.Builder
	s := strconv.Itoa(len(el.list))
	builder.WriteString(s)
	builder.WriteString(" errors: [")
	for i, err := range el.list {
		if i > 0 {
			builder.WriteString(", ")
		}
		s = "<nil>"
		if err != nil {
			s = err.Error()
		}
		s = strconv.Quote(s)
		builder.WriteString(s)
	}
	builder.WriteString("]")
	return builder.String()
}

// Unwrap is like ToList, but drops nil-value errors.
func (el *errorList) Unwrap() []error {
	if el.ignoreNil {
		return el.ToList()
	}
	var n int
	for _, err := range el.list {
		if err != nil {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	errs := make([]error, 0, n)
	for _, err := range el.list {
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (el *errorList) Len() int {
	return len(el.list)
}

func (el *errorList) Erroneous() bool {
	if el.ignoreNil {
		return len(el.list) > 0
	}
	for _, err := range el.list {
		if err != nil {
			return true
		}
	}
	return false
}

func (el *errorList) ToList() []error {
	if len(el.list) == 0 {
		return nil
	}
	errs := make([]error, len(el.list))
	copy(errs, el.list)
	return errs
}

func (el *errorList) ToError() error {
	switch len(el.list) {
	case 0:
		return nil
	case 1:
		return el.list[0]
	default:
		return el
	}
}

func (el *errorList) Range(handler func(i int, err error) (cont bool)) {
	if handler != nil {
		for i, err := range el.list {
			if !handler(i, err) {
				return
			}
		}
	}
}

func (el *errorList) RangeBackward(handler func(i int, err error) (cont bool)) {
	if handler != nil {
		for i := len(el.list) - 1; i >= 0; i-- {
			if !handler(i, el.list[i]) {
				return
			}
		}
	}
}

func (el *errorList) IterErrors() iter.Seq[error] {
	return func(yield func(error) bool) {
		if yield != nil {
			for _, err := range el.list {
				if !yield(err) {
					return
				}
			}
		}
	}
}

func (el *errorList) IterErrorsBackward() iter.Seq[error] {
	return func(yield func(error) bool) {
		if yield != nil {
			for i := len(el.list) - 1; i >= 0; i-- {
				if !yield(el.list[i]) {
					return
				}
			}
		}
	}
}

func (el *errorList) IterIndexErrors() iter.Seq2[int, error] {
	return el.Range
}

func (el *errorList) IterIndexErrorsBackward() iter.Seq2[int, error] {
	return el.RangeBackward
}

func (el *errorList) Append(err ...error) {
	for _, e := range err {
		if e != nil || !el.ignoreNil {
			el.list = append(el.list, e)
		}
	}
}

func (el *errorList) Deduplicate() {
	if len(el.list) == 0 {
		return
	}
	set, n := make(map[string]struct{}), 0
	for i := range el.list {
		x := el.list[i]
		if x != nil {
			errStr := x.Error()
			if _, ok := set[errStr]; !ok {
				set[errStr] = struct{}{}
				el.list[n] = x
				n++
			}
		}
	}
	if n == len(el.list) {
		return
	}
	clear(el.list[n:]) // avoid memory leak
	el.list = el.list[:n]
}

// Combine collects multiple non-nil errors into an error list.
//
// It discards all nil errors.
//
// If there is no non-nil error, it returns nil.
// If there is only one non-nil error, it returns this error.
// Otherwise, it returns an ErrorList containing all non-nil errors.
func Combine(err ...error) error {
	if len(err) == 0 {
		return nil
	}
	return NewErrorList(true, err...).ToError()
}
