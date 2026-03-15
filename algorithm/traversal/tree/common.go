// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2026  Yuan Gao
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

import (
	"sync"

	"github.com/donyori/gogo/algorithm/traversal/vseq"
	"github.com/donyori/gogo/container/queue"
	"github.com/donyori/gogo/container/set/mapset"
	"github.com/donyori/gogo/container/stack"
)

// Node represents a node in an ordered tree.
type Node interface {
	// FirstChild returns the first child of this node.
	//
	// It returns nil if this node has no children.
	FirstChild() Node

	// NextSibling returns the next sibling of this node.
	//
	// It returns nil if this node is the root or the last child of its parent.
	NextSibling() Node
}

// Options are common options for tree traversal algorithms.
type Options struct {
	// MaxStep is the maximum number of nodes to be visited.
	//
	// Nonpositive values for no limit.
	MaxStep int64

	// MaxDepth is the maximum depth during traversal,
	// where the depth is the distance from the root to the current node
	// (e.g., the depth of the root is 0,
	// the depth of a child of the root is 1, and so on).
	//
	// 0 for only visiting the root.
	// Negative values for no limit.
	MaxDepth int

	// LocalBuf indicates whether to use a local buffer (e.g., stack or queue)
	// instead of a buffer got from the global pool when needed.
	LocalBuf bool
}

// VisitorCommon is an interface common to NodeVisitor and PathVisitor.
// It groups common methods required for one tree traversal,
// including specifying the root of the tree to be traversed and
// specifying the options for the tree traversal algorithm.
type VisitorCommon interface {
	// Root returns the root of the tree to be traversed.
	//
	// If it returns nil, the traversal does nothing.
	Root() Node

	// Options returns the options for the tree traversal algorithm.
	//
	// If it returns nil, the default options are used by the traversal.
	// The default options are as follows:
	//   - MaxStep: 0
	//   - MaxDepth: -1
	//   - LocalBuf: false
	Options() *Options
}

// NodeVisitor is an interface that extends VisitorCommon
// with a method to visit each node.
//
// When visiting a node, it examines the node
// as well as the current step and depth,
// where the step is the number of nodes already visited
// (including the current node),
// and the depth is the distance from the root to the current node
// (e.g., the depth of the root is 0,
// the depth of a child of the root is 1, and so on).
type NodeVisitor interface {
	VisitorCommon

	// VisitNode examines the current node
	// as well as the current step and depth,
	// where the step is the number of nodes already visited
	// (including the current node),
	// and the depth is the distance from the root to the current node
	// (e.g., the depth of the root is 0,
	// the depth of a child of the root is 1, and so on).
	//
	// It returns two indicators cont and skipChildren,
	// where cont reports whether the traversal continues and
	// skipChildren reports whether to skip
	// the unvisited children of the current node.
	// skipChildren makes sense only when cont is true.
	VisitNode(step int64, node Node, depth int) (cont, skipChildren bool)
}

// PathVisitor is an interface that extends VisitorCommon with a method to
// visit each node as well as the path from the root to the node.
//
// When visiting a node, it examines the path from the root to the node
// as well as the current step and depth,
// where the step is the number of nodes already visited
// (including the current node),
// and the depth is the distance from the root to the current node
// (e.g., the depth of the root is 0,
// the depth of a child of the root is 1, and so on).
// The depth equals the number of nodes in the path minus 1.
type PathVisitor interface {
	VisitorCommon

	// VisitPath examines the path from the root to the current node
	// (represented by a node sequence) as well as the current step and depth,
	// where the step is the number of nodes already visited
	// (including the current node),
	// and the depth is the distance from the root to the current node
	// (e.g., the depth of the root is 0,
	// the depth of a child of the root is 1, and so on).
	// The depth equals the number of nodes in the path minus 1.
	//
	// It returns two indicators cont and skipChildren,
	// where cont reports whether the traversal continues and
	// skipChildren reports whether to skip
	// the unvisited children of the current node.
	// skipChildren makes sense only when cont is true.
	VisitPath(
		step int64,
		nodePath vseq.VertexSequence[Node],
		depth int,
	) (cont, skipChildren bool)
}

// nodeDepth consists of node and depth.
type nodeDepth struct {
	node  Node
	depth int
}

// defaultOptions are the default common options for tree traversal algorithms.
var defaultOptions = &Options{MaxDepth: -1}

