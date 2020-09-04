// gogo. A Golang toolbox.
// Copyright (C) 2019-2020 Yuan Gao
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
	"github.com/donyori/gogo/function"
)

// Priority queue, mini version.
type PriorityQueueMini interface {
	// Return the number of items in the queue.
	Len() int

	// Add items x into the queue.
	// Time complexity: O(m log(m + n)), where m = len(x), n = pq.Len().
	Enqueue(x ...interface{})

	// Pop the minimum item in the queue.
	// It panics if the queue is nil or empty.
	// Time complexity: O(log n), where n = pq.Len().
	Dequeue() interface{}
}

// Priority queue (standard version).
type PriorityQueue interface {
	PriorityQueueMini

	// Return the capacity of the queue.
	Cap() int

	// Discard all items in the queue and release the memory.
	Clear()

	// Return the minimum item in the queue, without popping it.
	// It panics if the queue is nil or empty.
	// Time complexity: O(1).
	Top() interface{}

	// Replace the minimum item with newX.
	// It panics if the queue is nil or empty.
	// Time complexity: O(log n), where n = pq.Len().
	ReplaceTop(newX interface{})

	// Maintain the priority queue to keep its underlying structure valid.
	// It is idempotent with respect to the priority queue.
	// Time complexity: O(n), where n = pq.Len().
	Maintain()
}

// Priority queue, extra version.
type PriorityQueueEx interface {
	PriorityQueue

	// Returns true if x is in the queue, otherwise false.
	// Time complexity: O(n), where n = pq.Len().
	Contain(x interface{}) bool

	// Remove the item x in the queue.
	// If x is in the queue and has been removed successfully, it returns true, otherwise false.
	// If there are multiple items equals to x in the queue, it removes one of them.
	// Time complexity: O(n), where n = pq.Len().
	Remove(x interface{}) (ok bool)

	// Replace oldX in the queue with newX.
	// If oldX is in the queue and has been replaced successfully, it returns true, otherwise false.
	// If there are multiple items equals to oldX in the queue, it replaces one of them.
	// Time complexity: O(n), where n = pq.Len().
	Replace(oldX, newX interface{}) (ok bool)

	// Scan the items in the queue as fast as possible.
	// Time complexity: O(n), where n = pq.Len().
	ScanWithoutOrder(handler func(x interface{}) (cont bool))
}

// An implementation of PriorityQueueMini and PriorityQueue,
// based on container/heap.
type priorityQueue heapa.DynamicArray

// Create a new priority queue (mini version).
// data is the initial items in the queue.
// It panics if less is nil.
func NewPriorityQueueMini(less function.LessFunc, data ...interface{}) PriorityQueueMini {
	return PriorityQueueMini(NewPriorityQueue(less, data...))
}

// Create a new priority queue (standard version).
// data is the initial items in the queue.
// It panics if less is nil.
func NewPriorityQueue(less function.LessFunc, data ...interface{}) PriorityQueue {
	if less == nil {
		panic(errors.AutoMsg("less is nil"))
	}
	gda := sequence.GeneralDynamicArray(append(data[:0:0], data...))
	pq := &priorityQueue{
		Data:   &gda,
		LessFn: less,
	}
	pq.Maintain()
	return pq
}

func (pq *priorityQueue) Len() int {
	return (*heapa.DynamicArray)(pq).Len()
}

func (pq *priorityQueue) Enqueue(x ...interface{}) {
	for _, item := range x {
		heap.Push((*heapa.DynamicArray)(pq), item)
	}
}

func (pq *priorityQueue) Dequeue() interface{} {
	if pq.Len() == 0 {
		panic(errors.AutoMsg("priority queue is nil or empty"))
	}
	return heap.Pop((*heapa.DynamicArray)(pq))
}

func (pq *priorityQueue) Cap() int {
	if pq == nil || pq.Data == nil {
		return 0
	}
	return pq.Data.Cap()
}

func (pq *priorityQueue) Clear() {
	pq.Data.Clear()
}

func (pq *priorityQueue) Top() interface{} {
	if pq.Len() == 0 {
		panic(errors.AutoMsg("priority queue is nil or empty"))
	}
	return pq.Data.Front()
}

func (pq *priorityQueue) ReplaceTop(newX interface{}) {
	if pq.Len() == 0 {
		panic(errors.AutoMsg("priority queue is nil or empty"))
	}
	pq.Data.SetFront(newX)
	heap.Fix((*heapa.DynamicArray)(pq), 0)
}

func (pq *priorityQueue) Maintain() {
	if pq.Len() == 0 {
		return
	}
	heap.Init((*heapa.DynamicArray)(pq))
}

// An implementation of PriorityQueueEx,
// based on container/heap.
type priorityQueueEx struct {
	priorityQueue
	EqualFn function.EqualFunc
}

// Create a new priority queue (extra version).
// data is the initial items in the queue.
// It panics if less is nil.
// equal can be nil. If equal is nil, it will be generated via less.
func NewPriorityQueueEx(less function.LessFunc, equal function.EqualFunc, data ...interface{}) PriorityQueueEx {
	if less == nil {
		panic(errors.AutoMsg("less is nil"))
	}
	if equal == nil {
		equal = function.GenerateEqualViaLess(less)
	}
	gda := sequence.GeneralDynamicArray(append(data[:0:0], data...))
	pq := &priorityQueueEx{
		priorityQueue: priorityQueue{
			Data:   &gda,
			LessFn: less,
		},
		EqualFn: equal,
	}
	pq.Maintain()
	return pq
}

func (pqx *priorityQueueEx) Contain(x interface{}) bool {
	return pqx.find(x) >= 0
}

func (pqx *priorityQueueEx) Remove(x interface{}) (ok bool) {
	idx := pqx.find(x)
	if idx < 0 {
		return false
	}
	heap.Remove((*heapa.DynamicArray)(&pqx.priorityQueue), idx)
	return true
}

func (pqx *priorityQueueEx) Replace(oldX, newX interface{}) (ok bool) {
	idx := pqx.find(oldX)
	if idx < 0 {
		return false
	}
	pqx.Data.Set(idx, newX)
	heap.Fix((*heapa.DynamicArray)(&pqx.priorityQueue), idx)
	return true
}

func (pqx *priorityQueueEx) ScanWithoutOrder(handler func(x interface{}) (cont bool)) {
	if handler == nil || pqx.Len() == 0 {
		return
	}
	pqx.Data.Scan(handler)
}

// Find item x in the priority queue, and return its index.
// If x is not found, it returns -1.
func (pqx *priorityQueueEx) find(x interface{}) int {
	if pqx.Len() == 0 {
		return -1
	}
	if pqx.EqualFn(x, pqx.Data.Front()) {
		return 0
	}
	if pqx.LessFn(x, pqx.Data.Front()) {
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
		if pqx.EqualFn(x, pqx.Data.Get(i)) {
			return i
		}
		if pqx.LessFn(x, pqx.Data.Get(i)) {
			maintainJumpMap(jmpMap, i, i)
		}
	}
	return -1
}

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
