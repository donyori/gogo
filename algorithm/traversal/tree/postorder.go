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

// Postorder executes post-order traversal for the specified visitor.
// Post-order traversal means that for each node,
// recursively traverse the subtrees of the node from first to last,
// and finally visit the node.
//
// The traversal uses a stack-based implementation in the iterative approach.
//
// If the specified visitor is nil or the root returned by the visitor is nil,
// Postorder does nothing.
//
// The return value skipChildren of the visitor's method VisitNode
// doesn't make sense for post-order traversal.
func Postorder(visitor NodeVisitor) {
	stackBasedVisitNodeFunc(visitor, stackBasedPostorderMain)
}

// stackBasedPostorderMain is the main body of Postorder
// that executes post-order traversal with a stack-based implementation
// in the iterative approach.
func stackBasedPostorderMain(
	visitor NodeVisitor,
	root Node,
	opts *Options,
	traversalStack stack.Stack[any],
) {
	var (
		step      int64
		lastDepth int
	)

	traversalStack.Push(nodeDepth{node: root})

	for traversalStack.Len() > 0 {
		top := traversalStack.Pop().(nodeDepth)

		// (top.depth != opts.MaxDepth) is equivalent to
		// (opts.MaxDepth < 0 || top.depth < opts.MaxDepth)
		// because the depth starts from 0 and increases by 1 each time.
		pushTopAndFirstChild := top.depth != opts.MaxDepth &&
			top.depth >= lastDepth
		lastDepth = top.depth

		if pushTopAndFirstChild {
			if fc := top.node.FirstChild(); fc != nil {
				traversalStack.Push(top) // the current node is pushed again
				traversalStack.Push(nodeDepth{node: fc, depth: top.depth + 1})

				continue
			}
		}

		step++
		cont, _ := visitor.VisitNode(step, top.node, top.depth)

		// (step == opts.MaxStep) is equivalent to
		// (opts.MaxStep > 0 && step >= opts.MaxStep)
		// because the step starts from 1 and increases by 1 each time.
		if !cont || step == opts.MaxStep {
			return
		}

		// The next sibling is pushed after visiting the current node.
		//
		// Consider siblings only when the current node is not the root
		// (equivalently, top.depth > 0) to allow that
		// the root returned by the visitor is the root of a subtree.
		if top.depth > 0 {
			if ns := top.node.NextSibling(); ns != nil {
				traversalStack.Push(nodeDepth{node: ns, depth: top.depth})
			}
		}
	}
}

// PostorderPath is similar to function Postorder,
// but it processes paths from the root to current nodes,
// instead of just processing current nodes.
func PostorderPath(visitor PathVisitor) {
	stackBasedVisitPathFunc(visitor, stackBasedPostorderPathMain)
}

// stackBasedPostorderPathMain is the main body of PostorderPath
// that executes post-order traversal with a stack-based implementation
// in the iterative approach.
func stackBasedPostorderPathMain(
	visitor PathVisitor,
	root Node,
	opts *Options,
	traversalStack stack.Stack[any],
) {
	var (
		step      int64
		lastDepth int
	)

	traversalStack.Push(vseq.NewVertexSequence(root))

	for traversalStack.Len() > 0 {
		top := traversalStack.Pop().(vseq.VertexSequence[Node])
		topDepth := top.Len() - 1
		pushTopAndFirstChild := topDepth != opts.MaxDepth &&
			topDepth >= lastDepth
		lastDepth = topDepth

		if pushTopAndFirstChild {
			if fc := top.Back().FirstChild(); fc != nil {
				traversalStack.Push(top)
				traversalStack.Push(top.Append(fc))

				continue
			}
		}

		step++
		cont, _ := visitor.VisitPath(step, top, topDepth)

		if !cont || step == opts.MaxStep {
			return
		}

		if topDepth > 0 {
			if ns := top.Back().NextSibling(); ns != nil {
				traversalStack.Push(top.ReplaceBack(ns))
			}
		}
	}
}

