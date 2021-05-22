// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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

	"github.com/donyori/gogo/adapter/containera/heapa"
	"github.com/donyori/gogo/container/sequence"
	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/function/compare"
)

// PriorityQueueMini is an interface representing a basic priority queue,
// also called a mini version priority queue.
type PriorityQueueMini interface {
	// Len returns the number of items in the queue.
	Len() int

	// Enqueue adds items x into the queue.
	//
	// Time complexity: O(m log(m + n)), where m = len(x), n = pq.Len().
	Enqueue(x ...interface{})

	// Dequeue pops the minimum item in the queue.
	//
	// It panics if the queue is nil or empty.
	//
	// Time complexity: O(log n), where n = pq.Len().
	Dequeue() interface{}
}

// PriorityQueue is an interface representing a standard priority queue.
type PriorityQueue interface {
	PriorityQueueMini

	// Cap returns the current capacity of the queue.
	Cap() int

	// Clear discards all items in the queue and asks to release the memory.
	Clear()

	// Top returns the minimum item in the queue, without popping it.
	//
	// It panics if the queue is nil or empty.
	//
	// Time complexity: O(1).
	Top() interface{}

	// ReplaceTop replaces the minimum item with newX.
	//
	// It panics if the queue is nil or empty.
	//
	// Time complexity: O(log n), where n = pq.Len().
	ReplaceTop(newX interface{})

	// Maintain safeguards the underlying structure of the priority queue valid.
	//
	// It is idempotent with respect to the priority queue.
	//
	// Time complexity: O(n), where n = pq.Len().
	Maintain()
}

// PriorityQueueEx is an interface representing an extended priority queue.
type PriorityQueueEx interface {
	PriorityQueue

	// Contain reports whether x is in the queue or not.
	//
	// Time complexity: O(n), where n = pq.Len().
	Contain(x interface{}) bool

	// Remove removes the item x in the queue.
	//
	// If x is in the queue and has been removed successfully,
	// it returns true, otherwise false.
	// If there are multiple items equals to x in the queue,
	// it removes one of them.
	// (Which one will be removed depends on the implementation.)
	//
	// Time complexity: O(n), where n = pq.Len().
	Remove(x interface{}) (ok bool)

	// Replace exchanges oldX in the queue to newX.
	//
	// If oldX is in the queue and has been replaced successfully,
	// it returns true, otherwise false.
	// If there are multiple items equals to oldX in the queue,
	// it replaces one of them.
	// (Which one will be replaced depends on the implementation.)
	//
	// Time complexity: O(n), where n = pq.Len().
	Replace(oldX, newX interface{}) (ok bool)

	// ScanWithoutOrder browses the items in the queue as fast as possible.
	//
	// Time complexity: O(n), where n = pq.Len().
	ScanWithoutOrder(handler func(x interface{}) (cont bool))
}

// priorityQueueMini is an implementation of interface PriorityQueueMini,
// based on container/heap.
type priorityQueueMini struct {
	Heap heapa.DynamicArray
}

// NewPriorityQueueMini creates a new mini version priority queue.
//
// less is a function to report whether a < b.
// It must describe a transitive ordering:
//  - if both less(a, b) and less(b, c) are true, then less(a, c) must be true as well.
//  - if both less(a, b) and less(b, c) are false, then less(a, c) must be false as well.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
//
// data is the initial items in the queue.
//
// It panics if less is nil.
func NewPriorityQueueMini(less compare.LessFunc, data ...interface{}) PriorityQueueMini {
	if less == nil {
		panic(errors.AutoMsg("less is nil"))
	}
	gda := sequence.GeneralDynamicArray(append(data[:0:0], data...))
	pqm := &priorityQueueMini{
		heapa.DynamicArray{
			Data:   &gda,
			LessFn: less,
		},
	}
	heap.Init(&pqm.Heap)
	return pqm
}

