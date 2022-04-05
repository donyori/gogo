// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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

package graphv

import (
	"fmt"
	"testing"
)

// testGraph implements all methods of interface IdsInterface[int].
type testGraph struct {
	Data          [][]int
	Goal          int
	GoalValid     bool
	AccessHistory []int

	// Index of the beginning of the access history
	// in the most recent search iteration.
	head int
}

// Init sets the search goal as well as resets the access history.
func (tg *testGraph) Init(args ...any) {
	tg.Goal, tg.GoalValid = -1, false
	if len(args) >= 1 {
		goal, ok := args[0].(int)
		if ok {
			tg.Goal, tg.GoalValid = goal, true
		}
	}
	tg.AccessHistory = tg.AccessHistory[:0] // Reuse the underlying array.
	tg.head = 0
}

func (tg *testGraph) Root() int {
	if len(tg.Data) == 0 {
		return -1
	}
	return 0
}

func (tg *testGraph) Adjacency(vertex int) []int {
	list := tg.Data[vertex]
	if len(list) == 0 {
		return nil
	}
	return append(list[:0:0], list...) // Return a copy of list.
}

func (tg *testGraph) Access(vertex, _ int) (found, cont bool) {
	tg.AccessHistory = append(tg.AccessHistory, vertex)
	if !tg.GoalValid {
		return
	}
	return vertex == tg.Goal, true
}

// Discovered reports whether the specified vertex
// has been examined by the method Access
// via checking the access history.
//
// It may be time-consuming,
// but it doesn't matter because the test graph is very small.
func (tg *testGraph) Discovered(vertex int) bool {
	for i := tg.head; i < len(tg.AccessHistory); i++ {
		if tg.AccessHistory[i] == vertex {
			return true
		}
	}
	return false
}

func (tg *testGraph) ResetSearchState() {
	tg.head = len(tg.AccessHistory)
}

// testNumUndirectedGraphVertices is the number of vertices
// in testUndirectedGraphData.
const testNumUndirectedGraphVertices int = 7

// testUndirectedGraphData represents an undirected graph as follows:
//      0
//     /|\
//    1 2 3
//   /| | |
//  4 5 6 |
//     \_/
//
// Assuming that the search starts at the vertex 0,
// and the left edges are chosen before the right edges,
// the expected orderings are as follows:
//
// Expected DFS ordering:
//  0, 1, 4, 5, 3, 2, 6
// Expected BFS ordering:
//  0, 1, 2, 3, 4, 5, 6
var testUndirectedGraphData = [][]int{
	{1, 2, 3}, // vertex 0
	{0, 4, 5}, // vertex 1
	{0, 6},    // vertex 2
	{0, 5},    // vertex 3
	{1},       // vertex 4
	{1, 3},    // vertex 5
	{2},       // vertex 6
}

// testUndirectedGraphDataOrderingMap is a mapping from algorithm short names
// to the expected vertex access orderings of testUndirectedGraphData.
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
var testUndirectedGraphDataOrderingMap = map[string][]int{
	"dfs":    {0, 1, 4, 5, 3, 2, 6},
	"bfs":    {0, 1, 2, 3, 4, 5, 6},
	"dls-0":  {0},
	"dls-1":  {0, 1, 2, 3},
	"dls-2":  {0, 1, 4, 5, 2, 6, 3},
	"dls-3":  nil, // It is the same as dfs and will be set in function init.
	"dls-m1": {},
	"ids":    {0, 1, 2, 3, 0, 1, 4, 5, 2, 6, 3, 0, 1, 4, 5, 3, 2, 6},
}

// testUndirectedGraphDataVertexPathMap is a mapping from algorithm short names
// to lists of paths from the root to each vertex of testUndirectedGraphData.
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
var testUndirectedGraphDataVertexPathMap = map[string][][]int{
	"dfs": {
		{0},
		{0, 1},
		{0, 2},
		{0, 1, 5, 3},
		{0, 1, 4},
		{0, 1, 5},
		{0, 2, 6},
	},
	"bfs": {
		{0},
		{0, 1},
		{0, 2},
		{0, 3},
		{0, 1, 4},
		{0, 1, 5},
		{0, 2, 6},
	},
	"dls-0": {
		{0},
	},
	"dls-1": {
		{0},
		{0, 1},
		{0, 2},
		{0, 3},
	},
	"dls-2":  nil, // It is the same as bfs and will be set in function init.
	"dls-3":  nil, // It is the same as dfs and will be set in function init.
	"dls-m1": {},
	"ids":    nil, // It is the same as bfs and will be set in function init.
}

