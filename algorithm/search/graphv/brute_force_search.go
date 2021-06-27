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

import "github.com/donyori/gogo/algorithm/search/internal"

// BasicInterface represents a graph used in the graph search algorithm,
// where only the vertices are concerned.
//
// It contains the basic methods required by every graph search algorithm.
type BasicInterface interface {
	// Root returns the vertex of the graph where the search algorithms start.
	//
	// It returns only one vertex because the search starts at
	// one specified vertex in most cases.
	// If there are multiple starting vertices, set the root to a dummy vertex
	// whose adjacent vertices are the starting vertices.
	Root() interface{}

	// Adjacency returns the adjacent vertices of the specified vertex
	// as an adjacency list.
	//
	// The first item in the list will be accessed first.
	Adjacency(vertex interface{}) []interface{}

	// SetGoal sets the search goal.
	//
	// It will be called once at the beginning of the search functions.
	//
	// Its implementation can do initialization for each search in this method.
	SetGoal(goal interface{})

	// Access examines the specified vertex.
	//
	// It returns an indicator found to reports whether the specified vertex
	// is the search goal.
	//
	// Sometimes it is also referred to as "visit".
	Access(vertex interface{}) (found bool)

	// Discovered reports whether the specified vertex
	// has been examined by the method Access.
	//
	// Sometimes it is also referred to as "visited".
	//
	// It is typically implemented by using a hash table or
	// associating an attribute "discovered" to each vertex.
	//
	// If it is not necessary to record the vertices discovered,
	// or the graph is too big to record the vertices discovered,
	// just always return false.
	// In this case, the search algorithms may access one vertex
	// multiple times, and may even fall into an infinite loop.
	Discovered(vertex interface{}) bool
}

// Dfs finds goal in itf using depth-first search algorithm,
// and returns the goal vertex found.
//
// It returns nil if goal is not found.
//
// goal is only used to call the method SetGoal of itf.
// It's OK to handle goal in your implementation of BasicInterface,
// and set goal to an arbitrary value, such as nil.
func Dfs(itf BasicInterface, goal interface{}) interface{} {
	itf.SetGoal(goal)
	v := itf.Root()
	if v == nil {
		return nil
	}
	stack, idx := [][]interface{}{{v}}, 0
	for idx >= 0 {
		adj, i := stack[idx], 0
		// Skip nil and discovered vertices.
		for i < len(adj) && (adj[i] == nil || itf.Discovered(adj[i])) {
			i++
		}
		if i >= len(adj) {
			// If all vertices in this adjacency list are nil or discovered,
			// pop this adjacency list from the stack and go to the next loop.
			stack, idx = stack[:idx], idx-1
			continue
		}
		v, adj = adj[i], adj[1+i:]
		if itf.Access(v) {
			return v
		}
		vAdj := itf.Adjacency(v)
		// Following code is a simplification of the procedure:
		//  1. Pop the old adjacency list from the stack;
		//  2. Push the updated adjacency list (adj) to the stack if it is nonempty;
		//  3. Push the adjacency list of the current vertex (vAdj) to the stack if it is nonempty.
		if len(adj) > 0 {
			stack[idx] = adj
			if len(vAdj) > 0 {
				stack, idx = append(stack, vAdj), idx+1
			}
		} else if len(vAdj) > 0 {
			stack[idx] = vAdj
		} else {
			stack, idx = stack[:idx], idx-1
		}
	}
	return nil
}

// DfsPath is similar to function Dfs,
// except that it returns the path from the root of itf
// to the goal vertex found, instead of only the goal vertex.
//
// It returns nil if goal is not found.
func DfsPath(itf BasicInterface, goal interface{}) []interface{} {
	itf.SetGoal(goal)
	v := itf.Root()
	if v == nil {
		return nil
	}
	// It is similar to function Dfs, but the item of the stack is
	// the list of Path instead of the adjacency list.
	stack, idx := [][]*internal.Path{{{X: v}}}, 0
	for idx >= 0 {
		pl, i := stack[idx], 0
		// Unlike in function Dfs, pl[i].X here must be non-nil.
		for i < len(pl) && itf.Discovered(pl[i].X) {
			i++
		}
		if i >= len(pl) {
			stack, idx = stack[:idx], idx-1
			continue
		}
		p, pl := pl[i], pl[1+i:]
		if itf.Access(p.X) {
			return p.ToList()
		}
		vAdj := itf.Adjacency(p.X)
		var vAdjPathList []*internal.Path
		if len(vAdj) > 0 {
			vAdjPathList = make([]*internal.Path, 0, len(vAdj))
			for _, x := range vAdj {
				if x != nil {
					vAdjPathList = append(vAdjPathList, &internal.Path{X: x, P: p})
				}
			}
		}
		if len(pl) > 0 {
			stack[idx] = pl
			if len(vAdjPathList) > 0 {
				stack, idx = append(stack, vAdjPathList), idx+1
			}
		} else if len(vAdjPathList) > 0 {
			stack[idx] = vAdjPathList
		} else {
			stack, idx = stack[:idx], idx-1
		}
	}
	return nil
}

