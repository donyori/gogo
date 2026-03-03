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
	// PreorderNodeIteratorTestCaseWants are 24 test case wanted outputs
	// for testing pre-order node iterators,
	// corresponding to IteratorTestCaseInputs.
	PreorderNodeIteratorTestCaseWants = [NumIteratorTestCase]string{
		"",
		"T0aL1aL2aL3aR3bL2bL3cL2cR3dR2dL3eL3fR3gR2eR3hR3i" +
			"L1bL1cL2fL2gL3jR3kR2hR2iL3lR3mR1dR2jR2kR1eL2lR3n",
		"L1b",
		"L1cL2fL2gL3jR3kR2hR2iL3lR3m",

		"",
		"T0aL1aL2aL3aR3bL2bL3c",
		"L1b",
		"L1cL2fL2gL3jR3kR2hR2i",

		"",
		"T0aL1aL2aL2bL2cR2dR2eL1bL1cL2fL2gR2hR2iR1dR2jR2kR1eL2l",
		"L1b",
		"L1cL2fL2gL3jR3kR2hR2iL3lR3m",

		"",
		"T0aL1aL2aL2bL2cR2dR2e",
		"L1b",
		"L1cL2fL2gL3jR3kR2hR2i",

		"",
		"T0aL1aL2aL3aR3bL2bL3cL2cR3dR2dL3eL3fR3gR2eR3hR3i" +
			"L1bL1cL2fL2gL3jR3kR2hR2iL3lR3mR1dR2jR2kR1eL2lR3n",
		"L1b",
		"L1cL2fL2gL3jR3kR2hR2iL3lR3m",

		"",
		"T0aL1aL2aL3aR3bL2bL3cL2cR3dR2dR2eR3hR3iL1bL1cR1dR2jR2kR1eL2l",
		"L1b",
		"L1c",
	}

	// PreorderPathIteratorTestCaseWants are 24 test case wanted outputs
	// for testing pre-order path iterators,
	// corresponding to IteratorTestCaseInputs.
	PreorderPathIteratorTestCaseWants = [NumIteratorTestCase]string{
		"",
		"(T0a)(L1aT0a)(L2aL1aT0a)(L3aL2aL1aT0a)(R3bL2aL1aT0a)(L2bL1aT0a)" +
			"(L3cL2bL1aT0a)(L2cL1aT0a)(R3dL2cL1aT0a)(R2dL1aT0a)(L3eR2dL1aT0a)" +
			"(L3fR2dL1aT0a)(R3gR2dL1aT0a)(R2eL1aT0a)(R3hR2eL1aT0a)" +
			"(R3iR2eL1aT0a)(L1bT0a)(L1cT0a)(L2fL1cT0a)(L2gL1cT0a)" +
			"(L3jL2gL1cT0a)(R3kL2gL1cT0a)(R2hL1cT0a)(R2iL1cT0a)(L3lR2iL1cT0a)" +
			"(R3mR2iL1cT0a)(R1dT0a)(R2jR1dT0a)(R2kR1dT0a)(R1eT0a)(L2lR1eT0a)" +
			"(R3nL2lR1eT0a)",
		"(L1b)",
		"(L1c)(L2fL1c)(L2gL1c)(L3jL2gL1c)(R3kL2gL1c)(R2hL1c)(R2iL1c)" +
			"(L3lR2iL1c)(R3mR2iL1c)",

		"",
		"(T0a)(L1aT0a)(L2aL1aT0a)(L3aL2aL1aT0a)(R3bL2aL1aT0a)(L2bL1aT0a)" +
			"(L3cL2bL1aT0a)",
		"(L1b)",
		"(L1c)(L2fL1c)(L2gL1c)(L3jL2gL1c)(R3kL2gL1c)(R2hL1c)(R2iL1c)",

		"",
		"(T0a)(L1aT0a)(L2aL1aT0a)(L2bL1aT0a)(L2cL1aT0a)(R2dL1aT0a)(R2eL1aT0a)" +
			"(L1bT0a)(L1cT0a)(L2fL1cT0a)(L2gL1cT0a)(R2hL1cT0a)(R2iL1cT0a)" +
			"(R1dT0a)(R2jR1dT0a)(R2kR1dT0a)(R1eT0a)(L2lR1eT0a)",
		"(L1b)",
		"(L1c)(L2fL1c)(L2gL1c)(L3jL2gL1c)(R3kL2gL1c)(R2hL1c)(R2iL1c)" +
			"(L3lR2iL1c)(R3mR2iL1c)",

		"",
		"(T0a)(L1aT0a)(L2aL1aT0a)(L2bL1aT0a)(L2cL1aT0a)(R2dL1aT0a)(R2eL1aT0a)",
		"(L1b)",
		"(L1c)(L2fL1c)(L2gL1c)(L3jL2gL1c)(R3kL2gL1c)(R2hL1c)(R2iL1c)",

		"",
		"(T0a)(L1aT0a)(L2aL1aT0a)(L3aL2aL1aT0a)(R3bL2aL1aT0a)(L2bL1aT0a)" +
			"(L3cL2bL1aT0a)(L2cL1aT0a)(R3dL2cL1aT0a)(R2dL1aT0a)(L3eR2dL1aT0a)" +
			"(L3fR2dL1aT0a)(R3gR2dL1aT0a)(R2eL1aT0a)(R3hR2eL1aT0a)" +
			"(R3iR2eL1aT0a)(L1bT0a)(L1cT0a)(L2fL1cT0a)(L2gL1cT0a)" +
			"(L3jL2gL1cT0a)(R3kL2gL1cT0a)(R2hL1cT0a)(R2iL1cT0a)(L3lR2iL1cT0a)" +
			"(R3mR2iL1cT0a)(R1dT0a)(R2jR1dT0a)(R2kR1dT0a)(R1eT0a)(L2lR1eT0a)" +
			"(R3nL2lR1eT0a)",
		"(L1b)",
		"(L1c)(L2fL1c)(L2gL1c)(L3jL2gL1c)(R3kL2gL1c)(R2hL1c)(R2iL1c)" +
			"(L3lR2iL1c)(R3mR2iL1c)",

		"",
		"(T0a)(L1aT0a)(L2aL1aT0a)(L3aL2aL1aT0a)(R3bL2aL1aT0a)(L2bL1aT0a)" +
			"(L3cL2bL1aT0a)(L2cL1aT0a)(R3dL2cL1aT0a)(R2dL1aT0a)(R2eL1aT0a)" +
			"(R3hR2eL1aT0a)(R3iR2eL1aT0a)(L1bT0a)(L1cT0a)(R1dT0a)(R2jR1dT0a)" +
			"(R2kR1dT0a)(R1eT0a)(L2lR1eT0a)",
		"(L1b)",
		"(L1c)",
	}

	// ReversePreorderNodeIteratorTestCaseWants are 24 test case wanted outputs
	// for testing reverse pre-order node iterators,
	// corresponding to IteratorTestCaseInputs.
	ReversePreorderNodeIteratorTestCaseWants = [NumIteratorTestCase]string{
		"",
		"T0aR1eL2lR3nR1dR2kR2jL1cR2iR3mL3lR2hL2gR3kL3jL2f" +
			"L1bL1aR2eR3iR3hR2dR3gL3fL3eL2cR3dL2bL3cL2aR3bL3a",
		"L1b",
		"L1cR2iR3mL3lR2hL2gR3kL3jL2f",

		"",
		"T0aR1eL2lR3nR1dR2kR2j",
		"L1b",
		"L1cR2iR3mL3lR2hL2gR3k",

		"",
		"T0aR1eL2lR1dR2kR2jL1cR2iR2hL2gL2fL1bL1aR2eR2dL2cL2bL2a",
		"L1b",
		"L1cR2iR3mL3lR2hL2gR3kL3jL2f",

		"",
		"T0aR1eL2lR1dR2kR2jL1c",
		"L1b",
		"L1cR2iR3mL3lR2hL2gR3k",

		"",
		"T0aR1eL2lR3nR1dR2kR2jL1cR2iR3mL3lR2hL2gR3kL3jL2f" +
			"L1bL1aR2eR3iR3hR2dR3gL3fL3eL2cR3dL2bL3cL2aR3bL3a",
		"L1b",
		"L1cR2iR3mL3lR2hL2gR3kL3jL2f",

		"",
		"T0aR1eL2lR1dR2kR2jL1cL1bL1aR2eR3iR3hR2dL2cR3dL2bL3cL2aR3bL3a",
		"L1b",
		"L1c",
	}

	// ReversePreorderPathIteratorTestCaseWants are 24 test case wanted outputs
	// for testing reverse pre-order path iterators,
	// corresponding to IteratorTestCaseInputs.
	ReversePreorderPathIteratorTestCaseWants = [NumIteratorTestCase]string{
		"",
		"(T0a)(R1eT0a)(L2lR1eT0a)(R3nL2lR1eT0a)(R1dT0a)(R2kR1dT0a)(R2jR1dT0a)" +
			"(L1cT0a)(R2iL1cT0a)(R3mR2iL1cT0a)(L3lR2iL1cT0a)(R2hL1cT0a)" +
			"(L2gL1cT0a)(R3kL2gL1cT0a)(L3jL2gL1cT0a)(L2fL1cT0a)(L1bT0a)" +
			"(L1aT0a)(R2eL1aT0a)(R3iR2eL1aT0a)(R3hR2eL1aT0a)(R2dL1aT0a)" +
			"(R3gR2dL1aT0a)(L3fR2dL1aT0a)(L3eR2dL1aT0a)(L2cL1aT0a)" +
			"(R3dL2cL1aT0a)(L2bL1aT0a)(L3cL2bL1aT0a)(L2aL1aT0a)(R3bL2aL1aT0a)" +
			"(L3aL2aL1aT0a)",
		"(L1b)",
		"(L1c)(R2iL1c)(R3mR2iL1c)(L3lR2iL1c)(R2hL1c)(L2gL1c)(R3kL2gL1c)" +
			"(L3jL2gL1c)(L2fL1c)",

		"",
		"(T0a)(R1eT0a)(L2lR1eT0a)(R3nL2lR1eT0a)(R1dT0a)(R2kR1dT0a)(R2jR1dT0a)",
		"(L1b)",
		"(L1c)(R2iL1c)(R3mR2iL1c)(L3lR2iL1c)(R2hL1c)(L2gL1c)(R3kL2gL1c)",

		"",
		"(T0a)(R1eT0a)(L2lR1eT0a)(R1dT0a)(R2kR1dT0a)(R2jR1dT0a)(L1cT0a)" +
			"(R2iL1cT0a)(R2hL1cT0a)(L2gL1cT0a)(L2fL1cT0a)(L1bT0a)(L1aT0a)" +
			"(R2eL1aT0a)(R2dL1aT0a)(L2cL1aT0a)(L2bL1aT0a)(L2aL1aT0a)",
		"(L1b)",
		"(L1c)(R2iL1c)(R3mR2iL1c)(L3lR2iL1c)(R2hL1c)(L2gL1c)(R3kL2gL1c)" +
			"(L3jL2gL1c)(L2fL1c)",

		"",
		"(T0a)(R1eT0a)(L2lR1eT0a)(R1dT0a)(R2kR1dT0a)(R2jR1dT0a)(L1cT0a)",
		"(L1b)",
		"(L1c)(R2iL1c)(R3mR2iL1c)(L3lR2iL1c)(R2hL1c)(L2gL1c)(R3kL2gL1c)",

		"",
		"(T0a)(R1eT0a)(L2lR1eT0a)(R3nL2lR1eT0a)(R1dT0a)(R2kR1dT0a)(R2jR1dT0a)" +
			"(L1cT0a)(R2iL1cT0a)(R3mR2iL1cT0a)(L3lR2iL1cT0a)(R2hL1cT0a)" +
			"(L2gL1cT0a)(R3kL2gL1cT0a)(L3jL2gL1cT0a)(L2fL1cT0a)(L1bT0a)" +
			"(L1aT0a)(R2eL1aT0a)(R3iR2eL1aT0a)(R3hR2eL1aT0a)(R2dL1aT0a)" +
			"(R3gR2dL1aT0a)(L3fR2dL1aT0a)(L3eR2dL1aT0a)(L2cL1aT0a)" +
			"(R3dL2cL1aT0a)(L2bL1aT0a)(L3cL2bL1aT0a)(L2aL1aT0a)(R3bL2aL1aT0a)" +
			"(L3aL2aL1aT0a)",
		"(L1b)",
		"(L1c)(R2iL1c)(R3mR2iL1c)(L3lR2iL1c)(R2hL1c)(L2gL1c)(R3kL2gL1c)" +
			"(L3jL2gL1c)(L2fL1c)",

		"",
		"(T0a)(R1eT0a)(L2lR1eT0a)(R1dT0a)(R2kR1dT0a)(R2jR1dT0a)(L1cT0a)" +
			"(L1bT0a)(L1aT0a)(R2eL1aT0a)(R3iR2eL1aT0a)(R3hR2eL1aT0a)" +
			"(R2dL1aT0a)(L2cL1aT0a)(R3dL2cL1aT0a)(L2bL1aT0a)(L3cL2bL1aT0a)" +
			"(L2aL1aT0a)(R3bL2aL1aT0a)(L3aL2aL1aT0a)",
		"(L1b)",
		"(L1c)",
	}
)

