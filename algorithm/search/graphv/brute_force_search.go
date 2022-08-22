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

import "github.com/donyori/gogo/algorithm/search/internal"

// Interface represents a graph used in the graph search algorithm,
// where only the vertices are concerned.
//
// It contains the basic methods required by every graph search algorithm.
type Interface[Vertex any] interface {
	// Init initializes all states for a new search
	// with specified arguments args (e.g., set the search goal).
	//
	// It will be called once at the beginning of the search functions.
	Init(args ...any)

	// Root returns the vertex of the graph where the search algorithm start.
	//
	// It returns only one vertex because the search starts at
	// one specified vertex in most cases.
	// If there are multiple starting vertices, set the root to a dummy vertex
	// whose adjacent vertices are the starting vertices.
	Root() Vertex

	// Adjacency returns the adjacent vertices of the specified vertex
	// as an adjacency list.
	//
	// The first item in the list will be accessed first.
	Adjacency(vertex Vertex) []Vertex

	// Access examines the specified vertex.
	//
	// It has two parameters:
	//  vertex - the vertex to examine;
	//  depth - the search depth from the root to the node.
	//
	// It returns two indicators:
	//  found - to report whether the specified vertex is the search goal;
	//  cont - to report whether to continue searching.
	//
	// The search algorithm should exit immediately if cont is false.
	// In this case, the search result may be invalid.
	//
	// Sometimes it is also referred to as "visit".
	Access(vertex Vertex, depth int) (found, cont bool)

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
	Discovered(vertex Vertex) bool
}

// adjListDepth consists of adjacency list and search depth.
type adjListDepth[Vertex any] struct {
	Adjacency []Vertex
	Depth     int
}

// pathListDepth consists of path list and search depth.
type pathListDepth[Vertex any] struct {
	PathList []*internal.Path[Vertex]
	Depth    int
}

// Dfs finds a vertex in itf using depth-first search algorithm
// and returns that vertex.
//
// It also returns an indicator found to report
// whether the vertex has been found.
//
// initArgs are the arguments to initialize itf.
func Dfs[Vertex any](itf Interface[Vertex], initArgs ...any) (vertexFound Vertex, found bool) {
	itf.Init(initArgs...)
	stack, idx := []adjListDepth[Vertex]{{Adjacency: []Vertex{itf.Root()}}}, 0 // Neither idx nor len(stack) is the depth.
	for idx >= 0 {
		adj, i := stack[idx].Adjacency, 0
		// Skip discovered vertices.
		for i < len(adj) && itf.Discovered(adj[i]) {
			i++
		}
		if i >= len(adj) {
			// If all vertices in this adjacency list have been discovered,
			// pop this adjacency list from the stack and go to the next loop.
			stack, idx = stack[:idx], idx-1
			continue
		}
		v, adj := adj[i], adj[1+i:]
		depth := stack[idx].Depth
		r, cont := itf.Access(v, depth)
		if r {
			return v, true
		}
		if !cont {
			return
		}
		// Following code is a simplification of the procedure:
		//  1. Pop the old adjacency list from the stack;
		//  2. Push the updated adjacency list (adj) to the stack if it is nonempty;
		//  3. Push the adjacency list of the current vertex (vAdj) to the stack if it is nonempty.
		if len(adj) > 0 {
			stack[idx].Adjacency = adj // Just update stack[idx].Adjacency.
		} else {
			stack, idx = stack[:idx], idx-1
		}
		if vAdj := itf.Adjacency(v); len(vAdj) > 0 {
			stack, idx = append(
				stack,
				adjListDepth[Vertex]{vAdj, depth + 1},
			), idx+1
		}
	}
	return
}

// DfsPath is similar to function Dfs,
// except that it returns the path from the root of itf to the vertex found
// instead of only the vertex.
//
// It returns nil if the vertex is not found.
func DfsPath[Vertex any](itf Interface[Vertex], initArgs ...any) []Vertex {
	itf.Init(initArgs...)
	// It is similar to function Dfs,
	// except that the item of the stack contains the list of Path
	// instead of the adjacency list.
	stack, idx := []pathListDepth[Vertex]{
		{PathList: []*internal.Path[Vertex]{{E: itf.Root()}}},
	}, 0
	for idx >= 0 {
		pl, i := stack[idx].PathList, 0
		for i < len(pl) && itf.Discovered(pl[i].E) {
			i++
		}
		if i >= len(pl) {
			stack, idx = stack[:idx], idx-1
			continue
		}
		p, pl := pl[i], pl[1+i:]
		depth := stack[idx].Depth
		r, cont := itf.Access(p.E, depth)
		if r {
			return p.ToList()
		}
		if !cont {
			return nil
		}
		if len(pl) > 0 {
			stack[idx].PathList = pl
		} else {
			stack, idx = stack[:idx], idx-1
		}
		if vAdj := itf.Adjacency(p.E); len(vAdj) > 0 {
			vAdjPathList := make([]*internal.Path[Vertex], len(vAdj))
			for i := range vAdj {
				vAdjPathList[i] = &internal.Path[Vertex]{vAdj[i], p}
			}
			stack, idx = append(
				stack,
				pathListDepth[Vertex]{vAdjPathList, depth + 1},
			), idx+1
		}
	}
	return nil
}

