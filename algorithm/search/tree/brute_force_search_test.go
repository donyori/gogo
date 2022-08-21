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

package tree_test

import (
	"fmt"
	"testing"

	"github.com/donyori/gogo/algorithm/search/tree"
)

type treeImpl struct {
	Data          [][2]int
	Goal          int
	GoalValid     bool
	AccessHistory []int
}

// Init sets the search goal as well as resets the access history.
func (t *treeImpl) Init(args ...any) {
	t.Goal, t.GoalValid = -1, false
	if len(args) >= 1 {
		goal, ok := args[0].(int)
		if ok {
			t.Goal, t.GoalValid = goal, true
		}
	}
	t.AccessHistory = t.AccessHistory[:0] // Reuse the underlying array.
}

func (t *treeImpl) Root() int {
	if len(t.Data) == 0 {
		return -1
	}
	return 0
}

func (t *treeImpl) FirstChild(node int) (fc int, ok bool) {
	idx := t.Data[node][0]
	if idx < 0 {
		return
	}
	return idx, true
}

func (t *treeImpl) NextSibling(node int) (ns int, ok bool) {
	idx := t.Data[node][1]
	if idx < 0 {
		return
	}
	return idx, true
}

func (t *treeImpl) Access(node, _ int) (found, cont bool) {
	t.AccessHistory = append(t.AccessHistory, node)
	if !t.GoalValid {
		return
	}
	return node == t.Goal, true
}

// numTreeNode is the number of nodes in treeData.
const numTreeNode int = 12

// treeData represents a tree as follows:
//
//	      0
//	     /|\
//	    1 2 3
//	   /|   |\
//	  4 5   6 7
//	 /|     |\
//	8 9    10 11
//
// Assuming that the left edges are chosen before the right edges,
// the expected orderings are as follows:
//
// Expected DFS ordering:
//
//	0, 1, 4, 8, 9, 5, 2, 3, 6, 10, 11, 7
//
// Expected BFS ordering:
//
//	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11
var treeData = [][2]int{
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

// treeDataOrderingMap is a mapping from algorithm short names
// to the expected node access orderings of treeData.
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
var treeDataOrderingMap = map[string][]int{
	"dfs":    {0, 1, 4, 8, 9, 5, 2, 3, 6, 10, 11, 7},
	"bfs":    {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	"dls-0":  {0},
	"dls-1":  {0, 1, 2, 3},
	"dls-2":  {0, 1, 4, 5, 2, 3, 6, 7},
	"dls-3":  nil, // It is the same as dfs and will be set in function init.
	"dls-m1": {},
	"ids":    {0, 1, 2, 3, 0, 1, 4, 5, 2, 3, 6, 7, 0, 1, 4, 8, 9, 5, 2, 3, 6, 10, 11, 7},
}

// treeDataNodePath is a list of paths from the root
// to each node of treeData.
var treeDataNodePath = [][]int{
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
	treeDataOrderingMap["dls-3"] = treeDataOrderingMap["dfs"]
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
	var f func(t *testing.T, itf tree.Interface[int], initArgs ...any) (int, bool)
	var ordering []int
	switch name {
	case "Dfs":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) (int, bool) {
			return tree.Dfs(itf, initArgs...)
		}
		ordering = treeDataOrderingMap["dfs"]
	case "Bfs":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) (int, bool) {
			return tree.Bfs(itf, initArgs...)
		}
		ordering = treeDataOrderingMap["bfs"]
	case "Dls-0":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) (int, bool) {
			node, found, more := tree.Dls(itf, 0, initArgs...)
			if isFirstInitArgInt(initArgs) && !found && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return node, found
		}
		ordering = treeDataOrderingMap["dls-0"]
	case "Dls-1":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) (int, bool) {
			node, found, more := tree.Dls(itf, 1, initArgs...)
			if isFirstInitArgInt(initArgs) && !found && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return node, found
		}
		ordering = treeDataOrderingMap["dls-1"]
	case "Dls-2":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) (int, bool) {
			node, found, more := tree.Dls(itf, 2, initArgs...)
			if isFirstInitArgInt(initArgs) && !found && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return node, found
		}
		ordering = treeDataOrderingMap["dls-2"]
	case "Dls-3":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) (int, bool) {
			node, found, _ := tree.Dls(itf, 3, initArgs...)
			// Both true and false are acceptable for the third return value.
			return node, found
		}
		ordering = treeDataOrderingMap["dls-3"]
	case "Dls-m1":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) (int, bool) {
			node, found, more := tree.Dls(itf, -1, initArgs...)
			if isFirstInitArgInt(initArgs) && !found && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return node, found
		}
		ordering = treeDataOrderingMap["dls-m1"]
	case "Ids":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) (int, bool) {
			return tree.Ids(itf, 1, initArgs...)
		}
		ordering = treeDataOrderingMap["ids"]
	case "Ids-m":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) (int, bool) {
			return tree.Ids(itf, -1, initArgs...)
		}
		ordering = treeDataOrderingMap["ids"]
	default:
		t.Errorf("unacceptable name %q", name)
		return
	}

	tt := &treeImpl{Data: treeData}
	for goal := 0; goal < numTreeNode; goal++ {
		t.Run(fmt.Sprintf("goal=%d", goal), func(t *testing.T) {
			var i int
			for i < len(ordering) && ordering[i] != goal {
				i++
			}
			var wantNode int   // The node expected to be found.
			var wantFound bool // Expected found value.
			wantHx := ordering // Expected history.
			if i < len(ordering) {
				wantNode, wantFound, wantHx = goal, true, wantHx[:1+i]
			}
			r, found := f(t, tt, goal)
			if found != wantFound || r != wantNode {
				t.Errorf("got <%d, %t>; want <%d, %t>", r, found, wantNode, wantFound)
			}
			checkAccessHistory(t, tt, wantHx)
		})
	}
	// Non-existent nodes:
	for _, goal := range []any{nil, -1, len(treeData), 1.2} {
		t.Run(fmt.Sprintf("goal=%v", goal), func(t *testing.T) {
			wantHx := ordering // Expected history.
			if len(wantHx) > 1 {
				if _, ok := goal.(int); !ok {
					wantHx = wantHx[:1]
				}
			}
			r, found := f(t, tt, goal)
			if r != 0 || found {
				t.Errorf("got <%d, %t>; want <0, false>", r, found)
			}
			checkAccessHistory(t, tt, wantHx)
		})
	}
}

