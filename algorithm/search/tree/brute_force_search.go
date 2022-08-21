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

package tree

import "github.com/donyori/gogo/algorithm/search/internal"

// Interface represents a tree used in the tree search algorithm.
type Interface[Node any] interface {
	// Init initializes all states for a new search
	// with the specified arguments args (e.g., set the search goal).
	//
	// It will be called once at the beginning of the search functions.
	Init(args ...any)

	// Root returns the node of the tree where the search algorithm start.
	Root() Node

	// FirstChild returns the first child of the specified node
	// and an indicator ok to report whether the node has a child.
	FirstChild(node Node) (fc Node, ok bool)

	// NextSibling returns the next sibling of the specified node
	// (i.e., the next child of the parent of the specified node)
	// and an indicator ok to report whether the node has the next sibling.
	// (If the node is the root or the last child of its parent,
	// ok will be false.)
	NextSibling(node Node) (ns Node, ok bool)

	// Access examines the specified node.
	//
	// It has two parameters:
	//  node - the node to examine;
	//  depth - the search depth from the root to the node.
	//
	// It returns two indicators:
	//  found - to report whether the specified node is the search goal;
	//  cont - to report whether to continue searching.
	//
	// The search algorithm should exit immediately if cont is false.
	// In this case, the search result may be invalid.
	//
	// Sometimes it is also referred to as "visit".
	Access(node Node, depth int) (found, cont bool)
}

// nodeDepth consists of node and search depth.
type nodeDepth[Node any] struct {
	Node  Node
	Depth int
}

// pathDepth consists of path and search depth.
type pathDepth[Node any] struct {
	Path  *internal.Path[Node]
	Depth int
}

// Dfs finds a node in itf using depth-first search algorithm
// and returns that node.
//
// It also returns an indicator found to report whether the node has been found.
//
// initArgs are the arguments to initialize itf.
func Dfs[Node any](itf Interface[Node], initArgs ...any) (nodeFound Node, found bool) {
	itf.Init(initArgs...)
	stack, idx := []nodeDepth[Node]{{Node: itf.Root()}}, 0 // Neither idx nor len(stack) is the depth.
	for idx >= 0 {
		top := stack[idx]
		r, cont := itf.Access(top.Node, top.Depth)
		if r {
			return top.Node, true
		}
		if !cont {
			return
		}
		// Following code is a simplification of the procedure:
		//  1. Pop the current node from the stack;
		//  2. Push the next sibling of the current node to the stack if exists;
		//  3. Push the first child of the current node to the stack if exists.
		ns, ok := itf.NextSibling(top.Node)
		if ok {
			stack[idx].Node = ns // Just update stack[idx].Node.
		} else {
			stack, idx = stack[:idx], idx-1
		}
		fc, ok := itf.FirstChild(top.Node)
		if ok {
			stack, idx = append(
				stack,
				nodeDepth[Node]{fc, top.Depth + 1},
			), idx+1
		}
	}
	return
}

// DfsPath is similar to function Dfs,
// except that it returns the path from the root of itf to the node found
// instead of only the node.
//
// It returns nil if the node is not found.
func DfsPath[Node any](itf Interface[Node], initArgs ...any) []Node {
	itf.Init(initArgs...)
	// It is similar to function Dfs,
	// except that the item of the stack contains the Path instead of the node.
	stack, idx := []pathDepth[Node]{
		{Path: &internal.Path[Node]{E: itf.Root()}},
	}, 0
	for idx >= 0 {
		top := stack[idx]
		r, cont := itf.Access(top.Path.E, top.Depth)
		if r {
			return top.Path.ToList()
		}
		if !cont {
			return nil
		}
		ns, ok := itf.NextSibling(top.Path.E)
		if ok {
			// Just update stack[idx].Path to a new Path.
			// Do not modify stack[idx].Path! Create a new Path.
			stack[idx].Path = &internal.Path[Node]{ns, top.Path.P}
		} else {
			stack, idx = stack[:idx], idx-1
		}
		fc, ok := itf.FirstChild(top.Path.E)
		if ok {
			stack, idx = append(
				stack,
				pathDepth[Node]{
					&internal.Path[Node]{fc, top.Path},
					top.Depth + 1,
				},
			), idx+1
		}
	}
	return nil
}

