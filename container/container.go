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

package container

import "iter"

// Container is an interface representing a general container.
type Container[Item any] interface {
	// Len returns the number of items in the container.
	Len() int

	// Range accesses the items in the container.
	// Each item is accessed once.
	// The order of access depends on the specific implementation.
	//
	// Its parameter handler is a function to deal with the item x in the
	// container and report whether to continue to access the next item.
	//
	// The client should do read-only operations on x
	// to avoid corrupting the container.
	Range(handler func(x Item) (cont bool))

	// IterItems returns an iterator over all items in the container.
	// The order of iteration is consistent with the method Range.
	//
	// The returned iterator is always non-nil.
	IterItems() iter.Seq[Item]
}
