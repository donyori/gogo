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

package set

import "github.com/donyori/gogo/container"

// Set is an interface representing a set.
//
// Set guarantees to contain no duplicate items.
type Set[Item any] interface {
	container.Container[Item]
	container.Filter[Item]

	// ContainsItem reports whether the item x is in the set.
	ContainsItem(x Item) bool

	// ContainsSet reports whether the set s is a subset of this set.
	ContainsSet(s Set[Item]) bool

	// ContainsAny reports whether any item in c is in this set.
	//
	// If c is nil or empty, it returns false.
	ContainsAny(c container.Container[Item]) bool

	// Add adds x to the set.
	Add(x ...Item)

	// Remove removes x from the set.
	//
	// It does nothing for the items in x that are not in the set.
	Remove(x ...Item)

	// Union adds the items in s to this set.
	// That is, perform the following assignment:
	//
	//	thisSet = thisSet ∪ s
	Union(s Set[Item])

	// Intersect removes the items not in s.
	// That is, perform the following assignment:
	//
	//	thisSet = thisSet ∩ s
	Intersect(s Set[Item])

	// Subtract removes the items in s.
	// That is, perform the following assignment:
	//
	//	thisSet = thisSet \ s
	Subtract(s Set[Item])

	// DisjunctiveUnion removes the items both in s and this set and
	// adds the items only in s.
	// That is, perform the following assignment:
	//
	//	thisSet = thisSet △ s
	DisjunctiveUnion(s Set[Item])

	// Clear removes all items in the set and asks to release the memory.
	Clear()
}