func init() {
	testUndirectedGraphDataOrderingMap["dls-3"] = testUndirectedGraphDataOrderingMap["dfs"]

	testUndirectedGraphDataVertexPathMap["dls-2"] = testUndirectedGraphDataVertexPathMap["bfs"]
	testUndirectedGraphDataVertexPathMap["dls-3"] = testUndirectedGraphDataVertexPathMap["dfs"]
	testUndirectedGraphDataVertexPathMap["ids"] = testUndirectedGraphDataVertexPathMap["bfs"]
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
	var f func(t *testing.T, itf IdsInterface[int], initArgs ...any) (int, bool)
	var ordering []int
	switch name {
	case "Dfs":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) (int, bool) {
			return Dfs[int](itf, initArgs...)
		}
		ordering = testUndirectedGraphDataOrderingMap["dfs"]
	case "Bfs":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) (int, bool) {
			return Bfs[int](itf, initArgs...)
		}
		ordering = testUndirectedGraphDataOrderingMap["bfs"]
	case "Dls-0":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) (int, bool) {
			vertex, found, more := Dls[int](itf, 0, initArgs...)
			if testIsFirstInitArgInt(initArgs) && !found && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return vertex, found
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-0"]
	case "Dls-1":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) (int, bool) {
			vertex, found, more := Dls[int](itf, 1, initArgs...)
			if testIsFirstInitArgInt(initArgs) && !found && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return vertex, found
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-1"]
	case "Dls-2":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) (int, bool) {
			vertex, found, _ := Dls[int](itf, 2, initArgs...)
			// Both true and false are acceptable for the third return value.
			return vertex, found
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-2"]
	case "Dls-3":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) (int, bool) {
			vertex, found, _ := Dls[int](itf, 3, initArgs...)
			// Both true and false are acceptable for the third return value.
			return vertex, found
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-3"]
	case "Dls-m1":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) (int, bool) {
			vertex, found, more := Dls[int](itf, -1, initArgs...)
			if testIsFirstInitArgInt(initArgs) && !found && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return vertex, found
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-m1"]
	case "Ids":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) (int, bool) {
			return Ids(itf, 1, initArgs...)
		}
		ordering = testUndirectedGraphDataOrderingMap["ids"]
	case "Ids-m":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) (int, bool) {
			return Ids(itf, -1, initArgs...)
		}
		ordering = testUndirectedGraphDataOrderingMap["ids"]
	default:
		t.Errorf("unacceptable name %q", name)
		return
	}

	tg := &testGraph{Data: testUndirectedGraphData}
	for goal := 0; goal < testNumUndirectedGraphVertices; goal++ {
		t.Run(fmt.Sprintf("goal=%d", goal), func(t *testing.T) {
			var i int
			for i < len(ordering) && ordering[i] != goal {
				i++
			}
			var wantVertex int // The vertex expected to be found.
			var wantFound bool // Expected found value.
			wantHx := ordering // Expected history.
			if i < len(ordering) {
				wantVertex, wantFound, wantHx = goal, true, wantHx[:1+i]
			}
			r, found := f(t, tg, goal)
			if found != wantFound || r != wantVertex {
				t.Errorf("got <%d, %t>; want <%d, %t>", r, found, wantVertex, wantFound)
			}
			testCheckAccessHistory(t, tg, wantHx)
		})
	}
	// Non-existent nodes:
	for _, goal := range []any{nil, -1, len(testUndirectedGraphData), 1.2} {
		t.Run(fmt.Sprintf("goal=%v", goal), func(t *testing.T) {
			wantHx := ordering // Expected history.
			if len(wantHx) > 1 {
				if _, ok := goal.(int); !ok {
					wantHx = wantHx[:1]
				}
			}
			r, found := f(t, tg, goal)
			if r != 0 || found {
				t.Errorf("got <%d, %t>; want <0, false>", r, found)
			}
			testCheckAccessHistory(t, tg, wantHx)
		})
	}
}

