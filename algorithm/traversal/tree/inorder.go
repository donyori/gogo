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
	"iter"

	"github.com/donyori/gogo/algorithm/traversal/vseq"
	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/container/stack"
)

// InorderNode extends interface Node for in-order traversal.
// It adds a method to report whether
// the node should be visited before its parent.
type InorderNode interface {
	Node

	// Left reports whether this node belongs to
	// the "left" children of its parent.
	// If true, this node should be visited before its parent;
	// otherwise, after its parent.
	//
	// All "left" children must come before
	// "right" children in the ordered tree.
	//
	// Left doesn't make sense for the root.
	Left() bool
}

// left reports whether the specified node belongs to
// the "left" children of its parent, as follows:
//   - If the node is nil, left returns false.
//   - If the node is of type InorderNode, left calls its method Left.
//   - Otherwise, left returns true if and only if the node has a next sibling.
func left(node Node) bool {
	if node == nil {
		return false
	} else if inorderNode, ok := node.(InorderNode); ok {
		return inorderNode.Left()
	}

	return node.NextSibling() != nil
}

// Inorder executes in-order traversal for the specified visitor.
// In-order traversal means that for each node,
// recursively traverse the "left" subtrees of the node from first to last,
// then visit the node, and finally
// recursively traverse the "right" subtrees of the node from first to last.
//
// The "left" and "right" of a node can be specified
// by implementing interface InorderNode.
// If a node does not implement interface InorderNode,
// it is treated as "left" if and only if it has a next sibling.
//
// The traversal uses a stack-based implementation in the iterative approach.
//
// If the specified visitor is nil or the root returned by the visitor is nil,
// Inorder does nothing.
//
// The return value skipChildren of the visitor's method VisitNode
// only affects the "right" children for in-order traversal.
func Inorder(visitor NodeVisitor) {
	stackBasedVisitNodeFunc(visitor, stackBasedInorderMain)
}

// stackBasedInorderMain is the main body of Inorder
// that executes in-order traversal with a stack-based implementation
// in the iterative approach.
func stackBasedInorderMain(
	visitor NodeVisitor,
	root Node,
	opts *Options,
	traversalStack stack.Stack[any],
) {
	var (
		step     int64
		lastPath array.SliceDynamicArray[Node]
	)

	traversalStack.Push(nodeDepth{node: root})

	for traversalStack.Len() > 0 {
		top := traversalStack.Pop().(nodeDepth)

		// (top.depth != opts.MaxDepth) is equivalent to
		// (opts.MaxDepth < 0 || top.depth < opts.MaxDepth)
		// because the depth starts from 0 and increases by 1 each time.
		pushTopAndFirstLeftChild := top.depth != opts.MaxDepth &&
			(top.depth >= lastPath.Len()-1 ||
				top.node != lastPath.Get(top.depth))
		stackBasedInorderAndReverseInorderUpdateLastPath(
			&lastPath,
			top.node,
			top.depth,
		)

		if pushTopAndFirstLeftChild {
			if fc := top.node.FirstChild(); left(fc) { // left checks the nilness
				traversalStack.Push(top) // the current node is pushed again
				traversalStack.Push(nodeDepth{node: fc, depth: top.depth + 1})

				continue
			}
		}

		step++
		cont, skipChildren := visitor.VisitNode(step, top.node, top.depth)

		// (step == opts.MaxStep) is equivalent to
		// (opts.MaxStep > 0 && step >= opts.MaxStep)
		// because the step starts from 1 and increases by 1 each time.
		if !cont || step == opts.MaxStep {
			return
		}

		stackBasedInorderPushNextSiblingAndFirstRightChild(
			opts.MaxDepth,
			traversalStack,
			&top,
			skipChildren,
		)
	}
}

