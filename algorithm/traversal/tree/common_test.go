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

package tree_test

import (
	"fmt"
	"iter"
	"strings"
	"testing"

	"github.com/donyori/gogo/algorithm/traversal/tree"
	"github.com/donyori/gogo/algorithm/traversal/vseq"
)

// StringNode is an implementation of interface
// github.com/donyori/gogo/algorithm/traversal/tree.InorderNode for testing.
//
// The tree is as follows:
//
//	T0a -+- L1a -+- L2a -+- L3a
//	     |       |       +- R3b
//	     |       |
//	     |       +- L2b --- L3c
//	     |       |
//	     |       +- L2c --- R3d
//	     |       |
//	     |       +- R2d -+- L3e
//	     |       |       +- L3f
//	     |       |       +- R3g
//	     |       |
//	     |       +- R2e -+- R3h
//	     |               +- R3i
//	     |
//	     +- L1b
//	     |
//	     +- L1c -+- L2f
//	     |       |
//	     |       +- L2g -+- L3j
//	     |       |       +- R3k
//	     |       |
//	     |       +- R2h
//	     |       |
//	     |       +- R2i -+- L3l
//	     |               +- R3m
//	     |
//	     +- R1d -+- R2j
//	     |       +- R2k
//	     |
//	     +- R1e --- L2l --- R3n
type StringNode string

var _ tree.InorderNode = StringNode("")

var (
	// stringNodeFirstChildMap records the first child of each StringNode.
	stringNodeFirstChildMap = map[StringNode]StringNode{
		"T0a": "L1a",

		"L1a": "L2a",
		"L2a": "L3a",
		"L2b": "L3c",
		"L2c": "R3d",
		"R2d": "L3e",
		"R2e": "R3h",

		"L1c": "L2f",
		"L2g": "L3j",
		"R2i": "L3l",

		"R1d": "R2j",

		"R1e": "L2l",
		"L2l": "R3n",
	}

	// stringNodeNextSiblingMap records the next sibling of each StringNode.
	stringNodeNextSiblingMap = map[StringNode]StringNode{
		"L1a": "L1b",
		"L1b": "L1c",
		"L1c": "R1d",
		"R1d": "R1e",

		"L2a": "L2b",
		"L2b": "L2c",
		"L2c": "R2d",
		"R2d": "R2e",

		"L2f": "L2g",
		"L2g": "R2h",
		"R2h": "R2i",

		"R2j": "R2k",

		"L3a": "R3b",
		"L3e": "L3f",
		"L3f": "R3g",
		"R3h": "R3i",

		"L3j": "R3k",
		"L3l": "R3m",
	}
)

func (n StringNode) FirstChild() tree.Node {
	if fc, ok := stringNodeFirstChildMap[n]; ok {
		return fc
	}

	return nil
}

func (n StringNode) NextSibling() tree.Node {
	if ns, ok := stringNodeNextSiblingMap[n]; ok {
		return ns
	}

	return nil
}

func (n StringNode) Left() bool {
	return strings.HasPrefix(string(n), "L")
}

func (n StringNode) Depth() int {
	if len(n) >= 2 {
		c := n[1]
		if c >= '0' && c <= '3' {
			return int(c - '0')
		}
	}

	return -1
}

// NumIteratorTestCase is the number of test cases for testing iterators.
const NumIteratorTestCase int = 24

// NodeSkipChildrenR2dL1bL1cL2l is a
// github.com/donyori/gogo/algorithm/traversal/tree.NodeSkipChildrenFunc
// that skips the children of StringNode "R2d", "L1b", "L1c", "L2l".
func NodeSkipChildrenR2dL1bL1cL2l(node tree.Node, _ int64, _ int) bool {
	if sn, ok := node.(StringNode); ok {
		switch sn {
		case "R2d", "L1b", "L1c", "L2l":
			return true
		}
	}

	return false
}

// PathSkipChildrenR2dL1bL1cL2l is a
// github.com/donyori/gogo/algorithm/traversal/tree.PathSkipChildrenFunc
// that skips the children of StringNode "R2d", "L1b", "L1c", "L2l".
func PathSkipChildrenR2dL1bL1cL2l(
	nodePath vseq.VertexSequence[tree.Node],
	_ int64,
	_ int,
) bool {
	if nodePath != nil && nodePath.Len() > 0 {
		return NodeSkipChildrenR2dL1bL1cL2l(nodePath.Back(), 0, 0)
	}

	return false
}

