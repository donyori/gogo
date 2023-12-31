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

package tree

import "github.com/donyori/gogo/algorithm/search/internal/pathlist"

// Common is the interface common to AccessNode and AccessPath.
// It represents a tree used in the tree search algorithm.
type Common[Node any] interface {
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
	// (If the node is the root or the last child of its parent, ok is false.)
	NextSibling(node Node) (ns Node, ok bool)
}

// AccessNode represents a tree used in the tree search algorithm.
//
// Its method AccessNode examines the current node.
type AccessNode[Node any] interface {
	Common[Node]

	// AccessNode examines the specified node.
	//
	// It has two parameters:
	//   node - the node to examine;
	//   depth - the search depth from the root to the node.
	//
	// It returns two indicators:
	//   found - to report whether the specified node is the search goal;
	//   cont - to report whether to continue searching.
	//
	// The search algorithm should exit immediately if cont is false.
	// In this case, the search result may be invalid.
	//
	// Sometimes it is also referred to as "visit".
	AccessNode(node Node, depth int) (found, cont bool)
}

// AccessPath represents a tree used in the tree search algorithm.
//
// Its method AccessPath examines the path from the search root
// to the current node.
type AccessPath[Node any] interface {
	Common[Node]

	// AccessPath examines the path from the search root to the current node.
	//
	// Its parameter is the path from the search root to
	// the current node to examine.
	// path must be non-empty.
	//
	// It returns two indicators:
	//   found - to report whether the specified node is the search goal;
	//   cont - to report whether to continue searching.
	//
	// The search algorithm should exit immediately if cont is false.
	// In this case, the search result may be invalid.
	//
	// Sometimes it is also referred to as "visit".
	AccessPath(path []Node) (found, cont bool)
}

// nodeDepth consists of node and search depth.
type nodeDepth[Node any] struct {
	node  Node
	depth int
}

// DFS finds a node in itf using depth-first search algorithm
// and returns that node.
//
// It also returns an indicator found to report whether the node has been found.
//
// initArgs are the arguments to initialize itf.
func DFS[Node any](itf AccessNode[Node], initArgs ...any) (
	nodeFound Node, found bool) {
	itf.Init(initArgs...)
	stack := []nodeDepth[Node]{{node: itf.Root()}} // len(stack) is not the depth
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		r, cont := itf.AccessNode(top.node, top.depth)
		if r {
			return top.node, true
		} else if !cont {
			return
		}
		// Following code is a simplification of the procedure:
		//   1. Pop the current node from the stack;
		//   2. Push the next sibling of the current node
		//      to the stack if exists;
		//   3. Push the first child of the current node
		//      to the stack if exists.
		ns, ok := itf.NextSibling(top.node)
		if ok {
			stack[len(stack)-1].node = ns // just update stack[len(stack)-1].node
		} else {
			stack = stack[:len(stack)-1]
		}
		fc, ok := itf.FirstChild(top.node)
		if ok {
			stack = append(stack, nodeDepth[Node]{
				node:  fc,
				depth: top.depth + 1,
			})
		}
	}
	return
}

// DFSPath is similar to function DFS,
// except that it returns the path from the root of itf to the node found
// instead of only the node.
//
// It returns nil if the node is not found.
func DFSPath[Node any](itf AccessPath[Node], initArgs ...any) []Node {
	itf.Init(initArgs...)
	// It is similar to function DFS,
	// except that the item of the stack contains the Path instead of the node.
	stack := []*pathlist.Path[Node]{{E: itf.Root()}}
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		pathList := top.ToList()
		r, cont := itf.AccessPath(pathList)
		if r {
			return pathList
		} else if !cont {
			return nil
		}
		ns, ok := itf.NextSibling(top.E)
		if ok {
			// Just update stack[len(stack)-1] to a new Path.
			// Do not modify stack[len(stack)-1]! Create a new Path.
			stack[len(stack)-1] = &pathlist.Path[Node]{E: ns, P: top.P}
		} else {
			stack = stack[:len(stack)-1]
		}
		fc, ok := itf.FirstChild(top.E)
		if ok {
			stack = append(stack, &pathlist.Path[Node]{
				E: fc,
				P: top,
			})
		}
	}
	return nil
}

