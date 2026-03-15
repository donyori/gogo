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
	"github.com/donyori/gogo/container/stack"
)

// Preorder executes pre-order traversal for the specified visitor.
// Pre-order traversal means that for each node, visit the node first,
// and then recursively traverse the subtrees of the node from first to last.
//
// The traversal uses a stack-based implementation in the iterative approach.
//
// If the specified visitor is nil or the root returned by the visitor is nil,
// Preorder does nothing.
func Preorder(visitor NodeVisitor) {
	stackBasedVisitNodeFunc(visitor, stackBasedPreorderMain)
}

// stackBasedPreorderMain is the main body of Preorder
// that executes pre-order traversal with a stack-based implementation
// in the iterative approach.
func stackBasedPreorderMain(
	visitor NodeVisitor,
	root Node,
	opts *Options,
	traversalStack stack.Stack[any],
) {
	var step int64

	traversalStack.Push(nodeDepth{node: root})

	for traversalStack.Len() > 0 {
		top := traversalStack.Pop().(nodeDepth)

		step++
		cont, skipChildren := visitor.VisitNode(step, top.node, top.depth)

		// (step == opts.MaxStep) is equivalent to
		// (opts.MaxStep > 0 && step >= opts.MaxStep)
		// because the step starts from 1 and increases by 1 each time.
		if !cont || step == opts.MaxStep {
			return
		}

		// The next sibling is pushed first
		// so that the children are visited first.
		//
		// Consider siblings only when the current node is not the root
		// (equivalently, top.depth > 0) to allow that
		// the root returned by the visitor is the root of a subtree.
		if top.depth > 0 {
			if ns := top.node.NextSibling(); ns != nil {
				traversalStack.Push(nodeDepth{node: ns, depth: top.depth})
			}
		}

		// (top.depth != opts.MaxDepth) is equivalent to
		// (opts.MaxDepth < 0 || top.depth < opts.MaxDepth)
		// because the depth starts from 0 and increases by 1 each time.
		if !skipChildren && top.depth != opts.MaxDepth {
			if fc := top.node.FirstChild(); fc != nil {
				traversalStack.Push(nodeDepth{node: fc, depth: top.depth + 1})
			}
		}
	}
}

// PreorderPath is similar to function Preorder,
// but it processes paths from the root to current nodes,
// instead of just processing current nodes.
func PreorderPath(visitor PathVisitor) {
	stackBasedVisitPathFunc(visitor, stackBasedPreorderPathMain)
}

// stackBasedPreorderPathMain is the main body of PreorderPath
// that executes pre-order traversal with a stack-based implementation
// in the iterative approach.
func stackBasedPreorderPathMain(
	visitor PathVisitor,
	root Node,
	opts *Options,
	traversalStack stack.Stack[any],
) {
	var step int64

	traversalStack.Push(vseq.NewVertexSequence(root))

	for traversalStack.Len() > 0 {
		top := traversalStack.Pop().(vseq.VertexSequence[Node])
		topDepth := top.Len() - 1

		step++
		cont, skipChildren := visitor.VisitPath(step, top, topDepth)

		if !cont || step == opts.MaxStep {
			return
		}

		if topDepth > 0 {
			if ns := top.Back().NextSibling(); ns != nil {
				traversalStack.Push(top.ReplaceBack(ns))
			}
		}

		if !skipChildren && topDepth != opts.MaxDepth {
			if fc := top.Back().FirstChild(); fc != nil {
				traversalStack.Push(top.Append(fc))
			}
		}
	}
}

// ReversePreorder executes reverse pre-order traversal
// for the specified visitor.
// Reverse pre-order traversal means that for each node, visit the node first,
// and then recursively traverse the subtrees of the node from last to first.
//
// The traversal uses a stack-based implementation in the iterative approach.
//
// If the specified visitor is nil or the root returned by the visitor is nil,
// ReversePreorder does nothing.
func ReversePreorder(visitor NodeVisitor) {
	stackBasedVisitNodeFunc(visitor, stackBasedReversePreorderMain)
}

// stackBasedReversePreorderMain is the main body of ReversePreorder
// that executes reverse pre-order traversal with a stack-based implementation
// in the iterative approach.
func stackBasedReversePreorderMain(
	visitor NodeVisitor,
	root Node,
	opts *Options,
	traversalStack stack.Stack[any],
) {
	var step int64

	traversalStack.Push(nodeDepth{node: root})

	for traversalStack.Len() > 0 {
		top := traversalStack.Pop().(nodeDepth)

		step++
		cont, skipChildren := visitor.VisitNode(step, top.node, top.depth)

		// (step == opts.MaxStep) is equivalent to
		// (opts.MaxStep > 0 && step >= opts.MaxStep)
		// because the step starts from 1 and increases by 1 each time.
		if !cont || step == opts.MaxStep {
			return
		}

		// (top.depth != opts.MaxDepth) is equivalent to
		// (opts.MaxDepth < 0 || top.depth < opts.MaxDepth)
		// because the depth starts from 0 and increases by 1 each time.
		if !skipChildren && top.depth != opts.MaxDepth {
			// The children are pushed from first to last
			// so that they are visited from last to first.
			for c := top.node.FirstChild(); c != nil; c = c.NextSibling() {
				traversalStack.Push(nodeDepth{node: c, depth: top.depth + 1})
			}
		}
	}
}

// ReversePreorderPath is similar to function ReversePreorder,
// but it processes paths from the root to current nodes,
// instead of just processing current nodes.
func ReversePreorderPath(visitor PathVisitor) {
	stackBasedVisitPathFunc(visitor, stackBasedReversePreorderPathMain)
}