// IteratorTestCaseInputs are 24 test case inputs for testing iterators.
var IteratorTestCaseInputs = [NumIteratorTestCase]struct {
	Root               tree.Node
	Opts               *tree.Options
	NodeSkipChildrenFn tree.NodeSkipChildrenFunc
	PathSkipChildrenFn tree.PathSkipChildrenFunc
}{
	{},
	{Root: StringNode("T0a")},
	{Root: StringNode("L1b")},
	{Root: StringNode("L1c")},

	{Opts: &tree.Options{MaxStep: 7, MaxDepth: -1}},
	{Root: StringNode("T0a"), Opts: &tree.Options{MaxStep: 7, MaxDepth: -1}},
	{Root: StringNode("L1b"), Opts: &tree.Options{MaxStep: 7, MaxDepth: -1}},
	{Root: StringNode("L1c"), Opts: &tree.Options{MaxStep: 7, MaxDepth: -1}},

	{Opts: &tree.Options{MaxDepth: 2}},
	{Root: StringNode("T0a"), Opts: &tree.Options{MaxDepth: 2}},
	{Root: StringNode("L1b"), Opts: &tree.Options{MaxDepth: 2}},
	{Root: StringNode("L1c"), Opts: &tree.Options{MaxDepth: 2}},

	{Opts: &tree.Options{MaxStep: 7, MaxDepth: 2}},
	{Root: StringNode("T0a"), Opts: &tree.Options{MaxStep: 7, MaxDepth: 2}},
	{Root: StringNode("L1b"), Opts: &tree.Options{MaxStep: 7, MaxDepth: 2}},
	{Root: StringNode("L1c"), Opts: &tree.Options{MaxStep: 7, MaxDepth: 2}},

	{Opts: &tree.Options{MaxDepth: -1, LocalBuf: true}},
	{
		Root: StringNode("T0a"),
		Opts: &tree.Options{MaxDepth: -1, LocalBuf: true},
	},
	{
		Root: StringNode("L1b"),
		Opts: &tree.Options{MaxDepth: -1, LocalBuf: true},
	},
	{
		Root: StringNode("L1c"),
		Opts: &tree.Options{MaxDepth: -1, LocalBuf: true},
	},

	{
		NodeSkipChildrenFn: NodeSkipChildrenR2dL1bL1cL2l,
		PathSkipChildrenFn: PathSkipChildrenR2dL1bL1cL2l,
	},
	{
		Root:               StringNode("T0a"),
		NodeSkipChildrenFn: NodeSkipChildrenR2dL1bL1cL2l,
		PathSkipChildrenFn: PathSkipChildrenR2dL1bL1cL2l,
	},
	{
		Root:               StringNode("L1b"),
		NodeSkipChildrenFn: NodeSkipChildrenR2dL1bL1cL2l,
		PathSkipChildrenFn: PathSkipChildrenR2dL1bL1cL2l,
	},
	{
		Root:               StringNode("L1c"),
		NodeSkipChildrenFn: NodeSkipChildrenR2dL1bL1cL2l,
		PathSkipChildrenFn: PathSkipChildrenR2dL1bL1cL2l,
	},
}

// stringNodeToName returns the name of the specified tree node
// displayed in subtest names, as follows:
//   - If the node is nil, it returns "<nil>".
//   - If the node is a StringNode, it returns the string value.
//   - Otherwise, it formats the name by fmt.Sprintf with the verb "%#v".
func stringNodeToName(n tree.Node) string {
	if n == nil {
		return "<nil>"
	} else if sn, ok := n.(StringNode); ok {
		return string(sn)
	}

	return fmt.Sprintf("%#v", n)
}

// optionsToName returns the name of the specified options
// displayed in subtest names.
func optionsToName(opts *tree.Options) string {
	if opts == nil {
		return "<nil>"
	}

	return fmt.Sprintf("{MaxStep=%d&MaxDepth=%d&LocalBuf=%t}",
		opts.MaxStep, opts.MaxDepth, opts.LocalBuf)
}