// ReversePostorder executes reverse post-order traversal
// for the specified visitor.
// Reverse post-order traversal means that for each node,
// recursively traverse the subtrees of the node from last to first,
// and finally visit the node.
//
// The traversal uses a stack-based implementation in the iterative approach.
//
// If the specified visitor is nil or the root returned by the visitor is nil,
// ReversePostorder does nothing.
//
// The return value skipChildren of the visitor's method VisitNode
// doesn't make sense for reverse post-order traversal.
func ReversePostorder(visitor NodeVisitor) {
	stackBasedVisitNodeFunc(visitor, stackBasedReversePostorderMain)
}

// stackBasedReversePostorderMain is the main body of ReversePostorder
// that executes reverse post-order traversal with a stack-based implementation
// in the iterative approach.
func stackBasedReversePostorderMain(
	visitor NodeVisitor,
	root Node,
	opts *Options,
	traversalStack stack.Stack[any],
) {
	var (
		step      int64
		lastDepth int
	)

	traversalStack.Push(nodeDepth{node: root})

	for traversalStack.Len() > 0 {
		top := traversalStack.Pop().(nodeDepth)

		// (top.depth != opts.MaxDepth) is equivalent to
		// (opts.MaxDepth < 0 || top.depth < opts.MaxDepth)
		// because the depth starts from 0 and increases by 1 each time.
		pushTopAndChildren := top.depth != opts.MaxDepth &&
			top.depth >= lastDepth
		lastDepth = top.depth

		if pushTopAndChildren {
			if fc := top.node.FirstChild(); fc != nil {
				traversalStack.Push(top) // the path from the root to the current node is pushed again

				// The children are pushed from first to last
				// so that they are visited from last to first.
				for c := fc; c != nil; c = c.NextSibling() {
					traversalStack.Push(nodeDepth{
						node:  c,
						depth: top.depth + 1,
					})
				}

				continue
			}
		}

		step++
		cont, _ := visitor.VisitNode(step, top.node, top.depth)

		// (step == opts.MaxStep) is equivalent to
		// (opts.MaxStep > 0 && step >= opts.MaxStep)
		// because the step starts from 1 and increases by 1 each time.
		if !cont || step == opts.MaxStep {
			return
		}
	}
}

// ReversePostorderPath is similar to function ReversePostorder,
// but it processes paths from the root to current nodes,
// instead of just processing current nodes.
func ReversePostorderPath(visitor PathVisitor) {
	stackBasedVisitPathFunc(visitor, stackBasedReversePostorderPathMain)
}

// stackBasedReversePostorderPathMain is the main body of ReversePostorderPath
// that executes reverse post-order traversal with a stack-based implementation
// in the iterative approach.
func stackBasedReversePostorderPathMain(
	visitor PathVisitor,
	root Node,
	opts *Options,
	traversalStack stack.Stack[any],
) {
	var (
		step      int64
		lastDepth int
	)

	traversalStack.Push(vseq.NewVertexSequence(root))

	for traversalStack.Len() > 0 {
		top := traversalStack.Pop().(vseq.VertexSequence[Node])
		topDepth := top.Len() - 1
		pushTopAndChildren := topDepth != opts.MaxDepth && topDepth >= lastDepth
		lastDepth = topDepth

		if pushTopAndChildren {
			if fc := top.Back().FirstChild(); fc != nil {
				traversalStack.Push(top)

				for c := fc; c != nil; c = c.NextSibling() {
					traversalStack.Push(top.Append(c))
				}

				continue
			}
		}

		step++
		cont, _ := visitor.VisitPath(step, top, topDepth)

		if !cont || step == opts.MaxStep {
			return
		}
	}
}