var (
	// stackPool is a set of temporary stacks
	// for depth-first traversal algorithms.
	//
	// The type of the stacks is
	// [github.com/donyori/gogo/container/stack.Stack[any]].
	stackPool = sync.Pool{
		New: func() any {
			return stack.New[any](0)
		},
	}

	// queuePool is a set of temporary queues
	// for breadth-first traversal algorithms.
	//
	// The type of the queues is
	// [github.com/donyori/gogo/container/queue.Queue[any]].
	queuePool = sync.Pool{
		New: func() any {
			return queue.New[any](0)
		},
	}

	// nodeSetPool is a set of temporary tree node sets
	// for recording tree nodes in level-order traversal algorithms.
	//
	// The type of the sets is
	// [github.com/donyori/gogo/container/set.Set[Node]].
	nodeSetPool = sync.Pool{
		New: func() any {
			return mapset.New[Node](0)
		},
	}
)

// stackBasedVisitNodeFunc is the common code for traversal
// with a stack-based implementation in the iterative approach.
//
// It checks the visitor and the root, then prepares options and a stack,
// and finally calls the specified function f.
// The visitor passed to f is exactly the specified visitor.
// The root passed to f is obtained from the visitor and is not nil.
// The options passed to f must never be nil.
// The stack passed to f is empty but not nil.
func stackBasedVisitNodeFunc(
	visitor NodeVisitor,
	f func(
		visitor NodeVisitor,
		root Node,
		opts *Options,
		traversalStack stack.Stack[any],
	),
) {
	if visitor == nil || f == nil {
		return
	}

	root := visitor.Root()
	if root == nil {
		return
	}

	opts := visitor.Options()
	if opts == nil {
		opts = defaultOptions
	}

	var traversalStack stack.Stack[any]
	if opts.LocalBuf {
		traversalStack = stack.New[any](0)
	} else {
		traversalStack = stackPool.Get().(stack.Stack[any])
		defer func(s stack.Stack[any]) {
			s.RemoveAll() // avoid memory leak
			stackPool.Put(s)
		}(traversalStack)
	}

	f(visitor, root, opts, traversalStack)
}

// stackBasedVisitPathFunc is similar to function stackBasedVisitNodeFunc,
// but it processes paths from the root to current nodes,
// instead of just processing current nodes.
func stackBasedVisitPathFunc(
	visitor PathVisitor,
	f func(
		visitor PathVisitor,
		root Node,
		opts *Options,
		traversalStack stack.Stack[any],
	),
) {
	if visitor == nil || f == nil {
		return
	}

	root := visitor.Root()
	if root == nil {
		return
	}

	opts := visitor.Options()
	if opts == nil {
		opts = defaultOptions
	}

	var traversalStack stack.Stack[any]
	if opts.LocalBuf {
		traversalStack = stack.New[any](0)
	} else {
		traversalStack = stackPool.Get().(stack.Stack[any])
		defer func(s stack.Stack[any]) {
			s.RemoveAll() // avoid memory leak
			stackPool.Put(s)
		}(traversalStack)
	}

	f(visitor, root, opts, traversalStack)
}

// queueBasedVisitNodeFunc is the common code for traversal
// with a queue-based implementation in the iterative approach.
//
// It checks the visitor and the root, then prepares options and a queue,
// and finally calls the specified function f.
// The visitor passed to f is exactly the specified visitor.
// The root passed to f is obtained from the visitor and is not nil.
// The options passed to f must never be nil.
// The queue passed to f is empty but not nil.
func queueBasedVisitNodeFunc(
	visitor NodeVisitor,
	f func(
		visitor NodeVisitor,
		root Node,
		opts *Options,
		traversalQueue queue.Queue[any],
	),
) {
	if visitor == nil || f == nil {
		return
	}

	root := visitor.Root()
	if root == nil {
		return
	}

	opts := visitor.Options()
	if opts == nil {
		opts = defaultOptions
	}

	var traversalQueue queue.Queue[any]
	if opts.LocalBuf {
		traversalQueue = queue.New[any](0)
	} else {
		traversalQueue = queuePool.Get().(queue.Queue[any])
		defer func(q queue.Queue[any]) {
			q.RemoveAll() // avoid memory leak
			queuePool.Put(q)
		}(traversalQueue)
	}

	f(visitor, root, opts, traversalQueue)
}

