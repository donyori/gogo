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

// An implementation of container/heap.Interface.
type intlHeap struct {
	Data   []interface{}
	LessFn LessFunc
}

func (h *intlHeap) Len() int {
	if h == nil {
		return 0
	}
	return len(h.Data)
}

func (h *intlHeap) Cap() int {
	if h == nil {
		return 0
	}
	return cap(h.Data)
}

func (h *intlHeap) Less(i, j int) bool {
	return h.LessFn(h.Data[i], h.Data[j])
}

func (h *intlHeap) Swap(i, j int) {
	h.Data[i], h.Data[j] = h.Data[j], h.Data[i]
}

func (h *intlHeap) Push(x interface{}) {
	h.Data = append(h.Data, x)
}

func (h *intlHeap) Pop() interface{} {
	last := len(h.Data) - 1
	item := h.Data[last]
	h.Data[last] = nil // avoid memory leak
	h.Data = h.Data[:last]
	return item
}

func (h *intlHeap) Clear() {
	if h == nil {
		return
	}
	h.Data = nil
}
