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
	"strconv"
	"strings"
)

// ErrorList is a list of errors, to collect multiple errors.
//
// If it is empty, it reports "no error".
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

// Combine combines multiple errors into an ErrorList.
// Note that nil error will be discarded.
// It always returns a non-nil *ErrorList.
func Combine(errs ...error) *ErrorList {
	el := make(ErrorList, 0, len(errs))
	el.Append(errs...)
	return &el
}

// Len returns the number of errors in the error list.
func (el ErrorList) Len() int {
	return len(el)
}

// Append appends errs to the error list.
// Note that nil error will be discarded.
func (el *ErrorList) Append(errs ...error) {
	for _, err := range errs {
		if err != nil {
			*el = append(*el, err)
		}
	}
}

// ToError returns a necessary error.
// If el.Len() == 0, it returns nil.
// If el.Len() == 1, it returns el[0].
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

// Deduplicate removes duplicated and nil errors.
// An error is regarded as duplicated if its method Error returns
// the same string as that of a previous error.
func (el *ErrorList) Deduplicate() {
	if el == nil || len(*el) == 0 {
		return
	}
	set := make(map[string]bool)
	n := 0
	for i := 0; i < len(*el); i++ {
		x := (*el)[i]
		if x != nil {
			errStr := x.Error()
			if !set[errStr] {
				set[errStr] = true
				(*el)[n] = x
				n++
			}
		}
	}
	if n == len(*el) {
		return
	}
	for i := n; i < len(*el); i++ {
		(*el)[i] = nil
	}
	*el = (*el)[:n]
}

// Error returns the same as el.String().
func (el *ErrorList) Error() string {
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
func (el *ErrorList) String() string {
	if el == nil || len(*el) == 0 {
		return "no error"
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
