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
	"sort"
	"testing"
)

type XBoolCase struct {
	X      interface{}
	Result bool
}

func TestPriorityQueueMini(t *testing.T) {
	samples := []int{0, -1, 1, 1, 2, 5, 0}
	pq := NewPriorityQueueMini(intLess)
	for _, x := range samples {
		pq.Enqueue(x)
	}
	t.Log("Data:", pq.(*priorityQueue).Data)
	sort.Ints(samples)
	for _, x := range samples {
		item, ok := pq.Dequeue()
		if !ok {
			t.Errorf("Item(%d) is not in the priority queue.", x)
			continue
		}
		if item != x {
			t.Errorf("Item(%v) != %d", item, x)
		}
	}
}

func TestPriorityQueue_Len(t *testing.T) {
	var intlPQ *priorityQueue
	n := intlPQ.Len()
	if n != 0 {
		t.Errorf("pq.Len() = %d != 0 when pq == nil.", n)
	}
	pq := NewPriorityQueue(intLess)
	n = pq.Len()
	if n != 0 {
		t.Errorf("pq.Len() = %d != 0 when pq is empty.", n)
	}
	for i := 0; i < 3; i++ {
		pq.Enqueue(i)
		n = pq.Len()
		if n != i+1 {
			t.Errorf("pq.Len() = %d != %d.", n, i+1)
		}
	}
	pq = NewPriorityQueue(intLess, 1, 2, 3, 4)
	n = pq.Len()
	if n != 4 {
		t.Errorf("pq.Len() = %d != 4.", n)
	}
}

func TestPriorityQueue_ReplaceTop(t *testing.T) {
	samples := []interface{}{1, 2, 3}
	pq := NewPriorityQueue(intLess, samples...)
	t.Log("Data:", pq.(*priorityQueue).Data)
	pq.ReplaceTop(0)
	t.Log("Data after replace top to 0:", pq.(*priorityQueue).Data)
	if x, ok := pq.Top(); x != 0 || !ok {
		t.Errorf("Top() = %v, %t != 0, true", x, ok)
	}
	pq.ReplaceTop(4)
	t.Log("Data after replace top to 4:", pq.(*priorityQueue).Data)
	if x, ok := pq.Top(); x != 2 || !ok {
		t.Errorf("Top() = %v, %t != 2, true", x, ok)
	}
}

func TestPriorityQueue_DoesContain(t *testing.T) {
	positiveSamples := []interface{}{5, 1, 1, 2, 7, 2, 0, 1, 8, 7}
	negativeSamples := []interface{}{-1, -2, 3, 4, 6, 9, 10}
	var cs []XBoolCase
	for _, x := range positiveSamples {
		cs = append(cs, XBoolCase{
			X:      x,
			Result: true,
		})
	}
	for _, x := range negativeSamples {
		cs = append(cs, XBoolCase{
			X:      x,
			Result: false,
		})
	}
	pq := NewPriorityQueueEx(intLess, positiveSamples...)
	t.Log("Data:", pq.(*priorityQueue).Data)
	for _, c := range cs {
		if pq.DoesContain(c.X) != c.Result {
			t.Errorf("pqueue.DoesContain(%v) != %t", c.X, c.Result)
		}
	}
}

func intLess(a, b interface{}) bool {
	return a.(int) < b.(int)
}
