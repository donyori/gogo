// gogo. A Golang toolbox.
// Copyright (C) 2019-2021 Yuan Gao
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

package internal

// Path recursively represents a path.
//
// The path represented by this Path is constructed by
// extending path P with element X.
type Path struct {
	X interface{}
	P *Path
}

// ToList describes the path represented by this Path as a list,
// whose items are the elements along the path, from the start to the end.
//
// The client should guarantee that there is no change along the path
// during the call to this method.
func (p *Path) ToList() []interface{} {
	var length int
	for x := p; x != nil; x = x.P {
		length++
	}
	list := make([]interface{}, length)
	for i, x := length-1, p; i >= 0; i, x = i-1, x.P {
		list[i] = x.X
	}
	return list
}
