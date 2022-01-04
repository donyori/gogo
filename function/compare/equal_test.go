// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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

package compare

import (
	"reflect"
	"testing"

	"github.com/donyori/gogo/errors"
)

func TestEqual(t *testing.T) {
	pairs := [][2]interface{}{
		{nil, nil},
		{1, nil},
		{nil, 1},
		{1, 1},
		{1, 0},
		{0, 1},
		{1., 1.},
		{1, 1.},
		{1., 1},
	}
	for _, pair := range pairs {
		if r := Equal(pair[0], pair[1]); r != (pair[0] == pair[1]) {
			t.Errorf("Equal(%v, %v): %t.", pair[0], pair[1], r)
		}
	}
}

func TestEqualFunc_Not(t *testing.T) {
	pairs := [][2]interface{}{
		{nil, nil},
		{1, nil},
		{nil, 1},
		{1, 1},
		{1, 0},
		{0, 1},
		{1., 1.},
		{1, 1.},
		{1., 1},
	}
	eq := Equal
	nEq := eq.Not()
	for _, pair := range pairs {
		r1 := !eq(pair[0], pair[1])
		r2 := nEq(pair[0], pair[1])
		if r1 != r2 {
			t.Errorf("nEq(%v, %v) != !eq(%[1]v, %v).", pair[0], pair[1])
		}
	}
}

