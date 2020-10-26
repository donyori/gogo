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

// PriorityQueueMini is an interface representing a basic priority queue.
// It is called a mini version priority queue.
type PriorityQueueMini interface {
	// Len returns the number of items in the queue.
	Len() int

	// Enqueue adds items x into the queue.
	//
	// Time complexity: O(m log(m + n)), where m = len(x), n = pq.Len().
	Enqueue(x ...interface{})

	// Dequeue pops the minimum item in the queue.
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
	// It panics if the queue is nil or empty.
	//
	// Time complexity: O(1).
	Top() interface{}

	// ReplaceTop replaces the minimum item with newX.
	// It panics if the queue is nil or empty.
	//
	// Time complexity: O(log n), where n = pq.Len().
	ReplaceTop(newX interface{})

	// Maintain safeguards the underlying structure of the priority queue valid.
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
	// If x is in the queue and has been removed successfully,
	// it returns true, otherwise false.
	// If there are multiple items equals to x in the queue,
	// it removes one of them.
	// (Which one will be removed depends on the implementation.)
	//
	// Time complexity: O(n), where n = pq.Len().
	Remove(x interface{}) (ok bool)

	// Replace exchanges oldX in the queue to newX.
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

// priorityQueue is an implementation of PriorityQueueMini and PriorityQueue,
// based on container/heap.
type priorityQueue heapa.DynamicArray

// NewPriorityQueueMini creates a new mini version priority queue.
// data is the initial items in the queue.
// It panics if less is nil.
func NewPriorityQueueMini(less function.LessFunc, data ...interface{}) PriorityQueueMini {
	return PriorityQueueMini(NewPriorityQueue(less, data...))
}

// NewPriorityQueue creates a new standard priority queue.
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
	if pq.Len() < len(x) {
		pq.Data.Append(sequence.GeneralDynamicArray(x))
		pq.Maintain()
	} else {
		for _, item := range x {
			heap.Push((*heapa.DynamicArray)(pq), item)
		}
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

// priorityQueueEx is an implementation of PriorityQueueEx,
// based on container/heap.
type priorityQueueEx struct {
	priorityQueue
	EqualFn function.EqualFunc
}

// NewPriorityQueueEx creates a new extended priority queue.
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

// find searches item x in the priority queue, and returns its index.
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
