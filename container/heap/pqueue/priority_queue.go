// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2024  Yuan Gao
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

package pqueue

import (
	"container/heap"
	"iter"

	"github.com/donyori/gogo/container"
	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/function/compare"
)

// PriorityQueue is an interface representing a priority queue.
//
// Its method Range may not access items in a priority-related order.
// It only guarantees that each item is accessed once.
type PriorityQueue[Item any] interface {
	container.Container[Item]
	container.Clearable
	container.CapacityReservable

	// Enqueue adds items x into the queue.
	//
	// Time complexity: O(m log(m + n)), where m = len(x), n = pq.Len().
	Enqueue(x ...Item)

	// Dequeue removes and returns the highest-priority item in the queue.
	//
	// It panics if the queue is nil or empty.
	//
	// Time complexity: O(log n), where n = pq.Len().
	Dequeue() Item

	// Top returns the highest-priority item in the queue,
	// without modifying the queue.
	//
	// It panics if the queue is nil or empty.
	//
	// Time complexity: O(1).
	Top() Item

	// ReplaceTop replaces the highest-priority item with newX and
	// returns the current highest-priority item.
	//
	// It panics if the queue is nil or empty.
	//
	// Time complexity: O(log n), where n = pq.Len().
	ReplaceTop(newX Item) Item
}

// defaultCapacity is the default capacity of the priority queue.
const defaultCapacity int = 16

// emptyQueuePanicMessage is the panic message
// indicating that the priority queue is empty.
const emptyQueuePanicMessage = "priority queue is empty"

// priorityQueue is an implementation of interface PriorityQueue,
// based on container/heap.
type priorityQueue[Item any] struct {
	oha odaHeapAdapter[Item]
}

// New creates a new priority queue.
// In this priority queue, the smaller the item
// (compared by the function lessFn), the higher its priority.
//
// lessFn is a function to report whether a < b.
// It must describe a strict weak ordering.
// See <https://en.wikipedia.org/wiki/Weak_ordering#Strict_weak_orderings>
// for details.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a strict weak ordering when not-a-number (NaN) values are involved.
//
// capacity asks to allocate enough space to hold the specified number of items.
// If capacity is nonpositive,
// New creates a priority queue with a small starting capacity.
//
// New panics if lessFn is nil.
func New[Item any](
	lessFn compare.LessFunc[Item],
	capacity int,
) PriorityQueue[Item] {
	if lessFn == nil {
		panic(errors.AutoMsg("lessFn is nil"))
	}
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	s := make([]Item, 0, capacity)
	pq := &priorityQueue[Item]{
		odaHeapAdapter[Item]{
			ODA: array.WrapSlice(&s, lessFn, nil),
		},
	}
	heap.Init(pq.oha)
	return pq
}

func (pq *priorityQueue[Item]) Len() int {
	return pq.oha.Len()
}

// Range accesses the items in the queue.
// Each item is accessed once.
// The order of access may not involve priority.
//
// Its parameter handler is a function to deal with the item x in the
// queue and report whether to continue to access the next item.
//
// The client should do read-only operations on x
// to avoid corrupting the priority queue.
func (pq *priorityQueue[Item]) Range(handler func(x Item) (cont bool)) {
	pq.oha.ODA.Range(handler)
}

func (pq *priorityQueue[Item]) IterItems() iter.Seq[Item] {
	return pq.Range
}

func (pq *priorityQueue[Item]) Clear() {
	pq.oha.ODA.Clear()
}

func (pq *priorityQueue[Item]) RemoveAll() {
	pq.oha.ODA.RemoveAll()
}

func (pq *priorityQueue[Item]) Cap() int {
	return pq.oha.ODA.Cap()
}

func (pq *priorityQueue[Item]) Reserve(capacity int) {
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	pq.oha.ODA.Reserve(capacity)
}

func (pq *priorityQueue[Item]) Enqueue(x ...Item) {
	if pq.oha.Len() < len(x) {
		pq.oha.ODA.Append((*array.SliceDynamicArray[Item])(&x))
		heap.Init(pq.oha)
	} else {
		for _, item := range x {
			heap.Push(pq.oha, item)
		}
	}
}

func (pq *priorityQueue[Item]) Dequeue() Item {
	if pq.oha.Len() == 0 {
		panic(errors.AutoMsg(emptyQueuePanicMessage))
	}
	return heap.Pop(pq.oha).(Item)
}

func (pq *priorityQueue[Item]) Top() Item {
	if pq.oha.Len() == 0 {
		panic(errors.AutoMsg(emptyQueuePanicMessage))
	}
	return pq.oha.ODA.Front()
}

func (pq *priorityQueue[Item]) ReplaceTop(newX Item) Item {
	if pq.oha.Len() == 0 {
		panic(errors.AutoMsg(emptyQueuePanicMessage))
	}
	pq.oha.ODA.SetFront(newX)
	heap.Fix(pq.oha, 0)
	return pq.oha.ODA.Front()
}

// odaHeapAdapter wraps
// github.com/donyori/gogo/container/sequence/array.OrderedDynamicArray
// to fit the interface container/heap.Interface.
type odaHeapAdapter[Item any] struct {
	ODA array.OrderedDynamicArray[Item]
}

func (oha odaHeapAdapter[Item]) Len() int {
	if oha.ODA == nil {
		return 0
	}
	return oha.ODA.Len()
}

func (oha odaHeapAdapter[Item]) Less(i, j int) bool {
	return oha.ODA.Less(i, j)
}

func (oha odaHeapAdapter[Item]) Swap(i, j int) {
	oha.ODA.Swap(i, j)
}

func (oha odaHeapAdapter[Item]) Push(x any) {
	oha.ODA.Push(x.(Item))
}

func (oha odaHeapAdapter[Item]) Pop() any {
	return oha.ODA.Pop()
}
