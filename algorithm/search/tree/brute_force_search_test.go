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

package tree

import "testing"

type testTree struct {
	Data         [][2]int
	Target       interface{}
	VisitHistory []interface{}
}

func (tt *testTree) Root() interface{} {
	if len(tt.Data) == 0 {
		return nil
	}
	return 0
}

func (tt *testTree) FirstChild(node interface{}) interface{} {
	idx := tt.Data[node.(int)][0]
	if idx < 0 {
		return nil
	}
	return idx
}

func (tt *testTree) NextSibling(node interface{}) interface{} {
	idx := tt.Data[node.(int)][1]
	if idx < 0 {
		return nil
	}
	return idx
}

// SetTarget sets the search target as well as resets the visit history.
func (tt *testTree) SetTarget(target interface{}) {
	tt.Target = target
	tt.VisitHistory = tt.VisitHistory[:0] // Reuse the underlying array.
}

func (tt *testTree) Visit(node interface{}) (found bool) {
	tt.VisitHistory = append(tt.VisitHistory, node)
	if tt.Target == nil {
		return false
	}
	return node == tt.Target
}

// testTreeData represents a tree as follows:
//         0
//       / | \
//      1  2  3
//     /|     |\
//    4 5     6 7
//   /|       |\
//  8 9      10 11
//
// Expected DFS order:
//  0, 1, 4, 8, 9, 5, 2, 3, 6, 10, 11, 7
// Expected BFS order:
//  0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11
var testTreeData = [][2]int{
	{1, -1},  // node 0
	{4, 2},   // node 1
	{-1, 3},  // node 2
	{6, -1},  // node 3
	{8, 5},   // node 4
	{-1, -1}, // node 5
	{10, 7},  // node 6
	{-1, -1}, // node 7
	{-1, 9},  // node 8
	{-1, -1}, // node 9
	{-1, 11}, // node 10
	{-1, -1}, // node 11
}

func TestDfs(t *testing.T) {
	testBruteForceSearch(t, "Dfs", Dfs, []int{0, 1, 4, 8, 9, 5, 2, 3, 6, 10, 11, 7})
}

func TestBfs(t *testing.T) {
	testBruteForceSearch(t, "Bfs", Bfs, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
}

func testBruteForceSearch(t *testing.T, name string, f func(itf Interface, target interface{}) interface{}, order []int) {
	tt := &testTree{Data: testTreeData}
	for i, node := range order {
		r := f(tt, node)
		if r != node {
			t.Errorf("%s returns %v != %v.", name, r, node)
		}
		testVisitHistoryCheck(t, tt, order[:1+i])
	}
	// Traverse all nodes:
	r := f(tt, nil)
	if r != nil {
		t.Errorf("%s returns %v != nil.", name, r)
	}
	testVisitHistoryCheck(t, tt, order)
}

func testVisitHistoryCheck(t *testing.T, tt *testTree, wanted []int) {
	if len(tt.VisitHistory) != len(wanted) {
		t.Errorf("Visit history: %v\nwanted: %v", tt.VisitHistory, wanted)
		return
	}
	for i := range tt.VisitHistory {
		if tt.VisitHistory[i] != wanted[i] {
			t.Errorf("Visit history: %v\nwanted: %v", tt.VisitHistory, wanted)
			return
		}
	}
}
