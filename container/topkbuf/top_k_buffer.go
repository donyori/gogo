// gogo. A Golang toolbox.
// Copyright (C) $date.year Yuan Gao
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

// Package topkbuf provides OOP-style buffer for saving the first K smallest items.
package topkbuf

import (
	"errors"
	"fmt"

	"github.com/donyori/gogo/container/pqueue"
)

// Buffer for saving the first K smallest items.
type TopKBuffer interface {
	// Return the parameter K.
	// It returns 0 if the buffer is nil.
	GetK() int
	// Return the number of items in the buffer.
	// It returns 0 if the buffer is nil.
	Len() int
	// Add an item, x, into the buffer.
	// Time complexity: O(log n), where n = b.Len().
	Add(x interface{})
	// Pop all items and return in ascending order.
	// Time complexity: O(n log n), where n = b.Len().
	Drain() []interface{}
	// Discard all items and release the memory.
	Clear()
}

// An implementation of TopKBuffer,
// based on github.com/donyori/gogo/container/pqueue.PriorityQueue.
type topKBuffer struct {
	K      int
	LessFn LessFunc
	PQ     pqueue.PriorityQueue
}

// Create a new TopKBuffer.
// data is the initial items in the buffer.
// It panics if k <= 0 or lessFunc is nil.
func NewTopKBuffer(k int, lessFunc LessFunc, data ...interface{}) TopKBuffer {
	if k <= 0 {
		panic(fmt.Errorf("K = %d <= 0", k))
	}
	if lessFunc == nil {
		panic(errors.New("lessFunc is nil"))
	}
	lessFn := lessFunc.Reverse()
	tkb := &topKBuffer{
		K:      k,
		LessFn: lessFn,
	}
	if len(data) <= k {
		tkb.PQ = pqueue.NewPriorityQueue(lessFn, data...)
	} else {
		tkb.PQ = pqueue.NewPriorityQueue(lessFn)
		for _, x := range data {
			tkb.Add(x)
		}
	}
	return tkb
}

func (tkb *topKBuffer) GetK() int {
	if tkb == nil {
		return 0
	}
	return tkb.K
}

func (tkb *topKBuffer) Len() int {
	if tkb == nil {
		return 0
	}
	return tkb.PQ.Len()
}

func (tkb *topKBuffer) Add(x interface{}) {
	if tkb.Len() < tkb.K {
		tkb.PQ.Enqueue(x)
		return
	}
	if top, _ := tkb.PQ.Top(); tkb.LessFn(top, x) {
		tkb.PQ.ReplaceTop(x)
	}
}

func (tkb *topKBuffer) Drain() []interface{} {
	n := tkb.Len()
	if n == 0 {
		return nil
	}
	result := make([]interface{}, n)
	for i := n - 1; i >= 0; i-- {
		result[i], _ = tkb.PQ.Dequeue()
	}
	return result
}

func (tkb *topKBuffer) Clear() {
	if tkb == nil {
		return
	}
	tkb.PQ.Clear()
}
