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

package unequal

import (
	"maps"
	"slices"

	"github.com/donyori/gogo/constraints"
	"github.com/donyori/gogo/errors"
)

// Slice tests whether two slices are unequal,
// where the type of slice items is comparable.
//
// A nil slice and an empty slice are considered unequal.
func Slice[S constraints.Slice[T], T comparable](s1, s2 S) bool {
	return (s1 == nil) != (s2 == nil) || !slices.Equal(s1, s2)
}

// Map tests whether two maps are unequal,
// where the type of keys and the type of values are comparable.
//
// A nil map and an empty map are considered unequal.
func Map[M constraints.Map[K, V], K, V comparable](m1, m2 M) bool {
	return (m1 == nil) != (m2 == nil) || !maps.Equal(m1, m2)
}

// ErrorUnwrapAuto tests whether two errors are unequal.
//
// The errors are unwrapped by
// github.com/donyori/gogo/errors.UnwrapAllAutoWrappedErrors
// and then compared by "!=".
func ErrorUnwrapAuto(err1, err2 error) bool {
	err1, _ = errors.UnwrapAllAutoWrappedErrors(err1)
	err2, _ = errors.UnwrapAllAutoWrappedErrors(err2)
	return err1 != err2 // compare the interface directly, don't use errors.Is
}
