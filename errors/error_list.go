// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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
	"strconv"
	"strings"

	"github.com/donyori/gogo/container/sequence"
)

// An error list, to collect multiple errors.
//
// If it is empty, it reports "no errors".
// If there is only one item, it performs the same as this item.
// If there are two or more items, it reports the number of errors,
// followed by an error array, in which every item is quoted.
//
// Note that *ErrorList is an implementation of interface error,
// but not ErrorList, in order to make it comparable.
//
// Recommend using function Combine and method Append to make a new ErrorList
// and add error(s) to ErrorList to discard nil errors.
type ErrorList []error

// Combine multiple errors into an ErrorList.
// Note that nil error will be discarded.
// It returns nil if len(errs) == 0.
func Combine(errs ...error) ErrorList {
	if len(errs) == 0 {
		return nil
	}
	el := make(ErrorList, 0, len(errs))
	el.Append(errs...)
	return el
}

// Return the number of errors in the error list.
func (el ErrorList) Len() int {
	return len(el)
}

// Append errors to the error list. Note that nil error will be discarded.
func (el *ErrorList) Append(errs ...error) {
	for _, err := range errs {
		if err != nil {
			*el = append(*el, err)
		}
	}
}

// Return a necessary error.
// If len(el) == 0, it returns nil.
// If len(el) == 1, it returns el[0].
// Otherwise, it returns el (itself).
func (el *ErrorList) ToError() error {
	if el == nil || len(*el) == 0 {
		return nil
	}
	if len(*el) == 1 {
		return (*el)[0]
	}
	return el
}

// Remove duplicated and nil errors. An error is regarded as duplicated if
// its method Error returns the same string as that of a previous error.
func (el *ErrorList) Deduplicate() {
	if el == nil || len(*el) == 0 {
		return
	}
	set := make(map[string]bool)
	sda := sequence.NewSliceDynamicArray(el)
	sda.Filter(func(x interface{}) (keep bool) {
		if x == nil || x.(error) == nil {
			return false
		}
		errStr := x.(error).Error()
		if set[errStr] {
			return false
		}
		set[errStr] = true
		return true
	})
}

// Return the same as el.String().
func (el *ErrorList) Error() string {
	return el.String()
}

// If there is only one item, return it.
// Otherwise, return nil indicating that it cannot be unwrapped.
func (el *ErrorList) Unwrap() error {
	if el != nil && len(*el) == 1 {
		return (*el)[0]
	}
	return nil
}

// If it is empty, return "no errors".
//
// If there is only one item (denoted by t), then
// return t.Error() if t != nil, or return <nil> otherwise.
//
// If there are two or more items, return the number of errors,
// followed by an error array, in which every item is double-quoted
// in Go string literal. Especially, nil error item will be "<nil>".
func (el *ErrorList) String() string {
	if el == nil || len(*el) == 0 {
		return "no errors"
	}
	if len(*el) == 1 {
		err := (*el)[0]
		if err != nil {
			return err.Error()
		}
		return "<nil>"
	}
	var builder strings.Builder
	s := strconv.Itoa(len(*el))
	builder.WriteString(s)
	builder.WriteString(" errors: [")
	for i, err := range *el {
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
