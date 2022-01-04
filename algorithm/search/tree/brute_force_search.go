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

import "github.com/donyori/gogo/algorithm/search/internal"

// Interface represents a tree used in the tree search algorithm.
type Interface interface {
	// Root returns the root of the tree.
	Root() interface{}

	// FirstChild returns the first child of the specified node.
	//
	// If the node has no child, FirstChild returns nil.
	FirstChild(node interface{}) interface{}

	// NextSibling returns the next sibling of the specified node,
	// i.e., the next child of the parent of the specified node.
	//
	// If the node is the root or the last child of its parent,
	// NextSibling returns nil.
	NextSibling(node interface{}) interface{}

	// SetGoal sets the search goal.
	//
	// It will be called once at the beginning of the search functions.
	//
	// Its implementation can do initialization for each search in this method.
	SetGoal(goal interface{})

	// Access examines the specified node.
	//
	// It returns an indicator found to reports whether the specified node
	// is the search goal.
	//
	// Sometimes it is also referred to as "visit".
	Access(node interface{}) (found bool)
}

// Dfs finds goal in itf using depth-first search algorithm,
// and returns the goal node found.
//
// It returns nil if goal is not found.
//
// goal is only used to call the method SetGoal of itf.
// It's OK to handle goal in your implementation of Interface,
// and set goal to an arbitrary value, such as nil.
func Dfs(itf Interface, goal interface{}) interface{} {
	itf.SetGoal(goal)
	node := itf.Root()
	if node == nil {
		return nil
	}
	stack, idx := []interface{}{node}, 0
	for idx >= 0 {
		node = stack[idx]
		if itf.Access(node) {
			return node
		}
		ns, fc := itf.NextSibling(node), itf.FirstChild(node)
		// Following code is a simplification of the procedure:
		//  1. Pop the current node from the stack;
		//  2. Push the next sibling of the current node to the stack if exists;
		//  3. Push the first child of the current node to the stack if exists.
		if ns != nil {
			stack[idx] = ns
			if fc != nil {
				stack, idx = append(stack, fc), idx+1
			}
		} else if fc != nil {
			stack[idx] = fc
		} else {
			stack, idx = stack[:idx], idx-1
		}
	}
	return nil
}

// DfsPath is similar to function Dfs,
// except that it returns the path from the root of itf
// to the goal node found, instead of only the goal node.
//
// It returns nil if goal is not found.
func DfsPath(itf Interface, goal interface{}) []interface{} {
	itf.SetGoal(goal)
	node := itf.Root()
	if node == nil {
		return nil
	}
	// It is similar to function Dfs, but the item of the stack is Path instead of the node.
	stack, idx := []*internal.Path{{X: node}}, 0
	for idx >= 0 {
		p := stack[idx]
		if itf.Access(p.X) {
			return p.ToList()
		}
		ns, fc := itf.NextSibling(p.X), itf.FirstChild(p.X)
		if ns != nil {
			stack[idx] = &internal.Path{X: ns, P: p.P} // Do not replace p.X with ns! Create a new Path.
			if fc != nil {
				stack, idx = append(stack, &internal.Path{X: fc, P: p}), idx+1
			}
		} else if fc != nil {
			stack[idx] = &internal.Path{X: fc, P: p} // Do not replace p.X and p.P with fc and p! Create a new Path.
		} else {
			stack, idx = stack[:idx], idx-1
		}
	}
	return nil
}

// Bfs finds goal in itf using breadth-first search algorithm,
// and returns the goal node found.
//
// It returns nil if goal is not found.
//
// goal is only used to call the method SetGoal of itf.
// It's OK to handle goal in your implementation of Interface,
// and set goal to an arbitrary value, such as nil.
func Bfs(itf Interface, goal interface{}) interface{} {
	itf.SetGoal(goal)
	node := itf.Root()
	if node == nil {
		return nil
	}
	queue := []interface{}{node} // A queue for the first child of each node.
	for len(queue) > 0 {
		node, queue = queue[0], queue[1:]
		for node != nil {
			if itf.Access(node) {
				return node
			}
			if fc := itf.FirstChild(node); fc != nil {
				queue = append(queue, fc)
			}
			node = itf.NextSibling(node)
		}
	}
	return nil
}

// BfsPath is similar to function Bfs,
// except that it returns the path from the root of itf
// to the goal node found, instead of only the goal node.
//
// It returns nil if goal is not found.
func BfsPath(itf Interface, goal interface{}) []interface{} {
	itf.SetGoal(goal)
	node := itf.Root()
	if node == nil {
		return nil
	}
	// It is similar to function Bfs, but the item of the queue is Path instead of the node.
	queue := []*internal.Path{{X: node}}
	for len(queue) > 0 {
		p := queue[0]
		for {
			if itf.Access(p.X) {
				return p.ToList()
			}
			if fc := itf.FirstChild(p.X); fc != nil {
				queue = append(queue, &internal.Path{X: fc, P: p})
			}
			node = itf.NextSibling(p.X)
			if node == nil {
				break
			}
			p = &internal.Path{X: node, P: p.P} // Do not replace p.X with node! Create a new Path.
		}
		queue = queue[1:]
	}
	return nil
}

// Dls finds goal in itf using depth-limited depth-first search algorithm.
//
// limit is the maximum depth during this search.
// The depth of the root is 0, of children of the root is 1, and so on.
//
// It returns the goal node found (nil if goal is not found)
// and an indicator more to report whether there is
// any undiscovered nodes because of the depth limit.
// This indicator makes sense only when the goal is not found.
// When more is false, all the nodes must have been discovered;
// when more is true, there must be at least one undiscovered node.
//
// goal is only used to call the method SetGoal of itf.
// It's OK to handle goal in your implementation of Interface,
// and set goal to an arbitrary value, such as nil.
func Dls(itf Interface, goal interface{}, limit int) (nodeFound interface{}, more bool) {
	itf.SetGoal(goal)
	root := itf.Root()
	if root == nil {
		return
	}
	return dls(itf, root, limit)
}