// stackBasedInorderPushNextSiblingAndFirstRightChild pushes
// the next sibling and the first "right" child of the current node
// onto the stack in the main loop of stackBasedInorderMain.
func stackBasedInorderPushNextSiblingAndFirstRightChild(
	maxDepth int,
	traversalStack stack.Stack[any],
	pTop *nodeDepth,
	skipChildren bool,
) {
	// The next sibling is pushed after visiting the current node.
	//
	// Consider siblings only when the current node is not the root
	// (equivalently, pTop.depth > 0) to allow that
	// the root returned by the visitor is the root of a subtree.
	if pTop.depth > 0 {
		ns := pTop.node.NextSibling()

		// The next sibling is pushed only when
		// both it and the current node are "left" or "right".
		if ns != nil && left(ns) == left(pTop.node) {
			traversalStack.Push(nodeDepth{node: ns, depth: pTop.depth})
		}
	}

	// The first "right" child is pushed after visiting the current node
	// and pushing the next sibling, unless it is skipped.
	//
	// (pTop.depth != maxDepth) is equivalent to
	// (maxDepth < 0 || pTop.depth < maxDepth)
	// because the depth starts from 0 and increases by 1 each time.
	if !skipChildren && pTop.depth != maxDepth {
		c := pTop.node.FirstChild()

		// Fast forward to the first "right" child.
		for left(c) { // left checks the nilness
			c = c.NextSibling()
		}

		if c != nil {
			traversalStack.Push(nodeDepth{node: c, depth: pTop.depth + 1})
		}
	}
}

// InorderPath is similar to function Inorder,
// but it processes paths from the root to current nodes,
// instead of just processing current nodes.
func InorderPath(visitor PathVisitor) {
	stackBasedVisitPathFunc(visitor, stackBasedInorderPathMain)
}

// stackBasedInorderPathMain is the main body of InorderPath
// that executes in-order traversal with a stack-based implementation
// in the iterative approach.
func stackBasedInorderPathMain(
	visitor PathVisitor,
	root Node,
	opts *Options,
	traversalStack stack.Stack[any],
) {
	var (
		step     int64
		lastPath array.SliceDynamicArray[Node]
	)

	traversalStack.Push(vseq.NewVertexSequence(root))

	for traversalStack.Len() > 0 {
		top := traversalStack.Pop().(vseq.VertexSequence[Node])
		topDepth := top.Len() - 1
		topNode := top.Back()

		pushTopAndFirstLeftChild := topDepth != opts.MaxDepth &&
			(topDepth >= lastPath.Len()-1 || topNode != lastPath.Get(topDepth))
		stackBasedInorderAndReverseInorderUpdateLastPath(
			&lastPath,
			topNode,
			topDepth,
		)

		if pushTopAndFirstLeftChild {
			if fc := topNode.FirstChild(); left(fc) {
				traversalStack.Push(top)
				traversalStack.Push(top.Append(fc))

				continue
			}
		}

		step++
		cont, skipChildren := visitor.VisitPath(step, top, topDepth)

		if !cont || step == opts.MaxStep {
			return
		}

		stackBasedInorderPathPushNextSiblingAndFirstRightChild(
			opts.MaxDepth,
			traversalStack,
			top,
			skipChildren,
		)
	}
}

// stackBasedInorderPathPushNextSiblingAndFirstRightChild pushes
// the paths from the root to the next sibling and the first "right" child
// of the current node onto the stack
// in the main loop of stackBasedInorderPathMain.
func stackBasedInorderPathPushNextSiblingAndFirstRightChild(
	maxDepth int,
	traversalStack stack.Stack[any],
	top vseq.VertexSequence[Node],
	skipChildren bool,
) {
	topDepth := top.Len() - 1
	topNode := top.Back()

	if topDepth > 0 {
		if ns := topNode.NextSibling(); ns != nil && left(ns) == left(topNode) {
			traversalStack.Push(top.ReplaceBack(ns))
		}
	}

	if !skipChildren && topDepth != maxDepth {
		c := topNode.FirstChild()
		for left(c) {
			c = c.NextSibling()
		}

		if c != nil {
			traversalStack.Push(top.Append(c))
		}
	}
}

// ReverseInorder executes reverse in-order traversal for the specified visitor.
// Reverse in-order traversal means that for each node,
// recursively traverse the "right" subtrees of the node from last to first,
// then visit the node, and finally
// recursively traverse the "left" subtrees of the node from last to first.
//
// The "left" and "right" of a node can be specified
// by implementing interface InorderNode.
// If a node does not implement interface InorderNode,
// it is treated as "left" if and only if it has a next sibling.
//
// The traversal uses a stack-based implementation in the iterative approach.
//
// If the specified visitor is nil or the root returned by the visitor is nil,
// ReverseInorder does nothing.
//
// The return value skipChildren of the visitor's method VisitNode
// only affects the "left" children for reverse in-order traversal.
func ReverseInorder(visitor NodeVisitor) {
	stackBasedVisitNodeFunc(visitor, stackBasedReverseInorderMain)
}

