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
	// InorderNodeIteratorTestCaseWants are 24 test case wanted outputs
	// for testing in-order node iterators,
	// corresponding to IteratorTestCaseInputs.
	InorderNodeIteratorTestCaseWants = [NumIteratorTestCase]string{
		"",
		"L3aL2aR3bL3cL2bL2cR3dL1aL3eL3fR2dR3gR2eR3hR3iL1b" +
			"L2fL3jL2gR3kL1cR2hL3lR2iR3mT0aR1dR2jR2kL2lR3nR1e",
		"L1b",
		"L2fL3jL2gR3kL1cR2hL3lR2iR3m",

		"",
		"L3aL2aR3bL3cL2bL2cR3d",
		"L1b",
		"L2fL3jL2gR3kL1cR2hL3l",

		"",
		"L2aL2bL2cL1aR2dR2eL1bL2fL2gL1cR2hR2iT0aR1dR2jR2kL2lR1e",
		"L1b",
		"L2fL3jL2gR3kL1cR2hL3lR2iR3m",

		"",
		"L2aL2bL2cL1aR2dR2eL1b",
		"L1b",
		"L2fL3jL2gR3kL1cR2hL3l",

		"",
		"L3aL2aR3bL3cL2bL2cR3dL1aL3eL3fR2dR3gR2eR3hR3iL1b" +
			"L2fL3jL2gR3kL1cR2hL3lR2iR3mT0aR1dR2jR2kL2lR3nR1e",
		"L1b",
		"L2fL3jL2gR3kL1cR2hL3lR2iR3m",

		"",
		"L3aL2aR3bL3cL2bL2cR3dL1aL3eL3fR2dR2eR3h" +
			"R3iL1bL2fL3jL2gR3kL1cT0aR1dR2jR2kL2lR1e",
		"L1b",
		"L2fL3jL2gR3kL1c",
	}

	// InorderPathIteratorTestCaseWants are 24 test case wanted outputs
	// for testing in-order path iterators,
	// corresponding to IteratorTestCaseInputs.
	InorderPathIteratorTestCaseWants = [NumIteratorTestCase]string{
		"",
		"(L3aL2aL1aT0a)(L2aL1aT0a)(R3bL2aL1aT0a)(L3cL2bL1aT0a)(L2bL1aT0a)" +
			"(L2cL1aT0a)(R3dL2cL1aT0a)(L1aT0a)(L3eR2dL1aT0a)(L3fR2dL1aT0a)" +
			"(R2dL1aT0a)(R3gR2dL1aT0a)(R2eL1aT0a)(R3hR2eL1aT0a)(R3iR2eL1aT0a)" +
			"(L1bT0a)(L2fL1cT0a)(L3jL2gL1cT0a)(L2gL1cT0a)(R3kL2gL1cT0a)" +
			"(L1cT0a)(R2hL1cT0a)(L3lR2iL1cT0a)(R2iL1cT0a)(R3mR2iL1cT0a)(T0a)" +
			"(R1dT0a)(R2jR1dT0a)(R2kR1dT0a)(L2lR1eT0a)(R3nL2lR1eT0a)(R1eT0a)",
		"(L1b)",
		"(L2fL1c)(L3jL2gL1c)(L2gL1c)(R3kL2gL1c)(L1c)(R2hL1c)(L3lR2iL1c)" +
			"(R2iL1c)(R3mR2iL1c)",

		"",
		"(L3aL2aL1aT0a)(L2aL1aT0a)(R3bL2aL1aT0a)(L3cL2bL1aT0a)(L2bL1aT0a)" +
			"(L2cL1aT0a)(R3dL2cL1aT0a)",
		"(L1b)",
		"(L2fL1c)(L3jL2gL1c)(L2gL1c)(R3kL2gL1c)(L1c)(R2hL1c)(L3lR2iL1c)",

		"",
		"(L2aL1aT0a)(L2bL1aT0a)(L2cL1aT0a)(L1aT0a)(R2dL1aT0a)(R2eL1aT0a)" +
			"(L1bT0a)(L2fL1cT0a)(L2gL1cT0a)(L1cT0a)(R2hL1cT0a)(R2iL1cT0a)" +
			"(T0a)(R1dT0a)(R2jR1dT0a)(R2kR1dT0a)(L2lR1eT0a)(R1eT0a)",
		"(L1b)",
		"(L2fL1c)(L3jL2gL1c)(L2gL1c)(R3kL2gL1c)(L1c)(R2hL1c)(L3lR2iL1c)" +
			"(R2iL1c)(R3mR2iL1c)",

		"",
		"(L2aL1aT0a)(L2bL1aT0a)(L2cL1aT0a)(L1aT0a)(R2dL1aT0a)(R2eL1aT0a)" +
			"(L1bT0a)",
		"(L1b)",
		"(L2fL1c)(L3jL2gL1c)(L2gL1c)(R3kL2gL1c)(L1c)(R2hL1c)(L3lR2iL1c)",

		"",
		"(L3aL2aL1aT0a)(L2aL1aT0a)(R3bL2aL1aT0a)(L3cL2bL1aT0a)(L2bL1aT0a)" +
			"(L2cL1aT0a)(R3dL2cL1aT0a)(L1aT0a)(L3eR2dL1aT0a)(L3fR2dL1aT0a)" +
			"(R2dL1aT0a)(R3gR2dL1aT0a)(R2eL1aT0a)(R3hR2eL1aT0a)(R3iR2eL1aT0a)" +
			"(L1bT0a)(L2fL1cT0a)(L3jL2gL1cT0a)(L2gL1cT0a)(R3kL2gL1cT0a)" +
			"(L1cT0a)(R2hL1cT0a)(L3lR2iL1cT0a)(R2iL1cT0a)(R3mR2iL1cT0a)(T0a)" +
			"(R1dT0a)(R2jR1dT0a)(R2kR1dT0a)(L2lR1eT0a)(R3nL2lR1eT0a)(R1eT0a)",
		"(L1b)",
		"(L2fL1c)(L3jL2gL1c)(L2gL1c)(R3kL2gL1c)(L1c)(R2hL1c)(L3lR2iL1c)" +
			"(R2iL1c)(R3mR2iL1c)",

		"",
		"(L3aL2aL1aT0a)(L2aL1aT0a)(R3bL2aL1aT0a)(L3cL2bL1aT0a)(L2bL1aT0a)" +
			"(L2cL1aT0a)(R3dL2cL1aT0a)(L1aT0a)(L3eR2dL1aT0a)(L3fR2dL1aT0a)" +
			"(R2dL1aT0a)(R2eL1aT0a)(R3hR2eL1aT0a)(R3iR2eL1aT0a)(L1bT0a)" +
			"(L2fL1cT0a)(L3jL2gL1cT0a)(L2gL1cT0a)(R3kL2gL1cT0a)(L1cT0a)(T0a)" +
			"(R1dT0a)(R2jR1dT0a)(R2kR1dT0a)(L2lR1eT0a)(R1eT0a)",
		"(L1b)",
		"(L2fL1c)(L3jL2gL1c)(L2gL1c)(R3kL2gL1c)(L1c)",
	}

	// ReverseInorderNodeIteratorTestCaseWants are 24 test case wanted outputs
	// for testing reverse in-order node iterators,
	// corresponding to IteratorTestCaseInputs.
	ReverseInorderNodeIteratorTestCaseWants = [NumIteratorTestCase]string{
		"",
		"R1eR3nL2lR2kR2jR1dT0aR3mR2iL3lR2hL1cR3kL2gL3jL2f" +
			"L1bR3iR3hR2eR3gR2dL3fL3eL1aR3dL2cL2bL3cR3bL2aL3a",
		"L1b",
		"R3mR2iL3lR2hL1cR3kL2gL3jL2f",

		"",
		"R1eR3nL2lR2kR2jR1dT0a",
		"L1b",
		"R3mR2iL3lR2hL1cR3kL2g",

		"",
		"R1eL2lR2kR2jR1dT0aR2iR2hL1cL2gL2fL1bR2eR2dL1aL2cL2bL2a",
		"L1b",
		"R3mR2iL3lR2hL1cR3kL2gL3jL2f",

		"",
		"R1eL2lR2kR2jR1dT0aR2i",
		"L1b",
		"R3mR2iL3lR2hL1cR3kL2g",

		"",
		"R1eR3nL2lR2kR2jR1dT0aR3mR2iL3lR2hL1cR3kL2gL3jL2f" +
			"L1bR3iR3hR2eR3gR2dL3fL3eL1aR3dL2cL2bL3cR3bL2aL3a",
		"L1b",
		"R3mR2iL3lR2hL1cR3kL2gL3jL2f",

		"",
		"R1eR3nL2lR2kR2jR1dT0aR3mR2iL3lR2hL1cL1b" +
			"R3iR3hR2eR3gR2dL1aR3dL2cL2bL3cR3bL2aL3a",
		"L1b",
		"R3mR2iL3lR2hL1c",
	}

	// ReverseInorderPathIteratorTestCaseWants are 24 test case wanted outputs
	// for testing reverse in-order path iterators,
	// corresponding to IteratorTestCaseInputs.
	ReverseInorderPathIteratorTestCaseWants = [NumIteratorTestCase]string{
		"",
		"(R1eT0a)(R3nL2lR1eT0a)(L2lR1eT0a)(R2kR1dT0a)(R2jR1dT0a)(R1dT0a)(T0a)" +
			"(R3mR2iL1cT0a)(R2iL1cT0a)(L3lR2iL1cT0a)(R2hL1cT0a)(L1cT0a)" +
			"(R3kL2gL1cT0a)(L2gL1cT0a)(L3jL2gL1cT0a)(L2fL1cT0a)(L1bT0a)" +
			"(R3iR2eL1aT0a)(R3hR2eL1aT0a)(R2eL1aT0a)(R3gR2dL1aT0a)(R2dL1aT0a)" +
			"(L3fR2dL1aT0a)(L3eR2dL1aT0a)(L1aT0a)(R3dL2cL1aT0a)(L2cL1aT0a)" +
			"(L2bL1aT0a)(L3cL2bL1aT0a)(R3bL2aL1aT0a)(L2aL1aT0a)(L3aL2aL1aT0a)",
		"(L1b)",
		"(R3mR2iL1c)(R2iL1c)(L3lR2iL1c)(R2hL1c)(L1c)(R3kL2gL1c)(L2gL1c)" +
			"(L3jL2gL1c)(L2fL1c)",

		"",
		"(R1eT0a)(R3nL2lR1eT0a)(L2lR1eT0a)(R2kR1dT0a)(R2jR1dT0a)(R1dT0a)(T0a)",
		"(L1b)",
		"(R3mR2iL1c)(R2iL1c)(L3lR2iL1c)(R2hL1c)(L1c)(R3kL2gL1c)(L2gL1c)",

		"",
		"(R1eT0a)(L2lR1eT0a)(R2kR1dT0a)(R2jR1dT0a)(R1dT0a)(T0a)(R2iL1cT0a)" +
			"(R2hL1cT0a)(L1cT0a)(L2gL1cT0a)(L2fL1cT0a)(L1bT0a)(R2eL1aT0a)" +
			"(R2dL1aT0a)(L1aT0a)(L2cL1aT0a)(L2bL1aT0a)(L2aL1aT0a)",
		"(L1b)",
		"(R3mR2iL1c)(R2iL1c)(L3lR2iL1c)(R2hL1c)(L1c)(R3kL2gL1c)(L2gL1c)" +
			"(L3jL2gL1c)(L2fL1c)",

		"",
		"(R1eT0a)(L2lR1eT0a)(R2kR1dT0a)(R2jR1dT0a)(R1dT0a)(T0a)(R2iL1cT0a)",
		"(L1b)",
		"(R3mR2iL1c)(R2iL1c)(L3lR2iL1c)(R2hL1c)(L1c)(R3kL2gL1c)(L2gL1c)",

		"",
		"(R1eT0a)(R3nL2lR1eT0a)(L2lR1eT0a)(R2kR1dT0a)(R2jR1dT0a)(R1dT0a)(T0a)" +
			"(R3mR2iL1cT0a)(R2iL1cT0a)(L3lR2iL1cT0a)(R2hL1cT0a)(L1cT0a)" +
			"(R3kL2gL1cT0a)(L2gL1cT0a)(L3jL2gL1cT0a)(L2fL1cT0a)(L1bT0a)" +
			"(R3iR2eL1aT0a)(R3hR2eL1aT0a)(R2eL1aT0a)(R3gR2dL1aT0a)(R2dL1aT0a)" +
			"(L3fR2dL1aT0a)(L3eR2dL1aT0a)(L1aT0a)(R3dL2cL1aT0a)(L2cL1aT0a)" +
			"(L2bL1aT0a)(L3cL2bL1aT0a)(R3bL2aL1aT0a)(L2aL1aT0a)(L3aL2aL1aT0a)",
		"(L1b)",
		"(R3mR2iL1c)(R2iL1c)(L3lR2iL1c)(R2hL1c)(L1c)(R3kL2gL1c)(L2gL1c)" +
			"(L3jL2gL1c)(L2fL1c)",

		"",
		"(R1eT0a)(R3nL2lR1eT0a)(L2lR1eT0a)(R2kR1dT0a)(R2jR1dT0a)(R1dT0a)(T0a)" +
			"(R3mR2iL1cT0a)(R2iL1cT0a)(L3lR2iL1cT0a)(R2hL1cT0a)(L1cT0a)" +
			"(L1bT0a)(R3iR2eL1aT0a)(R3hR2eL1aT0a)(R2eL1aT0a)(R3gR2dL1aT0a)" +
			"(R2dL1aT0a)(L1aT0a)(R3dL2cL1aT0a)(L2cL1aT0a)(L2bL1aT0a)" +
			"(L3cL2bL1aT0a)(R3bL2aL1aT0a)(L2aL1aT0a)(L3aL2aL1aT0a)",
		"(L1b)",
		"(R3mR2iL1c)(R2iL1c)(L3lR2iL1c)(R2hL1c)(L1c)",
	}
)

