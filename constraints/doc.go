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

// Package constraints provides some basic constraints that can be used by
// other types and functions and embedded in other constraints.
//
// Currently, this package provides constraints for numeric types
// (including signed and unsigned integers, floating-point numbers,
// and complex numbers), byte sequence types
// (including byte slices and strings), slices, and maps.
//
// These constraints are helpful to apply arithmetic operators and
// comparison operators in generic code.
//
// The arithmetic operators include sum (+), difference (-), product (*),
// quotient (/), remainder (%), bitwise AND (&), bitwise OR (|),
// bitwise XOR (^), bit clear (AND NOT, &^), left shift (<<),
// and right shift (>>).
// See <https://go.dev/ref/spec#Arithmetic_operators> for details.
//
// The comparison operators include equal (==), not equal (!=), less (<),
// less or equal (<=), greater (>), greater or equal (>=).
// See <https://go.dev/ref/spec#Comparison_operators> for details.
package constraints
