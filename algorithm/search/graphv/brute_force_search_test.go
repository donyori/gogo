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

package graphv

import "testing"

// testGraphBase implements all methods of interface IdsInterface
// except Adjacency.
type testGraphBase struct {
	Data          [][]int
	Goal          interface{}
	AccessHistory []interface{}

	// Index of the beginning of the access history
	// in the most recent search iteration.
	head int
}

func (tg *testGraphBase) Root() interface{} {
	return 0
}

// SetGoal sets the search goal as well as resets the access history.
func (tg *testGraphBase) SetGoal(goal interface{}) {
	tg.Goal = goal
	tg.AccessHistory = tg.AccessHistory[:0] // Reuse the underlying array.
	tg.head = 0
}

func (tg *testGraphBase) Access(vertex interface{}) (found bool) {
	tg.AccessHistory = append(tg.AccessHistory, vertex)
	if tg.Goal == nil {
		return false
	}
	return vertex == tg.Goal
}

// Discovered reports whether the specified vertex
// has been examined by the method Access
// via checking the access history.
//
// It may be time-consuming,
// but it doesn't matter because the test graph is very small.
func (tg *testGraphBase) Discovered(vertex interface{}) bool {
	for i := tg.head; i < len(tg.AccessHistory); i++ {
		if tg.AccessHistory[i] == vertex {
			return true
		}
	}
	return false
}

func (tg *testGraphBase) ResetSearchState() {
	tg.head = len(tg.AccessHistory)
}

// testGraphNormal attaches the method Adjacency with normal performance
// to testGraphBase.
type testGraphNormal struct {
	testGraphBase
}

func (tg *testGraphNormal) Adjacency(vertex interface{}) []interface{} {
	list := tg.Data[vertex.(int)]
	if len(list) == 0 {
		return nil
	}
	adj := make([]interface{}, len(list))
	for i := range list {
		adj[i] = list[i]
	}
	return adj
}

// testGraphNilAdjacentVertices attaches the method Adjacency to testGraphBase.
// Its method Adjacency returns an adjacency list with
// some additional nil vertices.
type testGraphNilAdjacentVertices struct {
	testGraphBase
}

func (tg *testGraphNilAdjacentVertices) Adjacency(vertex interface{}) []interface{} {
	list := tg.Data[vertex.(int)]
	if len(list) == 0 {
		return nil
	}
	adj := make([]interface{}, len(list)+4)
	// Add 4 nil vertices for testing.
	switch len(list) {
	case 1:
		// adj: {nil, nil, nil, list[0], nil}
		adj[4] = list[0]
	case 2:
		// adj: {nil, list[0], nil, nil, list[1], nil}
		adj[1] = list[0]
		adj[5] = list[1]
	default:
		// adj: {list[0], nil, list[1], ..., list[n-2], nil, nil, nil, list[n-1]}
		// where n = len(list)
		adj[0] = list[0]
		for i := 1; i < len(list)-1; i++ {
			adj[i+1] = list[i]
		}
		adj[len(adj)-1] = list[len(list)-1]
	}
	return adj
}

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
//  ids
var testUndirectedGraphDataOrderingMap = map[string][]int{
	"dfs":   {0, 1, 4, 5, 3, 2, 6},
	"bfs":   {0, 1, 2, 3, 4, 5, 6},
	"dls-0": {0},
	"dls-1": {0, 1, 2, 3},
	"dls-2": {0, 1, 4, 5, 2, 6, 3},
	"dls-3": nil, // It is the same as dfs and will be set in function init.
	"ids":   {0, 0, 1, 2, 3, 0, 1, 4, 5, 2, 6, 3, 0, 1, 4, 5, 3, 2, 6},
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
//  ids
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
	"dls-2": nil, // It is the same as bfs and will be set in function init.
	"dls-3": nil, // It is the same as dfs and will be set in function init.
	"ids":   nil, // It is the same as bfs and will be set in function init.
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
	testBruteForceSearch(t, "Dls-0")
	testBruteForceSearch(t, "Dls-1")
	testBruteForceSearch(t, "Dls-2")
	testBruteForceSearch(t, "Dls-3")
	testBruteForceSearch(t, "Dls-m1")
}

func TestDlsPath(t *testing.T) {
	testBruteForceSearchPath(t, "DlsPath-0")
	testBruteForceSearchPath(t, "DlsPath-1")
	testBruteForceSearchPath(t, "DlsPath-2")
	testBruteForceSearchPath(t, "DlsPath-3")
	testBruteForceSearchPath(t, "DlsPath-m1")
}