// BFS finds a node in itf using breadth-first search algorithm
// and returns that node.
//
// It also returns an indicator found to report whether the node has been found.
//
// initArgs are the arguments to initialize itf.
func BFS[Node any](itf AccessNode[Node], initArgs ...any) (
	nodeFound Node, found bool) {
	itf.Init(initArgs...)
	queue := []nodeDepth[Node]{{node: itf.Root()}} // a queue for the first child of each node
	for len(queue) > 0 {
		head := queue[0]
		queue = queue[1:]
		for node, ok := head.node, true; ok; node, ok = itf.NextSibling(node) {
			r, cont := itf.AccessNode(node, head.depth)
			if r {
				return node, true
			} else if !cont {
				return
			} else if fc, ok := itf.FirstChild(node); ok {
				queue = append(queue, nodeDepth[Node]{
					node:  fc,
					depth: head.depth + 1,
				})
			}
		}
	}
	return
}

// BFSPath is similar to function BFS,
// except that it returns the path from the root of itf to the node found
// instead of only the node.
//
// It returns nil if the node is not found.
func BFSPath[Node any](itf AccessPath[Node], initArgs ...any) []Node {
	itf.Init(initArgs...)
	// It is similar to function BFS,
	// except that the item of the queue contains the Path instead of the node.
	queue := []*pathlist.Path[Node]{{E: itf.Root()}}
	for len(queue) > 0 {
		head := queue[0]
		queue = queue[1:]
		for node, ok := head.E, true; ok; node, ok = itf.NextSibling(node) {
			// The path to head (in the first loop)
			// or one of its siblings (in other loops).
			path := &pathlist.Path[Node]{E: node, P: head.P}
			pathList := path.ToList()
			r, cont := itf.AccessPath(pathList)
			if r {
				return pathList
			} else if !cont {
				return nil
			} else if fc, ok := itf.FirstChild(node); ok {
				queue = append(queue, &pathlist.Path[Node]{E: fc, P: path})
			}
		}
	}
	return nil
}

// DLS finds a node in itf using depth-limited depth-first search algorithm.
//
// limit is the maximum depth during this search.
// The depth of the root is 0, of children of the root is 1, and so on.
//
// initArgs are the arguments to initialize itf.
//
// It returns the node found and two indicators:
//
//	found - to report whether the node has been found;
//	more - to report whether there is any undiscovered node because of the depth limit.
//
// The indicator more makes sense only when the node is not found.
// When more is false, all the nodes must have been discovered;
// when more is true, there must be at least one undiscovered node.
func DLS[Node any](itf AccessNode[Node], limit int, initArgs ...any) (
	nodeFound Node, found, more bool) {
	itf.Init(initArgs...)
	nodeFound, found, more, _ = dls(itf, itf.Root(), limit)
	return
}

// dls is the main body of function DLS,
// without initializing itf and acquiring the root from itf.
//
// It requires the root to avoid redundant calls to itf.Root
// in some functions such as IDS.
// The client should guarantee that root is itf.Root().
//
// It returns one more indicator quit to report whether
// itf.AccessNode asked to stop the search
// (i.e., set its return value cont to false).
func dls[Node any](itf AccessNode[Node], root Node, limit int) (
	nodeFound Node, found, more, quit bool) {
	if limit < 0 {
		more = true // there must be an undiscovered node because of the depth limit: the root
		return
	}
	// It is similar to function DFS,
	// except that it examines the depth before pushing a new item to the stack
	// to guarantee that the depth does not exceed the limit.
	stack := []nodeDepth[Node]{{node: root}} // len(stack) is not the depth
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		r, cont := itf.AccessNode(top.node, top.depth)
		if r {
			nodeFound, found = top.node, true
			return
		} else if !cont {
			quit = true
			return
		}
		ns, ok := itf.NextSibling(top.node)
		if ok {
			stack[len(stack)-1].node = ns // just update stack[len(stack)-1].node
		} else {
			stack = stack[:len(stack)-1]
		}
		fc, ok := itf.FirstChild(top.node)
		if ok {
			if top.depth < limit {
				// If the depth does not exceed the limit, push a new item.
				stack = append(stack, nodeDepth[Node]{
					node:  fc,
					depth: top.depth + 1,
				})
			} else {
				// If the depth of the child exceeds the limit,
				// set more to true.
				more = true
			}
		}
	}
	return
}

