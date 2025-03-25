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

package container

// CapacityReservable is an interface representing a container
// whose capacity can be reserved.
type CapacityReservable interface {
	// Cap returns the current capacity.
	Cap() int

	// Reserve requires that the capacity be at least the specified capacity.
	//
	// If capacity is nonpositive, Reserve uses a small capacity.
	// Reserve does nothing if the new capacity is not greater than the current.
	Reserve(capacity int)
}