// IterPostorder returns an iterator over nodes
// of the tree with the specified root,
// traversing it in post-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// The returned iterator is never nil.
func IterPostorder(root Node, opts *Options) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		if yield != nil && root != nil {
			Postorder(&iteratorNodeVisitor{
				yield: yield,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterPostorderDepth returns an iterator over node-depth pairs
// of the tree with the specified root,
// traversing it in post-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// The returned iterator is never nil.
func IterPostorderDepth(root Node, opts *Options) iter.Seq2[Node, int] {
	return func(yield func(Node, int) bool) {
		if yield != nil && root != nil {
			Postorder(&iteratorNodeDepthVisitor{
				yield: yield,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterPostorderStep returns an iterator over step-node pairs
// of the tree with the specified root,
// traversing it in post-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// The returned iterator is never nil.
func IterPostorderStep(root Node, opts *Options) iter.Seq2[int64, Node] {
	return func(yield func(int64, Node) bool) {
		if yield != nil && root != nil {
			Postorder(&iteratorStepNodeVisitor{
				yield: yield,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterPostorderPath returns an iterator over paths
// from the root to nodes of the tree with the specified root,
// traversing it in post-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// The returned iterator is never nil.
func IterPostorderPath(
	root Node,
	opts *Options,
) iter.Seq[vseq.VertexSequence[Node]] {
	return func(yield func(vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			PostorderPath(&iteratorPathVisitor{
				yield: yield,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterPostorderPathStep returns an iterator over step-path pairs where
// the paths are from the root to nodes of the tree with the specified root,
// traversing it in post-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// The returned iterator is never nil.
func IterPostorderPathStep(
	root Node,
	opts *Options,
) iter.Seq2[int64, vseq.VertexSequence[Node]] {
	return func(yield func(int64, vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			PostorderPath(&iteratorStepPathVisitor{
				yield: yield,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterReversePostorder returns an iterator over nodes
// of the tree with the specified root,
// traversing it in reverse post-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// The returned iterator is never nil.
func IterReversePostorder(root Node, opts *Options) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		if yield != nil && root != nil {
			ReversePostorder(&iteratorNodeVisitor{
				yield: yield,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterReversePostorderDepth returns an iterator over node-depth pairs
// of the tree with the specified root,
// traversing it in reverse post-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// The returned iterator is never nil.
func IterReversePostorderDepth(root Node, opts *Options) iter.Seq2[Node, int] {
	return func(yield func(Node, int) bool) {
		if yield != nil && root != nil {
			ReversePostorder(&iteratorNodeDepthVisitor{
				yield: yield,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterReversePostorderStep returns an iterator over step-node pairs
// of the tree with the specified root,
// traversing it in reverse post-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// The returned iterator is never nil.
func IterReversePostorderStep(root Node, opts *Options) iter.Seq2[int64, Node] {
	return func(yield func(int64, Node) bool) {
		if yield != nil && root != nil {
			ReversePostorder(&iteratorStepNodeVisitor{
				yield: yield,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterReversePostorderPath returns an iterator over paths
// from the root to nodes of the tree with the specified root,
// traversing it in reverse post-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// The returned iterator is never nil.
func IterReversePostorderPath(
	root Node,
	opts *Options,
) iter.Seq[vseq.VertexSequence[Node]] {
	return func(yield func(vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			ReversePostorderPath(&iteratorPathVisitor{
				yield: yield,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}

// IterReversePostorderPathStep returns an iterator over step-path pairs where
// the paths are from the root to nodes of the tree with the specified root,
// traversing it in reverse post-order traversal with the specified options.
//
// If opts are nil, the default options are used.
// The default options are as follows:
//   - MaxStep: 0
//   - MaxDepth: -1
//   - LocalBuf: false
//
// The returned iterator is never nil.
func IterReversePostorderPathStep(
	root Node,
	opts *Options,
) iter.Seq2[int64, vseq.VertexSequence[Node]] {
	return func(yield func(int64, vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			ReversePostorderPath(&iteratorStepPathVisitor{
				yield: yield,
				visitorCommon: visitorCommon{
					root: root,
					opts: opts,
				},
			})
		}
	}
}