func TestIterInorder(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		skipChildrenFn := IteratorTestCaseInputs[i].NodeSkipChildrenFn
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				skipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq := tree.IterInorder(root, opts, skipChildrenFn)
				checkNodeIterator(t, seq, InorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterInorderDepth(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		skipChildrenFn := IteratorTestCaseInputs[i].NodeSkipChildrenFn
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				skipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq2 := tree.IterInorderDepth(root, opts, skipChildrenFn)
				checkNodeDepthIterator(
					t, root, seq2, InorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterInorderStep(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		skipChildrenFn := IteratorTestCaseInputs[i].NodeSkipChildrenFn
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				skipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq2 := tree.IterInorderStep(root, opts, skipChildrenFn)
				checkStepNodeIterator(
					t, seq2, InorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterInorderPath(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		skipChildrenFn := IteratorTestCaseInputs[i].PathSkipChildrenFn
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				skipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq := tree.IterInorderPath(root, opts, skipChildrenFn)
				checkPathIterator(t, seq, InorderPathIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterInorderPathStep(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		skipChildrenFn := IteratorTestCaseInputs[i].PathSkipChildrenFn
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				skipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq2 := tree.IterInorderPathStep(root, opts, skipChildrenFn)
				checkStepPathIterator(
					t, seq2, InorderPathIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReverseInorder(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		skipChildrenFn := IteratorTestCaseInputs[i].NodeSkipChildrenFn
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				skipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq := tree.IterReverseInorder(root, opts, skipChildrenFn)
				checkNodeIterator(
					t, seq, ReverseInorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReverseInorderDepth(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		skipChildrenFn := IteratorTestCaseInputs[i].NodeSkipChildrenFn
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				skipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq2 := tree.IterReverseInorderDepth(root, opts, skipChildrenFn)
				checkNodeDepthIterator(
					t, root, seq2, ReverseInorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReverseInorderStep(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		skipChildrenFn := IteratorTestCaseInputs[i].NodeSkipChildrenFn
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				skipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq2 := tree.IterReverseInorderStep(root, opts, skipChildrenFn)
				checkStepNodeIterator(
					t, seq2, ReverseInorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReverseInorderPath(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		skipChildrenFn := IteratorTestCaseInputs[i].PathSkipChildrenFn
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				skipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq := tree.IterReverseInorderPath(root, opts, skipChildrenFn)
				checkPathIterator(
					t, seq, ReverseInorderPathIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReverseInorderPathStep(t *testing.T) {
	t.Parallel()

	for i := range NumIteratorTestCase {
		root := IteratorTestCaseInputs[i].Root
		opts := IteratorTestCaseInputs[i].Opts
		skipChildrenFn := IteratorTestCaseInputs[i].PathSkipChildrenFn
		t.Run(
			fmt.Sprintf(
				"root=%s&opts=%s&skipChildren=%t",
				stringNodeToName(root),
				optionsToName(opts),
				skipChildrenFn != nil,
			),
			func(t *testing.T) {
				t.Parallel()

				seq2 := tree.IterReverseInorderPathStep(
					root, opts, skipChildrenFn)
				checkStepPathIterator(
					t, seq2, ReverseInorderPathIteratorTestCaseWants[i])
			},
		)
	}
}