// queueBasedVisitPathFunc is similar to function queueBasedVisitNodeFunc,
// but it processes paths from the root to current nodes,
// instead of just processing current nodes.
func queueBasedVisitPathFunc(
	visitor PathVisitor,
	f func(
		visitor PathVisitor,
		root Node,
		opts *Options,
		traversalQueue queue.Queue[any],
	),
) {
	if visitor == nil || f == nil {
		return
	}

	root := visitor.Root()
	if root == nil {
		return
	}

	opts := visitor.Options()
	if opts == nil {
		opts = defaultOptions
	}

	var traversalQueue queue.Queue[any]
	if opts.LocalBuf {
		traversalQueue = queue.New[any](0)
	} else {
		traversalQueue = queuePool.Get().(queue.Queue[any])
		defer func(q queue.Queue[any]) {
			q.RemoveAll() // avoid memory leak
			queuePool.Put(q)
		}(traversalQueue)
	}

	f(visitor, root, opts, traversalQueue)
}

// NodeSkipChildrenFunc is a function that reports whether to skip
// the unvisited children of the current node
// with the current step and depth during tree traversal.
type NodeSkipChildrenFunc func(step int64, node Node, depth int) bool

// PathSkipChildrenFunc is a function that reports whether to skip
// the unvisited children of the current node
// (i.e., the last node in the specified node path)
// with the current step and depth during tree traversal.
type PathSkipChildrenFunc func(
	step int64,
	nodePath vseq.VertexSequence[Node],
	depth int,
) bool

// visitorCommon is a trivial implementation of interface VisitorCommon.
type visitorCommon struct {
	root Node
	opts *Options
}

func (vc *visitorCommon) Root() Node {
	return vc.root
}

func (vc *visitorCommon) Options() *Options {
	return vc.opts
}

// iteratorNodeVisitor is an implementation of interface NodeVisitor
// for iterators over tree nodes.
type iteratorNodeVisitor struct {
	yield          func(Node) bool
	skipChildrenFn NodeSkipChildrenFunc
	visitorCommon
}

func (inv *iteratorNodeVisitor) VisitNode(
	step int64,
	node Node,
	depth int,
) (cont, skipChildren bool) {
	if inv.yield != nil {
		cont = inv.yield(node)
		skipChildren = inv.skipChildrenFn != nil && inv.skipChildrenFn(
			step, node, depth)
	}

	return
}

// iteratorNodeDepthVisitor is an implementation of interface NodeVisitor
// for iterators over node-depth pairs.
type iteratorNodeDepthVisitor struct {
	yield          func(Node, int) bool
	skipChildrenFn NodeSkipChildrenFunc
	visitorCommon
}

func (indv *iteratorNodeDepthVisitor) VisitNode(
	step int64,
	node Node,
	depth int,
) (cont, skipChildren bool) {
	if indv.yield != nil {
		cont = indv.yield(node, depth)
		skipChildren = indv.skipChildrenFn != nil && indv.skipChildrenFn(
			step, node, depth)
	}

	return
}

// iteratorStepNodeVisitor is an implementation of interface NodeVisitor
// for iterators over step-node pairs.
type iteratorStepNodeVisitor struct {
	yield          func(int64, Node) bool
	skipChildrenFn NodeSkipChildrenFunc
	visitorCommon
}

func (isnv *iteratorStepNodeVisitor) VisitNode(
	step int64,
	node Node,
	depth int,
) (cont, skipChildren bool) {
	if isnv.yield != nil {
		cont = isnv.yield(step, node)
		skipChildren = isnv.skipChildrenFn != nil && isnv.skipChildrenFn(
			step, node, depth)
	}

	return
}

// iteratorPathVisitor is an implementation of interface PathVisitor
// for iterators over tree node paths.
type iteratorPathVisitor struct {
	yield          func(vseq.VertexSequence[Node]) bool
	skipChildrenFn PathSkipChildrenFunc
	visitorCommon
}

func (ipv *iteratorPathVisitor) VisitPath(
	step int64,
	nodePath vseq.VertexSequence[Node],
	depth int,
) (cont, skipChildren bool) {
	if ipv.yield != nil {
		cont = ipv.yield(nodePath)
		skipChildren = ipv.skipChildrenFn != nil && ipv.skipChildrenFn(
			step, nodePath, depth)
	}

	return
}

// iteratorStepPathVisitor is an implementation of interface PathVisitor
// for iterators over step-path pairs.
type iteratorStepPathVisitor struct {
	yield          func(int64, vseq.VertexSequence[Node]) bool
	skipChildrenFn PathSkipChildrenFunc
	visitorCommon
}

func (ispv *iteratorStepPathVisitor) VisitPath(
	step int64,
	nodePath vseq.VertexSequence[Node],
	depth int,
) (cont, skipChildren bool) {
	if ispv.yield != nil {
		cont = ispv.yield(step, nodePath)
		skipChildren = ispv.skipChildrenFn != nil && ispv.skipChildrenFn(
			step, nodePath, depth)
	}

	return
}
