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

package topkbuf

import (
	"fmt"
	"iter"

	"github.com/donyori/gogo/container"
	"github.com/donyori/gogo/container/heap/pqueue"
	"github.com/donyori/gogo/errors"
	"github.com/donyori/gogo/function/compare"
)

// TopKBuffer is a buffer for storing the top-K greatest items.
//
// Its method Range may not access items in ascending or descending order.
// It only guarantees that each item is accessed once.
type TopKBuffer[Item any] interface {
	container.Container[Item]
	container.Clearable

	// K returns the parameter K,
	// which limits the maximum number of items the buffer can hold.
	K() int

	// Add adds items x into the buffer.
	//
	// Time complexity: O(m log(m + n)), where m = len(x), n = tkb.Len().
	Add(x ...Item)

	// Drain pops all items and returns them in descending order.
	//
	// Time complexity: O(n log n), where n = tkb.Len().
	Drain() []Item
}

// topKBuffer is an implementation of interface TopKBuffer,
// based on github.com/donyori/gogo/container/heap/pqueue.PriorityQueue.
//
// Never use it with a nil pointer.
type topKBuffer[Item any] struct {
	k      int
	lessFn compare.LessFunc[Item]
	pq     pqueue.PriorityQueue[Item]
}

// New creates a new TopKBuffer with the parameter k.
// The buffer holds the top-k greatest items.
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
// New panics if lessFn is nil or k is nonpositive.
func New[Item any](lessFn compare.LessFunc[Item], k int) TopKBuffer[Item] {
	if lessFn == nil {
		panic(errors.AutoMsg("lessFn is nil"))
	} else if k <= 0 {
		panic(errors.AutoMsg(fmt.Sprintf("k (%d) is nonpositive", k)))
	}
	return &topKBuffer[Item]{
		k:      k,
		lessFn: lessFn,
		pq:     pqueue.New(lessFn, k),
	}
}

func (tkb *topKBuffer[Item]) Len() int {
	return tkb.pq.Len()
}

// Range accesses the items in the buffer.
// Each item is accessed once.
// The order of access may not be ascending or descending.
//
// Its parameter handler is a function to deal with the item x in the
// buffer and report whether to continue to access the next item.
//
// The client should do read-only operations on x
// to avoid corrupting the top-K buffer.
func (tkb *topKBuffer[Item]) Range(handler func(x Item) (cont bool)) {
	tkb.pq.Range(handler)
}

func (tkb *topKBuffer[Item]) IterItems() iter.Seq[Item] {
	return tkb.Range
}

func (tkb *topKBuffer[Item]) Clear() {
	tkb.pq.Clear()
}

func (tkb *topKBuffer[Item]) RemoveAll() {
	tkb.pq.RemoveAll()
}

func (tkb *topKBuffer[Item]) K() int {
	return tkb.k
}

func (tkb *topKBuffer[Item]) Add(x ...Item) {
	if len(x) == 0 {
		return
	}
	tkb.pq.Reserve(tkb.k)
	r := tkb.k - tkb.pq.Len()
	if len(x) <= r {
		tkb.pq.Enqueue(x...)
		return
	} else if r > 0 {
		tkb.pq.Enqueue(x[:r]...)
	}
	for _, item := range x[r:] {
		if top := tkb.pq.Top(); tkb.lessFn(top, item) {
			tkb.pq.ReplaceTop(item)
		}
	}
}

func (tkb *topKBuffer[Item]) Drain() []Item {
	n := tkb.pq.Len()
	if n == 0 {
		return nil
	}
	result := make([]Item, n)
	for i := n - 1; i >= 0; i-- {
		result[i] = tkb.pq.Dequeue()
	}
	return result
}
