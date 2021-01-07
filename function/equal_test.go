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

import "testing"

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
	var eq EqualFunc = Equal
	nEq := eq.Not()
	for _, pair := range pairs {
		r1 := !eq(pair[0], pair[1])
		r2 := nEq(pair[0], pair[1])
		if r1 != r2 {
			t.Errorf("nEq(%v, %v) != !eq(%[1]v, %v).", pair[0], pair[1])
		}
	}
}

func TestGenerateEqualViaLess(t *testing.T) {
	intPairs := [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}}
	eq := GenerateEqualViaLess(IntLess)
	for _, pair := range intPairs {
		if r := eq(pair[0], pair[1]); r != (pair[0] == pair[1]) {
			t.Errorf("eq(%d, %d): %t.", pair[0], pair[1], r)
		}
	}
}

func TestIntsEqual(t *testing.T) {
	equalPairs := [][2][]int{
		{{}, {}},                           // Empty.
		{nil, nil},                         // Nil.
		{{1, 2, 3, 4}, {1, 2, 3, 4}},       // Even length.
		{{1, 2, 3, 4, 5}, {1, 2, 3, 4, 5}}, // Odd length.
	}
	unequalPairs := [][2][]int{
		{nil, {}},                                // Nil - empty.
		{{}, nil},                                // Empty - nil.
		{{1, 2, 3}, {1, 2, 3, 4}},                // Different length.
		{{1, 2, 3, 4, 5}, {1, 2, 5, 4, 5}},       // Odd length.
		{{1, 2, 3, 4, 5, 6}, {1, 2, 3, 3, 5, 6}}, // Even length.
	}
	for i, pair := range equalPairs {
		if r := IntsEqual(pair[0], pair[1]); !r {
			t.Errorf("equalPairs Case %d: IntsEqual(a, b): false. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
	}
	for i, pair := range unequalPairs {
		if r := IntsEqual(pair[0], pair[1]); r {
			t.Errorf("unequalPairs Case %d: IntsEqual(a, b): true. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
	}
}

func TestFloat64sEqual(t *testing.T) {
	equalPairs := [][2][]float64{
		{{}, {}},                           // Empty.
		{nil, nil},                         // Nil.
		{{1, 2, 3, 4}, {1, 2, 3, 4}},       // Even length.
		{{1, 2, 3, 4, 5}, {1, 2, 3, 4, 5}}, // Odd length.
	}
	unequalPairs := [][2][]float64{
		{nil, {}},                                // Nil - empty.
		{{}, nil},                                // Empty - nil.
		{{1, 2, 3}, {1, 2, 3, 4}},                // Different length.
		{{1, 2, 3, 4, 5}, {1, 2, 5, 4, 5}},       // Odd length.
		{{1, 2, 3, 4, 5, 6}, {1, 2, 3, 3, 5, 6}}, // Even length.
	}
	for i, pair := range equalPairs {
		if r := Float64sEqual(pair[0], pair[1]); !r {
			t.Errorf("equalPairs Case %d: Float64sEqual(a, b): false. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
	}
	for i, pair := range unequalPairs {
		if r := Float64sEqual(pair[0], pair[1]); r {
			t.Errorf("unequalPairs Case %d: Float64sEqual(a, b): true. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
	}
}

func TestStringsEqual(t *testing.T) {
	equalPairs := [][2][]string{
		{{}, {}},   // Empty.
		{nil, nil}, // Nil.
		{{"1", "2", "3", "4"}, {"1", "2", "3", "4"}},           // Even length.
		{{"1", "2", "3", "4", "5"}, {"1", "2", "3", "4", "5"}}, // Odd length.
	}
	unequalPairs := [][2][]string{
		{nil, {}},                               // Nil - empty.
		{{}, nil},                               // Empty - nil.
		{{"1", "2", "3"}, {"1", "2", "3", "4"}}, // Different length.
		{{"1", "2", "3", "4", "5"}, {"1", "2", "5", "4", "5"}},           // Odd length.
		{{"1", "2", "3", "4", "5", "6"}, {"1", "2", "3", "3", "5", "6"}}, // Even length.
	}
	for i, pair := range equalPairs {
		if r := StringsEqual(pair[0], pair[1]); !r {
			t.Errorf("equalPairs Case %d: StringsEqual(a, b): false. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
	}
	for i, pair := range unequalPairs {
		if r := StringsEqual(pair[0], pair[1]); r {
			t.Errorf("unequalPairs Case %d: StringsEqual(a, b): true. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
	}
}

func TestGeneralSliceEqual(t *testing.T) {
	equalPairs := [][2][]interface{}{
		{{}, {}},                           // Empty.
		{nil, nil},                         // Nil.
		{{1, 2, 3, 4}, {1, 2, 3, 4}},       // Even length.
		{{1, 2, 3, 4, 5}, {1, 2, 3, 4, 5}}, // Odd length.
	}
	unequalPairs := [][2][]interface{}{
		{nil, {}},                                // Nil - empty.
		{{}, nil},                                // Empty - nil.
		{{1, 2, 3}, {1, 2, 3, 4}},                // Different length.
		{{1, 2, 3, 4, 5}, {1, 2, 5, 4, 5}},       // Odd length.
		{{1, 2, 3, 4, 5, 6}, {1, 2, 3, 3, 5, 6}}, // Even length.
	}
	panicPair := [2][]interface{}{
		{[]int{1, 2, 3}},
		{[]int{1, 2, 3}},
	}
	for i, pair := range equalPairs {
		if r := GeneralSliceEqual(pair[0], pair[1]); !r {
			t.Errorf("equalPairs Case %d: GeneralSliceEqual(a, b): false. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
	}
	for i, pair := range unequalPairs {
		if r := GeneralSliceEqual(pair[0], pair[1]); r {
			t.Errorf("unequalPairs Case %d: GeneralSliceEqual(a, b): true. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("No panic when calling GeneralSliceEqual on %#v.", panicPair)
		}
	}()
	GeneralSliceEqual(panicPair[0], panicPair[1])
}

func TestSliceEqual(t *testing.T) {
	// Cases from other slice EqualFunc prefab tests.
	iEqualPairs := [][2][]int{
		{{}, {}},                           // Empty.
		{nil, nil},                         // Nil.
		{{1, 2, 3, 4}, {1, 2, 3, 4}},       // Even length.
		{{1, 2, 3, 4, 5}, {1, 2, 3, 4, 5}}, // Odd length.
	}
	iUnequalPairs := [][2][]int{
		{nil, {}},                                // Nil - empty.
		{{}, nil},                                // Empty - nil.
		{{1, 2, 3}, {1, 2, 3, 4}},                // Different length.
		{{1, 2, 3, 4, 5}, {1, 2, 5, 4, 5}},       // Odd length.
		{{1, 2, 3, 4, 5, 6}, {1, 2, 3, 3, 5, 6}}, // Even length.
	}
	fEqualPairs := [][2][]float64{
		{{}, {}},                           // Empty.
		{nil, nil},                         // Nil.
		{{1, 2, 3, 4}, {1, 2, 3, 4}},       // Even length.
		{{1, 2, 3, 4, 5}, {1, 2, 3, 4, 5}}, // Odd length.
	}
	fUnequalPairs := [][2][]float64{
		{nil, {}},                                // Nil - empty.
		{{}, nil},                                // Empty - nil.
		{{1, 2, 3}, {1, 2, 3, 4}},                // Different length.
		{{1, 2, 3, 4, 5}, {1, 2, 5, 4, 5}},       // Odd length.
		{{1, 2, 3, 4, 5, 6}, {1, 2, 3, 3, 5, 6}}, // Even length.
	}
	sEqualPairs := [][2][]string{
		{{}, {}},   // Empty.
		{nil, nil}, // Nil.
		{{"1", "2", "3", "4"}, {"1", "2", "3", "4"}},           // Even length.
		{{"1", "2", "3", "4", "5"}, {"1", "2", "3", "4", "5"}}, // Odd length.
	}
	sUnequalPairs := [][2][]string{
		{nil, {}},                               // Nil - empty.
		{{}, nil},                               // Empty - nil.
		{{"1", "2", "3"}, {"1", "2", "3", "4"}}, // Different length.
		{{"1", "2", "3", "4", "5"}, {"1", "2", "5", "4", "5"}},           // Odd length.
		{{"1", "2", "3", "4", "5", "6"}, {"1", "2", "3", "3", "5", "6"}}, // Even length.
	}
	gEqualPairs := [][2][]interface{}{
		{{}, {}},                           // Empty.
		{nil, nil},                         // Nil.
		{{1, 2, 3, 4}, {1, 2, 3, 4}},       // Even length.
		{{1, 2, 3, 4, 5}, {1, 2, 3, 4, 5}}, // Odd length.
	}
	gUnequalPairs := [][2][]interface{}{
		{nil, {}},                                // Nil - empty.
		{{}, nil},                                // Empty - nil.
		{{1, 2, 3}, {1, 2, 3, 4}},                // Different length.
		{{1, 2, 3, 4, 5}, {1, 2, 5, 4, 5}},       // Odd length.
		{{1, 2, 3, 4, 5, 6}, {1, 2, 3, 3, 5, 6}}, // Even length.
	}

	// Cases for SliceEqual only.
	equalPairs := [][2]interface{}{
		{nil, nil},                // Nil interface{}.
		{nil, []interface{}(nil)}, // Nil - nil interface{}.
		{[]interface{}(nil), nil}, // Nil interface{} - nil.
	}
	unequalPairs := [][2]interface{}{
		{nil, []interface{}{}},             // Nil - empty.
		{[]interface{}{}, nil},             // Empty - nil.
		{[]interface{}{}, []int{}},         // Different element types, empty.
		{[]interface{}{1, 2}, []int{1, 2}}, // Different element types, non-empty, same underlying values.
	}
	panicPair := [2]interface{}{
		[]interface{}{[]int{1, 2, 3}},
		[]interface{}{[]int{1, 2, 3}},
	}

	// Append cases from other slice EqualFunc prefab tests.
	for _, pair := range iEqualPairs {
		equalPairs = append(equalPairs, [2]interface{}{pair[0], pair[1]})
	}
	for _, pair := range iUnequalPairs {
		unequalPairs = append(unequalPairs, [2]interface{}{pair[0], pair[1]})
	}
	for _, pair := range fEqualPairs {
		equalPairs = append(equalPairs, [2]interface{}{pair[0], pair[1]})
	}
	for _, pair := range fUnequalPairs {
		unequalPairs = append(unequalPairs, [2]interface{}{pair[0], pair[1]})
	}
	for _, pair := range sEqualPairs {
		equalPairs = append(equalPairs, [2]interface{}{pair[0], pair[1]})
	}
	for _, pair := range sUnequalPairs {
		unequalPairs = append(unequalPairs, [2]interface{}{pair[0], pair[1]})
	}
	for _, pair := range gEqualPairs {
		equalPairs = append(equalPairs, [2]interface{}{pair[0], pair[1]})
	}
	for _, pair := range gUnequalPairs {
		unequalPairs = append(unequalPairs, [2]interface{}{pair[0], pair[1]})
	}

	for i, pair := range equalPairs {
		if r := SliceEqual(pair[0], pair[1]); !r {
			t.Errorf("equalPairs Case %d: SliceEqual(a, b): false. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
	}
	for i, pair := range unequalPairs {
		if r := SliceEqual(pair[0], pair[1]); r {
			t.Errorf("unequalPairs Case %d: SliceEqual(a, b): true. a: %#v, b: %#v.", i, pair[0], pair[1])
		}
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("No panic when calling SliceEqual on %#v.", panicPair)
		}
	}()
	SliceEqual(panicPair[0], panicPair[1])
}