func TestBytesEqual(t *testing.T) {
	eqPairs, neqPairs := testMkEqNeqPairs([][]interface{}{
		{nil, []byte(nil), []byte{}, ""}, // Empty.
		{[]byte("1234"), "1234"},         // Even length - 1.
		{[]byte("1224"), "1224"},         // Even length - 2.
		{[]byte("12345"), "12345"},       // Odd length - 1.
		{[]byte("12245"), "12245"},       // Odd length - 2.
	}, 0, 0)
	for i, pair := range eqPairs {
		if r := BytesEqual(pair[0], pair[1]); !r {
			t.Errorf("eqPairs Case %d: BytesEqual(a, b): false. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
		if r := BytesEqual(pair[1], pair[0]); !r {
			t.Errorf("eqPairs Case %d Reverse: BytesEqual(a, b): false. a: %#v, b: %#v.", i, pair[1], pair[0])
		}
	}
	for i, pair := range neqPairs {
		if r := BytesEqual(pair[0], pair[1]); r {
			t.Errorf("neqPairs Case %d: BytesEqual(a, b): true. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
		if r := BytesEqual(pair[1], pair[0]); r {
			t.Errorf("neqPairs Case %d Reverse: BytesEqual(a, b): true. a: %#v, b: %#v.", i, pair[1], pair[0])
		}
	}
}

func TestIntsEqual(t *testing.T) {
	eqPairs, neqPairs := testMkEqNeqPairs([][]interface{}{
		{nil, []int(nil), []int{}}, // Empty.
		{[]int{1, 2, 3, 4}},        // Even length - 1.
		{[]int{1, 2, 2, 4}},        // Even length - 2.
		{[]int{1, 2, 3, 4, 5}},     // Odd length - 1.
		{[]int{1, 2, 2, 4, 5}},     // Odd length - 2.
	}, 0, 0)
	for i, pair := range eqPairs {
		if r := IntsEqual(pair[0], pair[1]); !r {
			t.Errorf("eqPairs Case %d: IntsEqual(a, b): false. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
		if r := IntsEqual(pair[1], pair[0]); !r {
			t.Errorf("eqPairs Case %d Reverse: IntsEqual(a, b): false. a: %#v, b: %#v.", i, pair[1], pair[0])
		}
	}
	for i, pair := range neqPairs {
		if r := IntsEqual(pair[0], pair[1]); r {
			t.Errorf("neqPairs Case %d: IntsEqual(a, b): true. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
		if r := IntsEqual(pair[1], pair[0]); r {
			t.Errorf("neqPairs Case %d Reverse: IntsEqual(a, b): true. a: %#v, b: %#v.", i, pair[1], pair[0])
		}
	}
}

func TestFloat64sEqual(t *testing.T) {
	eqPairs, neqPairs := testMkEqNeqPairs([][]interface{}{
		{nil, []float64(nil), []float64{}}, // Empty.
		{[]float64{1., 2., 3., 4.}},        // Even length - 1.
		{[]float64{1., 2., 2., 4.}},        // Even length - 2.
		{[]float64{1., 2., 3., 4., 5.}},    // Odd length - 1.
		{[]float64{1., 2., 2., 4., 5.}},    // Odd length - 2.
	}, 0, 0)
	for i, pair := range eqPairs {
		if r := Float64sEqual(pair[0], pair[1]); !r {
			t.Errorf("eqPairs Case %d: Float64sEqual(a, b): false. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
		if r := Float64sEqual(pair[1], pair[0]); !r {
			t.Errorf("eqPairs Case %d Reverse: Float64sEqual(a, b): false. a: %#v, b: %#v.", i, pair[1], pair[0])
		}
	}
	for i, pair := range neqPairs {
		if r := Float64sEqual(pair[0], pair[1]); r {
			t.Errorf("neqPairs Case %d: Float64sEqual(a, b): true. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
		if r := Float64sEqual(pair[1], pair[0]); r {
			t.Errorf("neqPairs Case %d Reverse: Float64sEqual(a, b): true. a: %#v, b: %#v.", i, pair[1], pair[0])
		}
	}
}

func TestStringsEqual(t *testing.T) {
	eqPairs, neqPairs := testMkEqNeqPairs([][]interface{}{
		{nil, []string(nil), []string{}},    // Empty.
		{[]string{"1", "2", "3", "4"}},      // Even length - 1.
		{[]string{"1", "2", "2", "4"}},      // Even length - 2.
		{[]string{"1", "2", "3", "4", "5"}}, // Odd length - 1.
		{[]string{"1", "2", "2", "4", "5"}}, // Odd length - 2.
	}, 0, 0)
	for i, pair := range eqPairs {
		if r := StringsEqual(pair[0], pair[1]); !r {
			t.Errorf("eqPairs Case %d: StringsEqual(a, b): false. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
		if r := StringsEqual(pair[1], pair[0]); !r {
			t.Errorf("eqPairs Case %d Reverse: StringsEqual(a, b): false. a: %#v, b: %#v.", i, pair[1], pair[0])
		}
	}
	for i, pair := range neqPairs {
		if r := StringsEqual(pair[0], pair[1]); r {
			t.Errorf("neqPairs Case %d: StringsEqual(a, b): true. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
		if r := StringsEqual(pair[1], pair[0]); r {
			t.Errorf("neqPairs Case %d Reverse: StringsEqual(a, b): true. a: %#v, b: %#v.", i, pair[1], pair[0])
		}
	}
}

func TestGeneralSliceEqual(t *testing.T) {
	eqPairs, neqPairs := testMkEqNeqPairs([][]interface{}{
		{nil, []interface{}(nil), []interface{}{}}, // Empty.
		{[]interface{}{1, 2, 3, 4}},                // Even length - 1.
		{[]interface{}{1., 2., 3., 4.}},            // Even length - 2.
		{[]interface{}{1, 2, 2, 4}},                // Even length - 3.
		{[]interface{}{1., 2., 2., 4.}},            // Even length - 4.
		{[]interface{}{1, 2, 3, 4, 5}},             // Odd length - 1.
		{[]interface{}{1., 2., 3., 4., 5.}},        // Odd length - 2.
		{[]interface{}{1, 2, 2, 4, 5}},             // Odd length - 3.
		{[]interface{}{1., 2., 2., 4., 5.}},        // Odd length - 4.
	}, 1, 2)
	eqPairs = append(eqPairs, [2]interface{}{
		[]interface{}{1, 2., '3', byte('4')},
		[]interface{}{1, 2., '3', byte('4')},
	})
	neqPairs = append(neqPairs, [2]interface{}{
		// It should regard as unequal since []int is not comparable.
		[]interface{}{[]int{1, 2, 3}},
		[]interface{}{[]int{1, 2, 3}},
	}, [2]interface{}{
		// 2. (type: float64) != 2 (type: int).
		[]interface{}{1, 2., '3', byte('4')},
		[]interface{}{1, 2, '3', byte('4')},
	})
	for i, pair := range eqPairs {
		if r := GeneralSliceEqual(pair[0], pair[1]); !r {
			t.Errorf("eqPairs Case %d: GeneralSliceEqual(a, b): false. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
		if r := GeneralSliceEqual(pair[1], pair[0]); !r {
			t.Errorf("eqPairs Case %d Reverse: GeneralSliceEqual(a, b): false. a: %#v, b: %#v.", i, pair[1], pair[0])
		}
	}
	for i, pair := range neqPairs {
		if r := GeneralSliceEqual(pair[0], pair[1]); r {
			t.Errorf("neqPairs Case %d: GeneralSliceEqual(a, b): true. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
		if r := GeneralSliceEqual(pair[1], pair[0]); r {
			t.Errorf("neqPairs Case %d Reverse: GeneralSliceEqual(a, b): true. a: %#v, b: %#v.", i, pair[1], pair[0])
		}
	}
}

func TestSliceItemEqual(t *testing.T) {
	eqPairs, neqPairs := testMkEqNeqPairs([][]interface{}{
		{nil, []byte(nil), []byte{}, "", []int(nil), []int{},
			[]float64(nil), []float64{}, []string(nil), []string{},
			[]interface{}(nil), []interface{}{}}, // Empty.
		{[]byte("1234"), "1234", []interface{}{byte('1'), byte('2'), byte('3'), byte('4')}},              // Bytes - Even length - 1.
		{[]byte("1224"), "1224", []interface{}{byte('1'), byte('2'), byte('2'), byte('4')}},              // Bytes - Even length - 2.
		{[]byte("12345"), "12345", []interface{}{byte('1'), byte('2'), byte('3'), byte('4'), byte('5')}}, // Bytes - Odd length - 1.
		{[]byte("12245"), "12245", []interface{}{byte('1'), byte('2'), byte('2'), byte('4'), byte('5')}}, // Bytes - Odd length - 2.
		{[]int{1, 2, 3, 4}, []interface{}{1, 2, 3, 4}},                                                   // Ints - Even length - 1.
		{[]int{1, 2, 2, 4}, []interface{}{1, 2, 2, 4}},                                                   // Ints - Even length - 2.
		{[]int{1, 2, 3, 4, 5}, []interface{}{1, 2, 3, 4, 5}},                                             // Ints - Odd length - 1.
		{[]int{1, 2, 2, 4, 5}, []interface{}{1, 2, 2, 4, 5}},                                             // Ints - Odd length - 2.
		{[]float64{1., 2., 3., 4.}, []interface{}{1., 2., 3., 4.}},                                       // Floats - Even length - 1.
		{[]float64{1., 2., 2., 4.}, []interface{}{1., 2., 2., 4.}},                                       // Floats - Even length - 2.
		{[]float64{1., 2., 3., 4., 5.}, []interface{}{1., 2., 3., 4., 5.}},                               // Floats - Odd length - 1.
		{[]float64{1., 2., 2., 4., 5.}, []interface{}{1., 2., 2., 4., 5.}},                               // Floats - Odd length - 2.
		{[]string{"1", "2", "3", "4"}, []interface{}{"1", "2", "3", "4"}},                                // Strings - Even length - 1.
		{[]string{"1", "2", "2", "4"}, []interface{}{"1", "2", "2", "4"}},                                // Strings - Even length - 2.
		{[]string{"1", "2", "3", "4", "5"}, []interface{}{"1", "2", "3", "4", "5"}},                      // Strings - Even length - 1.
		{[]string{"1", "2", "2", "4", "5"}, []interface{}{"1", "2", "2", "4", "5"}},                      // Strings - Even length - 2.
	}, 1, 2)
	eqPairs = append(eqPairs, [2]interface{}{
		[]interface{}{1, 2., '3', byte('4')},
		[]interface{}{1, 2., '3', byte('4')},
	})
	neqPairs = append(neqPairs, [2]interface{}{
		// It should regard as unequal since []int is not comparable.
		[]interface{}{[]int{1, 2, 3}},
		[]interface{}{[]int{1, 2, 3}},
	}, [2]interface{}{
		// 2. (type: float64) != 2 (type: int).
		[]interface{}{1, 2., '3', byte('4')},
		[]interface{}{1, 2, '3', byte('4')},
	})

	for i, pair := range eqPairs {
		if r := SliceItemEqual(pair[0], pair[1]); !r {
			t.Errorf("eqPairs Case %d: SliceItemEqual(a, b): false. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
		if r := SliceItemEqual(pair[1], pair[0]); !r {
			t.Errorf("eqPairs Case %d Reverse: SliceItemEqual(a, b): false. a: %#v, b: %#v.", i, pair[1], pair[0])
		}
	}
	for i, pair := range neqPairs {
		if r := SliceItemEqual(pair[0], pair[1]); r {
			t.Errorf("neqPairs Case %d: SliceItemEqual(a, b): true. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
		if r := SliceItemEqual(pair[1], pair[0]); r {
			t.Errorf("neqPairs Case %d Reverse: SliceItemEqual(a, b): true. a: %#v, b: %#v.", i, pair[1], pair[0])
		}
	}
}

// testMkEqNeqPairs generates eqPairs and neqPairs for
// testing the prefab functions for EqualFunc.
//
// An item in eqPairs consists of two equal elements.
// An item in neqPairs consists of two unequal elements.
//
// eqGroups consist of groups of equivalent elements.
// Items in the same group are equal to each other.
// Items in the different groups are unequal to each other.
//
// eqExCap is the additional capacity of eqPairs.
// neqExCap is the additional capacity of neqPairs.
// These two arguments are useful to avoid unnecessary memory allocation
// when the caller wants to append custom data to eqPairs and neqPairs.
func testMkEqNeqPairs(eqGroups [][]interface{}, eqExCap, neqExCap int) (eqPairs, neqPairs [][2]interface{}) {
	if eqExCap < 0 {
		eqExCap = 0
	}
	if neqExCap < 0 {
		neqExCap = 0
	}
	gn := len(eqGroups)
	if gn == 0 {
		return
	}
	var eqPairsLen, neqPairsLen, eqIdx, neqIdx int
	for i, group := range eqGroups {
		n := len(group)
		eqPairsLen += n * (n + 1) / 2
		for k := i + 1; k < gn; k++ {
			neqPairsLen += n * len(eqGroups[k])
		}
	}
	eqPairs = make([][2]interface{}, eqPairsLen, eqPairsLen+eqExCap)
	neqPairs = make([][2]interface{}, neqPairsLen, neqPairsLen+neqExCap)
	for i, group := range eqGroups {
		for k := range group {
			for m := k; m < len(group); m++ {
				eqPairs[eqIdx][0], eqPairs[eqIdx][1], eqIdx = group[k], group[m], eqIdx+1
			}
		}
		for k := i + 1; k < gn; k++ {
			for _, a := range group {
				for _, b := range eqGroups[k] {
					neqPairs[neqIdx][0], neqPairs[neqIdx][1], neqIdx = a, b, neqIdx+1
				}
			}
		}
	}
	return
}

func BenchmarkIntsEqual(b *testing.B) {
	data := make([]int, 9999)
	for i := range data {
		data[i] = i % 100
	}

	bms := []struct {
		name string
		fn   EqualFunc
	}{
		{"One direction", IntsEqual},
		{"Two directions", testIntsEqualBidirectionalComparison},
	}
	for _, bm := range bms {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if !bm.fn(data, data) {
					b.Error(bm.name, "case reports false.")
				}
			}
		})
	}
}

func BenchmarkBytesEqual(b *testing.B) {
	const span = 'Z' - 'A' + 1
	bs := make([]byte, 9999)
	for i := range bs {
		bs[i] = 'A' + byte(i%span)
	}
	s := string(bs)

	bms := []struct {
		name string
		fn   EqualFunc
	}{
		{"One direction", BytesEqual},
		{"Two directions", testBytesEqualBidirectionalComparison},
	}
	for _, bm := range bms {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if !bm.fn(bs, s) {
					b.Error(bm.name, "case reports false.")
				}
			}
		})
	}
}

func BenchmarkSliceItemEqual(b *testing.B) {
	const span = 'Z' - 'A' + 1
	bs := make([]byte, 9999)
	for i := range bs {
		bs[i] = 'A' + byte(i%span)
	}
	s := string(bs)

	bms := []struct {
		name string
		fn   EqualFunc
	}{
		{"One direction", SliceItemEqual},
		{"Two directions", testSliceItemEqualBidirectionalComparison},
	}
	for _, bm := range bms {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if !bm.fn(bs, s) {
					b.Error(bm.name, "case reports false.")
				}
			}
		})
	}
}

var testIntsEqualBidirectionalComparison EqualFunc = func(a, b interface{}) bool {
	var ia, ib []int
	var ok bool
	if a != nil {
		ia, ok = a.([]int)
		if !ok {
			panic(errors.AutoMsg("a is not []int"))
		}
	}
	if b != nil {
		ib, ok = b.([]int)
		if !ok {
			panic(errors.AutoMsg("b is not []int"))
		}
	}
	if len(ia) != len(ib) {
		return false
	}
	for i, k := 0, len(ia)-1; i <= k; i, k = i+1, k-1 {
		// ia and ib will never be nil here.
		if ia[i] != ib[i] || ia[k] != ib[k] {
			return false
		}
	}
	return true
}

var testBytesEqualBidirectionalComparison EqualFunc = func(a, b interface{}) bool {
	if a == nil {
		a = ""
	}
	if b == nil {
		b = ""
	}
	if sa, ok := a.(string); ok {
		if sb, ok := b.(string); ok {
			return sa == sb
		}
		if bb, ok := b.([]byte); ok {
			return testBytesEqualBytesStringBidirectionalComparison(bb, sa)
		}
	} else if ba, ok := a.([]byte); ok {
		if sb, ok := b.(string); ok {
			return testBytesEqualBytesStringBidirectionalComparison(ba, sb)
		}
		if bb, ok := b.([]byte); ok {
			if len(ba) != len(bb) {
				return false
			}
			for i, k := 0, len(ba)-1; i <= k; i, k = i+1, k-1 {
				if ba[i] != bb[i] || ba[k] != bb[k] {
					return false
				}
			}
			return true
		}
	} else {
		panic(errors.AutoMsg("a is neither []byte nor string"))
	}
	panic(errors.AutoMsg("b is neither []byte nor string"))
}

var testSliceItemEqualBidirectionalComparison EqualFunc = func(a, b interface{}) bool {
	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
	var na, nb int
	switch va.Kind() {
	case reflect.Slice, reflect.String:
		na = va.Len()
	case reflect.Invalid:
	default:
		panic(errors.AutoMsg("a is neither slice nor string"))
	}
	switch vb.Kind() {
	case reflect.Slice, reflect.String:
		nb = vb.Len()
	case reflect.Invalid:
	default:
		panic(errors.AutoMsg("b is neither slice nor string"))
	}

	if na != nb {
		return false
	}
	for i, k := 0, na-1; i <= k; i, k = i+1, k-1 {
		if !(equal(va.Index(i).Interface(), vb.Index(i).Interface()) &&
			equal(va.Index(k).Interface(), vb.Index(k).Interface())) {
			return false
		}
	}
	return true
}

func testBytesEqualBytesStringBidirectionalComparison(b []byte, s string) bool {
	if len(b) != len(s) {
		return false
	}
	for i, k := 0, len(b)-1; i <= k; i, k = i+1, k-1 {
		if b[i] != s[i] || b[k] != s[k] {
			return false
		}
	}
	return true
}
