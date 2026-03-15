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
	"github.com/donyori/gogo/container/queue"
	"github.com/donyori/gogo/container/set"
	"github.com/donyori/gogo/container/set/mapset"
	"github.com/donyori/gogo/container/stack"
)

// LevelOrder executes level-order traversal for the specified visitor.
// Level-order traversal means that start at the root and
// visit all nodes at the present depth completely
// prior to moving on to the nodes at the next depth level.
//
// The traversal uses a stack-based implementation
// in the iterative deepening approach.
//
// If the specified visitor is nil or the root returned by the visitor is nil,
// LevelOrder does nothing.
func LevelOrder(visitor NodeVisitor) {
	stackBasedVisitNodeFunc(visitor, iterativeDeepeningLevelOrderMain)
}

// iterativeDeepeningLevelOrderMain is the main body of LevelOrder
// that executes level-order traversal with a stack-based implementation
// in the iterative deepening approach.
func iterativeDeepeningLevelOrderMain(
	visitor NodeVisitor,
	root Node,
	opts *Options,
	traversalStack stack.Stack[any],
) {
	rootItem := nodeDepth{node: root}

	var skipChildSet set.Set[Node]
	if opts.LocalBuf {
		skipChildSet = mapset.New[Node](0)
	} else {
		skipChildSet = nodeSetPool.Get().(set.Set[Node])
		defer func(s set.Set[Node]) {
			s.RemoveAll() // avoid memory leak
			nodeSetPool.Put(s)
		}(skipChildSet)
	}

	var step int64

	for visitDepth, goNextLevel := 0, true; goNextLevel; visitDepth++ {
		goNextLevel = false

		traversalStack.RemoveAll()
		traversalStack.Push(rootItem)

		for traversalStack.Len() > 0 {
			cont, more := iterativeDeepeningLevelOrderHandleStackTop(
				visitor,
				opts,
				traversalStack,
				skipChildSet,
				&step,
				visitDepth,
			)
			if !cont {
				return
			}

			goNextLevel = goNextLevel || more
		}
	}
}

// iterativeDeepeningLevelOrderHandleStackTop is a subprocess of
// iterativeDeepeningLevelOrderMain that handles the current node
// (i.e., the top of the stack).
//
// It reports whether the loop continues and
// whether any nodes exceed the depth limit (i.e., visitDepth).
func iterativeDeepeningLevelOrderHandleStackTop(
	visitor NodeVisitor,
	opts *Options,
	traversalStack stack.Stack[any],
	skipChildSet set.Set[Node],
	pStep *int64,
	visitDepth int,
) (cont, more bool) {
	top := traversalStack.Pop().(nodeDepth)

	var skipChildren bool

	if top.depth == visitDepth {
		*pStep++
		cont, skipChildren = visitor.VisitNode(*pStep, top.node, top.depth)

		if skipChildren {
			skipChildSet.Add(top.node)
		}

		// (*pStep == opts.MaxStep) is equivalent to
		// (opts.MaxStep > 0 && *pStep >= opts.MaxStep)
		// because the step starts from 1 and increases by 1 each time.
		if !cont || *pStep == opts.MaxStep {
			return false, false
		}
	} else {
		skipChildren = skipChildSet.ContainsItem(top.node)
	}

	cont = true

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
			if top.depth < visitDepth {
				traversalStack.Push(nodeDepth{node: fc, depth: top.depth + 1})
			} else {
				more = true
			}
		}
	}

	return
}

// LevelOrderPath is similar to function LevelOrder,
// but it processes paths from the root to current nodes,
// instead of just processing current nodes.
func LevelOrderPath(visitor PathVisitor) {
	stackBasedVisitPathFunc(visitor, iterativeDeepeningLevelOrderPathMain)
}

// iterativeDeepeningLevelOrderPathMain is the main body of LevelOrderPath
// that executes level-order traversal with a stack-based implementation
// in the iterative deepening approach.
func iterativeDeepeningLevelOrderPathMain(
	visitor PathVisitor,
	root Node,
	opts *Options,
	traversalStack stack.Stack[any],
) {
	rootItem := vseq.NewVertexSequence(root)

	var skipChildSet set.Set[Node]
	if opts.LocalBuf {
		skipChildSet = mapset.New[Node](0)
	} else {
		skipChildSet = nodeSetPool.Get().(set.Set[Node])
		defer func(s set.Set[Node]) {
			s.RemoveAll() // avoid memory leak
			nodeSetPool.Put(s)
		}(skipChildSet)
	}

	var step int64

	for visitDepth, goNextLevel := 0, true; goNextLevel; visitDepth++ {
		goNextLevel = false

		traversalStack.RemoveAll()
		traversalStack.Push(rootItem)

		for traversalStack.Len() > 0 {
			cont, more := iterativeDeepeningLevelOrderPathHandleStackTop(
				visitor,
				opts,
				traversalStack,
				skipChildSet,
				&step,
				visitDepth,
			)
			if !cont {
				return
			}

			goNextLevel = goNextLevel || more
		}
	}
}

