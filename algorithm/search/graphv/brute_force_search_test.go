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

package graphv_test

import (
	"fmt"
	"testing"

	"github.com/donyori/gogo/algorithm/search/graphv"
)

// graphImpl implements all methods of interface IdsInterface[int].
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
	g.AccessHistory = g.AccessHistory[:0] // Reuse the underlying array.
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
	return append(list[:0:0], list...) // Return a copy of list.
}

func (g *graphImpl) Access(vertex, _ int) (found, cont bool) {
	g.AccessHistory = append(g.AccessHistory, vertex)
	if !g.GoalValid {
		return
	}
	return vertex == g.Goal, true
}

// Discovered reports whether the specified vertex
// has been examined by the method Access
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

func (g *graphImpl) ResetSearchState() {
	g.head = len(g.AccessHistory)
}

// numUndirectedGraphVertices is the number of vertices
// in undirectedGraphData.
const numUndirectedGraphVertices int = 7

// undirectedGraphData represents an undirected graph as follows:
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
//  dfs
//  bfs
//  dls-0
//  dls-1
//  dls-2
//  dls-3
//  dls-m1
//  ids
// where dls is followed by the depth limit, and m1 is minus 1 (-1).
var undirectedGraphDataOrderingMap = map[string][]int{
	"dfs":    {0, 1, 4, 5, 3, 2, 6},
	"bfs":    {0, 1, 2, 3, 4, 5, 6},
	"dls-0":  {0},
	"dls-1":  {0, 1, 2, 3},
	"dls-2":  {0, 1, 4, 5, 2, 6, 3},
	"dls-3":  nil, // It is the same as dfs and will be set in function init.
	"dls-m1": {},
	"ids":    {0, 1, 2, 3, 0, 1, 4, 5, 2, 6, 3, 0, 1, 4, 5, 3, 2, 6},
}

// undirectedGraphDataVertexPathMap is a mapping from algorithm short names
// to lists of paths from the root to each vertex of undirectedGraphData.
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
	"dls-2":  nil, // It is the same as bfs and will be set in function init.
	"dls-3":  nil, // It is the same as dfs and will be set in function init.
	"dls-m1": {},
	"ids":    nil, // It is the same as bfs and will be set in function init.
}

func init() {
	undirectedGraphDataOrderingMap["dls-3"] = undirectedGraphDataOrderingMap["dfs"]

	undirectedGraphDataVertexPathMap["dls-2"] = undirectedGraphDataVertexPathMap["bfs"]
	undirectedGraphDataVertexPathMap["dls-3"] = undirectedGraphDataVertexPathMap["dfs"]
	undirectedGraphDataVertexPathMap["ids"] = undirectedGraphDataVertexPathMap["bfs"]
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
	var f func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) (int, bool)
	var ordering []int
	switch name {
	case "Dfs":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) (int, bool) {
			return graphv.Dfs[int](itf, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["dfs"]
	case "Bfs":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) (int, bool) {
			return graphv.Bfs[int](itf, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["bfs"]
	case "Dls-0":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) (int, bool) {
			vertex, found, more := graphv.Dls[int](itf, 0, initArgs...)
			if isFirstInitArgInt(initArgs) && !found && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return vertex, found
		}
		ordering = undirectedGraphDataOrderingMap["dls-0"]
	case "Dls-1":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) (int, bool) {
			vertex, found, more := graphv.Dls[int](itf, 1, initArgs...)
			if isFirstInitArgInt(initArgs) && !found && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return vertex, found
		}
		ordering = undirectedGraphDataOrderingMap["dls-1"]
	case "Dls-2":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) (int, bool) {
			vertex, found, _ := graphv.Dls[int](itf, 2, initArgs...)
			// Both true and false are acceptable for the third return value.
			return vertex, found
		}
		ordering = undirectedGraphDataOrderingMap["dls-2"]
	case "Dls-3":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) (int, bool) {
			vertex, found, _ := graphv.Dls[int](itf, 3, initArgs...)
			// Both true and false are acceptable for the third return value.
			return vertex, found
		}
		ordering = undirectedGraphDataOrderingMap["dls-3"]
	case "Dls-m1":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) (int, bool) {
			vertex, found, more := graphv.Dls[int](itf, -1, initArgs...)
			if isFirstInitArgInt(initArgs) && !found && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return vertex, found
		}
		ordering = undirectedGraphDataOrderingMap["dls-m1"]
	case "Ids":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) (int, bool) {
			return graphv.Ids(itf, 1, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["ids"]
	case "Ids-m":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) (int, bool) {
			return graphv.Ids(itf, -1, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["ids"]
	default:
		t.Errorf("unacceptable name %q", name)
		return
	}

	tg := &graphImpl{Data: undirectedGraphData}
	for goal := 0; goal < numUndirectedGraphVertices; goal++ {
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
			checkAccessHistory(t, tg, wantHx)
		})
	}
	// Non-existent nodes:
	for _, goal := range []any{nil, -1, len(undirectedGraphData), 1.2} {
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
			checkAccessHistory(t, tg, wantHx)
		})
	}
}

