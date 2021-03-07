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
	"fmt"
	"strconv"
	"strings"
)

// ErrorList is a list of errors, to collect multiple errors in sequence.
//
// If it is empty, it reports "no error".
// If there is only one item, it performs the same as this item.
// If there are two or more items, it reports the number of errors,
// followed by an error array, in which every item is quoted.
//
// Note that *ErrorList, not ErrorList,
// is the implementation of interface error.
type ErrorList interface {
	error
	fmt.Stringer

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

	// Range calls handler for all items in the list one by one.
	//
	// handler has two parameters: i (the index of the error) and
	// err (the error value),
	// and returns an indicator cont to report whether to continue the iteration.
	Range(handler func(i int, err error) (cont bool))

	// Append appends errs to the error list.
	Append(errs ...error)

	// Deduplicate removes duplicated and nil errors.
	//
	// An error is regarded as duplicated if its method Error returns
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
// If ignoreNil is true, the ErrorList will discard all nil errors.
// errs is a list of errors adding to the ErrorList initially.
func NewErrorList(ignoreNil bool, errs ...error) ErrorList {
	el := &errorList{ignoreNil: ignoreNil}
	if len(errs) > 0 {
		el.list = make([]error, 0, len(errs))
		el.Append(errs...)
	}
	return el
}

// Error returns the same as el.String().
func (el *errorList) Error() string {
	return el.String()
}

// String returns the error message of this error list, as follows:
//
// If the error list is empty, it returns "no error".
//
// If there is only one item (denoted by t), then
// it returns t.Error() if t != nil, or returns "<nil>" otherwise.
//
// If there are two or more items, it returns the number of errors,
// followed by an error array, in which every item is double-quoted
// in Go string literal.
// Especially, nil error item will be "<nil>".
func (el *errorList) String() string {
	if len(el.list) == 0 {
		return "no error"
	}
	if len(el.list) == 1 {
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

// Len returns the number of errors in the error list.
func (el *errorList) Len() int {
	return len(el.list)
}

// Erroneous reports whether the error list contains non-nil errors.
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

// ToList returns a copy of the list of errors, as type []error.
//
// If there is no errors in the list, it returns nil.
func (el *errorList) ToList() []error {
	if len(el.list) == 0 {
		return nil
	}
	errs := make([]error, len(el.list)) // Explicitly set the capacity of the slice to len(el.list).
	copy(errs, el.list)
	return errs
}

// ToError returns a necessary error.
//
// If there is no error, it returns nil.
// If there is only one item, it returns this item.
// Otherwise, it returns the error list itself.
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

// Range calls handler for all items in the list one by one.
//
// handler has two parameters: i (the index of the error) and
// err (the error value),
// and returns an indicator cont to report whether to continue the iteration.
func (el *errorList) Range(handler func(i int, err error) (cont bool)) {
	if len(el.list) == 0 {
		return
	}
	for i, err := range el.list {
		if !handler(i, err) {
			return
		}
	}
}

// Append appends errs to the error list.
func (el *errorList) Append(errs ...error) {
	for _, err := range errs {
		if err != nil || !el.ignoreNil {
			el.list = append(el.list, err)
		}
	}
}

// Deduplicate removes duplicated and nil errors.
//
// An error is regarded as duplicated if its method Error returns
// the same string as that of a previous error.
func (el *errorList) Deduplicate() {
	if len(el.list) == 0 {
		return
	}
	set := make(map[string]bool)
	n := 0
	for i := 0; i < len(el.list); i++ {
		x := el.list[i]
		if x != nil {
			errStr := x.Error()
			if !set[errStr] {
				set[errStr] = true
				el.list[n] = x
				n++
			}
		}
	}
	if n == len(el.list) {
		return
	}
	for i := n; i < len(el.list); i++ {
		el.list[i] = nil
	}
	el.list = el.list[:n]
}

// Combine collects multiple non-nil errors into an error list.
//
// It discards all nil errors.
//
// If there is no non-nil error, it returns nil.
// If there is only one non-nil error, it returns this error.
// Otherwise, it returns an ErrorList containing all non-nil errors.
func Combine(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}
	return NewErrorList(true, errs...).ToError()
}
