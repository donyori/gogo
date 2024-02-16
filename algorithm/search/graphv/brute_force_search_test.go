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

package graphv_test

import (
	"fmt"
	"testing"

	"github.com/donyori/gogo/algorithm/search/graphv"
)

// graphImpl implements all methods of interface IDSAccessVertex[int].
type graphImpl struct {
	Data          [][]int
	Goal          int
	GoalValid     bool
	AccessHistory []int

	// Index of the beginning of the access history
	// in the most recent search iteration.
	head int
}

// Init sets the search goal as well as resets the access history.
func (g *graphImpl) Init(args ...any) {
	g.Goal, g.GoalValid = -1, false
	if len(args) >= 1 {
		goal, ok := args[0].(int)
		if ok {
			g.Goal, g.GoalValid = goal, true
		}
	}
	g.AccessHistory = g.AccessHistory[:0] // reuse the underlying array
	g.head = 0
}

func (g *graphImpl) Root() int {
	if len(g.Data) == 0 {
		return -1
	}
	return 0
}

func (g *graphImpl) Adjacency(vertex int) []int {
	list := g.Data[vertex]
	if len(list) == 0 {
		return nil
	}
	return append(list[:0:0], list...) // return a copy of list
}

// Discovered reports whether the specified vertex
// has been examined by the method AccessVertex or AccessPath
// via checking the access history.
//
// It may be time-consuming,
// but it doesn't matter because the graph is very small.
func (g *graphImpl) Discovered(vertex int) bool {
	for i := g.head; i < len(g.AccessHistory); i++ {
		if g.AccessHistory[i] == vertex {
			return true
		}
	}
	return false
}

func (g *graphImpl) AccessVertex(vertex, _ int) (found, cont bool) {
	g.AccessHistory = append(g.AccessHistory, vertex)
	if !g.GoalValid {
		return
	}
	return vertex == g.Goal, true
}

func (g *graphImpl) AccessPath(path []int) (found, cont bool) {
	return g.AccessVertex(path[len(path)-1], len(path)-1)
}

func (g *graphImpl) ResetSearchState() {
	g.head = len(g.AccessHistory)
}

// NumUndirectedGraphVertex is the number of vertices
// in undirectedGraphData.
const NumUndirectedGraphVertex int = 7

// undirectedGraphData represents an undirected graph as follows:
//
//	    0
//	   /|\
//	  1 2 3
//	 /| | |
//	4 5 6 |
//	   \_/
//
// Assuming that the search starts at the vertex 0,
// and the left edges are chosen before the right edges,
// the expected orderings are as follows:
//
// Expected DFS ordering:
//
//	0, 1, 4, 5, 3, 2, 6
//
// Expected BFS ordering:
//
//	0, 1, 2, 3, 4, 5, 6
var undirectedGraphData = [][]int{
	{1, 2, 3}, // vertex 0
	{0, 4, 5}, // vertex 1
	{0, 6},    // vertex 2
	{0, 5},    // vertex 3
	{1},       // vertex 4
	{1, 3},    // vertex 5
	{2},       // vertex 6
}

// undirectedGraphDataOrderingMap is a mapping from algorithm short names
// to the expected vertex access orderings of undirectedGraphData.
//
// Valid keys:
//
//	dfs
//	bfs
//	dls-0
//	dls-1
//	dls-2
//	dls-3
//	dls-m1
//	ids
//
// where dls is followed by the depth limit, and m1 is minus 1 (-1).
var undirectedGraphDataOrderingMap = map[string][]int{
	"dfs":    {0, 1, 4, 5, 3, 2, 6},
	"bfs":    {0, 1, 2, 3, 4, 5, 6},
	"dls-0":  {0},
	"dls-1":  {0, 1, 2, 3},
	"dls-2":  {0, 1, 4, 5, 2, 6, 3},
	"dls-3":  nil, // it is the same as dfs and is set in function init
	"dls-m1": {},
	"ids":    {0, 1, 2, 3, 0, 1, 4, 5, 2, 6, 3, 0, 1, 4, 5, 3, 2, 6},
}

// undirectedGraphDataVertexPathMap is a mapping from algorithm short names
// to lists of paths from the root to each vertex of undirectedGraphData.
//
// Valid keys:
//
//	dfs
//	bfs
//	dls-0
//	dls-1
//	dls-2
//	dls-3
//	dls-m1
//	ids
//
// where dls is followed by the depth limit, and m1 is minus 1 (-1).
var undirectedGraphDataVertexPathMap = map[string][][]int{
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
	"dls-2":  nil, // it is the same as bfs and is set in function init
	"dls-3":  nil, // it is the same as dfs and is set in function init
	"dls-m1": {},
	"ids":    nil, // it is the same as bfs and is set in function init
}