// stackBasedReversePreorderPathMain is the main body of ReversePreorderPath
// that executes reverse pre-order traversal with a stack-based implementation
// in the iterative approach.
func stackBasedReversePreorderPathMain(
	visitor PathVisitor,
	root Node,
	opts *Options,
	traversalStack stack.Stack[any],
) {
	var step int64

	traversalStack.Push(vseq.NewVertexSequence(root))

	for traversalStack.Len() > 0 {
		top := traversalStack.Pop().(vseq.VertexSequence[Node])
		topDepth := top.Len() - 1

		step++
		cont, skipChildren := visitor.VisitPath(step, top, topDepth)

		if !cont || step == opts.MaxStep {
			return
		}

		if !skipChildren && topDepth != opts.MaxDepth {
			for c := top.Back().FirstChild(); c != nil; c = c.NextSibling() {
				traversalStack.Push(top.Append(c))
			}
		}
	}
}

// IterPreorder returns an iterator over nodes
// of the tree with the specified root,
// traversing it in pre-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the children of the current node.
// If it is nil, no nodes will be skipped.
//
// The returned iterator is never nil.
func IterPreorder(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		if yield != nil && root != nil {
			Preorder(&iteratorNodeVisitor{
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

// IterPreorderDepth returns an iterator over node-depth pairs
// of the tree with the specified root,
// traversing it in pre-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the children of the current node.
// If it is nil, no nodes will be skipped.
//
// The returned iterator is never nil.
func IterPreorderDepth(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq2[Node, int] {
	return func(yield func(Node, int) bool) {
		if yield != nil && root != nil {
			Preorder(&iteratorNodeDepthVisitor{
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

// IterPreorderStep returns an iterator over step-node pairs
// of the tree with the specified root,
// traversing it in pre-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the children of the current node.
// If it is nil, no nodes will be skipped.
//
// The returned iterator is never nil.
func IterPreorderStep(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq2[int64, Node] {
	return func(yield func(int64, Node) bool) {
		if yield != nil && root != nil {
			Preorder(&iteratorStepNodeVisitor{
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

// IterPreorderPath returns an iterator over paths
// from the root to nodes of the tree with the specified root,
// traversing it in pre-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the children of the current node.
// If it is nil, no nodes will be skipped.
//
// The returned iterator is never nil.
func IterPreorderPath(
	root Node,
	opts *Options,
	skipChildrenFn PathSkipChildrenFunc,
) iter.Seq[vseq.VertexSequence[Node]] {
	return func(yield func(vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			PreorderPath(&iteratorPathVisitor{
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

// IterPreorderPathStep returns an iterator over step-path pairs where
// the paths are from the root to nodes of the tree with the specified root,
// traversing it in pre-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the children of the current node.
// If it is nil, no nodes will be skipped.
//
// The returned iterator is never nil.
func IterPreorderPathStep(
	root Node,
	opts *Options,
	skipChildrenFn PathSkipChildrenFunc,
) iter.Seq2[int64, vseq.VertexSequence[Node]] {
	return func(yield func(int64, vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			PreorderPath(&iteratorStepPathVisitor{
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

// IterReversePreorder returns an iterator over nodes
// of the tree with the specified root,
// traversing it in reverse pre-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the children of the current node.
// If it is nil, no nodes will be skipped.
//
// The returned iterator is never nil.
func IterReversePreorder(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		if yield != nil && root != nil {
			ReversePreorder(&iteratorNodeVisitor{
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

// IterReversePreorderDepth returns an iterator over node-depth pairs
// of the tree with the specified root,
// traversing it in reverse pre-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the children of the current node.
// If it is nil, no nodes will be skipped.
//
// The returned iterator is never nil.
func IterReversePreorderDepth(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq2[Node, int] {
	return func(yield func(Node, int) bool) {
		if yield != nil && root != nil {
			ReversePreorder(&iteratorNodeDepthVisitor{
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

// IterReversePreorderStep returns an iterator over step-node pairs
// of the tree with the specified root,
// traversing it in reverse pre-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the children of the current node.
// If it is nil, no nodes will be skipped.
//
// The returned iterator is never nil.
func IterReversePreorderStep(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq2[int64, Node] {
	return func(yield func(int64, Node) bool) {
		if yield != nil && root != nil {
			ReversePreorder(&iteratorStepNodeVisitor{
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

// IterReversePreorderPath returns an iterator over paths
// from the root to nodes of the tree with the specified root,
// traversing it in reverse pre-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the children of the current node.
// If it is nil, no nodes will be skipped.
//
// The returned iterator is never nil.
func IterReversePreorderPath(
	root Node,
	opts *Options,
	skipChildrenFn PathSkipChildrenFunc,
) iter.Seq[vseq.VertexSequence[Node]] {
	return func(yield func(vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			ReversePreorderPath(&iteratorPathVisitor{
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

// IterReversePreorderPathStep returns an iterator over step-path pairs where
// the paths are from the root to nodes of the tree with the specified root,
// traversing it in reverse pre-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// skipChildrenFn is a function that reports
// whether to skip the children of the current node.
// If it is nil, no nodes will be skipped.
//
// The returned iterator is never nil.
func IterReversePreorderPathStep(
	root Node,
	opts *Options,
	skipChildrenFn PathSkipChildrenFunc,
) iter.Seq2[int64, vseq.VertexSequence[Node]] {
	return func(yield func(int64, vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			ReversePreorderPath(&iteratorStepPathVisitor{
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
