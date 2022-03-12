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

package tree

import (
	"fmt"
	"testing"
)

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

// testNumTreeNodes is the number of nodes in testTreeData.
const testNumTreeNodes int = 12

// testTreeData represents a tree as follows:
//        0
//       /|\
//      1 2 3
//     /|   |\
//    4 5   6 7
//   /|     |\
//  8 9    10 11
//
// Assuming that the left edges are chosen before the right edges,
// the expected orderings are as follows:
//
// Expected DFS ordering:
//  0, 1, 4, 8, 9, 5, 2, 3, 6, 10, 11, 7
// Expected BFS ordering:
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

// testTreeDataOrderingMap is a mapping from algorithm short names
// to the expected node access orderings of testTreeData.
//
// Valid keys:
//  dfs
//  bfs
//  dls-0
//  dls-1
//  dls-2
//  dls-3
//  dls-m1
//  ids
// where dls is followed by the depth limit, and m1 is minus 1 (-1).
var testTreeDataOrderingMap = map[string][]int{
	"dfs":    {0, 1, 4, 8, 9, 5, 2, 3, 6, 10, 11, 7},
	"bfs":    {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	"dls-0":  {0},
	"dls-1":  {0, 1, 2, 3},
	"dls-2":  {0, 1, 4, 5, 2, 3, 6, 7},
	"dls-3":  nil, // It is the same as dfs and will be set in function init.
	"dls-m1": {},
	"ids":    {0, 1, 2, 3, 0, 1, 4, 5, 2, 3, 6, 7, 0, 1, 4, 8, 9, 5, 2, 3, 6, 10, 11, 7},
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

func init() {
	testTreeDataOrderingMap["dls-3"] = testTreeDataOrderingMap["dfs"]
}

func TestDfs(t *testing.T) {
	testBruteForceSearch(t, "Dfs")
}

func TestDfsPath(t *testing.T) {
	testBruteForceSearchPath(t, "DfsPath")
}

func TestBfs(t *testing.T) {
	testBruteForceSearch(t, "Bfs")
}

func TestBfsPath(t *testing.T) {
	testBruteForceSearchPath(t, "BfsPath")
}

func TestDls(t *testing.T) {
	for _, limit := range []int{0, 1, 2, 3, -1} {
		var name string
		if limit >= 0 {
			name = fmt.Sprintf("Dls-%d", limit)
		} else {
			name = fmt.Sprintf("Dls-m%d", -limit)
		}
		t.Run(fmt.Sprintf("limit=%d", limit), func(t *testing.T) {
			testBruteForceSearch(t, name)
		})
	}
}

func TestDlsPath(t *testing.T) {
	for _, limit := range []int{0, 1, 2, 3, -1} {
		var name string
		if limit >= 0 {
			name = fmt.Sprintf("DlsPath-%d", limit)
		} else {
			name = fmt.Sprintf("DlsPath-m%d", -limit)
		}
		t.Run(fmt.Sprintf("limit=%d", limit), func(t *testing.T) {
			testBruteForceSearchPath(t, name)
		})
	}
}

func TestIds(t *testing.T) {
	t.Run("initLimit=1", func(t *testing.T) {
		testBruteForceSearch(t, "Ids")
	})
	t.Run("initLimit=-1", func(t *testing.T) {
		testBruteForceSearch(t, "Ids-m")
	})
}

func TestIdsPath(t *testing.T) {
	t.Run("initLimit=1", func(t *testing.T) {
		testBruteForceSearchPath(t, "IdsPath")
	})
	t.Run("initLimit=-1", func(t *testing.T) {
		testBruteForceSearchPath(t, "IdsPath-m")
	})
}

func testBruteForceSearch(t *testing.T, name string) {
	var f func(itf Interface, goal interface{}) interface{}
	var ordering []int
	switch name {
	case "Dfs":
		f = Dfs
		ordering = testTreeDataOrderingMap["dfs"]
	case "Bfs":
		f = Bfs
		ordering = testTreeDataOrderingMap["bfs"]
	case "Dls-0":
		f = func(itf Interface, goal interface{}) interface{} {
			node, more := Dls(itf, goal, 0)
			if node == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return node
		}
		ordering = testTreeDataOrderingMap["dls-0"]
	case "Dls-1":
		f = func(itf Interface, goal interface{}) interface{} {
			node, more := Dls(itf, goal, 1)
			if node == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return node
		}
		ordering = testTreeDataOrderingMap["dls-1"]
	case "Dls-2":
		f = func(itf Interface, goal interface{}) interface{} {
			node, more := Dls(itf, goal, 2)
			if node == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return node
		}
		ordering = testTreeDataOrderingMap["dls-2"]
	case "Dls-3":
		f = func(itf Interface, goal interface{}) interface{} {
			node, _ := Dls(itf, goal, 3)
			// Both true and false are acceptable for the second return value.
			return node
		}
		ordering = testTreeDataOrderingMap["dls-3"]
	case "Dls-m1":
		f = func(itf Interface, goal interface{}) interface{} {
			node, more := Dls(itf, goal, -1)
			if node == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return node
		}
		ordering = testTreeDataOrderingMap["dls-m1"]
	case "Ids":
		f = func(itf Interface, goal interface{}) interface{} {
			return Ids(itf, goal, 1)
		}
		ordering = testTreeDataOrderingMap["ids"]
	case "Ids-m":
		f = func(itf Interface, goal interface{}) interface{} {
			return Ids(itf, goal, -1)
		}
		ordering = testTreeDataOrderingMap["ids"]
	default:
		t.Errorf("unacceptable name %q", name)
		return
	}

	tt := &testTree{Data: testTreeData}
	for node := 0; node < testNumTreeNodes; node++ {
		var i int
		for i < len(ordering) && ordering[i] != node {
			i++
		}
		var wantNode interface{} // The node expected to be found.
		wantHx := ordering       // Expected history.
		if i < len(ordering) {
			wantNode = node
			wantHx = wantHx[:1+i]
		}
		r := f(tt, node)
		if r != wantNode {
			t.Errorf("got %v; want %v", r, wantNode)
		}
		testCheckAccessHistory(t, tt, wantHx)
	}
	// Non-existent nodes:
	for _, goal := range []interface{}{nil, -1, len(testTreeData), 1.2} {
		r := f(tt, goal)
		if r != nil {
			t.Errorf("got %v; want <nil>", r)
		}
		testCheckAccessHistory(t, tt, ordering)
	}
}

func testBruteForceSearchPath(t *testing.T, name string) {
	var f func(itf Interface, goal interface{}) []interface{}
	var ordering []int
	switch name {
	case "DfsPath":
		f = DfsPath
		ordering = testTreeDataOrderingMap["dfs"]
	case "BfsPath":
		f = BfsPath
		ordering = testTreeDataOrderingMap["bfs"]
	case "DlsPath-0":
		f = func(itf Interface, goal interface{}) []interface{} {
			p, more := DlsPath(itf, goal, 0)
			if p == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return p
		}
		ordering = testTreeDataOrderingMap["dls-0"]
	case "DlsPath-1":
		f = func(itf Interface, goal interface{}) []interface{} {
			p, more := DlsPath(itf, goal, 1)
			if p == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return p
		}
		ordering = testTreeDataOrderingMap["dls-1"]
	case "DlsPath-2":
		f = func(itf Interface, goal interface{}) []interface{} {
			p, more := DlsPath(itf, goal, 2)
			if p == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return p
		}
		ordering = testTreeDataOrderingMap["dls-2"]
	case "DlsPath-3":
		f = func(itf Interface, goal interface{}) []interface{} {
			p, _ := DlsPath(itf, goal, 3)
			// Both true and false are acceptable for the second return value.
			return p
		}
		ordering = testTreeDataOrderingMap["dls-3"]
	case "DlsPath-m1":
		f = func(itf Interface, goal interface{}) []interface{} {
			p, more := DlsPath(itf, goal, -1)
			if p == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return p
		}
		ordering = testTreeDataOrderingMap["dls-m1"]
	case "IdsPath":
		f = func(itf Interface, goal interface{}) []interface{} {
			return IdsPath(itf, goal, 1)
		}
		ordering = testTreeDataOrderingMap["ids"]
	case "IdsPath-m":
		f = func(itf Interface, goal interface{}) []interface{} {
			return IdsPath(itf, goal, -1)
		}
		ordering = testTreeDataOrderingMap["ids"]
	default:
		t.Errorf("unacceptable name %q", name)
		return
	}

	tt := &testTree{Data: testTreeData}
	for node := 0; node < testNumTreeNodes; node++ {
		pl := f(tt, node)
		testCheckPath(t, name, node, pl)
		var i int
		for i < len(ordering) && ordering[i] != node {
			i++
		}
		wantHx := ordering // Expected history.
		if i < len(ordering) {
			wantHx = wantHx[:1+i]
		}
		testCheckAccessHistory(t, tt, wantHx)
	}
	// Non-existent nodes:
	for _, goal := range []interface{}{nil, -1, len(testTreeData), 1.2} {
		pl := f(tt, goal)
		testCheckPath(t, name, goal, pl)
		testCheckAccessHistory(t, tt, ordering)
	}
}

func testCheckAccessHistory(t *testing.T, tt *testTree, want []int) {
	if len(tt.AccessHistory) != len(want) {
		t.Errorf("got access history %v;\nwant %v", tt.AccessHistory, want)
		return
	}
	for i := range want {
		if tt.AccessHistory[i] != want[i] {
			t.Errorf("got access history %v;\nwant %v", tt.AccessHistory, want)
			return
		}
	}
}

func testCheckPath(t *testing.T, name string, node interface{}, pathList []interface{}) {
	var wantPath []int
	var ordering []int
	switch name {
	case "DfsPath":
		ordering = testTreeDataOrderingMap["dfs"]
	case "BfsPath":
		ordering = testTreeDataOrderingMap["bfs"]
	case "DlsPath-0":
		ordering = testTreeDataOrderingMap["dls-0"]
	case "DlsPath-1":
		ordering = testTreeDataOrderingMap["dls-1"]
	case "DlsPath-2":
		ordering = testTreeDataOrderingMap["dls-2"]
	case "DlsPath-3":
		ordering = testTreeDataOrderingMap["dls-3"]
	case "DlsPath-m1":
		ordering = testTreeDataOrderingMap["dls-m1"]
	case "IdsPath", "IdsPath-m":
		ordering = testTreeDataOrderingMap["ids"]
	default:
		// This should never happen, but will act as a safeguard for later,
		// as a default value doesn't make sense here.
		t.Errorf("unacceptable name %q", name)
		return
	}
	for _, n := range ordering {
		if n == node {
			idx, ok := node.(int)
			if ok && idx >= 0 && idx < len(testTreeDataNodePath) {
				wantPath = testTreeDataNodePath[idx]
			}
			break
		}
	}
	if len(pathList) != len(wantPath) {
		t.Errorf("path of %v %v;\nwant %v", node, pathList, wantPath)
		return
	}
	for i := range wantPath {
		if pathList[i] != wantPath[i] {
			t.Errorf("path of %v %v;\nwant %v", node, pathList, wantPath)
			return
		}
	}
}