func init() {
	undirectedGraphDataOrderingMap["dls-3"] = undirectedGraphDataOrderingMap["dfs"]

	undirectedGraphDataVertexPathMap["dls-2"] = undirectedGraphDataVertexPathMap["bfs"]
	undirectedGraphDataVertexPathMap["dls-3"] = undirectedGraphDataVertexPathMap["dfs"]
	undirectedGraphDataVertexPathMap["ids"] = undirectedGraphDataVertexPathMap["bfs"]
}

func TestDFS(t *testing.T) {
	testBruteForceSearch(t, "DFS")
}

func TestDFSPath(t *testing.T) {
	testBruteForceSearchPath(t, "DFSPath")
}

func TestBFS(t *testing.T) {
	testBruteForceSearch(t, "BFS")
}

func TestBFSPath(t *testing.T) {
	testBruteForceSearchPath(t, "BFSPath")
}

func TestDLS(t *testing.T) {
	for _, limit := range []int{0, 1, 2, 3, -1} {
		var name string
		if limit >= 0 {
			name = fmt.Sprintf("DLS-%d", limit)
		} else {
			name = fmt.Sprintf("DLS-m%d", -limit)
		}
		t.Run(fmt.Sprintf("limit=%d", limit), func(t *testing.T) {
			testBruteForceSearch(t, name)
		})
	}
}

func TestDLSPath(t *testing.T) {
	for _, limit := range []int{0, 1, 2, 3, -1} {
		var name string
		if limit >= 0 {
			name = fmt.Sprintf("DLSPath-%d", limit)
		} else {
			name = fmt.Sprintf("DLSPath-m%d", -limit)
		}
		t.Run(fmt.Sprintf("limit=%d", limit), func(t *testing.T) {
			testBruteForceSearchPath(t, name)
		})
	}
}

func TestIDS(t *testing.T) {
	t.Run("initLimit=1", func(t *testing.T) {
		testBruteForceSearch(t, "IDS")
	})
	t.Run("initLimit=-1", func(t *testing.T) {
		testBruteForceSearch(t, "IDS-m")
	})
}

func TestIDSPath(t *testing.T) {
	t.Run("initLimit=1", func(t *testing.T) {
		testBruteForceSearchPath(t, "IDSPath")
	})
	t.Run("initLimit=-1", func(t *testing.T) {
		testBruteForceSearchPath(t, "IDSPath-m")
	})
}

// testBruteForceSearchFunc is the type of function
// testing the brute force search function.
type testBruteForceSearchFunc func(
	t *testing.T,
	itf graphv.IDSAccessVertex[int],
	initArgs ...any,
) (vertexFound int, found bool)

func testBruteForceSearch(t *testing.T, name string) {
	f, ordering := getTestBruteForceSearchFunctionAndOrdering(t, name)
	if t.Failed() {
		return
	}

	tg := &graphImpl{Data: undirectedGraphData}
	for goal := range NumUndirectedGraphVertex {
		t.Run(fmt.Sprintf("goal=%d", goal), func(t *testing.T) {
			var i int
			for i < len(ordering) && ordering[i] != goal {
				i++
			}
			var wantVertex int // the vertex expected to be found
			var wantFound bool // expected found value
			wantHx := ordering // expected history
			if i < len(ordering) {
				wantVertex, wantFound, wantHx = goal, true, wantHx[:1+i]
			}
			r, found := f(t, tg, goal)
			if found != wantFound || r != wantVertex {
				t.Errorf("got <%d, %t>; want <%d, %t>",
					r, found, wantVertex, wantFound)
			}
			checkAccessHistory(t, tg, wantHx)
		})
	}
	// Non-existent vertices:
	for _, goal := range []any{nil, -1, len(undirectedGraphData), 1.2} {
		t.Run(fmt.Sprintf("goal=%v", goal), func(t *testing.T) {
			wantHx := ordering // expected history
			if len(wantHx) > 1 {
				if _, ok := goal.(int); !ok {
					wantHx = wantHx[:1]
				}
			}
			r, found := f(t, tg, goal)
			if r != 0 || found {
				t.Errorf("got <%d, %t>; want <0, false>", r, found)
			}
			checkAccessHistory(t, tg, wantHx)
		})
	}
}