// Len returns the number of items in the queue.
func (pqm *priorityQueueMini) Len() int {
	return pqm.Heap.Len()
}

// Enqueue adds items x into the queue.
//
// Time complexity: O(m log(m + n)), where m = len(x), n = pq.Len().
func (pqm *priorityQueueMini) Enqueue(x ...interface{}) {
	if pqm.Len() < len(x) {
		pqm.Heap.Data.Append(sequence.GeneralDynamicArray(x))
		heap.Init(&pqm.Heap)
	} else {
		for _, item := range x {
			heap.Push(&pqm.Heap, item)
		}
	}
}

// Dequeue pops the minimum item in the queue.
//
// It panics if the queue is nil or empty.
//
// Time complexity: O(log n), where n = pq.Len().
func (pqm *priorityQueueMini) Dequeue() interface{} {
	if pqm.Len() == 0 {
		panic(errors.AutoMsg("priority queue is nil or empty"))
	}
	return heap.Pop(&pqm.Heap)
}

// priorityQueue is an implementation of interface PriorityQueue,
// based on container/heap.
type priorityQueue struct {
	priorityQueueMini
}

// NewPriorityQueue creates a new standard priority queue.
//
// less is a function to report whether a < b.
// It must describe a transitive ordering:
//  - if both less(a, b) and less(b, c) are true, then less(a, c) must be true as well.
//  - if both less(a, b) and less(b, c) are false, then less(a, c) must be false as well.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
//
// data is the initial items in the queue.
//
// It panics if less is nil.
func NewPriorityQueue(less compare.LessFunc, data ...interface{}) PriorityQueue {
	if less == nil {
		panic(errors.AutoMsg("less is nil"))
	}
	gda := sequence.GeneralDynamicArray(append(data[:0:0], data...))
	pq := &priorityQueue{priorityQueueMini{
		heapa.DynamicArray{
			Data:   &gda,
			LessFn: less,
		},
	}}
	pq.Maintain()
	return pq
}

// Cap returns the current capacity of the queue.
func (pq *priorityQueue) Cap() int {
	return pq.Heap.Data.Cap()
}

// Clear discards all items in the queue and asks to release the memory.
func (pq *priorityQueue) Clear() {
	pq.Heap.Data.Clear()
}

// Top returns the minimum item in the queue, without popping it.
//
// It panics if the queue is nil or empty.
//
// Time complexity: O(1).
func (pq *priorityQueue) Top() interface{} {
	if pq.Len() == 0 {
		panic(errors.AutoMsg("priority queue is nil or empty"))
	}
	return pq.Heap.Data.Front()
}

// ReplaceTop replaces the minimum item with newX.
//
// It panics if the queue is nil or empty.
//
// Time complexity: O(log n), where n = pq.Len().
func (pq *priorityQueue) ReplaceTop(newX interface{}) {
	if pq.Len() == 0 {
		panic(errors.AutoMsg("priority queue is nil or empty"))
	}
	pq.Heap.Data.SetFront(newX)
	heap.Fix(&pq.Heap, 0)
}

// Maintain safeguards the underlying structure of the priority queue valid.
//
// It is idempotent with respect to the priority queue.
//
// Time complexity: O(n), where n = pq.Len().
func (pq *priorityQueue) Maintain() {
	if pq.Len() == 0 {
		return
	}
	heap.Init(&pq.Heap)
}

// priorityQueueEx is an implementation of interface PriorityQueueEx,
// based on container/heap.
type priorityQueueEx struct {
	priorityQueue
	EqualFn compare.EqualFunc
}

