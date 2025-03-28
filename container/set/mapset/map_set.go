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

package mapset

import (
	"iter"

	"github.com/donyori/gogo/container"
	"github.com/donyori/gogo/container/set"
)

// mapSet is an implementation of interface
// github.com/donyori/container/set.Set based on Go map.
type mapSet[Item comparable] struct {
	m map[Item]struct{}
}

// New creates a new Go-map-based set.
//
// The method Range of the set accesses items in random order.
// The access order in two calls to Range may be different.
//
// capacity asks to allocate enough space to hold the specified number of items.
// If capacity is nonpositive, New creates a set with a small starting capacity.
func New[Item comparable](capacity int) set.Set[Item] {
	ms := new(mapSet[Item])
	if capacity > 0 {
		ms.m = make(map[Item]struct{}, capacity)
	} else {
		ms.m = make(map[Item]struct{})
	}
	return ms
}

func (ms *mapSet[Item]) Len() int {
	return len(ms.m)
}

// Range accesses the items in the set.
// Each item is accessed once.
// The order of access is random.
//
// Its parameter handler is a function to deal with the item x in the
// set and report whether to continue to access the next item.
func (ms *mapSet[Item]) Range(handler func(x Item) (cont bool)) {
	if handler != nil {
		for x := range ms.m {
			if !handler(x) {
				return
			}
		}
	}
}

func (ms *mapSet[Item]) IterItems() iter.Seq[Item] {
	return ms.Range
}

func (ms *mapSet[Item]) Clear() {
	ms.m = make(map[Item]struct{})
}

func (ms *mapSet[Item]) RemoveAll() {
	clear(ms.m)
}

func (ms *mapSet[Item]) Filter(filter func(x Item) (keep bool)) {
	for x := range ms.m {
		if !filter(x) {
			delete(ms.m, x)
		}
	}
}

func (ms *mapSet[Item]) ContainsItem(x Item) bool {
	_, ok := ms.m[x]
	return ok
}

func (ms *mapSet[Item]) ContainsSet(s set.Set[Item]) bool {
	if s == nil {
		return true
	}
	n := s.Len()
	if n == 0 {
		return true
	} else if n > len(ms.m) {
		return false
	}
	var ok bool
	s.Range(func(x Item) (cont bool) {
		_, ok = ms.m[x]
		return ok
	})
	return ok
}

func (ms *mapSet[Item]) ContainsAny(c container.Container[Item]) bool {
	if c == nil || c.Len() == 0 {
		return false
	}
	var ok bool
	c.Range(func(x Item) (cont bool) {
		_, ok = ms.m[x]
		return !ok
	})
	return ok
}

func (ms *mapSet[Item]) Add(x ...Item) {
	for _, item := range x {
		ms.m[item] = struct{}{}
	}
}

func (ms *mapSet[Item]) Remove(x ...Item) {
	for _, item := range x {
		delete(ms.m, item)
	}
}

func (ms *mapSet[Item]) Union(s set.Set[Item]) {
	if s == nil || s.Len() == 0 {
		return
	}
	s.Range(func(x Item) (cont bool) {
		ms.m[x] = struct{}{}
		return true
	})
}

func (ms *mapSet[Item]) Intersect(s set.Set[Item]) {
	if s == nil || s.Len() == 0 {
		ms.m = make(map[Item]struct{})
		return
	}
	for x := range ms.m {
		if !s.ContainsItem(x) {
			delete(ms.m, x)
		}
	}
}

func (ms *mapSet[Item]) Subtract(s set.Set[Item]) {
	if s == nil || s.Len() == 0 {
		return
	}
	s.Range(func(x Item) (cont bool) {
		delete(ms.m, x)
		return true
	})
}

func (ms *mapSet[Item]) DisjunctiveUnion(s set.Set[Item]) {
	if s == nil || s.Len() == 0 {
		return
	}
	s.Range(func(x Item) (cont bool) {
		if _, ok := ms.m[x]; ok {
			delete(ms.m, x)
		} else {
			ms.m[x] = struct{}{}
		}
		return true
	})
}
