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

import stderrors "errors"

// ErrorReadOnlySet is a read-only set of errors.
//
// It is designed to specify a set of errors and
// test whether future encountered errors belong to that set.
//
// The criterion for "belong to" depends on the specific implementation.
type ErrorReadOnlySet interface {
	// Len returns the number of errors in the error set.
	Len() int

	// Contain reports whether err belongs to this error set.
	//
	// The criterion for "belong to" depends on the specific implementation.
	Contain(err error) bool

	// Range calls handler on all items in the set.
	//
	// handler has one parameter: err (the error value),
	// and returns an indicator cont to report
	// whether to continue the iteration.
	Range(handler func(err error) (cont bool))
}

// errorReadOnlySetEqual is an implementation of interface ErrorReadOnlySet.
//
// An error is regarded as "belonging to" this set
// if it is equal to any item in this set.
type errorReadOnlySetEqual struct {
	set map[error]bool
}

// NewErrorReadOnlySetEqual creates a new ErrorReadOnlySet.
//
// The returned set regards that an error "belongs to" it
// if the error is equal to any item in this set.
func NewErrorReadOnlySetEqual(errs ...error) ErrorReadOnlySet {
	erose := &errorReadOnlySetEqual{make(map[error]bool, len(errs))}
	for _, err := range errs {
		erose.set[err] = true
	}
	return erose
}

// Len returns the number of errors in the error set.
func (erose *errorReadOnlySetEqual) Len() int {
	return len(erose.set)
}

// Contain reports whether err belongs to this error set.
//
// err is regarded as "belonging to" this set
// if it is equal to any item in this set.
func (erose *errorReadOnlySetEqual) Contain(err error) bool {
	return erose.set[err]
}

// Range calls handler on all items in the set.
//
// handler has one parameter: err (the error value),
// and returns an indicator cont to report whether to continue the iteration.
func (erose *errorReadOnlySetEqual) Range(handler func(err error) (cont bool)) {
	for err := range erose.set {
		if !handler(err) {
			return
		}
	}
}

// errorReadOnlySetEqual is an implementation of interface ErrorReadOnlySet.
//
// An error err is regarded as "belonging to" this set
// if there is an item x in this set such that errors.Is(err, x) returns true.
type errorReadOnlySetIs struct {
	set map[error]map[error]bool
	cis map[error]bool // a set of errors that have custom Is method along the Unwrap error chain
}

// NewErrorReadOnlySetIs creates a new ErrorReadOnlySet.
//
// The returned set regards that an error err "belongs to" it
// if there is an item x in this set such that errors.Is(err, x) returns true.
func NewErrorReadOnlySetIs(errs ...error) ErrorReadOnlySet {
	erosi := &errorReadOnlySetIs{set: make(map[error]map[error]bool, len(errs))}
	for _, err := range errs {
		var hasIs bool
		if _, ok := err.(ErrorIs); ok {
			hasIs = true
		}
		root := err
		for tmp := stderrors.Unwrap(err); tmp != nil; tmp = stderrors.Unwrap(tmp) {
			root = tmp
			if !hasIs {
				if _, ok := tmp.(ErrorIs); ok {
					hasIs = true
				}
			}
		}
		subset := erosi.set[root]
		if subset == nil {
			subset = map[error]bool{err: true}
			erosi.set[root] = subset
		} else {
			subset[err] = true
		}
		if hasIs {
			if erosi.cis == nil {
				erosi.cis = map[error]bool{err: true}
			} else {
				erosi.cis[err] = true
			}
		}
	}
	return erosi
}

// Len returns the number of errors in the error set.
func (erosi *errorReadOnlySetIs) Len() int {
	var n int
	for _, subset := range erosi.set {
		n += len(subset)
	}
	return n
}

// Contain reports whether err belongs to this error set.
//
// err is regarded as "belonging to" this set if there is an item x
// in this set such that errors.Is(err, x) returns true.
func (erosi *errorReadOnlySetIs) Contain(err error) bool {
	root := err
	for tmp := stderrors.Unwrap(err); tmp != nil; tmp = stderrors.Unwrap(tmp) {
		root = tmp
	}
	subset := erosi.set[root]
	for x := range subset {
		if stderrors.Is(err, x) {
			return true
		}
	}
	for x := range erosi.cis {
		if (subset == nil || !subset[x]) && stderrors.Is(err, x) {
			return true
		}
	}
	return false
}

// Range calls handler on all items in the set.
//
// handler has one parameter: err (the error value),
// and returns an indicator cont to report
// whether to continue the iteration.
func (erosi *errorReadOnlySetIs) Range(handler func(err error) (cont bool)) {
	for _, subset := range erosi.set {
		for err := range subset {
			if !handler(err) {
				return
			}
		}
	}
}

// errorReadOnlySetSameMessage is an implementation of
// interface ErrorReadOnlySet.
//
// An error is regarded as "belonging to" this set
// if it has the same message as any item in this set.
type errorReadOnlySetSameMessage struct {
	set map[string]map[error]bool
}

// NewErrorReadOnlySetSameMessage creates a new ErrorReadOnlySet.
//
// The returned set regards that an error "belongs to" it
// if the error has the same message as any item in this set.
func NewErrorReadOnlySetSameMessage(errs ...error) ErrorReadOnlySet {
	erossm := &errorReadOnlySetSameMessage{make(map[string]map[error]bool, len(errs))}
	for _, err := range errs {
		msg := "<nil>"
		if err != nil {
			msg = err.Error()
		}
		subset := erossm.set[msg]
		if subset == nil {
			subset = map[error]bool{err: true}
			erossm.set[msg] = subset
		} else {
			subset[err] = true
		}
	}
	return erossm
}

// Len returns the number of errors in the error set.
func (erossm *errorReadOnlySetSameMessage) Len() int {
	var n int
	for _, subset := range erossm.set {
		n += len(subset)
	}
	return n
}

// Contain reports whether err belongs to this error set.
//
// err is regarded as "belonging to" this set
// if it has the same message as any item in this set.
func (erossm *errorReadOnlySetSameMessage) Contain(err error) bool {
	if err == nil {
		return erossm.set["<nil>"] != nil
	}
	return erossm.set[err.Error()] != nil
}

// Range calls handler on all items in the set.
//
// handler has one parameter: err (the error value),
// and returns an indicator cont to report
// whether to continue the iteration.
func (erossm *errorReadOnlySetSameMessage) Range(handler func(err error) (cont bool)) {
	for _, subset := range erossm.set {
		for err := range subset {
			if !handler(err) {
				return
			}
		}
	}
}