// iterativeDeepeningLevelOrderPathHandleStackTop is a subprocess of
// iterativeDeepeningLevelOrderPathMain that handles the current node
// (i.e., the top of the stack).
//
// It reports whether the loop continues and
// whether any nodes exceed the depth limit (i.e., visitDepth).
func iterativeDeepeningLevelOrderPathHandleStackTop(
	visitor PathVisitor,
	opts *Options,
	traversalStack stack.Stack[any],
	skipChildSet set.Set[Node],
	pStep *int64,
	visitDepth int,
) (cont, more bool) {
	top := traversalStack.Pop().(vseq.VertexSequence[Node])
	topDepth := top.Len() - 1

	var skipChildren bool

	if topDepth == visitDepth {
		*pStep++
		cont, skipChildren = visitor.VisitPath(*pStep, top, topDepth)

		if skipChildren {
			skipChildSet.Add(top.Back())
		}

		if !cont || *pStep == opts.MaxStep {
			return false, false
		}
	} else {
		skipChildren = skipChildSet.ContainsItem(top.Back())
	}

	cont = true

	if topDepth > 0 {
		if ns := top.Back().NextSibling(); ns != nil {
			traversalStack.Push(top.ReplaceBack(ns))
		}
	}

	if !skipChildren && topDepth != opts.MaxDepth {
		if fc := top.Back().FirstChild(); fc != nil {
			if topDepth < visitDepth {
				traversalStack.Push(top.Append(fc))
			} else {
				more = true
			}
		}
	}

	return
}

// LevelOrderQueueBased executes level-order traversal
// for the specified visitor.
// Level-order traversal means that start at the root and
// visit all nodes at the present depth completely
// prior to moving on to the nodes at the next depth level.
//
// The traversal uses a queue-based implementation in the iterative approach.
//
// If the specified visitor is nil or the root returned by the visitor is nil,
// LevelOrderQueueBased does nothing.
func LevelOrderQueueBased(visitor NodeVisitor) {
	queueBasedVisitNodeFunc(visitor, queueBasedLevelOrderMain)
}

// queueBasedLevelOrderMain is the main body of LevelOrderQueueBased
// that executes level-order traversal with a queue-based implementation
// in the iterative approach.
func queueBasedLevelOrderMain(
	visitor NodeVisitor,
	root Node,
	opts *Options,
	traversalQueue queue.Queue[any],
) {
	var step int64

	traversalQueue.Enqueue(nodeDepth{node: root})

	for traversalQueue.Len() > 0 {
		head := traversalQueue.Dequeue().(nodeDepth)
		node := head.node

		// It is sufficient to just add the first child
		// of each node to the queue.
		// Other children can be obtained by method NextSibling within a loop.
		for node != nil {
			step++
			cont, skipChildren := visitor.VisitNode(step, node, head.depth)

			// (step == opts.MaxStep) is equivalent to
			// (opts.MaxStep > 0 && step >= opts.MaxStep)
			// because the step starts from 1 and increases by 1 each time.
			if !cont || step == opts.MaxStep {
				return
			}

			// (head.depth != opts.MaxDepth) is equivalent to
			// (opts.MaxDepth < 0 || head.depth < opts.MaxDepth)
			// because the depth starts from 0 and increases by 1 each time.
			if !skipChildren && head.depth != opts.MaxDepth {
				if fc := node.FirstChild(); fc != nil {
					traversalQueue.Enqueue(nodeDepth{
						node:  fc,
						depth: head.depth + 1,
					})
				}
			}

			// Don't consider siblings when the current node is the root
			// (equivalently, head.depth == 0) to allow that
			// the root returned by the visitor is the root of a subtree.
			if head.depth == 0 {
				break
			}

			node = node.NextSibling()
		}
	}
}

// LevelOrderPathQueueBased is similar to function LevelOrderQueueBased,
// but it processes paths from the root to current nodes,
// instead of just processing current nodes.
func LevelOrderPathQueueBased(visitor PathVisitor) {
	queueBasedVisitPathFunc(visitor, queueBasedLevelOrderPathMain)
}

// queueBasedLevelOrderPathMain is the main body of LevelOrderPathQueueBased
// that executes level-order traversal with a queue-based implementation
// in the iterative approach.
func queueBasedLevelOrderPathMain(
	visitor PathVisitor,
	root Node,
	opts *Options,
	traversalQueue queue.Queue[any],
) {
	var step int64

	traversalQueue.Enqueue(vseq.NewVertexSequence(root))

	for traversalQueue.Len() > 0 {
		nSeq := traversalQueue.Dequeue().(vseq.VertexSequence[Node])
		headDepth := nSeq.Len() - 1

		for nSeq != nil {
			step++
			cont, skipChildren := visitor.VisitPath(step, nSeq, headDepth)

			if !cont || step == opts.MaxStep {
				return
			}

			if !skipChildren && headDepth != opts.MaxDepth {
				if fc := nSeq.Back().FirstChild(); fc != nil {
					traversalQueue.Enqueue(nSeq.Append(fc))
				}
			}

			nSeq = queueBasedLevelOrderPathUpdateNodePath(nSeq, headDepth)
		}
	}
}