// stackBasedReverseInorderMain is the main body of ReverseInorder
// that executes reverse in-order traversal with a stack-based implementation
// in the iterative approach.
func stackBasedReverseInorderMain(
	visitor NodeVisitor,
	root Node,
	opts *Options,
	traversalStack stack.Stack[any],
) {
	var (
		step     int64
		lastPath array.SliceDynamicArray[Node]
	)

	traversalStack.Push(nodeDepth{node: root})

	for traversalStack.Len() > 0 {
		top := traversalStack.Pop().(nodeDepth)

		// (top.depth != opts.MaxDepth) is equivalent to
		// (opts.MaxDepth < 0 || top.depth < opts.MaxDepth)
		// because the depth starts from 0 and increases by 1 each time.
		pushTopAndRightChildren := top.depth != opts.MaxDepth &&
			(top.depth >= lastPath.Len()-1 ||
				top.node != lastPath.Get(top.depth))
		stackBasedInorderAndReverseInorderUpdateLastPath(
			&lastPath,
			top.node,
			top.depth,
		)

		if pushTopAndRightChildren &&
			stackBasedReverseInorderPushTopAndRightChildren(
				traversalStack,
				&top,
			) {
			continue
		}

		step++
		cont, skipChildren := visitor.VisitNode(step, top.node, top.depth)

		// (step == opts.MaxStep) is equivalent to
		// (opts.MaxStep > 0 && step >= opts.MaxStep)
		// because the step starts from 1 and increases by 1 each time.
		if !cont || step == opts.MaxStep {
			return
		}

		// The "left" children are pushed after visiting the current node,
		// unless they are skipped.
		if !skipChildren && top.depth != opts.MaxDepth {
			// The "left" children are pushed from first to last
			// so that they are visited from last to first.
			for c := top.node.FirstChild(); left(c); c = c.NextSibling() { // left checks the nilness
				traversalStack.Push(nodeDepth{node: c, depth: top.depth + 1})
			}
		}
	}
}

// stackBasedReverseInorderPushTopAndRightChildren pushes
// the current node and its "right" children onto the stack
// in the main loop of stackBasedReverseInorderMain,
// and reports whether to skip visiting the current node.
func stackBasedReverseInorderPushTopAndRightChildren(
	traversalStack stack.Stack[any],
	pTop *nodeDepth,
) (skipVisitingNode bool) {
	c := pTop.node.FirstChild()

	// Fast forward to the first "right" child.
	for left(c) { // left checks the nilness
		c = c.NextSibling()
	}

	// If there are any "right" children (i.e., (c != nil) now),
	// the current node and "right" children are pushed.
	// Otherwise, the current node is visited.
	if c != nil {
		traversalStack.Push(*pTop) // the current node is pushed again

		// The "right" children are pushed from first to last
		// so that they are visited from last to first.
		for c != nil {
			traversalStack.Push(nodeDepth{node: c, depth: pTop.depth + 1})
			c = c.NextSibling()
		}

		skipVisitingNode = true
	}

	return
}

// ReverseInorderPath is similar to function ReverseInorder,
// but it processes paths from the root to current nodes,
// instead of just processing current nodes.
func ReverseInorderPath(visitor PathVisitor) {
	stackBasedVisitPathFunc(visitor, stackBasedReverseInorderPathMain)
}

// stackBasedReverseInorderPathMain is the main body of ReverseInorderPath
// that executes reverse in-order traversal with a stack-based implementation
// in the iterative approach.
func stackBasedReverseInorderPathMain(
	visitor PathVisitor,
	root Node,
	opts *Options,
	traversalStack stack.Stack[any],
) {
	var (
		step     int64
		lastPath array.SliceDynamicArray[Node]
	)

	traversalStack.Push(vseq.NewVertexSequence(root))

	for traversalStack.Len() > 0 {
		top := traversalStack.Pop().(vseq.VertexSequence[Node])
		topDepth := top.Len() - 1
		topNode := top.Back()

		pushTopAndRightChildren := topDepth != opts.MaxDepth &&
			(topDepth >= lastPath.Len()-1 || topNode != lastPath.Get(topDepth))
		stackBasedInorderAndReverseInorderUpdateLastPath(
			&lastPath,
			topNode,
			topDepth,
		)

		if pushTopAndRightChildren &&
			stackBasedReverseInorderPathPushTopAndRightChildren(
				traversalStack,
				top,
			) {
			continue
		}

		step++
		cont, skipChildren := visitor.VisitPath(step, top, topDepth)

		if !cont || step == opts.MaxStep {
			return
		}

		if !skipChildren && topDepth != opts.MaxDepth {
			for c := topNode.FirstChild(); left(c); c = c.NextSibling() {
				traversalStack.Push(top.Append(c))
			}
		}
	}
}

