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

package function

import (
	"reflect"

	"github.com/donyori/gogo/errors"
)

// LessFunc is a function to test whether a < b.
//
// To use LessFunc in cases where a transitive ordering is required,
// such as sorting, it must implement a transitive ordering.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
type LessFunc func(a, b interface{}) bool

// Not returns a negative function to test whether !(a < b).
func (lf LessFunc) Not() LessFunc {
	return func(a, b interface{}) bool {
		return !lf(a, b)
	}
}

// Reverse returns a reverse function to test whether b < a.
func (lf LessFunc) Reverse() LessFunc {
	return func(a, b interface{}) bool {
		return lf(b, a)
	}
}

// ToEqual returns an EqualFunc to test whether a == b.
// The return function reports true if and only if
//  !(less(a, b) || less(b, a))
func (lf LessFunc) ToEqual() EqualFunc {
	return func(a, b interface{}) bool {
		return !(lf(a, b) || lf(b, a))
	}
}

// IntLess is a prefab LessFunc for int.
//
// The nil interface{} is treated as a zero int 0 for convenience.
//
// It panics if any non-nil input variable is not int.
var IntLess LessFunc = intLess

// intLess is an implementation of function IntLess.
func intLess(a, b interface{}) bool {
	var ia, ib int
	var ok bool
	if a != nil {
		ia, ok = a.(int)
		if !ok {
			panic(errors.AutoMsg("a is not int"))
		}
	}
	if b != nil {
		ib, ok = b.(int)
		if !ok {
			panic(errors.AutoMsg("b is not int"))
		}
	}
	return ia < ib
}

// Float64Less is a prefab LessFunc for float64.
//
// It implements a transitive ordering:
//  - if both Float64Less(a, b) and Float64Less(b, c) are true, then Float64Less(a, c) must be true as well.
//  - if both Float64Less(a, b) and Float64Less(b, c) are false, then Float64Less(a, c) must be false as well.
// It treats NaN values as less than any others.
//
// The nil interface{} is treated as a zero float64 0.0 for convenience.
//
// It panics if any non-nil input variable is not float64.
var Float64Less LessFunc = float64Less

// float64Less is an implementation of function Float64Less.
func float64Less(a, b interface{}) bool {
	var fa, fb float64
	var ok bool
	if a != nil {
		fa, ok = a.(float64)
		if !ok {
			panic(errors.AutoMsg("a is not float64"))
		}
	}
	if b != nil {
		fb, ok = b.(float64)
		if !ok {
			panic(errors.AutoMsg("b is not float64"))
		}
	}
	return fa < fb || (isNaN(fa) && !isNaN(fb))
}

// StringLess is a prefab LessFunc for string.
//
// The nil interface{} is treated as an empty string "" for convenience.
//
// It panics if any non-nil input variable is not string.
var StringLess LessFunc = stringLess

// stringLess is an implementation of function StringLess.
func stringLess(a, b interface{}) bool {
	var sa, sb string
	var ok bool
	if a != nil {
		sa, ok = a.(string)
		if !ok {
			panic(errors.AutoMsg("a is not string"))
		}
	}
	if b != nil {
		sb, ok = b.(string)
		if !ok {
			panic(errors.AutoMsg("b is not string"))
		}
	}
	return sa < sb
}

// BuiltinRealNumberLess is a prefab LessFunc for built-in real numbers,
// including
//  int, int8, int16, int32, int64,
//  uint, uintptr, uint8, uint16, uint32, uint64,
//  float32, float64,
//  byte, // an alias for uint8
//  rune. // an alias for int32
//
// It implements a transitive ordering:
//  - if both BuiltinRealNumberLess(a, b) and BuiltinRealNumberLess(b, c) are true, then BuiltinRealNumberLess(a, c) must be true as well.
//  - if both BuiltinRealNumberLess(a, b) and BuiltinRealNumberLess(b, c) are false, then BuiltinRealNumberLess(a, c) must be false as well.
// It treats NaN values as less than any others.
//
// The nil interface{} is treated as a zero real number for convenience.
//
// It panics if any non-nil input variable is not a built-in real number,
// as listed above.
var BuiltinRealNumberLess LessFunc = builtinRealNumberLess

// builtinRealNumberLess is an implementation of function BuiltinRealNumberLess.
func builtinRealNumberLess(a, b interface{}) bool {
	var flag uint8
	var ia, ib int64
	var ua, ub uint64
	var fa, fb float64

	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
	switch va.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ia = va.Int()
		flag = 0b00_01
	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ua = va.Uint()
		flag = 0b00_10
	case reflect.Float32, reflect.Float64:
		fa = va.Float()
		flag = 0b00_11
	case reflect.Invalid:
	default:
		panic(errors.AutoMsg("a is not a built-in real number"))
	}
	switch vb.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ib = vb.Int()
		flag |= 0b01_00
	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ub = vb.Uint()
		flag |= 0b10_00
	case reflect.Float32, reflect.Float64:
		fb = vb.Float()
		flag |= 0b11_00
	case reflect.Invalid:
	default:
		panic(errors.AutoMsg("b is not a built-in real number"))
	}

	switch flag {
	case 0b00_00, 0b00_10:
		return false // "0 < 0" and "unsigned integer < 0" are false
	case 0b00_01:
		return ia < 0
	case 0b00_11:
		return fa < 0. || isNaN(fa)
	case 0b01_00:
		return ib > 0
	case 0b01_01:
		return ia < ib
	case 0b01_10:
		return ib > 0 && ua < uint64(ib)
	case 0b01_11:
		return fa < float64(ib) || isNaN(fa)
	case 0b10_00:
		return ub > 0
	case 0b10_01:
		return ia < 0 || uint64(ia) < ub
	case 0b10_10:
		return ua < ub
	case 0b10_11:
		return fa < float64(ub) || isNaN(fa)
	case 0b11_00:
		return fb > 0.
	case 0b11_01:
		return float64(ia) < fb
	case 0b11_10:
		return float64(ua) < fb
	case 0b11_11:
		return fa < fb || (isNaN(fa) && !isNaN(fb))
	default:
		// This should never happen, but will act as a safeguard for later,
		// as a default value doesn't make sense here.
		panic(errors.AutoMsg("flag is invalid, which should never happen"))
	}
}

// isNaN is a copy of math.IsNaN to avoid a dependency on the math package.
func isNaN(f float64) bool {
	return f != f
}