func testBruteForceSearchPath(t *testing.T, name string) {
	var f func(t *testing.T, itf IdsInterface[int], initArgs ...any) []int
	var ordering []int
	switch name {
	case "DfsPath":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) []int {
			return DfsPath[int](itf, initArgs...)
		}
		ordering = testUndirectedGraphDataOrderingMap["dfs"]
	case "BfsPath":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) []int {
			return BfsPath[int](itf, initArgs...)
		}
		ordering = testUndirectedGraphDataOrderingMap["bfs"]
	case "DlsPath-0":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) []int {
			path, more := DlsPath[int](itf, 0, initArgs...)
			if testIsFirstInitArgInt(initArgs) && path == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return path
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-0"]
	case "DlsPath-1":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) []int {
			path, more := DlsPath[int](itf, 1, initArgs...)
			if testIsFirstInitArgInt(initArgs) && path == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return path
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-1"]
	case "DlsPath-2":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) []int {
			path, _ := DlsPath[int](itf, 2, initArgs...)
			// Both true and false are acceptable for the second return value.
			return path
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-2"]
	case "DlsPath-3":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) []int {
			path, _ := DlsPath[int](itf, 3, initArgs...)
			// Both true and false are acceptable for the second return value.
			return path
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-3"]
	case "DlsPath-m1":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) []int {
			path, more := DlsPath[int](itf, -1, initArgs...)
			if testIsFirstInitArgInt(initArgs) && path == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return path
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-m1"]
	case "IdsPath":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) []int {
			return IdsPath(itf, 1, initArgs...)
		}
		ordering = testUndirectedGraphDataOrderingMap["ids"]
	case "IdsPath-m":
		f = func(t *testing.T, itf IdsInterface[int], initArgs ...any) []int {
			return IdsPath(itf, -1, initArgs...)
		}
		ordering = testUndirectedGraphDataOrderingMap["ids"]
	default:
		t.Errorf("unacceptable name %q", name)
		return
	}

	tg := &testGraph{Data: testUndirectedGraphData}
	for goal := 0; goal < testNumUndirectedGraphVertices; goal++ {
		t.Run(fmt.Sprintf("goal=%d", goal), func(t *testing.T) {
			path := f(t, tg, goal)
			testCheckPath(t, name, goal, path)
			var i int
			for i < len(ordering) && ordering[i] != goal {
				i++
			}
			wantHx := ordering // Expected history.
			if i < len(ordering) {
				wantHx = wantHx[:1+i]
			}
			testCheckAccessHistory(t, tg, wantHx)
		})
	}
	// Non-existent nodes:
	for _, goal := range []any{nil, -1, len(testUndirectedGraphData), 1.2} {
		t.Run(fmt.Sprintf("goal=%v", goal), func(t *testing.T) {
			wantHx := ordering // Expected history.
			goalVertex := -1
			if g, ok := goal.(int); ok {
				goalVertex = g
			} else if len(wantHx) > 1 {
				wantHx = wantHx[:1]
			}
			path := f(t, tg, goal)
			testCheckPath(t, name, goalVertex, path)
			testCheckAccessHistory(t, tg, wantHx)
		})
	}
}

func testIsFirstInitArgInt(initArgs []any) bool {
	if len(initArgs) < 1 {
		return false
	}
	_, ok := initArgs[0].(int)
	return ok
}

func testCheckAccessHistory(t *testing.T, tg *testGraph, want []int) {
	if len(tg.AccessHistory) != len(want) {
		t.Errorf("got access history %v;\nwant %v", tg.AccessHistory, want)
		return
	}
	for i := range want {
		if tg.AccessHistory[i] != want[i] {
			t.Errorf("got access history %v;\nwant %v", tg.AccessHistory, want)
			return
		}
	}
}

func testCheckPath(t *testing.T, name string, vertex int, pathList []int) {
	var wantPath []int
	var list [][]int
	switch name {
	case "DfsPath":
		list = testUndirectedGraphDataVertexPathMap["dfs"]
	case "BfsPath":
		list = testUndirectedGraphDataVertexPathMap["bfs"]
	case "DlsPath-0":
		list = testUndirectedGraphDataVertexPathMap["dls-0"]
	case "DlsPath-1":
		list = testUndirectedGraphDataVertexPathMap["dls-1"]
	case "DlsPath-2":
		list = testUndirectedGraphDataVertexPathMap["dls-2"]
	case "DlsPath-3":
		list = testUndirectedGraphDataVertexPathMap["dls-3"]
	case "DlsPath-m1":
		list = testUndirectedGraphDataVertexPathMap["dls-m1"]
	case "IdsPath", "IdsPath-m":
		list = testUndirectedGraphDataVertexPathMap["ids"]
	default:
		// This should never happen, but will act as a safeguard for later,
		// as a default value doesn't make sense here.
		t.Errorf("unacceptable name %q", name)
		return
	}
	if vertex >= 0 && vertex < len(list) {
		wantPath = list[vertex]
	}
	if len(pathList) != len(wantPath) {
		t.Errorf("path of %d %v;\nwant %v", vertex, pathList, wantPath)
		return
	}
	for i := range wantPath {
		if pathList[i] != wantPath[i] {
			t.Errorf("path of %d %v;\nwant %v", vertex, pathList, wantPath)
			return
		}
	}
}
