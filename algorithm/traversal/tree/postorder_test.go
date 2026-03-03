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
	"testing"

	"github.com/donyori/gogo/algorithm/traversal/tree"
)

var (
	// PostorderNodeIteratorTestCaseWants are 24 test case wanted outputs
	// for testing post-order node iterators,
	// corresponding to IteratorTestCaseInputs.
	PostorderNodeIteratorTestCaseWants = [NumIteratorTestCase]string{
		"",
		"L3aR3bL2aL3cL2bR3dL2cL3eL3fR3gR2dR3hR3iR2eL1aL1b" +
			"L2fL3jR3kL2gR2hL3lR3mR2iL1cR2jR2kR1dR3nL2lR1eT0a",
		"L1b",
		"L2fL3jR3kL2gR2hL3lR3mR2iL1c",

		"",
		"L3aR3bL2aL3cL2bR3dL2c",
		"L1b",
		"L2fL3jR3kL2gR2hL3lR3m",

		"",
		"L2aL2bL2cR2dR2eL1aL1bL2fL2gR2hR2iL1cR2jR2kR1dL2lR1eT0a",
		"L1b",
		"L2fL3jR3kL2gR2hL3lR3mR2iL1c",

		"",
		"L2aL2bL2cR2dR2eL1aL1b",
		"L1b",
		"L2fL3jR3kL2gR2hL3lR3m",

		"",
		"L3aR3bL2aL3cL2bR3dL2cL3eL3fR3gR2dR3hR3iR2eL1aL1b" +
			"L2fL3jR3kL2gR2hL3lR3mR2iL1cR2jR2kR1dR3nL2lR1eT0a",
		"L1b",
		"L2fL3jR3kL2gR2hL3lR3mR2iL1c",

		"",
		"L3aR3bL2aL3cL2bR3dL2cL3eL3fR3gR2dR3hR3iR2eL1aL1b" +
			"L2fL3jR3kL2gR2hL3lR3mR2iL1cR2jR2kR1dR3nL2lR1eT0a",
		"L1b",
		"L2fL3jR3kL2gR2hL3lR3mR2iL1c",
	}

	// PostorderPathIteratorTestCaseWants are 24 test case wanted outputs
	// for testing post-order path iterators,
	// corresponding to IteratorTestCaseInputs.
	PostorderPathIteratorTestCaseWants = [NumIteratorTestCase]string{
		"",
		"(L3aL2aL1aT0a)(R3bL2aL1aT0a)(L2aL1aT0a)(L3cL2bL1aT0a)(L2bL1aT0a)" +
			"(R3dL2cL1aT0a)(L2cL1aT0a)(L3eR2dL1aT0a)(L3fR2dL1aT0a)" +
			"(R3gR2dL1aT0a)(R2dL1aT0a)(R3hR2eL1aT0a)(R3iR2eL1aT0a)(R2eL1aT0a)" +
			"(L1aT0a)(L1bT0a)(L2fL1cT0a)(L3jL2gL1cT0a)(R3kL2gL1cT0a)" +
			"(L2gL1cT0a)(R2hL1cT0a)(L3lR2iL1cT0a)(R3mR2iL1cT0a)(R2iL1cT0a)" +
			"(L1cT0a)(R2jR1dT0a)(R2kR1dT0a)(R1dT0a)(R3nL2lR1eT0a)(L2lR1eT0a)" +
			"(R1eT0a)(T0a)",
		"(L1b)",
		"(L2fL1c)(L3jL2gL1c)(R3kL2gL1c)(L2gL1c)(R2hL1c)(L3lR2iL1c)(R3mR2iL1c)" +
			"(R2iL1c)(L1c)",

		"",
		"(L3aL2aL1aT0a)(R3bL2aL1aT0a)(L2aL1aT0a)(L3cL2bL1aT0a)(L2bL1aT0a)" +
			"(R3dL2cL1aT0a)(L2cL1aT0a)",
		"(L1b)",
		"(L2fL1c)(L3jL2gL1c)(R3kL2gL1c)(L2gL1c)(R2hL1c)(L3lR2iL1c)(R3mR2iL1c)",

		"",
		"(L2aL1aT0a)(L2bL1aT0a)(L2cL1aT0a)(R2dL1aT0a)(R2eL1aT0a)(L1aT0a)" +
			"(L1bT0a)(L2fL1cT0a)(L2gL1cT0a)(R2hL1cT0a)(R2iL1cT0a)(L1cT0a)" +
			"(R2jR1dT0a)(R2kR1dT0a)(R1dT0a)(L2lR1eT0a)(R1eT0a)(T0a)",
		"(L1b)",
		"(L2fL1c)(L3jL2gL1c)(R3kL2gL1c)(L2gL1c)(R2hL1c)(L3lR2iL1c)(R3mR2iL1c)" +
			"(R2iL1c)(L1c)",

		"",
		"(L2aL1aT0a)(L2bL1aT0a)(L2cL1aT0a)(R2dL1aT0a)(R2eL1aT0a)(L1aT0a)" +
			"(L1bT0a)",
		"(L1b)",
		"(L2fL1c)(L3jL2gL1c)(R3kL2gL1c)(L2gL1c)(R2hL1c)(L3lR2iL1c)(R3mR2iL1c)",

		"",
		"(L3aL2aL1aT0a)(R3bL2aL1aT0a)(L2aL1aT0a)(L3cL2bL1aT0a)(L2bL1aT0a)" +
			"(R3dL2cL1aT0a)(L2cL1aT0a)(L3eR2dL1aT0a)(L3fR2dL1aT0a)" +
			"(R3gR2dL1aT0a)(R2dL1aT0a)(R3hR2eL1aT0a)(R3iR2eL1aT0a)(R2eL1aT0a)" +
			"(L1aT0a)(L1bT0a)(L2fL1cT0a)(L3jL2gL1cT0a)(R3kL2gL1cT0a)" +
			"(L2gL1cT0a)(R2hL1cT0a)(L3lR2iL1cT0a)(R3mR2iL1cT0a)(R2iL1cT0a)" +
			"(L1cT0a)(R2jR1dT0a)(R2kR1dT0a)(R1dT0a)(R3nL2lR1eT0a)(L2lR1eT0a)" +
			"(R1eT0a)(T0a)",
		"(L1b)",
		"(L2fL1c)(L3jL2gL1c)(R3kL2gL1c)(L2gL1c)(R2hL1c)(L3lR2iL1c)(R3mR2iL1c)" +
			"(R2iL1c)(L1c)",

		"",
		"(L3aL2aL1aT0a)(R3bL2aL1aT0a)(L2aL1aT0a)(L3cL2bL1aT0a)(L2bL1aT0a)" +
			"(R3dL2cL1aT0a)(L2cL1aT0a)(L3eR2dL1aT0a)(L3fR2dL1aT0a)" +
			"(R3gR2dL1aT0a)(R2dL1aT0a)(R3hR2eL1aT0a)(R3iR2eL1aT0a)(R2eL1aT0a)" +
			"(L1aT0a)(L1bT0a)(L2fL1cT0a)(L3jL2gL1cT0a)(R3kL2gL1cT0a)" +
			"(L2gL1cT0a)(R2hL1cT0a)(L3lR2iL1cT0a)(R3mR2iL1cT0a)(R2iL1cT0a)" +
			"(L1cT0a)(R2jR1dT0a)(R2kR1dT0a)(R1dT0a)(R3nL2lR1eT0a)(L2lR1eT0a)" +
			"(R1eT0a)(T0a)",
		"(L1b)",
		"(L2fL1c)(L3jL2gL1c)(R3kL2gL1c)(L2gL1c)(R2hL1c)(L3lR2iL1c)(R3mR2iL1c)" +
			"(R2iL1c)(L1c)",
	}

	// ReversePostorderNodeIteratorTestCaseWants are 24 test case wanted outputs
	// for testing reverse post-order node iterators,
	// corresponding to IteratorTestCaseInputs.
	ReversePostorderNodeIteratorTestCaseWants = [NumIteratorTestCase]string{
		"",
		"R3nL2lR1eR2kR2jR1dR3mL3lR2iR2hR3kL3jL2gL2fL1cL1b" +
			"R3iR3hR2eR3gL3fL3eR2dR3dL2cL3cL2bR3bL3aL2aL1aT0a",
		"L1b",
		"R3mL3lR2iR2hR3kL3jL2gL2fL1c",

		"",
		"R3nL2lR1eR2kR2jR1dR3m",
		"L1b",
		"R3mL3lR2iR2hR3kL3jL2g",

		"",
		"L2lR1eR2kR2jR1dR2iR2hL2gL2fL1cL1bR2eR2dL2cL2bL2aL1aT0a",
		"L1b",
		"R3mL3lR2iR2hR3kL3jL2gL2fL1c",

		"",
		"L2lR1eR2kR2jR1dR2iR2h",
		"L1b",
		"R3mL3lR2iR2hR3kL3jL2g",

		"",
		"R3nL2lR1eR2kR2jR1dR3mL3lR2iR2hR3kL3jL2gL2fL1cL1b" +
			"R3iR3hR2eR3gL3fL3eR2dR3dL2cL3cL2bR3bL3aL2aL1aT0a",
		"L1b",
		"R3mL3lR2iR2hR3kL3jL2gL2fL1c",

		"",
		"R3nL2lR1eR2kR2jR1dR3mL3lR2iR2hR3kL3jL2gL2fL1cL1b" +
			"R3iR3hR2eR3gL3fL3eR2dR3dL2cL3cL2bR3bL3aL2aL1aT0a",
		"L1b",
		"R3mL3lR2iR2hR3kL3jL2gL2fL1c",
	}

	// ReversePostorderPathIteratorTestCaseWants are 24 test case wanted outputs
	// for testing reverse post-order path iterators,
	// corresponding to IteratorTestCaseInputs.
	ReversePostorderPathIteratorTestCaseWants = [NumIteratorTestCase]string{
		"",
		"(R3nL2lR1eT0a)(L2lR1eT0a)(R1eT0a)(R2kR1dT0a)(R2jR1dT0a)(R1dT0a)" +
			"(R3mR2iL1cT0a)(L3lR2iL1cT0a)(R2iL1cT0a)(R2hL1cT0a)(R3kL2gL1cT0a)" +
			"(L3jL2gL1cT0a)(L2gL1cT0a)(L2fL1cT0a)(L1cT0a)(L1bT0a)" +
			"(R3iR2eL1aT0a)(R3hR2eL1aT0a)(R2eL1aT0a)(R3gR2dL1aT0a)" +
			"(L3fR2dL1aT0a)(L3eR2dL1aT0a)(R2dL1aT0a)(R3dL2cL1aT0a)(L2cL1aT0a)" +
			"(L3cL2bL1aT0a)(L2bL1aT0a)(R3bL2aL1aT0a)(L3aL2aL1aT0a)(L2aL1aT0a)" +
			"(L1aT0a)(T0a)",
		"(L1b)",
		"(R3mR2iL1c)(L3lR2iL1c)(R2iL1c)(R2hL1c)(R3kL2gL1c)(L3jL2gL1c)(L2gL1c)" +
			"(L2fL1c)(L1c)",

		"",
		"(R3nL2lR1eT0a)(L2lR1eT0a)(R1eT0a)(R2kR1dT0a)(R2jR1dT0a)(R1dT0a)" +
			"(R3mR2iL1cT0a)",
		"(L1b)",
		"(R3mR2iL1c)(L3lR2iL1c)(R2iL1c)(R2hL1c)(R3kL2gL1c)(L3jL2gL1c)(L2gL1c)",

		"",
		"(L2lR1eT0a)(R1eT0a)(R2kR1dT0a)(R2jR1dT0a)(R1dT0a)(R2iL1cT0a)" +
			"(R2hL1cT0a)(L2gL1cT0a)(L2fL1cT0a)(L1cT0a)(L1bT0a)(R2eL1aT0a)" +
			"(R2dL1aT0a)(L2cL1aT0a)(L2bL1aT0a)(L2aL1aT0a)(L1aT0a)(T0a)",
		"(L1b)",
		"(R3mR2iL1c)(L3lR2iL1c)(R2iL1c)(R2hL1c)(R3kL2gL1c)(L3jL2gL1c)(L2gL1c)" +
			"(L2fL1c)(L1c)",

		"",
		"(L2lR1eT0a)(R1eT0a)(R2kR1dT0a)(R2jR1dT0a)(R1dT0a)(R2iL1cT0a)" +
			"(R2hL1cT0a)",
		"(L1b)",
		"(R3mR2iL1c)(L3lR2iL1c)(R2iL1c)(R2hL1c)(R3kL2gL1c)(L3jL2gL1c)(L2gL1c)",

		"",
		"(R3nL2lR1eT0a)(L2lR1eT0a)(R1eT0a)(R2kR1dT0a)(R2jR1dT0a)(R1dT0a)" +
			"(R3mR2iL1cT0a)(L3lR2iL1cT0a)(R2iL1cT0a)(R2hL1cT0a)(R3kL2gL1cT0a)" +
			"(L3jL2gL1cT0a)(L2gL1cT0a)(L2fL1cT0a)(L1cT0a)(L1bT0a)" +
			"(R3iR2eL1aT0a)(R3hR2eL1aT0a)(R2eL1aT0a)(R3gR2dL1aT0a)" +
			"(L3fR2dL1aT0a)(L3eR2dL1aT0a)(R2dL1aT0a)(R3dL2cL1aT0a)(L2cL1aT0a)" +
			"(L3cL2bL1aT0a)(L2bL1aT0a)(R3bL2aL1aT0a)(L3aL2aL1aT0a)(L2aL1aT0a)" +
			"(L1aT0a)(T0a)",
		"(L1b)",
		"(R3mR2iL1c)(L3lR2iL1c)(R2iL1c)(R2hL1c)(R3kL2gL1c)(L3jL2gL1c)(L2gL1c)" +
			"(L2fL1c)(L1c)",

		"",
		"(R3nL2lR1eT0a)(L2lR1eT0a)(R1eT0a)(R2kR1dT0a)(R2jR1dT0a)(R1dT0a)" +
			"(R3mR2iL1cT0a)(L3lR2iL1cT0a)(R2iL1cT0a)(R2hL1cT0a)(R3kL2gL1cT0a)" +
			"(L3jL2gL1cT0a)(L2gL1cT0a)(L2fL1cT0a)(L1cT0a)(L1bT0a)" +
			"(R3iR2eL1aT0a)(R3hR2eL1aT0a)(R2eL1aT0a)(R3gR2dL1aT0a)" +
			"(L3fR2dL1aT0a)(L3eR2dL1aT0a)(R2dL1aT0a)(R3dL2cL1aT0a)(L2cL1aT0a)" +
			"(L3cL2bL1aT0a)(L2bL1aT0a)(R3bL2aL1aT0a)(L3aL2aL1aT0a)(L2aL1aT0a)" +
			"(L1aT0a)(T0a)",
		"(L1b)",
		"(R3mR2iL1c)(L3lR2iL1c)(R2iL1c)(R2hL1c)(R3kL2gL1c)(L3jL2gL1c)(L2gL1c)" +
			"(L2fL1c)(L1c)",
	}
)