func testBruteForceSearchPath(t *testing.T, name string) {
	var f func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) []int
	var ordering []int
	switch name {
	case "DfsPath":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) []int {
			return graphv.DfsPath[int](itf, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["dfs"]
	case "BfsPath":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) []int {
			return graphv.BfsPath[int](itf, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["bfs"]
	case "DlsPath-0":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) []int {
			path, more := graphv.DlsPath[int](itf, 0, initArgs...)
			if isFirstInitArgInt(initArgs) && path == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return path
		}
		ordering = undirectedGraphDataOrderingMap["dls-0"]
	case "DlsPath-1":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) []int {
			path, more := graphv.DlsPath[int](itf, 1, initArgs...)
			if isFirstInitArgInt(initArgs) && path == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return path
		}
		ordering = undirectedGraphDataOrderingMap["dls-1"]
	case "DlsPath-2":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) []int {
			path, _ := graphv.DlsPath[int](itf, 2, initArgs...)
			// Both true and false are acceptable for the second return value.
			return path
		}
		ordering = undirectedGraphDataOrderingMap["dls-2"]
	case "DlsPath-3":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) []int {
			path, _ := graphv.DlsPath[int](itf, 3, initArgs...)
			// Both true and false are acceptable for the second return value.
			return path
		}
		ordering = undirectedGraphDataOrderingMap["dls-3"]
	case "DlsPath-m1":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) []int {
			path, more := graphv.DlsPath[int](itf, -1, initArgs...)
			if isFirstInitArgInt(initArgs) && path == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return path
		}
		ordering = undirectedGraphDataOrderingMap["dls-m1"]
	case "IdsPath":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) []int {
			return graphv.IdsPath(itf, 1, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["ids"]
	case "IdsPath-m":
		f = func(t *testing.T, itf graphv.IdsInterface[int], initArgs ...any) []int {
			return graphv.IdsPath(itf, -1, initArgs...)
		}
		ordering = undirectedGraphDataOrderingMap["ids"]
	default:
		t.Errorf("unacceptable name %q", name)
		return
	}

	tg := &graphImpl{Data: undirectedGraphData}
	for goal := 0; goal < numUndirectedGraphVertices; goal++ {
		t.Run(fmt.Sprintf("goal=%d", goal), func(t *testing.T) {
			path := f(t, tg, goal)
			checkPath(t, name, goal, path)
			var i int
			for i < len(ordering) && ordering[i] != goal {
				i++
			}
			wantHx := ordering // Expected history.
			if i < len(ordering) {
				wantHx = wantHx[:1+i]
			}
			checkAccessHistory(t, tg, wantHx)
		})
	}
	// Non-existent nodes:
	for _, goal := range []any{nil, -1, len(undirectedGraphData), 1.2} {
		t.Run(fmt.Sprintf("goal=%v", goal), func(t *testing.T) {
			wantHx := ordering // Expected history.
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
	var wantPath []int
	var list [][]int
	switch name {
	case "DfsPath":
		list = undirectedGraphDataVertexPathMap["dfs"]
	case "BfsPath":
		list = undirectedGraphDataVertexPathMap["bfs"]
	case "DlsPath-0":
		list = undirectedGraphDataVertexPathMap["dls-0"]
	case "DlsPath-1":
		list = undirectedGraphDataVertexPathMap["dls-1"]
	case "DlsPath-2":
		list = undirectedGraphDataVertexPathMap["dls-2"]
	case "DlsPath-3":
		list = undirectedGraphDataVertexPathMap["dls-3"]
	case "DlsPath-m1":
		list = undirectedGraphDataVertexPathMap["dls-m1"]
	case "IdsPath", "IdsPath-m":
		list = undirectedGraphDataVertexPathMap["ids"]
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
