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

import stderrors "errors"

// Export functions from standard package errors for convenience.

// New directly calls standard errors.New.
func New(msg string) error {
	return stderrors.New(msg)
}

// Unwrap directly calls standard errors.Unwrap.
func Unwrap(err error) error {
	return stderrors.Unwrap(err)
}

// Is directly calls standard errors.Is.
func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

// As directly calls standard errors.As.
func As(err error, target interface{}) bool {
	return stderrors.As(err, target)
}
