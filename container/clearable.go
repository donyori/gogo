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

// Clearable is an interface representing a container that can be cleared.
//
// It contains two methods to remove all items in the container.
// One removes all items and asks to release the memory;
// the other removes all items but may keep the allocated space for future use.
type Clearable interface {
	// Clear removes all items in the container and asks to release the memory.
	Clear()

	// RemoveAll removes all items in the container
	// but may keep the allocated space for future use.
	RemoveAll()
}
