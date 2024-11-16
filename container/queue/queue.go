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

package queue

import (
	"fmt"
	"iter"
	"math/bits"

	"github.com/donyori/gogo/container"
	"github.com/donyori/gogo/errors"
)

// Queue is an interface representing a queue.
//
// Its method Range accesses the items from the front (the earliest added)
// to the rear (the most recently added).
type Queue[Item any] interface {
	container.Container[Item]
	container.Clearable
	container.CapacityReservable

	// Enqueue adds x to the rear of the queue.
	Enqueue(x Item)

	// Dequeue removes and returns the item at the front of the queue.
	//
	// It panics if the queue is nil or empty.
	Dequeue() Item

	// Peek returns the item at the front of the queue,
	// without modifying the queue.
	//
	// It panics if the queue is nil or empty.
	Peek() Item
}

// defaultCapacity is the default capacity of the queue.
const defaultCapacity int = 16

// emptyQueuePanicMessage is the panic message
// indicating that the queue is empty.
const emptyQueuePanicMessage = "Queue[...] is empty"

// queueSliceRing is an implementation of interface Queue,
// based on Go slice and the ring (circular buffer) structure.
type queueSliceRing[Item any] struct {
	// buf is the buffer.
	// Its length is always a power of two.
	buf []Item

	// r and w are positions for reading and writing, respectively.
	//
	// r is the index of the front of the queue.
	// buf[r] is the next item to be dequeued.
	//
	// w is the index of the item following the rear of the queue.
	// buf[w] is the location where the next enqueued item is stored.
	//
	// In particular, when the queue is empty,
	// r and w are -1 and 0, respectively;
	// when the buffer is full, r equals w.
	r, w int
}

// New creates a new queue.
//
// capacity asks to allocate enough space to hold the specified number of items.
// If capacity is nonpositive,
// New creates a queue with a small starting capacity.
func New[Item any](capacity int) Queue[Item] {
	size := defaultCapacity
	if capacity > 0 {
		size = getBufferSize(capacity)
	}
	return &queueSliceRing[Item]{
		buf: make([]Item, size),
		r:   -1,
	}
}

func (q *queueSliceRing[Item]) Len() int {
	if q.r < 0 {
		return 0
	}
	n := q.w - q.r
	if n <= 0 {
		n += len(q.buf)
	}
	return n
}

func (q *queueSliceRing[Item]) Range(handler func(x Item) (cont bool)) {
	if handler == nil || q.r < 0 {
		return
	} else if q.r < q.w {
		for i := q.r; i < q.w; i++ {
			if !handler(q.buf[i]) {
				return
			}
		}
		return
	}
	for i := q.r; i < len(q.buf); i++ {
		if !handler(q.buf[i]) {
			return
		}
	}
	for i := range q.w {
		if !handler(q.buf[i]) {
			return
		}
	}
}

func (q *queueSliceRing[Item]) IterItems() iter.Seq[Item] {
	return q.Range
}

func (q *queueSliceRing[Item]) Clear() {
	q.buf, q.r, q.w = nil, -1, 0
}

func (q *queueSliceRing[Item]) RemoveAll() {
	switch {
	case q.r < 0:
		return
	case q.r < q.w:
		clear(q.buf[q.r:q.w]) // avoid memory leak
	default:
		clear(q.buf[:q.w]) // avoid memory leak
		clear(q.buf[q.r:]) // avoid memory leak
	}
	q.r, q.w = -1, 0
}

func (q *queueSliceRing[Item]) Cap() int {
	return len(q.buf)
}

func (q *queueSliceRing[Item]) Reserve(capacity int) {
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	if capacity <= len(q.buf) {
		return
	}
	q.resize(getBufferSize(capacity))
}

func (q *queueSliceRing[Item]) Enqueue(x Item) {
	if q.r < 0 { // the queue is empty
		if len(q.buf) == 0 { // there is no buffer, possibly caused by Clear()
			q.buf = make([]Item, defaultCapacity) // make a buffer with the default capacity
		}
		q.r = 0
	} else if q.r == q.w { // the buffer is full
		n := len(q.buf) << 1
		if n < len(q.buf) {
			panic(errors.AutoMsg(fmt.Sprintf(
				"buffer size overflows; required %d (%#[1]x)", uint(n))))
		}
		q.resize(n)
	}
	q.buf[q.w] = x
	// (q.w+1)&^len(q.buf) is equivalent to (q.w+1)%len(q.buf)
	// as 0 <= q.w < len(q.buf) and len(q.buf) is a power of two.
	q.w = (q.w + 1) &^ len(q.buf)
}

func (q *queueSliceRing[Item]) Dequeue() Item {
	if q.r < 0 {
		panic(errors.AutoMsg(emptyQueuePanicMessage))
	}
	x := q.buf[q.r]
	clear(q.buf[q.r:][:1]) // avoid memory leak
	// (q.r+1)&^len(q.buf) is equivalent to (q.r+1)%len(q.buf)
	// as 0 <= q.r < len(q.buf) and len(q.buf) is a power of two.
	q.r = (q.r + 1) &^ len(q.buf)
	if q.r == q.w { // r meets w: the queue is now empty
		q.r, q.w = -1, 0
	}
	return x
}

func (q *queueSliceRing[Item]) Peek() Item {
	if q.r < 0 {
		panic(errors.AutoMsg(emptyQueuePanicMessage))
	}
	return q.buf[q.r]
}

// resize allocates a new buffer of the specified size and
// copies the data from the old buffer to the front of the new buffer.
//
// Caller should guarantee that size >= q.Len() and size is a power of two.
func (q *queueSliceRing[Item]) resize(size int) {
	buf := make([]Item, size)
	if q.r >= 0 { // the queue is nonempty
		// Copy data from q.buf to the front of buf and
		// set q.r and q.w.
		if q.r < q.w {
			copy(buf, q.buf[q.r:q.w])
			q.w -= q.r
		} else {
			copy(buf[copy(buf, q.buf[q.r:]):], q.buf[:q.w])
			q.w += len(q.buf) - q.r
		}
		q.r = 0
	}
	q.buf = buf
}

// getBufferSize returns the smallest buffer size that is:
//   - a power of two;
//   - equal to or greater than the specified capacity.
//
// It panics if the buffer size overflows.
//
// Caller should guarantee that capacity > 0.
func getBufferSize(capacity int) int {
	size := 1 << (bits.Len(uint(capacity)) - 1)
	if size != capacity {
		size <<= 1
	}
	if size < capacity {
		panic(errors.AutoMsg(fmt.Sprintf(
			"buffer size overflows; "+
				"specified capacity %d (%#[1]x), "+
				"required buffer size %d (%#[2]x) "+
				"(buffer size must be a power of two)",
			capacity, uint(size))))
	}
	return size
}
