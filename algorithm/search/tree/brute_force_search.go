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
	// It will be called once at the beginning of the search function.
	SetTarget(target interface{})

	// Visit handles the specified node.
	//
	// It returns an indicator found to reports whether the specified node
	// is the search target.
	Visit(node interface{}) (found bool)
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
		if itf.Visit(node) {
			return node
		}
		ns, fc := itf.NextSibling(node), itf.FirstChild(node)
		if ns != nil {
			stack[idx] = ns
			if fc != nil {
				stack, idx = append(stack, fc), idx+1
			}
		} else {
			if fc != nil {
				stack[idx] = fc
			} else {
				stack, idx = stack[:idx], idx-1
			}
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
			if itf.Visit(node) {
				return node
			}
			if n := itf.FirstChild(node); n != nil {
				queue = append(queue, n)
			}
			node = itf.NextSibling(node)
		}
	}
	return nil
}
