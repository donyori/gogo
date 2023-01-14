// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2023  Yuan Gao
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

package constraints_test

import (
	"testing"

	"github.com/donyori/gogo/constraints"
)

type (
	myInt        int
	myInt8       int8
	myInt16      int16
	myInt32      int32
	myRune       rune
	myInt64      int64
	myUint       uint
	myUint8      uint8
	myByte       byte
	myUint16     uint16
	myUint32     uint32
	myUint64     uint64
	myUintptr    uintptr
	myFloat32    float32
	myFloat64    float64
	myComplex64  complex64
	myComplex128 complex128
	myString     string
	myByteSlice  []byte
)

var (
	ints        []int
	int8s       []int8
	int16s      []int16
	int32s      []int32
	runes       []rune
	int64s      []int64
	uints       []uint
	uint8s      []uint8
	bytes       []byte
	uint16s     []uint16
	uint32s     []uint32
	uint64s     []uint64
	uintptrs    []uintptr
	float32s    []float32
	float64s    []float64
	complex64s  []complex64
	complex128s []complex128
	strings     []string
)

var (
	myInts        []myInt
	myInt8s       []myInt8
	myInt16s      []myInt16
	myInt32s      []myInt32
	myRunes       []myRune
	myInt64s      []myInt64
	myUints       []myUint
	myUint8s      []myUint8
	myBytes       []myByte
	myUint16s     []myUint16
	myUint32s     []myUint32
	myUint64s     []myUint64
	myUintptrs    []myUintptr
	myFloat32s    []myFloat32
	myFloat64s    []myFloat64
	myComplex64s  []myComplex64
	myComplex128s []myComplex128
	myStrings     []myString
)

func TestCompilePredeclaredOrdered(t *testing.T) {
	predeclaredSmallest(ints)
	predeclaredSmallest(int8s)
	predeclaredSmallest(int16s)
	predeclaredSmallest(int32s)
	predeclaredSmallest(runes)
	predeclaredSmallest(int64s)
	predeclaredSmallest(uints)
	predeclaredSmallest(uint8s)
	predeclaredSmallest(bytes)
	predeclaredSmallest(uint16s)
	predeclaredSmallest(uint32s)
	predeclaredSmallest(uint64s)
	predeclaredSmallest(uintptrs)
	predeclaredSmallest(float32s)
	predeclaredSmallest(float64s)
	predeclaredSmallest(strings)

	// The following statements should be invalid.
	//
	//  predeclaredSmallest(myInts)
	//	predeclaredSmallest(myInt8s)
	//	predeclaredSmallest(myInt16s)
	//	predeclaredSmallest(myInt32s)
	//	predeclaredSmallest(myRunes)
	//	predeclaredSmallest(myInt64s)
	//	predeclaredSmallest(myUints)
	//	predeclaredSmallest(myUint8s)
	//	predeclaredSmallest(myBytes)
	//	predeclaredSmallest(myUint16s)
	//	predeclaredSmallest(myUint32s)
	//	predeclaredSmallest(myUint64s)
	//	predeclaredSmallest(myUintptrs)
	//	predeclaredSmallest(myFloat32s)
	//	predeclaredSmallest(myFloat64s)
	//	predeclaredSmallest(myStrings)
	//
	//  predeclaredSmallest(complex64s)
	//  predeclaredSmallest(complex128s)
	//  predeclaredSmallest(myComplex64s)
	//  predeclaredSmallest(myComplex128s)
}

func predeclaredSmallest[T constraints.PredeclaredOrdered](s []T) (r T, ok bool) {
	if len(s) == 0 {
		return
	}
	r = s[0]
	for _, x := range s[1:] {
		if x < r {
			r = x
		}
	}
	return r, true
}

func TestCompileOrdered(t *testing.T) {
	smallest(ints)
	smallest(int8s)
	smallest(int16s)
	smallest(int32s)
	smallest(runes)
	smallest(int64s)
	smallest(uints)
	smallest(uint8s)
	smallest(bytes)
	smallest(uint16s)
	smallest(uint32s)
	smallest(uint64s)
	smallest(uintptrs)
	smallest(float32s)
	smallest(float64s)
	smallest(strings)

	smallest(myInts)
	smallest(myInt8s)
	smallest(myInt16s)
	smallest(myInt32s)
	smallest(myRunes)
	smallest(myInt64s)
	smallest(myUints)
	smallest(myUint8s)
	smallest(myBytes)
	smallest(myUint16s)
	smallest(myUint32s)
	smallest(myUint64s)
	smallest(myUintptrs)
	smallest(myFloat32s)
	smallest(myFloat64s)
	smallest(myStrings)

	// The following statements should be invalid.
	//
	//  smallest(complex64s)
	//  smallest(complex128s)
	//  smallest(myComplex64s)
	//  smallest(myComplex128s)
}

func smallest[T constraints.Ordered](s []T) (r T, ok bool) {
	if len(s) == 0 {
		return
	}
	r = s[0]
	for _, x := range s[1:] {
		if x < r {
			r = x
		}
	}
	return r, true
}

func TestCompileAddable(t *testing.T) {
	sum(ints)
	sum(int8s)
	sum(int16s)
	sum(int32s)
	sum(runes)
	sum(int64s)
	sum(uints)
	sum(uint8s)
	sum(bytes)
	sum(uint16s)
	sum(uint32s)
	sum(uint64s)
	sum(uintptrs)
	sum(float32s)
	sum(float64s)
	sum(complex64s)
	sum(complex128s)
	sum(strings)

	sum(myInts)
	sum(myInt8s)
	sum(myInt16s)
	sum(myInt32s)
	sum(myRunes)
	sum(myInt64s)
	sum(myUints)
	sum(myUint8s)
	sum(myBytes)
	sum(myUint16s)
	sum(myUint32s)
	sum(myUint64s)
	sum(myUintptrs)
	sum(myFloat32s)
	sum(myFloat64s)
	sum(myComplex64s)
	sum(myComplex128s)
	sum(myStrings)
}

func sum[T constraints.Addable](s []T) T {
	var r T
	if len(s) == 0 {
		return r
	}
	for _, x := range s {
		r += x
	}
	return r
}

func TestCompileByteString(t *testing.T) {
	var str string
	var bs []byte
	var myS myString
	var myB myByteSlice

	concatBytes(str, bs)
	concatBytes(bs, myS)
	concatBytes(myS, myB)
	concatBytes(myB, str)
}

func concatBytes[T1, T2 constraints.ByteString](s1 T1, s2 T2) []byte {
	s := make([]byte, len(s1)+len(s2))
	n := copy(s, s1)
	copy(s[n:], s2)
	return s
}