// Bfs finds goal in itf using breadth-first search algorithm,
// and returns the goal vertex found.
//
// It returns nil if goal is not found.
//
// goal is only used to call the method SetGoal of itf.
// It's OK to handle goal in your implementation of BasicInterface,
// and set goal to an arbitrary value, such as nil.
func Bfs(itf BasicInterface, goal interface{}) interface{} {
	itf.SetGoal(goal)
	v := itf.Root()
	if v == nil {
		return nil
	}
	queue := [][]interface{}{{v}}
	for len(queue) > 0 {
		for _, v = range queue[0] {
			if v == nil || itf.Discovered(v) {
				// Skip nil and discovered vertices.
				continue
			}
			if itf.Access(v) {
				return v
			}
			if vAdj := itf.Adjacency(v); len(vAdj) > 0 {
				queue = append(queue, vAdj)
			}
		}
		queue = queue[1:]
	}
	return nil
}

// BfsPath is similar to function Bfs,
// except that it returns the path from the root of itf
// to the goal vertex found, instead of only the goal vertex.
//
// It returns nil if goal is not found.
func BfsPath(itf BasicInterface, goal interface{}) []interface{} {
	itf.SetGoal(goal)
	v := itf.Root()
	if v == nil {
		return nil
	}
	// It is similar to function Bfs, but the item of the queue is
	// the list of Path instead of the adjacency list.
	queue := [][]*internal.Path{{{X: v}}}
	for len(queue) > 0 {
		for _, p := range queue[0] {
			// Unlike in function Bfs, p.X here must be non-nil.
			if itf.Discovered(p.X) {
				continue
			}
			if itf.Access(p.X) {
				return p.ToList()
			}
			if vAdj := itf.Adjacency(p.X); len(vAdj) > 0 {
				vAdjPathList := make([]*internal.Path, 0, len(vAdj))
				for _, x := range vAdj {
					if x != nil {
						vAdjPathList = append(vAdjPathList, &internal.Path{X: x, P: p})
					}
				}
				if len(vAdjPathList) > 0 {
					queue = append(queue, vAdjPathList)
				}
			}
		}
		queue = queue[1:]
	}
	return nil
}

// Dls finds goal in itf using depth-limited depth-first search algorithm.
//
// limit is the maximum depth during this search.
// The depth of the root is 0, of adjacent vertices of the root is 1, and so on.
//
// It returns the goal vertex found (nil if goal is not found)
// and an indicator more to report whether there may be
// any undiscovered vertices because of the depth limit.
// This indicator makes sense only when the goal is not found.
// When more is false, all the vertices must have been discovered.
// However, when more is true, it does not guarantee that
// there must be an undiscovered vertex,
// because the vertex may be discovered in another search path.
//
// goal is only used to call the method SetGoal of itf.
// It's OK to handle goal in your implementation of BasicInterface,
// and set goal to an arbitrary value, such as nil.
func Dls(itf BasicInterface, goal interface{}, limit int) (vertexFound interface{}, more bool) {
	itf.SetGoal(goal)
	root := itf.Root()
	if root == nil {
		return
	}
	return dls(itf, root, limit)
}

// dlsStackItem is the item in the stack used in function dls.
// It consists of adjacency list and search depth.
type dlsStackItem struct {
	Adjacency []interface{}
	Depth     int
}

// dls is the main body of function Dls,
// without setting the goal and acquiring the root from itf.
//
// It requires the root to avoid redundant calls to itf.Root
// in some functions such as Ids.
// The client should guarantee that root is itf.Root() and root != nil.
func dls(itf BasicInterface, root interface{}, limit int) (vertexFound interface{}, more bool) {
	if limit < 0 {
		return nil, true // There must be an undiscovered vertex because of the depth limit: the root.
	}
	// It is similar to function Dfs,
	// except that it examines the depth before pushing a new item to the stack
	// to guarantee that the depth does not exceed the limit.
	stack, idx := []dlsStackItem{{Adjacency: []interface{}{root}}}, 0 // Neither idx nor len(stack) is the depth.
	for idx >= 0 {
		adj, i := stack[idx].Adjacency, 0
		for i < len(adj) && (adj[i] == nil || itf.Discovered(adj[i])) {
			i++
		}
		if i >= len(adj) {
			stack, idx = stack[:idx], idx-1
			continue
		}
		v, adj := adj[i], adj[1+i:]
		if itf.Access(v) {
			vertexFound = v
			return
		}
		depth := stack[idx].Depth
		if len(adj) > 0 {
			stack[idx].Adjacency = adj // Just update stack[idx].Adjacency.
		} else {
			stack, idx = stack[:idx], idx-1
		}
		vAdj := itf.Adjacency(v)
		if len(vAdj) > 0 {
			if depth < limit {
				// If the depth is less than the limit, push a new item.
				stack, idx = append(stack, dlsStackItem{Adjacency: vAdj, Depth: depth + 1}), idx+1
			} else if !more {
				// If the depth reaches the limit,
				// examine whether there is any more undiscovered vertex
				// and update more.
				for _, a := range vAdj {
					if a != nil && !itf.Discovered(a) {
						more = true
						break
					}
				}
			}
		}
	}
	return
}