// Bfs finds a vertex in itf using breadth-first search algorithm
// and returns that vertex.
//
// It also returns an indicator found to report
// whether the vertex has been found.
//
// initArgs are the arguments to initialize itf.
func Bfs[Vertex any](itf Interface[Vertex], initArgs ...any) (vertexFound Vertex, found bool) {
	itf.Init(initArgs...)
	queue := []adjListDepth[Vertex]{{Adjacency: []Vertex{itf.Root()}}}
	for len(queue) > 0 {
		head := queue[0]
		queue = queue[1:]
		for _, vertex := range head.Adjacency {
			if !itf.Discovered(vertex) {
				r, cont := itf.Access(vertex, head.Depth)
				if r {
					return vertex, true
				}
				if !cont {
					return
				}
				if vAdj := itf.Adjacency(vertex); len(vAdj) > 0 {
					queue = append(
						queue,
						adjListDepth[Vertex]{vAdj, head.Depth + 1},
					)
				}
			}
		}
	}
	return
}

// BfsPath is similar to function Bfs,
// except that it returns the path from the root of itf to the vertex found
// instead of only the vertex.
//
// It returns nil if the vertex is not found.
func BfsPath[Vertex any](itf Interface[Vertex], initArgs ...any) []Vertex {
	itf.Init(initArgs...)
	// It is similar to function Bfs,
	// except that the item of the queue contains the list of the Path
	// instead of the adjacency list.
	queue := []pathListDepth[Vertex]{
		{PathList: []*internal.Path[Vertex]{{E: itf.Root()}}},
	}
	for len(queue) > 0 {
		head := queue[0]
		queue = queue[1:]
		for _, p := range head.PathList {
			if !itf.Discovered(p.E) {
				r, cont := itf.Access(p.E, head.Depth)
				if r {
					return p.ToList()
				}
				if !cont {
					return nil
				}
				if vAdj := itf.Adjacency(p.E); len(vAdj) > 0 {
					vAdjPathList := make([]*internal.Path[Vertex], len(vAdj))
					for i := range vAdj {
						vAdjPathList[i] = &internal.Path[Vertex]{vAdj[i], p}
					}
					queue = append(
						queue,
						pathListDepth[Vertex]{vAdjPathList, head.Depth + 1},
					)
				}
			}
		}
	}
	return nil
}

// Dls finds a vertex in itf using depth-limited depth-first search algorithm.
//
// limit is the maximum depth during this search.
// The depth of the root is 0, of adjacent vertices of the root is 1, and so on.
//
// initArgs are the arguments to initialize itf.
//
// It returns the vertex found and two indicators:
//
//	found - to report whether the vertex has been found;
//	more - to report whether there is any undiscovered vertex because of the depth limit.
//
// The indicator more makes sense only when the vertex is not found.
// When more is false, all the vertices must have been discovered.
// However, when more is true, it does not guarantee that
// there must be an undiscovered vertex,
// because the vertex may be discovered in another search path.
func Dls[Vertex any](itf Interface[Vertex], limit int, initArgs ...any) (vertexFound Vertex, found, more bool) {
	itf.Init(initArgs...)
	vertexFound, found, more, _ = dls(itf, itf.Root(), limit)
	return
}

