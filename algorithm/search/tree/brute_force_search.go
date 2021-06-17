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

	// SetTarget sets the search target.
	//
	// It will be called once at the beginning of the search functions.
	SetTarget(target interface{})

	// Access examines the specified node.
	//
	// It returns an indicator found to reports whether the specified node
	// is the search target.
	Access(node interface{}) (found bool)
}

// Dfs finds target in itf using depth-first search algorithm,
// and returns the target node found.
//
// It returns nil if target is not found.
//
// target is only used to call the method SetTarget of itf.
// It's OK to handle target in your implementation of Interface,
// and set target to an arbitrary value, such as nil.
func Dfs(itf Interface, target interface{}) interface{} {
	itf.SetTarget(target)
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

// Bfs finds target in itf using breadth-first search algorithm,
// and returns the target node found.
//
// It returns nil if target is not found.
//
// target is only used to call the method SetTarget of itf.
// It's OK to handle target in your implementation of Interface,
// and set target to an arbitrary value, such as nil.
func Bfs(itf Interface, target interface{}) interface{} {
	itf.SetTarget(target)
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

// DfsPath finds target in itf using depth-first search algorithm,
// and returns the path from the root of itf to the target node found.
//
// It returns nil if target is not found.
//
// target is only used to call the method SetTarget of itf.
// It's OK to handle target in your implementation of Interface,
// and set target to an arbitrary value, such as nil.
func DfsPath(itf Interface, target interface{}) []interface{} {
	itf.SetTarget(target)
	node := itf.Root()
	if node == nil {
		return nil
	}
	// It is similar to function Dfs, but the item of the stack is NodePath instead of the node.
	np := &internal.NodePath{N: node}
	stack, idx := []*internal.NodePath{np}, 0
	for idx >= 0 {
		np = stack[idx]
		if itf.Access(np.N) {
			return np.ToList()
		}
		ns, fc := itf.NextSibling(np.N), itf.FirstChild(np.N)
		if ns != nil {
			stack[idx] = &internal.NodePath{N: ns, P: np.P} // Do not replace np.N with ns! Create a new NodePath.
			if fc != nil {
				stack, idx = append(stack, &internal.NodePath{N: fc, P: np}), idx+1
			}
		} else if fc != nil {
			stack[idx] = &internal.NodePath{N: fc, P: np} // Do not replace np.N and np.P with fc and np! Create a new NodePath.
		} else {
			stack, idx = stack[:idx], idx-1
		}
	}
	return nil
}

// BfsPath finds target in itf using breadth-first search algorithm,
// and returns the path from the root of itf to the target node found.
//
// It returns nil if target is not found.
//
// target is only used to call the method SetTarget of itf.
// It's OK to handle target in your implementation of Interface,
// and set target to an arbitrary value, such as nil.
func BfsPath(itf Interface, target interface{}) []interface{} {
	itf.SetTarget(target)
	node := itf.Root()
	if node == nil {
		return nil
	}
	// It is similar to function Bfs, but the item of the queue is NodePath instead of the node.
	np := &internal.NodePath{N: node}
	queue := []*internal.NodePath{np}
	for len(queue) > 0 {
		np, queue = queue[0], queue[1:]
		for {
			if itf.Access(np.N) {
				return np.ToList()
			}
			if fc := itf.FirstChild(np.N); fc != nil {
				queue = append(queue, &internal.NodePath{N: fc, P: np})
			}
			node = itf.NextSibling(np.N)
			if node == nil {
				break
			}
			np = &internal.NodePath{N: node, P: np.P} // Do not replace np.N with node! Create a new NodePath.
		}
	}
	return nil
}