// Bfs finds a node in itf using breadth-first search algorithm
// and returns that node.
//
// It also returns an indicator found to report whether the node has been found.
//
// initArgs are the arguments to initialize itf.
func Bfs[Node any](itf Interface[Node], initArgs ...any) (nodeFound Node, found bool) {
	itf.Init(initArgs...)
	queue := []nodeDepth[Node]{{Node: itf.Root()}} // A queue for the first child of each node.
	for len(queue) > 0 {
		head := queue[0]
		queue = queue[1:]
		for node, ok := head.Node, true; ok; node, ok = itf.NextSibling(node) {
			r, cont := itf.Access(node, head.Depth)
			if r {
				return node, true
			}
			if !cont {
				return
			}
			if fc, ok := itf.FirstChild(node); ok {
				queue = append(queue, nodeDepth[Node]{fc, head.Depth + 1})
			}
		}
	}
	return
}

// BfsPath is similar to function Bfs,
// except that it returns the path from the root of itf to the node found
// instead of only the node.
//
// It returns nil if the node is not found.
func BfsPath[Node any](itf Interface[Node], initArgs ...any) []Node {
	itf.Init(initArgs...)
	// It is similar to function Bfs,
	// except that the item of the queue contains the Path instead of the node.
	queue := []pathDepth[Node]{{Path: &internal.Path[Node]{E: itf.Root()}}}
	for len(queue) > 0 {
		head := queue[0]
		queue = queue[1:]
		for node, ok := head.Path.E, true; ok; node, ok = itf.NextSibling(node) {
			r, cont := itf.Access(node, head.Depth)
			if r {
				// The path to head (in the first loop) or one of its siblings (in other loops).
				path := &internal.Path[Node]{node, head.Path.P}
				return path.ToList()
			}
			if !cont {
				return nil
			}
			if fc, ok := itf.FirstChild(node); ok {
				queue = append(
					queue,
					pathDepth[Node]{
						&internal.Path[Node]{
							fc,
							&internal.Path[Node]{node, head.Path.P},
						},
						head.Depth + 1,
					},
				)
			}
		}
	}
	return nil
}

