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

// DfsPath finds goal in itf using depth-first search algorithm,
// and returns the path from the root of itf to the goal node found.
//
// It returns nil if goal is not found.
//
// goal is only used to call the method SetGoal of itf.
// It's OK to handle goal in your implementation of Interface,
// and set goal to an arbitrary value, such as nil.
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

// BfsPath finds goal in itf using breadth-first search algorithm,
// and returns the path from the root of itf to the goal node found.
//
// It returns nil if goal is not found.
//
// goal is only used to call the method SetGoal of itf.
// It's OK to handle goal in your implementation of Interface,
// and set goal to an arbitrary value, such as nil.
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
