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

// NodePath recursively represents a path of nodes.
//
// The path represented by this NodePath is constructed by
// extending path P with node N.
type NodePath struct {
	N interface{}
	P *NodePath
}

// ToList describes the path represented by this NodePath as a node list,
// whose items are the nodes along the path, from the start node to the end.
func (np *NodePath) ToList() []interface{} {
	var length int
	for x := np; x != nil; x = x.P {
		length++
	}
	list := make([]interface{}, length)
	for i, x := length-1, np; i >= 0; i, x = i-1, x.P {
		list[i] = x.N
	}
	return list
}
