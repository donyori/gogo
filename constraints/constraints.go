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

package constraints

// PredeclaredSignedInteger is a constraint that matches the five predeclared
// signed integer types: int, int8, int16, int32 (rune), and int64.
type PredeclaredSignedInteger interface {
	int | int8 | int16 | int32 | int64
}

// SignedInteger is a constraint for signed integers.
// It matches any type whose underlying type is one of int, int8, int16,
// int32 (rune), and int64.
type SignedInteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// PredeclaredUnsignedInteger is a constraint that matches the six predeclared
// unsigned integer types: uint, uint8 (byte), uint16, uint32, uint64,
// and uintptr.
type PredeclaredUnsignedInteger interface {
	uint | uint8 | uint16 | uint32 | uint64 | uintptr
}

// UnsignedInteger is a constraint for unsigned integers.
// It matches any type whose underlying type is one of uint, uint8 (byte),
// uint16, uint32, uint64, and uintptr.
type UnsignedInteger interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// PredeclaredInteger is a constraint that matches the eleven predeclared
// integer types: int, int8, int16, int32 (rune), int64, uint, uint8 (byte),
// uint16, uint32, uint64, and uintptr.
type PredeclaredInteger interface {
	PredeclaredSignedInteger | PredeclaredUnsignedInteger
}

// Integer is a constraint for integers.
// It matches any type whose underlying type is one of int, int8, int16,
// int32 (rune), int64, uint, uint8 (byte), uint16, uint32, uint64, and uintptr.
type Integer interface {
	SignedInteger | UnsignedInteger
}

// PredeclaredFloat is a constraint that matches the two predeclared
// floating-point number types: float32 and float64.
type PredeclaredFloat interface {
	float32 | float64
}

// Float is a constraint for floating-point numbers.
// It matches any type whose underlying type is float32 or float64.
type Float interface {
	~float32 | ~float64
}

// PredeclaredReal is a constraint that matches the predeclared real number
// types: int, int8, int16, int32 (rune), int64, uint, uint8 (byte), uint16,
// uint32, uint64, uintptr, float32 and float64.
type PredeclaredReal interface {
	PredeclaredInteger | PredeclaredFloat
}

// Real is a constraint for real numbers.
// It matches any type whose underlying type is one of int, int8, int16,
// int32 (rune), int64, uint, uint8 (byte), uint16, uint32, uint64, uintptr,
// float32, and float64.
type Real interface {
	Integer | Float
}

// PredeclaredComplex is a constraint that matches the two predeclared
// complex number types: complex64 and complex128.
type PredeclaredComplex interface {
	complex64 | complex128
}

// Complex is a constraint for complex numbers.
// It matches any type whose underlying type is complex64 or complex128.
type Complex interface {
	~complex64 | ~complex128
}

// PredeclaredNumeric is a constraint that matches the predeclared numeric
// types: int, int8, int16, int32 (rune), int64, uint, uint8 (byte), uint16,
// uint32, uint64, uintptr, float32, float64, complex64, and complex128.
type PredeclaredNumeric interface {
	PredeclaredReal | PredeclaredComplex
}

// Numeric is a constraint for numerics.
// It matches any type whose underlying type is one of int, int8, int16,
// int32 (rune), int64, uint, uint8 (byte), uint16, uint32, uint64, uintptr,
// float32, float64, complex64, and complex128.
type Numeric interface {
	Real | Complex
}

// PredeclaredOrdered is a constraint that matches the predeclared ordered
// types: int, int8, int16, int32 (rune), int64, uint, uint8 (byte), uint16,
// uint32, uint64, uintptr, float32, float64, and string.
//
// An ordered type is one that supports the <, <=, >, and >= operators.
type PredeclaredOrdered interface {
	PredeclaredReal | string
}

// Ordered is a constraint that matches any ordered type.
//
// An ordered type is one that supports the <, <=, >, and >= operators.
type Ordered interface {
	Real | ~string
}

// PredeclaredStrictWeakOrdered is a constraint that matches the predeclared
// ordered types that implement a strict weak ordering, including int, int8,
// int16, int32 (rune), int64, uint, uint8 (byte), uint16, uint32, uint64,
// uintptr, and string.
//
// An ordered type is one that supports the <, <=, >, and >= operators.
//
// For strict weak ordering,
// see <https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings>.
type PredeclaredStrictWeakOrdered interface {
	PredeclaredInteger | string
}

// StrictWeakOrdered is a constraint that matches any ordered type that
// implements a strict weak ordering.
//
// An ordered type is one that supports the <, <=, >, and >= operators.
//
// For strict weak ordering,
// see <https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings>.
type StrictWeakOrdered interface {
	Integer | ~string
}

// PredeclaredAddable is a constraint that matches the predeclared addable
// types: int, int8, int16, int32 (rune), int64, uint, uint8 (byte), uint16,
// uint32, uint64, uintptr, float32, float64, complex64, complex128, and string.
//
// An addable type is one that supports the + operator.
type PredeclaredAddable interface {
	PredeclaredNumeric | string
}

// Addable is a constraint that matches any addable type.
//
// An addable type is one that supports the + operator.
type Addable interface {
	Numeric | ~string
}

// PredeclaredByteString is a constraint that matches the predeclared
// byte string types: []byte and string.
type PredeclaredByteString interface {
	[]byte | string
}

// ByteString is a constraint for byte strings.
// It matches any type whose underlying type is []byte or string.
type ByteString interface {
	~[]byte | ~string
}

// Slice is a constraint for slices.
// It matches any type whose underlying type is []Elem,
// where Elem can be any type.
type Slice[Elem any] interface {
	~[]Elem
}

// Map is a constraint for maps.
// It matches any type whose underlying type is map[Key]Value,
// where Key is a comparable type and Value can be any type.
type Map[Key comparable, Value any] interface {
	~map[Key]Value
}