// dls is the main body of function Dls,
// without initializing itf and acquiring the root from itf.
//
// It requires the root to avoid redundant calls to itf.Root
// in some functions such as Ids.
// The client should guarantee that root is itf.Root().
//
// It returns one more indicator quit to report whether
// itf.Access asked to stop the search
// (i.e., set its return value cont to false).
func dls[Vertex any](itf Interface[Vertex], root Vertex, limit int) (vertexFound Vertex, found, more, quit bool) {
	if limit < 0 {
		more = true // There must be an undiscovered vertex because of the depth limit: the root.
		return
	}
	// It is similar to function Dfs,
	// except that it examines the depth before pushing a new item to the stack
	// to guarantee that the depth does not exceed the limit.
	stack, idx := []adjListDepth[Vertex]{{Adjacency: []Vertex{root}}}, 0 // Neither idx nor len(stack) is the depth.
	for idx >= 0 {
		adj, i := stack[idx].Adjacency, 0
		for i < len(adj) && itf.Discovered(adj[i]) {
			i++
		}
		if i >= len(adj) {
			stack, idx = stack[:idx], idx-1
			continue
		}
		v, adj := adj[i], adj[1+i:]
		depth := stack[idx].Depth
		r, cont := itf.Access(v, depth)
		if r {
			vertexFound, found = v, true
			return
		}
		if !cont {
			quit = true
			return
		}
		if len(adj) > 0 {
			stack[idx].Adjacency = adj // Just update stack[idx].Adjacency.
		} else {
			stack, idx = stack[:idx], idx-1
		}
		if vAdj := itf.Adjacency(v); len(vAdj) > 0 {
			if depth < limit {
				// If the depth is less than the limit, push a new item.
				stack, idx = append(
					stack,
					adjListDepth[Vertex]{vAdj, depth + 1},
				), idx+1
			} else if !more {
				// If the depth reaches the limit,
				// examine whether there is any more undiscovered vertex
				// and update more.
				for _, v := range vAdj {
					if !itf.Discovered(v) {
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
// except that it returns the path from the root of itf to the vertex found
// instead of only the vertex.
//
// It returns nil for the path if the vertex is not found.
func DlsPath[Vertex any](itf Interface[Vertex], limit int, initArgs ...any) (pathFound []Vertex, more bool) {
	itf.Init(initArgs...)
	pathFound, more, _ = dlsPath(itf, itf.Root(), limit)
	return
}

// dlsPath is the main body of function DlsPath,
// without initializing itf and acquiring the root from itf.
//
// It requires the root to avoid redundant calls to itf.Root
// in some functions such as IdsPath.
// The client should guarantee that root is itf.Root().
//
// It returns one more indicator quit to report whether
// itf.Access asked to stop the search
// (i.e., set its return value cont to false).
func dlsPath[Vertex any](itf Interface[Vertex], root Vertex, limit int) (pathFound []Vertex, more, quit bool) {
	if limit < 0 {
		more = true // There must be an undiscovered vertex because of the depth limit: the root.
		return
	}
	// It is similar to function dls,
	// except that the item of the stack contains the list of Path
	// instead of the adjacency list.
	stack, idx := []pathListDepth[Vertex]{
		{PathList: []*internal.Path[Vertex]{{E: root}}},
	}, 0
	for idx >= 0 {
		pl, i := stack[idx].PathList, 0
		for i < len(pl) && itf.Discovered(pl[i].E) {
			i++
		}
		if i >= len(pl) {
			stack, idx = stack[:idx], idx-1
			continue
		}
		p, pl := pl[i], pl[1+i:]
		depth := stack[idx].Depth
		r, cont := itf.Access(p.E, depth)
		if r {
			pathFound = p.ToList()
			return
		}
		if !cont {
			quit = true
			return
		}
		if len(pl) > 0 {
			stack[idx].PathList = pl
		} else {
			stack, idx = stack[:idx], idx-1
		}
		if vAdj := itf.Adjacency(p.E); len(vAdj) > 0 {
			if depth < limit {
				vAdjPathList := make([]*internal.Path[Vertex], len(vAdj))
				for i := range vAdj {
					vAdjPathList[i] = &internal.Path[Vertex]{vAdj[i], p}
				}
				stack, idx = append(
					stack,
					pathListDepth[Vertex]{vAdjPathList, depth + 1},
				), idx+1
			} else if !more {
				for _, v := range vAdj {
					if !itf.Discovered(v) {
						more = true
						break
					}
				}
			}
		}
	}
	return
}

// IdsInterface extends interface Interface for
// functions Ids and IdsPath.
//
// It appends a new method ResetSearchState to reset
// the search state for each iteration.
type IdsInterface[Vertex any] interface {
	Interface[Vertex]

	// ResetSearchState resets the search state for the next iteration.
	//
	// It will be called before each iteration in functions Ids and IdsPath,
	// except for the first iteration.
	//
	// Its implementation must reset all the vertices to undiscovered,
	// and can reset any other states associated with each iteration
	// in this method.
	ResetSearchState()
}

// Ids finds a vertex in itf using iterative deepening depth-first
// search algorithm and returns that vertex.
//
// It also returns an indicator found to report
// whether the vertex has been found.
//
// initLimit is the depth limit used in the first iteration.
// The depth of the root is 0, of adjacent vertices of the root is 1, and so on.
// If initLimit < 1, the depth limit in the first iteration will be 1.
//
// initArgs are the arguments to initialize itf.
func Ids[Vertex any](itf IdsInterface[Vertex], initLimit int, initArgs ...any) (vertexFound Vertex, found bool) {
	itf.Init(initArgs...)
	root, limit := itf.Root(), initLimit
	if limit < 1 {
		limit = 1
	}
	for { // The loop ends when the vertex is found, more is false, or quit is true.
		vertex, r, more, quit := dls[Vertex](itf, root, limit)
		if r {
			return vertex, true
		}
		if !more || quit {
			return
		}
		itf.ResetSearchState()
		limit++
	}
}

// IdsPath is similar to function Ids,
// except that it returns the path from the root of itf to the vertex found
// instead of only the vertex.
//
// It returns nil if the vertex is not found.
func IdsPath[Vertex any](itf IdsInterface[Vertex], initLimit int, initArgs ...any) []Vertex {
	itf.Init(initArgs...)
	root, limit := itf.Root(), initLimit
	if limit < 1 {
		limit = 1
	}
	for { // The loop ends when the vertex is found, more is false, or quit is true.
		path, more, quit := dlsPath[Vertex](itf, root, limit)
		if path != nil {
			return path
		}
		if !more || quit {
			return nil
		}
		itf.ResetSearchState()
		limit++
	}
}