// queueBasedLevelOrderPathUpdateNodePath updates the node path
// in the inner loop of queueBasedLevelOrderPathMain.
//
// It returns nil if depth is 0 (i.e., the current node is the root)
// or the current node is the last child of its parent.
// Otherwise, it returns the node path for the next sibling.
func queueBasedLevelOrderPathUpdateNodePath(
	nodePath vseq.VertexSequence[Node],
	depth int,
) vseq.VertexSequence[Node] {
	if depth == 0 {
		return nil
	}

	ns := nodePath.Back().NextSibling()
	if ns == nil {
		return nil
	}

	return nodePath.ReplaceBack(ns)
}

// IterLevelOrder returns an iterator over nodes
// of the tree with the specified root,
// traversing it in stack-based level-order traversal
// with the specified options.
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
func IterLevelOrder(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		if yield != nil && root != nil {
			LevelOrder(&iteratorNodeVisitor{
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

// IterLevelOrderDepth returns an iterator over node-depth pairs
// of the tree with the specified root,
// traversing it in stack-based level-order traversal
// with the specified options.
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
func IterLevelOrderDepth(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq2[Node, int] {
	return func(yield func(Node, int) bool) {
		if yield != nil && root != nil {
			LevelOrder(&iteratorNodeDepthVisitor{
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

// IterLevelOrderStep returns an iterator over step-node pairs
// of the tree with the specified root,
// traversing it in stack-based level-order traversal
// with the specified options.
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
func IterLevelOrderStep(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq2[int64, Node] {
	return func(yield func(int64, Node) bool) {
		if yield != nil && root != nil {
			LevelOrder(&iteratorStepNodeVisitor{
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

// IterLevelOrderPath returns an iterator over paths
// from the root to nodes of the tree with the specified root,
// traversing it in stack-based level-order traversal
// with the specified options.
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
func IterLevelOrderPath(
	root Node,
	opts *Options,
	skipChildrenFn PathSkipChildrenFunc,
) iter.Seq[vseq.VertexSequence[Node]] {
	return func(yield func(vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			LevelOrderPath(&iteratorPathVisitor{
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

// IterLevelOrderPathStep returns an iterator over step-path pairs where
// the paths are from the root to nodes of the tree with the specified root,
// traversing it in stack-based level-order traversal
// with the specified options.
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
func IterLevelOrderPathStep(
	root Node,
	opts *Options,
	skipChildrenFn PathSkipChildrenFunc,
) iter.Seq2[int64, vseq.VertexSequence[Node]] {
	return func(yield func(int64, vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			LevelOrderPath(&iteratorStepPathVisitor{
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

// IterLevelOrderQueueBased returns an iterator over nodes
// of the tree with the specified root,
// traversing it in queue-based level-order traversal
// with the specified options.
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
func IterLevelOrderQueueBased(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		if yield != nil && root != nil {
			LevelOrderQueueBased(&iteratorNodeVisitor{
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

// IterLevelOrderDepthQueueBased returns an iterator over node-depth pairs
// of the tree with the specified root,
// traversing it in queue-based level-order traversal
// with the specified options.
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
func IterLevelOrderDepthQueueBased(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq2[Node, int] {
	return func(yield func(Node, int) bool) {
		if yield != nil && root != nil {
			LevelOrderQueueBased(&iteratorNodeDepthVisitor{
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

// IterLevelOrderStepQueueBased returns an iterator over step-node pairs
// of the tree with the specified root,
// traversing it in queue-based level-order traversal
// with the specified options.
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
func IterLevelOrderStepQueueBased(
	root Node,
	opts *Options,
	skipChildrenFn NodeSkipChildrenFunc,
) iter.Seq2[int64, Node] {
	return func(yield func(int64, Node) bool) {
		if yield != nil && root != nil {
			LevelOrderQueueBased(&iteratorStepNodeVisitor{
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

// IterLevelOrderPathQueueBased returns an iterator over paths
// from the root to nodes of the tree with the specified root,
// traversing it in queue-based level-order traversal
// with the specified options.
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
func IterLevelOrderPathQueueBased(
	root Node,
	opts *Options,
	skipChildrenFn PathSkipChildrenFunc,
) iter.Seq[vseq.VertexSequence[Node]] {
	return func(yield func(vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			LevelOrderPathQueueBased(&iteratorPathVisitor{
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

// IterLevelOrderPathStepQueueBased returns an iterator over
// step-path pairs where the paths are from the root to
// nodes of the tree with the specified root,
// traversing it in queue-based level-order traversal
// with the specified options.
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
func IterLevelOrderPathStepQueueBased(
	root Node,
	opts *Options,
	skipChildrenFn PathSkipChildrenFunc,
) iter.Seq2[int64, vseq.VertexSequence[Node]] {
	return func(yield func(int64, vseq.VertexSequence[Node]) bool) {
		if yield != nil && root != nil {
			LevelOrderPathQueueBased(&iteratorStepPathVisitor{
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