func TestIterPostorder(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				IteratorTestCaseInputs[i].NodeSkipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq := tree.IterPostorder(root, opts)
				checkNodeIterator(t, seq, PostorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterPostorderDepth(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				IteratorTestCaseInputs[i].NodeSkipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq2 := tree.IterPostorderDepth(root, opts)
				checkNodeDepthIterator(
					t, root, seq2, PostorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterPostorderStep(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				IteratorTestCaseInputs[i].NodeSkipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq2 := tree.IterPostorderStep(root, opts)
				checkStepNodeIterator(
					t, seq2, PostorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterPostorderPath(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				IteratorTestCaseInputs[i].PathSkipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq := tree.IterPostorderPath(root, opts)
				checkPathIterator(t, seq, PostorderPathIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterPostorderPathStep(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				IteratorTestCaseInputs[i].PathSkipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq2 := tree.IterPostorderPathStep(root, opts)
				checkStepPathIterator(
					t, seq2, PostorderPathIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReversePostorder(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				IteratorTestCaseInputs[i].NodeSkipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq := tree.IterReversePostorder(root, opts)
				checkNodeIterator(
					t, seq, ReversePostorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReversePostorderDepth(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				IteratorTestCaseInputs[i].NodeSkipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq2 := tree.IterReversePostorderDepth(root, opts)
				checkNodeDepthIterator(
					t, root, seq2, ReversePostorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReversePostorderStep(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				IteratorTestCaseInputs[i].NodeSkipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq2 := tree.IterReversePostorderStep(root, opts)
				checkStepNodeIterator(
					t, seq2, ReversePostorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReversePostorderPath(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				IteratorTestCaseInputs[i].PathSkipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq := tree.IterReversePostorderPath(root, opts)
				checkPathIterator(
					t, seq, ReversePostorderPathIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReversePostorderPathStep(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				IteratorTestCaseInputs[i].PathSkipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq2 := tree.IterReversePostorderPathStep(root, opts)
				checkStepPathIterator(
					t, seq2, ReversePostorderPathIteratorTestCaseWants[i])
			},
		)
	}
}