// getTestBruteForceSearchFunctionAndOrdering returns the function
// testing the brute force search function
// and the expected vertex access ordering
// according to the specified name.
//
// It reports an error using t.Errorf if the name is unacceptable.
func getTestBruteForceSearchFunctionAndOrdering(t *testing.T, name string) (
	f testBruteForceSearchFunc, ordering []int) {
	checkMore := func(t *testing.T, initArgs []any, found bool, more bool) {
		if isFirstInitArgInt(initArgs) && !found && !more {
			t.Error("more is false but there are undiscovered vertices")
		}
	}
	switch name {
	case "DFS":
		f = func(
			_ *testing.T,
			itf graphv.IDSAccessVertex[int],
			initArgs ...any,
		) (vertexFound int, found bool) {
			return graphv.DFS[int](itf, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["dfs"]
	case "BFS":
		f = func(
			_ *testing.T,
			itf graphv.IDSAccessVertex[int],
			initArgs ...any,
		) (vertexFound int, found bool) {
			return graphv.BFS[int](itf, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["bfs"]
	case "DLS-0":
		f = func(
			t *testing.T,
			itf graphv.IDSAccessVertex[int],
			initArgs ...any,
		) (vertexFound int, found bool) {
			vertex, found, more := graphv.DLS[int](itf, 0, initArgs...)
			checkMore(t, initArgs, found, more)
			return vertex, found
		}
		ordering = undirectedGraphDataOrderingMap["dls-0"]
	case "DLS-1":
		f = func(
			t *testing.T,
			itf graphv.IDSAccessVertex[int],
			initArgs ...any,
		) (vertexFound int, found bool) {
			vertex, found, more := graphv.DLS[int](itf, 1, initArgs...)
			checkMore(t, initArgs, found, more)
			return vertex, found
		}
		ordering = undirectedGraphDataOrderingMap["dls-1"]
	case "DLS-2":
		f = func(
			_ *testing.T,
			itf graphv.IDSAccessVertex[int],
			initArgs ...any,
		) (vertexFound int, found bool) {
			vertex, found, _ := graphv.DLS[int](itf, 2, initArgs...)
			// Both true and false are acceptable for the third return value.
			return vertex, found
		}
		ordering = undirectedGraphDataOrderingMap["dls-2"]
	case "DLS-3":
		f = func(
			_ *testing.T,
			itf graphv.IDSAccessVertex[int],
			initArgs ...any,
		) (vertexFound int, found bool) {
			vertex, found, _ := graphv.DLS[int](itf, 3, initArgs...)
			// Both true and false are acceptable for the third return value.
			return vertex, found
		}
		ordering = undirectedGraphDataOrderingMap["dls-3"]
	case "DLS-m1":
		f = func(
			t *testing.T,
			itf graphv.IDSAccessVertex[int],
			initArgs ...any,
		) (vertexFound int, found bool) {
			vertex, found, more := graphv.DLS[int](itf, -1, initArgs...)
			checkMore(t, initArgs, found, more)
			return vertex, found
		}
		ordering = undirectedGraphDataOrderingMap["dls-m1"]
	case "IDS":
		f = func(
			_ *testing.T,
			itf graphv.IDSAccessVertex[int],
			initArgs ...any,
		) (vertexFound int, found bool) {
			return graphv.IDS(itf, 1, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["ids"]
	case "IDS-m":
		f = func(
			_ *testing.T,
			itf graphv.IDSAccessVertex[int],
			initArgs ...any,
		) (vertexFound int, found bool) {
			return graphv.IDS(itf, -1, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["ids"]
	default:
		t.Errorf("unacceptable name %q", name)
	}
	return
}

// testBruteForceSearchPathFunc is the type of function
// testing the brute force search path function.
type testBruteForceSearchPathFunc func(
	t *testing.T,
	itf graphv.IDSAccessPath[int],
	initArgs ...any,
) []int

func testBruteForceSearchPath(t *testing.T, name string) {
	f, ordering := getTestBruteForceSearchPathFunctionAndOrdering(t, name)
	if t.Failed() {
		return
	}

	tg := &graphImpl{Data: undirectedGraphData}
	for goal := range NumUndirectedGraphVertex {
		t.Run(fmt.Sprintf("goal=%d", goal), func(t *testing.T) {
			path := f(t, tg, goal)
			checkPath(t, name, goal, path)
			var i int
			for i < len(ordering) && ordering[i] != goal {
				i++
			}
			wantHx := ordering // expected history
			if i < len(ordering) {
				wantHx = wantHx[:1+i]
			}
			checkAccessHistory(t, tg, wantHx)
		})
	}
	// Non-existent vertices:
	for _, goal := range []any{nil, -1, len(undirectedGraphData), 1.2} {
		t.Run(fmt.Sprintf("goal=%v", goal), func(t *testing.T) {
			wantHx := ordering // expected history
			goalVertex := -1
			if g, ok := goal.(int); ok {
				goalVertex = g
			} else if len(wantHx) > 1 {
				wantHx = wantHx[:1]
			}
			path := f(t, tg, goal)
			checkPath(t, name, goalVertex, path)
			checkAccessHistory(t, tg, wantHx)
		})
	}
}

// getTestBruteForceSearchPathFunctionAndOrdering returns the function
// testing the brute force search path function
// and the expected vertex access ordering
// according to the specified name.
//
// It reports an error using t.Errorf if the name is unacceptable.
func getTestBruteForceSearchPathFunctionAndOrdering(t *testing.T, name string) (
	f testBruteForceSearchPathFunc, ordering []int) {
	checkMore := func(t *testing.T, initArgs []any, path []int, more bool) {
		if isFirstInitArgInt(initArgs) && path == nil && !more {
			t.Error("more is false but there are undiscovered vertices")
		}
	}
	switch name {
	case "DFSPath":
		f = func(
			_ *testing.T,
			itf graphv.IDSAccessPath[int],
			initArgs ...any,
		) []int {
			return graphv.DFSPath[int](itf, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["dfs"]
	case "BFSPath":
		f = func(
			_ *testing.T,
			itf graphv.IDSAccessPath[int],
			initArgs ...any,
		) []int {
			return graphv.BFSPath[int](itf, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["bfs"]
	case "DLSPath-0":
		f = func(
			t *testing.T,
			itf graphv.IDSAccessPath[int],
			initArgs ...any,
		) []int {
			path, more := graphv.DLSPath[int](itf, 0, initArgs...)
			checkMore(t, initArgs, path, more)
			return path
		}
		ordering = undirectedGraphDataOrderingMap["dls-0"]
	case "DLSPath-1":
		f = func(
			t *testing.T,
			itf graphv.IDSAccessPath[int],
			initArgs ...any,
		) []int {
			path, more := graphv.DLSPath[int](itf, 1, initArgs...)
			checkMore(t, initArgs, path, more)
			return path
		}
		ordering = undirectedGraphDataOrderingMap["dls-1"]
	case "DLSPath-2":
		f = func(
			_ *testing.T,
			itf graphv.IDSAccessPath[int],
			initArgs ...any,
		) []int {
			path, _ := graphv.DLSPath[int](itf, 2, initArgs...)
			// Both true and false are acceptable for the second return value.
			return path
		}
		ordering = undirectedGraphDataOrderingMap["dls-2"]
	case "DLSPath-3":
		f = func(
			_ *testing.T,
			itf graphv.IDSAccessPath[int],
			initArgs ...any,
		) []int {
			path, _ := graphv.DLSPath[int](itf, 3, initArgs...)
			// Both true and false are acceptable for the second return value.
			return path
		}
		ordering = undirectedGraphDataOrderingMap["dls-3"]
	case "DLSPath-m1":
		f = func(
			t *testing.T,
			itf graphv.IDSAccessPath[int],
			initArgs ...any,
		) []int {
			path, more := graphv.DLSPath[int](itf, -1, initArgs...)
			checkMore(t, initArgs, path, more)
			return path
		}
		ordering = undirectedGraphDataOrderingMap["dls-m1"]
	case "IDSPath":
		f = func(
			_ *testing.T,
			itf graphv.IDSAccessPath[int],
			initArgs ...any,
		) []int {
			return graphv.IDSPath(itf, 1, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["ids"]
	case "IDSPath-m":
		f = func(
			_ *testing.T,
			itf graphv.IDSAccessPath[int],
			initArgs ...any,
		) []int {
			return graphv.IDSPath(itf, -1, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["ids"]
	default:
		t.Errorf("unacceptable name %q", name)
	}
	return
}

func isFirstInitArgInt(initArgs []any) bool {
	if len(initArgs) < 1 {
		return false
	}
	_, ok := initArgs[0].(int)
	return ok
}

func checkAccessHistory(t *testing.T, tg *graphImpl, want []int) {
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

func checkPath(t *testing.T, name string, vertex int, pathList []int) {
	var list [][]int
	switch name {
	case "DFSPath":
		list = undirectedGraphDataVertexPathMap["dfs"]
	case "BFSPath":
		list = undirectedGraphDataVertexPathMap["bfs"]
	case "DLSPath-0":
		list = undirectedGraphDataVertexPathMap["dls-0"]
	case "DLSPath-1":
		list = undirectedGraphDataVertexPathMap["dls-1"]
	case "DLSPath-2":
		list = undirectedGraphDataVertexPathMap["dls-2"]
	case "DLSPath-3":
		list = undirectedGraphDataVertexPathMap["dls-3"]
	case "DLSPath-m1":
		list = undirectedGraphDataVertexPathMap["dls-m1"]
	case "IDSPath", "IDSPath-m":
		list = undirectedGraphDataVertexPathMap["ids"]
	default:
		// This should never happen, but will act as a safeguard for later,
		// as a default value doesn't make sense here.
		t.Errorf("unacceptable name %q", name)
		return
	}

	var wantPath []int
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
