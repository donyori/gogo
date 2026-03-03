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

package vseq

import (
	"iter"

	"github.com/donyori/gogo/container"
	"github.com/donyori/gogo/errors"
)

// VertexSequence represents a sequence of vertices/nodes in a graph.
//
// Its method Range accesses the vertices from first to last.
//
// The sequence is immutable.
// The methods Append, TruncateBack, Truncate, ReplaceBack, and Replace
// can create new sequences based on existing sequences,
// but must not modify existing sequences.
type VertexSequence[Vertex any] interface {
	container.Container[Vertex]

	// RangeBackward is like Range,
	// but the order of access is from last to first.
	RangeBackward(handler func(v Vertex) (cont bool))

	// IterItemsBackward returns an iterator over all vertices in the sequence,
	// traversing it from last to first.
	//
	// The returned iterator is never nil.
	IterItemsBackward() iter.Seq[Vertex]

	// IterIndexItems returns an iterator over index-vertex pairs
	// in the sequence, traversing it from first to last.
	//
	// The returned iterator is never nil.
	IterIndexItems() iter.Seq2[int, Vertex]

	// IterIndexItemsBackward returns an iterator over index-vertex pairs
	// in the sequence, traversing it from last to first
	// with descending indices.
	//
	// The returned iterator is never nil.
	IterIndexItemsBackward() iter.Seq2[int, Vertex]

	// Front returns the first vertex in the sequence.
	//
	// It panics if the sequence is nil or empty.
	Front() Vertex

	// Back returns the last vertex in the sequence.
	//
	// It panics if the sequence is nil or empty.
	Back() Vertex

	// ToList describes the sequence as a vertex list,
	// with items representing the vertices from first to last in the sequence.
	//
	// It returns nil if the sequence is nil or empty.
	ToList() []Vertex

	// Append returns a new sequence that
	// extends this sequence with the specified vertices.
	//
	// In particular, if there are no specified vertices,
	// Append returns this sequence itself.
	Append(v ...Vertex) VertexSequence[Vertex]

	// TruncateBack returns a new sequence that
	// trims the last vertex off this sequence.
	//
	// In particular, if this sequence is nil or has at most one vertex,
	// TruncateBack returns a non-nil empty vertex sequence.
	TruncateBack() VertexSequence[Vertex]

	// Truncate returns a new sequence that trims the vertex at index i
	// and all subsequent vertices off this sequence.
	// The returned sequence may be empty, but is never nil.
	//
	// In particular, if i is nonpositive or this sequence is nil or empty,
	// Truncate returns a non-nil empty vertex sequence.
	// If i is greater than the index of the last vertex,
	// Truncate returns this sequence itself.
	Truncate(i int) VertexSequence[Vertex]

	// ReplaceBack returns a new sequence that trims the last vertex
	// off this sequence, and then appends the specified vertices to the end.
	//
	//	vs.ReplaceBack(v...)
	//
	// is equivalent to
	//
	//	vs.TruncateBack().Append(v...)
	//
	// The special cases are consistent with
	// those of methods TruncateBack and Append.
	ReplaceBack(v ...Vertex) VertexSequence[Vertex]

	// Replace returns a new sequence that trims the vertex at index i
	// and all subsequent vertices off this sequence,
	// and then appends the specified vertices to the end.
	//
	//	vs.Replace(i, v...)
	//
	// is equivalent to
	//
	//	vs.Truncate(i).Append(v...)
	//
	// The special cases are consistent with
	// those of methods Truncate and Append.
	Replace(i int, v ...Vertex) VertexSequence[Vertex]
}

// emptyVertexSequencePanicMessage is the panic message
// indicating that the vertex sequence is empty.
const emptyVertexSequencePanicMessage = "VertexSequence[...] is empty"

// vertexSequence is an implementation of interface VertexSequence.
//
// It represents a vertex sequence recursively,
// constructed by extending vertex sequence p with vertex v.
//
// For example, the vertex sequence ("A", "B", "C") is represented as
//
//	&vertexSequence[string]{
//	    i: 2,
//	    v: "C",
//	    p: &vertexSequence[string]{
//	        i: 1,
//	        v: "B",
//	        p: &vertexSequence[string]{
//	            i: 0,
//	            v: "A",
//	            p: nil,
//	        }
//	    },
//	}
type vertexSequence[Vertex any] struct {
	i int
	v Vertex
	p *vertexSequence[Vertex]
}