// Dls finds a node in itf using depth-limited depth-first search algorithm.
//
// limit is the maximum depth during this search.
// The depth of the root is 0, of children of the root is 1, and so on.
//
// initArgs are the arguments to initialize itf.
//
// It returns the node found and two indicators:
//  found - to report whether the node has been found;
//  more - to report whether there is any undiscovered node because of the depth limit.
//
// The indicator more makes sense only when the node is not found.
// When more is false, all the nodes must have been discovered;
// when more is true, there must be at least one undiscovered node.
func Dls[Node any](itf Interface[Node], limit int, initArgs ...any) (nodeFound Node, found, more bool) {
	itf.Init(initArgs...)
	nodeFound, found, more, _ = dls(itf, itf.Root(), limit)
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
func dls[Node any](itf Interface[Node], root Node, limit int) (nodeFound Node, found, more, quit bool) {
	if limit < 0 {
		more = true // There must be an undiscovered node because of the depth limit: the root.
		return
	}
	// It is similar to function Dfs,
	// except that it examines the depth before pushing a new item to the stack
	// to guarantee that the depth does not exceed the limit.
	stack, idx := []nodeDepth[Node]{{Node: root}}, 0 // Neither idx nor len(stack) is the depth.
	for idx >= 0 {
		top := stack[idx]
		r, cont := itf.Access(top.Node, top.Depth)
		if r {
			nodeFound, found = top.Node, true
			return
		}
		if !cont {
			quit = true
			return
		}
		ns, ok := itf.NextSibling(top.Node)
		if ok {
			stack[idx].Node = ns // Just update stack[idx].Node.
		} else {
			stack, idx = stack[:idx], idx-1
		}
		fc, ok := itf.FirstChild(top.Node)
		if ok {
			if top.Depth < limit {
				// If the depth does not exceed the limit, push a new item.
				stack, idx = append(
					stack,
					nodeDepth[Node]{fc, top.Depth + 1},
				), idx+1
			} else {
				// If the depth of the child exceeds the limit,
				// set more to true.
				more = true
			}
		}
	}
	return
}

// DlsPath is similar to function Dls,
// except that it returns the path from the root of itf to the node found
// instead of only the node.
//
// It returns nil for the path if the node is not found.
func DlsPath[Node any](itf Interface[Node], limit int, initArgs ...any) (pathFound []Node, more bool) {
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
func dlsPath[Node any](itf Interface[Node], root Node, limit int) (pathFound []Node, more, quit bool) {
	if limit < 0 {
		more = true // There must be an undiscovered node because of the depth limit: the root.
		return
	}
	// It is similar to function dls,
	// except that the item of the stack contains the Path instead of the node.
	stack, idx := []pathDepth[Node]{{Path: &internal.Path[Node]{E: root}}}, 0
	for idx >= 0 {
		top := stack[idx]
		r, cont := itf.Access(top.Path.E, top.Depth)
		if r {
			pathFound = top.Path.ToList()
			return
		}
		if !cont {
			quit = true
			return
		}
		ns, ok := itf.NextSibling(top.Path.E)
		if ok {
			// Just update stack[idx].Path to a new Path.
			// Do not modify stack[idx].Path! Create a new Path.
			stack[idx].Path = &internal.Path[Node]{ns, top.Path.P}
		} else {
			stack, idx = stack[:idx], idx-1
		}
		fc, ok := itf.FirstChild(top.Path.E)
		if ok {
			if top.Depth < limit {
				stack, idx = append(
					stack,
					pathDepth[Node]{
						&internal.Path[Node]{fc, top.Path},
						top.Depth + 1,
					},
				), idx+1
			} else {
				more = true
			}
		}
	}
	return
}

// Ids finds a node in itf using iterative deepening depth-first
// search algorithm and returns that node.
//
// It also returns an indicator found to report whether the node has been found.
//
// initLimit is the depth limit used in the first iteration.
// The depth of the root is 0, of children of the root is 1, and so on.
// If initLimit < 1, the depth limit in the first iteration will be 1.
//
// initArgs are the arguments to initialize itf.
//
// If the client needs to reset any search state
// at the beginning of each iteration,
// just add the method ResetSearchState to itf.
// This method will be called before each iteration except for the first one.
//
// The method signature should be
//
//  ResetSearchState()
//
// And the client should define this method like
//
//  func (m MyInterface) ResetSearchState() {
//  	// Reset your search state.
//  }
func Ids[Node any](itf Interface[Node], initLimit int, initArgs ...any) (nodeFound Node, found bool) {
	itf.Init(initArgs...)
	root := itf.Root()
	resetter, resettable := (any)(itf).(interface{ ResetSearchState() })
	limit := initLimit
	if limit < 1 {
		limit = 1
	}
	for { // The loop ends when the node is found, more is false, or quit is true.
		node, r, more, quit := dls(itf, root, limit)
		if r {
			return node, true
		}
		if !more || quit {
			return
		}
		if resettable {
			resetter.ResetSearchState()
		}
		limit++
	}
}

// IdsPath is similar to function Ids,
// except that it returns the path from the root of itf to the node found
// instead of only the node.
//
// It returns nil if the node is not found.
//
// Same as function Ids, if the client needs to reset any search state
// at the beginning of each iteration,
// just add the method ResetSearchState to itf.
// This method will be called before each iteration except for the first one.
//
// The method signature should be
//
//  ResetSearchState()
//
// And the client should define this method like
//
//  func (m MyInterface) ResetSearchState() {
//  	// Reset your search state.
//  }
func IdsPath[Node any](itf Interface[Node], initLimit int, initArgs ...any) []Node {
	itf.Init(initArgs...)
	root := itf.Root()
	resetter, resettable := (any)(itf).(interface{ ResetSearchState() })
	limit := initLimit
	if limit < 1 {
		limit = 1
	}
	for { // The loop ends when the node is found, more is false, or quit is true.
		path, more, quit := dlsPath(itf, root, limit)
		if path != nil {
			return path
		}
		if !more || quit {
			return nil
		}
		if resettable {
			resetter.ResetSearchState()
		}
		limit++
	}
}