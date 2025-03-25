// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2025  Yuan Gao
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

package stack

import (
	"iter"

	"github.com/donyori/gogo/container"
	"github.com/donyori/gogo/errors"
)

// Stack is an interface representing a stack.
//
// Its method Range accesses the items
// from the most recently added to the earliest added.
type Stack[Item any] interface {
	container.Container[Item]
	container.Clearable
	container.CapacityReservable

	// Push adds x to the stack.
	Push(x Item)

	// Pop removes and returns the most recently added item.
	//
	// It panics if the stack is nil or empty.
	Pop() Item

	// Peek returns the most recently added item, without modifying the stack.
	//
	// It panics if the stack is nil or empty.
	Peek() Item
}

// defaultCapacity is the default capacity of the stack.
const defaultCapacity int = 16

// emptyStackPanicMessage is the panic message
// indicating that the stack is empty.
const emptyStackPanicMessage = "Stack[...] is empty"

// stackSlice is an implementation of interface Stack,
// based on Go slice.
type stackSlice[Item any] struct {
	buf []Item
}

// New creates a new stack.
//
// capacity asks to allocate enough space to hold the specified number of items.
// If capacity is nonpositive,
// New creates a stack with a small starting capacity.
func New[Item any](capacity int) Stack[Item] {
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	return &stackSlice[Item]{
		buf: make([]Item, 0, capacity),
	}
}

func (s *stackSlice[Item]) Len() int {
	return len(s.buf)
}

func (s *stackSlice[Item]) Range(handler func(x Item) (cont bool)) {
	if handler != nil {
		for i := len(s.buf) - 1; i >= 0; i-- {
			if !handler(s.buf[i]) {
				return
			}
		}
	}
}

func (s *stackSlice[Item]) IterItems() iter.Seq[Item] {
	return s.Range
}

func (s *stackSlice[Item]) Clear() {
	s.buf = nil
}

func (s *stackSlice[Item]) RemoveAll() {
	clear(s.buf) // avoid memory leak
	s.buf = s.buf[:0]
}

func (s *stackSlice[Item]) Cap() int {
	return cap(s.buf)
}

func (s *stackSlice[Item]) Reserve(capacity int) {
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	if capacity <= cap(s.buf) {
		return
	}
	buf := make([]Item, len(s.buf), capacity)
	copy(buf, s.buf)
	s.buf = buf
}

func (s *stackSlice[Item]) Push(x Item) {
	s.buf = append(s.buf, x)
}

func (s *stackSlice[Item]) Pop() Item {
	if len(s.buf) == 0 {
		panic(errors.AutoMsg(emptyStackPanicMessage))
	}
	x := s.buf[len(s.buf)-1]
	clear(s.buf[len(s.buf)-1:]) // avoid memory leak
	s.buf = s.buf[:len(s.buf)-1]
	return x
}

func (s *stackSlice[Item]) Peek() Item {
	if len(s.buf) == 0 {
		panic(errors.AutoMsg(emptyStackPanicMessage))
	}
	return s.buf[len(s.buf)-1]
}