// stackBasedReverseInorderPathPushTopAndRightChildren pushes
// the paths from the root to the current node and its "right" children
// onto the stack in the main loop of stackBasedReverseInorderPathMain,
// and reports whether to skip visiting the current path.
func stackBasedReverseInorderPathPushTopAndRightChildren(
	traversalStack stack.Stack[any],
	top vseq.VertexSequence[Node],
) (skipVisitingPath bool) {
	c := top.Back().FirstChild()
	for left(c) {
		c = c.NextSibling()
	}

	if c != nil {
		traversalStack.Push(top)

		for c != nil {
			traversalStack.Push(top.Append(c))
			c = c.NextSibling()
		}

		skipVisitingPath = true
	}

	return
}

// stackBasedInorderAndReverseInorderUpdateLastPath updates the last path
// in stack-based implementations for in-order and reverse in-order traversal.
func stackBasedInorderAndReverseInorderUpdateLastPath(
	lastPath array.DynamicArray[Node],
	topNode Node,
	topDepth int,
) {
	if topDepth < lastPath.Len() {
		lastPath.Truncate(topDepth + 1)
		lastPath.SetBack(topNode)
	} else {
		lastPath.Push(topNode)
	}
}

// IterInorder returns an iterator over nodes
// of the tree with the specified root,
// traversing it in in-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the "right" children of the current node.
// If it is nil, no nodes will be skipped.
//
// The "left" and "right" of a node can be specified
// by implementing interface InorderNode.
// If a node does not implement interface InorderNode,
// it is treated as "left" if and only if it has a next sibling.
//
// The returned iterator is never nil.
func IterInorder(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		if yield != nil && root != nil {
			Inorder(&iteratorNodeVisitor{
				yield:          yield,
				skipChildrenFn: skipChildrenFn,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterInorderDepth returns an iterator over node-depth pairs
// of the tree with the specified root,
// traversing it in in-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the "right" children of the current node.
// If it is nil, no nodes will be skipped.
//
// The "left" and "right" of a node can be specified
// by implementing interface InorderNode.
// If a node does not implement interface InorderNode,
// it is treated as "left" if and only if it has a next sibling.
//
// The returned iterator is never nil.
func IterInorderDepth(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq2[Node, int] {
	return func(yield func(Node, int) bool) {
		if yield != nil && root != nil {
			Inorder(&iteratorNodeDepthVisitor{
				yield:          yield,
				skipChildrenFn: skipChildrenFn,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterInorderStep returns an iterator over step-node pairs
// of the tree with the specified root,
// traversing it in in-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the "right" children of the current node.
// If it is nil, no nodes will be skipped.
//
// The "left" and "right" of a node can be specified
// by implementing interface InorderNode.
// If a node does not implement interface InorderNode,
// it is treated as "left" if and only if it has a next sibling.
//
// The returned iterator is never nil.
func IterInorderStep(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq2[int64, Node] {
	return func(yield func(int64, Node) bool) {
		if yield != nil && root != nil {
			Inorder(&iteratorStepNodeVisitor{
				yield:          yield,
				skipChildrenFn: skipChildrenFn,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterInorderPath returns an iterator over paths
// from the root to nodes of the tree with the specified root,
// traversing it in in-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the "right" children of the current node.
// If it is nil, no nodes will be skipped.
//
// The "left" and "right" of a node can be specified
// by implementing interface InorderNode.
// If a node does not implement interface InorderNode,
// it is treated as "left" if and only if it has a next sibling.
//
// The returned iterator is never nil.
func IterInorderPath(
	root Node,
	opts *Options,
	skipChildrenFn PathSkipChildrenFunc,
) iter.Seq[vseq.VertexSequence[Node]] {
	return func(yield func(vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			InorderPath(&iteratorPathVisitor{
				yield:          yield,
				skipChildrenFn: skipChildrenFn,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterInorderPathStep returns an iterator over step-path pairs where
// the paths are from the root to nodes of the tree with the specified root,
// traversing it in in-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the "right" children of the current node.
// If it is nil, no nodes will be skipped.
//
// The "left" and "right" of a node can be specified
// by implementing interface InorderNode.
// If a node does not implement interface InorderNode,
// it is treated as "left" if and only if it has a next sibling.
//
// The returned iterator is never nil.
func IterInorderPathStep(
	root Node,
	opts *Options,
	skipChildrenFn PathSkipChildrenFunc,
) iter.Seq2[int64, vseq.VertexSequence[Node]] {
	return func(yield func(int64, vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			InorderPath(&iteratorStepPathVisitor{
				yield:          yield,
				skipChildrenFn: skipChildrenFn,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterReverseInorder returns an iterator over nodes
// of the tree with the specified root,
// traversing it in reverse in-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the "left" children of the current node.
// If it is nil, no nodes will be skipped.
//
// The "left" and "right" of a node can be specified
// by implementing interface InorderNode.
// If a node does not implement interface InorderNode,
// it is treated as "left" if and only if it has a next sibling.
//
// The returned iterator is never nil.
func IterReverseInorder(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		if yield != nil && root != nil {
			ReverseInorder(&iteratorNodeVisitor{
				yield:          yield,
				skipChildrenFn: skipChildrenFn,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterReverseInorderDepth returns an iterator over node-depth pairs
// of the tree with the specified root,
// traversing it in reverse in-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the "left" children of the current node.
// If it is nil, no nodes will be skipped.
//
// The "left" and "right" of a node can be specified
// by implementing interface InorderNode.
// If a node does not implement interface InorderNode,
// it is treated as "left" if and only if it has a next sibling.
//
// The returned iterator is never nil.
func IterReverseInorderDepth(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq2[Node, int] {
	return func(yield func(Node, int) bool) {
		if yield != nil && root != nil {
			ReverseInorder(&iteratorNodeDepthVisitor{
				yield:          yield,
				skipChildrenFn: skipChildrenFn,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterReverseInorderStep returns an iterator over step-node pairs
// of the tree with the specified root,
// traversing it in reverse in-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the "left" children of the current node.
// If it is nil, no nodes will be skipped.
//
// The "left" and "right" of a node can be specified
// by implementing interface InorderNode.
// If a node does not implement interface InorderNode,
// it is treated as "left" if and only if it has a next sibling.
//
// The returned iterator is never nil.
func IterReverseInorderStep(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq2[int64, Node] {
	return func(yield func(int64, Node) bool) {
		if yield != nil && root != nil {
			ReverseInorder(&iteratorStepNodeVisitor{
				yield:          yield,
				skipChildrenFn: skipChildrenFn,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterReverseInorderPath returns an iterator over paths
// from the root to nodes of the tree with the specified root,
// traversing it in reverse in-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the "left" children of the current node.
// If it is nil, no nodes will be skipped.
//
// The "left" and "right" of a node can be specified
// by implementing interface InorderNode.
// If a node does not implement interface InorderNode,
// it is treated as "left" if and only if it has a next sibling.
//
// The returned iterator is never nil.
func IterReverseInorderPath(
	root Node,
	opts *Options,
	skipChildrenFn PathSkipChildrenFunc,
) iter.Seq[vseq.VertexSequence[Node]] {
	return func(yield func(vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			ReverseInorderPath(&iteratorPathVisitor{
				yield:          yield,
				skipChildrenFn: skipChildrenFn,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterReverseInorderPathStep returns an iterator over step-path pairs where
// the paths are from the root to nodes of the tree with the specified root,
// traversing it in reverse in-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the "left" children of the current node.
// If it is nil, no nodes will be skipped.
//
// The "left" and "right" of a node can be specified
// by implementing interface InorderNode.
// If a node does not implement interface InorderNode,
// it is treated as "left" if and only if it has a next sibling.
//
// The returned iterator is never nil.
func IterReverseInorderPathStep(
	root Node,
	opts *Options,
	skipChildrenFn PathSkipChildrenFunc,
) iter.Seq2[int64, vseq.VertexSequence[Node]] {
	return func(yield func(int64, vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			ReverseInorderPath(&iteratorStepPathVisitor{
				yield:          yield,
				skipChildrenFn: skipChildrenFn,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}