// dlsStackItem is the item in the stack used in function dls.
// It consists of node and search depth.
type dlsStackItem struct {
	Node  interface{}
	Depth int
}

// dls is the main body of function Dls,
// without setting the goal and acquiring the root from itf.
//
// It requires the root to avoid redundant calls to itf.Root
// in some functions such as Ids.
// The client should guarantee that root is itf.Root() and root != nil.
func dls(itf Interface, root interface{}, limit int) (nodeFound interface{}, more bool) {
	if limit < 0 {
		return nil, true // There must be an undiscovered node because of the depth limit: the root.
	}
	// It is similar to function Dfs,
	// except that it examines the depth before pushing a new item to the stack
	// to guarantee that the depth does not exceed the limit.
	stack, idx := []dlsStackItem{
		{Node: root},
	}, 0 // Neither idx nor len(stack) is the depth.
	for idx >= 0 {
		node := stack[idx].Node
		if itf.Access(node) {
			nodeFound = node
			return
		}
		cd := stack[idx].Depth + 1 // Search depth of the children of node.
		ns, fc := itf.NextSibling(node), itf.FirstChild(node)
		if ns != nil {
			stack[idx].Node = ns // Just update stack[idx].Node.
		} else {
			stack, idx = stack[:idx], idx-1
		}
		if fc != nil {
			if cd <= limit {
				// If the depth does not exceed the limit, push a new item.
				stack, idx = append(
					stack,
					dlsStackItem{Node: fc, Depth: cd},
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
// except that it returns the path from the root of itf
// to the goal node found, instead of only the goal node.
//
// It returns nil for the path if goal is not found.
func DlsPath(itf Interface, goal interface{}, limit int) (pathFound []interface{}, more bool) {
	itf.SetGoal(goal)
	root := itf.Root()
	if root == nil {
		return
	}
	return dlsPath(itf, root, limit)
}

// dlsPathStackItem is the item in the stack used in function dlsPath.
// It consists of path and search depth.
type dlsPathStackItem struct {
	Path  *internal.Path
	Depth int
}

// dlsPath is the main body of function DlsPath,
// without setting the goal and acquiring the root from itf.
//
// It requires the root to avoid redundant calls to itf.Root
// in some functions such as IdsPath.
// The client should guarantee that root is itf.Root() and root != nil.
func dlsPath(itf Interface, root interface{}, limit int) (pathFound []interface{}, more bool) {
	if limit < 0 {
		return nil, true // There must be an undiscovered node because of the depth limit: the root.
	}
	// It is similar to function dls, but the item of the stack contains
	// the Path instead of the node.
	stack, idx := []dlsPathStackItem{
		{Path: &internal.Path{X: root}},
	}, 0 // Neither idx nor len(stack) is the depth.
	for idx >= 0 {
		p := stack[idx].Path
		if itf.Access(p.X) {
			pathFound = p.ToList()
			return
		}
		cd := stack[idx].Depth + 1 // Search depth of the children of node.
		ns, fc := itf.NextSibling(p.X), itf.FirstChild(p.X)
		if ns != nil {
			stack[idx].Path = &internal.Path{X: ns, P: p.P} // Just update stack[idx].Path to a new Path.
		} else {
			stack, idx = stack[:idx], idx-1
		}
		if fc != nil {
			if cd <= limit {
				stack, idx = append(
					stack,
					dlsPathStackItem{
						Path:  &internal.Path{X: fc, P: p},
						Depth: cd,
					},
				), idx+1
			} else {
				more = true
			}
		}
	}
	return
}

// Ids finds goal in itf using iterative deepening depth-first
// search algorithm, and returns the goal node found.
//
// initLimit is the depth limit used in the first iteration.
// The depth of the root is 0, of children of the root is 1, and so on.
// If initLimit < 1, the depth limit in the first iteration will be 1.
//
// It returns nil if goal is not found.
//
// goal is only used to call the method SetGoal of itf.
// It's OK to handle goal in your implementation of Interface,
// and set goal to an arbitrary value, such as nil.
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
func Ids(itf Interface, goal interface{}, initLimit int) interface{} {
	itf.SetGoal(goal)
	root := itf.Root()
	if root == nil {
		return nil
	}
	resetter, resettable := itf.(interface{ ResetSearchState() })
	limit := initLimit
	if limit < 1 {
		limit = 1
	}
	for { // The loop ends when goal is found or more is false.
		nodeFound, more := dls(itf, root, limit)
		if nodeFound != nil {
			return nodeFound
		}
		if !more {
			return nil
		}
		if resettable {
			resetter.ResetSearchState()
		}
		limit++
	}
}

// IdsPath is similar to function Ids,
// except that it returns the path from the root of itf
// to the goal node found, instead of only the goal node.
//
// It returns nil if goal is not found.
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
func IdsPath(itf Interface, goal interface{}, initLimit int) []interface{} {
	itf.SetGoal(goal)
	root := itf.Root()
	if root == nil {
		return nil
	}
	resetter, resettable := itf.(interface{ ResetSearchState() })
	limit := initLimit
	if limit < 1 {
		limit = 1
	}
	for { // The loop ends when goal is found or more is false.
		pathFound, more := dlsPath(itf, root, limit)
		if pathFound != nil {
			return pathFound
		}
		if !more {
			return nil
		}
		if resettable {
			resetter.ResetSearchState()
		}
		limit++
	}
}
