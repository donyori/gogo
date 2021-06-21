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
	Data          [][2]int
	Goal          interface{}
	AccessHistory []interface{}
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

// SetGoal sets the search goal as well as resets the access history.
func (tt *testTree) SetGoal(goal interface{}) {
	tt.Goal = goal
	tt.AccessHistory = tt.AccessHistory[:0] // Reuse the underlying array.
}

func (tt *testTree) Access(node interface{}) (found bool) {
	tt.AccessHistory = append(tt.AccessHistory, node)
	if tt.Goal == nil {
		return false
	}
	return node == tt.Goal
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

// testTreeDataNodePath is a list of paths from the root
// to each node of testTreeData.
var testTreeDataNodePath = [][]int{
	{0},
	{0, 1},
	{0, 2},
	{0, 3},
	{0, 1, 4},
	{0, 1, 5},
	{0, 3, 6},
	{0, 3, 7},
	{0, 1, 4, 8},
	{0, 1, 4, 9},
	{0, 3, 6, 10},
	{0, 3, 6, 11},
}

func TestDfs(t *testing.T) {
	testBruteForceSearch(t, "Dfs", Dfs, []int{0, 1, 4, 8, 9, 5, 2, 3, 6, 10, 11, 7})
}

func TestBfs(t *testing.T) {
	testBruteForceSearch(t, "Bfs", Bfs, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
}

func TestDfsPath(t *testing.T) {
	testBruteForceSearchPath(t, "DfsPath", DfsPath, []int{0, 1, 4, 8, 9, 5, 2, 3, 6, 10, 11, 7})
}

func TestBfsPath(t *testing.T) {
	testBruteForceSearchPath(t, "BfsPath", BfsPath, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
}

func testBruteForceSearch(t *testing.T, name string, f func(itf Interface, goal interface{}) interface{}, order []int) {
	tt := &testTree{Data: testTreeData}
	for i, node := range order {
		r := f(tt, node)
		if r != node {
			t.Errorf("%s returns %v != %v.", name, r, node)
		}
		testCheckAccessHistory(t, name, tt, order[:1+i])
	}
	// Non-existent nodes:
	for _, goal := range []interface{}{nil, -1, 1.2} {
		r := f(tt, goal)
		if r != nil {
			t.Errorf("%s returns %v != nil.", name, r)
		}
		testCheckAccessHistory(t, name, tt, order)
	}
}

func testBruteForceSearchPath(t *testing.T, name string, f func(itf Interface, goal interface{}) []interface{}, order []int) {
	tt := &testTree{Data: testTreeData}
	for i, node := range order {
		list := f(tt, node)
		testCheckNodePath(t, name, node, list)
		testCheckAccessHistory(t, name, tt, order[:1+i])
	}
	// Non-existent nodes:
	for _, goal := range []interface{}{nil, -1, 1.2} {
		list := f(tt, goal)
		testCheckNodePath(t, name, goal, list)
		testCheckAccessHistory(t, name, tt, order)
	}
}

func testCheckAccessHistory(t *testing.T, name string, tt *testTree, wanted []int) {
	if len(tt.AccessHistory) != len(wanted) {
		t.Errorf("%s - Access history: %v\nwanted: %v", name, tt.AccessHistory, wanted)
		return
	}
	for i := range wanted {
		if tt.AccessHistory[i] != wanted[i] {
			t.Errorf("%s - Access history: %v\nwanted: %v", name, tt.AccessHistory, wanted)
			return
		}
	}
}

func testCheckNodePath(t *testing.T, name string, node interface{}, nodePathList []interface{}) {
	var p []int
	idx, ok := node.(int)
	if ok && idx >= 0 && idx < len(testTreeDataNodePath) {
		p = testTreeDataNodePath[idx]
	}
	if len(p) != len(nodePathList) {
		t.Errorf("%s - NodePath of %v: %v\nwanted: %v", name, node, nodePathList, p)
		return
	}
	for i := range p {
		if nodePathList[i] != p[i] {
			t.Errorf("%s - NodePath of %v: %v\nwanted: %v", name, node, nodePathList, p)
			return
		}
	}
}