// checkNodeIterator checks the iterator over tree nodes.
func checkNodeIterator(t *testing.T, seq iter.Seq[tree.Node], want string) {
	t.Helper()

	if seq == nil {
		t.Error("got nil iterator")
		return
	}

	var b strings.Builder

	for n := range seq {
		if n == nil {
			b.WriteString("<nil>")
		} else if sn, ok := n.(StringNode); ok {
			b.WriteString(string(sn))
		} else {
			b.WriteString("!notStringNode!")
		}
	}

	if b.String() != want {
		t.Errorf("got %q; want %q", b.String(), want)
	}
}

// checkNodeDepthIterator checks the iterator over node-depth pairs of the tree.
func checkNodeDepthIterator(
	t *testing.T,
	root tree.Node,
	seq2 iter.Seq2[tree.Node, int],
	want string,
) {
	t.Helper()

	if seq2 == nil {
		t.Error("got nil iterator")
		return
	}

	var (
		rootDepth int
		b         strings.Builder
	)
	if snRoot, ok := root.(StringNode); ok {
		rootDepth = snRoot.Depth()
	}

	for n, d := range seq2 {
		if n == nil {
			b.WriteString("<nil>")
			t.Errorf("got depth %d for a nil node", d)
		} else if sn, ok := n.(StringNode); ok {
			b.WriteString(string(sn))

			wantDepth := sn.Depth() - rootDepth
			if d != wantDepth {
				t.Errorf("got depth %d for node %q; want %d", d, sn, wantDepth)
			}
		} else {
			b.WriteString("!notStringNode!")
			t.Errorf("got depth %d for a node not of type StringNode", d)
		}
	}

	if b.String() != want {
		t.Errorf("got %q; want %q", b.String(), want)
	}
}

// checkStepNodeIterator checks the iterator over step-node pairs of the tree.
func checkStepNodeIterator(
	t *testing.T,
	seq2 iter.Seq2[int64, tree.Node],
	want string,
) {
	t.Helper()

	if seq2 == nil {
		t.Error("got nil iterator")
		return
	}

	var (
		wantStep int64
		b        strings.Builder
	)

	for i, n := range seq2 {
		wantStep++
		if i != wantStep {
			t.Errorf("got step %d; want %d", i, wantStep)
		}

		if n == nil {
			b.WriteString("<nil>")
		} else if sn, ok := n.(StringNode); ok {
			b.WriteString(string(sn))
		} else {
			b.WriteString("!notStringNode!")
		}
	}

	if b.String() != want {
		t.Errorf("got %q; want %q", b.String(), want)
	}
}

// checkPathIterator checks the iterator over paths
// from the root to nodes of the tree.
func checkPathIterator(
	t *testing.T,
	seq iter.Seq[vseq.VertexSequence[tree.Node]],
	want string,
) {
	t.Helper()

	if seq == nil {
		t.Error("got nil iterator")
		return
	}

	var b strings.Builder

	for p := range seq {
		if p == nil {
			b.WriteString("<nil>")
			continue
		}

		b.WriteByte('(')

		for n := range p.IterItemsBackward() {
			if n == nil {
				b.WriteString("<nil>")
			} else if sn, ok := n.(StringNode); ok {
				b.WriteString(string(sn))
			} else {
				b.WriteString("!notStringNode!")
			}
		}

		b.WriteByte(')')
	}

	if b.String() != want {
		t.Errorf("got %q; want %q", b.String(), want)
	}
}

// checkStepPathIterator checks the iterator over step-path pairs where
// the paths are from the root to nodes of the tree.
func checkStepPathIterator(
	t *testing.T,
	seq2 iter.Seq2[int64, vseq.VertexSequence[tree.Node]],
	want string,
) {
	t.Helper()

	if seq2 == nil {
		t.Error("got nil iterator")
		return
	}

	var (
		wantStep int64
		b        strings.Builder
	)

	for i, p := range seq2 {
		wantStep++
		if i != wantStep {
			t.Errorf("got step %d; want %d", i, wantStep)
		}

		if p == nil {
			b.WriteString("<nil>")
			continue
		}

		b.WriteByte('(')

		for n := range p.IterItemsBackward() {
			if n == nil {
				b.WriteString("<nil>")
			} else if sn, ok := n.(StringNode); ok {
				b.WriteString(string(sn))
			} else {
				b.WriteString("!notStringNode!")
			}
		}

		b.WriteByte(')')
	}

	if b.String() != want {
		t.Errorf("got %q; want %q", b.String(), want)
	}
}