// DLSPath is similar to function DLS,
// except that it returns the path from the root of itf to the node found
// instead of only the node.
//
// It returns nil for the path if the node is not found.
func DLSPath[Node any](itf AccessPath[Node], limit int, initArgs ...any) (
	pathFound []Node, more bool) {
	itf.Init(initArgs...)
	pathFound, more, _ = dlsPath(itf, itf.Root(), limit)
	return
}

// dlsPath is the main body of function DLSPath,
// without initializing itf and acquiring the root from itf.
//
// It requires the root to avoid redundant calls to itf.Root
// in some functions such as IDSPath.
// The client should guarantee that root is itf.Root().
//
// It returns one more indicator quit to report whether
// itf.AccessPath asked to stop the search
// (i.e., set its return value cont to false).
func dlsPath[Node any](itf AccessPath[Node], root Node, limit int) (
	pathFound []Node, more, quit bool) {
	if limit < 0 {
		more = true // there must be an undiscovered node because of the depth limit: the root
		return
	}
	// It is similar to function dls,
	// except that the item of the stack contains the Path instead of the node.
	stack := []*pathlist.Path[Node]{{E: root}}
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		pathList := top.ToList()
		r, cont := itf.AccessPath(pathList)
		if r {
			pathFound = pathList
			return
		} else if !cont {
			quit = true
			return
		}
		ns, ok := itf.NextSibling(top.E)
		if ok {
			// Just update stack[len(stack)-1] to a new Path.
			// Do not modify stack[len(stack)-1]! Create a new Path.
			stack[len(stack)-1] = &pathlist.Path[Node]{E: ns, P: top.P}
		} else {
			stack = stack[:len(stack)-1]
		}
		fc, ok := itf.FirstChild(top.E)
		if ok {
			if len(pathList) <= limit {
				stack = append(stack, &pathlist.Path[Node]{
					E: fc,
					P: top,
				})
			} else {
				more = true
			}
		}
	}
	return
}

// IDS finds a node in itf using iterative deepening depth-first
// search algorithm and returns that node.
//
// It also returns an indicator found to report whether the node has been found.
//
// initLimit is the depth limit used in the first iteration.
// The depth of the root is 0, of children of the root is 1, and so on.
// If initLimit < 1, the depth limit in the first iteration is 1.
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
//	ResetSearchState()
//
// And the client should define this method like
//
//	func (m MyInterface) ResetSearchState() {
//		// Reset your search state.
//	}
func IDS[Node any](
	itf AccessNode[Node],
	initLimit int,
	initArgs ...any,
) (nodeFound Node, found bool) {
	itf.Init(initArgs...)
	root := itf.Root()
	resetter, resettable := (any)(itf).(interface{ ResetSearchState() })
	limit := initLimit
	if limit < 1 {
		limit = 1
	}
	for { // the loop ends when the node is found, more is false, or quit is true
		node, r, more, quit := dls(itf, root, limit)
		switch {
		case r:
			return node, true
		case !more || quit:
			return
		case resettable:
			resetter.ResetSearchState()
		}
		limit++
	}
}

// IDSPath is similar to function IDS,
// except that it returns the path from the root of itf to the node found
// instead of only the node.
//
// It returns nil if the node is not found.
//
// Same as function IDS, if the client needs to reset any search state
// at the beginning of each iteration,
// just add the method ResetSearchState to itf.
// This method will be called before each iteration except for the first one.
//
// The method signature should be
//
//	ResetSearchState()
//
// And the client should define this method like
//
//	func (m MyInterface) ResetSearchState() {
//		// Reset your search state.
//	}
func IDSPath[Node any](
	itf AccessPath[Node],
	initLimit int,
	initArgs ...any,
) []Node {
	itf.Init(initArgs...)
	root := itf.Root()
	resetter, resettable := (any)(itf).(interface{ ResetSearchState() })
	limit := initLimit
	if limit < 1 {
		limit = 1
	}
	for { // the loop ends when the node is found, more is false, or quit is true
		path, more, quit := dlsPath(itf, root, limit)
		switch {
		case path != nil:
			return path
		case !more || quit:
			return nil
		case resettable:
			resetter.ResetSearchState()
		}
		limit++
	}
}