func TestIterPreorder(t *testing.T) {
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

				seq := tree.IterPreorder(root, opts, skipChildrenFn)
				checkNodeIterator(t, seq, PreorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterPreorderDepth(t *testing.T) {
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

				seq2 := tree.IterPreorderDepth(root, opts, skipChildrenFn)
				checkNodeDepthIterator(
					t, root, seq2, PreorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterPreorderStep(t *testing.T) {
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

				seq2 := tree.IterPreorderStep(root, opts, skipChildrenFn)
				checkStepNodeIterator(
					t, seq2, PreorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterPreorderPath(t *testing.T) {
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

				seq := tree.IterPreorderPath(root, opts, skipChildrenFn)
				checkPathIterator(t, seq, PreorderPathIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterPreorderPathStep(t *testing.T) {
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

				seq2 := tree.IterPreorderPathStep(root, opts, skipChildrenFn)
				checkStepPathIterator(
					t, seq2, PreorderPathIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReversePreorder(t *testing.T) {
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

				seq := tree.IterReversePreorder(root, opts, skipChildrenFn)
				checkNodeIterator(
					t, seq, ReversePreorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReversePreorderDepth(t *testing.T) {
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

				seq2 := tree.IterReversePreorderDepth(
					root, opts, skipChildrenFn)
				checkNodeDepthIterator(
					t, root, seq2, ReversePreorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReversePreorderStep(t *testing.T) {
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

				seq2 := tree.IterReversePreorderStep(root, opts, skipChildrenFn)
				checkStepNodeIterator(
					t, seq2, ReversePreorderNodeIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReversePreorderPath(t *testing.T) {
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

				seq := tree.IterReversePreorderPath(root, opts, skipChildrenFn)
				checkPathIterator(
					t, seq, ReversePreorderPathIteratorTestCaseWants[i])
			},
		)
	}
}

func TestIterReversePreorderPathStep(t *testing.T) {
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

				seq2 := tree.IterReversePreorderPathStep(
					root, opts, skipChildrenFn)
				checkStepPathIterator(
					t, seq2, ReversePreorderPathIteratorTestCaseWants[i])
			},
		)
	}
}