// NewVertexSequence creates a new vertex sequence
// consisting of the specified vertices.
//
// If no vertices are specified, New returns a non-nil empty vertex sequence.
func NewVertexSequence[Vertex any](v ...Vertex) VertexSequence[Vertex] {
	return (*vertexSequence[Vertex])(nil).Append(v...)
}

func (vs *vertexSequence[Vertex]) Len() int {
	if vs == nil {
		return 0
	}

	return vs.i + 1
}

func (vs *vertexSequence[Vertex]) Range(handler func(v Vertex) (cont bool)) {
	if handler != nil && vs != nil {
		list := vs.ToList()
		for _, v := range list {
			if !handler(v) {
				return
			}
		}
	}
}

func (vs *vertexSequence[Vertex]) IterItems() iter.Seq[Vertex] {
	return vs.Range
}

func (vs *vertexSequence[Vertex]) RangeBackward(
	handler func(v Vertex) (cont bool),
) {
	if handler != nil && vs != nil {
		for t := vs; t != nil; t = t.p {
			if !handler(t.v) {
				return
			}
		}
	}
}

func (vs *vertexSequence[Vertex]) IterItemsBackward() iter.Seq[Vertex] {
	return vs.RangeBackward
}

func (vs *vertexSequence[Vertex]) IterIndexItems() iter.Seq2[int, Vertex] {
	return func(yield func(int, Vertex) bool) {
		if yield != nil && vs != nil {
			list := vs.ToList()
			for i, v := range list {
				if !yield(i, v) {
					return
				}
			}
		}
	}
}

func (vs *vertexSequence[Vertex]) IterIndexItemsBackward() iter.Seq2[
	int,
	Vertex,
] {
	return func(yield func(int, Vertex) bool) {
		if yield != nil && vs != nil {
			for t := vs; t != nil; t = t.p {
				if !yield(t.i, t.v) {
					return
				}
			}
		}
	}
}

func (vs *vertexSequence[Vertex]) Front() Vertex {
	if vs == nil {
		panic(errors.AutoMsg(emptyVertexSequencePanicMessage))
	}

	t := vs
	for {
		if t.p == nil {
			return t.v
		}

		t = t.p
	}
}

func (vs *vertexSequence[Vertex]) Back() Vertex {
	if vs == nil {
		panic(errors.AutoMsg(emptyVertexSequencePanicMessage))
	}

	return vs.v
}

func (vs *vertexSequence[Vertex]) ToList() []Vertex {
	if vs == nil {
		return nil
	}

	list := make([]Vertex, vs.i+1)
	for t := vs; t != nil; t = t.p {
		list[t.i] = t.v
	}

	return list
}

func (vs *vertexSequence[Vertex]) Append(v ...Vertex) VertexSequence[Vertex] {
	t := vs
	for _, x := range v {
		t = &vertexSequence[Vertex]{v: x, p: t}
		if t.p != nil {
			t.i = t.p.i + 1
		}
	}

	return t
}

func (vs *vertexSequence[Vertex]) TruncateBack() VertexSequence[Vertex] {
	if vs == nil {
		return vs
	}

	return vs.p
}

func (vs *vertexSequence[Vertex]) Truncate(i int) VertexSequence[Vertex] {
	if i <= 0 || vs == nil { // a short path
		return (*vertexSequence[Vertex])(nil)
	}

	t := vs
	for t != nil && t.i >= i {
		t = t.p
	}

	return t
}

func (vs *vertexSequence[Vertex]) ReplaceBack(
	v ...Vertex,
) VertexSequence[Vertex] {
	return vs.TruncateBack().Append(v...)
}

func (vs *vertexSequence[Vertex]) Replace(
	i int,
	v ...Vertex,
) VertexSequence[Vertex] {
	return vs.Truncate(i).Append(v...)
}
