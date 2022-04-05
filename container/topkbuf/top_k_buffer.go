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

package topkbuf

import (
	"fmt"

	"github.com/donyori/gogo/container/pqueue"
	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/function/compare"
)

// TopKBuffer is a buffer for saving the first K smallest items.
type TopKBuffer interface {
	// K returns the parameter K.
	K() int

	// Len returns the number of items in the buffer.
	Len() int

	// Add adds items x into the buffer.
	//
	// Time complexity: O(m log(m + n)), where m = len(x), n = tkb.Len().
	Add(x ...interface{})

	// Drain pops all items and returns them in ascending order.
	//
	// Time complexity: O(n log n), where n = tkb.Len().
	Drain() []interface{}

	// Clear discards all items and asks to release the memory.
	Clear()
}

// topKBuffer is an implementation of interface TopKBuffer,
// based on github.com/donyori/gogo/container/pqueue.PriorityQueue.
type topKBuffer struct {
	ParamK    int
	GreaterFn compare.LessFunc
	Pq        pqueue.PriorityQueue
}

// NewTopKBuffer creates a new TopKBuffer.
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
// data is the initial items in the buffer.
//
// It panics if k <= 0 or less is nil.
func NewTopKBuffer(k int, less compare.LessFunc, data ...interface{}) TopKBuffer {
	if k <= 0 {
		panic(errors.AutoMsg(fmt.Sprintf("k: %d <= 0", k)))
	}
	if less == nil {
		panic(errors.AutoMsg("less is nil"))
	}
	greater := less.Reverse()
	tkb := &topKBuffer{
		ParamK:    k,
		GreaterFn: greater,
	}
	if len(data) <= k {
		tkb.Pq = pqueue.NewPriorityQueue(greater, data...)
	} else {
		tkb.Pq = pqueue.NewPriorityQueue(greater, data[:k]...)
		for i := k; i < len(data); i++ {
			tkb.Add(data[i])
		}
	}
	return tkb
}

// K returns the parameter K.
func (tkb *topKBuffer) K() int {
	if tkb == nil {
		return 0
	}
	return tkb.ParamK
}

// Len returns the number of items in the buffer.
func (tkb *topKBuffer) Len() int {
	if tkb == nil {
		return 0
	}
	return tkb.Pq.Len()
}

// Add adds items x into the buffer.
//
// Time complexity: O(m log(m + n)), where m = len(x), n = tkb.Len().
func (tkb *topKBuffer) Add(x ...interface{}) {
	r := tkb.ParamK - tkb.Len()
	if len(x) <= r {
		tkb.Pq.Enqueue(x...)
		return
	}
	if r > 0 {
		tkb.Pq.Enqueue(x[:r]...)
	}
	for _, item := range x[r:] {
		if top := tkb.Pq.Top(); tkb.GreaterFn(top, item) {
			tkb.Pq.ReplaceTop(item)
		}
	}
}

// Drain pops all items and returns them in ascending order.
//
// Time complexity: O(n log n), where n = tkb.Len().
func (tkb *topKBuffer) Drain() []interface{} {
	n := tkb.Len()
	if n == 0 {
		return nil
	}
	result := make([]interface{}, n)
	for i := n - 1; i >= 0; i-- {
		result[i] = tkb.Pq.Dequeue()
	}
	return result
}

// Clear discards all items and asks to release the memory.
func (tkb *topKBuffer) Clear() {
	if tkb == nil {
		return
	}
	tkb.Pq.Clear()
}
