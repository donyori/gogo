// gogo. A Golang toolbox.
// Copyright (C) 2019-2022 Yuan Gao
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

	"github.com/donyori/gogo/function/compare"
)

type xBoolCase struct {
	X      interface{}
	Result bool
}

func TestPriorityQueueMini(t *testing.T) {
	samples := []int{0, -1, 1, 1, 2, 5, 0}
	pq := NewPriorityQueueMini(compare.IntLess)
	for _, x := range samples {
		pq.Enqueue(x)
	}
	t.Log("Data:", pq.(*priorityQueueMini).Heap.Data)
	sort.Ints(samples)
	for _, x := range samples {
		item := pq.Dequeue()
		if item != x {
			t.Errorf("Item(%v) != %d", item, x)
		}
	}
}

func TestPriorityQueue_Len(t *testing.T) {
	pq := NewPriorityQueue(compare.IntLess)
	n := pq.Len()
	if n != 0 {
		t.Errorf("pq.Len(): %d != 0 when pq is empty.", n)
	}
	for i := 0; i < 3; i++ {
		pq.Enqueue(i)
		n = pq.Len()
		if n != i+1 {
			t.Errorf("pq.Len(): %d != %d.", n, i+1)
		}
	}
	pq = NewPriorityQueue(compare.IntLess, 1, 2, 3, 4)
	n = pq.Len()
	if n != 4 {
		t.Errorf("pq.Len(): %d != 4.", n)
	}
}

func TestPriorityQueue_ReplaceTop(t *testing.T) {
	samples := []interface{}{1, 2, 3}
	pq := NewPriorityQueue(compare.IntLess, samples...)
	t.Log("Data:", pq.(*priorityQueue).Heap.Data)
	pq.ReplaceTop(0)
	t.Log("Data after replace top to 0:", pq.(*priorityQueue).Heap.Data)
	if x := pq.Top(); x != 0 {
		t.Errorf("Top(): %v != 0.", x)
	}
	pq.ReplaceTop(4)
	t.Log("Data after replace top to 4:", pq.(*priorityQueue).Heap.Data)
	if x := pq.Top(); x != 2 {
		t.Errorf("Top(): %v != 2.", x)
	}
}

func TestPriorityQueueEx_Contain(t *testing.T) {
	positiveSamples := []interface{}{5, 1, 1, 2, 7, 2, 0, 1, 8, 7}
	negativeSamples := []interface{}{-1, -2, 3, 4, 6, 9, 10}
	var cs []xBoolCase
	for _, x := range positiveSamples {
		cs = append(cs, xBoolCase{
			X:      x,
			Result: true,
		})
	}
	for _, x := range negativeSamples {
		cs = append(cs, xBoolCase{
			X:      x,
			Result: false,
		})
	}
	pq := NewPriorityQueueEx(compare.IntLess, nil, positiveSamples...)
	t.Log("Data:", pq.(*priorityQueueEx).Heap.Data)
	for _, c := range cs {
		if pq.Contain(c.X) != c.Result {
			t.Errorf("pqueue.Contain(%v) != %t", c.X, c.Result)
		}
	}
}

func TestPriorityQueueMini_CannotToStandard(t *testing.T) {
	mini := NewPriorityQueueMini(compare.IntLess)
	_, ok := mini.(PriorityQueue)
	if ok {
		t.Error("PriorityQueueMini can be converted to PriorityQueue via type assertion, which should not happen.")
	}
}
