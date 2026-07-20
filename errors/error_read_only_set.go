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

package errors

import (
	stderrors "errors"
	"iter"
	"maps"
	"slices"
)

// ErrorReadOnlySet is a read-only set of errors.
//
// It is designed to specify a set of errors and
// test whether future encountered errors belong to that set.
//
// The criterion for "belong to" depends on the specific implementation.
type ErrorReadOnlySet interface {
	// Len returns the number of errors in the error set.
	Len() int

	// Contains reports whether err belongs to this error set.
	//
	// The criterion for "belong to" depends on the specific implementation.
	Contains(err error) bool

	// Range calls handler on all items in the set.
	//
	// handler has one parameter: err (the error value),
	// and returns an indicator cont to report whether the iteration continues.
	Range(handler func(err error) (cont bool))

	// IterErrors returns an iterator over all errors in the set.
	// The order of iteration is consistent with the method Range.
	//
	// The returned iterator is never nil.
	IterErrors() iter.Seq[error]
}

// errorReadOnlySetEqual is an implementation of interface ErrorReadOnlySet.
//
// An error is treated as "belonging to" this set
// if it is equal to any item in this set.
type errorReadOnlySetEqual struct {
	errSet map[error]struct{}
}

// NewErrorReadOnlySetEqual creates a new ErrorReadOnlySet.
//
// The returned set treats that an error "belongs to" it
// if the error is equal to any item in this set.
func NewErrorReadOnlySetEqual(err ...error) ErrorReadOnlySet {
	eros := &errorReadOnlySetEqual{errSet: make(map[error]struct{}, len(err))}
	for _, e := range err {
		eros.errSet[e] = setMapV
	}

	return eros
}

func (eros *errorReadOnlySetEqual) Len() int {
	return len(eros.errSet)
}

// Contains reports whether err belongs to this error set.
//
// err is treated as "belonging to" this set
// if it is equal to any item in this set.
func (eros *errorReadOnlySetEqual) Contains(err error) bool {
	_, ok := eros.errSet[err]
	return ok
}

func (eros *errorReadOnlySetEqual) Range(
	handler func(err error) (cont bool),
) {
	if handler == nil {
		return
	}

	maps.Keys(eros.errSet)(handler)
}

func (eros *errorReadOnlySetEqual) IterErrors() iter.Seq[error] {
	return eros.Range
}

// errorReadOnlySetEqual is an implementation of interface ErrorReadOnlySet.
//
// An error err is treated as "belonging to" this set
// if there is an item x in this set such that errors.Is(err, x) returns true.
type errorReadOnlySetIs struct {
	errs []error
}

// NewErrorReadOnlySetIs creates a new ErrorReadOnlySet.
//
// The returned set treats that an error err "belongs to" it
// if there is an item x in this set such that errors.Is(err, x) returns true.
func NewErrorReadOnlySetIs(err ...error) ErrorReadOnlySet {
	errSet := make(map[error]struct{}, len(err))
	for _, e := range err {
		errSet[e] = setMapV
	}

	eros := &errorReadOnlySetIs{errs: make([]error, 0, len(errSet))}
	for e := range errSet {
		eros.errs = append(eros.errs, e)
	}

	return eros
}

func (eros *errorReadOnlySetIs) Len() int {
	return len(eros.errs)
}

// Contains reports whether err belongs to this error set.
//
// err is treated as "belonging to" this set if there is an item x
// in this set such that errors.Is(err, x) returns true.
func (eros *errorReadOnlySetIs) Contains(err error) bool {
	for _, target := range eros.errs {
		if stderrors.Is(err, target) {
			return true
		}
	}

	return false
}

func (eros *errorReadOnlySetIs) Range(handler func(err error) (cont bool)) {
	if handler == nil {
		return
	}

	slices.Values(eros.errs)(handler)
}

func (eros *errorReadOnlySetIs) IterErrors() iter.Seq[error] {
	return eros.Range
}

// errorReadOnlySetSameMessage is an implementation of
// interface ErrorReadOnlySet.
//
// An error is treated as "belonging to" this set
// if it has the same message as any item in this set.
//
// In particular, the message of nil error is treated as "<nil>".
type errorReadOnlySetSameMessage struct {
	errsSet map[string][]error
}

// NewErrorReadOnlySetSameMessage creates a new ErrorReadOnlySet.
//
// The returned set treats that an error "belongs to" it
// if the error has the same message as any item in this set.
//
// In particular, the message of nil error is treated as "<nil>".
func NewErrorReadOnlySetSameMessage(err ...error) ErrorReadOnlySet {
	errSetMap := make(map[string]map[error]struct{}, len(err))
	for _, e := range err {
		msg := "<nil>"
		if e != nil {
			msg = e.Error()
		}

		set := errSetMap[msg]
		if set == nil {
			set = map[error]struct{}{e: setMapV}
			errSetMap[msg] = set
		} else {
			set[e] = setMapV
		}
	}

	eros := &errorReadOnlySetSameMessage{
		errsSet: make(map[string][]error, len(errSetMap)),
	}
	for k, set := range errSetMap {
		errs := make([]error, 0, len(set))
		for e := range set {
			errs = append(errs, e)
		}

		eros.errsSet[k] = errs
	}

	return eros
}

func (eros *errorReadOnlySetSameMessage) Len() int {
	var n int
	for _, errs := range eros.errsSet {
		n += len(errs)
	}

	return n
}

// Contains reports whether err belongs to this error set.
//
// err is treated as "belonging to" this set
// if it has the same message as any item in this set.
//
// In particular, the message of nil error is treated as "<nil>".
func (eros *errorReadOnlySetSameMessage) Contains(err error) bool {
	if err == nil {
		return eros.errsSet[nilErrorString] != nil
	}

	return eros.errsSet[err.Error()] != nil
}

func (eros *errorReadOnlySetSameMessage) Range(
	handler func(err error) (cont bool),
) {
	if handler == nil {
		return
	}

	for _, errs := range eros.errsSet {
		slices.Values(errs)(handler)
	}
}

func (eros *errorReadOnlySetSameMessage) IterErrors() iter.Seq[error] {
	return eros.Range
}

// setMapV is the value for map[Type]struct{}.
// May be redundant.
var setMapV = struct{}{}
