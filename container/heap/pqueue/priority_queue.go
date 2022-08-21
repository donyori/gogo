// gogo.  A Go (Golang) toolbox.
// Copyright (C) 2019-2022  Yuan Gao
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

	"github.com/donyori/gogo/container"
	"github.com/donyori/gogo/container/sequence/array"
	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/function/compare"
)

// PriorityQueue is an interface representing a priority queue.
//
// Its method Range may not access items in a priority-related order.
// It only guarantees that each item will be accessed once.
type PriorityQueue[Item any] interface {
	container.Container[Item]

	// Cap returns the current capacity of the queue.
	Cap() int

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

	// Clear removes all items in the queue and asks to release the memory.
	Clear()
}

const emptyQueuePanicMessage string = "priority queue is empty"

// priorityQueue is an implementation of interface PriorityQueue,
// based on container/heap.
type priorityQueue[Item any] struct {
	oha odaHeapAdapter[Item]
}

// NewPriorityQueue creates a new priority queue.
// In this priority queue, the smaller the item
// (compared by the function lessFn), the higher its priority.
//
// lessFn is a function to report whether a < b.
// It must describe a transitive ordering:
//   - if both lessFn(a, b) and lessFn(b, c) are true, then lessFn(a, c) must be true as well.
//   - if both lessFn(a, b) and lessFn(b, c) are false, then lessFn(a, c) must be false as well.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
//
// data is the initial items in the queue.
//
// It panics if lessFn is nil.
func NewPriorityQueue[Item any](lessFn compare.LessFunc[Item], data ...Item) PriorityQueue[Item] {
	if lessFn == nil {
		panic(errors.AutoMsg("lessFn is nil"))
	}
	dataCopy := make([]Item, len(data))
	copy(dataCopy, data)
	pq := &priorityQueue[Item]{
		odaHeapAdapter[Item]{
			Oda: array.WrapSliceLess(&dataCopy, lessFn),
		},
	}
	heap.Init(pq.oha)
	return pq
}

// Len returns the number of items in the queue.
func (pq *priorityQueue[Item]) Len() int {
	return pq.oha.Len()
}

// Range accesses the items in the queue.
// Each item will be accessed once.
// The order of the access may not involve priority.
//
// Its argument handler is a function to deal with the item x in the
// queue and report whether to continue to access the next item.
//
// The client should do read-only operations on x
// to avoid corrupting the priority queue.
func (pq *priorityQueue[Item]) Range(handler func(x Item) (cont bool)) {
	pq.oha.Oda.Range(handler)
}

// Cap returns the current capacity of the queue.
func (pq *priorityQueue[Item]) Cap() int {
	return pq.oha.Oda.Cap()
}

// Enqueue adds items x into the queue.
//
// Time complexity: O(m log(m + n)), where m = len(x), n = pq.Len().
func (pq *priorityQueue[Item]) Enqueue(x ...Item) {
	if pq.oha.Len() < len(x) {
		pq.oha.Oda.Append(array.SliceDynamicArray[Item](x))
		heap.Init(pq.oha)
	} else {
		for _, item := range x {
			heap.Push(pq.oha, item)
		}
	}
}

// Dequeue removes and returns the highest-priority item in the queue.
//
// It panics if the queue is nil or empty.
//
// Time complexity: O(log n), where n = pq.Len().
func (pq *priorityQueue[Item]) Dequeue() Item {
	if pq.oha.Len() == 0 {
		panic(errors.AutoMsg(emptyQueuePanicMessage))
	}
	return heap.Pop(pq.oha).(Item)
}

// Top returns the highest-priority item in the queue,
// without modifying the queue.
//
// It panics if the queue is nil or empty.
//
// Time complexity: O(1).
func (pq *priorityQueue[Item]) Top() Item {
	if pq.oha.Len() == 0 {
		panic(errors.AutoMsg(emptyQueuePanicMessage))
	}
	return pq.oha.Oda.Front()
}

// ReplaceTop replaces the highest-priority item with newX and
// returns the current highest-priority item.
//
// It panics if the queue is nil or empty.
//
// Time complexity: O(log n), where n = pq.Len().
func (pq *priorityQueue[Item]) ReplaceTop(newX Item) Item {
	if pq.oha.Len() == 0 {
		panic(errors.AutoMsg(emptyQueuePanicMessage))
	}
	pq.oha.Oda.SetFront(newX)
	heap.Fix(pq.oha, 0)
	return pq.oha.Oda.Front()
}

// Clear removes all items in the queue and asks to release the memory.
func (pq *priorityQueue[Item]) Clear() {
	pq.oha.Oda.Clear()
}

// odaHeapAdapter wraps
// github.com/donyori/gogo/container/sequence/array.OrderedDynamicArray
// to fit the interface container/heap.Interface.
type odaHeapAdapter[Item any] struct {
	Oda array.OrderedDynamicArray[Item]
}

func (oha odaHeapAdapter[Item]) Len() int {
	if oha.Oda == nil {
		return 0
	}
	return oha.Oda.Len()
}

func (oha odaHeapAdapter[Item]) Less(i, j int) bool {
	return oha.Oda.Less(i, j)
}

func (oha odaHeapAdapter[Item]) Swap(i, j int) {
	oha.Oda.Swap(i, j)
}

func (oha odaHeapAdapter[Item]) Push(x any) {
	oha.Oda.Push(x.(Item))
}

func (oha odaHeapAdapter[Item]) Pop() any {
	return oha.Oda.Pop()
}