// DlsPath is similar to function Dls,
// except that it returns the path from the root of itf
// to the goal vertex found, instead of only the goal vertex.
//
// It returns nil for the path if goal is not found.
func DlsPath(itf BasicInterface, goal interface{}, limit int) (pathFound []interface{}, more bool) {
	itf.SetGoal(goal)
	root := itf.Root()
	if root == nil {
		return
	}
	return dlsPath(itf, root, limit)
}

// dlsPathStackItem is the item in the stack used in function dlsPath.
// It consists of path list and search depth.
type dlsPathStackItem struct {
	PathList []*internal.Path
	Depth    int
}

// dlsPath is the main body of function DlsPath,
// without setting the goal and acquiring the root from itf.
//
// It requires the root to avoid redundant calls to itf.Root
// in some functions such as IdsPath.
// The client should guarantee that root is itf.Root() and root != nil.
func dlsPath(itf BasicInterface, root interface{}, limit int) (pathFound []interface{}, more bool) {
	if limit < 0 {
		return nil, true // There must be an undiscovered vertex because of the depth limit: the root.
	}
	// It is similar to function dls, but the item of the stack contains
	// the list of Path instead of the adjacency list.
	stack, idx := []dlsPathStackItem{{PathList: []*internal.Path{{X: root}}}}, 0 // Neither idx nor len(stack) is the depth.
	for idx >= 0 {
		pl, i := stack[idx].PathList, 0
		// Unlike in function dls, pl[i].X here must be non-nil.
		for i < len(pl) && itf.Discovered(pl[i].X) {
			i++
		}
		if i >= len(pl) {
			stack, idx = stack[:idx], idx-1
			continue
		}
		p, pl := pl[i], pl[1+i:]
		if itf.Access(p.X) {
			pathFound = p.ToList()
			return
		}
		depth := stack[idx].Depth
		if len(pl) > 0 {
			stack[idx].PathList = pl
		} else {
			stack, idx = stack[:idx], idx-1
		}
		vAdj := itf.Adjacency(p.X)
		var vAdjPathList []*internal.Path
		if len(vAdj) > 0 {
			vAdjPathList = make([]*internal.Path, 0, len(vAdj))
			for _, x := range vAdj {
				if x != nil {
					vAdjPathList = append(vAdjPathList, &internal.Path{X: x, P: p})
				}
			}
		}
		if len(vAdjPathList) > 0 {
			if depth < limit {
				stack, idx = append(stack, dlsPathStackItem{PathList: vAdjPathList, Depth: depth + 1}), idx+1
			} else if !more {
				for _, a := range vAdjPathList {
					if !itf.Discovered(a.X) {
						more = true
						break
					}
				}
			}
		}
	}
	return
}

// IdsInterface extends interface BasicInterface for
// functions Ids and IdsPath.
//
// It contains a new method ResetSearchState to reset
// the search state for each iteration.
type IdsInterface interface {
	BasicInterface

	// ResetSearchState resets the search state for the next iteration.
	//
	// It will be called before each iteration in function Ids,
	// except for the first iteration.
	//
	// Its implementation must reset all the vertices to undiscovered,
	// and can reset any other states associated with each iteration
	// in this method.
	ResetSearchState()
}

// Ids finds goal in itf using iterative deepening depth-first
// search algorithm, and returns the goal vertex found.
//
// initLimit is the depth limit used in the first iteration.
// The depth of the root is 0, of adjacent vertices of the root is 1, and so on.
// If initLimit < 1, the depth limit in the first iteration will be 1.
//
// It returns nil if goal is not found.
//
// goal is only used to call the method SetGoal of itf.
// It's OK to handle goal in your implementation of IdsInterface,
// and set goal to an arbitrary value, such as nil.
func Ids(itf IdsInterface, goal interface{}, initLimit int) interface{} {
	itf.SetGoal(goal)
	root := itf.Root()
	if root == nil {
		return nil
	}
	limit := initLimit
	if limit < 1 {
		limit = 1
	}
	for { // The loop ends when goal is found or more is false.
		vertexFound, more := dls(itf, root, limit)
		if vertexFound != nil {
			return vertexFound
		} else if !more {
			return nil
		}
		itf.ResetSearchState()
		limit++
	}
}

// IdsPath is similar to function Ids,
// except that it returns the path from the root of itf
// to the goal vertex found, instead of only the goal vertex.
//
// It returns nil if goal is not found.
func IdsPath(itf IdsInterface, goal interface{}, initLimit int) []interface{} {
	itf.SetGoal(goal)
	root := itf.Root()
	if root == nil {
		return nil
	}
	limit := initLimit
	if limit < 1 {
		limit = 1
	}
	for { // The loop ends when goal is found or more is false.
		pathFound, more := dlsPath(itf, root, limit)
		if pathFound != nil {
			return pathFound
		} else if !more {
			return nil
		}
		itf.ResetSearchState()
		limit++
	}
}