// NewPriorityQueueEx creates a new extended priority queue.
//
// less is a function to report whether a < b.
// It must describe a transitive ordering:
//  - if both less(a, b) and less(b, c) are true, then less(a, c) must be true as well.
//  - if both less(a, b) and less(b, c) are false, then less(a, c) must be false as well.
//
// Note that floating-point comparison
// (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
//
// data is the initial items in the queue.
//
// It panics if less is nil.
//
// equal can be nil. If equal is nil, it will be generated via less.
func NewPriorityQueueEx(less compare.LessFunc, equal compare.EqualFunc, data ...interface{}) PriorityQueueEx {
	if less == nil {
		panic(errors.AutoMsg("less is nil"))
	}
	if equal == nil {
		equal = less.ToEqual()
	}
	gda := sequence.GeneralDynamicArray(append(data[:0:0], data...))
	pq := &priorityQueueEx{
		priorityQueue: priorityQueue{
			priorityQueueMini{heapa.DynamicArray{
				Data:   &gda,
				LessFn: less,
			}},
		},
		EqualFn: equal,
	}
	pq.Maintain()
	return pq
}

// Contain reports whether x is in the queue or not.
//
// Time complexity: O(n), where n = pq.Len().
func (pqx *priorityQueueEx) Contain(x interface{}) bool {
	return pqx.find(x) >= 0
}

// Remove removes the item x in the queue.
//
// If x is in the queue and has been removed successfully,
// it returns true, otherwise false.
// If there are multiple items equals to x in the queue,
// it removes one of them.
// (Which one will be removed depends on the implementation.)
//
// Time complexity: O(n), where n = pq.Len().
func (pqx *priorityQueueEx) Remove(x interface{}) (ok bool) {
	idx := pqx.find(x)
	if idx < 0 {
		return false
	}
	heap.Remove(&pqx.Heap, idx)
	return true
}

// Replace exchanges oldX in the queue to newX.
//
// If oldX is in the queue and has been replaced successfully,
// it returns true, otherwise false.
// If there are multiple items equals to oldX in the queue,
// it replaces one of them.
// (Which one will be replaced depends on the implementation.)
//
// Time complexity: O(n), where n = pq.Len().
func (pqx *priorityQueueEx) Replace(oldX, newX interface{}) (ok bool) {
	idx := pqx.find(oldX)
	if idx < 0 {
		return false
	}
	pqx.Heap.Data.Set(idx, newX)
	heap.Fix(&pqx.Heap, idx)
	return true
}

// ScanWithoutOrder browses the items in the queue as fast as possible.
//
// Time complexity: O(n), where n = pq.Len().
func (pqx *priorityQueueEx) ScanWithoutOrder(handler func(x interface{}) (cont bool)) {
	if handler == nil || pqx.Len() == 0 {
		return
	}
	pqx.Heap.Data.Range(handler)
}

// find searches item x in the priority queue, and returns its index.
//
// If x is not found, it returns -1.
func (pqx *priorityQueueEx) find(x interface{}) int {
	if pqx.Len() == 0 {
		return -1
	}
	if pqx.EqualFn(x, pqx.Heap.Data.Front()) {
		return 0
	}
	if pqx.Heap.LessFn(x, pqx.Heap.Data.Front()) {
		return -1
	}
	jmpMap := make(map[int]int)
	for i, n := 1, pqx.Len(); i < n; i++ {
		to := jmpMap[i]
		if to > 0 {
			maintainJumpMap(jmpMap, i, to)
			i = to
			continue
		} else if to < 0 { // if int overflow
			return -1
		}
		if pqx.EqualFn(x, pqx.Heap.Data.Get(i)) {
			return i
		}
		if pqx.Heap.LessFn(x, pqx.Heap.Data.Get(i)) {
			maintainJumpMap(jmpMap, i, i)
		}
	}
	return -1
}

// maintainJumpMap maintains the map for jumping offset, used by method find.
func maintainJumpMap(jmpMap map[int]int, from, to int) {
	delete(jmpMap, from)
	i, j := from*2+1, to*2+2
	if i < 0 { // if int overflow
		return
	}
	if j >= 0 {
		for t := jmpMap[j+1]; t > 0; t = jmpMap[t+1] {
			j = t
		}
	}
	jmpMap[i] = j
}