func TestIds(t *testing.T) {
	testBruteForceSearch(t, "Ids")
	testBruteForceSearch(t, "Ids-m")
}

func TestIdsPath(t *testing.T) {
	testBruteForceSearchPath(t, "IdsPath")
	testBruteForceSearchPath(t, "IdsPath-m")
}

func testBruteForceSearch(t *testing.T, name string) {
	var f func(itf BasicInterface, goal interface{}) interface{}
	var ordering []int
	switch name {
	case "Dfs":
		f = Dfs
		ordering = testUndirectedGraphDataOrderingMap["dfs"]
	case "Bfs":
		f = Bfs
		ordering = testUndirectedGraphDataOrderingMap["bfs"]
	case "Dls-0":
		f = func(itf BasicInterface, goal interface{}) interface{} {
			v, more := Dls(itf, goal, 0)
			if v == nil && !more {
				t.Error("Dls-0 - more is false but there are undiscovered vertices.")
			}
			return v
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-0"]
	case "Dls-1":
		f = func(itf BasicInterface, goal interface{}) interface{} {
			v, more := Dls(itf, goal, 1)
			if v == nil && !more {
				t.Error("Dls-1 - more is false but there are undiscovered vertices.")
			}
			return v
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-1"]
	case "Dls-2":
		f = func(itf BasicInterface, goal interface{}) interface{} {
			v, _ := Dls(itf, goal, 2)
			// Both true and false are acceptable for the second return value.
			return v
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-2"]
	case "Dls-3":
		f = func(itf BasicInterface, goal interface{}) interface{} {
			v, _ := Dls(itf, goal, 3)
			// Both true and false are acceptable for the second return value.
			return v
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-3"]
	case "Dls-m1":
		f = func(itf BasicInterface, goal interface{}) interface{} {
			v, more := Dls(itf, goal, -1)
			if v == nil && !more {
				t.Error("Dls-m1 - more is false but there are undiscovered vertices.")
			}
			return v
		}
		ordering = []int{}
	case "Ids":
		f = func(itf BasicInterface, goal interface{}) interface{} {
			return Ids(itf.(IdsInterface), goal, 0)
		}
		ordering = testUndirectedGraphDataOrderingMap["ids"]
	case "Ids-m":
		f = func(itf BasicInterface, goal interface{}) interface{} {
			return Ids(itf.(IdsInterface), goal, -1)
		}
		ordering = testUndirectedGraphDataOrderingMap["ids"]
	default:
		t.Error("Unacceptable name:", name)
		return
	}

	for _, tg := range []IdsInterface{
		&testGraphNormal{testGraphBase{Data: testUndirectedGraphData}},
		&testGraphNilAdjacentVertices{testGraphBase{Data: testUndirectedGraphData}},
	} {
		var tgb *testGraphBase
		switch tg.(type) {
		case *testGraphNormal:
			tgb = &tg.(*testGraphNormal).testGraphBase
		case *testGraphNilAdjacentVertices:
			tgb = &tg.(*testGraphNilAdjacentVertices).testGraphBase
		default:
			// This should never happen, but will act as a safeguard for later,
			// as a default value doesn't make sense here.
			t.Errorf("tg is neither of type *testGraphNormal nor of type *testGraphNilAdjacentVertices, type: %T", tg)
			return
		}
		tested := make([]bool, len(testUndirectedGraphData))
		for i, v := range ordering {
			if tested[v] {
				continue
			}
			tested[v] = true
			r := f(tg, v)
			if r != v {
				t.Errorf("%s returns %v != %v.", name, r, v)
			}
			testCheckAccessHistory(t, name, tgb, ordering[:1+i])
		}
		// Non-existent nodes:
		for _, goal := range []interface{}{nil, -1, len(testUndirectedGraphData), 1.2} {
			r := f(tg, goal)
			if r != nil {
				t.Errorf("%s returns %v != nil.", name, r)
			}
			testCheckAccessHistory(t, name, tgb, ordering)
		}
	}
}

func testBruteForceSearchPath(t *testing.T, name string) {
	var f func(itf BasicInterface, goal interface{}) []interface{}
	var ordering []int
	switch name {
	case "DfsPath":
		f = DfsPath
		ordering = testUndirectedGraphDataOrderingMap["dfs"]
	case "BfsPath":
		f = BfsPath
		ordering = testUndirectedGraphDataOrderingMap["bfs"]
	case "DlsPath-0":
		f = func(itf BasicInterface, goal interface{}) []interface{} {
			p, more := DlsPath(itf, goal, 0)
			if p == nil && !more {
				t.Error("DlsPath-0 - more is false but there are undiscovered vertices.")
			}
			return p
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-0"]
	case "DlsPath-1":
		f = func(itf BasicInterface, goal interface{}) []interface{} {
			p, more := DlsPath(itf, goal, 1)
			if p == nil && !more {
				t.Error("DlsPath-1 - more is false but there are undiscovered vertices.")
			}
			return p
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-1"]
	case "DlsPath-2":
		f = func(itf BasicInterface, goal interface{}) []interface{} {
			p, _ := DlsPath(itf, goal, 2)
			// Both true and false are acceptable for the second return value.
			return p
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-2"]
	case "DlsPath-3":
		f = func(itf BasicInterface, goal interface{}) []interface{} {
			p, _ := DlsPath(itf, goal, 3)
			// Both true and false are acceptable for the second return value.
			return p
		}
		ordering = testUndirectedGraphDataOrderingMap["dls-3"]
	case "DlsPath-m1":
		f = func(itf BasicInterface, goal interface{}) []interface{} {
			p, more := DlsPath(itf, goal, -1)
			if p == nil && !more {
				t.Error("DlsPath-m1 - more is false but there are undiscovered vertices.")
			}
			return p
		}
		ordering = []int{}
	case "IdsPath":
		f = func(itf BasicInterface, goal interface{}) []interface{} {
			return IdsPath(itf.(IdsInterface), goal, 0)
		}
		ordering = testUndirectedGraphDataOrderingMap["ids"]
	case "IdsPath-m":
		f = func(itf BasicInterface, goal interface{}) []interface{} {
			return IdsPath(itf.(IdsInterface), goal, -1)
		}
		ordering = testUndirectedGraphDataOrderingMap["ids"]
	default:
		t.Error("Unacceptable name:", name)
		return
	}

	for _, tg := range []IdsInterface{
		&testGraphNormal{testGraphBase{Data: testUndirectedGraphData}},
		&testGraphNilAdjacentVertices{testGraphBase{Data: testUndirectedGraphData}},
	} {
		var tgb *testGraphBase
		switch tg.(type) {
		case *testGraphNormal:
			tgb = &tg.(*testGraphNormal).testGraphBase
		case *testGraphNilAdjacentVertices:
			tgb = &tg.(*testGraphNilAdjacentVertices).testGraphBase
		default:
			// This should never happen, but will act as a safeguard for later,
			// as a default value doesn't make sense here.
			t.Errorf("tg is neither of type *testGraphNormal nor of type *testGraphNilAdjacentVertices, type: %T", tg)
			return
		}
		tested := make([]bool, len(testUndirectedGraphData))
		for i, v := range ordering {
			if tested[v] {
				continue
			}
			tested[v] = true
			list := f(tg, v)
			testCheckPath(t, name, v, list)
			testCheckAccessHistory(t, name, tgb, ordering[:1+i])
		}
		// Non-existent nodes:
		for _, goal := range []interface{}{nil, -1, len(testUndirectedGraphData), 1.2} {
			list := f(tg, goal)
			testCheckPath(t, name, goal, list)
			testCheckAccessHistory(t, name, tgb, ordering)
		}
	}
}

func testCheckAccessHistory(t *testing.T, name string, tg *testGraphBase, wanted []int) {
	if len(tg.AccessHistory) != len(wanted) {
		t.Errorf("%s - Access history: %v\nwanted: %v", name, tg.AccessHistory, wanted)
		return
	}
	for i := range wanted {
		if tg.AccessHistory[i] != wanted[i] {
			t.Errorf("%s - Access history: %v\nwanted: %v", name, tg.AccessHistory, wanted)
			return
		}
	}
}

func testCheckPath(t *testing.T, name string, vertex interface{}, pathList []interface{}) {
	var p []int
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
		list = [][]int{}
	case "IdsPath", "IdsPath-m":
		list = testUndirectedGraphDataVertexPathMap["ids"]
	default:
		// This should never happen, but will act as a safeguard for later,
		// as a default value doesn't make sense here.
		t.Error("Unacceptable name:", name)
		return
	}
	idx, ok := vertex.(int)
	if ok && idx >= 0 && idx < len(list) {
		p = list[idx]
	}
	if len(pathList) != len(p) {
		t.Errorf("%s - Path of %v: %v\nwanted: %v", name, vertex, pathList, p)
		return
	}
	for i := range p {
		if pathList[i] != p[i] {
			t.Errorf("%s - Path of %v: %v\nwanted: %v", name, vertex, pathList, p)
			return
		}
	}
}
