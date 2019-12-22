// gogo. A Golang toolbox.
// Copyright (C) 2019 Yuan Gao
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
	"errors"

	"github.com/donyori/gogo/function"
)

// Priority queue, mini version.
// It contains two basic method: Enqueue and Dequeue.
type PriorityQueueMini interface {
	// Add an item, x, into the priority queue.
	// Time complexity: O(log n), where n = pq.Len().
	Enqueue(x interface{})
	// Pop the minimum item in the priority queue.
	// x is the minimum item. It is nil if the queue is nil or empty.
	// ok is an indicator. It is false if the queue is nil or empty, and true otherwise.
	// Time complexity: O(log n), where n = pq.Len().
	Dequeue() (x interface{}, ok bool)
}

// Priority queue (standard version).
type PriorityQueue interface {
	PriorityQueueMini
	// Return the number of items in the queue.
	// It returns 0 if the queue is nil.
	Len() int
	// Return the capacity of the queue.
	// It returns 0 if the queue is nil.
	Cap() int
	// Discard all items in the queue and release the memory.
	Clear()
	// Return the minimum item in the queue, without popping it.
	// x is the minimum item. It is nil if the queue is nil or empty.
	// ok is an indicator. It is false if the queue is nil or empty, and true otherwise.
	Top() (x interface{}, ok bool)
	// Replace the minimum item with newX.
	// If the queue is nil or empty, it returns false.
	// Otherwise it replaces the minimum item and returns true.
	// Time complexity: O(log n), where n = pq.Len().
	ReplaceTop(newX interface{}) (ok bool)
	// Maintain the priority queue to keep its underlying structure valid.
	// It is idempotent with respect to the priority queue.
	// Time complexity: O(n), where n = pq.Len().
	Maintain()
}

// Priority queue, ex version.
type PriorityQueueEx interface {
	PriorityQueue
	// Returns true if x is in the queue, otherwise false.
	// Time complexity: O(n), where n = pq.Len().
	DoesContain(x interface{}) bool
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
	DisorderlyScan(handler func(x interface{}) (doesContinue bool))
}

// Export github.com/donyori/gogo/function.LessFunc.
type LessFunc = function.LessFunc

// An implementation of PriorityQueueMini, PriorityQueue and PriorityQueueEx,
// based on container/heap.
type priorityQueue intlHeap

// Create a new priority queue (mini version).
// data is the initial items in the queue.
// It panics if lessFunc is nil.
func NewPriorityQueueMini(lessFunc LessFunc, data ...interface{}) PriorityQueueMini {
	return PriorityQueueMini(NewPriorityQueueEx(lessFunc, data...))
}

// Create a new priority queue (standard version).
// data is the initial items in the queue.
// It panics if lessFunc is nil.
func NewPriorityQueue(lessFunc LessFunc, data ...interface{}) PriorityQueue {
	return PriorityQueue(NewPriorityQueueEx(lessFunc, data...))
}

// Create a new priority queue (ex version).
// data is the initial items in the queue.
// It panics if lessFunc is nil.
func NewPriorityQueueEx(lessFunc LessFunc, data ...interface{}) PriorityQueueEx {
	if lessFunc == nil {
		panic(errors.New("lessFunc is nil"))
	}
	pq := &priorityQueue{
		Data:     append(data[:0:0], data...),
		LessFunc: lessFunc,
	}
	pq.Maintain()
	return pq
}

func (pq *priorityQueue) Enqueue(x interface{}) {
	heap.Push((*intlHeap)(pq), x)
}

func (pq *priorityQueue) Dequeue() (x interface{}, ok bool) {
	if pq.Len() == 0 {
		return nil, false
	}
	return heap.Pop((*intlHeap)(pq)), true
}

func (pq *priorityQueue) Len() int {
	return (*intlHeap)(pq).Len()
}

func (pq *priorityQueue) Cap() int {
	return (*intlHeap)(pq).Cap()
}

func (pq *priorityQueue) Clear() {
	(*intlHeap)(pq).Clear()
}

func (pq *priorityQueue) Top() (x interface{}, ok bool) {
	if pq.Len() == 0 {
		return nil, false
	}
	return pq.Data[0], true
}

func (pq *priorityQueue) ReplaceTop(newX interface{}) (ok bool) {
	if pq.Len() == 0 {
		return false
	}
	pq.Data[0] = newX
	heap.Fix((*intlHeap)(pq), 0)
	return true
}

func (pq *priorityQueue) Maintain() {
	if pq.Len() == 0 {
		return
	}
	heap.Init((*intlHeap)(pq))
}

func (pq *priorityQueue) DoesContain(x interface{}) bool {
	return pq.find(x) >= 0
}

func (pq *priorityQueue) Remove(x interface{}) (ok bool) {
	idx := pq.find(x)
	if idx < 0 {
		return false
	}
	heap.Remove((*intlHeap)(pq), idx)
	return true
}

func (pq *priorityQueue) Replace(oldX, newX interface{}) (ok bool) {
	idx := pq.find(oldX)
	if idx < 0 {
		return false
	}
	pq.Data[idx] = newX
	heap.Fix((*intlHeap)(pq), idx)
	return true
}

func (pq *priorityQueue) DisorderlyScan(handler func(x interface{}) (doesContinue bool)) {
	if handler == nil || pq.Len() == 0 {
		return
	}
	for _, x := range pq.Data {
		if !handler(x) {
			return
		}
	}
}

func (pq *priorityQueue) find(x interface{}) int {
	if pq.Len() == 0 {
		return -1
	}
	if x == pq.Data[0] {
		return 0
	}
	if pq.LessFunc(x, pq.Data[0]) {
		return -1
	}
	jmpMap := make(map[int]int)
	for i, n := 1, pq.Len(); i < n; i++ {
		to := jmpMap[i]
		if to > 0 {
			maintainJumpMap(jmpMap, i, to)
			i = to
			continue
		} else if to < 0 { // if int overflow
			return -1
		}
		if x == pq.Data[i] {
			return i
		}
		if pq.LessFunc(x, pq.Data[i]) {
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