func testBruteForceSearchPath(t *testing.T, name string) {
	var f func(t *testing.T, itf tree.Interface[int], initArgs ...any) []int
	var ordering []int
	switch name {
	case "DfsPath":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) []int {
			return tree.DfsPath(itf, initArgs...)
		}
		ordering = treeDataOrderingMap["dfs"]
	case "BfsPath":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) []int {
			return tree.BfsPath(itf, initArgs...)
		}
		ordering = treeDataOrderingMap["bfs"]
	case "DlsPath-0":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) []int {
			path, more := tree.DlsPath(itf, 0, initArgs...)
			if isFirstInitArgInt(initArgs) && path == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return path
		}
		ordering = treeDataOrderingMap["dls-0"]
	case "DlsPath-1":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) []int {
			path, more := tree.DlsPath(itf, 1, initArgs...)
			if isFirstInitArgInt(initArgs) && path == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return path
		}
		ordering = treeDataOrderingMap["dls-1"]
	case "DlsPath-2":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) []int {
			path, more := tree.DlsPath(itf, 2, initArgs...)
			if isFirstInitArgInt(initArgs) && path == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return path
		}
		ordering = treeDataOrderingMap["dls-2"]
	case "DlsPath-3":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) []int {
			path, _ := tree.DlsPath(itf, 3, initArgs...)
			// Both true and false are acceptable for the second return value.
			return path
		}
		ordering = treeDataOrderingMap["dls-3"]
	case "DlsPath-m1":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) []int {
			path, more := tree.DlsPath(itf, -1, initArgs...)
			if isFirstInitArgInt(initArgs) && path == nil && !more {
				t.Error("more is false but there are undiscovered vertices")
			}
			return path
		}
		ordering = treeDataOrderingMap["dls-m1"]
	case "IdsPath":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) []int {
			return tree.IdsPath(itf, 1, initArgs...)
		}
		ordering = treeDataOrderingMap["ids"]
	case "IdsPath-m":
		f = func(t *testing.T, itf tree.Interface[int], initArgs ...any) []int {
			return tree.IdsPath(itf, -1, initArgs...)
		}
		ordering = treeDataOrderingMap["ids"]
	default:
		t.Errorf("unacceptable name %q", name)
		return
	}

	tt := &treeImpl{Data: treeData}
	for goal := 0; goal < numTreeNode; goal++ {
		t.Run(fmt.Sprintf("goal=%d", goal), func(t *testing.T) {
			path := f(t, tt, goal)
			checkPath(t, name, goal, path)
			var i int
			for i < len(ordering) && ordering[i] != goal {
				i++
			}
			wantHx := ordering // Expected history.
			if i < len(ordering) {
				wantHx = wantHx[:1+i]
			}
			checkAccessHistory(t, tt, wantHx)
		})
	}
	// Non-existent nodes:
	for _, goal := range []any{nil, -1, len(treeData), 1.2} {
		t.Run(fmt.Sprintf("goal=%v", goal), func(t *testing.T) {
			wantHx := ordering // Expected history.
			goalNode := -1
			if g, ok := goal.(int); ok {
				goalNode = g
			} else if len(wantHx) > 1 {
				wantHx = wantHx[:1]
			}
			path := f(t, tt, goal)
			checkPath(t, name, goalNode, path)
			checkAccessHistory(t, tt, wantHx)
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

func checkAccessHistory(t *testing.T, tt *treeImpl, want []int) {
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

func checkPath(t *testing.T, name string, node int, pathList []int) {
	var wantPath []int
	var ordering []int
	switch name {
	case "DfsPath":
		ordering = treeDataOrderingMap["dfs"]
	case "BfsPath":
		ordering = treeDataOrderingMap["bfs"]
	case "DlsPath-0":
		ordering = treeDataOrderingMap["dls-0"]
	case "DlsPath-1":
		ordering = treeDataOrderingMap["dls-1"]
	case "DlsPath-2":
		ordering = treeDataOrderingMap["dls-2"]
	case "DlsPath-3":
		ordering = treeDataOrderingMap["dls-3"]
	case "DlsPath-m1":
		ordering = treeDataOrderingMap["dls-m1"]
	case "IdsPath", "IdsPath-m":
		ordering = treeDataOrderingMap["ids"]
	default:
		// This should never happen, but will act as a safeguard for later,
		// as a default value doesn't make sense here.
		t.Errorf("unacceptable name %q", name)
		return
	}
	for _, n := range ordering {
		if n == node {
			if node >= 0 && node < len(treeDataNodePath) {
				wantPath = treeDataNodePath[node]
			}
			break
		}
	}
	if len(pathList) != len(wantPath) {
		t.Errorf("path of %d %v;\nwant %v", node, pathList, wantPath)
		return
	}
	for i := range wantPath {
		if pathList[i] != wantPath[i] {
			t.Errorf("path of %d %v;\nwant %v", node, pathList, wantPath)
			return
		}
	}
}